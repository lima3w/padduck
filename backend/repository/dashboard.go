package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"padduck/models"
)

// Dashboard operations

// GetDashboardSummary returns aggregate IPAM counts and top utilised subnets.
func (r *Repository) GetDashboardSummary(ctx context.Context) (*models.DashboardSummary, error) {
	summary := &models.DashboardSummary{}

	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM networks`).Scan(&summary.TotalNetworks); err != nil {
		return nil, fmt.Errorf("count sections: %w", err)
	}
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM subnets`).Scan(&summary.TotalSubnets); err != nil {
		return nil, fmt.Errorf("count subnets: %w", err)
	}
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM ip_addresses`).Scan(&summary.TotalIPs); err != nil {
		return nil, fmt.Errorf("count ips: %w", err)
	}
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM ip_addresses WHERE status = 'assigned'`).Scan(&summary.UsedIPs); err != nil {
		return nil, fmt.Errorf("count used ips: %w", err)
	}
	if summary.TotalIPs > 0 {
		summary.UtilisationPct = float64(summary.UsedIPs) / float64(summary.TotalIPs) * 100
	}

	topQuery := `
		SELECT
			s.id,
			host(s.network_address) || '/' || s.prefix_length AS cidr,
			s.description,
			COUNT(CASE WHEN ip.status = 'assigned' THEN 1 END) AS used,
			GREATEST((POWER(2, 32 - s.prefix_length) - 2)::bigint, 1) AS total
		FROM subnets s
		LEFT JOIN ip_addresses ip ON ip.subnet_id = s.id
		GROUP BY s.id, s.network_address, s.prefix_length, s.description
		HAVING COUNT(ip.id) > 0
		ORDER BY
			COUNT(CASE WHEN ip.status = 'assigned' THEN 1 END)::float /
			GREATEST((POWER(2, 32 - s.prefix_length) - 2)::bigint, 1) DESC
		LIMIT 5`

	topRows, err := r.db.Query(ctx, topQuery)
	if err != nil {
		return nil, fmt.Errorf("top subnets: %w", err)
	}
	defer topRows.Close()

	summary.TopSubnets = make([]models.SubnetUtilisation, 0)
	for topRows.Next() {
		su := models.SubnetUtilisation{}
		if err := topRows.Scan(&su.ID, &su.CIDR, &su.Description, &su.Used, &su.Total); err != nil {
			return nil, err
		}
		if su.Total > 0 {
			su.UtilisationPct = float64(su.Used) / float64(su.Total) * 100
		}
		summary.TopSubnets = append(summary.TopSubnets, su)
	}
	if err := topRows.Err(); err != nil {
		return nil, err
	}

	// Pending request counts (tables may not exist yet in older deployments — treat as 0)
	_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM subnet_requests WHERE status = 'pending'`).Scan(&summary.PendingSubnetRequests)
	_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM ip_requests WHERE status = 'pending'`).Scan(&summary.PendingIPRequests)

	return summary, nil
}

// GetDashboardRecentActivity returns the last 20 relevant audit log entries.
func (r *Repository) GetDashboardRecentActivity(ctx context.Context) ([]*models.DashboardActivity, error) {
	query := `
		SELECT id, action, resource_type, resource_id, user_id, username, COALESCE(resource_name, ''), created_at
		FROM audit_logs
		WHERE action IN ('ip_assigned','ip_released','subnet_created','subnet_deleted','subnet_updated')
		ORDER BY created_at DESC
		LIMIT 20`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	activities := make([]*models.DashboardActivity, 0)
	for rows.Next() {
		a := &models.DashboardActivity{}
		var createdAt time.Time
		if err := rows.Scan(&a.ID, &a.Action, &a.EntityType, &a.EntityID, &a.UserID, &a.Username, &a.Description, &createdAt); err != nil {
			return nil, err
		}
		a.CreatedAt = createdAt.Format(time.RFC3339)
		activities = append(activities, a)
	}
	return activities, rows.Err()
}

// ListNetworksPaginated returns sections with pagination.
func (r *Repository) ListNetworksPaginated(ctx context.Context, limit, offset int) ([]*models.Network, int64, error) {
	return r.ListNetworksPaginatedWithOptions(ctx, ListOptions{Limit: limit, Offset: offset})
}

