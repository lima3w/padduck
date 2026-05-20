package handlers

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"ipam-next/models"
)

// All three audit handlers use a direct user.Role == "admin" check.

// ---------------------------------------------------------------------------
// GetAuditLogs — GET /api/v1/admin/audit-logs
// ---------------------------------------------------------------------------

func TestGetAuditLogs_NoUser_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Get("/admin/audit-logs", h.GetAuditLogs)
	resp, err := app.Test(httptest.NewRequest("GET", "/admin/audit-logs", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetAuditLogs_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", nonAdminUser)
		return c.Next()
	})
	app.Get("/admin/audit-logs", h.GetAuditLogs)
	resp, err := app.Test(httptest.NewRequest("GET", "/admin/audit-logs", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ExportAuditLogs — GET /api/v1/admin/audit-logs/export
// ---------------------------------------------------------------------------

func TestExportAuditLogs_NoUser_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Get("/admin/audit-logs/export", h.ExportAuditLogs)
	resp, err := app.Test(httptest.NewRequest("GET", "/admin/audit-logs/export", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestExportAuditLogs_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", nonAdminUser)
		return c.Next()
	})
	app.Get("/admin/audit-logs/export", h.ExportAuditLogs)
	resp, err := app.Test(httptest.NewRequest("GET", "/admin/audit-logs/export", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// PurgeAuditLogs — POST /api/v1/admin/audit-logs/purge
// ---------------------------------------------------------------------------

func TestPurgeAuditLogs_NoUser_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/admin/audit-logs/purge", h.PurgeAuditLogs)
	resp, err := app.Test(httptest.NewRequest("POST", "/admin/audit-logs/purge", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestPurgeAuditLogs_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", nonAdminUser)
		return c.Next()
	})
	app.Post("/admin/audit-logs/purge", h.PurgeAuditLogs)
	resp, err := app.Test(httptest.NewRequest("POST", "/admin/audit-logs/purge", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestFormatAuditLogsRedactsSensitiveValues(t *testing.T) {
	raw := `{"snmp_community":"public","name":"scan"}`
	logs := []*models.AuditLog{{
		ID:        1,
		CreatedAt: time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC),
		Username:  "admin",
		Action:    "updated",
		Status:    "success",
		NewValues: &raw,
	}}

	formatted := formatAuditLogs(logs)
	assert.Len(t, formatted, 1)
	assert.NotContains(t, *formatted[0].NewValues, "public")
	assert.Contains(t, *formatted[0].NewValues, "***REDACTED***")
}
