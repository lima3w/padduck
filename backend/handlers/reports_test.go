package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"padduck/models"
)

// ─────────────────────────────────────────────────────────────────────────────
// GetSubnetUtilisationHistory (#220)
// ─────────────────────────────────────────────────────────────────────────────

func TestGetSubnetUtilisationHistory_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/subnets/:id/utilisation/history", h.GetSubnetUtilisationHistory)

	req := httptest.NewRequest("GET", "/subnets/1/utilisation/history", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetSubnetUtilisationHistory_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/subnets/:id/utilisation/history", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "viewer"})
		return h.GetSubnetUtilisationHistory(c)
	})

	req := httptest.NewRequest("GET", "/subnets/1/utilisation/history", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	// "subnets:read" is denied for viewer (not in their allowed perms), so 403
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetUtilisationTrends (#220)
// ─────────────────────────────────────────────────────────────────────────────

func TestGetUtilisationTrends_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/reports/utilisation-trends", h.GetUtilisationTrends)

	req := httptest.NewRequest("GET", "/admin/reports/utilisation-trends", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetUtilisationTrends_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/reports/utilisation-trends", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.GetUtilisationTrends(c)
	})

	req := httptest.NewRequest("GET", "/admin/reports/utilisation-trends", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetSubnetsNearCapacity (#221)
// ─────────────────────────────────────────────────────────────────────────────

func TestGetSubnetsNearCapacity_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/reports/subnets-near-capacity", h.GetSubnetsNearCapacity)

	req := httptest.NewRequest("GET", "/admin/reports/subnets-near-capacity", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetSubnetsNearCapacity_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/reports/subnets-near-capacity", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.GetSubnetsNearCapacity(c)
	})

	req := httptest.NewRequest("GET", "/admin/reports/subnets-near-capacity", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// ListScheduledReports (#222)
// ─────────────────────────────────────────────────────────────────────────────

func TestListScheduledReports_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/reports/scheduled", h.ListScheduledReports)

	req := httptest.NewRequest("GET", "/admin/reports/scheduled", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListScheduledReports_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/reports/scheduled", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.ListScheduledReports(c)
	})

	req := httptest.NewRequest("GET", "/admin/reports/scheduled", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// CreateScheduledReport (#222)
// ─────────────────────────────────────────────────────────────────────────────

func TestCreateScheduledReport_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/reports/scheduled", h.CreateScheduledReport)

	req := httptest.NewRequest("POST", "/admin/reports/scheduled", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateScheduledReport_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/reports/scheduled", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.CreateScheduledReport(c)
	})

	req := httptest.NewRequest("POST", "/admin/reports/scheduled", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetScheduledReport (#222)
// ─────────────────────────────────────────────────────────────────────────────

func TestGetScheduledReport_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/reports/scheduled/:id", h.GetScheduledReport)

	req := httptest.NewRequest("GET", "/admin/reports/scheduled/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetScheduledReport_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/reports/scheduled/:id", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.GetScheduledReport(c)
	})

	req := httptest.NewRequest("GET", "/admin/reports/scheduled/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// UpdateScheduledReport (#222)
// ─────────────────────────────────────────────────────────────────────────────

func TestUpdateScheduledReport_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/admin/reports/scheduled/:id", h.UpdateScheduledReport)

	req := httptest.NewRequest("PUT", "/admin/reports/scheduled/1", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateScheduledReport_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/admin/reports/scheduled/:id", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.UpdateScheduledReport(c)
	})

	req := httptest.NewRequest("PUT", "/admin/reports/scheduled/1", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// DeleteScheduledReport (#222)
// ─────────────────────────────────────────────────────────────────────────────

func TestDeleteScheduledReport_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/admin/reports/scheduled/:id", h.DeleteScheduledReport)

	req := httptest.NewRequest("DELETE", "/admin/reports/scheduled/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteScheduledReport_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/admin/reports/scheduled/:id", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.DeleteScheduledReport(c)
	})

	req := httptest.NewRequest("DELETE", "/admin/reports/scheduled/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// RunScheduledReportNow (#222)
// ─────────────────────────────────────────────────────────────────────────────

func TestRunScheduledReportNow_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/reports/scheduled/:id/run", h.RunScheduledReportNow)

	req := httptest.NewRequest("POST", "/admin/reports/scheduled/1/run", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestRunScheduledReportNow_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/reports/scheduled/:id/run", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.RunScheduledReportNow(c)
	})

	req := httptest.NewRequest("POST", "/admin/reports/scheduled/1/run", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// ExportSubnets (#223)
// ─────────────────────────────────────────────────────────────────────────────

func TestExportSubnets_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/reports/export/subnets", h.ExportSubnets)

	req := httptest.NewRequest("GET", "/admin/reports/export/subnets", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestExportSubnets_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/reports/export/subnets", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.ExportSubnets(c)
	})

	req := httptest.NewRequest("GET", "/admin/reports/export/subnets", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// ExportIPs (#223)
// ─────────────────────────────────────────────────────────────────────────────

func TestExportIPs_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/reports/export/ips", h.ExportIPs)

	req := httptest.NewRequest("GET", "/admin/reports/export/ips?subnet_id=1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestExportIPs_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/reports/export/ips", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.ExportIPs(c)
	})

	req := httptest.NewRequest("GET", "/admin/reports/export/ips?subnet_id=1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// ExportInactiveIPs (#223)
// ─────────────────────────────────────────────────────────────────────────────

func TestExportInactiveIPs_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/reports/export/inactive-ips", h.ExportInactiveIPs)

	req := httptest.NewRequest("GET", "/admin/reports/export/inactive-ips", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestExportInactiveIPs_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/reports/export/inactive-ips", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.ExportInactiveIPs(c)
	})

	req := httptest.NewRequest("GET", "/admin/reports/export/inactive-ips", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// GetInactiveIPs (#224)
// ─────────────────────────────────────────────────────────────────────────────

func TestGetInactiveIPs_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/reports/inactive-ips", h.GetInactiveIPs)

	req := httptest.NewRequest("GET", "/admin/reports/inactive-ips", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetInactiveIPs_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/reports/inactive-ips", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.GetInactiveIPs(c)
	})

	req := httptest.NewRequest("GET", "/admin/reports/inactive-ips", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// BulkReleaseIPs (#224)
// ─────────────────────────────────────────────────────────────────────────────

func TestBulkReleaseIPs_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/ip-addresses/bulk-release", h.BulkReleaseIPs)

	req := httptest.NewRequest("POST", "/admin/ip-addresses/bulk-release",
		strings.NewReader(`{"ip_ids":[1,2]}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestBulkReleaseIPs_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/ip-addresses/bulk-release", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.BulkReleaseIPs(c)
	})

	req := httptest.NewRequest("POST", "/admin/ip-addresses/bulk-release",
		strings.NewReader(`{"ip_ids":[1,2]}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
