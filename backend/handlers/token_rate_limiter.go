package handlers

import (
	"sync"
	"time"
)

type tokenRateLimiter struct {
	mu      sync.Mutex
	windows map[int64][]time.Time
}

func newTokenRateLimiter() *tokenRateLimiter {
	return &tokenRateLimiter{windows: make(map[int64][]time.Time)}
}

// Allow returns true if the token is within its rate limit.
func (r *tokenRateLimiter) Allow(tokenID int64, limitPerMinute int) bool {
	if limitPerMinute <= 0 {
		return true
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-time.Minute)

	times := r.windows[tokenID]
	valid := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= limitPerMinute {
		r.windows[tokenID] = valid
		return false
	}
	r.windows[tokenID] = append(valid, now)
	return true
}
