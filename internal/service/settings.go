package service

import (
	"context"
	"os"
	"strconv"
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

	TOTPIssuer(ctx context.Context) StringSetting
	WallpaperLimit(ctx context.Context) IntSetting
	BookmarkletDays(ctx context.Context) IntSetting
	SetRuntime(ctx context.Context, totpIssuer string, wallpaperLimit, bookmarkletDays int) error

	SystemInfo() SystemInfo
}

// SystemInfo holds compose-managed values shown read-only in the admin panel.
// Secrets are never included as plaintext — only a "set" boolean.
type SystemInfo struct {
	Addr             string
	Env              string
	BaseURL          string
	DBPath           string
	MediaPath        string
	MaxUploadSize    int64
	TrustedProxy     bool
	SessionSecretSet bool
	SMTPHost         string
	SMTPPort         int
	SMTPUser         string
	SMTPFrom         string
	SMTPTLS          bool
	SMTPPassSet      bool
}

// StringSetting / IntSetting carry a resolved value plus whether it is
// env-locked (configured in the compose, read-only in the admin panel).
type StringSetting struct {
	Value  string
	Locked bool
}
type IntSetting struct {
	Value  int
	Locked bool
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

func (s *settingsService) SystemInfo() SystemInfo {
	return SystemInfo{
		Addr:             s.cfg.Addr,
		Env:              s.cfg.Env,
		BaseURL:          s.cfg.BaseURL,
		DBPath:           s.cfg.DBPath,
		MediaPath:        s.cfg.MediaPath,
		MaxUploadSize:    s.cfg.MaxUploadSize,
		TrustedProxy:     s.cfg.TrustedProxy,
		SessionSecretSet: s.cfg.SessionSecret != "",
		SMTPHost:         s.cfg.SMTPHost,
		SMTPPort:         s.cfg.SMTPPort,
		SMTPUser:         s.cfg.SMTPUser,
		SMTPFrom:         s.cfg.SMTPFrom,
		SMTPTLS:          s.cfg.SMTPTLS,
		SMTPPassSet:      s.cfg.SMTPPass != "",
	}
}

// ── Runtime settings (env > DB > default) ──────────────────────────────────

func (s *settingsService) TOTPIssuer(ctx context.Context) StringSetting {
	if _, ok := os.LookupEnv("CAIRN_TOTP_ISSUER"); ok {
		return StringSetting{Value: s.cfg.TOTPIssuer, Locked: true}
	}
	if v, err := s.repos.Invitations.GetSetting(ctx, "totp_issuer"); err == nil && v != "" {
		return StringSetting{Value: v}
	}
	return StringSetting{Value: orDefault(s.cfg.TOTPIssuer, "Cairn")}
}

func (s *settingsService) WallpaperLimit(ctx context.Context) IntSetting {
	if _, ok := os.LookupEnv("CAIRN_DEFAULT_WALLPAPER_LIMIT"); ok {
		return IntSetting{Value: s.cfg.DefaultWallpaperLimit, Locked: true}
	}
	if v, err := s.repos.Invitations.GetSetting(ctx, "wallpaper_limit"); err == nil && v != "" {
		if n, e := strconv.Atoi(v); e == nil {
			return IntSetting{Value: n}
		}
	}
	return IntSetting{Value: s.cfg.DefaultWallpaperLimit}
}

func (s *settingsService) BookmarkletDays(ctx context.Context) IntSetting {
	if _, ok := os.LookupEnv("CAIRN_BOOKMARKLET_TOKEN_LIFETIME"); ok {
		return IntSetting{Value: s.cfg.BookmarkletTokenLifetime, Locked: true}
	}
	if v, err := s.repos.Invitations.GetSetting(ctx, "bookmarklet_days"); err == nil && v != "" {
		if n, e := strconv.Atoi(v); e == nil {
			return IntSetting{Value: n}
		}
	}
	return IntSetting{Value: s.cfg.BookmarkletTokenLifetime}
}

// SetRuntime persists the admin-editable runtime settings. Env-locked values
// are ignored (cannot be overridden from the panel).
func (s *settingsService) SetRuntime(ctx context.Context, totpIssuer string, wallpaperLimit, bookmarkletDays int) error {
	if !s.TOTPIssuer(ctx).Locked && strings.TrimSpace(totpIssuer) != "" {
		if err := s.repos.Invitations.SetSetting(ctx, "totp_issuer", strings.TrimSpace(totpIssuer)); err != nil {
			return err
		}
	}
	if !s.WallpaperLimit(ctx).Locked && wallpaperLimit > 0 {
		if err := s.repos.Invitations.SetSetting(ctx, "wallpaper_limit", strconv.Itoa(wallpaperLimit)); err != nil {
			return err
		}
	}
	if !s.BookmarkletDays(ctx).Locked && bookmarkletDays > 0 {
		if err := s.repos.Invitations.SetSetting(ctx, "bookmarklet_days", strconv.Itoa(bookmarkletDays)); err != nil {
			return err
		}
	}
	return nil
}