func (r *Repository) ListNetworksPaginatedWithOptions(ctx context.Context, opts ListOptions) ([]*models.Network, int64, error) {
	where := ""
	args := []interface{}{}
	if opts.Query != "" {
		args = append(args, "%"+opts.Query+"%")
		where = fmt.Sprintf(" WHERE name ILIKE $%d OR description ILIKE $%d", len(args), len(args))
	}

	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM networks`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sortCol := sortExpr(opts.Sort, "created_at", map[string]string{"name": "name", "created_at": "created_at", "updated_at": "updated_at"})
	args = append(args, opts.Limit, opts.Offset)
	query := fmt.Sprintf(`SELECT id, name, description, created_by, created_at, updated_at FROM networks%s ORDER BY %s %s LIMIT $%d OFFSET $%d`, where, sortCol, orderDirection(opts.Order), len(args)-1, len(args))
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	sections := make([]*models.Network, 0)
	for rows.Next() {
		section := &models.Network{}
		if err := rows.Scan(&section.ID, &section.Name, &section.Description, &section.CreatedBy, &section.CreatedAt, &section.UpdatedAt); err != nil {
			return nil, 0, err
		}
		sections = append(sections, section)
	}
	return sections, total, rows.Err()
}

// ListSubnetsBySectionPaginated returns subnets for a section with pagination.
func (r *Repository) ListSubnetsBySectionPaginated(ctx context.Context, networkID int64, limit, offset int) ([]*models.Subnet, int64, error) {
	return r.ListSubnetsBySectionPaginatedWithOptions(ctx, networkID, ListOptions{Limit: limit, Offset: offset})
}

func (r *Repository) ListSubnetsBySectionPaginatedWithOptions(ctx context.Context, networkID int64, opts ListOptions) ([]*models.Subnet, int64, error) {
	args := []interface{}{networkID}
	where := ` WHERE network_id = $1`
	if opts.Query != "" {
		args = append(args, "%"+opts.Query+"%")
		where += fmt.Sprintf(" AND (host(network_address) ILIKE $%d OR description ILIKE $%d)", len(args), len(args))
	}

	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM subnets`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sortCol := sortExpr(opts.Sort, "network_address", map[string]string{"network": "network_address", "prefix": "prefix_length", "created_at": "created_at", "updated_at": "updated_at"})
	args = append(args, opts.Limit, opts.Offset)
	query := fmt.Sprintf(`SELECT id, network_id, host(network_address), prefix_length, description, created_at, updated_at FROM subnets%s ORDER BY %s %s LIMIT $%d OFFSET $%d`, where, sortCol, orderDirection(opts.Order), len(args)-1, len(args))
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	subnets := make([]*models.Subnet, 0)
	for rows.Next() {
		subnet := &models.Subnet{}
		if err := rows.Scan(&subnet.ID, &subnet.NetworkID, &subnet.NetworkAddress, &subnet.PrefixLength, &subnet.Description, &subnet.CreatedAt, &subnet.UpdatedAt); err != nil {
			return nil, 0, err
		}
		subnets = append(subnets, subnet)
	}
	return subnets, total, rows.Err()
}

// ListIPAddressesBySubnetPaginated returns IP addresses for a subnet with pagination.
func (r *Repository) ListIPAddressesBySubnetPaginated(ctx context.Context, subnetID int64, limit, offset int) ([]*models.IPAddress, int64, error) {
	return r.ListIPAddressesBySubnetPaginatedWithOptions(ctx, subnetID, ListOptions{Limit: limit, Offset: offset})
}

