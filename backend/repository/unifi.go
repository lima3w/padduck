package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"

	"padduck/models"
)

// FindVLANByNumber returns the VLAN with the given 802.1Q number (unique).
func (r *Repository) FindVLANByNumber(ctx context.Context, number int) (*models.VLAN, error) {
	vlan := &models.VLAN{}
	err := r.db.QueryRow(ctx, `SELECT id, vrf_id, domain_id, group_id, vlan_id, name, description, created_at, updated_at FROM vlans WHERE vlan_id = $1`, number).Scan(
		&vlan.ID, &vlan.VRFID, &vlan.DomainID, &vlan.GroupID, &vlan.VlanID, &vlan.Name, &vlan.Description, &vlan.CreatedAt, &vlan.UpdatedAt)
	return vlan, err
}

// UpsertUniFiVLANLink links a VLAN to a UniFi network. A UniFi network ID that
// was linked to another VLAN (its tag changed) is moved to this one.
func (r *Repository) UpsertUniFiVLANLink(ctx context.Context, vlanID int64, unifiID, name string) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM unifi_links WHERE vlan_id IS NOT NULL AND unifi_id=$1 AND vlan_id<>$2`, unifiID, vlanID); err != nil {
		return err
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO unifi_links (vlan_id, unifi_id, unifi_name) VALUES ($1,$2,$3)
		ON CONFLICT (vlan_id) WHERE vlan_id IS NOT NULL DO UPDATE SET
			unifi_id=EXCLUDED.unifi_id, unifi_name=EXCLUDED.unifi_name, last_synced_at=CURRENT_TIMESTAMP`,
		vlanID, unifiID, name)
	return err
}

// UpsertUniFiDeviceLink links a device to a UniFi device and stores its live details.
func (r *Repository) UpsertUniFiDeviceLink(ctx context.Context, deviceID int64, unifiID, name string, details any) error {
	raw, err := json.Marshal(details)
	if err != nil {
		return err
	}
	if _, err := r.db.Exec(ctx, `DELETE FROM unifi_links WHERE device_id IS NOT NULL AND unifi_id=$1 AND device_id<>$2`, unifiID, deviceID); err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `
		INSERT INTO unifi_links (device_id, unifi_id, unifi_name, details) VALUES ($1,$2,$3,$4)
		ON CONFLICT (device_id) WHERE device_id IS NOT NULL DO UPDATE SET
			unifi_id=EXCLUDED.unifi_id, unifi_name=EXCLUDED.unifi_name, details=EXCLUDED.details, last_synced_at=CURRENT_TIMESTAMP`,
		deviceID, unifiID, name, raw)
	return err
}

// ListUniFiLinks returns every VLAN and device link.
func (r *Repository) ListUniFiLinks(ctx context.Context) ([]*models.UniFiLink, error) {
	rows, err := r.db.Query(ctx, `SELECT id, vlan_id, device_id, unifi_id, unifi_name, details, last_synced_at FROM unifi_links ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*models.UniFiLink{}
	for rows.Next() {
		l := &models.UniFiLink{}
		if err := rows.Scan(&l.ID, &l.VLANID, &l.DeviceID, &l.UniFiID, &l.UniFiName, &l.Details, &l.LastSyncedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// FindDeviceIDByUniFiID returns the device linked to a UniFi device id.
func (r *Repository) FindDeviceIDByUniFiID(ctx context.Context, unifiID string) (int64, bool, error) {
	var id int64
	err := r.db.QueryRow(ctx, `SELECT device_id FROM unifi_links WHERE device_id IS NOT NULL AND unifi_id=$1`, unifiID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	return id, err == nil, err
}

// FindUnlinkedDeviceIDByHostname returns an existing device with this hostname
// (case-insensitive) that is not yet linked to a UniFi device.
func (r *Repository) FindUnlinkedDeviceIDByHostname(ctx context.Context, hostname string) (int64, bool, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		SELECT d.id FROM devices d
		WHERE lower(d.hostname)=lower($1)
		  AND NOT EXISTS (SELECT 1 FROM unifi_links l WHERE l.device_id=d.id)
		ORDER BY d.id LIMIT 1`, hostname).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	return id, err == nil, err
}

// UpdateDeviceFromUniFi applies UniFi-reported state to a linked device:
// online status, model and firmware always; vendor only when it is blank.
func (r *Repository) UpdateDeviceFromUniFi(ctx context.Context, id int64, vendor, model, osVersion string, online bool) error {
	_, err := r.db.Exec(ctx, `
		UPDATE devices SET
			vendor = COALESCE(NULLIF(vendor,''), NULLIF($2,'')),
			model = COALESCE(NULLIF($3,''), model),
			os_version = COALESCE(NULLIF($4,''), os_version),
			is_online = $5,
			last_ping_at = CASE WHEN $5 THEN CURRENT_TIMESTAMP ELSE last_ping_at END,
			updated_at = CURRENT_TIMESTAMP
		WHERE id=$1`, id, vendor, model, osVersion, online)
	return err
}

