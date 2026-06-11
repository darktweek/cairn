package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/darktweek/cairn/internal/config"
)

// OIDCService implements the OpenID Connect Authorization Code flow with PKCE,
// using only the standard library. Claims are read from the userinfo endpoint
// (the token response and userinfo call both come directly from the IdP over
// TLS), so no local JWT signature verification is required for a confidential
// client.
type OIDCService interface {
	Config(ctx context.Context) OIDCConfig
	AuthURL(ctx context.Context) (authURL, state string, err error)
	Exchange(ctx context.Context, state, code string) (*OIDCUser, error)
}

// OIDCUser holds the identity claims returned by the provider.
type OIDCUser struct {
	Subject  string
	Email    string
	Username string
	Name     string
}

type oidcDiscovery struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
	fetchedAt             time.Time
}

type oidcState struct {
	verifier  string
	createdAt time.Time
}

type oidcService struct {
	cfg      *config.Config
	settings SettingsService
	client   *http.Client

	mu        sync.Mutex
	discovery map[string]*oidcDiscovery // keyed by issuer
	states    map[string]oidcState      // state → PKCE verifier
}

func newOIDCService(cfg *config.Config, settings SettingsService) OIDCService {
	s := &oidcService{
		cfg:       cfg,
		settings:  settings,
		client:    &http.Client{Timeout: 10 * time.Second},
		discovery: make(map[string]*oidcDiscovery),
		states:    make(map[string]oidcState),
	}
	return s
}

func (s *oidcService) Config(ctx context.Context) OIDCConfig {
	return s.settings.OIDC(ctx)
}

func (s *oidcService) redirectURI() string {
	return strings.TrimRight(s.cfg.BaseURL, "/") + "/api/auth/sso/callback"
}

func (s *oidcService) discover(ctx context.Context, issuer string) (*oidcDiscovery, error) {
	s.mu.Lock()
	if d, ok := s.discovery[issuer]; ok && time.Since(d.fetchedAt) < time.Hour {
		s.mu.Unlock()
		return d, nil
	}
	s.mu.Unlock()

	wellKnown := strings.TrimRight(issuer, "/") + "/.well-known/openid-configuration"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wellKnown, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oidc discovery: status %d", resp.StatusCode)
	}
	var d oidcDiscovery
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return nil, fmt.Errorf("oidc discovery decode: %w", err)
	}
	if d.AuthorizationEndpoint == "" || d.TokenEndpoint == "" {
		return nil, fmt.Errorf("oidc discovery: incomplete document")
	}
	d.fetchedAt = time.Now()

	s.mu.Lock()
	s.discovery[issuer] = &d
	s.mu.Unlock()
	return &d, nil
}

func (s *oidcService) AuthURL(ctx context.Context) (string, string, error) {
	cfg := s.settings.OIDC(ctx)
	if !cfg.Enabled() {
		return "", "", ErrNotFound
	}
	d, err := s.discover(ctx, cfg.Issuer)
	if err != nil {
		return "", "", err
	}

	state := randToken()
	verifier := randToken()
	challenge := pkceChallenge(verifier)

	s.mu.Lock()
	s.gcStatesLocked()
	s.states[state] = oidcState{verifier: verifier, createdAt: time.Now()}
	s.mu.Unlock()

	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", cfg.ClientID)
	q.Set("redirect_uri", s.redirectURI())
	q.Set("scope", cfg.Scopes)
	q.Set("state", state)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")

	return d.AuthorizationEndpoint + "?" + q.Encode(), state, nil
}

func (s *oidcService) Exchange(ctx context.Context, state, code string) (*OIDCUser, error) {
	if state == "" || code == "" {
		return nil, ErrInvalidInput
	}
	s.mu.Lock()
	st, ok := s.states[state]
	if ok {
		delete(s.states, state)
	}
	s.mu.Unlock()
	if !ok || time.Since(st.createdAt) > 10*time.Minute {
		return nil, fmt.Errorf("%w: invalid or expired state", ErrUnauthorized)
	}

	cfg := s.settings.OIDC(ctx)
	if !cfg.Enabled() {
		return nil, ErrNotFound
	}
	d, err := s.discover(ctx, cfg.Issuer)
	if err != nil {
		return nil, err
	}

	// Exchange the code for tokens (confidential client + PKCE verifier).
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", s.redirectURI())
	form.Set("client_id", cfg.ClientID)
	form.Set("client_secret", cfg.ClientSecret)
	form.Set("code_verifier", st.verifier)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.TokenEndpoint,
		strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("oidc token exchange: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: token endpoint status %d", ErrUnauthorized, resp.StatusCode)
	}
	var tok struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return nil, fmt.Errorf("oidc token decode: %w", err)
	}
	if tok.AccessToken == "" {
		return nil, fmt.Errorf("%w: no access token", ErrUnauthorized)
	}

	// Fetch verified claims from the userinfo endpoint.
	if d.UserinfoEndpoint == "" {
		return nil, fmt.Errorf("oidc: provider has no userinfo endpoint")
	}
	ureq, err := http.NewRequestWithContext(ctx, http.MethodGet, d.UserinfoEndpoint, nil)
	if err != nil {
		return nil, err
	}
	ureq.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	ureq.Header.Set("Accept", "application/json")

	uresp, err := s.client.Do(ureq)
	if err != nil {
		return nil, fmt.Errorf("oidc userinfo: %w", err)
	}
	defer uresp.Body.Close()
	if uresp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: userinfo status %d", ErrUnauthorized, uresp.StatusCode)
	}
	var claims struct {
		Sub               string `json:"sub"`
		Email             string `json:"email"`
		PreferredUsername string `json:"preferred_username"`
		Name              string `json:"name"`
	}
	if err := json.NewDecoder(uresp.Body).Decode(&claims); err != nil {
		return nil, fmt.Errorf("oidc userinfo decode: %w", err)
	}
	if claims.Email == "" {
		return nil, fmt.Errorf("%w: provider returned no email claim", ErrUnauthorized)
	}

	return &OIDCUser{
		Subject:  claims.Sub,
		Email:    strings.ToLower(claims.Email),
		Username: claims.PreferredUsername,
		Name:     claims.Name,
	}, nil
}

func (s *oidcService) gcStatesLocked() {
	for k, v := range s.states {
		if time.Since(v.createdAt) > 10*time.Minute {
			delete(s.states, k)
		}
	}
}

func randToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func pkceChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
