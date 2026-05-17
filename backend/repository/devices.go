package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"ipam-next/models"
)

// ---- Device management (v1.3.0) ----

// DeviceParams holds fields for creating or updating a device.
type DeviceParams struct {
	Hostname        string             `json:"hostname"`
	Description     *string            `json:"description"`
	TypeID          *int64             `json:"type_id"`
	SectionID       *int64             `json:"section_id"`
	Vendor          *string            `json:"vendor"`
	Model           *string            `json:"model"`
	OSVersion       *string            `json:"os_version"`
	SNMPCommunity   *string            `json:"snmp_community"`
	SNMPVersion     string             `json:"snmp_version"`
	SNMPV3User      *string            `json:"snmp_v3_user"`
	SNMPV3AuthProto *string            `json:"snmp_v3_auth_proto"`
	SNMPV3AuthPass  *string            `json:"snmp_v3_auth_pass"`
	SNMPV3PrivProto *string            `json:"snmp_v3_priv_proto"`
	SNMPV3PrivPass  *string            `json:"snmp_v3_priv_pass"`
	LocationID      *int64             `json:"location_id"`
	RackID          *int64             `json:"rack_id"`
	RackUnitStart   *int               `json:"rack_unit_start"`
	RackUnitSize    int                `json:"rack_unit_size"`
	CustomFields    map[string]*string `json:"custom_fields"`
}

// DeviceInterfaceParams holds fields for creating or updating a device interface.
type DeviceInterfaceParams struct {
	Name                   string  `json:"name"`
	Description            *string `json:"description"`
	SpeedMbps              *int    `json:"speed_mbps"`
	MediaType              *string `json:"media_type"`
	VLANID                 *int64  `json:"vlan_id"`
	IPAddressID            *int64  `json:"ip_address_id"`
	ConnectedToDeviceID    *int64  `json:"connected_to_device_id"`
	ConnectedToInterfaceID *int64  `json:"connected_to_interface_id"`
}

// DeviceSearchFilter holds optional criteria for device search.
type DeviceSearchFilter struct {
	Query     string  `json:"query"`
	TypeID    *int64  `json:"type_id"`
	SectionID *int64  `json:"section_id"`
	Vendor    *string `json:"vendor"`
	IsOnline  *bool   `json:"is_online"`
	VLANID    *int64  `json:"vlan_id"`
}

// ListDeviceTypes returns all device types ordered by name.
func (r *Repository) ListDeviceTypes(ctx context.Context) ([]*models.DeviceType, error) {
	query := `SELECT id, name, COALESCE(icon, ''), description, created_at, updated_at FROM device_types ORDER BY name`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	types := make([]*models.DeviceType, 0)
	for rows.Next() {
		dt := &models.DeviceType{}
		if err := rows.Scan(&dt.ID, &dt.Name, &dt.Icon, &dt.Description, &dt.CreatedAt, &dt.UpdatedAt); err != nil {
			return nil, err
		}
		types = append(types, dt)
	}
	return types, rows.Err()
}

const deviceSelectCols = `d.id, d.hostname, d.description, d.type_id, d.section_id, d.vendor, d.model, d.os_version, d.is_online, d.last_ping_at, d.location_id, d.rack_id, d.rack_unit_start, d.rack_unit_size, d.created_at, d.updated_at`
const deviceTypeSelectCols = `dt.id, dt.name, COALESCE(dt.icon, ''), dt.description, dt.created_at, dt.updated_at`

