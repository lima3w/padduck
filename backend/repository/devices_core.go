package repository

import (
	"context"
	"fmt"

	"padduck/models"
)

// ---- Device management (v1.3.0) ----

// DeviceParams holds fields for creating or updating a device.
type DeviceParams struct {
	Hostname        string             `json:"hostname"`
	Description     *string            `json:"description"`
	TypeID          *int64             `json:"type_id"`
	NetworkID       *int64             `json:"network_id"`
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
	NetworkID *int64  `json:"network_id"`
	Vendor    *string `json:"vendor"`
	IsOnline  *bool   `json:"is_online"`
	VLANID    *int64  `json:"vlan_id"`
}

const deviceSelectCols = `d.id, d.hostname, d.description, d.type_id, d.network_id, d.vendor, d.model, d.os_version, d.is_online, d.last_ping_at, d.location_id, d.rack_id, d.rack_unit_start, d.rack_unit_size, d.created_at, d.updated_at`

// NULL-safe for LEFT JOINs against devices with no type: dt.id scans into a
// nullable temp that gates whether the type is attached, so the COALESCE
// fallbacks below are placeholders that are never exposed.
const deviceTypeSelectCols = `dt.id, COALESCE(dt.name, ''), COALESCE(dt.icon, ''), dt.description, COALESCE(dt.created_at, to_timestamp(0)), COALESCE(dt.updated_at, to_timestamp(0))`

func scanDevice(row interface{ Scan(dest ...any) error }) (*models.Device, error) {
	d := &models.Device{}
	err := row.Scan(
		&d.ID, &d.Hostname, &d.Description, &d.TypeID, &d.NetworkID,
		&d.Vendor, &d.Model, &d.OSVersion, &d.IsOnline, &d.LastPingAt,
		&d.LocationID, &d.RackID, &d.RackUnitStart, &d.RackUnitSize,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (r *Repository) ListDevicesWithOptions(ctx context.Context, opts ListOptions) ([]*models.Device, int64, error) {
	args := []interface{}{}
	where := ""
	if opts.Query != "" {
		args = append(args, "%"+opts.Query+"%")
		where = fmt.Sprintf(" WHERE d.hostname ILIKE $%d OR d.vendor ILIKE $%d OR d.model ILIKE $%d", len(args), len(args), len(args))
	}
	var total int64
	countQuery := `SELECT COUNT(*) FROM devices d` + where
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sortCol := sortExpr(opts.Sort, "d.hostname", map[string]string{"hostname": "d.hostname", "vendor": "d.vendor", "model": "d.model", "last_ping_at": "d.last_ping_at", "created_at": "d.created_at"})
	args = append(args, opts.Limit, opts.Offset)
	query := `
		SELECT ` + deviceSelectCols + `,
		       ` + deviceTypeSelectCols + `,
		       COUNT(ip.id) AS ip_count
		FROM devices d
		LEFT JOIN device_types dt ON dt.id = d.type_id
		LEFT JOIN ip_addresses ip ON ip.device_id = d.id
		` + where + `
		GROUP BY d.id, dt.id
		ORDER BY ` + sortCol + ` ` + orderDirection(opts.Order) + `
		LIMIT $` + fmt.Sprint(len(args)-1) + ` OFFSET $` + fmt.Sprint(len(args))

	rows, err := r.db.Query(ctx, query, args...)
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
			&d.ID, &d.Hostname, &d.Description, &d.TypeID, &d.NetworkID,
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
			&d.ID, &d.Hostname, &d.Description, &d.TypeID, &d.NetworkID,
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
		INSERT INTO devices (hostname, description, type_id, network_id, vendor, model, os_version,
		                     snmp_community, snmp_version, snmp_v3_user, snmp_v3_auth_proto,
		                     snmp_v3_auth_pass, snmp_v3_priv_proto, snmp_v3_priv_pass,
		                     location_id, rack_id, rack_unit_start, rack_unit_size)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
		RETURNING id, hostname, description, type_id, network_id, vendor, model, os_version,
		          is_online, last_ping_at, location_id, rack_id, rack_unit_start, rack_unit_size, created_at, updated_at`
	row := r.db.QueryRow(ctx, query,
		p.Hostname, p.Description, p.TypeID, p.NetworkID, p.Vendor, p.Model, p.OSVersion,
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
		&d.ID, &d.Hostname, &d.Description, &d.TypeID, &d.NetworkID,
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
		  hostname=$1, description=$2, type_id=$3, network_id=$4, vendor=$5, model=$6, os_version=$7,
		  snmp_community=$8, snmp_version=$9, snmp_v3_user=$10, snmp_v3_auth_proto=$11,
		  snmp_v3_auth_pass=$12, snmp_v3_priv_proto=$13, snmp_v3_priv_pass=$14,
		  location_id=$15, rack_id=$16, rack_unit_start=$17, rack_unit_size=$18,
		  updated_at=now()
		WHERE id=$19
		RETURNING id, hostname, description, type_id, network_id, vendor, model, os_version,
		          is_online, last_ping_at, location_id, rack_id, rack_unit_start, rack_unit_size, created_at, updated_at`
	row := r.db.QueryRow(ctx, query,
		p.Hostname, p.Description, p.TypeID, p.NetworkID, p.Vendor, p.Model, p.OSVersion,
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
			&d.ID, &d.Hostname, &d.Description, &d.TypeID, &d.NetworkID,
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
			&d.ID, &d.Hostname, &d.Description, &d.TypeID, &d.NetworkID,
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
