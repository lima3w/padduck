package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"padduck/models"
)

// ─────────────────────────────────────────────────────────────────────────────
// Utilisation history (#220)
// ─────────────────────────────────────────────────────────────────────────────

// RecordUtilisationSnapshot inserts a utilisation data point for a subnet.
func (r *Repository) RecordUtilisationSnapshot(ctx context.Context, subnetID int64, used, total int, pct float64) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO subnet_utilisation_history (subnet_id, used_count, total_count, utilisation_pct)
		 VALUES ($1, $2, $3, $4)`,
		subnetID, used, total, pct,
	)
	return err
}

// GetUtilisationHistory returns ordered utilisation data points for a subnet over the last N days.
func (r *Repository) GetUtilisationHistory(ctx context.Context, subnetID int64, days int) ([]*models.SubnetUtilisationPoint, error) {
	rows, err := r.db.Query(ctx,
		`SELECT recorded_at, used_count, total_count, utilisation_pct
		 FROM subnet_utilisation_history
		 WHERE subnet_id = $1 AND recorded_at >= now() - ($2 * INTERVAL '1 day')
		 ORDER BY recorded_at ASC`,
		subnetID, days,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []*models.SubnetUtilisationPoint
	for rows.Next() {
		p := &models.SubnetUtilisationPoint{}
		if err := rows.Scan(&p.RecordedAt, &p.UsedCount, &p.TotalCount, &p.UtilisationPct); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

// GetUtilisationTrends returns current utilisation and the delta vs 7 days ago for all subnets.
func (r *Repository) GetUtilisationTrends(ctx context.Context) ([]*models.SubnetUtilisationTrend, error) {
	rows, err := r.db.Query(ctx, `
		WITH latest AS (
			SELECT DISTINCT ON (subnet_id)
				subnet_id, utilisation_pct AS current_pct
			FROM subnet_utilisation_history
			ORDER BY subnet_id, recorded_at DESC
		),
		week_ago AS (
			SELECT DISTINCT ON (subnet_id)
				subnet_id, utilisation_pct AS week_ago_pct
			FROM subnet_utilisation_history
			WHERE recorded_at BETWEEN now() - interval '8 days' AND now() - interval '6 days'
			ORDER BY subnet_id, recorded_at DESC
		)
		SELECT s.id,
		       host(s.network_address) || '/' || s.prefix_length AS cidr,
		       s.description,
		       COALESCE(l.current_pct, 0) AS current_pct,
		       COALESCE(w.week_ago_pct, 0) AS week_ago_pct,
		       COALESCE(l.current_pct, 0) - COALESCE(w.week_ago_pct, 0) AS delta_pct
		FROM subnets s
		LEFT JOIN latest l ON l.subnet_id = s.id
		LEFT JOIN week_ago w ON w.subnet_id = s.id
		ORDER BY delta_pct DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trends []*models.SubnetUtilisationTrend
	for rows.Next() {
		t := &models.SubnetUtilisationTrend{}
		if err := rows.Scan(&t.SubnetID, &t.CIDR, &t.Description, &t.CurrentPct, &t.WeekAgoPct, &t.DeltaPct); err != nil {
			return nil, err
		}
		trends = append(trends, t)
	}
	return trends, rows.Err()
}

// ─────────────────────────────────────────────────────────────────────────────
// Alert cooldowns (#221)
// ─────────────────────────────────────────────────────────────────────────────

// GetAlertCooldown returns the cooldown record for a subnet, or nil if none exists.
func (r *Repository) GetAlertCooldown(ctx context.Context, subnetID int64) (*models.AlertCooldown, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, subnet_id, alerted_at, alerted_pct FROM alert_cooldowns WHERE subnet_id = $1`,
		subnetID,
	)
	cd := &models.AlertCooldown{}
	err := row.Scan(&cd.ID, &cd.SubnetID, &cd.AlertedAt, &cd.AlertedPct)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return cd, nil
}

// SetAlertCooldown upserts a cooldown record for a subnet.
func (r *Repository) SetAlertCooldown(ctx context.Context, subnetID int64, pct float64) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO alert_cooldowns (subnet_id, alerted_at, alerted_pct)
		 VALUES ($1, now(), $2)
		 ON CONFLICT (subnet_id) DO UPDATE SET alerted_at = now(), alerted_pct = EXCLUDED.alerted_pct`,
		subnetID, pct,
	)
	return err
}

// ClearAlertCooldown removes the cooldown record for a subnet.
func (r *Repository) ClearAlertCooldown(ctx context.Context, subnetID int64) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM alert_cooldowns WHERE subnet_id = $1`,
		subnetID,
	)
	return err
}

// ListSubnetsWithThresholds returns subnets that have alert_threshold_pct set.
func (r *Repository) ListSubnetsWithThresholds(ctx context.Context) ([]*models.Subnet, error) {
	query := `SELECT ` + subnetSelectCols + ` ` + subnetFromJoin + `
	          WHERE s.alert_threshold_pct IS NOT NULL
	          ORDER BY s.network_address`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subnets []*models.Subnet
	for rows.Next() {
		subnet, err := scanSubnet(rows)
		if err != nil {
			return nil, err
		}
		subnets = append(subnets, subnet)
	}
	return subnets, rows.Err()
}

// ListAllSubnets returns every subnet across all sections.
func (r *Repository) ListAllSubnets(ctx context.Context) ([]*models.Subnet, error) {
	query := `SELECT ` + subnetSelectCols + ` ` + subnetFromJoin + ` ORDER BY s.network_address`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subnets []*models.Subnet
	for rows.Next() {
		subnet, err := scanSubnet(rows)
		if err != nil {
			return nil, err
		}
		subnets = append(subnets, subnet)
	}
	return subnets, rows.Err()
}

// GetLatestUtilisationForSubnet returns the most recent utilisation record for a subnet.
func (r *Repository) GetLatestUtilisationForSubnet(ctx context.Context, subnetID int64) (*models.SubnetUtilisationPoint, error) {
	row := r.db.QueryRow(ctx,
		`SELECT recorded_at, used_count, total_count, utilisation_pct
		 FROM subnet_utilisation_history
		 WHERE subnet_id = $1
		 ORDER BY recorded_at DESC
		 LIMIT 1`,
		subnetID,
	)
	p := &models.SubnetUtilisationPoint{}
	err := row.Scan(&p.RecordedAt, &p.UsedCount, &p.TotalCount, &p.UtilisationPct)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

// GetSubnetsByUtilisationThreshold returns subnets whose latest utilisation exceeds the given percentage.
func (r *Repository) GetSubnetsByUtilisationThreshold(ctx context.Context, thresholdPct float64) ([]*models.SubnetUtilisationTrend, error) {
	rows, err := r.db.Query(ctx, `
		WITH latest AS (
			SELECT DISTINCT ON (subnet_id)
				subnet_id, utilisation_pct
			FROM subnet_utilisation_history
			ORDER BY subnet_id, recorded_at DESC
		)
		SELECT s.id,
		       host(s.network_address) || '/' || s.prefix_length,
		       s.description,
		       l.utilisation_pct,
		       0::numeric,
		       0::numeric
		FROM latest l
		JOIN subnets s ON s.id = l.subnet_id
		WHERE l.utilisation_pct > $1
		ORDER BY l.utilisation_pct DESC
	`, thresholdPct)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*models.SubnetUtilisationTrend
	for rows.Next() {
		t := &models.SubnetUtilisationTrend{}
		if err := rows.Scan(&t.SubnetID, &t.CIDR, &t.Description, &t.CurrentPct, &t.WeekAgoPct, &t.DeltaPct); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
