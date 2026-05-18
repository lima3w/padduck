package handlers

import (
	"context"
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

// StartCleanup launches a background goroutine that periodically evicts stale
// entries from the window map. It stops when ctx is cancelled.
func (r *tokenRateLimiter) StartCleanup(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.cleanup()
			}
		}
	}()
}

func (r *tokenRateLimiter) cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()
	cutoff := time.Now().Add(-time.Minute)
	for id, times := range r.windows {
		valid := times[:0]
		for _, t := range times {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(r.windows, id)
		} else {
			r.windows[id] = valid
		}
	}
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
