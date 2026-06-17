package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/repository"
	"padduck/services"
)

// GetReconciliationReport handles GET /api/v1/admin/reports/reconciliation
func (h *Handler) GetReconciliationReport(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}
	ctx := c.Context()

	staleIPs, _ := h.service.Reports.GetInactiveIPs(ctx, 30, nil)
	if staleIPs == nil {
		staleIPs = []*models.InactiveIPReport{}
	}

	dnsEntries, _ := h.service.Reports.GetDNSAudit(ctx)
	if dnsEntries == nil {
		dnsEntries = []*repository.DNSAuditRow{}
	}

	overlaps, _ := h.service.OverlapReport(ctx)
	if overlaps == nil {
		overlaps = []*services.OverlapPair{}
	}

	return c.JSON(fiber.Map{
		"stale_ips": staleIPs,
		"dns_drift": dnsEntries,
		"overlaps":  overlaps,
	})
}
