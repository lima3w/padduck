package repository

import (
	"context"
	"fmt"

	"padduck/models"
)

// ListIPAddressesByDevice returns all IP addresses linked to a device.
func (r *Repository) ListIPAddressesByDevice(ctx context.Context, deviceID int64) ([]*models.IPAddress, error) {
	query := `
		SELECT id, subnet_id, host(address), hostname, status, created_at, updated_at,
		       device_id, interface_name, is_primary
		FROM ip_addresses
		WHERE device_id=$1
		ORDER BY address`
	rows, err := r.db.Query(ctx, query, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ips := make([]*models.IPAddress, 0)
	for rows.Next() {
		ip := &models.IPAddress{}
		if err := rows.Scan(
			&ip.ID, &ip.SubnetID, &ip.Address, &ip.Hostname, &ip.Status,
			&ip.CreatedAt, &ip.UpdatedAt, &ip.DeviceID, &ip.InterfaceName, &ip.IsPrimary,
		); err != nil {
			return nil, err
		}
		ips = append(ips, ip)
	}
	return ips, rows.Err()
}

// AssociateIPToDevice links an IP address to a device.
func (r *Repository) AssociateIPToDevice(ctx context.Context, deviceID, ipID int64, interfaceName *string, isPrimary bool) error {
	_, err := r.db.Exec(ctx,
		`UPDATE ip_addresses SET device_id=$1, interface_name=$2, is_primary=$3, updated_at=now() WHERE id=$4`,
		deviceID, interfaceName, isPrimary, ipID,
	)
	return err
}

// UnlinkIPFromDevice removes the device association from an IP address.
func (r *Repository) UnlinkIPFromDevice(ctx context.Context, deviceID, ipID int64) error {
	ct, err := r.db.Exec(ctx,
		`UPDATE ip_addresses SET device_id=NULL, interface_name=NULL, is_primary=false, updated_at=now()
		 WHERE id=$1 AND device_id=$2`,
		ipID, deviceID,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("ip address not associated with this device")
	}
	return nil
}

// GetSubnetTreeBySection returns all subnets for a section with utilisation counts, ordered by network address.
func (r *Repository) GetSubnetTreeBySection(ctx context.Context, networkID int64) ([]models.SubnetTreeNode, error) {
	query := `
		SELECT
			s.id,
			host(s.network_address) || '/' || s.prefix_length AS cidr,
			s.description,
			COUNT(CASE WHEN ip.status = 'assigned' THEN 1 END) AS used,
			COUNT(ip.id) AS total
		FROM subnets s
		LEFT JOIN ip_addresses ip ON ip.subnet_id = s.id
		WHERE s.network_id = $1
		GROUP BY s.id, s.network_address, s.prefix_length, s.description
		ORDER BY s.network_address`

	rows, err := r.db.Query(ctx, query, networkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodes := make([]models.SubnetTreeNode, 0)
	for rows.Next() {
		var n models.SubnetTreeNode
		if err := rows.Scan(&n.ID, &n.CIDR, &n.Description, &n.Used, &n.Total); err != nil {
			return nil, err
		}
		if n.Total > 0 {
			n.UtilizationPct = float64(n.Used) / float64(n.Total) * 100
		}
		nodes = append(nodes, n)
	}
	return nodes, rows.Err()
}
