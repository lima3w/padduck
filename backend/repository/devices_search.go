package repository

import (
	"context"
	"fmt"
	"strings"

	"padduck/models"
)

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
	if f.NetworkID != nil {
		where = append(where, fmt.Sprintf("d.network_id = $%d", n))
		args = append(args, *f.NetworkID)
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
	if f.NetworkID != nil {
		where = append(where, fmt.Sprintf("d.network_id = $%d", n))
		args = append(args, *f.NetworkID)
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
