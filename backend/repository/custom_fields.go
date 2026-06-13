package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"padduck/models"
)

// Custom field operations

// CustomFieldDefinitionParams holds input for creating or updating a custom field definition
type CustomFieldDefinitionParams struct {
	EntityType   string          `json:"entity_type"`
	Name         string          `json:"name"`
	Label        string          `json:"label"`
	FieldType    string          `json:"field_type"`
	Options      json.RawMessage `json:"options"`
	IsRequired   bool            `json:"is_required"`
	DefaultValue *string         `json:"default_value"`
	Placeholder  *string         `json:"placeholder"`
	DisplayOrder int             `json:"display_order"`
	IsSearchable bool            `json:"is_searchable"`
}

func scanCustomFieldDefinition(row interface {
	Scan(dest ...any) error
}) (*models.CustomFieldDefinition, error) {
	d := &models.CustomFieldDefinition{}
	var options []byte
	err := row.Scan(&d.ID, &d.EntityType, &d.Name, &d.Label, &d.FieldType, &options,
		&d.IsRequired, &d.DefaultValue, &d.Placeholder, &d.DisplayOrder, &d.IsSearchable, &d.CreatedAt)
	if err != nil {
		return nil, err
	}
	if options != nil {
		var v interface{}
		if err := json.Unmarshal(options, &v); err == nil {
			d.Options = v
		}
	}
	return d, nil
}

const cfdSelectCols = `id, entity_type, name, label, field_type, options, is_required, default_value, placeholder, display_order, is_searchable, created_at`

func (r *Repository) ListCustomFieldDefinitions(ctx context.Context, entityType string) ([]*models.CustomFieldDefinition, error) {
	q := `SELECT ` + cfdSelectCols + ` FROM custom_field_definitions`
	args := []interface{}{}
	if entityType != "" {
		q += ` WHERE entity_type = $1`
		args = append(args, entityType)
	}
	q += ` ORDER BY display_order ASC`

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	defs := make([]*models.CustomFieldDefinition, 0)
	for rows.Next() {
		d, err := scanCustomFieldDefinition(rows)
		if err != nil {
			return nil, err
		}
		defs = append(defs, d)
	}
	return defs, rows.Err()
}

func (r *Repository) CreateCustomFieldDefinition(ctx context.Context, p *CustomFieldDefinitionParams) (*models.CustomFieldDefinition, error) {
	q := `INSERT INTO custom_field_definitions (entity_type, name, label, field_type, options, is_required, default_value, placeholder, display_order, is_searchable)
	      VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	      RETURNING ` + cfdSelectCols
	row := r.db.QueryRow(ctx, q, p.EntityType, p.Name, p.Label, p.FieldType, p.Options, p.IsRequired, p.DefaultValue, p.Placeholder, p.DisplayOrder, p.IsSearchable)
	return scanCustomFieldDefinition(row)
}

func (r *Repository) GetCustomFieldDefinition(ctx context.Context, id int64) (*models.CustomFieldDefinition, error) {
	q := `SELECT ` + cfdSelectCols + ` FROM custom_field_definitions WHERE id = $1`
	row := r.db.QueryRow(ctx, q, id)
	d, err := scanCustomFieldDefinition(row)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("custom field definition not found")
		}
		return nil, err
	}
	return d, nil
}

func (r *Repository) UpdateCustomFieldDefinition(ctx context.Context, id int64, p *CustomFieldDefinitionParams) (*models.CustomFieldDefinition, error) {
	q := `UPDATE custom_field_definitions
	      SET entity_type=$1, name=$2, label=$3, field_type=$4, options=$5, is_required=$6,
	          default_value=$7, placeholder=$8, display_order=$9, is_searchable=$10
	      WHERE id=$11
	      RETURNING ` + cfdSelectCols
	row := r.db.QueryRow(ctx, q, p.EntityType, p.Name, p.Label, p.FieldType, p.Options, p.IsRequired, p.DefaultValue, p.Placeholder, p.DisplayOrder, p.IsSearchable, id)
	d, err := scanCustomFieldDefinition(row)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("custom field definition not found")
		}
		return nil, err
	}
	return d, nil
}

