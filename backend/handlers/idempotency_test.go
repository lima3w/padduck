package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestIdempotencyMiddleware_ReplaysMatchingRequest(t *testing.T) {
	h := &Handler{idempotency: newIdempotencyStore()}
	app := fiber.New()
	calls := 0
	app.Post("/automation/ip-addresses/allocate", h.IdempotencyMiddleware, func(c *fiber.Ctx) error {
		calls++
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"call": calls})
	})

	body := `{"subnet_id":1,"assigned_to":"ops"}`
	req1 := httptest.NewRequest("POST", "/automation/ip-addresses/allocate", strings.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Idempotency-Key", "retry-1")
	resp1, err := app.Test(req1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp1.StatusCode)

	req2 := httptest.NewRequest("POST", "/automation/ip-addresses/allocate", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Idempotency-Key", "retry-1")
	resp2, err := app.Test(req2)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp2.StatusCode)
	assert.Equal(t, "true", resp2.Header.Get("X-Idempotent-Replay"))
	assert.Equal(t, 1, calls)
}

func TestIdempotencyMiddleware_RejectsKeyReuseWithDifferentBody(t *testing.T) {
	h := &Handler{idempotency: newIdempotencyStore()}
	app := fiber.New()
	app.Post("/automation/ip-addresses/allocate", h.IdempotencyMiddleware, func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"ok": true})
	})

	req1 := httptest.NewRequest("POST", "/automation/ip-addresses/allocate", strings.NewReader(`{"subnet_id":1}`))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Idempotency-Key", "retry-1")
	_, err := app.Test(req1)
	assert.NoError(t, err)

	req2 := httptest.NewRequest("POST", "/automation/ip-addresses/allocate", strings.NewReader(`{"subnet_id":2}`))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Idempotency-Key", "retry-1")
	resp2, err := app.Test(req2)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusConflict, resp2.StatusCode)
}
