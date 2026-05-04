package repository

import (
	"context"

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
	query := `INSERT INTO users (username, email) VALUES ($1, $2) RETURNING id, username, email, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, username, email)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	query := `SELECT id, username, email, created_at, updated_at FROM users WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
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
	query := `INSERT INTO subnets (section_id, network_address, prefix_length, description) VALUES ($1, $2, $3, $4) RETURNING id, section_id, network_address, prefix_length, description, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, sectionID, networkAddress, prefixLength, description)

	subnet := &models.Subnet{}
	err := row.Scan(&subnet.ID, &subnet.SectionID, &subnet.NetworkAddress, &subnet.PrefixLength, &subnet.Description, &subnet.CreatedAt, &subnet.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return subnet, nil
}

func (r *Repository) GetSubnetByID(ctx context.Context, id int64) (*models.Subnet, error) {
	query := `SELECT id, section_id, network_address, prefix_length, description, created_at, updated_at FROM subnets WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	subnet := &models.Subnet{}
	err := row.Scan(&subnet.ID, &subnet.SectionID, &subnet.NetworkAddress, &subnet.PrefixLength, &subnet.Description, &subnet.CreatedAt, &subnet.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return subnet, nil
}

func (r *Repository) ListSubnetsBySection(ctx context.Context, sectionID int64) ([]*models.Subnet, error) {
	query := `SELECT id, section_id, network_address, prefix_length, description, created_at, updated_at FROM subnets WHERE section_id = $1 ORDER BY network_address`
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
	query := `UPDATE subnets SET description = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 RETURNING id, section_id, network_address, prefix_length, description, created_at, updated_at`
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
	query := `INSERT INTO ip_addresses (subnet_id, address, hostname, status, assigned_to) VALUES ($1, $2, $3, $4, $5) RETURNING id, subnet_id, address, hostname, status, assigned_to, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, subnetID, address, hostname, status, assignedTo)

	ip := &models.IPAddress{}
	err := row.Scan(&ip.ID, &ip.SubnetID, &ip.Address, &ip.Hostname, &ip.Status, &ip.AssignedTo, &ip.CreatedAt, &ip.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return ip, nil
}

func (r *Repository) GetIPAddressByID(ctx context.Context, id int64) (*models.IPAddress, error) {
	query := `SELECT id, subnet_id, address, hostname, status, assigned_to, created_at, updated_at FROM ip_addresses WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	ip := &models.IPAddress{}
	err := row.Scan(&ip.ID, &ip.SubnetID, &ip.Address, &ip.Hostname, &ip.Status, &ip.AssignedTo, &ip.CreatedAt, &ip.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return ip, nil
}

func (r *Repository) ListIPAddressesBySubnet(ctx context.Context, subnetID int64) ([]*models.IPAddress, error) {
	query := `SELECT id, subnet_id, address, hostname, status, assigned_to, created_at, updated_at FROM ip_addresses WHERE subnet_id = $1 ORDER BY address`
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
	query := `UPDATE ip_addresses SET status = $2, assigned_to = $3 WHERE id = $1 RETURNING id, subnet_id, address, hostname, status, assigned_to, created_at, updated_at`
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
