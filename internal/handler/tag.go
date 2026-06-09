package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/darktweek/cairn/internal/middleware"
)

func (h *Handler) ListTags(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	tags, err := h.Bookmark.ListTags(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tags)
}

func (h *Handler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	if err := h.Bookmark.DeleteTag(r.Context(), user.ID, id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
