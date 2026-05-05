package handlers

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RateLimiter tracks requests per IP address
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
// limit: maximum requests allowed
// window: time window to count requests
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// Cleanup old entries every minute
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()

	return rl
}

// Allow checks if a request from the given IP should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Get or create request list for this IP
	requests, exists := rl.requests[ip]
	if !exists {
		rl.requests[ip] = []time.Time{now}
		return true
	}

	// Remove old requests outside the window
	validRequests := make([]time.Time, 0)
	for _, t := range requests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}

	// Check if limit exceeded
	if len(validRequests) >= rl.limit {
		rl.requests[ip] = validRequests
		return false
	}

	// Add new request
	validRequests = append(validRequests, now)
	rl.requests[ip] = validRequests
	return true
}

// cleanup removes old IP entries
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window * 2) // Keep 2 windows of history

	for ip, requests := range rl.requests {
		validRequests := make([]time.Time, 0)
		for _, t := range requests {
			if t.After(windowStart) {
				validRequests = append(validRequests, t)
			}
		}

		if len(validRequests) == 0 {
			delete(rl.requests, ip)
		} else {
			rl.requests[ip] = validRequests
		}
	}
}

// RateLimitMiddleware returns a Fiber middleware for rate limiting
// limit: requests per window
// window: time window (e.g., 1 minute)
func (h *Handler) RateLimitMiddleware(limit int, window time.Duration) fiber.Handler {
	rl := NewRateLimiter(limit, window)

	return func(c *fiber.Ctx) error {
		ip := c.IP()

		if !rl.Allow(ip) {
			return RespondError(c, fiber.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED",
				"Too many requests. Please try again later.",
				fmt.Sprintf("Rate limit: %d requests per %s", limit, window.String()))
		}

		return c.Next()
	}
}
