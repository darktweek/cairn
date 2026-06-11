package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	// Server
	Addr    string
	Env     string
	BaseURL string

	// Database
	DBPath    string
	MediaPath string

	// Security
	SessionSecret string

	// Limits
	DefaultWallpaperLimit    int
	MaxUploadSize            int64  // max size of a single uploaded file (default 50 MB)
	DefaultStorageQuota      int64  // max total media storage per user (default 200 MB)
	BookmarkletTokenLifetime int

	// TOTP
	TOTPIssuer string

	// Invitations
	OpenRegistration bool
	InviteLifetime   int // hours

	// Menu — bang that opens the full-page hub. Empty = admin-editable (default !menu).
	MenuBang string

	// OIDC / SSO — when issuer+client are set here, they are locked (env-managed).
	// Otherwise an admin may configure them from the panel.
	OIDCIssuer       string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCProviderName string
	OIDCScopes       string

	// Proxy
	TrustedProxy bool

	// SMTP
	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
	SMTPFrom string
	SMTPTLS  bool
}

func Load() (*Config, error) {
	cfg := &Config{
		Addr:                     getEnv("CAIRN_ADDR", ":8080"),
		Env:                      getEnv("CAIRN_ENV", "production"),
		BaseURL:                  getEnv("CAIRN_BASE_URL", ""),
		DBPath:                   getEnv("CAIRN_DB_PATH", "/data/db.sqlite"),
		MediaPath:                getEnv("CAIRN_MEDIA_PATH", "/data/media"),
		SessionSecret:            getEnv("CAIRN_SESSION_SECRET", ""),
		DefaultWallpaperLimit:    getEnvInt("CAIRN_DEFAULT_WALLPAPER_LIMIT", 10),
		MaxUploadSize:            getEnvInt64("CAIRN_MAX_UPLOAD_SIZE", 52428800),   // 50 MB
		DefaultStorageQuota:      getEnvInt64("CAIRN_STORAGE_QUOTA", 209715200),   // 200 MB
		BookmarkletTokenLifetime: getEnvInt("CAIRN_BOOKMARKLET_TOKEN_LIFETIME", 90),
		TOTPIssuer:               getEnv("CAIRN_TOTP_ISSUER", "Cairn"),
		OpenRegistration:         getEnvBool("CAIRN_OPEN_REGISTRATION", true),
		InviteLifetime:           getEnvInt("CAIRN_INVITE_LIFETIME", 72),
		MenuBang:                 getEnv("CAIRN_MENU_BANG", ""),
		OIDCIssuer:               getEnv("CAIRN_OIDC_ISSUER", ""),
		OIDCClientID:             getEnv("CAIRN_OIDC_CLIENT_ID", ""),
		OIDCClientSecret:         getEnv("CAIRN_OIDC_CLIENT_SECRET", ""),
		OIDCProviderName:         getEnv("CAIRN_OIDC_PROVIDER_NAME", ""),
		OIDCScopes:               getEnv("CAIRN_OIDC_SCOPES", "openid profile email"),
		TrustedProxy:             getEnvBool("CAIRN_TRUSTED_PROXY", true),
		SMTPHost:                 getEnv("CAIRN_SMTP_HOST", ""),
		SMTPPort:                 getEnvInt("CAIRN_SMTP_PORT", 587),
		SMTPUser:                 getEnv("CAIRN_SMTP_USER", ""),
		SMTPPass:                 getEnv("CAIRN_SMTP_PASS", ""),
		SMTPFrom:                 getEnv("CAIRN_SMTP_FROM", ""),
		SMTPTLS:                  getEnvBool("CAIRN_SMTP_TLS", true),
	}

	return cfg, cfg.validate()
}

func (c *Config) validate() error {
	var errs []error

	if c.BaseURL == "" {
		errs = append(errs, errors.New("CAIRN_BASE_URL is required"))
	}
	if c.SessionSecret == "" {
		errs = append(errs, errors.New("CAIRN_SESSION_SECRET is required"))
	} else if len(c.SessionSecret) < 32 {
		errs = append(errs, errors.New("CAIRN_SESSION_SECRET must be at least 32 characters"))
	}
	if c.Env != "production" && c.Env != "development" {
		errs = append(errs, fmt.Errorf("CAIRN_ENV must be 'production' or 'development', got %q", c.Env))
	}

	// SMTP is intentionally not required here: if it is not provided via the
	// environment, an admin configures it through the web setup, after which it
	// is stored in the database and used at send time.

	return errors.Join(errs...)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvInt64(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}
