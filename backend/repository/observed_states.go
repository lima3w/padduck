package repository

import (
	"context"
	"encoding/json"
	"time"

	"padduck/models"
)

// UpsertObservedState inserts or updates the observed state for a resource.
// For registered resources (resource_id IS NOT NULL) the unique key is (resource_type, resource_id).
// For unregistered IPs (resource_id IS NULL) the unique key is ip_address.
func (r *Repository) UpsertObservedState(ctx context.Context, s *models.ObservedState) error {
	dataJSON, err := json.Marshal(s.ObservedData)
	if err != nil {
		return err
	}

	if s.ResourceID != nil {
		_, err = r.db.Exec(ctx, `
			INSERT INTO observed_states
			  (organization_id, resource_type, resource_id, ip_address, observed_data, source, scan_result_id, last_seen_at)
			VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7, NOW())
			ON CONFLICT (resource_type, resource_id) WHERE resource_id IS NOT NULL
			DO UPDATE SET
			  observed_data  = EXCLUDED.observed_data,
			  source         = EXCLUDED.source,
			  scan_result_id = EXCLUDED.scan_result_id,
			  last_seen_at   = NOW()`,
			s.OrganizationID, s.ResourceType, s.ResourceID, s.IPAddress,
			string(dataJSON), s.Source, s.ScanResultID)
	} else {
		_, err = r.db.Exec(ctx, `
			INSERT INTO observed_states
			  (organization_id, resource_type, ip_address, observed_data, source, scan_result_id, last_seen_at)
			VALUES ($1, $2, $3, $4::jsonb, $5, $6, NOW())
			ON CONFLICT (ip_address) WHERE resource_id IS NULL AND ip_address IS NOT NULL
			DO UPDATE SET
			  observed_data  = EXCLUDED.observed_data,
			  source         = EXCLUDED.source,
			  scan_result_id = EXCLUDED.scan_result_id,
			  last_seen_at   = NOW()`,
			s.OrganizationID, s.ResourceType, s.IPAddress,
			string(dataJSON), s.Source, s.ScanResultID)
	}
	return err
}

// GetObservedState returns the observed state for a specific registered resource.
func (r *Repository) GetObservedState(ctx context.Context, resourceType string, resourceID int64) (*models.ObservedState, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, organization_id, resource_type, resource_id, ip_address,
		       observed_data::text, source, scan_result_id, first_seen_at, last_seen_at
		FROM observed_states
		WHERE resource_type = $1 AND resource_id = $2`,
		resourceType, resourceID)

	s := &models.ObservedState{}
	var rawData string
	var lastSeen time.Time
	err := row.Scan(&s.ID, &s.OrganizationID, &s.ResourceType, &s.ResourceID, &s.IPAddress,
		&rawData, &s.Source, &s.ScanResultID, &s.FirstSeenAt, &lastSeen)
	if err != nil {
		return nil, err
	}
	s.LastSeenAt = lastSeen
	_ = json.Unmarshal([]byte(rawData), &s.ObservedData)
	return s, nil
}

// ListUnregisteredHosts returns observed states for IPs seen by scanner but not matched
// to any authoritative ip_addresses record (resource_id IS NULL).
func (r *Repository) ListUnregisteredHosts(ctx context.Context, orgID *int64) ([]*models.ObservedState, error) {
	query := `
		SELECT id, organization_id, resource_type, resource_id, ip_address,
		       observed_data::text, source, scan_result_id, first_seen_at, last_seen_at
		FROM observed_states
		WHERE resource_id IS NULL AND resource_type = 'ip_address'`
	args := []any{}
	if orgID != nil {
		query += ` AND organization_id = $1`
		args = append(args, *orgID)
	}
	query += ` ORDER BY last_seen_at DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*models.ObservedState
	for rows.Next() {
		s := &models.ObservedState{}
		var rawData string
		if err := rows.Scan(&s.ID, &s.OrganizationID, &s.ResourceType, &s.ResourceID, &s.IPAddress,
			&rawData, &s.Source, &s.ScanResultID, &s.FirstSeenAt, &s.LastSeenAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(rawData), &s.ObservedData)
		out = append(out, s)
	}
	return out, rows.Err()
}
