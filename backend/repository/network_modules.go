package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"padduck/models"
)

type NATRuleParams struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	InternalCIDR string `json:"internal_cidr"`
	ExternalCIDR string `json:"external_cidr"`
	Protocol     string `json:"protocol"`
	InternalPort *int   `json:"internal_port"`
	ExternalPort *int   `json:"external_port"`
	DeviceID     *int64 `json:"device_id"`
	CustomerID   *int64 `json:"customer_id"`
	Description  string `json:"description"`
	Status       string `json:"status"`
}

type DHCPServerParams struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	Vendor      string `json:"vendor"`
	Version     string `json:"version"`
	LocationID  *int64 `json:"location_id"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type DHCPLeaseParams struct {
	ServerID   int64   `json:"server_id"`
	IPAddress  string  `json:"ip_address"`
	MACAddress string  `json:"mac_address"`
	Hostname   string  `json:"hostname"`
	SubnetID   *int64  `json:"subnet_id"`
	IPID       *int64  `json:"ip_id"`
	CustomerID *int64  `json:"customer_id"`
	StartsAt   *string `json:"starts_at"`
	EndsAt     *string `json:"ends_at"`
	State      string  `json:"state"`
}

type CircuitProviderParams struct {
	Name         string `json:"name"`
	AccountNo    string `json:"account_no"`
	SupportEmail string `json:"support_email"`
	SupportPhone string `json:"support_phone"`
	PortalURL    string `json:"portal_url"`
	Notes        string `json:"notes"`
}

type PhysicalCircuitParams struct {
	ProviderID    int64   `json:"provider_id"`
	CircuitID     string  `json:"circuit_id"`
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	Status        string  `json:"status"`
	BandwidthMbps *int    `json:"bandwidth_mbps"`
	LocationAID   *int64  `json:"location_a_id"`
	LocationBID   *int64  `json:"location_b_id"`
	CustomerID    *int64  `json:"customer_id"`
	InstallDate   *string `json:"install_date"`
	Notes         string  `json:"notes"`
}

type LogicalCircuitParams struct {
	PhysicalCircuitID *int64 `json:"physical_circuit_id"`
	Name              string `json:"name"`
	ServiceID         string `json:"service_id"`
	Type              string `json:"type"`
	Status            string `json:"status"`
	VLANID            *int64 `json:"vlan_id"`
	VRFID             *int64 `json:"vrf_id"`
	CustomerID        *int64 `json:"customer_id"`
	BandwidthMbps     *int   `json:"bandwidth_mbps"`
	Notes             string `json:"notes"`
}

type CustomerAssociationParams struct {
	CustomerID   int64  `json:"customer_id"`
	ObjectType   string `json:"object_type"`
	ObjectID     int64  `json:"object_id"`
	Relationship string `json:"relationship"`
	Notes        string `json:"notes"`
}

type FirewallZoneParams struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Status      string `json:"status"`
}

type FirewallZoneMappingParams struct {
	ZoneID      int64  `json:"zone_id"`
	ObjectType  string `json:"object_type"`
	ObjectID    *int64 `json:"object_id"`
	CIDR        string `json:"cidr"`
	Direction   string `json:"direction"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func (r *Repository) ListNATRules(ctx context.Context) ([]*models.NATRule, error) {
	rows, err := r.db.Query(ctx, `SELECT n.id, n.name, n.type, n.internal_cidr, n.external_cidr, n.protocol, n.internal_port, n.external_port, n.device_id, n.customer_id, n.description, n.status, c.name, d.hostname, n.created_at, n.updated_at FROM nat_rules n LEFT JOIN customers c ON c.id=n.customer_id LEFT JOIN devices d ON d.id=n.device_id ORDER BY n.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.NATRule
	for rows.Next() {
		item := &models.NATRule{}
		if err := rows.Scan(&item.ID, &item.Name, &item.Type, &item.InternalCIDR, &item.ExternalCIDR, &item.Protocol, &item.InternalPort, &item.ExternalPort, &item.DeviceID, &item.CustomerID, &item.Description, &item.Status, &item.CustomerName, &item.DeviceName, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *Repository) GetNATRuleByID(ctx context.Context, id int64) (*models.NATRule, error) {
	item := &models.NATRule{}
	err := r.db.QueryRow(ctx, `SELECT n.id, n.name, n.type, n.internal_cidr, n.external_cidr, n.protocol, n.internal_port, n.external_port, n.device_id, n.customer_id, n.description, n.status, c.name, d.hostname, n.created_at, n.updated_at FROM nat_rules n LEFT JOIN customers c ON c.id=n.customer_id LEFT JOIN devices d ON d.id=n.device_id WHERE n.id=$1`, id).Scan(&item.ID, &item.Name, &item.Type, &item.InternalCIDR, &item.ExternalCIDR, &item.Protocol, &item.InternalPort, &item.ExternalPort, &item.DeviceID, &item.CustomerID, &item.Description, &item.Status, &item.CustomerName, &item.DeviceName, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) CreateNATRule(ctx context.Context, p *NATRuleParams) (*models.NATRule, error) {
	id := int64(0)
	err := r.db.QueryRow(ctx, `INSERT INTO nat_rules (name, type, internal_cidr, external_cidr, protocol, internal_port, external_port, device_id, customer_id, description, status) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id`, p.Name, p.Type, p.InternalCIDR, p.ExternalCIDR, p.Protocol, p.InternalPort, p.ExternalPort, p.DeviceID, p.CustomerID, p.Description, p.Status).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetNATRuleByID(ctx, id)
}

func (r *Repository) UpdateNATRule(ctx context.Context, id int64, p *NATRuleParams) (*models.NATRule, error) {
	tag, err := r.db.Exec(ctx, `UPDATE nat_rules SET name=$1, type=$2, internal_cidr=$3, external_cidr=$4, protocol=$5, internal_port=$6, external_port=$7, device_id=$8, customer_id=$9, description=$10, status=$11, updated_at=CURRENT_TIMESTAMP WHERE id=$12`, p.Name, p.Type, p.InternalCIDR, p.ExternalCIDR, p.Protocol, p.InternalPort, p.ExternalPort, p.DeviceID, p.CustomerID, p.Description, p.Status, id)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, pgx.ErrNoRows
	}
	return r.GetNATRuleByID(ctx, id)
}

func (r *Repository) DeleteNATRule(ctx context.Context, id int64) error {
	return deleteByID(ctx, r, "nat_rules", id)
}

func (r *Repository) ListFirewallZones(ctx context.Context) ([]*models.FirewallZone, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, description, color, status, created_at, updated_at FROM firewall_zones ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.FirewallZone
	for rows.Next() {
		item := &models.FirewallZone{}
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Color, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *Repository) GetFirewallZoneByID(ctx context.Context, id int64) (*models.FirewallZone, error) {
	item := &models.FirewallZone{}
	err := r.db.QueryRow(ctx, `SELECT id, name, description, color, status, created_at, updated_at FROM firewall_zones WHERE id=$1`, id).Scan(&item.ID, &item.Name, &item.Description, &item.Color, &item.Status, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) CreateFirewallZone(ctx context.Context, p *FirewallZoneParams) (*models.FirewallZone, error) {
	id := int64(0)
	err := r.db.QueryRow(ctx, `INSERT INTO firewall_zones (name, description, color, status) VALUES ($1,$2,$3,$4) RETURNING id`, p.Name, p.Description, p.Color, p.Status).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetFirewallZoneByID(ctx, id)
}

func (r *Repository) UpdateFirewallZone(ctx context.Context, id int64, p *FirewallZoneParams) (*models.FirewallZone, error) {
	tag, err := r.db.Exec(ctx, `UPDATE firewall_zones SET name=$1, description=$2, color=$3, status=$4, updated_at=CURRENT_TIMESTAMP WHERE id=$5`, p.Name, p.Description, p.Color, p.Status, id)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, pgx.ErrNoRows
	}
	return r.GetFirewallZoneByID(ctx, id)
}

func (r *Repository) DeleteFirewallZone(ctx context.Context, id int64) error {
	return deleteByID(ctx, r, "firewall_zones", id)
}

func (r *Repository) ListFirewallZoneMappings(ctx context.Context, zoneID int64) ([]*models.FirewallZoneMapping, error) {
	query := `SELECT m.id, m.zone_id, m.object_type, m.object_id, COALESCE(m.cidr::text, ''), m.direction, m.description, m.status, z.name, m.created_at, m.updated_at FROM firewall_zone_mappings m JOIN firewall_zones z ON z.id=m.zone_id`
	args := []any{}
	if zoneID > 0 {
		query += ` WHERE m.zone_id=$1`
		args = append(args, zoneID)
	}
	query += ` ORDER BY z.name, m.object_type, m.object_id, m.cidr`
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.FirewallZoneMapping
	for rows.Next() {
		item := &models.FirewallZoneMapping{}
		if err := rows.Scan(&item.ID, &item.ZoneID, &item.ObjectType, &item.ObjectID, &item.CIDR, &item.Direction, &item.Description, &item.Status, &item.ZoneName, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *Repository) GetFirewallZoneMappingByID(ctx context.Context, id int64) (*models.FirewallZoneMapping, error) {
	item := &models.FirewallZoneMapping{}
	err := r.db.QueryRow(ctx, `SELECT m.id, m.zone_id, m.object_type, m.object_id, COALESCE(m.cidr::text, ''), m.direction, m.description, m.status, z.name, m.created_at, m.updated_at FROM firewall_zone_mappings m JOIN firewall_zones z ON z.id=m.zone_id WHERE m.id=$1`, id).Scan(&item.ID, &item.ZoneID, &item.ObjectType, &item.ObjectID, &item.CIDR, &item.Direction, &item.Description, &item.Status, &item.ZoneName, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) CreateFirewallZoneMapping(ctx context.Context, p *FirewallZoneMappingParams) (*models.FirewallZoneMapping, error) {
	id := int64(0)
	err := r.db.QueryRow(ctx, `INSERT INTO firewall_zone_mappings (zone_id, object_type, object_id, cidr, direction, description, status) VALUES ($1,$2,$3,NULLIF($4, '')::cidr,$5,$6,$7) RETURNING id`, p.ZoneID, p.ObjectType, p.ObjectID, p.CIDR, p.Direction, p.Description, p.Status).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetFirewallZoneMappingByID(ctx, id)
}

func (r *Repository) UpdateFirewallZoneMapping(ctx context.Context, id int64, p *FirewallZoneMappingParams) (*models.FirewallZoneMapping, error) {
	tag, err := r.db.Exec(ctx, `UPDATE firewall_zone_mappings SET zone_id=$1, object_type=$2, object_id=$3, cidr=NULLIF($4, '')::cidr, direction=$5, description=$6, status=$7, updated_at=CURRENT_TIMESTAMP WHERE id=$8`, p.ZoneID, p.ObjectType, p.ObjectID, p.CIDR, p.Direction, p.Description, p.Status, id)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, pgx.ErrNoRows
	}
	return r.GetFirewallZoneMappingByID(ctx, id)
}

func (r *Repository) DeleteFirewallZoneMapping(ctx context.Context, id int64) error {
	return deleteByID(ctx, r, "firewall_zone_mappings", id)
}

func (r *Repository) ListDHCPServers(ctx context.Context) ([]*models.DHCPServer, error) {
	rows, err := r.db.Query(ctx, `SELECT s.id, s.name, s.address, s.vendor, s.version, s.location_id, s.description, s.status, l.name, s.created_at, s.updated_at FROM dhcp_servers s LEFT JOIN locations l ON l.id=s.location_id ORDER BY s.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.DHCPServer
	for rows.Next() {
		item := &models.DHCPServer{}
		if err := rows.Scan(&item.ID, &item.Name, &item.Address, &item.Vendor, &item.Version, &item.LocationID, &item.Description, &item.Status, &item.LocationName, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *Repository) GetDHCPServerByID(ctx context.Context, id int64) (*models.DHCPServer, error) {
	item := &models.DHCPServer{}
	err := r.db.QueryRow(ctx, `SELECT s.id, s.name, s.address, s.vendor, s.version, s.location_id, s.description, s.status, l.name, s.created_at, s.updated_at FROM dhcp_servers s LEFT JOIN locations l ON l.id=s.location_id WHERE s.id=$1`, id).Scan(&item.ID, &item.Name, &item.Address, &item.Vendor, &item.Version, &item.LocationID, &item.Description, &item.Status, &item.LocationName, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) CreateDHCPServer(ctx context.Context, p *DHCPServerParams) (*models.DHCPServer, error) {
	id := int64(0)
	err := r.db.QueryRow(ctx, `INSERT INTO dhcp_servers (name, address, vendor, version, location_id, description, status) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`, p.Name, p.Address, p.Vendor, p.Version, p.LocationID, p.Description, p.Status).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetDHCPServerByID(ctx, id)
}

func (r *Repository) UpdateDHCPServer(ctx context.Context, id int64, p *DHCPServerParams) (*models.DHCPServer, error) {
	tag, err := r.db.Exec(ctx, `UPDATE dhcp_servers SET name=$1, address=$2, vendor=$3, version=$4, location_id=$5, description=$6, status=$7, updated_at=CURRENT_TIMESTAMP WHERE id=$8`, p.Name, p.Address, p.Vendor, p.Version, p.LocationID, p.Description, p.Status, id)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, pgx.ErrNoRows
	}
	return r.GetDHCPServerByID(ctx, id)
}

func (r *Repository) DeleteDHCPServer(ctx context.Context, id int64) error {
	return deleteByID(ctx, r, "dhcp_servers", id)
}

func (r *Repository) ListDHCPLeases(ctx context.Context, serverID int64) ([]*models.DHCPLease, error) {
	query := `SELECT l.id, l.server_id, l.ip_address::text, l.mac_address, l.hostname, l.subnet_id, l.ip_id, l.customer_id, l.starts_at, l.ends_at, l.state, s.name, c.name, l.created_at, l.updated_at FROM dhcp_leases l JOIN dhcp_servers s ON s.id=l.server_id LEFT JOIN customers c ON c.id=l.customer_id`
	args := []any{}
	if serverID > 0 {
		query += ` WHERE l.server_id=$1`
		args = append(args, serverID)
	}
	query += ` ORDER BY l.ip_address`
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.DHCPLease
	for rows.Next() {
		item := &models.DHCPLease{}
		if err := rows.Scan(&item.ID, &item.ServerID, &item.IPAddress, &item.MACAddress, &item.Hostname, &item.SubnetID, &item.IPID, &item.CustomerID, &item.StartsAt, &item.EndsAt, &item.State, &item.ServerName, &item.CustomerName, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *Repository) GetDHCPLeaseByID(ctx context.Context, id int64) (*models.DHCPLease, error) {
	item := &models.DHCPLease{}
	err := r.db.QueryRow(ctx, `SELECT l.id, l.server_id, l.ip_address::text, l.mac_address, l.hostname, l.subnet_id, l.ip_id, l.customer_id, l.starts_at, l.ends_at, l.state, s.name, c.name, l.created_at, l.updated_at FROM dhcp_leases l JOIN dhcp_servers s ON s.id=l.server_id LEFT JOIN customers c ON c.id=l.customer_id WHERE l.id=$1`, id).Scan(&item.ID, &item.ServerID, &item.IPAddress, &item.MACAddress, &item.Hostname, &item.SubnetID, &item.IPID, &item.CustomerID, &item.StartsAt, &item.EndsAt, &item.State, &item.ServerName, &item.CustomerName, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) CreateDHCPLease(ctx context.Context, p *DHCPLeaseParams) (*models.DHCPLease, error) {
	id := int64(0)
	err := r.db.QueryRow(ctx, `INSERT INTO dhcp_leases (server_id, ip_address, mac_address, hostname, subnet_id, ip_id, customer_id, starts_at, ends_at, state) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`, p.ServerID, p.IPAddress, p.MACAddress, p.Hostname, p.SubnetID, p.IPID, p.CustomerID, p.StartsAt, p.EndsAt, p.State).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetDHCPLeaseByID(ctx, id)
}

func (r *Repository) UpdateDHCPLease(ctx context.Context, id int64, p *DHCPLeaseParams) (*models.DHCPLease, error) {
	tag, err := r.db.Exec(ctx, `UPDATE dhcp_leases SET server_id=$1, ip_address=$2, mac_address=$3, hostname=$4, subnet_id=$5, ip_id=$6, customer_id=$7, starts_at=$8, ends_at=$9, state=$10, updated_at=CURRENT_TIMESTAMP WHERE id=$11`, p.ServerID, p.IPAddress, p.MACAddress, p.Hostname, p.SubnetID, p.IPID, p.CustomerID, p.StartsAt, p.EndsAt, p.State, id)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, pgx.ErrNoRows
	}
	return r.GetDHCPLeaseByID(ctx, id)
}

func (r *Repository) DeleteDHCPLease(ctx context.Context, id int64) error {
	return deleteByID(ctx, r, "dhcp_leases", id)
}

func deleteByID(ctx context.Context, r *Repository, table string, id int64) error {
	tag, err := r.db.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE id=$1", table), id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
