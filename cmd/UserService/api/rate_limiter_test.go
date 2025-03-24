package api

import (
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	// Initialize a new RateLimiter
	rateLimiter := NewRateLimiter()

	// Check that the map is initialized
	if rateLimiter == nil || rateLimiter.lastSent == nil {
		t.Fatalf("expected initialized RateLimiter with non-nil lastSent map, got nil")
	}
}

func TestSetAndGetLastSent(t *testing.T) {
	// Initialize a new RateLimiter
	rateLimiter := NewRateLimiter()

	uid := "user123"
	now := time.Now()

	// Set the last sent time for the user
	rateLimiter.SetLastSent(uid, now)

	// Get the last sent time for the user
	lastSent, exists := rateLimiter.GetLastSent(uid)

	// Ensure that the time was set correctly
	if !exists {
		t.Fatalf("expected lastSent to exist for uid %s", uid)
	}

	if lastSent != now {
		t.Fatalf("expected lastSent to be %v, got %v", now, lastSent)
	}
}

func TestRateLimiting(t *testing.T) {
	// Initialize a new RateLimiter
	rateLimiter := NewRateLimiter()

	uid := "user123"
	now := time.Now()
	rateLimitDuration := 1 * time.Minute

	// Set the last sent time for the user
	rateLimiter.SetLastSent(uid, now)

	// Try to resend before 1 minute has passed (30 seconds later)
	tooSoon := now.Add(30 * time.Second)

	// Check if the user can send again within 1 minute (should be false)
	if rateLimiter.CanSend(uid, rateLimitDuration, tooSoon) {
		t.Fatalf("expected rate limiting to prevent resending within 1 minute")
	}

	// Now simulate waiting for more than 1 minute
	later := now.Add(1 * time.Minute).Add(1 * time.Second)

	// Ensure that the user can resend after the limit has passed
	if !rateLimiter.CanSend(uid, rateLimitDuration, later) {
		t.Fatalf("expected user to be able to send after the 1-minute limit")
	}
}
