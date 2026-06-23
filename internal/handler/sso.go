package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/darktweek/cairn/internal/service"
)

// SSOConfig — public: GET /api/auth/sso/config
// Tells the login page whether to show a "Sign in with <provider>" button.
func (h *Handler) SSOConfig(w http.ResponseWriter, r *http.Request) {
	cfg := h.OIDC.Config(r.Context())
	writeJSON(w, http.StatusOK, map[string]any{
		"enabled":       cfg.Enabled(),
		"provider_name": cfg.ProviderName,
	})
}

// SSOLogin — public: GET /api/auth/sso/login → redirect to the provider.
func (h *Handler) SSOLogin(w http.ResponseWriter, r *http.Request) {
	authURL, _, err := h.OIDC.AuthURL(r.Context())
	if err != nil {
		slog.Error("sso: build auth URL", "err", err)
		http.Redirect(w, r, "/?sso_error=unavailable", http.StatusFound)
		return
	}
	http.Redirect(w, r, authURL, http.StatusFound)
}

// SSOCallback — public: GET /api/auth/sso/callback?code=&state=
func (h *Handler) SSOCallback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if e := q.Get("error"); e != "" {
		slog.Warn("sso: provider returned error", "error", e, "description", q.Get("error_description"))
		http.Redirect(w, r, "/?sso_error="+e, http.StatusFound)
		return
	}
	state := q.Get("state")
	code := q.Get("code")

	claims, err := h.OIDC.Exchange(r.Context(), state, code)
	if err != nil {
		slog.Error("sso: exchange failed", "err", err)
		http.Redirect(w, r, "/?sso_error=exchange", http.StatusFound)
		return
	}

	user, err := h.User.ProvisionSSO(r.Context(), claims.Email, claims.Username, claims.Name, clientIP(r), r.UserAgent())
	if err != nil {
		slog.Error("sso: provision failed", "email", claims.Email, "err", err)
		http.Redirect(w, r, "/?sso_error=provision", http.StatusFound)
		return
	}

	_, token, err := h.Auth.CreateSessionForUser(r.Context(), user.ID, clientIP(r), r.UserAgent())
	if err != nil {
		slog.Error("sso: create session failed", "userID", user.ID, "err", err)
		http.Redirect(w, r, "/?sso_error=session", http.StatusFound)
		return
	}

	h.setSessionCookie(w, token, sessionMaxAge)
	http.Redirect(w, r, "/", http.StatusFound)
}

// AdminGetSSOSettings — GET /api/admin/settings/sso (secret never returned)
func (h *Handler) AdminGetSSOSettings(w http.ResponseWriter, r *http.Request) {
	cfg := h.OIDC.Config(r.Context())
	writeJSON(w, http.StatusOK, map[string]any{
		"issuer":        cfg.Issuer,
		"client_id":     cfg.ClientID,
		"provider_name": cfg.ProviderName,
		"scopes":        cfg.Scopes,
		"has_secret":    cfg.ClientSecret != "",
		"locked":        cfg.Locked,
		"enabled":       cfg.Enabled(),
	})
}

// AdminSetSSOSettings — PUT /api/admin/settings/sso
func (h *Handler) AdminSetSSOSettings(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Issuer       string `json:"issuer"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		ProviderName string `json:"provider_name"`
		Scopes       string `json:"scopes"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}
	err := h.Settings.SetOIDC(r.Context(), service.OIDCConfig{
		Issuer:       body.Issuer,
		ClientID:     body.ClientID,
		ClientSecret: body.ClientSecret,
		ProviderName: body.ProviderName,
		Scopes:       body.Scopes,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	h.AdminGetSSOSettings(w, r)
}
