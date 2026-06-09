package repository

import (
	"context"
	"time"

	"padduck/models"
)

// Subnet operations

// subnetSelectCols is the base column list for subnets (no JOIN).
const subnetSelectCols = `s.id, s.network_id, host(s.network_address), s.prefix_length, s.description, s.gateway, s.auto_reserve_first, s.auto_reserve_last, s.location_id, s.nameserver_id, s.vlan_id, s.parent_subnet_id, s.is_container, s.alert_threshold_pct, s.alert_email_override, s.created_at, s.updated_at, ns.id, ns.name, ns.server1, ns.server2, ns.server3, ns.description, ns.created_at, ns.updated_at, s.scan_profile_id`

const subnetFromJoin = `FROM subnets s LEFT JOIN nameservers ns ON s.nameserver_id = ns.id`

func scanSubnet(row interface {
	Scan(dest ...any) error
}) (*models.Subnet, error) {
	subnet := &models.Subnet{}
	var nsID *int64
	var nsName, nsServer1 *string
	var nsServer2, nsServer3, nsDesc *string
	var nsCreatedAt, nsUpdatedAt *time.Time
	err := row.Scan(
		&subnet.ID, &subnet.NetworkID, &subnet.NetworkAddress, &subnet.PrefixLength,
		&subnet.Description, &subnet.Gateway, &subnet.AutoReserveFirst, &subnet.AutoReserveLast,
		&subnet.LocationID, &subnet.NameserverID, &subnet.VLANID, &subnet.ParentSubnetID, &subnet.IsContainer,
		&subnet.AlertThresholdPct, &subnet.AlertEmailOverride,
		&subnet.CreatedAt, &subnet.UpdatedAt,
		&nsID, &nsName, &nsServer1, &nsServer2, &nsServer3, &nsDesc, &nsCreatedAt, &nsUpdatedAt,
		&subnet.ScanProfileID,
	)
	if err != nil {
		return nil, err
	}
	if nsID != nil {
		subnet.Nameserver = &models.Nameserver{
			ID:          *nsID,
			Name:        *nsName,
			Server1:     *nsServer1,
			Server2:     nsServer2,
			Server3:     nsServer3,
			Description: nsDesc,
			CreatedAt:   *nsCreatedAt,
			UpdatedAt:   *nsUpdatedAt,
		}
	}
	return subnet, nil
}

