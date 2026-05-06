package repository

import (
	"context"
	"fmt"
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
	query := `INSERT INTO users (username, email, role) VALUES ($1, $2, 'user') RETURNING id, username, email, password_hash, role, state, last_login_at, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, username, email)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, role, state, last_login_at, created_at, updated_at FROM users WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, role, state, last_login_at, created_at, updated_at FROM users WHERE username = $1`
	row := r.db.QueryRow(ctx, query, username)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) ListAllUsers(ctx context.Context) ([]*models.User, error) {
	query := `SELECT id, username, email, password_hash, role, state, last_login_at, created_at, updated_at FROM users ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*models.User, 0)
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *Repository) CreateUserWithPassword(ctx context.Context, username, email, passwordHash, role string) (*models.User, error) {
	query := `INSERT INTO users (username, email, password_hash, role) VALUES ($1, $2, $3, $4) RETURNING id, username, email, password_hash, role, state, last_login_at, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, username, email, passwordHash, role)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) CreateUserWithState(ctx context.Context, username, email, passwordHash, role, state string) (*models.User, error) {
	query := `INSERT INTO users (username, email, password_hash, role, state) VALUES ($1, $2, $3, $4, $5) RETURNING id, username, email, password_hash, role, state, last_login_at, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, username, email, passwordHash, role, state)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) UpdateUserState(ctx context.Context, userID int64, state string) error {
	query := `UPDATE users SET state = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID, state)
	return err
}

func (r *Repository) UpdateUserRole(ctx context.Context, userID int64, role string) (*models.User, error) {
	query := `UPDATE users SET role = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1 RETURNING id, username, email, password_hash, role, state, last_login_at, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, userID, role)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) DeleteUser(ctx context.Context, userID int64) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

func (r *Repository) UpdateLastLogin(ctx context.Context, userID int64) error {
	query := `UPDATE users SET last_login_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, role, state, last_login_at, created_at, updated_at FROM users WHERE email = $1`
	row := r.db.QueryRow(ctx, query, email)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) CreatePasswordReset(ctx context.Context, userID int64, tokenHash string) (*models.PasswordReset, error) {
	query := `INSERT INTO password_resets (user_id, token_hash, expires_at) VALUES ($1, $2, CURRENT_TIMESTAMP + INTERVAL '1 hour') RETURNING id, user_id, token_hash, expires_at, used_at, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, userID, tokenHash)

	reset := &models.PasswordReset{}
	err := row.Scan(&reset.ID, &reset.UserID, &reset.TokenHash, &reset.ExpiresAt, &reset.UsedAt, &reset.CreatedAt, &reset.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return reset, nil
}

