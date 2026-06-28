package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/darktweek/cairn/internal/service"
)

// Handler holds all services and is the base for all HTTP handlers.
type Handler struct {
	Auth         service.AuthService
	User         service.UserService
	Bookmark     service.BookmarkService
	Collection   service.CollectionService
	Group        service.GroupService
	RBAC         service.RBACService
	Wallpaper    service.WallpaperService
	Admin        service.AdminService
	Email        service.EmailService
	Invitation   service.InvitationService
	Settings      service.SettingsService
	OIDC          service.OIDCService
	secureCookie  bool
	sessionMaxAge int
}

func New(svcs *service.Services, production bool, sessionDays int) *Handler {
	maxAge := defaultSessionMaxAge
	if sessionDays > 0 {
		maxAge = sessionDays * 24 * 60 * 60
	}
	return &Handler{
		Auth:          svcs.Auth,
		User:          svcs.User,
		Bookmark:      svcs.Bookmark,
		Collection:    svcs.Collection,
		Group:         svcs.Group,
		RBAC:          svcs.RBAC,
		Wallpaper:     svcs.Wallpaper,
		Admin:         svcs.Admin,
		Email:         svcs.Email,
		Invitation:    svcs.Invitation,
		Settings:      svcs.Settings,
		OIDC:          svcs.OIDC,
		secureCookie:  production,
		sessionMaxAge: maxAge,
	}
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError writes a structured JSON error response.
func writeError(w http.ResponseWriter, err error) {
	status, code := mapError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
		"code":  code,
	})
}

func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return http.StatusBadRequest, "INVALID_INPUT"
	case errors.Is(err, service.ErrUnauthorized):
		return http.StatusUnauthorized, "UNAUTHORIZED"
	case errors.Is(err, service.ErrTOTPRequired):
		return http.StatusUnauthorized, "TOTP_REQUIRED"
	case errors.Is(err, service.ErrRateLimited):
		return http.StatusTooManyRequests, "RATE_LIMITED"
	case errors.Is(err, service.ErrForbidden):
		return http.StatusForbidden, "FORBIDDEN"
	case errors.Is(err, service.ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND"
	case errors.Is(err, service.ErrConflict):
		return http.StatusConflict, "CONFLICT"
	case errors.Is(err, service.ErrWallpaperLimit):
		return http.StatusConflict, "WALLPAPER_LIMIT"
	case errors.Is(err, service.ErrUnsupportedFile):
		return http.StatusBadRequest, "INVALID_INPUT"
	default:
		return http.StatusInternalServerError, "INTERNAL"
	}
}

// decode parses the JSON request body into v.
func decode(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// pageParams extracts offset and limit from query params with defaults.
func pageParams(r *http.Request) (offset, limit int) {
	limit = 50
	offset = 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := parseInt(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := parseInt(v); err == nil && n >= 0 {
			offset = n
		}
	}
	return
}

func parseInt(s string) (int, error) {
	var n int
	err := json.Unmarshal([]byte(s), &n)
	return n, err
}