func (r *Repository) DeleteCustomFieldDefinition(ctx context.Context, id int64) error {
	res, err := r.db.Exec(ctx, `DELETE FROM custom_field_definitions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("custom field definition not found")
	}
	return nil
}

func (r *Repository) ReorderCustomFieldDefinitions(ctx context.Context, ids []int64) error {
	for i, id := range ids {
		_, err := r.db.Exec(ctx, `UPDATE custom_field_definitions SET display_order = $1 WHERE id = $2`, i, id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) GetCustomFieldValues(ctx context.Context, entityType string, entityID int64) (map[string]*string, error) {
	q := `SELECT cfd.name, cfv.value
	      FROM custom_field_values cfv
	      JOIN custom_field_definitions cfd ON cfd.id = cfv.definition_id
	      WHERE cfv.entity_type = $1 AND cfv.entity_id = $2`
	rows, err := r.db.Query(ctx, q, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]*string)
	for rows.Next() {
		var name string
		var value *string
		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		result[name] = value
	}
	return result, rows.Err()
}

func (r *Repository) SetCustomFieldValues(ctx context.Context, entityType string, entityID int64, defs []*models.CustomFieldDefinition, values map[string]*string) error {
	for _, def := range defs {
		val, ok := values[def.Name]
		if !ok {
			continue
		}
		_, err := r.db.Exec(ctx,
			`INSERT INTO custom_field_values (definition_id, entity_id, entity_type, value)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (definition_id, entity_id, entity_type) DO UPDATE SET value = EXCLUDED.value`,
			def.ID, entityID, entityType, val)
		if err != nil {
			return err
		}
	}
	return nil
}

// textLikeFieldTypes is the set of field types where ILIKE should be used in search
var textLikeFieldTypes = map[string]bool{
	"text": true, "textarea": true, "url": true, "email": true,
}

// SearchSubnetsWithCustomFields searches subnets and optionally filters by custom field values.
func (r *Repository) SearchSubnetsWithCustomFields(ctx context.Context, networkID int64, query string, limit, offset int64, cfFilters map[string]string) ([]*models.Subnet, error) {
	if len(cfFilters) == 0 {
		return r.SearchSubnets(ctx, networkID, query, limit, offset)
	}

	defs, err := r.ListCustomFieldDefinitions(ctx, "subnet")
	if err != nil {
		return nil, err
	}
	defByName := make(map[string]*models.CustomFieldDefinition, len(defs))
	for _, d := range defs {
		defByName[d.Name] = d
	}

	base := `SELECT ` + subnetSelectCols + ` ` + subnetFromJoin + `
	         WHERE s.network_id = $1 AND (host(s.network_address) ILIKE $2 OR s.description ILIKE $2)`
	args := []interface{}{networkID, "%" + query + "%"}
	n := 3

	for fname, fval := range cfFilters {
		def, ok := defByName[fname]
		if !ok {
			continue
		}
		placeholder := fmt.Sprintf("$%d", n)
		n++
		if textLikeFieldTypes[def.FieldType] {
			base += fmt.Sprintf(` AND EXISTS (SELECT 1 FROM custom_field_values cfv WHERE cfv.entity_type='subnet' AND cfv.entity_id=s.id AND cfv.definition_id=%d AND cfv.value ILIKE %s)`, def.ID, placeholder)
			args = append(args, "%"+fval+"%")
		} else {
			base += fmt.Sprintf(` AND EXISTS (SELECT 1 FROM custom_field_values cfv WHERE cfv.entity_type='subnet' AND cfv.entity_id=s.id AND cfv.definition_id=%d AND cfv.value = %s)`, def.ID, placeholder)
			args = append(args, fval)
		}
	}

	base += fmt.Sprintf(` ORDER BY s.network_address ASC LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, base, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subnets := make([]*models.Subnet, 0)
	for rows.Next() {
		s, err := scanSubnet(rows)
		if err != nil {
			return nil, err
		}
		subnets = append(subnets, s)
	}
	return subnets, rows.Err()
}

// SearchIPAddressesWithCustomFields searches IP addresses and optionally filters by custom field values.
func (r *Repository) SearchIPAddressesWithCustomFields(ctx context.Context, subnetID int64, query, status string, limit, offset int64, filter IPSearchFilter, cfFilters map[string]string) ([]*models.IPAddress, error) {
	if len(cfFilters) == 0 {
		return r.SearchIPAddresses(ctx, subnetID, query, status, limit, offset, filter)
	}

	defs, err := r.ListCustomFieldDefinitions(ctx, "ip_address")
	if err != nil {
		return nil, err
	}
	defByName := make(map[string]*models.CustomFieldDefinition, len(defs))
	for _, d := range defs {
		defByName[d.Name] = d
	}

	sql := `SELECT ` + ipSelectCols + ` ` + ipFromJoin + `
	        WHERE ip.subnet_id = $1 AND (ip.address::text ILIKE $2 OR ip.hostname ILIKE $2)`
	args := []interface{}{subnetID, "%" + query + "%"}
	n := 3

	if status != "" {
		sql += fmt.Sprintf(" AND ip.status = $%d", n)
		args = append(args, status)
		n++
	}
	if filter.TagID != nil {
		sql += fmt.Sprintf(" AND ip.tag_id = $%d", n)
		args = append(args, *filter.TagID)
		n++
	}
	if filter.MACAddress != "" {
		sql += fmt.Sprintf(" AND ip.mac_address ILIKE $%d", n)
		args = append(args, "%"+filter.MACAddress+"%")
		n++
	}
	if filter.PTRRecord != "" {
		sql += fmt.Sprintf(" AND ip.ptr_record ILIKE $%d", n)
		args = append(args, "%"+filter.PTRRecord+"%")
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
			sql += fmt.Sprintf(` AND EXISTS (SELECT 1 FROM custom_field_values cfv WHERE cfv.entity_type='ip_address' AND cfv.entity_id=ip.id AND cfv.definition_id=%d AND cfv.value ILIKE %s)`, def.ID, placeholder)
			args = append(args, "%"+fval+"%")
		} else {
			sql += fmt.Sprintf(` AND EXISTS (SELECT 1 FROM custom_field_values cfv WHERE cfv.entity_type='ip_address' AND cfv.entity_id=ip.id AND cfv.definition_id=%d AND cfv.value = %s)`, def.ID, placeholder)
			args = append(args, fval)
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
