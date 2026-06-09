package repository

import (
	"context"
)

// GrafanaSubnetRow holds per-subnet utilisation data for the Grafana datasource.
type GrafanaSubnetRow struct {
	CIDR           string
	NetworkName    string
	Description    string
	Used           int64
	Total          int64
	UtilisationPct float64
}

// GrafanaIPStatusRow holds an IP count for a single status value.
type GrafanaIPStatusRow struct {
	Status string
	Count  int64
}

// GrafanaSectionRow holds aggregate counts for a single section.
type GrafanaSectionRow struct {
	NetworkName string
	SubnetCount int64
	IPCount     int64
	UsedIPs     int64
}

func (r *Repository) GrafanaGetSubnetUtilisation(ctx context.Context) ([]GrafanaSubnetRow, error) {
	query := `
		SELECT
			host(s.network_address) || '/' || s.prefix_length AS cidr,
			sec.name AS section_name,
			s.description,
			COUNT(CASE WHEN ip.status = 'assigned' THEN 1 END) AS used,
			COUNT(ip.id) AS total
		FROM subnets s
		JOIN networks sec ON sec.id = s.network_id
		LEFT JOIN ip_addresses ip ON ip.subnet_id = s.id
		GROUP BY s.id, s.network_address, s.prefix_length, s.description, sec.name
		ORDER BY sec.name ASC, s.network_address ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []GrafanaSubnetRow
	for rows.Next() {
		var row GrafanaSubnetRow
		if err := rows.Scan(&row.CIDR, &row.NetworkName, &row.Description, &row.Used, &row.Total); err != nil {
			return nil, err
		}
		if row.Total > 0 {
			row.UtilisationPct = float64(row.Used) / float64(row.Total) * 100
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (r *Repository) GrafanaGetIPCountsByStatus(ctx context.Context) ([]GrafanaIPStatusRow, error) {
	query := `SELECT status, COUNT(*) FROM ip_addresses GROUP BY status ORDER BY status ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []GrafanaIPStatusRow
	for rows.Next() {
		var row GrafanaIPStatusRow
		if err := rows.Scan(&row.Status, &row.Count); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (r *Repository) GrafanaGetSectionSummary(ctx context.Context) ([]GrafanaSectionRow, error) {
	query := `
		SELECT
			sec.name,
			COUNT(DISTINCT s.id) AS subnet_count,
			COUNT(ip.id) AS ip_count,
			COUNT(CASE WHEN ip.status = 'assigned' THEN 1 END) AS used_ips
		FROM networks sec
		LEFT JOIN subnets s ON s.network_id = sec.id
		LEFT JOIN ip_addresses ip ON ip.subnet_id = s.id
		GROUP BY sec.id, sec.name
		ORDER BY sec.name ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []GrafanaSectionRow
	for rows.Next() {
		var row GrafanaSectionRow
		if err := rows.Scan(&row.NetworkName, &row.SubnetCount, &row.IPCount, &row.UsedIPs); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}
