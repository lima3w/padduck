package repository

import "context"

// TelemetryCounts holds aggregate counts gathered in a single round-trip for
// the telemetry snapshot.
type TelemetryCounts struct {
	UsersTotal        int64
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
}

// GetTelemetryCounts fetches all object counts needed for a telemetry snapshot
// in a single database round-trip.
func (r *Repository) GetTelemetryCounts(ctx context.Context) (*TelemetryCounts, error) {
	const q = `
		SELECT
			(SELECT COUNT(*) FROM users)     AS users_total,
			(SELECT COUNT(*) FROM customers) AS customers_total,
			(SELECT COUNT(*) FROM locations) AS locations_total,
			(SELECT COUNT(*) FROM vlans)     AS vlans_total,
			(SELECT COUNT(*) FROM subnets)   AS subnets_total,
			(SELECT COUNT(*) FROM subnets WHERE family(network_address) = 4)                                                  AS ipv4_subnets_total,
			(SELECT COUNT(*) FROM subnets WHERE family(network_address) = 6)                                                  AS ipv6_subnets_total,
			(SELECT COUNT(*) FROM subnets WHERE family(network_address) = 4 AND prefix_length BETWEEN 29 AND 32)              AS ipv4_29_to_32,
			(SELECT COUNT(*) FROM subnets WHERE family(network_address) = 4 AND prefix_length BETWEEN 25 AND 28)              AS ipv4_25_to_28,
			(SELECT COUNT(*) FROM subnets WHERE family(network_address) = 4 AND prefix_length = 24)                           AS ipv4_24,
			(SELECT COUNT(*) FROM subnets WHERE family(network_address) = 4 AND prefix_length BETWEEN 16 AND 23)              AS ipv4_16_to_23,
			(SELECT COUNT(*) FROM subnets WHERE family(network_address) = 4 AND prefix_length BETWEEN 8  AND 15)              AS ipv4_8_to_15
	`
	c := &TelemetryCounts{}
	err := r.db.QueryRow(ctx, q).Scan(
		&c.UsersTotal,
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
	)
	return c, err
}
