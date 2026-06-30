package middleware

import (
	"net/http"
	"strings"
)

// bookmarkletPath is the only endpoint that must accept cross-origin requests
// from arbitrary sites. Auth is token-based (no session cookie), so wildcard
// origin is safe.
const bookmarkletPath = "/api/bookmarks/quick"

// CORS returns a CORS middleware restricted to the given origin(s).
// The bookmarklet endpoint is always open to any origin.
func CORS(baseURL, env string) func(http.Handler) http.Handler {
	allowed := []string{baseURL}
	if env == "development" {
		allowed = append(allowed, "http://localhost:8080", "http://localhost:3000")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Bookmarklet: allow any origin, handle preflight and bail out.
			if r.URL.Path == bookmarkletPath {
				if origin != "" {
					w.Header().Set("Access-Control-Allow-Origin", "*")
					w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
					w.Header().Set("Access-Control-Max-Age", "86400")
				}
				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			// All other routes: restrict to known origins.
			if origin != "" && isAllowed(origin, allowed) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Vary", "Origin")
			}

			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Max-Age", "86400")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if strings.EqualFold(origin, a) {
			return true
		}
	}
	return false
}
