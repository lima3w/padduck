package repository

import (
	"context"
	"fmt"
	"time"

	"padduck/models"
)

// Dashboard operations

// GetDashboardSummary returns aggregate IPAM counts and top utilised subnets.
func (r *Repository) GetDashboardSummary(ctx context.Context) (*models.DashboardSummary, error) {
	summary := &models.DashboardSummary{}

	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM sections`).Scan(&summary.TotalSections); err != nil {
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
			COUNT(ip.id) AS total
		FROM subnets s
		LEFT JOIN ip_addresses ip ON ip.subnet_id = s.id
		GROUP BY s.id, s.network_address, s.prefix_length, s.description
		HAVING COUNT(ip.id) > 0
		ORDER BY
			CASE WHEN COUNT(ip.id) > 0
				THEN COUNT(CASE WHEN ip.status = 'assigned' THEN 1 END)::float / COUNT(ip.id)
				ELSE 0
			END DESC
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

// ListSectionsPaginated returns sections with pagination.
func (r *Repository) ListSectionsPaginated(ctx context.Context, limit, offset int) ([]*models.Section, int64, error) {
	return r.ListSectionsPaginatedWithOptions(ctx, ListOptions{Limit: limit, Offset: offset})
}

func (r *Repository) ListSectionsPaginatedWithOptions(ctx context.Context, opts ListOptions) ([]*models.Section, int64, error) {
	where := ""
	args := []interface{}{}
	if opts.Query != "" {
		args = append(args, "%"+opts.Query+"%")
		where = fmt.Sprintf(" WHERE name ILIKE $%d OR description ILIKE $%d", len(args), len(args))
	}

	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM sections`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sortCol := sortExpr(opts.Sort, "created_at", map[string]string{"name": "name", "created_at": "created_at", "updated_at": "updated_at"})
	args = append(args, opts.Limit, opts.Offset)
	query := fmt.Sprintf(`SELECT id, name, description, created_by, created_at, updated_at FROM sections%s ORDER BY %s %s LIMIT $%d OFFSET $%d`, where, sortCol, orderDirection(opts.Order), len(args)-1, len(args))
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	sections := make([]*models.Section, 0)
	for rows.Next() {
		section := &models.Section{}
		if err := rows.Scan(&section.ID, &section.Name, &section.Description, &section.CreatedBy, &section.CreatedAt, &section.UpdatedAt); err != nil {
			return nil, 0, err
		}
		sections = append(sections, section)
	}
	return sections, total, rows.Err()
}

// ListSubnetsBySectionPaginated returns subnets for a section with pagination.
func (r *Repository) ListSubnetsBySectionPaginated(ctx context.Context, sectionID int64, limit, offset int) ([]*models.Subnet, int64, error) {
	return r.ListSubnetsBySectionPaginatedWithOptions(ctx, sectionID, ListOptions{Limit: limit, Offset: offset})
}

func (r *Repository) ListSubnetsBySectionPaginatedWithOptions(ctx context.Context, sectionID int64, opts ListOptions) ([]*models.Subnet, int64, error) {
	args := []interface{}{sectionID}
	where := ` WHERE section_id = $1`
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
	query := fmt.Sprintf(`SELECT id, section_id, host(network_address), prefix_length, description, created_at, updated_at FROM subnets%s ORDER BY %s %s LIMIT $%d OFFSET $%d`, where, sortCol, orderDirection(opts.Order), len(args)-1, len(args))
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	subnets := make([]*models.Subnet, 0)
	for rows.Next() {
		subnet := &models.Subnet{}
		if err := rows.Scan(&subnet.ID, &subnet.SectionID, &subnet.NetworkAddress, &subnet.PrefixLength, &subnet.Description, &subnet.CreatedAt, &subnet.UpdatedAt); err != nil {
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
	where := ` WHERE subnet_id = $1`
	if opts.Status != "" {
		args = append(args, opts.Status)
		where += fmt.Sprintf(" AND status = $%d", len(args))
	}
	if opts.Query != "" {
		args = append(args, "%"+opts.Query+"%")
		where += fmt.Sprintf(" AND (address::text ILIKE $%d OR hostname ILIKE $%d OR assigned_to ILIKE $%d)", len(args), len(args), len(args))
	}

	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM ip_addresses`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sortCol := sortExpr(opts.Sort, "address", map[string]string{"address": "address", "hostname": "hostname", "status": "status", "created_at": "created_at", "updated_at": "updated_at"})
	args = append(args, opts.Limit, opts.Offset)
	query := fmt.Sprintf(`SELECT id, subnet_id, address::text, hostname, status, assigned_to, created_at, updated_at FROM ip_addresses%s ORDER BY %s %s LIMIT $%d OFFSET $%d`, where, sortCol, orderDirection(opts.Order), len(args)-1, len(args))
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	ips := make([]*models.IPAddress, 0)
	for rows.Next() {
		ip := &models.IPAddress{}
		if err := rows.Scan(&ip.ID, &ip.SubnetID, &ip.Address, &ip.Hostname, &ip.Status, &ip.AssignedTo, &ip.CreatedAt, &ip.UpdatedAt); err != nil {
			return nil, 0, err
		}
		ips = append(ips, ip)
	}
	return ips, total, rows.Err()
}
