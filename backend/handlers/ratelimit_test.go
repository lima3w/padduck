package handlers

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// RateLimiter.Allow — pure in-memory logic
// ---------------------------------------------------------------------------

func TestRateLimiter_AllowsUpToLimit(t *testing.T) {
	t.Parallel()

	rl := NewRateLimiter(3, 1*time.Minute)
	assert.True(t, rl.Allow("10.0.0.1"))
	assert.True(t, rl.Allow("10.0.0.1"))
	assert.True(t, rl.Allow("10.0.0.1"))
	assert.False(t, rl.Allow("10.0.0.1"))
}

func TestRateLimiter_DifferentIPsAreIndependent(t *testing.T) {
	t.Parallel()

	rl := NewRateLimiter(1, 1*time.Minute)
	assert.True(t, rl.Allow("10.0.0.1"))
	assert.False(t, rl.Allow("10.0.0.1"))
	assert.True(t, rl.Allow("10.0.0.2")) // different IP — not affected
}

func TestRateLimiter_WindowExpiry_AllowsAgain(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC)
	rl := newRateLimiterWithClock(1, time.Minute, func() time.Time { return now })
	assert.True(t, rl.Allow("10.0.0.3"))
	assert.False(t, rl.Allow("10.0.0.3"))
	now = now.Add(time.Minute + time.Nanosecond)
	assert.True(t, rl.Allow("10.0.0.3")) // window expired — allowed again
}

// ---------------------------------------------------------------------------
// RateLimitMiddleware — via fiber test
// ---------------------------------------------------------------------------

func TestRateLimitMiddleware_Returns429WhenLimitExceeded(t *testing.T) {
	t.Parallel()

	h := &Handler{}
	app := fiber.New()
	app.Use(h.RateLimitMiddleware(2, 1*time.Minute))
	app.Get("/test", func(c *fiber.Ctx) error { return c.SendStatus(200) })

	for i := 0; i < 2; i++ {
		resp, err := app.Test(httptest.NewRequest("GET", "/test", nil))
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	}

	resp, err := app.Test(httptest.NewRequest("GET", "/test", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusTooManyRequests, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// tokenRateLimiter.Allow — pure in-memory logic
// ---------------------------------------------------------------------------

func TestTokenRateLimiter_ZeroLimit_AlwaysAllows(t *testing.T) {
	t.Parallel()

	r := newTokenRateLimiter()
	for i := 0; i < 10; i++ {
		assert.True(t, r.Allow(1, 0))
	}
}

func TestTokenRateLimiter_AllowsUpToLimit(t *testing.T) {
	t.Parallel()

	r := newTokenRateLimiter()
	assert.True(t, r.Allow(42, 2))
	assert.True(t, r.Allow(42, 2))
	assert.False(t, r.Allow(42, 2))
}

func TestTokenRateLimiter_DifferentTokensAreIndependent(t *testing.T) {
	t.Parallel()

	r := newTokenRateLimiter()
	assert.True(t, r.Allow(1, 1))
	assert.False(t, r.Allow(1, 1))
	assert.True(t, r.Allow(2, 1)) // different token — not affected
}