func (r *Repository) CreateSubnet(ctx context.Context, networkID int64, networkAddress string, prefixLength int, description string, gateway *string, autoFirst, autoLast bool) (*models.Subnet, error) {
	query := `INSERT INTO subnets (network_id, network_address, prefix_length, description, gateway, auto_reserve_first, auto_reserve_last)
	          VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	var id int64
	if err := r.db.QueryRow(ctx, query, networkID, networkAddress, prefixLength, description, gateway, autoFirst, autoLast).Scan(&id); err != nil {
		return nil, err
	}
	return r.GetSubnetByID(ctx, id)
}

// CreateSubnetWithLocation inserts a new subnet with an optional location.
func (r *Repository) CreateSubnetWithLocation(ctx context.Context, networkID int64, networkAddress string, prefixLength int, description string, gateway *string, autoFirst, autoLast bool, locationID *int64, nameserverID ...*int64) (*models.Subnet, error) {
	var nsID *int64
	if len(nameserverID) > 0 {
		nsID = nameserverID[0]
	}
	query := `INSERT INTO subnets (network_id, network_address, prefix_length, description, gateway, auto_reserve_first, auto_reserve_last, location_id, nameserver_id)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`
	var id int64
	if err := r.db.QueryRow(ctx, query, networkID, networkAddress, prefixLength, description, gateway, autoFirst, autoLast, locationID, nsID).Scan(&id); err != nil {
		return nil, err
	}
	return r.GetSubnetByID(ctx, id)
}

func (r *Repository) CreateSubnetWithVLAN(ctx context.Context, networkID int64, networkAddress string, prefixLength int, description string, gateway *string, autoFirst, autoLast bool, locationID *int64, nameserverID *int64, vlanID *int64) (*models.Subnet, error) {
	query := `INSERT INTO subnets (network_id, network_address, prefix_length, description, gateway, auto_reserve_first, auto_reserve_last, location_id, nameserver_id, vlan_id)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`
	var id int64
	if err := r.db.QueryRow(ctx, query, networkID, networkAddress, prefixLength, description, gateway, autoFirst, autoLast, locationID, nameserverID, vlanID).Scan(&id); err != nil {
		return nil, err
	}
	return r.GetSubnetByID(ctx, id)
}

func (r *Repository) GetSubnetByID(ctx context.Context, id int64) (*models.Subnet, error) {
	query := `SELECT ` + subnetSelectCols + ` ` + subnetFromJoin + ` WHERE s.id = $1`
	row := r.db.QueryRow(ctx, query, id)
	return scanSubnet(row)
}

// GetSubnetByCIDR looks up a subnet by CIDR notation (e.g. "192.168.1.0/24").
// network_address stores only the host IP; prefix_length is in a separate column.
func (r *Repository) GetSubnetByCIDR(ctx context.Context, cidr string) (*models.Subnet, error) {
	query := `SELECT ` + subnetSelectCols + ` ` + subnetFromJoin + ` WHERE host(s.network_address) = host($1::inet) AND s.prefix_length = masklen($1::inet)`
	row := r.db.QueryRow(ctx, query, cidr)
	return scanSubnet(row)
}

func (r *Repository) ListSubnetsBySection(ctx context.Context, networkID int64) ([]*models.Subnet, error) {
	query := `SELECT ` + subnetSelectCols + ` ` + subnetFromJoin + ` WHERE s.network_id = $1 ORDER BY s.network_address`
	rows, err := r.db.Query(ctx, query, networkID)
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

// ListSubnetsByLocation returns subnets assigned to a specific location.
func (r *Repository) ListSubnetsByLocation(ctx context.Context, locationID int64) ([]*models.Subnet, error) {
	query := `SELECT ` + subnetSelectCols + ` ` + subnetFromJoin + ` WHERE s.location_id=$1 ORDER BY s.network_address`
	rows, err := r.db.Query(ctx, query, locationID)
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

func (r *Repository) UpdateSubnet(ctx context.Context, id int64, description string, gateway *string, autoFirst, autoLast bool) (*models.Subnet, error) {
	query := `UPDATE subnets SET description = $1, gateway = $2, auto_reserve_first = $3, auto_reserve_last = $4,
	          updated_at = CURRENT_TIMESTAMP WHERE id = $5`
	if _, err := r.db.Exec(ctx, query, description, gateway, autoFirst, autoLast, id); err != nil {
		return nil, err
	}
	return r.GetSubnetByID(ctx, id)
}

// UpdateSubnetWithLocation updates a subnet including its location and nameserver assignment.
func (r *Repository) UpdateSubnetWithLocation(ctx context.Context, id int64, description string, gateway *string, autoFirst, autoLast bool, locationID *int64, nameserverID ...*int64) (*models.Subnet, error) {
	var nsID *int64
	if len(nameserverID) > 0 {
		nsID = nameserverID[0]
	}
	query := `UPDATE subnets SET description=$1, gateway=$2, auto_reserve_first=$3, auto_reserve_last=$4,
	          location_id=$5, nameserver_id=$6, updated_at=CURRENT_TIMESTAMP WHERE id=$7`
	if _, err := r.db.Exec(ctx, query, description, gateway, autoFirst, autoLast, locationID, nsID, id); err != nil {
		return nil, err
	}
	return r.GetSubnetByID(ctx, id)
}

// UpdateSubnetWithVLAN updates a subnet including vlan_id assignment.
func (r *Repository) UpdateSubnetWithVLAN(ctx context.Context, id int64, description string, gateway *string, autoFirst, autoLast bool, locationID *int64, nameserverID *int64, vlanID *int64) (*models.Subnet, error) {
	query := `UPDATE subnets SET description=$1, gateway=$2, auto_reserve_first=$3, auto_reserve_last=$4,
	          location_id=$5, nameserver_id=$6, vlan_id=$7, updated_at=CURRENT_TIMESTAMP WHERE id=$8`
	if _, err := r.db.Exec(ctx, query, description, gateway, autoFirst, autoLast, locationID, nameserverID, vlanID, id); err != nil {
		return nil, err
	}
	return r.GetSubnetByID(ctx, id)
}

// AssignSubnetToVLAN updates only the VLAN association for a subnet.
func (r *Repository) AssignSubnetToVLAN(ctx context.Context, subnetID int64, vlanID *int64) (*models.Subnet, error) {
	query := `UPDATE subnets SET vlan_id=$1, updated_at=CURRENT_TIMESTAMP WHERE id=$2`
	if _, err := r.db.Exec(ctx, query, vlanID, subnetID); err != nil {
		return nil, err
	}
	return r.GetSubnetByID(ctx, subnetID)
}

func (r *Repository) DeleteSubnet(ctx context.Context, id int64) error {
	query := `DELETE FROM subnets WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
