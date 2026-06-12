package services

import (
	"context"
	"encoding/json"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"padduck/models"
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
		b, err := json.Marshal(redactSensitiveValue(e.OldValues))
		if err == nil {
			s := string(b)
			entry.OldValues = &s
		}
	}
	if e.NewValues != nil {
		b, err := json.Marshal(redactSensitiveValue(e.NewValues))
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
			OldValues:    redactSensitiveValue(e.OldValues),
			NewValues:    redactSensitiveValue(e.NewValues),
			OccurredAt:   time.Now().UTC(),
		})
	}
}

const redactedValue = "***REDACTED***"

var sensitiveFieldFragments = []string{
	"snmp_community",
	"community",
	"password",
	"pass",
	"secret",
	"token",
	"api_key",
	"apikey",
	"private_key",
	"certificate",
}

func isSensitiveField(name string) bool {
	name = strings.ToLower(name)
	for _, fragment := range sensitiveFieldFragments {
		if strings.Contains(name, fragment) {
			return true
		}
	}
	return false
}

func redactSensitiveValue(value interface{}) interface{} {
	if value == nil {
		return nil
	}
	switch typed := value.(type) {
	case map[string]string:
		out := make(map[string]string, len(typed))
		for k, v := range typed {
			if isSensitiveField(k) {
				out[k] = redactedValue
			} else {
				out[k] = v
			}
		}
		return out
	case map[string]interface{}:
		out := make(map[string]interface{}, len(typed))
		for k, v := range typed {
			if isSensitiveField(k) {
				out[k] = redactedValue
			} else {
				out[k] = redactSensitiveValue(v)
			}
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(typed))
		for i, v := range typed {
			out[i] = redactSensitiveValue(v)
		}
		return out
	}

	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Pointer && !rv.IsNil() {
		return redactSensitiveValue(rv.Elem().Interface())
	}
	return value
}

func RedactSensitiveJSON(value *string) *string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return value
	}
	var decoded interface{}
	if err := json.Unmarshal([]byte(*value), &decoded); err != nil {
		return value
	}
	redacted, err := json.Marshal(redactSensitiveValue(decoded))
	if err != nil {
		return value
	}
	out := string(redacted)
	return &out
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
	if val, err := a.svc.Config.GetCtx(ctx, "audit_log_retention_days"); err == nil && val != "" {
		if days, err := strconv.Atoi(val); err == nil && days > 0 {
			retentionDays = days
		}
	}
	before := time.Now().UTC().AddDate(0, 0, -retentionDays)
	return a.svc.repository.DeleteAuditLogsBefore(ctx, before)
}

// GetRetentionSettings returns the audit retention settings row.
func (a *AuditService) GetRetentionSettings(ctx context.Context) (*models.AuditRetentionSettings, error) {
	return a.svc.repository.GetAuditRetentionSettings(ctx)
}

// UpdateRetentionSettings updates the audit retention settings row.
func (a *AuditService) UpdateRetentionSettings(ctx context.Context, retentionDays int, archiveEnabled bool) (*models.AuditRetentionSettings, error) {
	return a.svc.repository.UpdateAuditRetentionSettings(ctx, retentionDays, archiveEnabled)
}

// PruneByRetentionSettings prunes audit logs using the retention_days from the settings table.
func (a *AuditService) PruneByRetentionSettings(ctx context.Context) (int64, error) {
	s, err := a.svc.repository.GetAuditRetentionSettings(ctx)
	if err != nil {
		// Fall back to 365 days if settings unavailable
		return a.svc.repository.PruneAuditLogs(ctx, 365)
	}
	return a.svc.repository.PruneAuditLogs(ctx, s.RetentionDays)
}
