package middleware

import (
	"crypto/sha256"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/darktweek/cairn/internal/ratelimit"
)

// RateLimitConfig defines limits for a route group.
type RateLimitConfig struct {
	Max    int
	Window time.Duration
}

// RateLimit returns a sliding window rate limiter middleware keyed by
// client IP. Behind Docker NAT or a spoofable X-Forwarded-For all clients
// can collapse onto one key, so this is a coarse spray guard — sensitive
// endpoints (login) add their own per-account limiter in the service.
func RateLimit(cfg RateLimitConfig, trustedProxy bool) func(http.Handler) http.Handler {
	l := ratelimit.New(ratelimit.Config{Max: cfg.Max, Window: cfg.Window})

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r, trustedProxy)
			if !l.Allow(hashIP(ip)) {
				retryAfter := int(l.Window().Seconds())
				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"too many requests","code":"RATE_LIMITED"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request, trusted bool) string {
	if trusted {
		if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
			return strings.TrimSpace(ip)
		}
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			// X-Forwarded-For: client, proxy1, proxy2 — take leftmost (real client).
			parts := strings.SplitN(forwarded, ",", 2)
			return strings.TrimSpace(parts[0])
		}
	}
	// r.RemoteAddr = "IP:port" — strip port.
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

func hashIP(ip string) string {
	h := sha256.Sum256([]byte(ip))
	return fmt.Sprintf("%x", h[:8])
}
