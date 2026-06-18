package handlers

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

// ─────────────────────────────────────────────────────────────────────────────
// Utilization history (#220)
// ─────────────────────────────────────────────────────────────────────────────

// GetSubnetUtilizationHistory handles GET /api/v1/subnets/:id/utilization/history
func (h *Handler) GetSubnetUtilizationHistory(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2SubnetRead) {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid subnet ID")
	}

	days := 30
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 {
			days = v
		}
	}

	points, err := h.ops.Reports.GetUtilizationHistory(c.Context(), int64(id), days)
	if err != nil {
		reqLogger(c).Error("get utilization history failed", "subnet_id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	if points == nil {
		points = make([]*models.SubnetUtilizationPoint, 0)
	}

	return c.JSON(fiber.Map{
		"subnet_id": id,
		"days":      days,
		"data":      points,
	})
}

// GetUtilizationTrends handles GET /api/v1/admin/reports/utilization-trends
func (h *Handler) GetUtilizationTrends(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}

	trends, err := h.ops.Reports.GetUtilizationTrends(c.Context())
	if err != nil {
		reqLogger(c).Error("get utilization trends failed", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	if trends == nil {
		trends = make([]*models.SubnetUtilizationTrend, 0)
	}

	return c.JSON(fiber.Map{"trends": trends})
}

// ─────────────────────────────────────────────────────────────────────────────
// Threshold alerts (#221)
// ─────────────────────────────────────────────────────────────────────────────

// GetSubnetsNearCapacity handles GET /api/v1/admin/reports/subnets-near-capacity
func (h *Handler) GetSubnetsNearCapacity(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}

	threshold := 80.0
	if t := c.Query("threshold"); t != "" {
		if v, err := strconv.ParseFloat(t, 64); err == nil && v > 0 {
			threshold = v
		}
	}

	subnets, err := h.ops.Reports.GetSubnetsNearCapacity(c.Context(), threshold)
	if err != nil {
		reqLogger(c).Error("get subnets near capacity failed", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	if subnets == nil {
		subnets = make([]*models.SubnetUtilizationTrend, 0)
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
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}

	reports, err := h.ops.Reports.ListScheduledReports(c.Context())
	if err != nil {
		reqLogger(c).Error("list scheduled reports failed", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	if reports == nil {
		reports = make([]*models.ScheduledReport, 0)
	}

	return c.JSON(fiber.Map{"reports": reports})
}

// CreateScheduledReport handles POST /api/v1/admin/reports/scheduled
func (h *Handler) CreateScheduledReport(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}

	req := new(createScheduledReportRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	if req.Name == "" || req.ReportType == "" || req.ScheduleCron == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "name, report_type and schedule_cron are required")
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

	report, err := h.ops.Reports.CreateScheduledReport(c.Context(),
		req.Name, req.ReportType, req.ScheduleCron,
		req.RecipientEmails, req.Filters, format, createdBy,
	)
	if err != nil {
		reqLogger(c).Error("create scheduled report failed", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	return c.Status(fiber.StatusCreated).JSON(report)
}

// GetScheduledReport handles GET /api/v1/admin/reports/scheduled/:id
func (h *Handler) GetScheduledReport(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid report ID")
	}

	report, err := h.ops.Reports.GetScheduledReport(c.Context(), int64(id))
	if err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "report not found")
	}

	return c.JSON(report)
}

// UpdateScheduledReport handles PUT /api/v1/admin/reports/scheduled/:id
func (h *Handler) UpdateScheduledReport(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid report ID")
	}

	req := new(updateScheduledReportRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	if req.Name == "" || req.ReportType == "" || req.ScheduleCron == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "name, report_type and schedule_cron are required")
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

	report, err := h.ops.Reports.UpdateScheduledReport(c.Context(),
		int64(id), req.Name, req.ReportType, req.ScheduleCron,
		req.RecipientEmails, req.Filters, format,
	)
	if err != nil {
		reqLogger(c).Error("update scheduled report failed", "report_id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	return c.JSON(report)
}

// DeleteScheduledReport handles DELETE /api/v1/admin/reports/scheduled/:id
func (h *Handler) DeleteScheduledReport(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid report ID")
	}

	if err := h.ops.Reports.DeleteScheduledReport(c.Context(), int64(id)); err != nil {
		reqLogger(c).Error("delete scheduled report failed", "report_id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// RunScheduledReportNow handles POST /api/v1/admin/reports/scheduled/:id/run
func (h *Handler) RunScheduledReportNow(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid report ID")
	}

	report, err := h.ops.Reports.GetScheduledReport(c.Context(), int64(id))
	if err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "report not found")
	}

	if c.QueryBool("async") {
		job := h.ops.Jobs.Enqueue("report", "Run scheduled report "+report.Name, fiber.Map{"report_id": report.ID}, 2, func(ctx context.Context, reporter *services.JobReporter) (interface{}, error) {
			reporter.Progress(0, 1, "running report")
			if err := h.ops.Reports.RunScheduledReport(ctx, report); err != nil {
				return nil, err
			}
			reporter.Progress(1, 1, "report complete")
			return fiber.Map{"report_id": report.ID}, nil
		})
		return c.Status(fiber.StatusAccepted).JSON(job)
	}

	if err := h.ops.Reports.RunScheduledReport(c.Context(), report); err != nil {
		reqLogger(c).Error("run scheduled report failed", "report_id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	return c.JSON(fiber.Map{"message": "report executed successfully"})
}

// ─────────────────────────────────────────────────────────────────────────────
// Export endpoints (#223)
// ─────────────────────────────────────────────────────────────────────────────

// ExportSubnets handles GET /api/v1/admin/reports/export/subnets
func (h *Handler) ExportSubnets(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}

	format := c.Query("format", "csv")

	data, filename, contentType, err := h.ops.Reports.ExportSubnets(c.Context(), format)
	if err != nil {
		reqLogger(c).Error("export subnets failed", "format", format, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	return c.Send(data)
}

// ExportIPs handles GET /api/v1/admin/reports/export/ips
func (h *Handler) ExportIPs(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}

	subnetIDStr := c.Query("subnet_id")
	if subnetIDStr == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "subnet_id is required")
	}

	subnetID, err := strconv.ParseInt(subnetIDStr, 10, 64)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid subnet_id")
	}

	format := c.Query("format", "csv")

	data, filename, contentType, err := h.ops.Reports.ExportIPs(c.Context(), subnetID, format)
	if err != nil {
		reqLogger(c).Error("export IPs failed", "subnet_id", subnetID, "format", format, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	return c.Send(data)
}

// ExportInactiveIPs handles GET /api/v1/admin/reports/export/inactive-ips
func (h *Handler) ExportInactiveIPs(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}

	days := 90
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 {
			days = v
		}
	}

	format := c.Query("format", "csv")

	data, filename, contentType, err := h.ops.Reports.ExportInactiveIPs(c.Context(), days, format)
	if err != nil {
		reqLogger(c).Error("export inactive IPs failed", "days", days, "format", format, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
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
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}

	days := 90
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 {
			days = v
		}
	}

	var networkID *int64
	if s := c.Query("network_id"); s != "" {
		if v, err := strconv.ParseInt(s, 10, 64); err == nil {
			networkID = &v
		}
	}

	ips, err := h.ops.Reports.GetInactiveIPs(c.Context(), days, networkID)
	if err != nil {
		reqLogger(c).Error("get inactive IPs failed", "days", days, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
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
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}

	req := new(BulkReleaseIPsRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	if len(req.IPIDs) == 0 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "ip_ids must not be empty")
	}

	user, _ := c.Locals("user").(*models.User)
	var operatorID int64
	if user != nil {
		operatorID = user.ID
	}

	count, err := h.ops.Reports.BulkReleaseIPs(c.Context(), req.IPIDs, operatorID)
	if err != nil {
		reqLogger(c).Error("bulk release IPs failed", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	return c.JSON(fiber.Map{
		"released_count": count,
	})
}

// BulkDeleteIPsRequest is the body for POST /api/v1/admin/ip-addresses/bulk-delete
type BulkDeleteIPsRequest struct {
	IPIDs []int64 `json:"ip_ids"`
}

// BulkDeleteIPs handles POST /api/v1/admin/ip-addresses/bulk-delete
func (h *Handler) BulkDeleteIPs(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}

	req := new(BulkDeleteIPsRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	if len(req.IPIDs) == 0 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "ip_ids must not be empty")
	}

	deleted, err := h.ops.IPAM.BulkDeleteIPAddresses(c.Context(), req.IPIDs)
	if err != nil {
		reqLogger(c).Error("bulk delete IPs failed", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to delete IP addresses")
	}

	return c.JSON(fiber.Map{
		"deleted_count": deleted,
	})
}