func scanDevice(row pgx.Row) (*models.Device, error) {
	d := &models.Device{}
	err := row.Scan(
		&d.ID, &d.Hostname, &d.Description, &d.TypeID, &d.SectionID,
		&d.Vendor, &d.Model, &d.OSVersion, &d.IsOnline, &d.LastPingAt,
		&d.LocationID, &d.RackID, &d.RackUnitStart, &d.RackUnitSize,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// ListDevices returns a paginated list of devices with their type and IP count.
func (r *Repository) ListDevices(ctx context.Context, limit, offset int) ([]*models.Device, int64, error) {
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM devices`).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT ` + deviceSelectCols + `,
		       ` + deviceTypeSelectCols + `,
		       COUNT(ip.id) AS ip_count
		FROM devices d
		LEFT JOIN device_types dt ON dt.id = d.type_id
		LEFT JOIN ip_addresses ip ON ip.device_id = d.id
		GROUP BY d.id, dt.id
		ORDER BY d.hostname
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	devices := make([]*models.Device, 0)
	for rows.Next() {
		d := &models.Device{}
		dt := &models.DeviceType{}
		var dtID *int64
		err := rows.Scan(
			&d.ID, &d.Hostname, &d.Description, &d.TypeID, &d.SectionID,
			&d.Vendor, &d.Model, &d.OSVersion, &d.IsOnline, &d.LastPingAt,
			&d.LocationID, &d.RackID, &d.RackUnitStart, &d.RackUnitSize,
			&d.CreatedAt, &d.UpdatedAt,
			&dtID, &dt.Name, &dt.Icon, &dt.Description, &dt.CreatedAt, &dt.UpdatedAt,
			&d.IPCount,
		)
		if err != nil {
			return nil, 0, err
		}
		if dtID != nil {
			dt.ID = *dtID
			d.Type = dt
		}
		devices = append(devices, d)
	}
	return devices, total, rows.Err()
}

// ListAllDevices returns all devices with their type and IP count (no pagination).
func (r *Repository) ListAllDevices(ctx context.Context) ([]*models.Device, error) {
	query := `
		SELECT ` + deviceSelectCols + `,
		       ` + deviceTypeSelectCols + `,
		       COUNT(ip.id) AS ip_count
		FROM devices d
		LEFT JOIN device_types dt ON dt.id = d.type_id
		LEFT JOIN ip_addresses ip ON ip.device_id = d.id
		GROUP BY d.id, dt.id
		ORDER BY d.hostname`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	devices := make([]*models.Device, 0)
	for rows.Next() {
		d := &models.Device{}
		dt := &models.DeviceType{}
		var dtID *int64
		err := rows.Scan(
			&d.ID, &d.Hostname, &d.Description, &d.TypeID, &d.SectionID,
			&d.Vendor, &d.Model, &d.OSVersion, &d.IsOnline, &d.LastPingAt,
			&d.LocationID, &d.RackID, &d.RackUnitStart, &d.RackUnitSize,
			&d.CreatedAt, &d.UpdatedAt,
			&dtID, &dt.Name, &dt.Icon, &dt.Description, &dt.CreatedAt, &dt.UpdatedAt,
			&d.IPCount,
		)
		if err != nil {
			return nil, err
		}
		if dtID != nil {
			dt.ID = *dtID
			d.Type = dt
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

// CreateDevice inserts a new device and returns the created device.
func (r *Repository) CreateDevice(ctx context.Context, p *DeviceParams) (*models.Device, error) {
	if p.SNMPVersion == "" {
		p.SNMPVersion = "v2c"
	}
	query := `
		INSERT INTO devices (hostname, description, type_id, section_id, vendor, model, os_version,
		                     snmp_community, snmp_version, snmp_v3_user, snmp_v3_auth_proto,
		                     snmp_v3_auth_pass, snmp_v3_priv_proto, snmp_v3_priv_pass,
		                     location_id, rack_id, rack_unit_start, rack_unit_size)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
		RETURNING id, hostname, description, type_id, section_id, vendor, model, os_version,
		          is_online, last_ping_at, location_id, rack_id, rack_unit_start, rack_unit_size, created_at, updated_at`
	row := r.db.QueryRow(ctx, query,
		p.Hostname, p.Description, p.TypeID, p.SectionID, p.Vendor, p.Model, p.OSVersion,
		p.SNMPCommunity, p.SNMPVersion, p.SNMPV3User, p.SNMPV3AuthProto,
		p.SNMPV3AuthPass, p.SNMPV3PrivProto, p.SNMPV3PrivPass,
		p.LocationID, p.RackID, p.RackUnitStart, p.RackUnitSize,
	)
	return scanDevice(row)
}

// GetDeviceByID returns a device with its type info and IP count.
func (r *Repository) GetDeviceByID(ctx context.Context, id int64) (*models.Device, error) {
	query := `
		SELECT ` + deviceSelectCols + `,
		       ` + deviceTypeSelectCols + `,
		       COUNT(ip.id) AS ip_count
		FROM devices d
		LEFT JOIN device_types dt ON dt.id = d.type_id
		LEFT JOIN ip_addresses ip ON ip.device_id = d.id
		WHERE d.id = $1
		GROUP BY d.id, dt.id`

	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("device not found")
	}

	d := &models.Device{}
	dt := &models.DeviceType{}
	var dtID *int64
	err = rows.Scan(
		&d.ID, &d.Hostname, &d.Description, &d.TypeID, &d.SectionID,
		&d.Vendor, &d.Model, &d.OSVersion, &d.IsOnline, &d.LastPingAt,
		&d.LocationID, &d.RackID, &d.RackUnitStart, &d.RackUnitSize,
		&d.CreatedAt, &d.UpdatedAt,
		&dtID, &dt.Name, &dt.Icon, &dt.Description, &dt.CreatedAt, &dt.UpdatedAt,
		&d.IPCount,
	)
	if err != nil {
		return nil, err
	}
	if dtID != nil {
		dt.ID = *dtID
		d.Type = dt
	}
	return d, rows.Err()
}

// UpdateDevice updates an existing device and returns the updated record.
func (r *Repository) UpdateDevice(ctx context.Context, id int64, p *DeviceParams) (*models.Device, error) {
	if p.SNMPVersion == "" {
		p.SNMPVersion = "v2c"
	}
	query := `
		UPDATE devices SET
		  hostname=$1, description=$2, type_id=$3, section_id=$4, vendor=$5, model=$6, os_version=$7,
		  snmp_community=$8, snmp_version=$9, snmp_v3_user=$10, snmp_v3_auth_proto=$11,
		  snmp_v3_auth_pass=$12, snmp_v3_priv_proto=$13, snmp_v3_priv_pass=$14,
		  location_id=$15, rack_id=$16, rack_unit_start=$17, rack_unit_size=$18,
		  updated_at=now()
		WHERE id=$19
		RETURNING id, hostname, description, type_id, section_id, vendor, model, os_version,
		          is_online, last_ping_at, location_id, rack_id, rack_unit_start, rack_unit_size, created_at, updated_at`
	row := r.db.QueryRow(ctx, query,
		p.Hostname, p.Description, p.TypeID, p.SectionID, p.Vendor, p.Model, p.OSVersion,
		p.SNMPCommunity, p.SNMPVersion, p.SNMPV3User, p.SNMPV3AuthProto,
		p.SNMPV3AuthPass, p.SNMPV3PrivProto, p.SNMPV3PrivPass,
		p.LocationID, p.RackID, p.RackUnitStart, p.RackUnitSize, id,
	)
	d, err := scanDevice(row)
	if err != nil {
		return nil, err
	}
	// Re-fetch with type info
	return r.GetDeviceByID(ctx, d.ID)
}

// DeleteDevice deletes a device by ID.
func (r *Repository) DeleteDevice(ctx context.Context, id int64) error {
	ct, err := r.db.Exec(ctx, `DELETE FROM devices WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("device not found")
	}
	return nil
}

// GetDeviceSNMP returns the raw (encrypted) SNMP credentials for a device.
func (r *Repository) GetDeviceSNMP(ctx context.Context, id int64) (*models.DeviceSNMP, error) {
	query := `
		SELECT id, snmp_community, COALESCE(snmp_version, 'v2c'), snmp_v3_user,
		       snmp_v3_auth_proto, snmp_v3_auth_pass, snmp_v3_priv_proto, snmp_v3_priv_pass
		FROM devices WHERE id=$1`
	row := r.db.QueryRow(ctx, query, id)
	creds := &models.DeviceSNMP{}
	err := row.Scan(
		&creds.DeviceID, &creds.SNMPCommunity, &creds.SNMPVersion,
		&creds.SNMPV3User, &creds.SNMPV3AuthProto, &creds.SNMPV3AuthPass,
		&creds.SNMPV3PrivProto, &creds.SNMPV3PrivPass,
	)
	if err != nil {
		return nil, err
	}
	return creds, nil
}

// ListIPAddressesByDevice returns all IP addresses linked to a device.
func (r *Repository) ListIPAddressesByDevice(ctx context.Context, deviceID int64) ([]*models.IPAddress, error) {
	query := `
		SELECT id, subnet_id, address::text, hostname, status, assigned_to, created_at, updated_at,
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
			&ip.ID, &ip.SubnetID, &ip.Address, &ip.Hostname, &ip.Status, &ip.AssignedTo,
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

const ifaceSelectCols = `id, device_id, name, description, speed_mbps, media_type, vlan_id, ip_address_id, connected_to_device_id, connected_to_interface_id, created_at, updated_at`

func scanInterface(row pgx.Row) (*models.DeviceInterface, error) {
	i := &models.DeviceInterface{}
	err := row.Scan(
		&i.ID, &i.DeviceID, &i.Name, &i.Description, &i.SpeedMbps, &i.MediaType,
		&i.VLANID, &i.IPAddressID, &i.ConnectedToDeviceID, &i.ConnectedToInterfaceID,
		&i.CreatedAt, &i.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return i, nil
}

// ListDeviceInterfaces returns all interfaces for a device.
func (r *Repository) ListDeviceInterfaces(ctx context.Context, deviceID int64) ([]*models.DeviceInterface, error) {
	query := `SELECT ` + ifaceSelectCols + ` FROM device_interfaces WHERE device_id=$1 ORDER BY name`
	rows, err := r.db.Query(ctx, query, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ifaces := make([]*models.DeviceInterface, 0)
	for rows.Next() {
		i := &models.DeviceInterface{}
		err := rows.Scan(
			&i.ID, &i.DeviceID, &i.Name, &i.Description, &i.SpeedMbps, &i.MediaType,
			&i.VLANID, &i.IPAddressID, &i.ConnectedToDeviceID, &i.ConnectedToInterfaceID,
			&i.CreatedAt, &i.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		ifaces = append(ifaces, i)
	}
	return ifaces, rows.Err()
}

// GetDeviceInterface returns a single interface by ID.
func (r *Repository) GetDeviceInterface(ctx context.Context, id int64) (*models.DeviceInterface, error) {
	query := `SELECT ` + ifaceSelectCols + ` FROM device_interfaces WHERE id=$1`
	row := r.db.QueryRow(ctx, query, id)
	return scanInterface(row)
}

// CreateDeviceInterface creates a new interface on a device.
func (r *Repository) CreateDeviceInterface(ctx context.Context, deviceID int64, p *DeviceInterfaceParams) (*models.DeviceInterface, error) {
	query := `
		INSERT INTO device_interfaces (device_id, name, description, speed_mbps, media_type,
		                               vlan_id, ip_address_id, connected_to_device_id, connected_to_interface_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING ` + ifaceSelectCols
	row := r.db.QueryRow(ctx, query,
		deviceID, p.Name, p.Description, p.SpeedMbps, p.MediaType,
		p.VLANID, p.IPAddressID, p.ConnectedToDeviceID, p.ConnectedToInterfaceID,
	)
	return scanInterface(row)
}

// UpdateDeviceInterface updates an existing device interface.
func (r *Repository) UpdateDeviceInterface(ctx context.Context, deviceID, id int64, p *DeviceInterfaceParams) (*models.DeviceInterface, error) {
	query := `
		UPDATE device_interfaces SET
		  name=$1, description=$2, speed_mbps=$3, media_type=$4,
		  vlan_id=$5, ip_address_id=$6, connected_to_device_id=$7, connected_to_interface_id=$8,
		  updated_at=now()
		WHERE id=$9 AND device_id=$10
		RETURNING ` + ifaceSelectCols
	row := r.db.QueryRow(ctx, query,
		p.Name, p.Description, p.SpeedMbps, p.MediaType,
		p.VLANID, p.IPAddressID, p.ConnectedToDeviceID, p.ConnectedToInterfaceID,
		id, deviceID,
	)
	return scanInterface(row)
}

// DeleteDeviceInterface deletes an interface by ID, ensuring it belongs to the given device.
func (r *Repository) DeleteDeviceInterface(ctx context.Context, deviceID, id int64) error {
	ct, err := r.db.Exec(ctx, `DELETE FROM device_interfaces WHERE id=$1 AND device_id=$2`, id, deviceID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("interface not found")
	}
	return nil
}

// SetInterfaceConnection sets the reverse connection on a target interface.
func (r *Repository) SetInterfaceConnection(ctx context.Context, ifaceID, connDeviceID, connIfaceID int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE device_interfaces SET connected_to_device_id=$1, connected_to_interface_id=$2, updated_at=now() WHERE id=$3`,
		connDeviceID, connIfaceID, ifaceID,
	)
	return err
}

// ClearInterfaceConnection removes the connection from an interface.
func (r *Repository) ClearInterfaceConnection(ctx context.Context, ifaceID int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE device_interfaces SET connected_to_device_id=NULL, connected_to_interface_id=NULL, updated_at=now() WHERE id=$1`,
		ifaceID,
	)
	return err
}

// SearchDevices searches devices based on filter criteria.
func (r *Repository) SearchDevices(ctx context.Context, f *DeviceSearchFilter) ([]*models.Device, error) {
	args := []interface{}{}
	n := 1
	where := []string{}
	vlanJoin := ""

	if f.VLANID != nil {
		vlanJoin = fmt.Sprintf(` JOIN device_interfaces di ON di.device_id = d.id AND di.vlan_id = $%d`, n)
		args = append(args, *f.VLANID)
		n++
	}

	if f.Query != "" {
		where = append(where, fmt.Sprintf("(d.hostname ILIKE $%d OR d.description ILIKE $%d)", n, n))
		args = append(args, "%"+f.Query+"%")
		n++
	}
	if f.TypeID != nil {
		where = append(where, fmt.Sprintf("d.type_id = $%d", n))
		args = append(args, *f.TypeID)
		n++
	}
	if f.SectionID != nil {
		where = append(where, fmt.Sprintf("d.section_id = $%d", n))
		args = append(args, *f.SectionID)
		n++
	}
	if f.Vendor != nil && *f.Vendor != "" {
		where = append(where, fmt.Sprintf("d.vendor ILIKE $%d", n))
		args = append(args, "%"+*f.Vendor+"%")
		n++
	}
	if f.IsOnline != nil {
		where = append(where, fmt.Sprintf("d.is_online = $%d", n))
		args = append(args, *f.IsOnline)
		n++
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE "
		for i, w := range where {
			if i > 0 {
				whereClause += " AND "
			}
			whereClause += w
		}
	}

	// Use subquery to get distinct device IDs, then join for full data
	query := `
		SELECT ` + deviceSelectCols + `,
		       ` + deviceTypeSelectCols + `,
		       COUNT(ip.id) AS ip_count
		FROM devices d
		LEFT JOIN device_types dt ON dt.id = d.type_id
		LEFT JOIN ip_addresses ip ON ip.device_id = d.id` +
		vlanJoin +
		whereClause + `
		GROUP BY d.id, dt.id
		ORDER BY d.hostname`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	devices := make([]*models.Device, 0)
	for rows.Next() {
		d := &models.Device{}
		dt := &models.DeviceType{}
		var dtID *int64
		err := rows.Scan(
			&d.ID, &d.Hostname, &d.Description, &d.TypeID, &d.SectionID,
			&d.Vendor, &d.Model, &d.OSVersion, &d.IsOnline, &d.LastPingAt,
			&d.LocationID, &d.RackID, &d.RackUnitStart, &d.RackUnitSize,
			&d.CreatedAt, &d.UpdatedAt,
			&dtID, &dt.Name, &dt.Icon, &dt.Description, &dt.CreatedAt, &dt.UpdatedAt,
			&d.IPCount,
		)
		if err != nil {
			return nil, err
		}
		if dtID != nil {
			dt.ID = *dtID
			d.Type = dt
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

// GetSubnetTreeBySection returns all subnets for a section with utilisation counts, ordered by network address.
func (r *Repository) GetSubnetTreeBySection(ctx context.Context, sectionID int64) ([]models.SubnetTreeNode, error) {
	query := `
		SELECT
			s.id,
			host(s.network_address) || '/' || s.prefix_length AS cidr,
			s.description,
			COUNT(CASE WHEN ip.status = 'assigned' THEN 1 END) AS used,
			COUNT(ip.id) AS total
		FROM subnets s
		LEFT JOIN ip_addresses ip ON ip.subnet_id = s.id
		WHERE s.section_id = $1
		GROUP BY s.id, s.network_address, s.prefix_length, s.description
		ORDER BY s.network_address`

	rows, err := r.db.Query(ctx, query, sectionID)
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
			n.UtilisationPct = float64(n.Used) / float64(n.Total) * 100
		}
		nodes = append(nodes, n)
	}
	return nodes, rows.Err()
}

// SearchDevicesWithCustomFields searches devices and optionally filters by custom field values.
func (r *Repository) SearchDevicesWithCustomFields(ctx context.Context, f *DeviceSearchFilter, cfFilters map[string]string) ([]*models.Device, error) {
	if len(cfFilters) == 0 {
		return r.SearchDevices(ctx, f)
	}

	defs, err := r.ListCustomFieldDefinitions(ctx, "device")
	if err != nil {
		return nil, err
	}
	defByName := make(map[string]*models.CustomFieldDefinition, len(defs))
	for _, d := range defs {
		defByName[d.Name] = d
	}

	vlanJoin := ""
	if f.VLANID != nil {
		vlanJoin = " JOIN device_interfaces di ON di.device_id = d.id AND di.vlan_id = " + fmt.Sprintf("%d", *f.VLANID)
	}

	var where []string
	var args []interface{}
	n := 1

	if f.Query != "" {
		where = append(where, fmt.Sprintf("(d.hostname ILIKE $%d OR d.description ILIKE $%d)", n, n))
		args = append(args, "%"+f.Query+"%")
		n++
	}
	if f.TypeID != nil {
		where = append(where, fmt.Sprintf("d.type_id = $%d", n))
		args = append(args, *f.TypeID)
		n++
	}
	if f.SectionID != nil {
		where = append(where, fmt.Sprintf("d.section_id = $%d", n))
		args = append(args, *f.SectionID)
		n++
	}
	if f.Vendor != nil && *f.Vendor != "" {
		where = append(where, fmt.Sprintf("d.vendor ILIKE $%d", n))
		args = append(args, "%"+*f.Vendor+"%")
		n++
	}
	if f.IsOnline != nil {
		where = append(where, fmt.Sprintf("d.is_online = $%d", n))
		args = append(args, *f.IsOnline)
		n++
	}

	for fname, fval := range cfFilters {
		def, ok := defByName[fname]
		if !ok {
			continue
		}
		placeholder := fmt.Sprintf("$%d", n)
		n++
		if textLikeFieldTypes[def.FieldType] {
			where = append(where, fmt.Sprintf("EXISTS (SELECT 1 FROM custom_field_values cfv WHERE cfv.entity_type='device' AND cfv.entity_id=d.id AND cfv.definition_id=%d AND cfv.value ILIKE %s)", def.ID, placeholder))
			args = append(args, "%"+fval+"%")
		} else {
			where = append(where, fmt.Sprintf("EXISTS (SELECT 1 FROM custom_field_values cfv WHERE cfv.entity_type='device' AND cfv.entity_id=d.id AND cfv.definition_id=%d AND cfv.value = %s)", def.ID, placeholder))
			args = append(args, fval)
		}
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE " + strings.Join(where, " AND ")
	}

	query := `SELECT ` + deviceSelectCols + `,
		       ` + deviceTypeSelectCols + `,
		       COUNT(ip.id) AS ip_count
		FROM devices d
		LEFT JOIN device_types dt ON dt.id = d.type_id
		LEFT JOIN ip_addresses ip ON ip.device_id = d.id` +
		vlanJoin +
		whereClause + `
		GROUP BY d.id, dt.id
		ORDER BY d.hostname`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	devices := make([]*models.Device, 0)
	for rows.Next() {
		d := &models.Device{}
		dt := &models.DeviceType{}
		var dtID *int64
		err := rows.Scan(
			&d.ID, &d.Hostname, &d.Description, &d.TypeID, &d.SectionID,
			&d.Vendor, &d.Model, &d.OSVersion, &d.IsOnline, &d.LastPingAt,
			&d.LocationID, &d.RackID, &d.RackUnitStart, &d.RackUnitSize,
			&d.CreatedAt, &d.UpdatedAt,
			&dtID, &dt.Name, &dt.Icon, &dt.Description, &dt.CreatedAt, &dt.UpdatedAt,
			&d.IPCount,
		)
		if err != nil {
			return nil, err
		}
		if dtID != nil {
			dt.ID = *dtID
			d.Type = dt
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

// ListDevicesInRack returns devices assigned to a rack, ordered by rack_unit_start.
func (r *Repository) ListDevicesInRack(ctx context.Context, rackID int64) ([]*models.Device, error) {
	query := `
		SELECT ` + deviceSelectCols + `,
		       ` + deviceTypeSelectCols + `,
		       COUNT(ip.id) AS ip_count
		FROM devices d
		LEFT JOIN device_types dt ON dt.id = d.type_id
		LEFT JOIN ip_addresses ip ON ip.device_id = d.id
		WHERE d.rack_id = $1
		GROUP BY d.id, dt.id
		ORDER BY d.rack_unit_start NULLS LAST, d.hostname`

	rows, err := r.db.Query(ctx, query, rackID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	devices := make([]*models.Device, 0)
	for rows.Next() {
		d := &models.Device{}
		dt := &models.DeviceType{}
		var dtID *int64
		err := rows.Scan(
			&d.ID, &d.Hostname, &d.Description, &d.TypeID, &d.SectionID,
			&d.Vendor, &d.Model, &d.OSVersion, &d.IsOnline, &d.LastPingAt,
			&d.LocationID, &d.RackID, &d.RackUnitStart, &d.RackUnitSize,
			&d.CreatedAt, &d.UpdatedAt,
			&dtID, &dt.Name, &dt.Icon, &dt.Description, &dt.CreatedAt, &dt.UpdatedAt,
			&d.IPCount,
		)
		if err != nil {
			return nil, err
		}
		if dtID != nil {
			dt.ID = *dtID
			d.Type = dt
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

// ListDevicesByLocation returns devices assigned to a specific location.
func (r *Repository) ListDevicesByLocation(ctx context.Context, locationID int64) ([]*models.Device, error) {
	query := `
		SELECT ` + deviceSelectCols + `,
		       ` + deviceTypeSelectCols + `,
		       COUNT(ip.id) AS ip_count
		FROM devices d
		LEFT JOIN device_types dt ON dt.id = d.type_id
		LEFT JOIN ip_addresses ip ON ip.device_id = d.id
		WHERE d.location_id = $1
		GROUP BY d.id, dt.id
		ORDER BY d.hostname`

	rows, err := r.db.Query(ctx, query, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	devices := make([]*models.Device, 0)
	for rows.Next() {
		d := &models.Device{}
		dt := &models.DeviceType{}
		var dtID *int64
		err := rows.Scan(
			&d.ID, &d.Hostname, &d.Description, &d.TypeID, &d.SectionID,
			&d.Vendor, &d.Model, &d.OSVersion, &d.IsOnline, &d.LastPingAt,
			&d.LocationID, &d.RackID, &d.RackUnitStart, &d.RackUnitSize,
			&d.CreatedAt, &d.UpdatedAt,
			&dtID, &dt.Name, &dt.Icon, &dt.Description, &dt.CreatedAt, &dt.UpdatedAt,
			&d.IPCount,
		)
		if err != nil {
			return nil, err
		}
		if dtID != nil {
			dt.ID = *dtID
			d.Type = dt
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}
