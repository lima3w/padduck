package repository

import (
	"context"
	"fmt"
	"time"

	"ipam-next/models"
)

// IPSearchFilter holds optional additional filters for IP search
type IPSearchFilter struct {
	TagID          *int64
	MACAddress     string
	PTRRecord      string
	IsAssigned     *bool
	LastSeenAfter  *time.Time
	LastSeenBefore *time.Time
}

// Search operations

func (r *Repository) SearchSections(ctx context.Context, query string, limit, offset int64) ([]*models.Section, error) {
	sql := `SELECT id, name, description, created_by, created_at, updated_at FROM sections
	        WHERE name ILIKE $1 OR description ILIKE $1
	        ORDER BY created_at DESC
	        LIMIT $2 OFFSET $3`
	searchQuery := "%" + query + "%"
	rows, err := r.db.Query(ctx, sql, searchQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sections := make([]*models.Section, 0)
	for rows.Next() {
		section := &models.Section{}
		err := rows.Scan(&section.ID, &section.Name, &section.Description, &section.CreatedBy, &section.CreatedAt, &section.UpdatedAt)
		if err != nil {
			return nil, err
		}
		sections = append(sections, section)
	}
	return sections, rows.Err()
}

func (r *Repository) SearchSubnets(ctx context.Context, sectionID int64, query string, limit, offset int64) ([]*models.Subnet, error) {
	sql := `SELECT ` + subnetSelectCols + ` ` + subnetFromJoin + `
	        WHERE s.section_id = $1 AND (host(s.network_address) ILIKE $2 OR s.description ILIKE $2)
	        ORDER BY s.network_address ASC
	        LIMIT $3 OFFSET $4`
	searchQuery := "%" + query + "%"
	rows, err := r.db.Query(ctx, sql, sectionID, searchQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subnets := make([]*models.Subnet, 0)
	for rows.Next() {
		subnet, err := scanSubnet(rows)
		if err != nil {
			return nil, err
		}
		subnets = append(subnets, subnet)
	}
	return subnets, rows.Err()
}

func (r *Repository) SearchIPAddresses(ctx context.Context, subnetID int64, query string, status string, limit, offset int64, filter ...IPSearchFilter) ([]*models.IPAddress, error) {
	sql := `SELECT ` + ipSelectCols + ` ` + ipFromJoin + `
	        WHERE ip.subnet_id = $1 AND (ip.address::text ILIKE $2 OR ip.hostname ILIKE $2 OR ip.assigned_to ILIKE $2)`
	args := []interface{}{subnetID, "%" + query + "%"}
	n := 3

	if status != "" {
		sql += fmt.Sprintf(" AND ip.status = $%d", n)
		args = append(args, status)
		n++
	}

	if len(filter) > 0 {
		f := filter[0]
		if f.TagID != nil {
			sql += fmt.Sprintf(" AND ip.tag_id = $%d", n)
			args = append(args, *f.TagID)
			n++
		}
		if f.MACAddress != "" {
			sql += fmt.Sprintf(" AND ip.mac_address ILIKE $%d", n)
			args = append(args, "%"+f.MACAddress+"%")
			n++
		}
		if f.PTRRecord != "" {
			sql += fmt.Sprintf(" AND ip.ptr_record ILIKE $%d", n)
			args = append(args, "%"+f.PTRRecord+"%")
			n++
		}
		if f.IsAssigned != nil {
			if *f.IsAssigned {
				sql += " AND ip.status = 'assigned'"
			} else {
				sql += " AND ip.status != 'assigned'"
			}
		}
		if f.LastSeenAfter != nil {
			sql += fmt.Sprintf(" AND ip.last_seen >= $%d", n)
			args = append(args, *f.LastSeenAfter)
			n++
		}
		if f.LastSeenBefore != nil {
			sql += fmt.Sprintf(" AND ip.last_seen <= $%d", n)
			args = append(args, *f.LastSeenBefore)
			n++
		}
	}

	sql += fmt.Sprintf(" ORDER BY ip.address ASC LIMIT $%d OFFSET $%d", n, n+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ips := make([]*models.IPAddress, 0)
	for rows.Next() {
		ip, err := scanIP(rows)
		if err != nil {
			return nil, err
		}
		ips = append(ips, ip)
	}
	return ips, rows.Err()
}