// FindIPInSubnet returns the IP record for an address within one subnet.
func (r *Repository) FindIPInSubnet(ctx context.Context, subnetID int64, address string) (*models.IPAddress, error) {
	var id int64
	if err := r.db.QueryRow(ctx, `SELECT id FROM ip_addresses WHERE subnet_id=$1 AND host(address)=$2`, subnetID, address).Scan(&id); err != nil {
		return nil, err
	}
	return r.GetIPAddressByID(ctx, id)
}

// SetIPHostname and SetIPMAC change a single field without touching the rest of the record.
func (r *Repository) SetIPHostname(ctx context.Context, id int64, hostname string) error {
	_, err := r.db.Exec(ctx, `UPDATE ip_addresses SET hostname=$2, updated_at=CURRENT_TIMESTAMP WHERE id=$1`, id, hostname)
	return err
}

func (r *Repository) SetIPMAC(ctx context.Context, id int64, mac string) error {
	_, err := r.db.Exec(ctx, `UPDATE ip_addresses SET mac_address=$2, updated_at=CURRENT_TIMESTAMP WHERE id=$1`, id, mac)
	return err
}

// CreateUniFiConflict records a difference for review. It does nothing when an
// identical open conflict exists, or when an admin already dismissed this exact
// UniFi value for the field, so dismissed differences do not keep reappearing.
// Reports whether a new conflict row was created.
func (r *Repository) CreateUniFiConflict(ctx context.Context, ipID int64, field, padduckValue, unifiValue string) (bool, error) {
	var dismissed bool
	if err := r.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM unifi_conflicts WHERE ip_id=$1 AND field=$2 AND status='dismissed' AND unifi_value=$3)`, ipID, field, unifiValue).Scan(&dismissed); err != nil {
		return false, err
	}
	if dismissed {
		return false, nil
	}
	tag, err := r.db.Exec(ctx, `
		INSERT INTO unifi_conflicts (ip_id, field, padduck_value, unifi_value) VALUES ($1,$2,$3,$4)
		ON CONFLICT (ip_id, field) WHERE status='open' DO UPDATE SET padduck_value=EXCLUDED.padduck_value, unifi_value=EXCLUDED.unifi_value
		WHERE unifi_conflicts.unifi_value<>EXCLUDED.unifi_value OR unifi_conflicts.padduck_value<>EXCLUDED.padduck_value`,
		ipID, field, padduckValue, unifiValue)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

const unifiConflictSelect = `SELECT c.id, c.ip_id, host(ip.address), ip.subnet_id, c.field, c.padduck_value, c.unifi_value, c.status, c.created_at, c.resolved_at FROM unifi_conflicts c JOIN ip_addresses ip ON ip.id=c.ip_id`

func scanUniFiConflict(row pgx.Row) (*models.UniFiConflict, error) {
	c := &models.UniFiConflict{}
	err := row.Scan(&c.ID, &c.IPID, &c.IPAddress, &c.SubnetID, &c.Field, &c.PadduckValue, &c.UniFiValue, &c.Status, &c.CreatedAt, &c.ResolvedAt)
	return c, err
}

// ListUniFiConflicts lists conflicts, optionally filtered by status ("" = all).
func (r *Repository) ListUniFiConflicts(ctx context.Context, status string) ([]*models.UniFiConflict, error) {
	q := unifiConflictSelect
	args := []any{}
	if status != "" {
		q += ` WHERE c.status=$1`
		args = append(args, status)
	}
	q += ` ORDER BY c.created_at DESC, c.id DESC`
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*models.UniFiConflict{}
	for rows.Next() {
		c, err := scanUniFiConflict(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) GetUniFiConflict(ctx context.Context, id int64) (*models.UniFiConflict, error) {
	return scanUniFiConflict(r.db.QueryRow(ctx, unifiConflictSelect+` WHERE c.id=$1`, id))
}

// ResolveUniFiConflict marks an open conflict accepted or dismissed.
func (r *Repository) ResolveUniFiConflict(ctx context.Context, id int64, status string) error {
	tag, err := r.db.Exec(ctx, `UPDATE unifi_conflicts SET status=$2, resolved_at=CURRENT_TIMESTAMP WHERE id=$1 AND status='open'`, id, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
