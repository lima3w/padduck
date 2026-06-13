package repository

import (
	"context"
	"fmt"

	"padduck/models"
)

const ifaceSelectCols = `id, device_id, name, description, speed_mbps, media_type, vlan_id, ip_address_id, connected_to_device_id, connected_to_interface_id, created_at, updated_at`

func scanInterface(row interface{ Scan(dest ...any) error }) (*models.DeviceInterface, error) {
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
