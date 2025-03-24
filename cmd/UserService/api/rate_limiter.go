package api

import "time"

type RateLimiter struct {
	lastSent map[string]time.Time
}

func (r *RateLimiter) GetLastSent(uid string) (time.Time, bool) {
	lastSent, exists := r.lastSent[uid]
	return lastSent, exists
}

func (r *RateLimiter) SetLastSent(uid string, timestamp time.Time) {
	r.lastSent[uid] = timestamp
}

// CanSend checks if the user is allowed to send another email based on the provided time and time limit
func (r *RateLimiter) CanSend(uid string, limit time.Duration, currentTime time.Time) bool {
	lastSent, exists := r.GetLastSent(uid)
	if !exists {
		return true // If the user hasn't sent an email before, they can send one
	}
	return currentTime.Sub(lastSent) >= limit
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{lastSent: make(map[string]time.Time)}
}