func (r *Repository) ListIPAddressesBySubnetPaginatedWithOptions(ctx context.Context, subnetID int64, opts ListOptions) ([]*models.IPAddress, int64, error) {
	args := []interface{}{subnetID}
	where := ` WHERE ip.subnet_id = $1`
	if opts.Status != "" {
		args = append(args, opts.Status)
		where += fmt.Sprintf(" AND ip.status = $%d", len(args))
	}
	if opts.HideAvailable {
		where += " AND ip.status != 'available'"
	}
	if opts.Query != "" {
		args = append(args, "%"+opts.Query+"%")
		where += fmt.Sprintf(" AND (ip.address::text ILIKE $%d OR ip.hostname ILIKE $%d)", len(args), len(args))
	}

	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM ip_addresses ip`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	allowedSorts := map[string]string{
		"address":     "ip.address",
		"hostname":    "ip.hostname",
		"status":      "ip.status",
		"mac_address": "ip.mac_address",
		"last_seen":   "ip.last_seen",
		"created_at":  "ip.created_at",
		"updated_at":  "ip.updated_at",
	}
	sortCol := sortExpr(opts.Sort, "ip.address", allowedSorts)
	args = append(args, opts.Limit, opts.Offset)
	query := fmt.Sprintf(`SELECT `+ipSelectCols+` `+ipFromJoin+`%s ORDER BY %s %s LIMIT $%d OFFSET $%d`, where, sortCol, orderDirection(opts.Order), len(args)-1, len(args))
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	ips := make([]*models.IPAddress, 0)
	for rows.Next() {
		ip, err := scanIP(rows)
		if err != nil {
			return nil, 0, err
		}
		ips = append(ips, ip)
	}
	return ips, total, rows.Err()
}

// ListIPAddressesFullRange returns every address in the subnet's IPv4 CIDR range, merged
// with any existing ip_addresses rows. Addresses with no database record are returned
// with Virtual=true and status "available". Uses generate_series with an explicit offset
// so only `limit` rows are generated — efficient even for large subnets.
func (r *Repository) ListIPAddressesFullRange(
	ctx context.Context,
	subnetID int64,
	networkAddr string,
	prefixLength int,
	offset, limit int,
) ([]*models.IPAddress, int64, error) {
	total := int64(1) << (32 - prefixLength)
	remaining := total - int64(offset)
	if remaining <= 0 {
		return []*models.IPAddress{}, total, nil
	}
	if int64(limit) > remaining {
		limit = int(remaining)
	}

	const query = `
		SELECT
			COALESCE(ip.id, 0)::bigint,
			host($1::inet + ($2::bigint + s)),
			COALESCE(ip.subnet_id, $4)::bigint,
			COALESCE(ip.hostname, ''),
			COALESCE(ip.status, 'available'),
			ip.tag_id,
			t.id, t.name, t.colour, t.description, t.is_system, t.created_at,
			ip.last_seen,
			ip.mac_address, ip.ptr_record,
			ip.dns_name, ip.dns_records::text, ip.dns_last_checked,
			ip.port_open,
			ip.created_at, ip.updated_at,
			ip.device_id, dv.hostname,
			(ip.id IS NULL) AS is_virtual
		FROM generate_series(0, $3::bigint - 1) s
		LEFT JOIN ip_addresses ip
			ON ip.address = ($1::inet + ($2::bigint + s))
			AND ip.subnet_id = $4
		LEFT JOIN ip_tags t ON ip.tag_id = t.id
		LEFT JOIN devices dv ON ip.device_id = dv.id`

	rows, err := r.db.Query(ctx, query, networkAddr, int64(offset), int64(limit), subnetID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	ips := make([]*models.IPAddress, 0, limit)
	for rows.Next() {
		ip, err := scanIPFull(rows)
		if err != nil {
			return nil, 0, err
		}
		ips = append(ips, ip)
	}
	return ips, total, rows.Err()
}

// scanIPFull scans a row produced by ListIPAddressesFullRange. Unlike scanIP it accepts
// nullable created_at/updated_at (virtual IPs have no DB timestamps) and sets Virtual.
func scanIPFull(row interface{ Scan(dest ...any) error }) (*models.IPAddress, error) {
	ip := &models.IPAddress{}
	var tagID, tagIDInner *int64
	var tagName, tagColour, tagDesc *string
	var tagIsSystem *bool
	var tagCreatedAt *time.Time
	var portOpenRaw []byte
	var deviceID *int64
	var deviceHostname *string
	var subnetID int64
	var createdAt, updatedAt *time.Time

	err := row.Scan(
		&ip.ID, &ip.Address, &subnetID, &ip.Hostname, &ip.Status,
		&tagID, &tagIDInner, &tagName, &tagColour, &tagDesc, &tagIsSystem, &tagCreatedAt,
		&ip.LastSeen, &ip.MACAddress, &ip.PTRRecord,
		&ip.DNSName, &ip.DNSRecords, &ip.DNSLastChecked,
		&portOpenRaw,
		&createdAt, &updatedAt,
		&deviceID, &deviceHostname,
		&ip.Virtual,
	)
	if err != nil {
		return nil, err
	}

	ip.SubnetID = subnetID
	if createdAt != nil {
		ip.CreatedAt = *createdAt
	}
	if updatedAt != nil {
		ip.UpdatedAt = *updatedAt
	}
	ip.TagID = tagID
	if tagIDInner != nil {
		ip.Tag = &models.IPTag{
			ID:          *tagIDInner,
			Name:        *tagName,
			Colour:      *tagColour,
			Description: tagDesc,
			IsSystem:    *tagIsSystem,
			CreatedAt:   *tagCreatedAt,
		}
	}
	if len(portOpenRaw) > 0 {
		if err2 := json.Unmarshal(portOpenRaw, &ip.PortOpen); err2 != nil {
			ip.PortOpen = nil
		}
	}
	ip.DeviceID = deviceID
	if deviceID != nil && deviceHostname != nil {
		ip.Device = &models.DeviceSummary{ID: *deviceID, Hostname: *deviceHostname}
	}
	return ip, nil
}
