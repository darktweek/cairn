package ratelimit

import (
	"testing"
	"time"
)

func TestAllowUpToMax(t *testing.T) {
	l := New(Config{Max: 3, Window: time.Minute})
	for i := 1; i <= 3; i++ {
		if !l.Allow("k") {
			t.Fatalf("attempt %d should be allowed", i)
		}
	}
	if l.Allow("k") {
		t.Fatal("attempt 4 should be denied")
	}
}

func TestKeysAreIsolated(t *testing.T) {
	l := New(Config{Max: 1, Window: time.Minute})
	if !l.Allow("a") {
		t.Fatal("first attempt on a should pass")
	}
	if l.Allow("a") {
		t.Fatal("second attempt on a should be denied")
	}
	if !l.Allow("b") {
		t.Fatal("b must not share a's bucket")
	}
}

func TestWindowSlides(t *testing.T) {
	l := New(Config{Max: 1, Window: 30 * time.Millisecond})
	if !l.Allow("k") {
		t.Fatal("first attempt should pass")
	}
	if l.Allow("k") {
		t.Fatal("second immediate attempt should be denied")
	}
	time.Sleep(40 * time.Millisecond)
	if !l.Allow("k") {
		t.Fatal("attempt after window expiry should pass")
	}
}

func TestDeniedAttemptDoesNotExtendWindow(t *testing.T) {
	l := New(Config{Max: 1, Window: 50 * time.Millisecond})
	l.Allow("k")
	// Hammering while denied must not push the unlock further away.
	for i := 0; i < 5; i++ {
		l.Allow("k")
		time.Sleep(5 * time.Millisecond)
	}
	time.Sleep(30 * time.Millisecond) // total > 50ms since the only counted attempt
	if !l.Allow("k") {
		t.Fatal("denied attempts must not count toward the window")
	}
}
