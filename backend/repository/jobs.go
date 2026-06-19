package repository

import (
	"context"
	"encoding/json"
	"time"
)

type JobRecord struct {
	ID           int64
	Kind         string
	Name         string
	Status       string
	Progress     int
	Diagnostic   string
	ErrorMessage string
	Payload      map[string]interface{}
	Result       map[string]interface{}
	Attempts     int
	MaxAttempts  int
	CreatedAt    time.Time
	StartedAt    *time.Time
	FinishedAt   *time.Time
}

func (r *Repository) InsertJob(ctx context.Context, kind, name string, payload map[string]interface{}, maxAttempts int) (int64, error) {
	var payloadJSON []byte
	if payload != nil {
		var err error
		payloadJSON, err = json.Marshal(payload)
		if err != nil {
			return 0, err
		}
	}
	var id int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO background_jobs (type, name, payload, max_attempts)
		VALUES ($1, $2, $3::jsonb, $4)
		RETURNING id`,
		kind, name, payloadJSON, maxAttempts,
	).Scan(&id)
	return id, err
}

func (r *Repository) MarkJobRunning(ctx context.Context, id int64, attempts int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE background_jobs
		SET status = 'running', attempts = $2, started_at = NOW()
		WHERE id = $1`,
		id, attempts,
	)
	return err
}

func (r *Repository) MarkJobFinished(ctx context.Context, id int64, status, errMsg, diagnostic string, progress int, result []byte) error {
	_, err := r.db.Exec(ctx, `
		UPDATE background_jobs
		SET status = $2, error_message = $3, diagnostic = $4, progress = $5, result = $6::jsonb, finished_at = NOW()
		WHERE id = $1`,
		id, status, errMsg, diagnostic, progress, result,
	)
	return err
}

func (r *Repository) MarkJobCanceled(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE background_jobs
		SET status = 'canceled', finished_at = NOW()
		WHERE id = $1`,
		id,
	)
	return err
}

func (r *Repository) RecoverStaleJobs(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `
		UPDATE background_jobs
		SET status = 'failed',
		    error_message = 'process restarted while job was running',
		    finished_at = NOW()
		WHERE status = 'running'`,
	)
	return err
}

func (r *Repository) GetJobRecord(ctx context.Context, id int64) (*JobRecord, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, type, name, status, progress, diagnostic, error_message,
		       payload::text, result::text, attempts, max_attempts,
		       created_at, started_at, finished_at
		FROM background_jobs
		WHERE id = $1`,
		id,
	)
	return scanJobRecord(row)
}

func (r *Repository) ListJobRecords(ctx context.Context, limit int) ([]*JobRecord, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, type, name, status, progress, diagnostic, error_message,
		       payload::text, result::text, attempts, max_attempts,
		       created_at, started_at, finished_at
		FROM background_jobs
		ORDER BY created_at DESC
		LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []*JobRecord
	for rows.Next() {
		rec, err := scanJobRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

type scannable interface {
	Scan(dest ...any) error
}

func scanJobRecord(s scannable) (*JobRecord, error) {
	rec := &JobRecord{}
	var payloadText, resultText *string
	if err := s.Scan(
		&rec.ID, &rec.Kind, &rec.Name, &rec.Status,
		&rec.Progress, &rec.Diagnostic, &rec.ErrorMessage,
		&payloadText, &resultText,
		&rec.Attempts, &rec.MaxAttempts,
		&rec.CreatedAt, &rec.StartedAt, &rec.FinishedAt,
	); err != nil {
		return nil, err
	}
	if payloadText != nil && *payloadText != "" {
		_ = json.Unmarshal([]byte(*payloadText), &rec.Payload)
	}
	if resultText != nil && *resultText != "" {
		_ = json.Unmarshal([]byte(*resultText), &rec.Result)
	}
	return rec, nil
}
