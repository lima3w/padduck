package handlers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"ipam-next/services"
)

// GetAuditRetention handles GET /api/v1/admin/audit/retention
func (h *Handler) GetAuditRetention(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	s, err := h.service.Audit.GetRetentionSettings(c.Context())
	if err != nil {
		return h.StatusInternalServerError(c, "Failed to fetch audit retention settings", err.Error())
	}
	return c.JSON(s)
}

// UpdateAuditRetention handles PUT /api/v1/admin/audit/retention
// Body: {"retention_days": 365, "archive_enabled": false}
func (h *Handler) UpdateAuditRetention(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	var body struct {
		RetentionDays  int  `json:"retention_days"`
		ArchiveEnabled bool `json:"archive_enabled"`
	}
	if err := c.BodyParser(&body); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if body.RetentionDays < 30 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "retention_days must be >= 30")
	}
	s, err := h.service.Audit.UpdateRetentionSettings(c.Context(), body.RetentionDays, body.ArchiveEnabled)
	if err != nil {
		return h.StatusInternalServerError(c, "Failed to update audit retention settings", err.Error())
	}
	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "audit_retention_updated",
		ResourceType: "audit_retention",
		NewValues: map[string]interface{}{
			"retention_days":  body.RetentionDays,
			"archive_enabled": body.ArchiveEnabled,
		},
	})
	return c.JSON(s)
}

// PruneAuditLogs handles POST /api/v1/admin/audit/prune
func (h *Handler) PruneAuditLogs(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	deleted, err := h.service.Audit.PruneByRetentionSettings(c.Context())
	if err != nil {
		return h.StatusInternalServerError(c, "Failed to prune audit logs", err.Error())
	}
	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "audit_logs_pruned",
		ResourceType: "audit_log",
		NewValues:    map[string]int64{"deleted_count": deleted},
	})
	return c.JSON(fiber.Map{
		"deleted": deleted,
		"message": fmt.Sprintf("Pruned %d audit log entries", deleted),
	})
}

// ExportAuditLog handles GET /api/v1/admin/audit/export?format=json&since=<RFC3339>&until=<RFC3339>
func (h *Handler) ExportAuditLog(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	filter := buildAuditFilter(c)
	filter.Limit = 10000 // allow large export

	logs, err := h.service.Audit.ListAuditLogs(c.Context(), filter)
	if err != nil {
		return h.StatusInternalServerError(c, "Failed to export audit logs", err.Error())
	}

	format := c.Query("format", "json")
	if format == "csv" {
		var buf bytes.Buffer
		w := csv.NewWriter(&buf)
		_ = w.Write([]string{
			"id", "user_id", "username", "action", "resource_type",
			"resource_id", "resource_name", "status", "ip_address", "created_at",
		})
		for _, l := range logs {
			uid := ""
			if l.UserID != nil {
				uid = strconv.FormatInt(*l.UserID, 10)
			}
			rid := ""
			if l.ResourceID != nil {
				rid = strconv.FormatInt(*l.ResourceID, 10)
			}
			_ = w.Write([]string{
				strconv.FormatInt(l.ID, 10),
				uid,
				l.Username,
				l.Action,
				l.ResourceType,
				rid,
				l.ResourceName,
				l.Status,
				l.IPAddress,
				l.CreatedAt.Format(time.RFC3339),
			})
		}
		w.Flush()
		c.Set("Content-Type", "text/csv")
		c.Set("Content-Disposition", `attachment; filename="audit-export.csv"`)
		return c.Send(buf.Bytes())
	}

	// JSON format
	return c.JSON(formatAuditLogs(logs))
}
