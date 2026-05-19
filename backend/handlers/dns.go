package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"ipam-next/services"
)

// CheckAllDNS handles POST /api/v1/admin/dns/check-all
// Triggers a background DNS check for all IPs that have a dns_name set.
func (h *Handler) CheckAllDNS(c *fiber.Ctx) error {
	// Admin-only via RBAC
	if err := h.permCheck(c, services.PermV2AuditRead); err != nil {
		return nil
	}
	job := h.service.Jobs.Enqueue("dns_check", "Check all DNS records", nil, 1, func(ctx context.Context, reporter *services.JobReporter) (interface{}, error) {
		reporter.Progress(0, 1, "checking DNS records")
		h.service.DNS.CheckAllDNS(ctx)
		reporter.Progress(1, 1, "DNS check complete")
		return fiber.Map{"message": "DNS check complete"}, nil
	})
	return c.Status(fiber.StatusAccepted).JSON(job)
}

// TestPowerDNSConnection handles POST /api/v1/admin/dns/test
// Tests connectivity to the configured PowerDNS API.
func (h *Handler) TestPowerDNSConnection(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AuditRead); err != nil {
		return nil
	}
	if err := h.service.DNS.TestPDNSConnection(c.Context()); err != nil {
		reqLogger(c).Error("PowerDNS connection test failed", "error", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "PowerDNS connection failed"})
	}
	return c.JSON(fiber.Map{"status": "ok"})
}

// TestTechnitiumConnection handles POST /api/v1/admin/dns/technitium/test
// Accepts optional JSON body {url, token, skip_tls} to test with unsaved values.
// Falls back to saved config when body fields are empty.
func (h *Handler) TestTechnitiumConnection(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AuditRead); err != nil {
		return nil
	}
	var body struct {
		URL     string `json:"url"`
		Token   string `json:"token"`
		SkipTLS bool   `json:"skip_tls"`
	}
	_ = c.BodyParser(&body)

	var err error
	if body.URL != "" && body.Token != "" {
		err = h.service.DNS.TestTechnitiumConnectionWith(c.Context(), body.URL, body.Token, body.SkipTLS)
	} else {
		err = h.service.DNS.TestTechnitiumConnection(c.Context())
	}
	if err != nil {
		reqLogger(c).Error("Technitium connection test failed", "error", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "Technitium connection failed"})
	}
	return c.JSON(fiber.Map{"status": "ok"})
}

// ListDNSZones handles GET /api/v1/dns/zones
// Returns the list of zones from the configured DNS provider, or {"configured": false} if none is set up.
func (h *Handler) ListDNSZones(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2NameserverList); err != nil {
		return nil
	}
	zones, configured, err := h.service.DNS.ListDNSZones(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing DNS zones", "error", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "DNS provider error"})
	}
	if !configured {
		return c.JSON(fiber.Map{"configured": false, "zones": []interface{}{}})
	}
	return c.JSON(fiber.Map{"configured": true, "zones": zones})
}

// GetDNSZoneRecords handles GET /api/v1/dns/zones/:zone/records
// Returns normalized records for a zone. Accepts optional ?type=A filter.
func (h *Handler) GetDNSZoneRecords(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2NameserverRead); err != nil {
		return nil
	}
	zone := c.Params("zone")
	if zone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "zone name is required"})
	}
	typeFilter := c.Query("type")
	records, err := h.service.DNS.GetDNSZoneRecords(c.Context(), zone, typeFilter)
	if err != nil {
		reqLogger(c).Error("error getting DNS zone records", "zone", zone, "error", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "DNS provider error"})
	}
	return c.JSON(fiber.Map{"zone": zone, "records": records})
}
