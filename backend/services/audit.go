package services

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"ipam-next/models"
)

// AuditService handles writing and querying audit logs.
type AuditService struct {
	svc *Service
}

func NewAuditService(svc *Service) *AuditService {
	return &AuditService{svc: svc}
}

// AuditEntry is the input to Log().
type AuditEntry struct {
	UserID       *int64
	Username     string
	Action       string
	ResourceType string
	ResourceID   *int64
	ResourceName string
	OldValues    interface{} // marshalled to JSON; nil = omitted
	NewValues    interface{} // marshalled to JSON; nil = omitted
	IPAddress    string
	UserAgent    string
	Status       string // "success" or "failure"
	ErrorMessage string
}

// Log writes an audit log entry. Errors are logged but never returned to callers.
func (a *AuditService) Log(ctx context.Context, e AuditEntry) {
	if e.Status == "" {
		e.Status = "success"
	}

	entry := &models.AuditLog{
		UserID:       e.UserID,
		Username:     e.Username,
		Action:       e.Action,
		ResourceType: e.ResourceType,
		ResourceID:   e.ResourceID,
		ResourceName: e.ResourceName,
		IPAddress:    e.IPAddress,
		UserAgent:    e.UserAgent,
		Status:       e.Status,
		ErrorMessage: e.ErrorMessage,
	}

	if e.OldValues != nil {
		b, err := json.Marshal(e.OldValues)
		if err == nil {
			s := string(b)
			entry.OldValues = &s
		}
	}
	if e.NewValues != nil {
		b, err := json.Marshal(e.NewValues)
		if err == nil {
			s := string(b)
			entry.NewValues = &s
		}
	}

	if err := a.svc.repository.CreateAuditLog(ctx, entry); err != nil {
		log.Printf("audit: failed to write log (action=%s): %v", e.Action, err)
	}
	if a.svc.Webhooks != nil {
		a.svc.Webhooks.Queue(ctx, WebhookEvent{
			EventType:    e.ResourceType + "." + e.Action,
			Action:       e.Action,
			ResourceType: e.ResourceType,
			ResourceID:   e.ResourceID,
			ResourceName: e.ResourceName,
			UserID:       e.UserID,
			Username:     e.Username,
			Status:       e.Status,
			OldValues:    e.OldValues,
			NewValues:    e.NewValues,
			OccurredAt:   time.Now().UTC(),
		})
	}
}

// ListAuditLogs returns audit log entries matching the filter.
func (a *AuditService) ListAuditLogs(ctx context.Context, filter *models.AuditLogFilter) ([]*models.AuditLog, error) {
	return a.svc.repository.ListAuditLogs(ctx, filter)
}

// CountAuditLogs returns the total count matching the filter (for pagination).
func (a *AuditService) CountAuditLogs(ctx context.Context, filter *models.AuditLogFilter) (int64, error) {
	return a.svc.repository.CountAuditLogs(ctx, filter)
}

// PurgeOldLogs deletes audit log entries older than the configured retention period.
// Returns the number of rows deleted.
func (a *AuditService) PurgeOldLogs(ctx context.Context) (int64, error) {
	retentionDays := 90
	if val, err := a.svc.Config.Get("audit_log_retention_days"); err == nil && val != "" {
		if days, err := strconv.Atoi(val); err == nil && days > 0 {
			retentionDays = days
		}
	}
	before := time.Now().AddDate(0, 0, -retentionDays)
	return a.svc.repository.DeleteAuditLogsBefore(ctx, before)
}
