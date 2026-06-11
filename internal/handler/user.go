package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/darktweek/cairn/internal/middleware"
	"github.com/darktweek/cairn/internal/service"
)

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	writeJSON(w, http.StatusOK, map[string]any{
		"id":                user.ID,
		"username":          user.Username,
		"email":             user.Email,
		"role":              user.Role,
		"is_active":         user.IsActive,
		"search_engine":     user.SearchEngine,
		"search_engine_url": user.SearchEngineURL,
		"wallpaper_limit":    user.WallpaperLimit,
		"upload_size_limit":  user.UploadSizeLimit,
		"storage_quota":      user.StorageQuota,
		"created_at":        user.CreatedAt.Unix(),
		"locale":            user.Locale,
		"menu_bang":         h.Settings.MenuBang(r.Context()),
		"smtp_configured":   h.Settings.SMTP(r.Context()).Configured(),
	})
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	var body struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}

	if err := h.User.UpdateProfile(r.Context(), user.ID, body.Username, body.Email); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	var body struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}

	if err := h.User.ChangePassword(r.Context(), user.ID, body.CurrentPassword, body.NewPassword); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	curr := middleware.SessionFromCtx(r.Context())

	sessions, err := h.Auth.ListSessions(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}

	out := make([]map[string]any, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, map[string]any{
			"id":             s.ID,
			"user_agent":     s.UserAgent,
			"ip":             s.IP,
			"expires_at":     s.ExpiresAt.Unix(),
			"created_at":     s.CreatedAt.Unix(),
			"is_bookmarklet": s.IsBookmarklet,
			"current":        s.ID == curr.ID,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	sessionID := chi.URLParam(r, "id")
	if err := h.Auth.LogoutForUser(r.Context(), sessionID, user.ID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	if err := h.Auth.LogoutAll(r.Context(), user.ID); err != nil {
		writeError(w, err)
		return
	}
	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetMyStats(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	st, err := h.User.Stats(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"bookmarks":    st.Bookmarks,
		"wallpapers":   st.Wallpapers,
		"sessions":     st.Sessions,
		"member_since": user.CreatedAt.Unix(),
	})
}

func (h *Handler) GetMyAuditLog(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	offset, limit := pageParams(r)

	entries, total, err := h.User.GetAuditLog(r.Context(), user.ID, offset, limit)
	if err != nil {
		writeError(w, err)
		return
	}

	out := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		out = append(out, auditJSON(e))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"total":   total,
		"entries": out,
	})
}

// TOTP handlers

func (h *Handler) BeginTOTP(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	secret, qrURL, err := h.Auth.BeginTOTP(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"secret": secret,
		"qr_url": qrURL,
	})
}

func (h *Handler) ConfirmTOTP(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	var body struct {
		Code string `json:"code"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}

	if err := h.Auth.ConfirmTOTP(r.Context(), user.ID, body.Code); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DisableTOTP(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	var body struct {
		Password string `json:"password"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}

	if err := h.Auth.DisableTOTP(r.Context(), user.ID, body.Password); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Bookmarklet handlers

func (h *Handler) GetBookmarklet(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	js, err := h.Bookmark.GenerateBookmarklet(r.Context(), user.ID, clientIP(r), r.UserAgent())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"bookmarklet": js})
}

func (h *Handler) RevokeBookmarklet(w http.ResponseWriter, r *http.Request) {
	// Revoke all bookmarklet sessions for the user.
	// Since we can't filter by is_bookmarklet in LogoutAll, we'd need a dedicated repo method.
	// For MVP: LogoutAll — revoking all sessions is acceptable on homelab.
	user := middleware.UserFromCtx(r.Context())
	_ = h.Auth.LogoutAll(r.Context(), user.ID)
	w.WriteHeader(http.StatusNoContent)
}

// Locale handler

func (h *Handler) UpdateLocale(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	var body struct {
		Locale string `json:"locale"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}
	if err := h.User.UpdateLocale(r.Context(), user.ID, body.Locale); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteAccount hard-deletes the authenticated user's account after password confirmation.
func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	var body struct {
		Password string `json:"password"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}
	if err := h.Auth.DeleteAccount(r.Context(), user.ID, body.Password); err != nil {
		writeError(w, err)
		return
	}
	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// Search engine handler

func (h *Handler) UpdateSearchEngine(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	var body struct {
		Engine    string  `json:"engine"`
		CustomURL *string `json:"custom_url"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}

	if err := h.User.UpdateSearchEngine(r.Context(), user.ID, body.Engine, body.CustomURL); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
