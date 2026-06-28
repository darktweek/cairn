package middleware

import "net/http"

func writeForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(`{"error":"forbidden","code":"FORBIDDEN"}`))
}

// RequirePermission gates a route on a single instance permission.
func RequirePermission(perm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := UserFromCtx(r.Context())
			if u == nil || !u.Can(perm) {
				writeForbidden(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission gates a route on holding at least one of the given
// permissions — used as the coarse guard for the whole admin area.
func RequireAnyPermission(perms ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := UserFromCtx(r.Context())
			if u == nil {
				writeForbidden(w)
				return
			}
			for _, p := range perms {
				if u.Can(p) {
					next.ServeHTTP(w, r)
					return
				}
			}
			writeForbidden(w)
		})
	}
}
