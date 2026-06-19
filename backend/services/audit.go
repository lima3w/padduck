package services

import (
	"context"
	"encoding/json"
	"log/slog"
	"reflect"
	"strconv"
	"strings"
	"time"

	"padduck/models"
)

type auditRepo interface {
	CreateAuditLog(ctx context.Context, entry *models.AuditLog) error
	ListAuditLogs(ctx context.Context, filter *models.AuditLogFilter) ([]*models.AuditLog, error)
	CountAuditLogs(ctx context.Context, filter *models.AuditLogFilter) (int64, error)
	DeleteAuditLogsBefore(ctx context.Context, before time.Time) (int64, error)
	GetAuditRetentionSettings(ctx context.Context) (*models.AuditRetentionSettings, error)
	UpdateAuditRetentionSettings(ctx context.Context, retentionDays int, archiveEnabled bool) (*models.AuditRetentionSettings, error)
	PruneAuditLogs(ctx context.Context, retentionDays int) (int64, error)
}

// AuditService handles writing and querying audit logs.
type AuditService struct {
	repo     auditRepo
	config   *ConfigService
	webhooks *WebhookService
}

func NewAuditService(repo auditRepo, config *ConfigService, webhooks *WebhookService) *AuditService {
	return &AuditService{repo: repo, config: config, webhooks: webhooks}
}

// SubscribeTo registers the AuditService as a subscriber on bus.
// Any AuditableEvent published on the bus is forwarded to Log().
func (a *AuditService) SubscribeTo(bus *EventBus) {
	bus.Subscribe("*", func(ctx context.Context, e Event) {
		if ae, ok := e.(AuditableEvent); ok {
			a.Log(ctx, ae.ToAuditEntry())
		}
	})
}

// AuditEntry is the input to Log().
type AuditEntry struct {
	OrgID        *int64
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
		OrganizationID: e.OrgID,
		UserID:         e.UserID,
		Username:       e.Username,
		Action:         e.Action,
		ResourceType:   e.ResourceType,
		ResourceID:     e.ResourceID,
		ResourceName:   e.ResourceName,
		IPAddress:      e.IPAddress,
		UserAgent:      e.UserAgent,
		Status:         e.Status,
		ErrorMessage:   e.ErrorMessage,
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

	if err := a.repo.CreateAuditLog(ctx, entry); err != nil {
		slog.Error("audit: failed to write log", "action", e.Action, "error", err)
	}
	if a.webhooks != nil {
		a.webhooks.Queue(ctx, WebhookEvent{
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
	return a.repo.ListAuditLogs(ctx, filter)
}

// CountAuditLogs returns the total count matching the filter (for pagination).
func (a *AuditService) CountAuditLogs(ctx context.Context, filter *models.AuditLogFilter) (int64, error) {
	return a.repo.CountAuditLogs(ctx, filter)
}

// PurgeOldLogs deletes audit log entries older than the configured retention period.
// Returns the number of rows deleted.
func (a *AuditService) PurgeOldLogs(ctx context.Context) (int64, error) {
	retentionDays := 90
	if val, err := a.config.GetCtx(ctx, "audit_log_retention_days"); err == nil && val != "" {
		if days, err := strconv.Atoi(val); err == nil && days > 0 {
			retentionDays = days
		}
	}
	before := time.Now().UTC().AddDate(0, 0, -retentionDays)
	return a.repo.DeleteAuditLogsBefore(ctx, before)
}

// GetRetentionSettings returns the audit retention settings row.
func (a *AuditService) GetRetentionSettings(ctx context.Context) (*models.AuditRetentionSettings, error) {
	return a.repo.GetAuditRetentionSettings(ctx)
}

// UpdateRetentionSettings updates the audit retention settings row.
func (a *AuditService) UpdateRetentionSettings(ctx context.Context, retentionDays int, archiveEnabled bool) (*models.AuditRetentionSettings, error) {
	return a.repo.UpdateAuditRetentionSettings(ctx, retentionDays, archiveEnabled)
}

// PruneByRetentionSettings prunes audit logs using the retention_days from the settings table.
func (a *AuditService) PruneByRetentionSettings(ctx context.Context) (int64, error) {
	s, err := a.repo.GetAuditRetentionSettings(ctx)
	if err != nil {
		// Fall back to 365 days if settings unavailable
		return a.repo.PruneAuditLogs(ctx, 365)
	}
	return a.repo.PruneAuditLogs(ctx, s.RetentionDays)
}
