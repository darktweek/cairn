package service

import (
	"context"
	"strings"

	"github.com/darktweek/cairn/internal/config"
	"github.com/darktweek/cairn/internal/repository"
)

// SettingsService centralizes runtime-configurable settings. Each setting can
// come from three sources, in priority order:
//
//	env     — hardcoded in the compose / environment, locked (read-only in admin)
//	db      — set by an admin through the panel
//	default — built-in fallback
//
// The "locked" flag tells the frontend whether the admin may edit a value.
type SettingsService interface {
	MenuBang(ctx context.Context) string
	SetMenuBang(ctx context.Context, bang string) error
	MenuBangLocked() bool

	OIDC(ctx context.Context) OIDCConfig
	SetOIDC(ctx context.Context, in OIDCConfig) error
}

// OIDCConfig is the resolved single-sign-on configuration.
type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	ProviderName string
	Scopes       string
	Locked       bool // env-managed, not editable from the admin panel
}

// Enabled reports whether SSO is fully configured.
func (c OIDCConfig) Enabled() bool {
	return c.Issuer != "" && c.ClientID != "" && c.ClientSecret != ""
}

const defaultMenuBang = "!menu"

type settingsService struct {
	repos *repository.Repositories
	cfg   *config.Config
}

func newSettingsService(repos *repository.Repositories, cfg *config.Config) SettingsService {
	return &settingsService{repos: repos, cfg: cfg}
}

func (s *settingsService) MenuBang(ctx context.Context) string {
	if s.cfg.MenuBang != "" {
		return s.cfg.MenuBang // env-locked
	}
	if v, err := s.repos.Invitations.GetSetting(ctx, "menu_bang"); err == nil && v != "" {
		return v
	}
	return defaultMenuBang
}

func (s *settingsService) SetMenuBang(ctx context.Context, bang string) error {
	if s.MenuBangLocked() {
		return ErrForbidden
	}
	if bang == "" || bang[0] != '!' || len(bang) < 2 {
		return ErrInvalidInput
	}
	return s.repos.Invitations.SetSetting(ctx, "menu_bang", bang)
}

func (s *settingsService) MenuBangLocked() bool { return s.cfg.MenuBang != "" }

// OIDC resolves the SSO config. Env (compose) takes priority and locks editing;
// otherwise values come from the admin-managed settings table.
func (s *settingsService) OIDC(ctx context.Context) OIDCConfig {
	if s.cfg.OIDCIssuer != "" {
		return OIDCConfig{
			Issuer:       s.cfg.OIDCIssuer,
			ClientID:     s.cfg.OIDCClientID,
			ClientSecret: s.cfg.OIDCClientSecret,
			ProviderName: orDefault(s.cfg.OIDCProviderName, "SSO"),
			Scopes:       orDefault(s.cfg.OIDCScopes, "openid profile email"),
			Locked:       true,
		}
	}
	get := func(k string) string {
		v, _ := s.repos.Invitations.GetSetting(ctx, k)
		return v
	}
	return OIDCConfig{
		Issuer:       get("oidc_issuer"),
		ClientID:     get("oidc_client_id"),
		ClientSecret: get("oidc_client_secret"),
		ProviderName: orDefault(get("oidc_provider_name"), "SSO"),
		Scopes:       orDefault(get("oidc_scopes"), "openid profile email"),
		Locked:       false,
	}
}

func (s *settingsService) SetOIDC(ctx context.Context, in OIDCConfig) error {
	if s.cfg.OIDCIssuer != "" {
		return ErrForbidden // env-locked
	}
	set := func(k, v string) error { return s.repos.Invitations.SetSetting(ctx, k, v) }
	if err := set("oidc_issuer", strings.TrimSpace(in.Issuer)); err != nil {
		return err
	}
	if err := set("oidc_client_id", strings.TrimSpace(in.ClientID)); err != nil {
		return err
	}
	// Only overwrite the secret when a new non-empty value is provided.
	if strings.TrimSpace(in.ClientSecret) != "" {
		if err := set("oidc_client_secret", strings.TrimSpace(in.ClientSecret)); err != nil {
			return err
		}
	}
	if err := set("oidc_provider_name", strings.TrimSpace(in.ProviderName)); err != nil {
		return err
	}
	if in.Scopes != "" {
		if err := set("oidc_scopes", strings.TrimSpace(in.Scopes)); err != nil {
			return err
		}
	}
	return nil
}

func orDefault(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}
