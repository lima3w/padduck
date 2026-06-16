package repository

import "context"

// TelemetryCounts holds aggregate counts gathered in a single round-trip for
// the telemetry snapshot.
type TelemetryCounts struct {
	UsersTotal        int64
	ActiveUsers7d     int64
	ActiveUsers30d    int64
	CustomersTotal    int64
	LocationsTotal    int64
	VLANsTotal        int64
	SubnetsTotal      int64
	IPv4SubnetsTotal  int64
	IPv6SubnetsTotal  int64
	IPv4Subnets29to32 int64
	IPv4Subnets25to28 int64
	IPv4Subnets24     int64
	IPv4Subnets16to23 int64
	IPv4Subnets8to15  int64
	DevicesTotal      int64
}

// GetTelemetryCounts fetches all object counts needed for a telemetry snapshot
// in a single database round-trip.
func (r *Repository) GetTelemetryCounts(ctx context.Context) (*TelemetryCounts, error) {
	const q = `
		SELECT
			(SELECT COUNT(*) FROM users)       AS users_total,
			(SELECT COUNT(DISTINCT user_id) FROM audit_logs
			 WHERE created_at >= NOW() - INTERVAL '7 days' AND user_id IS NOT NULL)  AS active_users_7d,
			(SELECT COUNT(DISTINCT user_id) FROM audit_logs
			 WHERE created_at >= NOW() - INTERVAL '30 days' AND user_id IS NOT NULL) AS active_users_30d,
			(SELECT COUNT(*) FROM customers)   AS customers_total,
			(SELECT COUNT(*) FROM locations)   AS locations_total,
			(SELECT COUNT(*) FROM vlans)       AS vlans_total,
			(SELECT COUNT(*) FROM subnets)     AS subnets_total,
			(SELECT COUNT(*) FROM subnets WHERE family(network_address) = 4)                                     AS ipv4_subnets_total,
			(SELECT COUNT(*) FROM subnets WHERE family(network_address) = 6)                                     AS ipv6_subnets_total,
			(SELECT COUNT(*) FROM subnets WHERE family(network_address) = 4 AND prefix_length BETWEEN 29 AND 32) AS ipv4_29_to_32,
			(SELECT COUNT(*) FROM subnets WHERE family(network_address) = 4 AND prefix_length BETWEEN 25 AND 28) AS ipv4_25_to_28,
			(SELECT COUNT(*) FROM subnets WHERE family(network_address) = 4 AND prefix_length = 24)              AS ipv4_24,
			(SELECT COUNT(*) FROM subnets WHERE family(network_address) = 4 AND prefix_length BETWEEN 16 AND 23) AS ipv4_16_to_23,
			(SELECT COUNT(*) FROM subnets WHERE family(network_address) = 4 AND prefix_length BETWEEN 8  AND 15) AS ipv4_8_to_15,
			(SELECT COUNT(*) FROM devices)     AS devices_total
	`
	c := &TelemetryCounts{}
	err := r.db.QueryRow(ctx, q).Scan(
		&c.UsersTotal,
		&c.ActiveUsers7d,
		&c.ActiveUsers30d,
		&c.CustomersTotal,
		&c.LocationsTotal,
		&c.VLANsTotal,
		&c.SubnetsTotal,
		&c.IPv4SubnetsTotal,
		&c.IPv6SubnetsTotal,
		&c.IPv4Subnets29to32,
		&c.IPv4Subnets25to28,
		&c.IPv4Subnets24,
		&c.IPv4Subnets16to23,
		&c.IPv4Subnets8to15,
		&c.DevicesTotal,
	)
	return c, err
}

// TelemetryUtilizationMetrics holds IPv4 subnet utilization statistics for the
// telemetry snapshot. Percentile fields are pointers because they are NULL when
// no IPv4 subnets exist.
type TelemetryUtilizationMetrics struct {
	AvgPct    *float64
	MedianPct *float64
	P75Pct    *float64
	P90Pct    *float64
	P95Pct    *float64
	Empty     int64
	Over50    int64
	Over80    int64
	Over90    int64
	Full      int64
}

// GetTelemetryUtilizationMetrics computes IPv4 subnet utilization statistics
// (mean, percentiles, threshold bucket counts) in a single query. Utilization
// per subnet = assigned+reserved IPs / theoretical capacity * 100.
func (r *Repository) GetTelemetryUtilizationMetrics(ctx context.Context) (*TelemetryUtilizationMetrics, error) {
	const q = `
		WITH utils AS (
			SELECT
				(COUNT(i.id)::float /
				 GREATEST((POWER(2, 32 - s.prefix_length) - 2)::bigint, 1) * 100) AS pct
			FROM subnets s
			LEFT JOIN ip_addresses i
				ON i.subnet_id = s.id AND i.status != 'available'
			WHERE family(s.network_address) = 4
			GROUP BY s.id, s.prefix_length
		)
		SELECT
			AVG(pct)                                               AS avg_pct,
			percentile_cont(0.50) WITHIN GROUP (ORDER BY pct)     AS median_pct,
			percentile_cont(0.75) WITHIN GROUP (ORDER BY pct)     AS p75_pct,
			percentile_cont(0.90) WITHIN GROUP (ORDER BY pct)     AS p90_pct,
			percentile_cont(0.95) WITHIN GROUP (ORDER BY pct)     AS p95_pct,
			COUNT(*) FILTER (WHERE pct = 0)                        AS subnets_empty,
			COUNT(*) FILTER (WHERE pct >= 50)                      AS subnets_over_50,
			COUNT(*) FILTER (WHERE pct >= 80)                      AS subnets_over_80,
			COUNT(*) FILTER (WHERE pct >= 90)                      AS subnets_over_90,
			COUNT(*) FILTER (WHERE pct >= 100)                     AS subnets_full
		FROM utils
	`
	m := &TelemetryUtilizationMetrics{}
	err := r.db.QueryRow(ctx, q).Scan(
		&m.AvgPct,
		&m.MedianPct,
		&m.P75Pct,
		&m.P90Pct,
		&m.P95Pct,
		&m.Empty,
		&m.Over50,
		&m.Over80,
		&m.Over90,
		&m.Full,
	)
	return m, err
}
