package handlers

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"padduck/models"
)

// auditLog itself calls h.service.Audit.Log with a nil service and is
// covered by integration tests instead. The pure context-extraction
// helpers below have no service/repo dependency and are fully testable.

func TestAuditUserFromCtx_WithUser(t *testing.T) {
	app := fiber.New()
	app.Get("/x", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{ID: 42, Username: "zack"})
		id, username := auditUserFromCtx(c)
		return c.JSON(fiber.Map{"id": id, "username": username})
	})
	req := httptest.NewRequest("GET", "/x", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	data, _ := io.ReadAll(resp.Body)
	var body map[string]interface{}
	assert.NoError(t, json.Unmarshal(data, &body))
	assert.Equal(t, float64(42), body["id"])
	assert.Equal(t, "zack", body["username"])
}

func TestAuditUserFromCtx_NoUser(t *testing.T) {
	app := fiber.New()
	app.Get("/x", func(c *fiber.Ctx) error {
		id, username := auditUserFromCtx(c)
		return c.JSON(fiber.Map{"id": id, "username": username})
	})
	req := httptest.NewRequest("GET", "/x", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	data, _ := io.ReadAll(resp.Body)
	var body map[string]interface{}
	assert.NoError(t, json.Unmarshal(data, &body))
	assert.Nil(t, body["id"])
	assert.Equal(t, "", body["username"])
}

func TestOrgIDFromCtx_Set(t *testing.T) {
	app := fiber.New()
	app.Get("/x", func(c *fiber.Ctx) error {
		orgID := int64(7)
		c.Locals("orgID", &orgID)
		got := orgIDFromCtx(c)
		return c.JSON(fiber.Map{"org_id": got})
	})
	req := httptest.NewRequest("GET", "/x", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	data, _ := io.ReadAll(resp.Body)
	var body map[string]interface{}
	assert.NoError(t, json.Unmarshal(data, &body))
	assert.Equal(t, float64(7), body["org_id"])
}

func TestOrgIDFromCtx_NotSet(t *testing.T) {
	app := fiber.New()
	app.Get("/x", func(c *fiber.Ctx) error {
		got := orgIDFromCtx(c)
		return c.JSON(fiber.Map{"org_id": got})
	})
	req := httptest.NewRequest("GET", "/x", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	data, _ := io.ReadAll(resp.Body)
	var body map[string]interface{}
	assert.NoError(t, json.Unmarshal(data, &body))
	assert.Nil(t, body["org_id"])
}

func TestCallerID_WithUser(t *testing.T) {
	app := fiber.New()
	app.Get("/x", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{ID: 99})
		got := callerID(c)
		return c.JSON(fiber.Map{"caller_id": got})
	})
	req := httptest.NewRequest("GET", "/x", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	data, _ := io.ReadAll(resp.Body)
	var body map[string]interface{}
	assert.NoError(t, json.Unmarshal(data, &body))
	assert.Equal(t, float64(99), body["caller_id"])
}

func TestCallerID_NoUser(t *testing.T) {
	app := fiber.New()
	app.Get("/x", func(c *fiber.Ctx) error {
		got := callerID(c)
		return c.JSON(fiber.Map{"caller_id": got})
	})
	req := httptest.NewRequest("GET", "/x", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	data, _ := io.ReadAll(resp.Body)
	var body map[string]interface{}
	assert.NoError(t, json.Unmarshal(data, &body))
	assert.Nil(t, body["caller_id"])
}

func TestBypassPolicyFromCtx(t *testing.T) {
	cases := []struct {
		name string
		set  bool
		val  bool
		want bool
	}{
		{"not set", false, false, false},
		{"set true", true, true, true},
		{"set false", true, false, false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/x", func(c *fiber.Ctx) error {
				if tc.set {
					c.Locals("bypassPolicy", tc.val)
				}
				got := bypassPolicyFromCtx(c)
				return c.JSON(fiber.Map{"bypass": got})
			})
			req := httptest.NewRequest("GET", "/x", nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			data, _ := io.ReadAll(resp.Body)
			var body map[string]interface{}
			assert.NoError(t, json.Unmarshal(data, &body))
			assert.Equal(t, tc.want, body["bypass"])
		})
	}
}
