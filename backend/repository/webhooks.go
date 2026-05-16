package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"ipam-next/models"
)

func (r *Repository) ListWebhookEndpoints(ctx context.Context) ([]*models.WebhookEndpoint, error) {
	query := `SELECT id, name, url, secret, events, is_active, created_by, created_at, updated_at
	          FROM webhook_endpoints
	          ORDER BY name ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	endpoints := make([]*models.WebhookEndpoint, 0)
	for rows.Next() {
		w := &models.WebhookEndpoint{}
		if err := rows.Scan(&w.ID, &w.Name, &w.URL, &w.Secret, &w.Events, &w.IsActive, &w.CreatedBy, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		endpoints = append(endpoints, w)
	}
	return endpoints, rows.Err()
}

func (r *Repository) ListActiveWebhookEndpoints(ctx context.Context) ([]*models.WebhookEndpoint, error) {
	query := `SELECT id, name, url, secret, events, is_active, created_by, created_at, updated_at
	          FROM webhook_endpoints
	          WHERE is_active = TRUE
	          ORDER BY id ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	endpoints := make([]*models.WebhookEndpoint, 0)
	for rows.Next() {
		w := &models.WebhookEndpoint{}
		if err := rows.Scan(&w.ID, &w.Name, &w.URL, &w.Secret, &w.Events, &w.IsActive, &w.CreatedBy, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		endpoints = append(endpoints, w)
	}
	return endpoints, rows.Err()
}

func (r *Repository) CreateWebhookEndpoint(ctx context.Context, endpoint *models.WebhookEndpoint) (*models.WebhookEndpoint, error) {
	query := `INSERT INTO webhook_endpoints (name, url, secret, events, is_active, created_by)
	          VALUES ($1, $2, $3, $4, $5, $6)
	          RETURNING id, name, url, secret, events, is_active, created_by, created_at, updated_at`
	w := &models.WebhookEndpoint{}
	err := r.db.QueryRow(ctx, query, endpoint.Name, endpoint.URL, endpoint.Secret, endpoint.Events, endpoint.IsActive, endpoint.CreatedBy).Scan(
		&w.ID, &w.Name, &w.URL, &w.Secret, &w.Events, &w.IsActive, &w.CreatedBy, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (r *Repository) UpdateWebhookEndpoint(ctx context.Context, endpoint *models.WebhookEndpoint) (*models.WebhookEndpoint, error) {
	query := `UPDATE webhook_endpoints
	          SET name = $2, url = $3, secret = $4, events = $5, is_active = $6, updated_at = NOW()
	          WHERE id = $1
	          RETURNING id, name, url, secret, events, is_active, created_by, created_at, updated_at`
	w := &models.WebhookEndpoint{}
	err := r.db.QueryRow(ctx, query, endpoint.ID, endpoint.Name, endpoint.URL, endpoint.Secret, endpoint.Events, endpoint.IsActive).Scan(
		&w.ID, &w.Name, &w.URL, &w.Secret, &w.Events, &w.IsActive, &w.CreatedBy, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (r *Repository) DeleteWebhookEndpoint(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM webhook_endpoints WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) CreateWebhookDelivery(ctx context.Context, endpointID int64, eventType, payloadJSON string) (*models.WebhookDelivery, error) {
	query := `INSERT INTO webhook_deliveries (endpoint_id, event_type, payload)
	          VALUES ($1, $2, $3::jsonb)
	          RETURNING id, endpoint_id, event_type, payload::text, status, retry_count, next_retry_at,
	                    delivered_at, response_status, error_msg, created_at, updated_at`
	d := &models.WebhookDelivery{}
	err := r.db.QueryRow(ctx, query, endpointID, eventType, payloadJSON).Scan(
		&d.ID, &d.EndpointID, &d.EventType, &d.Payload, &d.Status, &d.RetryCount, &d.NextRetryAt,
		&d.DeliveredAt, &d.ResponseStatus, &d.ErrorMsg, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (r *Repository) GetPendingWebhookDeliveries(ctx context.Context, limit int) ([]*models.WebhookDelivery, error) {
	query := `SELECT id, endpoint_id, event_type, payload::text, status, retry_count, next_retry_at,
	                 delivered_at, response_status, error_msg, created_at, updated_at
	          FROM webhook_deliveries
	          WHERE status IN ('pending', 'retrying')
	            AND (next_retry_at IS NULL OR next_retry_at <= NOW())
	          ORDER BY created_at ASC
	          LIMIT $1`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	deliveries := make([]*models.WebhookDelivery, 0)
	for rows.Next() {
		d := &models.WebhookDelivery{}
		if err := rows.Scan(
			&d.ID, &d.EndpointID, &d.EventType, &d.Payload, &d.Status, &d.RetryCount, &d.NextRetryAt,
			&d.DeliveredAt, &d.ResponseStatus, &d.ErrorMsg, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		deliveries = append(deliveries, d)
	}
	return deliveries, rows.Err()
}

func (r *Repository) ListWebhookDeliveries(ctx context.Context, limit int) ([]*models.WebhookDelivery, error) {
	query := `SELECT id, endpoint_id, event_type, payload::text, status, retry_count, next_retry_at,
	                 delivered_at, response_status, error_msg, created_at, updated_at
	          FROM webhook_deliveries
	          ORDER BY created_at DESC
	          LIMIT $1`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	deliveries := make([]*models.WebhookDelivery, 0)
	for rows.Next() {
		d := &models.WebhookDelivery{}
		if err := rows.Scan(
			&d.ID, &d.EndpointID, &d.EventType, &d.Payload, &d.Status, &d.RetryCount, &d.NextRetryAt,
			&d.DeliveredAt, &d.ResponseStatus, &d.ErrorMsg, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		deliveries = append(deliveries, d)
	}
	return deliveries, rows.Err()
}

func (r *Repository) GetWebhookEndpoint(ctx context.Context, id int64) (*models.WebhookEndpoint, error) {
	query := `SELECT id, name, url, secret, events, is_active, created_by, created_at, updated_at
	          FROM webhook_endpoints WHERE id = $1`
	w := &models.WebhookEndpoint{}
	err := r.db.QueryRow(ctx, query, id).Scan(&w.ID, &w.Name, &w.URL, &w.Secret, &w.Events, &w.IsActive, &w.CreatedBy, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (r *Repository) MarkWebhookDelivered(ctx context.Context, id int64, statusCode int) error {
	query := `UPDATE webhook_deliveries
	          SET status = 'delivered', delivered_at = NOW(), response_status = $2, updated_at = NOW()
	          WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, statusCode)
	return err
}

func (r *Repository) MarkWebhookFailed(ctx context.Context, id int64, errMsg string, retryCount int, nextRetryAt *time.Time, statusCode *int) error {
	status := "failed"
	if nextRetryAt != nil {
		status = "retrying"
	}
	query := `UPDATE webhook_deliveries
	          SET status = $2, error_msg = $3, retry_count = $4, next_retry_at = $5,
	              response_status = $6, updated_at = NOW()
	          WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, status, errMsg, retryCount, nextRetryAt, statusCode)
	return err
}
