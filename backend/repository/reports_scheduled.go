package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"padduck/models"
)

// ─────────────────────────────────────────────────────────────────────────────
// Scheduled reports (#222)
// ─────────────────────────────────────────────────────────────────────────────

// CreateScheduledReport inserts a new scheduled report.
func (r *Repository) CreateScheduledReport(ctx context.Context, orgID *int64, name, reportType, scheduleCron string, recipientEmails []string, filters map[string]any, format string, createdBy int64) (*models.ScheduledReport, error) {
	emailsJSON, err := json.Marshal(recipientEmails)
	if err != nil {
		return nil, fmt.Errorf("marshalling recipient emails: %w", err)
	}
	filtersJSON, err := json.Marshal(filters)
	if err != nil {
		return nil, fmt.Errorf("marshalling filters: %w", err)
	}

	row := r.db.QueryRow(ctx,
		`INSERT INTO scheduled_reports (organization_id, name, report_type, schedule_cron, recipient_emails, filters, format, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, name, report_type, schedule_cron, recipient_emails, filters, format, last_run_at, created_by, created_at, updated_at`,
		orgID, name, reportType, scheduleCron, emailsJSON, filtersJSON, format, createdBy,
	)
	return scanScheduledReport(row)
}

// GetScheduledReportByID returns a scheduled report by primary key.
func (r *Repository) GetScheduledReportByID(ctx context.Context, id int64) (*models.ScheduledReport, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, name, report_type, schedule_cron, recipient_emails, filters, format, last_run_at, created_by, created_at, updated_at
		 FROM scheduled_reports WHERE id = $1`,
		id,
	)
	return scanScheduledReport(row)
}

// ListScheduledReports returns scheduled reports, optionally filtered to a single org.
// Pass nil to return all reports (used by the background scheduler).
func (r *Repository) ListScheduledReports(ctx context.Context, orgID *int64) ([]*models.ScheduledReport, error) {
	query := `SELECT id, name, report_type, schedule_cron, recipient_emails, filters, format, last_run_at, created_by, created_at, updated_at
		 FROM scheduled_reports`
	var args []interface{}
	if orgID != nil {
		query += ` WHERE organization_id = $1`
		args = append(args, *orgID)
	}
	query += ` ORDER BY name`
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*models.ScheduledReport
	for rows.Next() {
		rpt, err := scanScheduledReport(rows)
		if err != nil {
			return nil, err
		}
		reports = append(reports, rpt)
	}
	return reports, rows.Err()
}

// UpdateScheduledReport updates the mutable fields of a scheduled report.
func (r *Repository) UpdateScheduledReport(ctx context.Context, id int64, name, reportType, scheduleCron string, recipientEmails []string, filters map[string]any, format string) (*models.ScheduledReport, error) {
	emailsJSON, err := json.Marshal(recipientEmails)
	if err != nil {
		return nil, fmt.Errorf("marshalling recipient emails: %w", err)
	}
	filtersJSON, err := json.Marshal(filters)
	if err != nil {
		return nil, fmt.Errorf("marshalling filters: %w", err)
	}

	row := r.db.QueryRow(ctx,
		`UPDATE scheduled_reports
		 SET name = $2, report_type = $3, schedule_cron = $4, recipient_emails = $5, filters = $6, format = $7, updated_at = now()
		 WHERE id = $1
		 RETURNING id, name, report_type, schedule_cron, recipient_emails, filters, format, last_run_at, created_by, created_at, updated_at`,
		id, name, reportType, scheduleCron, emailsJSON, filtersJSON, format,
	)
	return scanScheduledReport(row)
}

// UpdateScheduledReportRunTime marks when a scheduled report was last run.
func (r *Repository) UpdateScheduledReportRunTime(ctx context.Context, id int64, t time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE scheduled_reports SET last_run_at = $2, updated_at = now() WHERE id = $1`,
		id, t,
	)
	return err
}

// DeleteScheduledReport removes a scheduled report.
func (r *Repository) DeleteScheduledReport(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM scheduled_reports WHERE id = $1`, id)
	return err
}

// scanScheduledReport scans a single row into a ScheduledReport.
func scanScheduledReport(row interface {
	Scan(dest ...any) error
}) (*models.ScheduledReport, error) {
	rpt := &models.ScheduledReport{}
	var emailsJSON, filtersJSON []byte
	err := row.Scan(
		&rpt.ID, &rpt.Name, &rpt.ReportType, &rpt.ScheduleCron,
		&emailsJSON, &filtersJSON,
		&rpt.Format, &rpt.LastRunAt, &rpt.CreatedBy,
		&rpt.CreatedAt, &rpt.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(emailsJSON, &rpt.RecipientEmails); err != nil {
		return nil, fmt.Errorf("unmarshalling recipient_emails: %w", err)
	}
	if err := json.Unmarshal(filtersJSON, &rpt.Filters); err != nil {
		return nil, fmt.Errorf("unmarshalling filters: %w", err)
	}
	return rpt, nil
}
