package handlers

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"padduck/services"
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
// Applies the dns_zone_filter_mode / dns_zone_filter_list / dns_zone_filter_auto_allow config settings.
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

	// Apply zone visibility filter
	zones = h.applyDNSZoneFilter(c.Context(), zones)

	return c.JSON(fiber.Map{"configured": true, "zones": zones})
}

// applyDNSZoneFilter applies the dns_zone_filter_mode setting to the given zone list.
// In "allow_all" mode (default): all zones shown, listed zones hidden.
// In "block_all" mode: only listed zones shown; with auto_allow, new zones are added to the list.
func (h *Handler) applyDNSZoneFilter(ctx context.Context, zones []services.ZoneInfo) []services.ZoneInfo {
	if h.service.Config == nil {
		return zones
	}
	mode, _ := h.service.Config.GetCtx(ctx, "dns_zone_filter_mode")
	listRaw, _ := h.service.Config.GetCtx(ctx, "dns_zone_filter_list")

	// Parse the exception list (one zone per line, trim whitespace)
	listed := map[string]struct{}{}
	for _, z := range strings.Split(listRaw, "\n") {
		if t := strings.TrimSpace(z); t != "" {
			listed[strings.ToLower(t)] = struct{}{}
		}
	}

	switch mode {
	case "block_all":
		autoAllow, _ := h.service.Config.GetCtx(ctx, "dns_zone_filter_auto_allow")
		if autoAllow == "true" {
			// Add any new zones to the list automatically
			var newEntries []string
			for _, z := range zones {
				name := strings.ToLower(z.Name)
				if _, known := listed[name]; !known {
					newEntries = append(newEntries, z.Name)
					listed[name] = struct{}{}
				}
			}
			if len(newEntries) > 0 {
				updated := listRaw
				for _, ne := range newEntries {
					if updated != "" {
						updated += "\n"
					}
					updated += ne
				}
				_ = h.service.Config.SetCtx(ctx, "dns_zone_filter_list", updated)
			}
			// After auto-allow, show everything in the updated list
			var result []services.ZoneInfo
			for _, z := range zones {
				if _, ok := listed[strings.ToLower(z.Name)]; ok {
					result = append(result, z)
				}
			}
			return result
		}
		// Strict block_all: only show listed zones
		var result []services.ZoneInfo
		for _, z := range zones {
			if _, ok := listed[strings.ToLower(z.Name)]; ok {
				result = append(result, z)
			}
		}
		return result

	default: // "allow_all"
		if len(listed) == 0 {
			return zones
		}
		var result []services.ZoneInfo
		for _, z := range zones {
			if _, hidden := listed[strings.ToLower(z.Name)]; !hidden {
				result = append(result, z)
			}
		}
		return result
	}
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