func (r *Repository) GetPasswordResetByToken(ctx context.Context, tokenHash string) (*models.PasswordReset, error) {
	query := `SELECT id, user_id, token_hash, expires_at, used_at, created_at, updated_at FROM password_resets WHERE token_hash = $1`
	row := r.db.QueryRow(ctx, query, tokenHash)

	reset := &models.PasswordReset{}
	err := row.Scan(&reset.ID, &reset.UserID, &reset.TokenHash, &reset.ExpiresAt, &reset.UsedAt, &reset.CreatedAt, &reset.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return reset, nil
}

func (r *Repository) MarkPasswordResetAsUsed(ctx context.Context, resetID int64) error {
	query := `UPDATE password_resets SET used_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, resetID)
	return err
}

func (r *Repository) UpdateUserPassword(ctx context.Context, userID int64, passwordHash string) error {
	query := `UPDATE users SET password_hash = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID, passwordHash)
	return err
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

// Session operations

func (r *Repository) CreateSession(ctx context.Context, userID int64, tokenHash, deviceName, ipAddress, userAgent string, absoluteExpiresAt time.Time) (*models.Session, error) {
	query := `INSERT INTO sessions (user_id, token_hash, device_name, ip_address, user_agent, absolute_expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, token_hash, device_name, ip_address, user_agent, last_used_at, absolute_expires_at, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, userID, tokenHash, deviceName, ipAddress, userAgent, absoluteExpiresAt)

	s := &models.Session{}
	err := row.Scan(&s.ID, &s.UserID, &s.TokenHash, &s.DeviceName, &s.IPAddress, &s.UserAgent, &s.LastUsedAt, &s.AbsoluteExpiresAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) GetSessionByHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	query := `SELECT id, user_id, token_hash, device_name, ip_address, user_agent, last_used_at, absolute_expires_at, created_at, updated_at FROM sessions WHERE token_hash = $1`
	row := r.db.QueryRow(ctx, query, tokenHash)

	s := &models.Session{}
	err := row.Scan(&s.ID, &s.UserID, &s.TokenHash, &s.DeviceName, &s.IPAddress, &s.UserAgent, &s.LastUsedAt, &s.AbsoluteExpiresAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) ListSessionsByUser(ctx context.Context, userID int64) ([]*models.Session, error) {
	query := `SELECT id, user_id, token_hash, device_name, ip_address, user_agent, last_used_at, absolute_expires_at, created_at, updated_at FROM sessions WHERE user_id = $1 ORDER BY last_used_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]*models.Session, 0)
	for rows.Next() {
		s := &models.Session{}
		err := rows.Scan(&s.ID, &s.UserID, &s.TokenHash, &s.DeviceName, &s.IPAddress, &s.UserAgent, &s.LastUsedAt, &s.AbsoluteExpiresAt, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func (r *Repository) UpdateSessionLastUsed(ctx context.Context, sessionID int64) error {
	query := `UPDATE sessions SET last_used_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, sessionID)
	return err
}

func (r *Repository) DeleteSession(ctx context.Context, sessionID int64) error {
	query := `DELETE FROM sessions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, sessionID)
	return err
}

func (r *Repository) DeleteSessionByHash(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM sessions WHERE token_hash = $1`
	_, err := r.db.Exec(ctx, query, tokenHash)
	return err
}

func (r *Repository) DeleteAllUserSessions(ctx context.Context, userID int64) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

func (r *Repository) DeleteExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE absolute_expires_at < CURRENT_TIMESTAMP`
	_, err := r.db.Exec(ctx, query)
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

// Config operations

func (r *Repository) GetConfig(ctx context.Context, key string) (*models.Config, error) {
	query := `SELECT key, value, created_at, updated_at FROM configs WHERE key = $1`
	cfg := &models.Config{}
	err := r.db.QueryRow(ctx, query, key).Scan(&cfg.Key, &cfg.Value, &cfg.CreatedAt, &cfg.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (r *Repository) ListConfigs(ctx context.Context) ([]*models.Config, error) {
	query := `SELECT key, value, created_at, updated_at FROM configs ORDER BY key ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	configs := make([]*models.Config, 0)
	for rows.Next() {
		cfg := &models.Config{}
		if err := rows.Scan(&cfg.Key, &cfg.Value, &cfg.CreatedAt, &cfg.UpdatedAt); err != nil {
			return nil, err
		}
		configs = append(configs, cfg)
	}
	return configs, rows.Err()
}

func (r *Repository) SetConfig(ctx context.Context, key, value string) error {
	query := `INSERT INTO configs (key, value) VALUES ($1, $2)
	          ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = CURRENT_TIMESTAMP`
	_, err := r.db.Exec(ctx, query, key, value)
	return err
}

// Email verification operations

func (r *Repository) CreateEmailVerification(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) (*models.EmailVerification, error) {
	query := `INSERT INTO email_verifications (user_id, token_hash, expires_at) VALUES ($1, $2, $3)
	          ON CONFLICT (token_hash) DO NOTHING
	          RETURNING id, user_id, token_hash, expires_at, used_at, created_at, updated_at`
	ev := &models.EmailVerification{}
	err := r.db.QueryRow(ctx, query, userID, tokenHash, expiresAt).Scan(
		&ev.ID, &ev.UserID, &ev.TokenHash, &ev.ExpiresAt, &ev.UsedAt, &ev.CreatedAt, &ev.UpdatedAt,
	)
	return ev, err
}

func (r *Repository) GetEmailVerificationByToken(ctx context.Context, tokenHash string) (*models.EmailVerification, error) {
	query := `SELECT id, user_id, token_hash, expires_at, used_at, created_at, updated_at FROM email_verifications WHERE token_hash = $1`
	ev := &models.EmailVerification{}
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&ev.ID, &ev.UserID, &ev.TokenHash, &ev.ExpiresAt, &ev.UsedAt, &ev.CreatedAt, &ev.UpdatedAt,
	)
	return ev, err
}

func (r *Repository) MarkEmailVerificationUsed(ctx context.Context, verificationID int64) error {
	query := `UPDATE email_verifications SET used_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, verificationID)
	return err
}

func (r *Repository) DeleteEmailVerificationsByUser(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM email_verifications WHERE user_id = $1`, userID)
	return err
}

// User approval operations

func (r *Repository) CreateUserApproval(ctx context.Context, userID int64) (*models.UserApproval, error) {
	query := `INSERT INTO user_approvals (user_id) VALUES ($1) RETURNING id, user_id, status, reviewed_by, reviewed_at, rejection_reason, created_at, updated_at`
	ua := &models.UserApproval{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&ua.ID, &ua.UserID, &ua.Status, &ua.ReviewedBy, &ua.ReviewedAt, &ua.RejectionReason, &ua.CreatedAt, &ua.UpdatedAt,
	)
	return ua, err
}

func (r *Repository) GetUserApprovalByUserID(ctx context.Context, userID int64) (*models.UserApproval, error) {
	query := `SELECT id, user_id, status, reviewed_by, reviewed_at, rejection_reason, created_at, updated_at FROM user_approvals WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`
	ua := &models.UserApproval{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&ua.ID, &ua.UserID, &ua.Status, &ua.ReviewedBy, &ua.ReviewedAt, &ua.RejectionReason, &ua.CreatedAt, &ua.UpdatedAt,
	)
	return ua, err
}

func (r *Repository) ListPendingApprovals(ctx context.Context) ([]*models.UserApproval, error) {
	query := `SELECT id, user_id, status, reviewed_by, reviewed_at, rejection_reason, created_at, updated_at FROM user_approvals WHERE status = 'pending' ORDER BY created_at ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	approvals := make([]*models.UserApproval, 0)
	for rows.Next() {
		ua := &models.UserApproval{}
		if err := rows.Scan(&ua.ID, &ua.UserID, &ua.Status, &ua.ReviewedBy, &ua.ReviewedAt, &ua.RejectionReason, &ua.CreatedAt, &ua.UpdatedAt); err != nil {
			return nil, err
		}
		approvals = append(approvals, ua)
	}
	return approvals, rows.Err()
}

func (r *Repository) UpdateUserApproval(ctx context.Context, approvalID int64, status string, reviewedBy int64, rejectionReason *string) error {
	query := `UPDATE user_approvals SET status = $2, reviewed_by = $3, reviewed_at = CURRENT_TIMESTAMP, rejection_reason = $4, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, approvalID, status, reviewedBy, rejectionReason)
	return err
}

func (r *Repository) GetUserApprovalByID(ctx context.Context, approvalID int64) (*models.UserApproval, error) {
	query := `SELECT id, user_id, status, reviewed_by, reviewed_at, rejection_reason, created_at, updated_at FROM user_approvals WHERE id = $1`
	ua := &models.UserApproval{}
	err := r.db.QueryRow(ctx, query, approvalID).Scan(
		&ua.ID, &ua.UserID, &ua.Status, &ua.ReviewedBy, &ua.ReviewedAt, &ua.RejectionReason, &ua.CreatedAt, &ua.UpdatedAt,
	)
	return ua, err
}

// MFA settings operations

func (r *Repository) GetMFASettings(ctx context.Context, userID int64) (*models.UserMFASettings, error) {
	query := `SELECT id, user_id, totp_enabled, backup_codes_generated_at, created_at, updated_at FROM user_mfa_settings WHERE user_id = $1`
	s := &models.UserMFASettings{}
	err := r.db.QueryRow(ctx, query, userID).Scan(&s.ID, &s.UserID, &s.TOTPEnabled, &s.BackupCodesGeneratedAt, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (r *Repository) UpsertMFASettings(ctx context.Context, userID int64, totpEnabled bool, backupCodesAt *time.Time) error {
	query := `INSERT INTO user_mfa_settings (user_id, totp_enabled, backup_codes_generated_at)
	          VALUES ($1, $2, $3)
	          ON CONFLICT (user_id) DO UPDATE SET totp_enabled = $2, backup_codes_generated_at = $3, updated_at = CURRENT_TIMESTAMP`
	_, err := r.db.Exec(ctx, query, userID, totpEnabled, backupCodesAt)
	return err
}

// TOTP secret operations

func (r *Repository) UpsertTOTPSecret(ctx context.Context, userID int64, encryptedSecret []byte) error {
	query := `INSERT INTO user_totp_secrets (user_id, encrypted_secret, verified)
	          VALUES ($1, $2, FALSE)
	          ON CONFLICT (user_id) DO UPDATE SET encrypted_secret = $2, verified = FALSE, updated_at = CURRENT_TIMESTAMP`
	_, err := r.db.Exec(ctx, query, userID, encryptedSecret)
	return err
}

func (r *Repository) GetTOTPSecret(ctx context.Context, userID int64) (*models.UserTOTPSecret, error) {
	query := `SELECT id, user_id, encrypted_secret, verified, created_at, updated_at FROM user_totp_secrets WHERE user_id = $1`
	s := &models.UserTOTPSecret{}
	err := r.db.QueryRow(ctx, query, userID).Scan(&s.ID, &s.UserID, &s.EncryptedSecret, &s.Verified, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (r *Repository) MarkTOTPVerified(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx, `UPDATE user_totp_secrets SET verified = TRUE, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1`, userID)
	return err
}

func (r *Repository) DeleteTOTPSecret(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM user_totp_secrets WHERE user_id = $1`, userID)
	return err
}

// Backup code operations

func (r *Repository) CreateBackupCodes(ctx context.Context, userID int64, hashes []string) error {
	// Delete existing codes first
	if _, err := r.db.Exec(ctx, `DELETE FROM user_backup_codes WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for _, h := range hashes {
		if _, err := r.db.Exec(ctx, `INSERT INTO user_backup_codes (user_id, code_hash) VALUES ($1, $2)`, userID, h); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ListBackupCodes(ctx context.Context, userID int64) ([]*models.UserBackupCode, error) {
	rows, err := r.db.Query(ctx, `SELECT id, user_id, code_hash, used, used_at, created_at FROM user_backup_codes WHERE user_id = $1 ORDER BY id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var codes []*models.UserBackupCode
	for rows.Next() {
		c := &models.UserBackupCode{}
		if err := rows.Scan(&c.ID, &c.UserID, &c.CodeHash, &c.Used, &c.UsedAt, &c.CreatedAt); err != nil {
			return nil, err
		}
		codes = append(codes, c)
	}
	return codes, rows.Err()
}

func (r *Repository) MarkBackupCodeUsed(ctx context.Context, codeID int64) error {
	_, err := r.db.Exec(ctx, `UPDATE user_backup_codes SET used = TRUE, used_at = CURRENT_TIMESTAMP WHERE id = $1`, codeID)
	return err
}

// MFA challenge operations

func (r *Repository) CreateMFAChallenge(ctx context.Context, userID int64, challengeHash string, expiresAt time.Time) (*models.MFAChallenge, error) {
	query := `INSERT INTO mfa_challenges (user_id, challenge_hash, expires_at) VALUES ($1, $2, $3) RETURNING id, user_id, challenge_hash, expires_at, completed_at, created_at`
	c := &models.MFAChallenge{}
	err := r.db.QueryRow(ctx, query, userID, challengeHash, expiresAt).Scan(&c.ID, &c.UserID, &c.ChallengeHash, &c.ExpiresAt, &c.CompletedAt, &c.CreatedAt)
	return c, err
}

func (r *Repository) GetMFAChallenge(ctx context.Context, challengeHash string) (*models.MFAChallenge, error) {
	query := `SELECT id, user_id, challenge_hash, expires_at, completed_at, created_at FROM mfa_challenges WHERE challenge_hash = $1`
	c := &models.MFAChallenge{}
	err := r.db.QueryRow(ctx, query, challengeHash).Scan(&c.ID, &c.UserID, &c.ChallengeHash, &c.ExpiresAt, &c.CompletedAt, &c.CreatedAt)
	return c, err
}

func (r *Repository) CompleteMFAChallenge(ctx context.Context, challengeID int64) error {
	_, err := r.db.Exec(ctx, `UPDATE mfa_challenges SET completed_at = CURRENT_TIMESTAMP WHERE id = $1`, challengeID)
	return err
}

// Login attempt operations

func (r *Repository) CreateLoginAttempt(ctx context.Context, username, ipAddress, userAgent string, success bool, failureReason string) error {
	query := `INSERT INTO login_attempts (username, ip_address, user_agent, success, failure_reason) VALUES ($1, $2::inet, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, username, nullableString(ipAddress), userAgent, success, nullableString(failureReason))
	return err
}

func (r *Repository) CountRecentFailedAttemptsByUsername(ctx context.Context, username string, since time.Time) (int, error) {
	query := `SELECT COUNT(*) FROM login_attempts WHERE username = $1 AND success = false AND created_at >= $2`
	var count int
	err := r.db.QueryRow(ctx, query, username, since).Scan(&count)
	return count, err
}

func (r *Repository) CountRecentFailedAttemptsByIP(ctx context.Context, username, ipAddress string, since time.Time) (int, error) {
	query := `SELECT COUNT(*) FROM login_attempts WHERE username = $1 AND ip_address = $2::inet AND success = false AND created_at >= $3`
	var count int
	err := r.db.QueryRow(ctx, query, username, ipAddress, since).Scan(&count)
	return count, err
}

func (r *Repository) GetLoginHistory(ctx context.Context, username string, limit int) ([]*models.LoginAttempt, error) {
	query := `SELECT id, username, COALESCE(ip_address::text, ''), COALESCE(user_agent, ''), success, COALESCE(failure_reason, ''), created_at
	          FROM login_attempts WHERE username = $1 ORDER BY created_at DESC LIMIT $2`
	rows, err := r.db.Query(ctx, query, username, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attempts := make([]*models.LoginAttempt, 0)
	for rows.Next() {
		a := &models.LoginAttempt{}
		if err := rows.Scan(&a.ID, &a.Username, &a.IPAddress, &a.UserAgent, &a.Success, &a.FailureReason, &a.CreatedAt); err != nil {
			return nil, err
		}
		attempts = append(attempts, a)
	}
	return attempts, rows.Err()
}

// Account lockout operations

func (r *Repository) CreateAccountLockout(ctx context.Context, userID int64, unlockAt time.Time, reason string, lockoutCount int) (*models.AccountLockout, error) {
	query := `INSERT INTO account_lockouts (user_id, unlock_at, reason, lockout_count)
	          VALUES ($1, $2, $3, $4)
	          RETURNING id, user_id, locked_at, unlock_at, unlock_token_hash, unlock_token_expires_at, unlock_token_used_at, reason, lockout_count, unlocked_at, unlocked_by, created_at`
	lo := &models.AccountLockout{}
	err := r.db.QueryRow(ctx, query, userID, unlockAt, reason, lockoutCount).Scan(
		&lo.ID, &lo.UserID, &lo.LockedAt, &lo.UnlockAt,
		&lo.UnlockTokenHash, &lo.UnlockTokenExpiresAt, &lo.UnlockTokenUsedAt,
		&lo.Reason, &lo.LockoutCount, &lo.UnlockedAt, &lo.UnlockedBy, &lo.CreatedAt,
	)
	return lo, err
}

func (r *Repository) GetActiveAccountLockout(ctx context.Context, userID int64) (*models.AccountLockout, error) {
	query := `SELECT id, user_id, locked_at, unlock_at, unlock_token_hash, unlock_token_expires_at, unlock_token_used_at, reason, lockout_count, unlocked_at, unlocked_by, created_at
	          FROM account_lockouts
	          WHERE user_id = $1 AND unlocked_at IS NULL AND unlock_at > NOW()
	          ORDER BY created_at DESC LIMIT 1`
	lo := &models.AccountLockout{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&lo.ID, &lo.UserID, &lo.LockedAt, &lo.UnlockAt,
		&lo.UnlockTokenHash, &lo.UnlockTokenExpiresAt, &lo.UnlockTokenUsedAt,
		&lo.Reason, &lo.LockoutCount, &lo.UnlockedAt, &lo.UnlockedBy, &lo.CreatedAt,
	)
	return lo, err
}

func (r *Repository) CountUserLockouts(ctx context.Context, userID int64) (int, error) {
	query := `SELECT COUNT(*) FROM account_lockouts WHERE user_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

func (r *Repository) UnlockAccount(ctx context.Context, lockoutID int64, unlockedBy *int64) error {
	_, err := r.db.Exec(ctx, `UPDATE account_lockouts SET unlocked_at = NOW(), unlocked_by = $2 WHERE id = $1`, lockoutID, unlockedBy)
	return err
}

func (r *Repository) SetUnlockToken(ctx context.Context, lockoutID int64, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, `UPDATE account_lockouts SET unlock_token_hash = $2, unlock_token_expires_at = $3 WHERE id = $1`, lockoutID, tokenHash, expiresAt)
	return err
}

func (r *Repository) GetLockoutByUnlockToken(ctx context.Context, tokenHash string) (*models.AccountLockout, error) {
	query := `SELECT id, user_id, locked_at, unlock_at, unlock_token_hash, unlock_token_expires_at, unlock_token_used_at, reason, lockout_count, unlocked_at, unlocked_by, created_at
	          FROM account_lockouts WHERE unlock_token_hash = $1`
	lo := &models.AccountLockout{}
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&lo.ID, &lo.UserID, &lo.LockedAt, &lo.UnlockAt,
		&lo.UnlockTokenHash, &lo.UnlockTokenExpiresAt, &lo.UnlockTokenUsedAt,
		&lo.Reason, &lo.LockoutCount, &lo.UnlockedAt, &lo.UnlockedBy, &lo.CreatedAt,
	)
	return lo, err
}

func (r *Repository) MarkUnlockTokenUsed(ctx context.Context, lockoutID int64) error {
	_, err := r.db.Exec(ctx, `UPDATE account_lockouts SET unlock_token_used_at = NOW() WHERE id = $1`, lockoutID)
	return err
}

// Security notification operations

func (r *Repository) CreateSecurityNotification(ctx context.Context, userID int64, notifType, ipAddress string) error {
	query := `INSERT INTO security_notifications (user_id, notification_type, ip_address) VALUES ($1, $2, $3::inet)`
	_, err := r.db.Exec(ctx, query, userID, notifType, nullableString(ipAddress))
	return err
}

func (r *Repository) CountRecentSecurityNotifications(ctx context.Context, userID int64, notifType string, since time.Time) (int, error) {
	query := `SELECT COUNT(*) FROM security_notifications WHERE user_id = $1 AND notification_type = $2 AND sent_at >= $3`
	var count int
	err := r.db.QueryRow(ctx, query, userID, notifType, since).Scan(&count)
	return count, err
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// Audit log operations

func (r *Repository) CreateAuditLog(ctx context.Context, entry *models.AuditLog) error {
	query := `INSERT INTO audit_logs
		(user_id, username, action, resource_type, resource_id, resource_name, old_values, new_values, ip_address, user_agent, status, error_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.db.Exec(ctx, query,
		entry.UserID, entry.Username, entry.Action,
		nullableString(entry.ResourceType), entry.ResourceID, nullableString(entry.ResourceName),
		entry.OldValues, entry.NewValues,
		nullableString(entry.IPAddress), nullableString(entry.UserAgent),
		entry.Status, nullableString(entry.ErrorMessage),
	)
	return err
}

func (r *Repository) ListAuditLogs(ctx context.Context, filter *models.AuditLogFilter) ([]*models.AuditLog, error) {
	args := []interface{}{}
	where := []string{}
	i := 1

	if filter.UserID != nil {
		where = append(where, fmt.Sprintf("user_id = $%d", i))
		args = append(args, *filter.UserID)
		i++
	}
	if filter.Username != "" {
		where = append(where, fmt.Sprintf("username ILIKE $%d", i))
		args = append(args, "%"+filter.Username+"%")
		i++
	}
	if filter.Action != "" {
		where = append(where, fmt.Sprintf("action = $%d", i))
		args = append(args, filter.Action)
		i++
	}
	if filter.ResourceType != "" {
		where = append(where, fmt.Sprintf("resource_type = $%d", i))
		args = append(args, filter.ResourceType)
		i++
	}
	if filter.IPAddress != "" {
		where = append(where, fmt.Sprintf("ip_address = $%d", i))
		args = append(args, filter.IPAddress)
		i++
	}
	if filter.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", i))
		args = append(args, filter.Status)
		i++
	}
	if filter.Since != nil {
		where = append(where, fmt.Sprintf("created_at >= $%d", i))
		args = append(args, *filter.Since)
		i++
	}
	if filter.Until != nil {
		where = append(where, fmt.Sprintf("created_at <= $%d", i))
		args = append(args, *filter.Until)
		i++
	}

	query := `SELECT id, user_id, username, action, resource_type, resource_id, resource_name,
		old_values, new_values, ip_address, user_agent, status, error_message, created_at
		FROM audit_logs`
	if len(where) > 0 {
		query += " WHERE " + joinStrings(where, " AND ")
	}
	query += " ORDER BY created_at DESC"

	limit := filter.Limit
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, limit, filter.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]*models.AuditLog, 0)
	for rows.Next() {
		l := &models.AuditLog{}
		err := rows.Scan(
			&l.ID, &l.UserID, &l.Username, &l.Action,
			scanNullString(&l.ResourceType), &l.ResourceID, scanNullString(&l.ResourceName),
			&l.OldValues, &l.NewValues,
			scanNullString(&l.IPAddress), scanNullString(&l.UserAgent),
			&l.Status, scanNullString(&l.ErrorMessage), &l.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func (r *Repository) CountAuditLogs(ctx context.Context, filter *models.AuditLogFilter) (int64, error) {
	args := []interface{}{}
	where := []string{}
	i := 1

	if filter.UserID != nil {
		where = append(where, fmt.Sprintf("user_id = $%d", i))
		args = append(args, *filter.UserID)
		i++
	}
	if filter.Action != "" {
		where = append(where, fmt.Sprintf("action = $%d", i))
		args = append(args, filter.Action)
		i++
	}
	if filter.ResourceType != "" {
		where = append(where, fmt.Sprintf("resource_type = $%d", i))
		args = append(args, filter.ResourceType)
		i++
	}
	if filter.Since != nil {
		where = append(where, fmt.Sprintf("created_at >= $%d", i))
		args = append(args, *filter.Since)
		i++
	}
	if filter.Until != nil {
		where = append(where, fmt.Sprintf("created_at <= $%d", i))
		args = append(args, *filter.Until)
		i++
	}

	query := `SELECT COUNT(*) FROM audit_logs`
	if len(where) > 0 {
		query += " WHERE " + joinStrings(where, " AND ")
	}

	var count int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

func (r *Repository) DeleteAuditLogsBefore(ctx context.Context, before time.Time) (int64, error) {
	query := `DELETE FROM audit_logs WHERE created_at < $1`
	result, err := r.db.Exec(ctx, query, before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// scanNullString returns a pointer that scan can write into; empty DB nulls become ""
func scanNullString(dest *string) *nullStringScanner {
	return &nullStringScanner{dest: dest}
}

type nullStringScanner struct{ dest *string }

func (n *nullStringScanner) Scan(src interface{}) error {
	if src == nil {
		*n.dest = ""
		return nil
	}
	switch v := src.(type) {
	case string:
		*n.dest = v
	case []byte:
		*n.dest = string(v)
	default:
		*n.dest = fmt.Sprintf("%v", v)
	}
	return nil
}

func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
