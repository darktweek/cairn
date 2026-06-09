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
		"wallpaper_limit":   user.WallpaperLimit,
		"created_at":        user.CreatedAt.Unix(),
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
	sess := middleware.SessionFromCtx(r.Context())

	// Delegate to auth — no ListSessions on auth service, so we'd need to add it.
	// For now, return current session only.
	writeJSON(w, http.StatusOK, []map[string]any{
		{
			"id":             sess.ID,
			"user_agent":     sess.UserAgent,
			"ip":             sess.IP,
			"expires_at":     sess.ExpiresAt.Unix(),
			"created_at":     sess.CreatedAt.Unix(),
			"is_bookmarklet": sess.IsBookmarklet,
			"current":        true,
		},
	})
	_ = user
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
	clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetMyAuditLog(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	offset, limit := pageParams(r)

	entries, total, err := h.User.GetAuditLog(r.Context(), user.ID, offset, limit)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"total":   total,
		"entries": entries,
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
		"secret":   secret,
		"qr_url":   qrURL,
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
