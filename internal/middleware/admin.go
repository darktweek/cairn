package middleware

import "net/http"

func Admin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := UserFromCtx(r.Context())
		if user == nil || user.Role != "admin" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":"forbidden","code":"FORBIDDEN"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}
