package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/darktweek/cairn/internal/middleware"
	"github.com/darktweek/cairn/internal/service"
)

func (h *Handler) ListCollections(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	collections, err := h.Collection.List(r.Context(), user.ID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"collections": collections})
}

func (h *Handler) GetCollection(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	id := chi.URLParam(r, "id")
	c, err := h.Collection.Get(r.Context(), user.ID, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *Handler) CreateCollection(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
		Icon        string `json:"icon"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}
	c, err := h.Collection.Create(r.Context(), user.ID, service.CollectionInput{
		Name:        body.Name,
		Description: body.Description,
		Color:       body.Color,
		Icon:        body.Icon,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (h *Handler) UpdateCollection(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	id := chi.URLParam(r, "id")
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
		Icon        string `json:"icon"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}
	if err := h.Collection.Update(r.Context(), user.ID, id, service.CollectionInput{
		Name:        body.Name,
		Description: body.Description,
		Color:       body.Color,
		Icon:        body.Icon,
	}); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteCollection(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.Collection.Delete(r.Context(), user.ID, id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListCollectionFolders(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	id := chi.URLParam(r, "id")
	folders, err := h.Collection.ListFolders(r.Context(), user.ID, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"folders": folders})
}

func (h *Handler) ListCollectionShares(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	id := chi.URLParam(r, "id")
	shares, err := h.Collection.ListShares(r.Context(), user.ID, id)
	if err != nil {
		writeError(w, err)
		return
	}
	groupShares, err := h.Collection.ListGroupShares(r.Context(), user.ID, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"shares": shares, "group_shares": groupShares})
}

func (h *Handler) SetCollectionShare(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	id := chi.URLParam(r, "id")
	var body struct {
		UserID string `json:"user_id"`
		Perm   string `json:"perm"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}
	if err := h.Collection.SetShare(r.Context(), user.ID, id, body.UserID, body.Perm); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemoveCollectionShare(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	id := chi.URLParam(r, "id")
	targetID := chi.URLParam(r, "userId")
	if err := h.Collection.RemoveShare(r.Context(), user.ID, id, targetID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.Collection.SearchUsers(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]map[string]string, 0, len(users))
	for _, u := range users {
		out = append(out, map[string]string{"id": u.ID, "username": u.Username})
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": out})
}

func (h *Handler) SetCollectionPublicLink(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	id := chi.URLParam(r, "id")
	var body struct {
		Enable bool `json:"enable"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}
	token, err := h.Collection.SetPublicLink(r.Context(), user.ID, id, body.Enable)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

// PublicCollectionView is unauthenticated — it serves a read-only collection by token.
func (h *Handler) PublicCollectionView(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	view, err := h.Collection.PublicView(r.Context(), token)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (h *Handler) GetPolicies(w http.ResponseWriter, r *http.Request) {
	pol, err := h.Collection.Policies(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"policies": pol})
}

func (h *Handler) SetPolicy(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	var body struct {
		Key   string `json:"key"`
		Value bool   `json:"value"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}
	if err := h.Collection.SetPolicy(r.Context(), user.ID, body.Key, body.Value); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CreateCollectionFolder(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromCtx(r.Context())
	id := chi.URLParam(r, "id")
	var body struct {
		Name     string  `json:"name"`
		ParentID *string `json:"parent_id"`
		Sort     int     `json:"sort"`
	}
	if err := decode(r, &body); err != nil {
		writeError(w, fmt.Errorf("%w: invalid JSON", service.ErrInvalidInput))
		return
	}
	f, err := h.Collection.CreateFolder(r.Context(), user.ID, id, service.FolderInput{
		Name:     body.Name,
		ParentID: body.ParentID,
		Sort:     body.Sort,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, f)
}
