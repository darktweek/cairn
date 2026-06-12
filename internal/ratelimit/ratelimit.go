// Package ratelimit provides a sliding-window limiter keyed by arbitrary
// strings. It backs both the per-IP HTTP middleware and the per-account
// login throttle: behind Docker NAT (or a spoofable X-Forwarded-For) every
// client can share one IP, so accounts need their own bucket to make
// brute-force impossible without locking out the whole network.
package ratelimit

import (
	"sync"
	"time"
)

// Config defines a window limit.
type Config struct {
	Max    int
	Window time.Duration
}

// Limiter is a thread-safe sliding-window rate limiter.
type Limiter struct {
	cfg     Config
	mu      sync.Mutex
	windows map[string][]time.Time
}

// New creates a limiter and starts its background cleanup.
func New(cfg Config) *Limiter {
	l := &Limiter{cfg: cfg, windows: make(map[string][]time.Time)}
	go func() {
		t := time.NewTicker(time.Minute)
		for range t.C {
			l.cleanup()
		}
	}()
	return l
}

// Allow records an attempt for key and reports whether it is within limits.
func (l *Limiter) Allow(key string) bool {
	now := time.Now()
	cutoff := now.Add(-l.cfg.Window)

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamps := l.windows[key]
	valid := timestamps[:0]
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= l.cfg.Max {
		l.windows[key] = valid
		return false
	}

	l.windows[key] = append(valid, now)
	return true
}

// Window returns the configured window (used for Retry-After headers).
func (l *Limiter) Window() time.Duration { return l.cfg.Window }

func (l *Limiter) cleanup() {
	cutoff := time.Now().Add(-l.cfg.Window)
	l.mu.Lock()
	defer l.mu.Unlock()
	for key, timestamps := range l.windows {
		valid := timestamps[:0]
		for _, t := range timestamps {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(l.windows, key)
		} else {
			l.windows[key] = valid
		}
	}
}
