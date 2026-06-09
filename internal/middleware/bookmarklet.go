package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/darktweek/cairn/internal/service"
)

// BookmarkletAuth reads the token from the JSON body, validates it as a bookmarklet session,
// then re-injects the full body so the downstream handler can decode it again.
func BookmarkletAuth(svc service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			buf, err := io.ReadAll(r.Body)
			if err != nil {
				writeUnauthorized(w)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(buf))

			var peek struct {
				Token string `json:"token"`
			}
			if err := json.Unmarshal(buf, &peek); err != nil || peek.Token == "" {
				writeUnauthorized(w)
				return
			}

			user, sess, err := svc.ValidateSession(r.Context(), peek.Token)
			if err != nil || !sess.IsBookmarklet {
				writeUnauthorized(w)
				return
			}

			// Re-inject body for the handler.
			r.Body = io.NopCloser(bytes.NewReader(buf))

			ctx := withUser(r.Context(), user)
			ctx = withSession(ctx, sess)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
