package middleware

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// RateLimitConfig defines limits for a route group.
type RateLimitConfig struct {
	Max    int
	Window time.Duration
}

type rateLimiter struct {
	cfg          RateLimitConfig
	trustedProxy bool
	mu           sync.Mutex
	windows      map[string][]time.Time
}

// RateLimit returns a sliding window rate limiter middleware.
func RateLimit(cfg RateLimitConfig, trustedProxy bool) func(http.Handler) http.Handler {
	rl := &rateLimiter{
		cfg:          cfg,
		trustedProxy: trustedProxy,
		windows:      make(map[string][]time.Time),
	}

	// Background cleanup every minute.
	go func() {
		t := time.NewTicker(time.Minute)
		for range t.C {
			rl.cleanup()
		}
	}()

	return rl.middleware
}

func (rl *rateLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r, rl.trustedProxy)
		key := hashIP(ip)

		now := time.Now()
		cutoff := now.Add(-rl.cfg.Window)

		rl.mu.Lock()
		timestamps := rl.windows[key]
		valid := timestamps[:0]
		for _, t := range timestamps {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}

		if len(valid) >= rl.cfg.Max {
			rl.mu.Unlock()
			retryAfter := int(rl.cfg.Window.Seconds())
			w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"too many requests","code":"RATE_LIMITED"}`))
			return
		}

		rl.windows[key] = append(valid, now)
		rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func (rl *rateLimiter) cleanup() {
	cutoff := time.Now().Add(-rl.cfg.Window)
	rl.mu.Lock()
	defer rl.mu.Unlock()
	for key, timestamps := range rl.windows {
		valid := timestamps[:0]
		for _, t := range timestamps {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.windows, key)
		} else {
			rl.windows[key] = valid
		}
	}
}

func clientIP(r *http.Request, trusted bool) string {
	if trusted {
		if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
			return ip
		}
		if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
			return ip
		}
	}
	return r.RemoteAddr
}

func hashIP(ip string) string {
	h := sha256.Sum256([]byte(ip))
	return fmt.Sprintf("%x", h[:8])
}
