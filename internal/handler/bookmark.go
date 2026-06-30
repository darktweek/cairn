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

func (h *Handler) ListBookmarks(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	q := r.URL.Query()
	offset, limit := pageParams(r)

	filter := repository.BookmarkFilter{
		Search:        q.Get("search"),
		IncludeHidden: q.Get("hidden") == "1",
		Offset:        offset,
		Limit:         limit,
	}
	if folderID := q.Get("folder_id"); folderID != "" {
		filter.FolderID = &folderID
	}
	if tagID := q.Get("tag_id"); tagID != "" {
		filter.TagID = &tagID
	}

	var (
		bookmarks []*model.Bookmark
		total     int
		err       error
	)
	// When a collection is given, list that collection (access-checked); otherwise
	// list the bookmarks the user authored (start-page view).
	if collectionID := q.Get("collection_id"); collectionID != "" {
		bookmarks, total, err = h.Bookmark.ListInCollection(r.Context(), user.ID, collectionID, filter)
	} else {
		bookmarks, total, err = h.Bookmark.List(r.Context(), user.ID, filter)
	}
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
		URL          string   `json:"url"`
		Title        string   `json:"title"`
		Hidden       bool     `json:"hidden"`
		CollectionID string   `json:"collection_id"`
		FolderID     *string  `json:"folder_id"`
		Tags         []string `json:"tags"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}

	b, err := h.Bookmark.Create(r.Context(), user.ID, service.BookmarkInput{
		URL:          body.URL,
		Title:        body.Title,
		Hidden:       body.Hidden,
		CollectionID: body.CollectionID,
		FolderID:     body.FolderID,
		Tags:         body.Tags,
	})
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
		URL          string   `json:"url"`
		Title        string   `json:"title"`
		Hidden       bool     `json:"hidden"`
		CollectionID string   `json:"collection_id"`
		FolderID     *string  `json:"folder_id"`
		Tags         []string `json:"tags"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}

	if err := h.Bookmark.Update(r.Context(), user.ID, id, service.BookmarkInput{
		URL:          body.URL,
		Title:        body.Title,
		Hidden:       body.Hidden,
		CollectionID: body.CollectionID,
		FolderID:     body.FolderID,
		Tags:         body.Tags,
	}); err != nil {
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

func (h *Handler) ClearBookmarks(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	if err := h.Bookmark.ClearAll(r.Context(), user.ID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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

	b, err := h.Bookmark.Create(r.Context(), user.ID, service.BookmarkInput{
		URL:   body.URL,
		Title: body.Title,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": b.ID})
}
