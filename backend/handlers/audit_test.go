package handlers

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"padduck/models"
)

// captureAuditFilter sends a GET to a minimal route and returns the parsed filter.
func captureAuditFilter(queryString string) *models.AuditLogFilter {
	var captured *models.AuditLogFilter
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		captured = buildAuditFilter(c)
		return c.SendStatus(fiber.StatusOK)
	})
	url := "/test"
	if queryString != "" {
		url += "?" + queryString
	}
	_, _ = app.Test(httptest.NewRequest("GET", url, nil))
	return captured
}

// All three audit handlers use a direct user.Role == "admin" check.

// ---------------------------------------------------------------------------
// GetAuditLogs — GET /api/v1/admin/audit-logs
// ---------------------------------------------------------------------------

func TestGetAuditLogs_NoUser_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/audit-logs", h.GetAuditLogs)
	resp, err := app.Test(httptest.NewRequest("GET", "/admin/audit-logs", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetAuditLogs_NonAdmin_Returns403(t *testing.T) {
	h := minHandler()
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
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/audit-logs/export", h.ExportAuditLogs)
	resp, err := app.Test(httptest.NewRequest("GET", "/admin/audit-logs/export", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestExportAuditLogs_NonAdmin_Returns403(t *testing.T) {
	h := minHandler()
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
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/audit-logs/purge", h.PurgeAuditLogs)
	resp, err := app.Test(httptest.NewRequest("POST", "/admin/audit-logs/purge", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestPurgeAuditLogs_NonAdmin_Returns403(t *testing.T) {
	h := minHandler()
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

// ---------------------------------------------------------------------------
// buildAuditFilter — query parameter parsing
// ---------------------------------------------------------------------------

func TestBuildAuditFilter_DefaultLimit(t *testing.T) {
	f := captureAuditFilter("")
	assert.Equal(t, 50, f.Limit)
	assert.Equal(t, 0, f.Offset)
}

func TestBuildAuditFilter_CustomLimit(t *testing.T) {
	f := captureAuditFilter("limit=25")
	assert.Equal(t, 25, f.Limit)
}

func TestBuildAuditFilter_InvalidLimitKeepsDefault(t *testing.T) {
	f := captureAuditFilter("limit=notanumber")
	assert.Equal(t, 50, f.Limit)
}

func TestBuildAuditFilter_NegativeLimitKeepsDefault(t *testing.T) {
	f := captureAuditFilter("limit=-5")
	assert.Equal(t, 50, f.Limit)
}

func TestBuildAuditFilter_Offset(t *testing.T) {
	f := captureAuditFilter("offset=100")
	assert.Equal(t, 100, f.Offset)
}

func TestBuildAuditFilter_StringFilters(t *testing.T) {
	f := captureAuditFilter("action=deleted&resource_type=subnet&username=alice&ip=192.168.1.1&status=success")
	assert.Equal(t, "deleted", f.Action)
	assert.Equal(t, "subnet", f.ResourceType)
	assert.Equal(t, "alice", f.Username)
	assert.Equal(t, "192.168.1.1", f.IPAddress)
	assert.Equal(t, "success", f.Status)
}

func TestBuildAuditFilter_ResourceID(t *testing.T) {
	f := captureAuditFilter("resource_id=42")
	assert.NotNil(t, f.ResourceID)
	assert.Equal(t, int64(42), *f.ResourceID)
}

func TestBuildAuditFilter_InvalidResourceIDIgnored(t *testing.T) {
	f := captureAuditFilter("resource_id=notanumber")
	assert.Nil(t, f.ResourceID)
}

func TestBuildAuditFilter_SinceParsed(t *testing.T) {
	f := captureAuditFilter("since=2026-01-01T00%3A00%3A00Z")
	assert.NotNil(t, f.Since)
	assert.Equal(t, 2026, f.Since.Year())
}

func TestBuildAuditFilter_InvalidSinceIgnored(t *testing.T) {
	f := captureAuditFilter("since=notadate")
	assert.Nil(t, f.Since)
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
