package handler

import (
	"fmt"
	"net/http"

	"github.com/darktweek/cairn/internal/service"
)

// AdminGetSystemSettings — GET /api/admin/settings/system
// Returns every configuration value: editable runtime settings, plus
// compose-managed values shown read-only (secrets masked).
func (h *Handler) AdminGetSystemSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	totp := h.Settings.TOTPIssuer(ctx)
	wp := h.Settings.WallpaperLimit(ctx)
	bm := h.Settings.BookmarkletDays(ctx)
	sys := h.Settings.SystemInfo()
	smtp := h.Settings.SMTP(ctx)

	writeJSON(w, http.StatusOK, map[string]any{
		"editable": map[string]any{
			"totp_issuer":      map[string]any{"value": totp.Value, "locked": totp.Locked},
			"wallpaper_limit":  map[string]any{"value": wp.Value, "locked": wp.Locked},
			"bookmarklet_days": map[string]any{"value": bm.Value, "locked": bm.Locked},
		},
		// SMTP — editable unless env-managed; password is write-only.
		"smtp": map[string]any{
			"host":         smtp.Host,
			"port":         smtp.Port,
			"user":         smtp.User,
			"from":         smtp.From,
			"tls":          smtp.TLS,
			"has_password": smtp.Pass != "",
			"locked":       smtp.Locked,
			"configured":   smtp.Configured(),
		},
		// Compose-managed — read-only, restart required to change.
		"system": map[string]any{
			"addr":            sys.Addr,
			"env":             sys.Env,
			"base_url":        sys.BaseURL,
			"db_path":         sys.DBPath,
			"media_path":      sys.MediaPath,
			"max_upload_size":         sys.MaxUploadSize,
				"default_storage_quota":   sys.DefaultStorageQuota,
			"trusted_proxy":   sys.TrustedProxy,
			"session_secret":  sys.SessionSecretSet, // bool only, never the value
		},
	})
}

// AdminSetSystemSettings — PUT /api/admin/settings/system
// Updates only the admin-editable runtime settings; env-locked values ignored.
func (h *Handler) AdminSetSystemSettings(w http.ResponseWriter, r *http.Request) {
	var body struct {
		TOTPIssuer      string `json:"totp_issuer"`
		WallpaperLimit  int    `json:"wallpaper_limit"`
		BookmarkletDays int    `json:"bookmarklet_days"`
		SMTP            *struct {
			Host string `json:"host"`
			Port int    `json:"port"`
			User string `json:"user"`
			Pass string `json:"pass"`
			From string `json:"from"`
			TLS  bool   `json:"tls"`
		} `json:"smtp"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}
	if err := h.Settings.SetRuntime(r.Context(), body.TOTPIssuer, body.WallpaperLimit, body.BookmarkletDays); err != nil {
		writeError(w, err)
		return
	}
	if body.SMTP != nil {
		err := h.Settings.SetSMTP(r.Context(), service.SMTPSettings{
			Host: body.SMTP.Host,
			Port: body.SMTP.Port,
			User: body.SMTP.User,
			Pass: body.SMTP.Pass,
			From: body.SMTP.From,
			TLS:  body.SMTP.TLS,
		})
		if err != nil {
			writeError(w, err)
			return
		}
	}
	h.AdminGetSystemSettings(w, r)
}
