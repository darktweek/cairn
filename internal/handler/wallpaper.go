package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/darktweek/cairn/internal/middleware"
	"github.com/darktweek/cairn/internal/service"
)

func (h *Handler) ListWallpapers(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	wallpapers, err := h.Wallpaper.List(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, wallpapers)
}

func (h *Handler) UploadWallpaper(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())

	if err := r.ParseMultipartForm(52 << 20); err != nil {
		writeError(w, fmt.Errorf("%w: %s", service.ErrInvalidInput, err.Error()))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, fmt.Errorf("%w: missing file", service.ErrInvalidInput))
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, fmt.Errorf("read file: %w", err))
		return
	}

	wp, err := h.Wallpaper.Upload(r.Context(), user.ID, header.Filename, data)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, wp)
}

func (h *Handler) DeleteWallpaper(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	if err := h.Wallpaper.Delete(r.Context(), user.ID, id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) SetWallpaperPinned(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	var body struct {
		Pinned bool `json:"pinned"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}

	if err := h.Wallpaper.SetPinned(r.Context(), user.ID, id, body.Pinned); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateWallpaperSort(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}

	if err := h.Wallpaper.UpdateSort(r.Context(), user.ID, body.IDs); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ServeMedia(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	filename := chi.URLParam(r, "filename")

	// Basic path traversal protection — chi ensures no slashes in params.
	if userID == "" || filename == "" {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, fmt.Sprintf("%s/%s/%s", mediaBasePath(r), userID, filename))
}

// mediaBasePath reads the media path from the request context (injected by main).
func mediaBasePath(r *http.Request) string {
	if v, ok := r.Context().Value(ctxKeyMediaPath).(string); ok {
		return v
	}
	return "/data/media"
}

type ctxKey int

const ctxKeyMediaPath ctxKey = iota
