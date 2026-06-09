package middleware

import (
	"net/http"

	"github.com/darktweek/cairn/internal/service"
)

const sessionCookieName = "cairn_session"

func Auth(svc service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(sessionCookieName)
			if err != nil {
				writeUnauthorized(w)
				return
			}

			user, sess, err := svc.ValidateSession(r.Context(), cookie.Value)
			if err != nil {
				writeUnauthorized(w)
				return
			}

			ctx := withUser(r.Context(), user)
			ctx = withSession(ctx, sess)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"unauthorized","code":"UNAUTHORIZED"}`))
}
