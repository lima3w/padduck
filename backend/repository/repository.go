package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"ipam-next/models"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Ping verifies database connectivity
func (r *Repository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

// User operations

func (r *Repository) CreateUser(ctx context.Context, username, email string) (*models.User, error) {
	query := `INSERT INTO users (username, email, role) VALUES ($1, $2, 'user') RETURNING id, username, email, role, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, username, email)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	query := `SELECT id, username, email, role, created_at, updated_at FROM users WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Section operations

func (r *Repository) CreateSection(ctx context.Context, name, description string, createdBy int64) (*models.Section, error) {
	query := `INSERT INTO sections (name, description, created_by) VALUES ($1, $2, $3) RETURNING id, name, description, created_by, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, name, description, createdBy)

	section := &models.Section{}
	err := row.Scan(&section.ID, &section.Name, &section.Description, &section.CreatedBy, &section.CreatedAt, &section.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return section, nil
}

func (r *Repository) GetSectionByID(ctx context.Context, id int64) (*models.Section, error) {
	query := `SELECT id, name, description, created_by, created_at, updated_at FROM sections WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	section := &models.Section{}
	err := row.Scan(&section.ID, &section.Name, &section.Description, &section.CreatedBy, &section.CreatedAt, &section.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return section, nil
}

func (r *Repository) ListAllSections(ctx context.Context) ([]*models.Section, error) {
	query := `SELECT id, name, description, created_by, created_at, updated_at FROM sections ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sections := make([]*models.Section, 0)
	for rows.Next() {
		section := &models.Section{}
		err := rows.Scan(&section.ID, &section.Name, &section.Description, &section.CreatedBy, &section.CreatedAt, &section.UpdatedAt)
		if err != nil {
			return nil, err
		}
		sections = append(sections, section)
	}
	return sections, rows.Err()
}

func (r *Repository) UpdateSection(ctx context.Context, id int64, name, description string) (*models.Section, error) {
	query := `UPDATE sections SET name = $2, description = $3 WHERE id = $1 RETURNING id, name, description, created_by, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, id, name, description)

	section := &models.Section{}
	err := row.Scan(&section.ID, &section.Name, &section.Description, &section.CreatedBy, &section.CreatedAt, &section.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return section, nil
}

func (r *Repository) DeleteSection(ctx context.Context, id int64) error {
	query := `DELETE FROM sections WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// Subnet operations

func (r *Repository) CreateSubnet(ctx context.Context, sectionID int64, networkAddress string, prefixLength int, description string) (*models.Subnet, error) {
	query := `INSERT INTO subnets (section_id, network_address, prefix_length, description) VALUES ($1, $2, $3, $4) RETURNING id, section_id, network_address::text, prefix_length, description, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, sectionID, networkAddress, prefixLength, description)

	subnet := &models.Subnet{}
	err := row.Scan(&subnet.ID, &subnet.SectionID, &subnet.NetworkAddress, &subnet.PrefixLength, &subnet.Description, &subnet.CreatedAt, &subnet.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return subnet, nil
}

func (r *Repository) GetSubnetByID(ctx context.Context, id int64) (*models.Subnet, error) {
	query := `SELECT id, section_id, network_address::text, prefix_length, description, created_at, updated_at FROM subnets WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	subnet := &models.Subnet{}
	err := row.Scan(&subnet.ID, &subnet.SectionID, &subnet.NetworkAddress, &subnet.PrefixLength, &subnet.Description, &subnet.CreatedAt, &subnet.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return subnet, nil
}

func (r *Repository) ListSubnetsBySection(ctx context.Context, sectionID int64) ([]*models.Subnet, error) {
	query := `SELECT id, section_id, network_address::text, prefix_length, description, created_at, updated_at FROM subnets WHERE section_id = $1 ORDER BY network_address`
	rows, err := r.db.Query(ctx, query, sectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subnets := make([]*models.Subnet, 0)
	for rows.Next() {
		subnet := &models.Subnet{}
		err := rows.Scan(&subnet.ID, &subnet.SectionID, &subnet.NetworkAddress, &subnet.PrefixLength, &subnet.Description, &subnet.CreatedAt, &subnet.UpdatedAt)
		if err != nil {
			return nil, err
		}
		subnets = append(subnets, subnet)
	}
	return subnets, rows.Err()
}

func (r *Repository) UpdateSubnet(ctx context.Context, id int64, description string) (*models.Subnet, error) {
	query := `UPDATE subnets SET description = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 RETURNING id, section_id, network_address::text, prefix_length, description, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, description, id)

	subnet := &models.Subnet{}
	err := row.Scan(&subnet.ID, &subnet.SectionID, &subnet.NetworkAddress, &subnet.PrefixLength, &subnet.Description, &subnet.CreatedAt, &subnet.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return subnet, nil
}

func (r *Repository) DeleteSubnet(ctx context.Context, id int64) error {
	query := `DELETE FROM subnets WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// IP Address operations

func (r *Repository) CreateIPAddress(ctx context.Context, subnetID int64, address, hostname string, status string, assignedTo *string) (*models.IPAddress, error) {
	query := `INSERT INTO ip_addresses (subnet_id, address, hostname, status, assigned_to) VALUES ($1, $2, $3, $4, $5) RETURNING id, subnet_id, address::text, hostname, status, assigned_to, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, subnetID, address, hostname, status, assignedTo)

	ip := &models.IPAddress{}
	err := row.Scan(&ip.ID, &ip.SubnetID, &ip.Address, &ip.Hostname, &ip.Status, &ip.AssignedTo, &ip.CreatedAt, &ip.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return ip, nil
}

func (r *Repository) GetIPAddressByID(ctx context.Context, id int64) (*models.IPAddress, error) {
	query := `SELECT id, subnet_id, address::text, hostname, status, assigned_to, created_at, updated_at FROM ip_addresses WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	ip := &models.IPAddress{}
	err := row.Scan(&ip.ID, &ip.SubnetID, &ip.Address, &ip.Hostname, &ip.Status, &ip.AssignedTo, &ip.CreatedAt, &ip.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return ip, nil
}

func (r *Repository) ListIPAddressesBySubnet(ctx context.Context, subnetID int64) ([]*models.IPAddress, error) {
	query := `SELECT id, subnet_id, address::text, hostname, status, assigned_to, created_at, updated_at FROM ip_addresses WHERE subnet_id = $1 ORDER BY address`
	rows, err := r.db.Query(ctx, query, subnetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ips := make([]*models.IPAddress, 0)
	for rows.Next() {
		ip := &models.IPAddress{}
		err := rows.Scan(&ip.ID, &ip.SubnetID, &ip.Address, &ip.Hostname, &ip.Status, &ip.AssignedTo, &ip.CreatedAt, &ip.UpdatedAt)
		if err != nil {
			return nil, err
		}
		ips = append(ips, ip)
	}
	return ips, rows.Err()
}

func (r *Repository) UpdateIPAddressStatus(ctx context.Context, id int64, status string, assignedTo *string) (*models.IPAddress, error) {
	query := `UPDATE ip_addresses SET status = $2, assigned_to = $3 WHERE id = $1 RETURNING id, subnet_id, address::text, hostname, status, assigned_to, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, id, status, assignedTo)

	ip := &models.IPAddress{}
	err := row.Scan(&ip.ID, &ip.SubnetID, &ip.Address, &ip.Hostname, &ip.Status, &ip.AssignedTo, &ip.CreatedAt, &ip.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return ip, nil
}

func (r *Repository) DeleteIPAddress(ctx context.Context, id int64) error {
	query := `DELETE FROM ip_addresses WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *Repository) ListAvailableIPsBySubnet(ctx context.Context, subnetID int64) ([]*models.IPAddress, error) {
	query := `SELECT id, subnet_id, address::text, hostname, status, assigned_to, created_at, updated_at FROM ip_addresses WHERE subnet_id = $1 AND status = 'available' ORDER BY address`
	rows, err := r.db.Query(ctx, query, subnetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ips := make([]*models.IPAddress, 0)
	for rows.Next() {
		ip := &models.IPAddress{}
		err := rows.Scan(&ip.ID, &ip.SubnetID, &ip.Address, &ip.Hostname, &ip.Status, &ip.AssignedTo, &ip.CreatedAt, &ip.UpdatedAt)
		if err != nil {
			return nil, err
		}
		ips = append(ips, ip)
	}
	return ips, rows.Err()
}

func (r *Repository) GetPool() *pgxpool.Pool {
	return r.db
}

// AllocateIPAddress atomically finds and assigns the next available IP
// Uses a transaction with SERIALIZABLE isolation to prevent duplicate allocation
func (r *Repository) AllocateIPAddress(ctx context.Context, subnetID int64, assignedTo string) (*models.IPAddress, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Find the first available IP in the subnet (ordered by address)
	query := `SELECT id, subnet_id, address::text, hostname, status, assigned_to, created_at, updated_at
	          FROM ip_addresses
	          WHERE subnet_id = $1 AND status = 'available'
	          ORDER BY address LIMIT 1`
	row := tx.QueryRow(ctx, query, subnetID)

	ip := &models.IPAddress{}
	err = row.Scan(&ip.ID, &ip.SubnetID, &ip.Address, &ip.Hostname, &ip.Status, &ip.AssignedTo, &ip.CreatedAt, &ip.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Atomically update the IP status to 'assigned'
	updateQuery := `UPDATE ip_addresses SET status = 'assigned', assigned_to = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 RETURNING id, subnet_id, address::text, hostname, status, assigned_to, created_at, updated_at`
	updateRow := tx.QueryRow(ctx, updateQuery, assignedTo, ip.ID)

	err = updateRow.Scan(&ip.ID, &ip.SubnetID, &ip.Address, &ip.Hostname, &ip.Status, &ip.AssignedTo, &ip.CreatedAt, &ip.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return ip, nil
}

// CountIPsByStatus counts IPs in a subnet by their status
func (r *Repository) CountIPsByStatus(ctx context.Context, subnetID int64, status string) (int64, error) {
	query := `SELECT COUNT(*) FROM ip_addresses WHERE subnet_id = $1 AND status = $2`
	row := r.db.QueryRow(ctx, query, subnetID, status)

	var count int64
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CountTotalIPsBySubnet counts all IPs in a subnet
func (r *Repository) CountTotalIPsBySubnet(ctx context.Context, subnetID int64) (int64, error) {
	query := `SELECT COUNT(*) FROM ip_addresses WHERE subnet_id = $1`
	row := r.db.QueryRow(ctx, query, subnetID)

	var count int64
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// UpdateIPAddressWithLease updates IP with lease information
func (r *Repository) UpdateIPAddressWithLease(ctx context.Context, id int64, status string, assignedTo *string, assignedAt *time.Time, expiresAt *time.Time) (*models.IPAddress, error) {
	query := `UPDATE ip_addresses SET status = $2, assigned_to = $3, assigned_at = $4, expires_at = $5, updated_at = CURRENT_TIMESTAMP WHERE id = $1 RETURNING id, subnet_id, address::text, hostname, status, assigned_to, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, id, status, assignedTo, assignedAt, expiresAt)

	ip := &models.IPAddress{}
	err := row.Scan(&ip.ID, &ip.SubnetID, &ip.Address, &ip.Hostname, &ip.Status, &ip.AssignedTo, &ip.CreatedAt, &ip.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return ip, nil
}

// API Token operations

func (r *Repository) CreateAPIToken(ctx context.Context, userID int64, tokenHash, name string) (*models.APIToken, error) {
	query := `INSERT INTO api_tokens (user_id, token_hash, name) VALUES ($1, $2, $3) RETURNING id, user_id, token_hash, name, last_used_at, expires_at, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, userID, tokenHash, name)

	token := &models.APIToken{}
	err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.Name, &token.LastUsedAt, &token.ExpiresAt, &token.CreatedAt, &token.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *Repository) GetAPITokenByHash(ctx context.Context, tokenHash string) (*models.APIToken, error) {
	query := `SELECT id, user_id, token_hash, name, last_used_at, expires_at, created_at, updated_at FROM api_tokens WHERE token_hash = $1`
	row := r.db.QueryRow(ctx, query, tokenHash)

	token := &models.APIToken{}
	err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.Name, &token.LastUsedAt, &token.ExpiresAt, &token.CreatedAt, &token.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *Repository) ListAPITokensByUser(ctx context.Context, userID int64) ([]*models.APIToken, error) {
	query := `SELECT id, user_id, token_hash, name, last_used_at, expires_at, created_at, updated_at FROM api_tokens WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens := make([]*models.APIToken, 0)
	for rows.Next() {
		token := &models.APIToken{}
		err := rows.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.Name, &token.LastUsedAt, &token.ExpiresAt, &token.CreatedAt, &token.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}

func (r *Repository) UpdateAPITokenLastUsed(ctx context.Context, tokenID int64) error {
	query := `UPDATE api_tokens SET last_used_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, tokenID)
	return err
}

func (r *Repository) DeleteAPIToken(ctx context.Context, tokenID int64) error {
	query := `DELETE FROM api_tokens WHERE id = $1`
	_, err := r.db.Exec(ctx, query, tokenID)
	return err
}

// Search operations

func (r *Repository) SearchSections(ctx context.Context, query string, limit, offset int64) ([]*models.Section, error) {
	sql := `SELECT id, name, description, created_by, created_at, updated_at FROM sections
	        WHERE name ILIKE $1 OR description ILIKE $1
	        ORDER BY created_at DESC
	        LIMIT $2 OFFSET $3`
	searchQuery := "%" + query + "%"
	rows, err := r.db.Query(ctx, sql, searchQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sections := make([]*models.Section, 0)
	for rows.Next() {
		section := &models.Section{}
		err := rows.Scan(&section.ID, &section.Name, &section.Description, &section.CreatedBy, &section.CreatedAt, &section.UpdatedAt)
		if err != nil {
			return nil, err
		}
		sections = append(sections, section)
	}
	return sections, rows.Err()
}

func (r *Repository) SearchSubnets(ctx context.Context, sectionID int64, query string, limit, offset int64) ([]*models.Subnet, error) {
	sql := `SELECT id, section_id, network_address::text, prefix_length, description, created_at, updated_at FROM subnets
	        WHERE section_id = $1 AND (network_address::text ILIKE $2 OR description ILIKE $2)
	        ORDER BY network_address ASC
	        LIMIT $3 OFFSET $4`
	searchQuery := "%" + query + "%"
	rows, err := r.db.Query(ctx, sql, sectionID, searchQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subnets := make([]*models.Subnet, 0)
	for rows.Next() {
		subnet := &models.Subnet{}
		err := rows.Scan(&subnet.ID, &subnet.SectionID, &subnet.NetworkAddress, &subnet.PrefixLength, &subnet.Description, &subnet.CreatedAt, &subnet.UpdatedAt)
		if err != nil {
			return nil, err
		}
		subnets = append(subnets, subnet)
	}
	return subnets, rows.Err()
}

func (r *Repository) SearchIPAddresses(ctx context.Context, subnetID int64, query string, status string, limit, offset int64) ([]*models.IPAddress, error) {
	sql := `SELECT id, subnet_id, address::text, hostname, status, assigned_to, created_at, updated_at FROM ip_addresses
	        WHERE subnet_id = $1 AND (address::text ILIKE $2 OR hostname ILIKE $2 OR assigned_to ILIKE $2)`
	args := []interface{}{subnetID, "%" + query + "%"}

	if status != "" {
		sql += " AND status = $3"
		args = append(args, status)
		sql += " ORDER BY address ASC LIMIT $4 OFFSET $5"
		args = append(args, limit, offset)
	} else {
		sql += " ORDER BY address ASC LIMIT $3 OFFSET $4"
		args = append(args, limit, offset)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ips := make([]*models.IPAddress, 0)
	for rows.Next() {
		ip := &models.IPAddress{}
		err := rows.Scan(&ip.ID, &ip.SubnetID, &ip.Address, &ip.Hostname, &ip.Status, &ip.AssignedTo, &ip.CreatedAt, &ip.UpdatedAt)
		if err != nil {
			return nil, err
		}
		ips = append(ips, ip)
	}
	return ips, rows.Err()
}

// VRF operations

func (r *Repository) CreateVRF(ctx context.Context, name, rd, description string) (*models.VRF, error) {
	query := `INSERT INTO vrfs (name, route_distinguisher, description)
	          VALUES ($1, $2, $3)
	          RETURNING id, name, route_distinguisher, description, created_at, updated_at`
	vrf := &models.VRF{}
	err := r.db.QueryRow(ctx, query, name, rd, description).Scan(
		&vrf.ID, &vrf.Name, &vrf.RouteDistinguisher, &vrf.Description, &vrf.CreatedAt, &vrf.UpdatedAt,
	)
	return vrf, err
}

func (r *Repository) GetVRFByID(ctx context.Context, id int64) (*models.VRF, error) {
	query := `SELECT id, name, route_distinguisher, description, created_at, updated_at FROM vrfs WHERE id = $1`
	vrf := &models.VRF{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&vrf.ID, &vrf.Name, &vrf.RouteDistinguisher, &vrf.Description, &vrf.CreatedAt, &vrf.UpdatedAt,
	)
	return vrf, err
}

func (r *Repository) ListAllVRFs(ctx context.Context) ([]*models.VRF, error) {
	query := `SELECT id, name, route_distinguisher, description, created_at, updated_at FROM vrfs ORDER BY name ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vrfs := make([]*models.VRF, 0)
	for rows.Next() {
		vrf := &models.VRF{}
		err := rows.Scan(&vrf.ID, &vrf.Name, &vrf.RouteDistinguisher, &vrf.Description, &vrf.CreatedAt, &vrf.UpdatedAt)
		if err != nil {
			return nil, err
		}
		vrfs = append(vrfs, vrf)
	}
	return vrfs, rows.Err()
}

func (r *Repository) UpdateVRF(ctx context.Context, id int64, name, rd, description string) (*models.VRF, error) {
	query := `UPDATE vrfs SET name = $1, route_distinguisher = $2, description = $3, updated_at = CURRENT_TIMESTAMP
	          WHERE id = $4
	          RETURNING id, name, route_distinguisher, description, created_at, updated_at`
	vrf := &models.VRF{}
	err := r.db.QueryRow(ctx, query, name, rd, description, id).Scan(
		&vrf.ID, &vrf.Name, &vrf.RouteDistinguisher, &vrf.Description, &vrf.CreatedAt, &vrf.UpdatedAt,
	)
	return vrf, err
}

func (r *Repository) DeleteVRF(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM vrfs WHERE id = $1`, id)
	return err
}

// VLAN operations

func (r *Repository) CreateVLAN(ctx context.Context, vrfID *int64, vlanID int, name, description string) (*models.VLAN, error) {
	query := `INSERT INTO vlans (vrf_id, vlan_id, name, description)
	          VALUES ($1, $2, $3, $4)
	          RETURNING id, vrf_id, vlan_id, name, description, created_at, updated_at`
	vlan := &models.VLAN{}
	err := r.db.QueryRow(ctx, query, vrfID, vlanID, name, description).Scan(
		&vlan.ID, &vlan.VRFID, &vlan.VlanID, &vlan.Name, &vlan.Description, &vlan.CreatedAt, &vlan.UpdatedAt,
	)
	return vlan, err
}

func (r *Repository) GetVLANByID(ctx context.Context, id int64) (*models.VLAN, error) {
	query := `SELECT id, vrf_id, vlan_id, name, description, created_at, updated_at FROM vlans WHERE id = $1`
	vlan := &models.VLAN{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&vlan.ID, &vlan.VRFID, &vlan.VlanID, &vlan.Name, &vlan.Description, &vlan.CreatedAt, &vlan.UpdatedAt,
	)
	return vlan, err
}

func (r *Repository) ListAllVLANs(ctx context.Context) ([]*models.VLAN, error) {
	query := `SELECT id, vrf_id, vlan_id, name, description, created_at, updated_at FROM vlans ORDER BY vlan_id ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vlans := make([]*models.VLAN, 0)
	for rows.Next() {
		vlan := &models.VLAN{}
		err := rows.Scan(&vlan.ID, &vlan.VRFID, &vlan.VlanID, &vlan.Name, &vlan.Description, &vlan.CreatedAt, &vlan.UpdatedAt)
		if err != nil {
			return nil, err
		}
		vlans = append(vlans, vlan)
	}
	return vlans, rows.Err()
}

func (r *Repository) ListVLANsByVRF(ctx context.Context, vrfID int64) ([]*models.VLAN, error) {
	query := `SELECT id, vrf_id, vlan_id, name, description, created_at, updated_at FROM vlans WHERE vrf_id = $1 ORDER BY vlan_id ASC`
	rows, err := r.db.Query(ctx, query, vrfID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vlans := make([]*models.VLAN, 0)
	for rows.Next() {
		vlan := &models.VLAN{}
		err := rows.Scan(&vlan.ID, &vlan.VRFID, &vlan.VlanID, &vlan.Name, &vlan.Description, &vlan.CreatedAt, &vlan.UpdatedAt)
		if err != nil {
			return nil, err
		}
		vlans = append(vlans, vlan)
	}
	return vlans, rows.Err()
}

func (r *Repository) UpdateVLAN(ctx context.Context, id int64, name, description string) (*models.VLAN, error) {
	query := `UPDATE vlans SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP
	          WHERE id = $3
	          RETURNING id, vrf_id, vlan_id, name, description, created_at, updated_at`
	vlan := &models.VLAN{}
	err := r.db.QueryRow(ctx, query, name, description, id).Scan(
		&vlan.ID, &vlan.VRFID, &vlan.VlanID, &vlan.Name, &vlan.Description, &vlan.CreatedAt, &vlan.UpdatedAt,
	)
	return vlan, err
}

func (r *Repository) DeleteVLAN(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM vlans WHERE id = $1`, id)
	return err
}
