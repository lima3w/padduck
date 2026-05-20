package handlers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	csvexport "ipam-next/internal/export"
	"ipam-next/models"
	"ipam-next/services"
)

// GetAuditLogs handles GET /api/v1/admin/audit-logs
// Supports query params: action, resource_type, username, ip, status, since, until, limit, offset
func (h *Handler) GetAuditLogs(c *fiber.Ctx) error {
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok || currentUser.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "admin access required"})
	}

	filter := buildAuditFilter(c)

	logs, err := h.service.Audit.ListAuditLogs(c.Context(), filter)
	if err != nil {
		return h.StatusInternalServerError(c, "Failed to fetch audit logs", err.Error())
	}

	total, err := h.service.Audit.CountAuditLogs(c.Context(), filter)
	if err != nil {
		return h.StatusInternalServerError(c, "Failed to count audit logs", err.Error())
	}

	return c.JSON(fiber.Map{
		"logs":   formatAuditLogs(logs),
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// ExportAuditLogs handles GET /api/v1/admin/audit-logs/export
// Returns a CSV download of matching audit log entries.
func (h *Handler) ExportAuditLogs(c *fiber.Ctx) error {
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok || currentUser.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "admin access required"})
	}

	filter := buildAuditFilter(c)
	filter.Limit = 10000 // allow larger export

	logs, err := h.service.Audit.ListAuditLogs(c.Context(), filter)
	if err != nil {
		return h.StatusInternalServerError(c, "Failed to export audit logs", err.Error())
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"id", "timestamp", "username", "action", "resource_type", "resource_id", "resource_name", "ip_address", "status", "error_message"})
	for _, l := range logs {
		rid := ""
		if l.ResourceID != nil {
			rid = strconv.FormatInt(*l.ResourceID, 10)
		}
		_ = w.Write([]string{
			strconv.FormatInt(l.ID, 10),
			l.CreatedAt.Format(time.RFC3339),
			csvexport.EscapeCSVCell(l.Username),
			csvexport.EscapeCSVCell(l.Action),
			csvexport.EscapeCSVCell(l.ResourceType),
			rid,
			csvexport.EscapeCSVCell(l.ResourceName),
			csvexport.EscapeCSVCell(l.IPAddress),
			csvexport.EscapeCSVCell(l.Status),
			csvexport.EscapeCSVCell(l.ErrorMessage),
		})
	}
	w.Flush()

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="audit-log-%s.csv"`, time.Now().Format("20060102-150405")))
	return c.Send(buf.Bytes())
}

// PurgeAuditLogs handles POST /api/v1/admin/audit-logs/purge
// Deletes log entries older than the configured retention period.
func (h *Handler) PurgeAuditLogs(c *fiber.Ctx) error {
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok || currentUser.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "admin access required"})
	}

	deleted, err := h.service.Audit.PurgeOldLogs(c.Context())
	if err != nil {
		return h.StatusInternalServerError(c, "Failed to purge audit logs", err.Error())
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "audit_logs_purged",
		ResourceType: "audit_log",
		NewValues:    map[string]int64{"deleted_count": deleted},
	})

	return c.JSON(fiber.Map{"deleted": deleted, "message": fmt.Sprintf("Purged %d audit log entries", deleted)})
}

// buildAuditFilter parses query parameters into an AuditLogFilter.
func buildAuditFilter(c *fiber.Ctx) *models.AuditLogFilter {
	filter := &models.AuditLogFilter{
		Action:       c.Query("action"),
		ResourceType: c.Query("resource_type"),
		Username:     c.Query("username"),
		IPAddress:    c.Query("ip"),
		Status:       c.Query("status"),
	}

	// parse resource_id if provided
	if ridStr := c.Query("resource_id"); ridStr != "" {
		if v, err := strconv.ParseInt(ridStr, 10, 64); err == nil {
			filter.ResourceID = &v
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
			filter.Limit = n
		}
	}
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if n, err := strconv.Atoi(offsetStr); err == nil && n >= 0 {
			filter.Offset = n
		}
	}

	if sinceStr := c.Query("since"); sinceStr != "" {
		if t, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			filter.Since = &t
		}
	}
	if untilStr := c.Query("until"); untilStr != "" {
		if t, err := time.Parse(time.RFC3339, untilStr); err == nil {
			filter.Until = &t
		}
	}

	return filter
}

type auditLogResponse struct {
	ID           int64   `json:"id"`
	Timestamp    string  `json:"timestamp"`
	UserID       *int64  `json:"user_id,omitempty"`
	Username     string  `json:"username"`
	Action       string  `json:"action"`
	ResourceType string  `json:"resource_type,omitempty"`
	ResourceID   *int64  `json:"resource_id,omitempty"`
	ResourceName string  `json:"resource_name,omitempty"`
	OldValues    *string `json:"old_values,omitempty"`
	NewValues    *string `json:"new_values,omitempty"`
	IPAddress    string  `json:"ip_address,omitempty"`
	Status       string  `json:"status"`
	ErrorMessage string  `json:"error_message,omitempty"`
}

func formatAuditLogs(logs []*models.AuditLog) []auditLogResponse {
	result := make([]auditLogResponse, 0, len(logs))
	for _, l := range logs {
		result = append(result, auditLogResponse{
			ID:           l.ID,
			Timestamp:    l.CreatedAt.Format(time.RFC3339),
			UserID:       l.UserID,
			Username:     l.Username,
			Action:       l.Action,
			ResourceType: l.ResourceType,
			ResourceID:   l.ResourceID,
			ResourceName: l.ResourceName,
			OldValues:    services.RedactSensitiveJSON(l.OldValues),
			NewValues:    services.RedactSensitiveJSON(l.NewValues),
			IPAddress:    l.IPAddress,
			Status:       l.Status,
			ErrorMessage: l.ErrorMessage,
		})
	}
	return result
}
