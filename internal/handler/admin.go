package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/darktweek/cairn/internal/middleware"
	"github.com/darktweek/cairn/internal/repository"
	"github.com/darktweek/cairn/internal/service"
)

func (h *Handler) AdminListUsers(w http.ResponseWriter, r *http.Request) {
	offset, limit := pageParams(r)
	users, total, err := h.Admin.ListUsers(r.Context(), offset, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"total": total,
		"users": users,
	})
}

func (h *Handler) AdminGetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := h.Admin.GetUser(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
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
	writeJSON(w, http.StatusOK, map[string]any{
		"total":   total,
		"entries": entries,
	})
}

func (h *Handler) AdminGetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.Admin.GetStats(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}
