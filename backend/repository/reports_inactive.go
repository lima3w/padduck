package repository

import (
	"context"

	"padduck/models"
)

// ─────────────────────────────────────────────────────────────────────────────
// Inactive IP reclamation (#224)
// ─────────────────────────────────────────────────────────────────────────────

// GetInactiveIPs returns assigned IPs that have not been seen within the given number of days.
// It excludes reserved IPs and gateway addresses.
func (r *Repository) GetInactiveIPs(ctx context.Context, days int, networkID *int64) ([]*models.InactiveIPReport, error) {
	query := `
		SELECT
			ip.id,
			host(ip.address),
			ip.hostname,
			host(s.network_address) || '/' || s.prefix_length AS subnet_cidr,
			sec.name AS section_name,
			ip.device_id,
			ip.last_seen,
			CASE
				WHEN ip.last_seen IS NULL THEN $1
				ELSE GREATEST(0, EXTRACT(DAY FROM now() - ip.last_seen)::int)
			END AS days_inactive
		FROM ip_addresses ip
		JOIN subnets s ON s.id = ip.subnet_id
		JOIN networks sec ON sec.id = s.network_id
		WHERE ip.status = 'assigned'
		  AND ip.device_id IS NOT NULL
		  AND (ip.last_seen IS NULL OR ip.last_seen < now() - ($1 * INTERVAL '1 day'))
		  AND host(ip.address) != s.gateway
		  AND ($2::bigint IS NULL OR sec.id = $2)
		ORDER BY ip.last_seen ASC NULLS FIRST
	`
	rows, err := r.db.Query(ctx, query, days, networkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.InactiveIPReport
	for rows.Next() {
		rec := &models.InactiveIPReport{}
		if err := rows.Scan(
			&rec.IPID, &rec.IPAddress, &rec.Hostname,
			&rec.SubnetCIDR, &rec.NetworkName,
			&rec.DeviceID, &rec.LastSeen, &rec.DaysInactive,
		); err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	return results, rows.Err()
}

// BulkReleaseIPs sets multiple IPs to 'available' and clears assignment fields.
// Returns the count of rows actually updated.
func (r *Repository) BulkReleaseIPs(ctx context.Context, ipIDs []int64) (int64, error) {
	tag, err := r.db.Exec(ctx,
		`UPDATE ip_addresses
		 SET status = 'available', device_id = NULL, updated_at = now()
		 WHERE id = ANY($1) AND status = 'assigned'`,
		ipIDs,
	)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
