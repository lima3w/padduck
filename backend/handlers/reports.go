package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
)

// ─────────────────────────────────────────────────────────────────────────────
// Utilisation history (#220)
// ─────────────────────────────────────────────────────────────────────────────

// GetSubnetUtilisationHistory handles GET /api/v1/subnets/:id/utilisation/history
func (h *Handler) GetSubnetUtilisationHistory(c *fiber.Ctx) error {
	if err := h.permCheck(c, "subnets:read"); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid subnet ID"})
	}

	days := 30
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 {
			days = v
		}
	}

	points, err := h.service.Reports.GetUtilisationHistory(c.Context(), int64(id), days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if points == nil {
		points = make([]*models.SubnetUtilisationPoint, 0)
	}

	return c.JSON(fiber.Map{
		"subnet_id": id,
		"days":      days,
		"data":      points,
	})
}

// GetUtilisationTrends handles GET /api/v1/admin/reports/utilisation-trends
func (h *Handler) GetUtilisationTrends(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}

	trends, err := h.service.Reports.GetUtilisationTrends(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if trends == nil {
		trends = make([]*models.SubnetUtilisationTrend, 0)
	}

	return c.JSON(fiber.Map{"trends": trends})
}

// ─────────────────────────────────────────────────────────────────────────────
// Threshold alerts (#221)
// ─────────────────────────────────────────────────────────────────────────────

// GetSubnetsNearCapacity handles GET /api/v1/admin/reports/subnets-near-capacity
func (h *Handler) GetSubnetsNearCapacity(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}

	threshold := 80.0
	if t := c.Query("threshold"); t != "" {
		if v, err := strconv.ParseFloat(t, 64); err == nil && v > 0 {
			threshold = v
		}
	}

	subnets, err := h.service.Reports.GetSubnetsNearCapacity(c.Context(), threshold)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if subnets == nil {
		subnets = make([]*models.SubnetUtilisationTrend, 0)
	}

	return c.JSON(fiber.Map{
		"threshold_pct": threshold,
		"subnets":       subnets,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Scheduled reports (#222)
// ─────────────────────────────────────────────────────────────────────────────

type createScheduledReportRequest struct {
	Name            string         `json:"name"`
	ReportType      string         `json:"report_type"`
	ScheduleCron    string         `json:"schedule_cron"`
	RecipientEmails []string       `json:"recipient_emails"`
	Filters         map[string]any `json:"filters"`
	Format          string         `json:"format"`
}

type updateScheduledReportRequest struct {
	Name            string         `json:"name"`
	ReportType      string         `json:"report_type"`
	ScheduleCron    string         `json:"schedule_cron"`
	RecipientEmails []string       `json:"recipient_emails"`
	Filters         map[string]any `json:"filters"`
	Format          string         `json:"format"`
}

// ListScheduledReports handles GET /api/v1/admin/reports/scheduled
func (h *Handler) ListScheduledReports(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}

	reports, err := h.service.Reports.ListScheduledReports(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if reports == nil {
		reports = make([]*models.ScheduledReport, 0)
	}

	return c.JSON(fiber.Map{"reports": reports})
}

// CreateScheduledReport handles POST /api/v1/admin/reports/scheduled
func (h *Handler) CreateScheduledReport(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	req := new(createScheduledReportRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Name == "" || req.ReportType == "" || req.ScheduleCron == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name, report_type and schedule_cron are required"})
	}

	format := req.Format
	if format == "" {
		format = "csv"
	}
	if req.Filters == nil {
		req.Filters = map[string]any{}
	}
	if req.RecipientEmails == nil {
		req.RecipientEmails = []string{}
	}

	user, _ := c.Locals("user").(*models.User)
	var createdBy int64
	if user != nil {
		createdBy = user.ID
	}

	report, err := h.service.Reports.CreateScheduledReport(c.Context(),
		req.Name, req.ReportType, req.ScheduleCron,
		req.RecipientEmails, req.Filters, format, createdBy,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(report)
}

// GetScheduledReport handles GET /api/v1/admin/reports/scheduled/:id
func (h *Handler) GetScheduledReport(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid report ID"})
	}

	report, err := h.service.Reports.GetScheduledReport(c.Context(), int64(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "report not found"})
	}

	return c.JSON(report)
}

// UpdateScheduledReport handles PUT /api/v1/admin/reports/scheduled/:id
func (h *Handler) UpdateScheduledReport(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid report ID"})
	}

	req := new(updateScheduledReportRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Name == "" || req.ReportType == "" || req.ScheduleCron == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name, report_type and schedule_cron are required"})
	}

	format := req.Format
	if format == "" {
		format = "csv"
	}
	if req.Filters == nil {
		req.Filters = map[string]any{}
	}
	if req.RecipientEmails == nil {
		req.RecipientEmails = []string{}
	}

	report, err := h.service.Reports.UpdateScheduledReport(c.Context(),
		int64(id), req.Name, req.ReportType, req.ScheduleCron,
		req.RecipientEmails, req.Filters, format,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(report)
}

// DeleteScheduledReport handles DELETE /api/v1/admin/reports/scheduled/:id
func (h *Handler) DeleteScheduledReport(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid report ID"})
	}

	if err := h.service.Reports.DeleteScheduledReport(c.Context(), int64(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// RunScheduledReportNow handles POST /api/v1/admin/reports/scheduled/:id/run
func (h *Handler) RunScheduledReportNow(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid report ID"})
	}

	report, err := h.service.Reports.GetScheduledReport(c.Context(), int64(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "report not found"})
	}

	if err := h.service.Reports.RunScheduledReport(c.Context(), report); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "report executed successfully"})
}

