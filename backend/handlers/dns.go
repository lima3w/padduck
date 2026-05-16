package handlers

import (
	"log"

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
	go func() {
		h.service.DNS.CheckAllDNS(c.Context())
	}()
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "DNS check started"})
}

// TestPowerDNSConnection handles POST /api/v1/admin/dns/test
// Tests connectivity to the configured PowerDNS API.
func (h *Handler) TestPowerDNSConnection(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AuditRead); err != nil {
		return nil
	}
	if err := h.service.DNS.TestPDNSConnection(c.Context()); err != nil {
		log.Printf("PowerDNS connection test failed: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "ok"})
}

// TestTechnitiumConnection handles POST /api/v1/admin/dns/technitium/test
// Tests connectivity to the configured Technitium DNS server.
func (h *Handler) TestTechnitiumConnection(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AuditRead); err != nil {
		return nil
	}
	if err := h.service.DNS.TestTechnitiumConnection(c.Context()); err != nil {
		log.Printf("Technitium connection test failed: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "ok"})
}

// ListDNSZones handles GET /api/v1/dns/zones
// Returns the list of zones from PowerDNS, or {"configured": false} if not set up.
func (h *Handler) ListDNSZones(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2NameserverList); err != nil {
		return nil
	}
	zones, configured, err := h.service.DNS.ListPDNSZones(c.Context())
	if err != nil {
		log.Printf("Error listing DNS zones: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
	}
	if !configured {
		return c.JSON(fiber.Map{"configured": false, "zones": []interface{}{}})
	}
	return c.JSON(fiber.Map{"configured": true, "zones": zones})
}

// GetDNSZoneRecords handles GET /api/v1/dns/zones/:zone/records
// Returns the rrsets for a zone. Accepts optional ?type=A filter.
func (h *Handler) GetDNSZoneRecords(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2NameserverRead); err != nil {
		return nil
	}
	zone := c.Params("zone")
	if zone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "zone name is required"})
	}
	detail, err := h.service.DNS.GetPDNSZone(c.Context(), zone)
	if err != nil {
		log.Printf("Error getting DNS zone %s: %v", zone, err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
	}

	records := detail.RRSets
	if typeFilter := c.Query("type"); typeFilter != "" {
		filtered := detail.RRSets[:0]
		for _, rr := range detail.RRSets {
			if rr.Type == typeFilter {
				filtered = append(filtered, rr)
			}
		}
		records = filtered
	}

	return c.JSON(fiber.Map{"zone": detail.Zone, "records": records})
}
