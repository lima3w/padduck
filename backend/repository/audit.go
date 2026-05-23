package repository

import (
	"context"
	"fmt"
	"time"

	"padduck/models"
)

// Audit log operations

func (r *Repository) CreateAuditLog(ctx context.Context, entry *models.AuditLog) error {
	query := `INSERT INTO audit_logs
		(user_id, username, action, resource_type, resource_id, resource_name, old_values, new_values, ip_address, user_agent, status, error_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.db.Exec(ctx, query,
		entry.UserID, entry.Username, entry.Action,
		nullableString(entry.ResourceType), entry.ResourceID, nullableString(entry.ResourceName),
		entry.OldValues, entry.NewValues,
		nullableString(entry.IPAddress), nullableString(entry.UserAgent),
		entry.Status, nullableString(entry.ErrorMessage),
	)
	return err
}

func (r *Repository) ListAuditLogs(ctx context.Context, filter *models.AuditLogFilter) ([]*models.AuditLog, error) {
	args := []interface{}{}
	where := []string{}
	i := 1

	if filter.UserID != nil {
		where = append(where, fmt.Sprintf("user_id = $%d", i))
		args = append(args, *filter.UserID)
		i++
	}
	if filter.Username != "" {
		where = append(where, fmt.Sprintf("username ILIKE $%d", i))
		args = append(args, "%"+filter.Username+"%")
		i++
	}
	if filter.Action != "" {
		where = append(where, fmt.Sprintf("action = $%d", i))
		args = append(args, filter.Action)
		i++
	}
	if filter.ResourceType != "" {
		where = append(where, fmt.Sprintf("resource_type = $%d", i))
		args = append(args, filter.ResourceType)
		i++
	}
	if filter.ResourceID != nil {
		where = append(where, fmt.Sprintf("resource_id = $%d", i))
		args = append(args, *filter.ResourceID)
		i++
	}
	if filter.IPAddress != "" {
		where = append(where, fmt.Sprintf("ip_address = $%d", i))
		args = append(args, filter.IPAddress)
		i++
	}
	if filter.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", i))
		args = append(args, filter.Status)
		i++
	}
	if filter.Since != nil {
		where = append(where, fmt.Sprintf("created_at >= $%d", i))
		args = append(args, *filter.Since)
		i++
	}
	if filter.Until != nil {
		where = append(where, fmt.Sprintf("created_at <= $%d", i))
		args = append(args, *filter.Until)
		i++
	}

	query := `SELECT id, user_id, username, action, resource_type, resource_id, resource_name,
		old_values, new_values, ip_address, user_agent, status, error_message, created_at
		FROM audit_logs`
	if len(where) > 0 {
		query += " WHERE " + joinStrings(where, " AND ")
	}
	query += " ORDER BY created_at DESC"

	limit := filter.Limit
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, limit, filter.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]*models.AuditLog, 0)
	for rows.Next() {
		l := &models.AuditLog{}
		err := rows.Scan(
			&l.ID, &l.UserID, &l.Username, &l.Action,
			scanNullString(&l.ResourceType), &l.ResourceID, scanNullString(&l.ResourceName),
			&l.OldValues, &l.NewValues,
			scanNullString(&l.IPAddress), scanNullString(&l.UserAgent),
			&l.Status, scanNullString(&l.ErrorMessage), &l.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func (r *Repository) CountAuditLogs(ctx context.Context, filter *models.AuditLogFilter) (int64, error) {
	args := []interface{}{}
	where := []string{}
	i := 1

	if filter.UserID != nil {
		where = append(where, fmt.Sprintf("user_id = $%d", i))
		args = append(args, *filter.UserID)
		i++
	}
	if filter.Action != "" {
		where = append(where, fmt.Sprintf("action = $%d", i))
		args = append(args, filter.Action)
		i++
	}
	if filter.ResourceType != "" {
		where = append(where, fmt.Sprintf("resource_type = $%d", i))
		args = append(args, filter.ResourceType)
		i++
	}
	if filter.ResourceID != nil {
		where = append(where, fmt.Sprintf("resource_id = $%d", i))
		args = append(args, *filter.ResourceID)
		i++
	}
	if filter.Since != nil {
		where = append(where, fmt.Sprintf("created_at >= $%d", i))
		args = append(args, *filter.Since)
		i++
	}
	if filter.Until != nil {
		where = append(where, fmt.Sprintf("created_at <= $%d", i))
		args = append(args, *filter.Until)
		i++
	}

	query := `SELECT COUNT(*) FROM audit_logs`
	if len(where) > 0 {
		query += " WHERE " + joinStrings(where, " AND ")
	}

	var count int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

func (r *Repository) DeleteAuditLogsBefore(ctx context.Context, before time.Time) (int64, error) {
	query := `DELETE FROM audit_logs WHERE created_at < $1`
	result, err := r.db.Exec(ctx, query, before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// GetAuditRetentionSettings returns the single retention settings row, creating
// it with sensible defaults (365 days, archive disabled) if it doesn't exist yet.
func (r *Repository) GetAuditRetentionSettings(ctx context.Context) (*models.AuditRetentionSettings, error) {
	s := &models.AuditRetentionSettings{}
	err := r.db.QueryRow(ctx,
		`SELECT id, retention_days, archive_enabled, updated_at FROM audit_retention_settings LIMIT 1`,
	).Scan(&s.ID, &s.RetentionDays, &s.ArchiveEnabled, &s.UpdatedAt)
	if err == nil {
		return s, nil
	}
	// No row exists — seed defaults and return them.
	err = r.db.QueryRow(ctx,
		`INSERT INTO audit_retention_settings (retention_days, archive_enabled)
		 VALUES (365, false)
		 ON CONFLICT DO NOTHING
		 RETURNING id, retention_days, archive_enabled, updated_at`,
	).Scan(&s.ID, &s.RetentionDays, &s.ArchiveEnabled, &s.UpdatedAt)
	return s, err
}

// UpdateAuditRetentionSettings updates the single retention settings row.
func (r *Repository) UpdateAuditRetentionSettings(ctx context.Context, retentionDays int, archiveEnabled bool) (*models.AuditRetentionSettings, error) {
	s := &models.AuditRetentionSettings{}
	err := r.db.QueryRow(ctx,
		`UPDATE audit_retention_settings SET retention_days = $1, archive_enabled = $2, updated_at = now()
		 WHERE id = (SELECT id FROM audit_retention_settings LIMIT 1)
		 RETURNING id, retention_days, archive_enabled, updated_at`,
		retentionDays, archiveEnabled,
	).Scan(&s.ID, &s.RetentionDays, &s.ArchiveEnabled, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// PruneAuditLogs deletes audit_logs entries older than retentionDays. Returns count deleted.
func (r *Repository) PruneAuditLogs(ctx context.Context, retentionDays int) (int64, error) {
	before := time.Now().AddDate(0, 0, -retentionDays)
	result, err := r.db.Exec(ctx, `DELETE FROM audit_logs WHERE created_at < $1`, before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// scanNullString returns a pointer that scan can write into; empty DB nulls become ""
func scanNullString(dest *string) *nullStringScanner {
	return &nullStringScanner{dest: dest}
}

type nullStringScanner struct{ dest *string }

func (n *nullStringScanner) Scan(src interface{}) error {
	if src == nil {
		*n.dest = ""
		return nil
	}
	switch v := src.(type) {
	case string:
		*n.dest = v
	case []byte:
		*n.dest = string(v)
	default:
		*n.dest = fmt.Sprintf("%v", v)
	}
	return nil
}

func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