// ─────────────────────────────────────────────────────────────────────────────
// Export endpoints (#223)
// ─────────────────────────────────────────────────────────────────────────────

// ExportSubnets handles GET /api/v1/admin/reports/export/subnets
func (h *Handler) ExportSubnets(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}

	format := c.Query("format", "csv")

	data, filename, contentType, err := h.service.Reports.ExportSubnets(c.Context(), format)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	return c.Send(data)
}

// ExportIPs handles GET /api/v1/admin/reports/export/ips
func (h *Handler) ExportIPs(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}

	subnetIDStr := c.Query("subnet_id")
	if subnetIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "subnet_id is required"})
	}

	subnetID, err := strconv.ParseInt(subnetIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid subnet_id"})
	}

	format := c.Query("format", "csv")

	data, filename, contentType, err := h.service.Reports.ExportIPs(c.Context(), subnetID, format)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	return c.Send(data)
}

// ExportInactiveIPs handles GET /api/v1/admin/reports/export/inactive-ips
func (h *Handler) ExportInactiveIPs(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}

	days := 90
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 {
			days = v
		}
	}

	format := c.Query("format", "csv")

	data, filename, contentType, err := h.service.Reports.ExportInactiveIPs(c.Context(), days, format)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	return c.Send(data)
}

// ─────────────────────────────────────────────────────────────────────────────
// Inactive IP reclamation (#224)
// ─────────────────────────────────────────────────────────────────────────────

// GetInactiveIPs handles GET /api/v1/admin/reports/inactive-ips
func (h *Handler) GetInactiveIPs(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}

	days := 90
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 {
			days = v
		}
	}

	var sectionID *int64
	if s := c.Query("section_id"); s != "" {
		if v, err := strconv.ParseInt(s, 10, 64); err == nil {
			sectionID = &v
		}
	}

	ips, err := h.service.Reports.GetInactiveIPs(c.Context(), days, sectionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if ips == nil {
		ips = make([]*models.InactiveIPReport, 0)
	}

	return c.JSON(fiber.Map{
		"days":     days,
		"count":    len(ips),
		"inactive": ips,
	})
}

// BulkReleaseIPsRequest is the body for POST /api/v1/admin/ip-addresses/bulk-release
type BulkReleaseIPsRequest struct {
	IPIDs []int64 `json:"ip_ids"`
}

// BulkReleaseIPs handles POST /api/v1/admin/ip-addresses/bulk-release
func (h *Handler) BulkReleaseIPs(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	req := new(BulkReleaseIPsRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if len(req.IPIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ip_ids must not be empty"})
	}

	user, _ := c.Locals("user").(*models.User)
	var operatorID int64
	if user != nil {
		operatorID = user.ID
	}

	count, err := h.service.Reports.BulkReleaseIPs(c.Context(), req.IPIDs, operatorID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"released_count": count,
	})
}
