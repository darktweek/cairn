package middleware

import (
	"net/http"
)

// BodyLimit returns a middleware that limits the request body size.
func BodyLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// UserBodyLimit limits the request body to the authenticated user's upload
// size limit, falling back to def when no override is set. Must run after
// the session middleware. A 1 MB margin covers multipart encoding overhead.
func UserBodyLimit(def int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limit := def
			if u := UserFromCtx(r.Context()); u != nil && u.UploadSizeLimit != nil {
				limit = *u.UploadSizeLimit
			}
			r.Body = http.MaxBytesReader(w, r.Body, limit+1<<20)
			next.ServeHTTP(w, r)
		})
	}
}
