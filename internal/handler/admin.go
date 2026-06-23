package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/darktweek/cairn/internal/middleware"
	"github.com/darktweek/cairn/internal/model"
	"github.com/darktweek/cairn/internal/repository"
	"github.com/darktweek/cairn/internal/service"
)

// userJSON serializes a user for admin views. Never includes the password hash.
func userJSON(u *model.User) map[string]any {
	out := map[string]any{
		"id":            u.ID,
		"username":      u.Username,
		"email":         u.Email,
		"role":          u.Role,
		"is_active":     u.IsActive,
		"search_engine": u.SearchEngine,
		"created_at":    u.CreatedAt.Unix(),
		"updated_at":    u.UpdatedAt.Unix(),
	}
	if u.WallpaperLimit != nil {
		out["wallpaper_limit"] = *u.WallpaperLimit
	}
	if u.UploadSizeLimit != nil {
		out["upload_size_limit"] = *u.UploadSizeLimit
	}
	if u.StorageQuota != nil {
		out["storage_quota"] = *u.StorageQuota
	}
	return out
}

// auditJSON serializes an audit entry with a unix timestamp.
func auditJSON(e *model.AuditEntry) map[string]any {
	out := map[string]any{
		"id":         e.ID,
		"action":     e.Action,
		"ip":         e.IP,
		"user_agent": e.UserAgent,
		"created_at": e.CreatedAt.Unix(),
	}
	if e.UserID != nil {
		out["user_id"] = *e.UserID
	}
	if e.Metadata != nil {
		out["metadata"] = e.Metadata
	}
	return out
}

func (h *Handler) AdminListUsers(w http.ResponseWriter, r *http.Request) {
	offset, limit := pageParams(r)
	users, total, err := h.Admin.ListUsers(r.Context(), offset, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]map[string]any, 0, len(users))
	for _, u := range users {
		out = append(out, userJSON(u))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"total": total,
		"users": out,
	})
}

func (h *Handler) AdminGetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := h.Admin.GetUser(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, userJSON(user))
}

func (h *Handler) AdminSuspendUser(w http.ResponseWriter, r *http.Request) {
	admin := middleware.UserFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	if err := h.Admin.SuspendUser(r.Context(), admin.ID, id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AdminActivateUser(w http.ResponseWriter, r *http.Request) {
	admin := middleware.UserFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	if err := h.Admin.ActivateUser(r.Context(), admin.ID, id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	admin := middleware.UserFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	if err := h.Admin.DeleteUser(r.Context(), admin.ID, id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AdminSetWallpaperLimit(w http.ResponseWriter, r *http.Request) {
	admin := middleware.UserFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	var body struct {
		Limit *int `json:"limit"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}

	if err := h.Admin.SetWallpaperLimit(r.Context(), admin.ID, id, body.Limit); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AdminSetUploadSizeLimit(w http.ResponseWriter, r *http.Request) {
	admin := middleware.UserFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	var body struct {
		Limit *int64 `json:"limit"` // bytes; null to reset to global default
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}

	if err := h.Admin.SetUploadSizeLimit(r.Context(), admin.ID, id, body.Limit); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AdminSetStorageQuota(w http.ResponseWriter, r *http.Request) {
	admin := middleware.UserFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	var body struct {
		Quota *int64 `json:"quota"` // bytes; null to reset to global default
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}

	if err := h.Admin.SetStorageQuota(r.Context(), admin.ID, id, body.Quota); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AdminListPendingRegistrations — GET /api/admin/pending-registrations
// Open-registration requests awaiting email confirmation. Token hashes and
// TOTP secrets are never exposed.
func (h *Handler) AdminListPendingRegistrations(w http.ResponseWriter, r *http.Request) {
	prs, err := h.Admin.ListPendingRegistrations(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]map[string]any, 0, len(prs))
	for _, pr := range prs {
		out = append(out, map[string]any{
			"id":         pr.ID,
			"username":   pr.Username,
			"email":      pr.Email,
			"expires_at": pr.ExpiresAt.Unix(),
			"created_at": pr.CreatedAt.Unix(),
			"completed":  pr.IsCompleted(),
			"expired":    pr.IsExpired(),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// AdminRevokePendingRegistration — DELETE /api/admin/pending-registrations/{id}
func (h *Handler) AdminRevokePendingRegistration(w http.ResponseWriter, r *http.Request) {
	admin := middleware.UserFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	if err := h.Admin.RevokePendingRegistration(r.Context(), admin.ID, id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AdminGetAuditLog(w http.ResponseWriter, r *http.Request) {
	offset, limit := pageParams(r)
	q := r.URL.Query()

	filter := repository.AuditFilter{}
	if v := q.Get("user_id"); v != "" {
		filter.UserID = &v
	}
	if v := q.Get("action"); v != "" {
		filter.Action = &v
	}

	entries, total, err := h.Admin.GetAuditLog(r.Context(), offset, limit, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	// Resolve actor usernames so the journal shows names, not raw IDs —
	// only the IDs present on this page, never a full user listing.
	idSet := map[string]struct{}{}
	for _, e := range entries {
		if e.UserID != nil {
			idSet[*e.UserID] = struct{}{}
		}
	}
	ids := make([]string, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	names, _ := h.Admin.ResolveUsernames(r.Context(), ids)

	out := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		j := auditJSON(e)
		if e.UserID != nil {
			if name, ok := names[*e.UserID]; ok {
				j["username"] = name
			}
		}
		out = append(out, j)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"total":   total,
		"entries": out,
	})
}

// AdminGetUserStats — GET /api/admin/users/{id}/stats
func (h *Handler) AdminGetUserStats(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := h.Admin.GetUser(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	st, err := h.User.Stats(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"bookmarks":     st.Bookmarks,
		"wallpapers":    st.Wallpapers,
		"sessions":      st.Sessions,
		"storage_bytes": st.StorageBytes,
		"member_since":  user.CreatedAt.Unix(),
	})
}

// AdminTestSMTP — POST /api/admin/settings/smtp/test
// Sends a test email to the logged-in admin's address to verify SMTP settings.
func (h *Handler) AdminTestSMTP(w http.ResponseWriter, r *http.Request) {
	admin := middleware.UserFromCtx(r.Context())
	if err := h.Admin.TestSMTP(r.Context(), admin.ID, admin.Email); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "to": admin.Email})
}

func (h *Handler) AdminGetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.Admin.GetStats(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"total_users":           stats.TotalUsers,
		"active_users":          stats.ActiveUsers,
		"total_bookmarks":       stats.TotalBookmarks,
		"total_wallpapers":      stats.TotalWallpapers,
		"db_size_bytes":         stats.DBSizeBytes,
		"media_bytes":           stats.MediaBytes,
		"active_sessions":       stats.ActiveSessions,
		"pending_invitations":   stats.PendingInvitations,
		"pending_registrations": stats.PendingRegistrations,
		"audit_entries":         stats.AuditEntries,
	})
}
