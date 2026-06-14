package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"padduck/models"
)

func (r *Repository) ListCircuitProviders(ctx context.Context) ([]*models.CircuitProvider, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, account_no, support_email, support_phone, portal_url, notes, created_at, updated_at FROM circuit_providers ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.CircuitProvider
	for rows.Next() {
		item := &models.CircuitProvider{}
		if err := rows.Scan(&item.ID, &item.Name, &item.AccountNo, &item.SupportEmail, &item.SupportPhone, &item.PortalURL, &item.Notes, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *Repository) GetCircuitProviderByID(ctx context.Context, id int64) (*models.CircuitProvider, error) {
	item := &models.CircuitProvider{}
	err := r.db.QueryRow(ctx, `SELECT id, name, account_no, support_email, support_phone, portal_url, notes, created_at, updated_at FROM circuit_providers WHERE id=$1`, id).Scan(&item.ID, &item.Name, &item.AccountNo, &item.SupportEmail, &item.SupportPhone, &item.PortalURL, &item.Notes, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) CreateCircuitProvider(ctx context.Context, p *CircuitProviderParams) (*models.CircuitProvider, error) {
	var id int64
	err := r.db.QueryRow(ctx, `INSERT INTO circuit_providers (name, account_no, support_email, support_phone, portal_url, notes) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`, p.Name, p.AccountNo, p.SupportEmail, p.SupportPhone, p.PortalURL, p.Notes).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetCircuitProviderByID(ctx, id)
}

func (r *Repository) UpdateCircuitProvider(ctx context.Context, id int64, p *CircuitProviderParams) (*models.CircuitProvider, error) {
	tag, err := r.db.Exec(ctx, `UPDATE circuit_providers SET name=$1, account_no=$2, support_email=$3, support_phone=$4, portal_url=$5, notes=$6, updated_at=CURRENT_TIMESTAMP WHERE id=$7`, p.Name, p.AccountNo, p.SupportEmail, p.SupportPhone, p.PortalURL, p.Notes, id)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, pgx.ErrNoRows
	}
	return r.GetCircuitProviderByID(ctx, id)
}

func (r *Repository) DeleteCircuitProvider(ctx context.Context, id int64) error {
	return deleteByID(ctx, r, "circuit_providers", id)
}

func (r *Repository) ListPhysicalCircuits(ctx context.Context) ([]*models.PhysicalCircuit, error) {
	rows, err := r.db.Query(ctx, `SELECT pc.id, pc.provider_id, pc.circuit_id, pc.name, pc.type, pc.status, pc.bandwidth_mbps, pc.location_a_id, pc.location_b_id, pc.customer_id, pc.install_date, cp.name, la.name, lb.name, c.name, pc.notes, pc.created_at, pc.updated_at FROM physical_circuits pc JOIN circuit_providers cp ON cp.id=pc.provider_id LEFT JOIN locations la ON la.id=pc.location_a_id LEFT JOIN locations lb ON lb.id=pc.location_b_id LEFT JOIN customers c ON c.id=pc.customer_id ORDER BY pc.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.PhysicalCircuit
	for rows.Next() {
		item := &models.PhysicalCircuit{}
		if err := rows.Scan(&item.ID, &item.ProviderID, &item.CircuitID, &item.Name, &item.Type, &item.Status, &item.BandwidthMbps, &item.LocationAID, &item.LocationBID, &item.CustomerID, &item.InstallDate, &item.ProviderName, &item.LocationAName, &item.LocationBName, &item.CustomerName, &item.Notes, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *Repository) GetPhysicalCircuitByID(ctx context.Context, id int64) (*models.PhysicalCircuit, error) {
	item := &models.PhysicalCircuit{}
	err := r.db.QueryRow(ctx, `SELECT pc.id, pc.provider_id, pc.circuit_id, pc.name, pc.type, pc.status, pc.bandwidth_mbps, pc.location_a_id, pc.location_b_id, pc.customer_id, pc.install_date, cp.name, la.name, lb.name, c.name, pc.notes, pc.created_at, pc.updated_at FROM physical_circuits pc JOIN circuit_providers cp ON cp.id=pc.provider_id LEFT JOIN locations la ON la.id=pc.location_a_id LEFT JOIN locations lb ON lb.id=pc.location_b_id LEFT JOIN customers c ON c.id=pc.customer_id WHERE pc.id=$1`, id).Scan(&item.ID, &item.ProviderID, &item.CircuitID, &item.Name, &item.Type, &item.Status, &item.BandwidthMbps, &item.LocationAID, &item.LocationBID, &item.CustomerID, &item.InstallDate, &item.ProviderName, &item.LocationAName, &item.LocationBName, &item.CustomerName, &item.Notes, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) CreatePhysicalCircuit(ctx context.Context, p *PhysicalCircuitParams) (*models.PhysicalCircuit, error) {
	var id int64
	err := r.db.QueryRow(ctx, `INSERT INTO physical_circuits (provider_id, circuit_id, name, type, status, bandwidth_mbps, location_a_id, location_b_id, customer_id, install_date, notes) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id`, p.ProviderID, p.CircuitID, p.Name, p.Type, p.Status, p.BandwidthMbps, p.LocationAID, p.LocationBID, p.CustomerID, p.InstallDate, p.Notes).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetPhysicalCircuitByID(ctx, id)
}

func (r *Repository) UpdatePhysicalCircuit(ctx context.Context, id int64, p *PhysicalCircuitParams) (*models.PhysicalCircuit, error) {
	tag, err := r.db.Exec(ctx, `UPDATE physical_circuits SET provider_id=$1, circuit_id=$2, name=$3, type=$4, status=$5, bandwidth_mbps=$6, location_a_id=$7, location_b_id=$8, customer_id=$9, install_date=$10, notes=$11, updated_at=CURRENT_TIMESTAMP WHERE id=$12`, p.ProviderID, p.CircuitID, p.Name, p.Type, p.Status, p.BandwidthMbps, p.LocationAID, p.LocationBID, p.CustomerID, p.InstallDate, p.Notes, id)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, pgx.ErrNoRows
	}
	return r.GetPhysicalCircuitByID(ctx, id)
}

func (r *Repository) DeletePhysicalCircuit(ctx context.Context, id int64) error {
	return deleteByID(ctx, r, "physical_circuits", id)
}

func (r *Repository) ListLogicalCircuits(ctx context.Context) ([]*models.LogicalCircuit, error) {
	rows, err := r.db.Query(ctx, `SELECT lc.id, lc.physical_circuit_id, lc.name, lc.service_id, lc.type, lc.status, lc.vlan_id, lc.vrf_id, lc.customer_id, lc.bandwidth_mbps, pc.name, c.name, lc.notes, lc.created_at, lc.updated_at FROM logical_circuits lc LEFT JOIN physical_circuits pc ON pc.id=lc.physical_circuit_id LEFT JOIN customers c ON c.id=lc.customer_id ORDER BY lc.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.LogicalCircuit
	for rows.Next() {
		item := &models.LogicalCircuit{}
		if err := rows.Scan(&item.ID, &item.PhysicalCircuitID, &item.Name, &item.ServiceID, &item.Type, &item.Status, &item.VLANID, &item.VRFID, &item.CustomerID, &item.BandwidthMbps, &item.PhysicalCircuitName, &item.CustomerName, &item.Notes, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *Repository) GetLogicalCircuitByID(ctx context.Context, id int64) (*models.LogicalCircuit, error) {
	item := &models.LogicalCircuit{}
	err := r.db.QueryRow(ctx, `SELECT lc.id, lc.physical_circuit_id, lc.name, lc.service_id, lc.type, lc.status, lc.vlan_id, lc.vrf_id, lc.customer_id, lc.bandwidth_mbps, pc.name, c.name, lc.notes, lc.created_at, lc.updated_at FROM logical_circuits lc LEFT JOIN physical_circuits pc ON pc.id=lc.physical_circuit_id LEFT JOIN customers c ON c.id=lc.customer_id WHERE lc.id=$1`, id).Scan(&item.ID, &item.PhysicalCircuitID, &item.Name, &item.ServiceID, &item.Type, &item.Status, &item.VLANID, &item.VRFID, &item.CustomerID, &item.BandwidthMbps, &item.PhysicalCircuitName, &item.CustomerName, &item.Notes, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) CreateLogicalCircuit(ctx context.Context, p *LogicalCircuitParams) (*models.LogicalCircuit, error) {
	var id int64
	err := r.db.QueryRow(ctx, `INSERT INTO logical_circuits (physical_circuit_id, name, service_id, type, status, vlan_id, vrf_id, customer_id, bandwidth_mbps, notes) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`, p.PhysicalCircuitID, p.Name, p.ServiceID, p.Type, p.Status, p.VLANID, p.VRFID, p.CustomerID, p.BandwidthMbps, p.Notes).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetLogicalCircuitByID(ctx, id)
}

func (r *Repository) UpdateLogicalCircuit(ctx context.Context, id int64, p *LogicalCircuitParams) (*models.LogicalCircuit, error) {
	tag, err := r.db.Exec(ctx, `UPDATE logical_circuits SET physical_circuit_id=$1, name=$2, service_id=$3, type=$4, status=$5, vlan_id=$6, vrf_id=$7, customer_id=$8, bandwidth_mbps=$9, notes=$10, updated_at=CURRENT_TIMESTAMP WHERE id=$11`, p.PhysicalCircuitID, p.Name, p.ServiceID, p.Type, p.Status, p.VLANID, p.VRFID, p.CustomerID, p.BandwidthMbps, p.Notes, id)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, pgx.ErrNoRows
	}
	return r.GetLogicalCircuitByID(ctx, id)
}

func (r *Repository) DeleteLogicalCircuit(ctx context.Context, id int64) error {
	return deleteByID(ctx, r, "logical_circuits", id)
}

func (r *Repository) ListCustomerAssociations(ctx context.Context, customerID int64) ([]*models.CustomerAssociation, error) {
	query := `SELECT ca.id, ca.customer_id, ca.object_type, ca.object_id, ca.object_name, ca.relationship, ca.notes, c.name, ca.created_at, ca.updated_at FROM customer_associations ca JOIN customers c ON c.id=ca.customer_id`
	args := []any{}
	if customerID > 0 {
		query += ` WHERE ca.customer_id=$1`
		args = append(args, customerID)
	}
	query += ` ORDER BY ca.object_type, ca.object_id`
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.CustomerAssociation
	for rows.Next() {
		item := &models.CustomerAssociation{}
		if err := rows.Scan(&item.ID, &item.CustomerID, &item.ObjectType, &item.ObjectID, &item.ObjectName, &item.Relationship, &item.Notes, &item.CustomerName, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *Repository) CreateCustomerAssociation(ctx context.Context, p *CustomerAssociationParams) (*models.CustomerAssociation, error) {
	var id int64
	err := r.db.QueryRow(ctx, `INSERT INTO customer_associations (customer_id, object_type, object_id, object_name, relationship, notes) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`, p.CustomerID, p.ObjectType, p.ObjectID, p.ObjectName, p.Relationship, p.Notes).Scan(&id)
	if err != nil {
		return nil, err
	}
	item := &models.CustomerAssociation{}
	err = r.db.QueryRow(ctx, `SELECT ca.id, ca.customer_id, ca.object_type, ca.object_id, ca.object_name, ca.relationship, ca.notes, c.name, ca.created_at, ca.updated_at FROM customer_associations ca JOIN customers c ON c.id=ca.customer_id WHERE ca.id=$1`, id).Scan(&item.ID, &item.CustomerID, &item.ObjectType, &item.ObjectID, &item.ObjectName, &item.Relationship, &item.Notes, &item.CustomerName, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *Repository) DeleteCustomerAssociation(ctx context.Context, id int64) error {
	return deleteByID(ctx, r, "customer_associations", id)
}
