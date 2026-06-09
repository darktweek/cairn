package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/darktweek/cairn/internal/middleware"
	"github.com/darktweek/cairn/internal/repository"
	"github.com/darktweek/cairn/internal/service"
)

func (h *Handler) ListBookmarks(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	q := r.URL.Query()
	offset, limit := pageParams(r)

	filter := repository.BookmarkFilter{
		Search: q.Get("search"),
		Offset: offset,
		Limit:  limit,
	}
	if folder := q.Get("folder"); folder != "" {
		filter.Folder = &folder
	}
	if tagID := q.Get("tag_id"); tagID != "" {
		filter.TagID = &tagID
	}

	bookmarks, total, err := h.Bookmark.List(r.Context(), user.ID, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"total":     total,
		"bookmarks": bookmarks,
	})
}

func (h *Handler) CreateBookmark(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	var body struct {
		URL    string   `json:"url"`
		Title  string   `json:"title"`
		Folder *string  `json:"folder"`
		Tags   []string `json:"tags"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}

	b, err := h.Bookmark.Create(r.Context(), user.ID, body.URL, body.Title, body.Folder, body.Tags)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

func (h *Handler) GetBookmark(w http.ResponseWriter, r *http.Request) {
	_ = middleware.UserFromCtx(r.Context())
	// GetByID is not on BookmarkService — use List with filter for now.
	// For MVP we expose the list endpoint; direct get is not critical for the frontend.
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *Handler) UpdateBookmark(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	var body struct {
		URL    string   `json:"url"`
		Title  string   `json:"title"`
		Folder *string  `json:"folder"`
		Tags   []string `json:"tags"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}

	if err := h.Bookmark.Update(r.Context(), user.ID, id, body.URL, body.Title, body.Folder, body.Tags); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteBookmark(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	id := chi.URLParam(r, "id")

	if err := h.Bookmark.Delete(r.Context(), user.ID, id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateBookmarkSort(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}

	if err := h.Bookmark.UpdateSort(r.Context(), user.ID, body.IDs); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ImportBookmarks(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, fmt.Errorf("%w: %s", service.ErrInvalidInput, err.Error()))
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, fmt.Errorf("%w: missing file", service.ErrInvalidInput))
		return
	}
	defer file.Close()

	buf := make([]byte, 10<<20)
	n, _ := file.Read(buf)

	imported, skipped, err := h.Bookmark.ImportNetscape(r.Context(), user.ID, buf[:n])
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]int{
		"imported": imported,
		"skipped":  skipped,
	})
}

func (h *Handler) ExportBookmarks(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())

	data, err := h.Bookmark.ExportNetscape(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.Header().Set("Content-Disposition", `attachment; filename="bookmarks.html"`)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// QuickBookmark is used by the bookmarklet — auth handled by BookmarkletAuth middleware.
func (h *Handler) QuickBookmark(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	var body struct {
		URL   string `json:"url"`
		Title string `json:"title"`
		// Token is already consumed by middleware, but still present in the re-injected body.
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}

	b, err := h.Bookmark.Create(r.Context(), user.ID, body.URL, body.Title, nil, nil)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": b.ID})
}
