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
	query := `INSERT INTO users (username, email, role) VALUES ($1, $2, 'user') RETURNING id, username, email, password_hash, role, state, last_login_at, suspended_at, suspended_by, suspension_reason, privacy_accepted_at, privacy_accepted_version, deletion_requested_at, anonymized_at, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, username, email)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.SuspendedAt, &user.SuspendedBy, &user.SuspensionReason, &user.PrivacyAcceptedAt, &user.PrivacyAcceptedVersion, &user.DeletionRequestedAt, &user.AnonymizedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, role, state, last_login_at, suspended_at, suspended_by, suspension_reason, privacy_accepted_at, privacy_accepted_version, deletion_requested_at, anonymized_at, created_at, updated_at FROM users WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.SuspendedAt, &user.SuspendedBy, &user.SuspensionReason, &user.PrivacyAcceptedAt, &user.PrivacyAcceptedVersion, &user.DeletionRequestedAt, &user.AnonymizedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, role, state, last_login_at, suspended_at, suspended_by, suspension_reason, privacy_accepted_at, privacy_accepted_version, deletion_requested_at, anonymized_at, created_at, updated_at FROM users WHERE username = $1`
	row := r.db.QueryRow(ctx, query, username)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.SuspendedAt, &user.SuspendedBy, &user.SuspensionReason, &user.PrivacyAcceptedAt, &user.PrivacyAcceptedVersion, &user.DeletionRequestedAt, &user.AnonymizedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) ListAllUsers(ctx context.Context) ([]*models.User, error) {
	query := `SELECT id, username, email, password_hash, role, state, last_login_at, suspended_at, suspended_by, suspension_reason, privacy_accepted_at, privacy_accepted_version, deletion_requested_at, anonymized_at, created_at, updated_at FROM users ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*models.User, 0)
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.SuspendedAt, &user.SuspendedBy, &user.SuspensionReason, &user.PrivacyAcceptedAt, &user.PrivacyAcceptedVersion, &user.DeletionRequestedAt, &user.AnonymizedAt, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *Repository) CreateUserWithPassword(ctx context.Context, username, email, passwordHash, role string) (*models.User, error) {
	query := `INSERT INTO users (username, email, password_hash, role) VALUES ($1, $2, $3, $4) RETURNING id, username, email, password_hash, role, state, last_login_at, suspended_at, suspended_by, suspension_reason, privacy_accepted_at, privacy_accepted_version, deletion_requested_at, anonymized_at, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, username, email, passwordHash, role)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.SuspendedAt, &user.SuspendedBy, &user.SuspensionReason, &user.PrivacyAcceptedAt, &user.PrivacyAcceptedVersion, &user.DeletionRequestedAt, &user.AnonymizedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) CreateUserWithState(ctx context.Context, username, email, passwordHash, role, state string) (*models.User, error) {
	query := `INSERT INTO users (username, email, password_hash, role, state) VALUES ($1, $2, $3, $4, $5) RETURNING id, username, email, password_hash, role, state, last_login_at, suspended_at, suspended_by, suspension_reason, privacy_accepted_at, privacy_accepted_version, deletion_requested_at, anonymized_at, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, username, email, passwordHash, role, state)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.SuspendedAt, &user.SuspendedBy, &user.SuspensionReason, &user.PrivacyAcceptedAt, &user.PrivacyAcceptedVersion, &user.DeletionRequestedAt, &user.AnonymizedAt, &user.CreatedAt, &user.UpdatedAt)
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

func (r *Repository) UpdateUserEmail(ctx context.Context, userID int64, email string) error {
	query := `UPDATE users SET email = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID, email)
	return err
}

func (r *Repository) UpdateUserRole(ctx context.Context, userID int64, role string) (*models.User, error) {
	query := `UPDATE users SET role = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1 RETURNING id, username, email, password_hash, role, state, last_login_at, suspended_at, suspended_by, suspension_reason, privacy_accepted_at, privacy_accepted_version, deletion_requested_at, anonymized_at, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, userID, role)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.SuspendedAt, &user.SuspendedBy, &user.SuspensionReason, &user.PrivacyAcceptedAt, &user.PrivacyAcceptedVersion, &user.DeletionRequestedAt, &user.AnonymizedAt, &user.CreatedAt, &user.UpdatedAt)
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
	query := `SELECT id, username, email, password_hash, role, state, last_login_at, suspended_at, suspended_by, suspension_reason, privacy_accepted_at, privacy_accepted_version, deletion_requested_at, anonymized_at, created_at, updated_at FROM users WHERE email = $1`
	row := r.db.QueryRow(ctx, query, email)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.State, &user.LastLoginAt, &user.SuspendedAt, &user.SuspendedBy, &user.SuspensionReason, &user.PrivacyAcceptedAt, &user.PrivacyAcceptedVersion, &user.DeletionRequestedAt, &user.AnonymizedAt, &user.CreatedAt, &user.UpdatedAt)
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

// InitAdminPassword sets the admin password only when it is currently NULL (i.e. first boot).
// Returns true if the password was set, false if it was already set.
func (r *Repository) InitAdminPassword(ctx context.Context, passwordHash string) (bool, error) {
	query := `UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE username = 'admin' AND password_hash IS NULL`
	result, err := r.db.Exec(ctx, query, passwordHash)
	if err != nil {
		return false, err
	}
	return result.RowsAffected() > 0, nil
}

// ForceSetAdminPassword unconditionally updates the admin user's password hash.
func (r *Repository) ForceSetAdminPassword(ctx context.Context, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE username = 'admin'`
	_, err := r.db.Exec(ctx, query, passwordHash)
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

func scanSubnet(row interface {
	Scan(dest ...any) error
}) (*models.Subnet, error) {
	subnet := &models.Subnet{}
	err := row.Scan(&subnet.ID, &subnet.SectionID, &subnet.NetworkAddress, &subnet.PrefixLength,
		&subnet.Description, &subnet.Gateway, &subnet.AutoReserveFirst, &subnet.AutoReserveLast,
		&subnet.CreatedAt, &subnet.UpdatedAt)
	return subnet, err
}

const subnetSelectCols = `id, section_id, host(network_address), prefix_length, description, gateway, auto_reserve_first, auto_reserve_last, created_at, updated_at`

func (r *Repository) CreateSubnet(ctx context.Context, sectionID int64, networkAddress string, prefixLength int, description string, gateway *string, autoFirst, autoLast bool) (*models.Subnet, error) {
	query := `INSERT INTO subnets (section_id, network_address, prefix_length, description, gateway, auto_reserve_first, auto_reserve_last)
	          VALUES ($1, $2, $3, $4, $5, $6, $7)
	          RETURNING ` + subnetSelectCols
	row := r.db.QueryRow(ctx, query, sectionID, networkAddress, prefixLength, description, gateway, autoFirst, autoLast)
	return scanSubnet(row)
}

func (r *Repository) GetSubnetByID(ctx context.Context, id int64) (*models.Subnet, error) {
	query := `SELECT ` + subnetSelectCols + ` FROM subnets WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	return scanSubnet(row)
}

func (r *Repository) ListSubnetsBySection(ctx context.Context, sectionID int64) ([]*models.Subnet, error) {
	query := `SELECT ` + subnetSelectCols + ` FROM subnets WHERE section_id = $1 ORDER BY network_address`
	rows, err := r.db.Query(ctx, query, sectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subnets := make([]*models.Subnet, 0)
	for rows.Next() {
		subnet, err := scanSubnet(rows)
		if err != nil {
			return nil, err
		}
		subnets = append(subnets, subnet)
	}
	return subnets, rows.Err()
}

func (r *Repository) UpdateSubnet(ctx context.Context, id int64, description string, gateway *string, autoFirst, autoLast bool) (*models.Subnet, error) {
	query := `UPDATE subnets SET description = $1, gateway = $2, auto_reserve_first = $3, auto_reserve_last = $4,
	          updated_at = CURRENT_TIMESTAMP WHERE id = $5
	          RETURNING ` + subnetSelectCols
	row := r.db.QueryRow(ctx, query, description, gateway, autoFirst, autoLast, id)
	return scanSubnet(row)
}

func (r *Repository) DeleteSubnet(ctx context.Context, id int64) error {
	query := `DELETE FROM subnets WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// IP Address operations

// ipSelectCols is the column list for ip_addresses JOINed with ip_tags
const ipSelectCols = `ip.id, ip.subnet_id, ip.address::text, ip.hostname, ip.status, ip.assigned_to,
	ip.tag_id, t.id, t.name, t.colour, t.description, t.is_system, t.created_at,
	ip.last_seen, ip.mac_address, ip.ptr_record,
	ip.created_at, ip.updated_at`

const ipFromJoin = `FROM ip_addresses ip LEFT JOIN ip_tags t ON ip.tag_id = t.id`

func scanIP(row interface {
	Scan(dest ...any) error
}) (*models.IPAddress, error) {
	ip := &models.IPAddress{}
	var tagID *int64
	var tagIDInner *int64
	var tagName *string
	var tagColour *string
	var tagDesc *string
	var tagIsSystem *bool
	var tagCreatedAt *time.Time

	err := row.Scan(
		&ip.ID, &ip.SubnetID, &ip.Address, &ip.Hostname, &ip.Status, &ip.AssignedTo,
		&tagID, &tagIDInner, &tagName, &tagColour, &tagDesc, &tagIsSystem, &tagCreatedAt,
		&ip.LastSeen, &ip.MACAddress, &ip.PTRRecord,
		&ip.CreatedAt, &ip.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	ip.TagID = tagID
	if tagIDInner != nil {
		ip.Tag = &models.IPTag{
			ID:          *tagIDInner,
			Name:        *tagName,
			Colour:      *tagColour,
			Description: tagDesc,
			IsSystem:    *tagIsSystem,
			CreatedAt:   *tagCreatedAt,
		}
	}
	return ip, nil
}

func (r *Repository) CreateIPAddress(ctx context.Context, subnetID int64, address, hostname string, status string, assignedTo *string, tagID *int64, macAddress, ptrRecord *string) (*models.IPAddress, error) {
	query := `WITH ins AS (
		INSERT INTO ip_addresses (subnet_id, address, hostname, status, assigned_to, tag_id, mac_address, ptr_record)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, subnet_id, address, hostname, status, assigned_to, tag_id, last_seen, mac_address, ptr_record, created_at, updated_at
	)
	SELECT ` + ipSelectCols + ` ` + ipFromJoin + ` WHERE ip.id = (SELECT id FROM ins)`
	row := r.db.QueryRow(ctx, query, subnetID, address, hostname, status, assignedTo, tagID, macAddress, ptrRecord)
	return scanIP(row)
}

func (r *Repository) GetIPAddressByID(ctx context.Context, id int64) (*models.IPAddress, error) {
	query := `SELECT ` + ipSelectCols + ` ` + ipFromJoin + ` WHERE ip.id = $1`
	row := r.db.QueryRow(ctx, query, id)
	return scanIP(row)
}

func (r *Repository) ListIPAddressesBySubnet(ctx context.Context, subnetID int64) ([]*models.IPAddress, error) {
	query := `SELECT ` + ipSelectCols + ` ` + ipFromJoin + ` WHERE ip.subnet_id = $1 ORDER BY ip.address`
	rows, err := r.db.Query(ctx, query, subnetID)
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

func (r *Repository) UpdateIPAddressStatus(ctx context.Context, id int64, status string, assignedTo *string) (*models.IPAddress, error) {
	query := `WITH upd AS (
		UPDATE ip_addresses SET status = $2, assigned_to = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 RETURNING id
	)
	SELECT ` + ipSelectCols + ` ` + ipFromJoin + ` WHERE ip.id = (SELECT id FROM upd)`
	row := r.db.QueryRow(ctx, query, id, status, assignedTo)
	return scanIP(row)
}

func (r *Repository) UpdateIPAddressFull(ctx context.Context, id int64, tagID *int64, macAddress, ptrRecord *string) (*models.IPAddress, error) {
	query := `WITH upd AS (
		UPDATE ip_addresses SET tag_id = $2, mac_address = $3, ptr_record = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 RETURNING id
	)
	SELECT ` + ipSelectCols + ` ` + ipFromJoin + ` WHERE ip.id = (SELECT id FROM upd)`
	row := r.db.QueryRow(ctx, query, id, tagID, macAddress, ptrRecord)
	return scanIP(row)
}

func (r *Repository) DeleteIPAddress(ctx context.Context, id int64) error {
	query := `DELETE FROM ip_addresses WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *Repository) ListAvailableIPsBySubnet(ctx context.Context, subnetID int64) ([]*models.IPAddress, error) {
	query := `SELECT ` + ipSelectCols + ` ` + ipFromJoin + ` WHERE ip.subnet_id = $1 AND ip.status = 'available' ORDER BY ip.address`
	rows, err := r.db.Query(ctx, query, subnetID)
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
	findQuery := `SELECT ip.id ` + ipFromJoin + ` WHERE ip.subnet_id = $1 AND ip.status = 'available' ORDER BY ip.address LIMIT 1`
	var ipID int64
	err = tx.QueryRow(ctx, findQuery, subnetID).Scan(&ipID)
	if err != nil {
		return nil, err
	}

	// Atomically update the IP status to 'assigned'
	updateQuery := `UPDATE ip_addresses SET status = 'assigned', assigned_to = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err = tx.Exec(ctx, updateQuery, assignedTo, ipID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetIPAddressByID(ctx, ipID)
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
	query := `WITH upd AS (
		UPDATE ip_addresses SET status = $2, assigned_to = $3, assigned_at = $4, expires_at = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 RETURNING id
	)
	SELECT ` + ipSelectCols + ` ` + ipFromJoin + ` WHERE ip.id = (SELECT id FROM upd)`
	row := r.db.QueryRow(ctx, query, id, status, assignedTo, assignedAt, expiresAt)
	return scanIP(row)
}

// UpdateLastSeen updates the last_seen timestamp for a discovered IP
func (r *Repository) UpdateLastSeen(ctx context.Context, id int64, lastSeen time.Time) error {
	_, err := r.db.Exec(ctx, `UPDATE ip_addresses SET last_seen = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, id, lastSeen)
	return err
}

// IP Tag operations

func scanTag(row interface {
	Scan(dest ...any) error
}) (*models.IPTag, error) {
	tag := &models.IPTag{}
	return tag, row.Scan(&tag.ID, &tag.Name, &tag.Colour, &tag.Description, &tag.IsSystem, &tag.CreatedAt)
}

func (r *Repository) CreateIPTag(ctx context.Context, name, colour string, description *string) (*models.IPTag, error) {
	query := `INSERT INTO ip_tags (name, colour, description) VALUES ($1, $2, $3)
	          RETURNING id, name, colour, description, is_system, created_at`
	row := r.db.QueryRow(ctx, query, name, colour, description)
	return scanTag(row)
}

func (r *Repository) GetIPTagByID(ctx context.Context, id int64) (*models.IPTag, error) {
	query := `SELECT id, name, colour, description, is_system, created_at FROM ip_tags WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	return scanTag(row)
}

func (r *Repository) ListIPTags(ctx context.Context) ([]*models.IPTag, error) {
	query := `SELECT id, name, colour, description, is_system, created_at FROM ip_tags ORDER BY name`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tags := make([]*models.IPTag, 0)
	for rows.Next() {
		tag, err := scanTag(rows)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (r *Repository) UpdateIPTag(ctx context.Context, id int64, name, colour string, description *string) (*models.IPTag, error) {
	query := `UPDATE ip_tags SET name = $2, colour = $3, description = $4 WHERE id = $1
	          RETURNING id, name, colour, description, is_system, created_at`
	row := r.db.QueryRow(ctx, query, id, name, colour, description)
	return scanTag(row)
}

func (r *Repository) DeleteIPTag(ctx context.Context, id int64) error {
	// Prevent deleting system tags
	var isSystem bool
	err := r.db.QueryRow(ctx, `SELECT is_system FROM ip_tags WHERE id = $1`, id).Scan(&isSystem)
	if err != nil {
		return fmt.Errorf("tag not found")
	}
	if isSystem {
		return fmt.Errorf("cannot delete system tag")
	}
	// Prevent deleting tags in use
	var count int64
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM ip_addresses WHERE tag_id = $1`, id).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("tag is in use by %d IP address(es)", count)
	}
	_, err = r.db.Exec(ctx, `DELETE FROM ip_tags WHERE id = $1`, id)
	return err
}

// API Token operations

func (r *Repository) CreateAPIToken(ctx context.Context, userID int64, tokenHash, name string) (*models.APIToken, error) {
	query := `INSERT INTO api_tokens (user_id, token_hash, name) VALUES ($1, $2, $3)
	          RETURNING id, user_id, token_hash, name, scope, usage_count, last_used_at, last_used_ip, expires_at, rotation_grace_expires_at, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, userID, tokenHash, name)

	token := &models.APIToken{}
	err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.Name, &token.Scope,
		&token.UsageCount, &token.LastUsedAt, &token.LastUsedIP,
		&token.ExpiresAt, &token.RotationGraceExpiresAt, &token.CreatedAt, &token.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *Repository) CreateAPITokenFull(ctx context.Context, userID int64, tokenHash, name, scope string, expiresAt *time.Time) (*models.APIToken, error) {
	query := `INSERT INTO api_tokens (user_id, token_hash, name, scope, expires_at)
	          VALUES ($1, $2, $3, $4, $5)
	          RETURNING id, user_id, token_hash, name, scope, usage_count, last_used_at, last_used_ip, expires_at, rotation_grace_expires_at, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, userID, tokenHash, name, scope, expiresAt)
	token := &models.APIToken{}
	err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.Name, &token.Scope,
		&token.UsageCount, &token.LastUsedAt, &token.LastUsedIP,
		&token.ExpiresAt, &token.RotationGraceExpiresAt, &token.CreatedAt, &token.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *Repository) GetAPITokenByHash(ctx context.Context, tokenHash string) (*models.APIToken, error) {
	query := `SELECT id, user_id, token_hash, name, scope, usage_count, last_used_at, last_used_ip, expires_at, rotation_grace_expires_at, created_at, updated_at FROM api_tokens WHERE token_hash = $1`
	row := r.db.QueryRow(ctx, query, tokenHash)

	token := &models.APIToken{}
	err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.Name, &token.Scope,
		&token.UsageCount, &token.LastUsedAt, &token.LastUsedIP,
		&token.ExpiresAt, &token.RotationGraceExpiresAt, &token.CreatedAt, &token.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *Repository) ListAPITokensByUser(ctx context.Context, userID int64) ([]*models.APIToken, error) {
	query := `SELECT id, user_id, token_hash, name, scope, usage_count, last_used_at, last_used_ip, expires_at, rotation_grace_expires_at, created_at, updated_at FROM api_tokens WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens := make([]*models.APIToken, 0)
	for rows.Next() {
		token := &models.APIToken{}
		err := rows.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.Name, &token.Scope,
			&token.UsageCount, &token.LastUsedAt, &token.LastUsedIP,
			&token.ExpiresAt, &token.RotationGraceExpiresAt, &token.CreatedAt, &token.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}

func (r *Repository) UpdateAPITokenLastUsed(ctx context.Context, tokenID int64, ip string) error {
	query := `UPDATE api_tokens SET last_used_at = CURRENT_TIMESTAMP, last_used_ip = $2, usage_count = usage_count + 1 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, tokenID, nullableString(ip))
	return err
}

func (r *Repository) DeleteAPIToken(ctx context.Context, tokenID int64) error {
	query := `DELETE FROM api_tokens WHERE id = $1`
	_, err := r.db.Exec(ctx, query, tokenID)
	return err
}

func (r *Repository) MarkAPITokenRotated(ctx context.Context, tokenID int64, graceExpiresAt time.Time) error {
	query := `UPDATE api_tokens SET rotation_grace_expires_at = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, tokenID, graceExpiresAt)
	return err
}

func (r *Repository) ExtendAPIToken(ctx context.Context, tokenID, userID int64, newExpiresAt time.Time) (*models.APIToken, error) {
	query := `UPDATE api_tokens SET expires_at = $3, updated_at = CURRENT_TIMESTAMP
	          WHERE id = $1 AND user_id = $2
	          RETURNING id, user_id, token_hash, name, scope, usage_count, last_used_at, last_used_ip, expires_at, rotation_grace_expires_at, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, tokenID, userID, newExpiresAt)
	token := &models.APIToken{}
	err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.Name, &token.Scope,
		&token.UsageCount, &token.LastUsedAt, &token.LastUsedIP,
		&token.ExpiresAt, &token.RotationGraceExpiresAt, &token.CreatedAt, &token.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *Repository) GetAPITokenByID(ctx context.Context, tokenID int64) (*models.APIToken, error) {
	query := `SELECT id, user_id, token_hash, name, scope, usage_count, last_used_at, last_used_ip, expires_at, rotation_grace_expires_at, created_at, updated_at FROM api_tokens WHERE id = $1`
	row := r.db.QueryRow(ctx, query, tokenID)
	token := &models.APIToken{}
	err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.Name, &token.Scope,
		&token.UsageCount, &token.LastUsedAt, &token.LastUsedIP,
		&token.ExpiresAt, &token.RotationGraceExpiresAt, &token.CreatedAt, &token.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *Repository) DeleteExpiredAPITokens(ctx context.Context) error {
	// Delete tokens expired more than 30 days ago with no grace period active
	query := `DELETE FROM api_tokens WHERE expires_at IS NOT NULL AND expires_at < NOW() - INTERVAL '30 days' AND (rotation_grace_expires_at IS NULL OR rotation_grace_expires_at < NOW() - INTERVAL '30 days')`
	_, err := r.db.Exec(ctx, query)
	return err
}

// Session operations

func (r *Repository) CreateSession(ctx context.Context, userID int64, tokenHash, deviceName, ipAddress, userAgent string, absoluteExpiresAt time.Time) (*models.Session, error) {
	query := `INSERT INTO sessions (user_id, token_hash, device_name, ip_address, user_agent, absolute_expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, token_hash, device_name, ip_address, user_agent, last_used_at, absolute_expires_at, is_impersonation, impersonated_by, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, userID, tokenHash, deviceName, ipAddress, userAgent, absoluteExpiresAt)

	s := &models.Session{}
	err := row.Scan(&s.ID, &s.UserID, &s.TokenHash, &s.DeviceName, &s.IPAddress, &s.UserAgent, &s.LastUsedAt, &s.AbsoluteExpiresAt, &s.IsImpersonation, &s.ImpersonatedBy, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) GetSessionByHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	query := `SELECT id, user_id, token_hash, device_name, ip_address, user_agent, last_used_at, absolute_expires_at, is_impersonation, impersonated_by, created_at, updated_at FROM sessions WHERE token_hash = $1`
	row := r.db.QueryRow(ctx, query, tokenHash)

	s := &models.Session{}
	err := row.Scan(&s.ID, &s.UserID, &s.TokenHash, &s.DeviceName, &s.IPAddress, &s.UserAgent, &s.LastUsedAt, &s.AbsoluteExpiresAt, &s.IsImpersonation, &s.ImpersonatedBy, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) ListSessionsByUser(ctx context.Context, userID int64) ([]*models.Session, error) {
	query := `SELECT id, user_id, token_hash, device_name, ip_address, user_agent, last_used_at, absolute_expires_at, is_impersonation, impersonated_by, created_at, updated_at FROM sessions WHERE user_id = $1 ORDER BY last_used_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]*models.Session, 0)
	for rows.Next() {
		s := &models.Session{}
		err := rows.Scan(&s.ID, &s.UserID, &s.TokenHash, &s.DeviceName, &s.IPAddress, &s.UserAgent, &s.LastUsedAt, &s.AbsoluteExpiresAt, &s.IsImpersonation, &s.ImpersonatedBy, &s.CreatedAt, &s.UpdatedAt)
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
	sql := `SELECT ` + subnetSelectCols + ` FROM subnets
	        WHERE section_id = $1 AND (host(network_address) ILIKE $2 OR description ILIKE $2)
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
		subnet, err := scanSubnet(rows)
		if err != nil {
			return nil, err
		}
		subnets = append(subnets, subnet)
	}
	return subnets, rows.Err()
}

// IPSearchFilter holds optional additional filters for IP search
type IPSearchFilter struct {
	TagID          *int64
	MACAddress     string
	PTRRecord      string
	IsAssigned     *bool
	LastSeenAfter  *time.Time
	LastSeenBefore *time.Time
}

func (r *Repository) SearchIPAddresses(ctx context.Context, subnetID int64, query string, status string, limit, offset int64, filter ...IPSearchFilter) ([]*models.IPAddress, error) {
	sql := `SELECT ` + ipSelectCols + ` ` + ipFromJoin + `
	        WHERE ip.subnet_id = $1 AND (ip.address::text ILIKE $2 OR ip.hostname ILIKE $2 OR ip.assigned_to ILIKE $2)`
	args := []interface{}{subnetID, "%" + query + "%"}
	n := 3

	if status != "" {
		sql += fmt.Sprintf(" AND ip.status = $%d", n)
		args = append(args, status)
		n++
	}

	if len(filter) > 0 {
		f := filter[0]
		if f.TagID != nil {
			sql += fmt.Sprintf(" AND ip.tag_id = $%d", n)
			args = append(args, *f.TagID)
			n++
		}
		if f.MACAddress != "" {
			sql += fmt.Sprintf(" AND ip.mac_address ILIKE $%d", n)
			args = append(args, "%"+f.MACAddress+"%")
			n++
		}
		if f.PTRRecord != "" {
			sql += fmt.Sprintf(" AND ip.ptr_record ILIKE $%d", n)
			args = append(args, "%"+f.PTRRecord+"%")
			n++
		}
		if f.IsAssigned != nil {
			if *f.IsAssigned {
				sql += " AND ip.status = 'assigned'"
			} else {
				sql += " AND ip.status != 'assigned'"
			}
		}
		if f.LastSeenAfter != nil {
			sql += fmt.Sprintf(" AND ip.last_seen >= $%d", n)
			args = append(args, *f.LastSeenAfter)
			n++
		}
		if f.LastSeenBefore != nil {
			sql += fmt.Sprintf(" AND ip.last_seen <= $%d", n)
			args = append(args, *f.LastSeenBefore)
			n++
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

// ---- RBAC ----

func (r *Repository) CreateRole(ctx context.Context, name, description string, isSystem bool) (*models.Role, error) {
	query := `INSERT INTO roles (name, description, is_system) VALUES ($1, $2, $3)
              RETURNING id, name, description, is_system, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, name, description, isSystem)
	role := &models.Role{}
	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt)
	return role, err
}

func (r *Repository) GetRoleByID(ctx context.Context, id int64) (*models.Role, error) {
	query := `SELECT id, name, description, is_system, created_at, updated_at FROM roles WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	role := &models.Role{}
	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, err
	}
	role.Permissions, err = r.GetRolePermissions(ctx, id)
	return role, err
}

func (r *Repository) ListRoles(ctx context.Context) ([]*models.Role, error) {
	query := `SELECT id, name, description, is_system, created_at, updated_at FROM roles ORDER BY is_system DESC, name`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []*models.Role
	for rows.Next() {
		role := &models.Role{}
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		role.Permissions, _ = r.GetRolePermissions(ctx, role.ID)
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *Repository) UpdateRole(ctx context.Context, id int64, name, description string) (*models.Role, error) {
	query := `UPDATE roles SET name=$1, description=$2, updated_at=CURRENT_TIMESTAMP WHERE id=$3 AND is_system=FALSE
              RETURNING id, name, description, is_system, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, name, description, id)
	role := &models.Role{}
	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, err
	}
	role.Permissions, _ = r.GetRolePermissions(ctx, id)
	return role, nil
}

func (r *Repository) DeleteRole(ctx context.Context, id int64) error {
	// Only delete non-system roles
	res, err := r.db.Exec(ctx, `DELETE FROM roles WHERE id=$1 AND is_system=FALSE`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("role not found or is a system role")
	}
	return nil
}

func (r *Repository) GetRolePermissions(ctx context.Context, roleID int64) ([]*models.RolePermission, error) {
	query := `SELECT id, role_id, permission, resource_type, resource_id, created_at FROM role_permissions WHERE role_id=$1 ORDER BY permission`
	rows, err := r.db.Query(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var perms []*models.RolePermission
	for rows.Next() {
		p := &models.RolePermission{}
		if err := rows.Scan(&p.ID, &p.RoleID, &p.Permission, &p.ResourceType, &p.ResourceID, &p.CreatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, rows.Err()
}

func (r *Repository) AddPermissionToRole(ctx context.Context, roleID int64, permission string, resourceType *string, resourceID *int64) (*models.RolePermission, error) {
	query := `INSERT INTO role_permissions (role_id, permission, resource_type, resource_id) VALUES ($1, $2, $3, $4)
              RETURNING id, role_id, permission, resource_type, resource_id, created_at`
	row := r.db.QueryRow(ctx, query, roleID, permission, resourceType, resourceID)
	p := &models.RolePermission{}
	err := row.Scan(&p.ID, &p.RoleID, &p.Permission, &p.ResourceType, &p.ResourceID, &p.CreatedAt)
	return p, err
}

func (r *Repository) RemovePermissionFromRole(ctx context.Context, permissionID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM role_permissions WHERE id=$1`, permissionID)
	return err
}

func (r *Repository) AssignRoleToUser(ctx context.Context, userID, roleID int64) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, roleID)
	return err
}

func (r *Repository) RemoveRoleFromUser(ctx context.Context, userID, roleID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM user_roles WHERE user_id=$1 AND role_id=$2`, userID, roleID)
	return err
}

func (r *Repository) GetUserRoles(ctx context.Context, userID int64) ([]*models.Role, error) {
	query := `SELECT r.id, r.name, r.description, r.is_system, r.created_at, r.updated_at
              FROM roles r JOIN user_roles ur ON r.id=ur.role_id WHERE ur.user_id=$1 ORDER BY r.name`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []*models.Role
	for rows.Next() {
		role := &models.Role{}
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *Repository) GetUserPermissions(ctx context.Context, userID int64) ([]*models.RolePermission, error) {
	query := `SELECT DISTINCT rp.id, rp.role_id, rp.permission, rp.resource_type, rp.resource_id, rp.created_at
              FROM role_permissions rp
              JOIN user_roles ur ON rp.role_id = ur.role_id
              WHERE ur.user_id = $1
              ORDER BY rp.permission`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var perms []*models.RolePermission
	for rows.Next() {
		p := &models.RolePermission{}
		if err := rows.Scan(&p.ID, &p.RoleID, &p.Permission, &p.ResourceType, &p.ResourceID, &p.CreatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, rows.Err()
}

func (r *Repository) CountUserRoles(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM user_roles WHERE user_id=$1`, userID).Scan(&count)
	return count, err
}

// --- Notification preferences ---

func (r *Repository) GetNotificationPreferences(ctx context.Context, userID int64) (*models.NotificationPreferences, error) {
	query := `SELECT id, user_id, login_success, login_failed, account_locked, password_changed,
	          mfa_changes, api_token_changes, role_changes, session_revoked, created_at, updated_at
	          FROM notification_preferences WHERE user_id = $1`
	p := &models.NotificationPreferences{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&p.ID, &p.UserID, &p.LoginSuccess, &p.LoginFailed, &p.AccountLocked, &p.PasswordChanged,
		&p.MFAChanges, &p.APITokenChanges, &p.RoleChanges, &p.SessionRevoked, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return &models.NotificationPreferences{
			UserID:          userID,
			LoginSuccess:    true,
			LoginFailed:     true,
			AccountLocked:   true,
			PasswordChanged: true,
			MFAChanges:      true,
			APITokenChanges: true,
			RoleChanges:     true,
			SessionRevoked:  true,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *Repository) UpsertNotificationPreferences(ctx context.Context, prefs *models.NotificationPreferences) (*models.NotificationPreferences, error) {
	query := `INSERT INTO notification_preferences
	          (user_id, login_success, login_failed, account_locked, password_changed,
	           mfa_changes, api_token_changes, role_changes, session_revoked)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	          ON CONFLICT (user_id) DO UPDATE SET
	              login_success    = EXCLUDED.login_success,
	              login_failed     = EXCLUDED.login_failed,
	              account_locked   = EXCLUDED.account_locked,
	              password_changed = EXCLUDED.password_changed,
	              mfa_changes      = EXCLUDED.mfa_changes,
	              api_token_changes = EXCLUDED.api_token_changes,
	              role_changes     = EXCLUDED.role_changes,
	              session_revoked  = EXCLUDED.session_revoked,
	              updated_at       = CURRENT_TIMESTAMP
	          RETURNING id, user_id, login_success, login_failed, account_locked, password_changed,
	                    mfa_changes, api_token_changes, role_changes, session_revoked, created_at, updated_at`
	p := &models.NotificationPreferences{}
	err := r.db.QueryRow(ctx, query,
		prefs.UserID, prefs.LoginSuccess, prefs.LoginFailed, prefs.AccountLocked, prefs.PasswordChanged,
		prefs.MFAChanges, prefs.APITokenChanges, prefs.RoleChanges, prefs.SessionRevoked,
	).Scan(
		&p.ID, &p.UserID, &p.LoginSuccess, &p.LoginFailed, &p.AccountLocked, &p.PasswordChanged,
		&p.MFAChanges, &p.APITokenChanges, &p.RoleChanges, &p.SessionRevoked, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// --- Notification queue ---

func (r *Repository) CreateNotificationQueueItem(ctx context.Context, userID int64, email, template, dataJSON string) (*models.NotificationQueue, error) {
	query := `INSERT INTO notification_queue (user_id, email, template, data)
	          VALUES ($1, $2, $3, $4::jsonb)
	          RETURNING id, user_id, email, template, data::text, status, retry_count,
	                    next_retry_at, sent_at, error_msg, created_at, updated_at`
	q := &models.NotificationQueue{}
	err := r.db.QueryRow(ctx, query, userID, email, template, dataJSON).Scan(
		&q.ID, &q.UserID, &q.Email, &q.Template, &q.Data, &q.Status, &q.RetryCount,
		&q.NextRetryAt, &q.SentAt, &q.ErrorMsg, &q.CreatedAt, &q.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (r *Repository) GetPendingNotifications(ctx context.Context, limit int) ([]*models.NotificationQueue, error) {
	query := `SELECT id, user_id, email, template, data::text, status, retry_count,
	          next_retry_at, sent_at, error_msg, created_at, updated_at
	          FROM notification_queue
	          WHERE status IN ('pending', 'retrying')
	            AND (next_retry_at IS NULL OR next_retry_at <= NOW())
	          ORDER BY created_at ASC
	          LIMIT $1`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*models.NotificationQueue, 0)
	for rows.Next() {
		q := &models.NotificationQueue{}
		if err := rows.Scan(
			&q.ID, &q.UserID, &q.Email, &q.Template, &q.Data, &q.Status, &q.RetryCount,
			&q.NextRetryAt, &q.SentAt, &q.ErrorMsg, &q.CreatedAt, &q.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, q)
	}
	return items, rows.Err()
}

func (r *Repository) MarkNotificationSent(ctx context.Context, id int64) error {
	query := `UPDATE notification_queue SET status = 'sent', sent_at = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *Repository) MarkNotificationFailed(ctx context.Context, id int64, errMsg string, retryCount int, nextRetryAt *time.Time) error {
	status := "failed"
	if nextRetryAt != nil {
		status = "retrying"
	}
	query := `UPDATE notification_queue
	          SET status = $2, error_msg = $3, retry_count = $4, next_retry_at = $5, updated_at = NOW()
	          WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, status, errMsg, retryCount, nextRetryAt)
	return err
}

func (r *Repository) CountRecentNotificationsSent(ctx context.Context, userID int64, since time.Time) (int64, error) {
	query := `SELECT COUNT(*) FROM notification_queue WHERE user_id = $1 AND sent_at >= $2`
	var count int64
	err := r.db.QueryRow(ctx, query, userID, since).Scan(&count)
	return count, err
}

func (r *Repository) GetNotificationStats(ctx context.Context) (map[string]int64, error) {
	query := `SELECT status, COUNT(*) FROM notification_queue GROUP BY status`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := map[string]int64{
		"pending":  0,
		"sent":     0,
		"failed":   0,
		"retrying": 0,
	}
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats[status] = count
	}
	return stats, rows.Err()
}

func (r *Repository) CleanupOldNotifications(ctx context.Context) error {
	query := `DELETE FROM notification_queue
	          WHERE (status = 'sent'   AND sent_at    < NOW() - INTERVAL '30 days')
	             OR (status = 'failed' AND updated_at < NOW() - INTERVAL '7 days')`
	_, err := r.db.Exec(ctx, query)
	return err
}

// SuspendUser sets a user's state to suspended with reason and admin tracking
func (r *Repository) SuspendUser(ctx context.Context, userID, adminID int64, reason string) error {
	query := `UPDATE users SET state = 'suspended', suspended_at = CURRENT_TIMESTAMP, suspended_by = $2, suspension_reason = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID, adminID, reason)
	return err
}

// UnsuspendUser restores a user to active state
func (r *Repository) UnsuspendUser(ctx context.Context, userID int64) error {
	query := `UPDATE users SET state = 'active', suspended_at = NULL, suspended_by = NULL, suspension_reason = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

// BulkUpdateUserState updates the state of multiple users
func (r *Repository) BulkUpdateUserState(ctx context.Context, userIDs []int64, state string) (int64, error) {
	query := `UPDATE users SET state = $1, updated_at = CURRENT_TIMESTAMP WHERE id = ANY($2)`
	result, err := r.db.Exec(ctx, query, state, userIDs)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// BulkDeleteUsers deletes multiple users
func (r *Repository) BulkDeleteUsers(ctx context.Context, userIDs []int64) (int64, error) {
	query := `DELETE FROM users WHERE id = ANY($1)`
	result, err := r.db.Exec(ctx, query, userIDs)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// UpdatePrivacyConsent records user acceptance of the privacy policy
func (r *Repository) UpdatePrivacyConsent(ctx context.Context, userID int64, version string) error {
	query := `UPDATE users SET privacy_accepted_at = CURRENT_TIMESTAMP, privacy_accepted_version = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID, version)
	return err
}

// RequestDeletion marks a user as having requested account deletion
func (r *Repository) RequestDeletion(ctx context.Context, userID int64) error {
	query := `UPDATE users SET deletion_requested_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

// AnonymizeUser replaces PII with anonymized values (GDPR right to erasure)
func (r *Repository) AnonymizeUser(ctx context.Context, userID int64) error {
	query := `UPDATE users SET
		username = 'deleted_' || id::text,
		email = 'deleted_' || id::text || '@deleted.invalid',
		password_hash = '',
		state = 'disabled',
		anonymized_at = CURRENT_TIMESTAMP,
		updated_at = CURRENT_TIMESTAMP
	WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

// CreateImpersonationSession creates a session flagged as impersonation
func (r *Repository) CreateImpersonationSession(ctx context.Context, targetUserID, adminID int64, tokenHash, deviceName, ipAddress, userAgent string, absoluteExpiresAt time.Time) (*models.Session, error) {
	query := `INSERT INTO sessions (user_id, token_hash, device_name, ip_address, user_agent, absolute_expires_at, is_impersonation, impersonated_by)
		VALUES ($1, $2, $3, $4, $5, $6, TRUE, $7)
		RETURNING id, user_id, token_hash, device_name, ip_address, user_agent, last_used_at, absolute_expires_at, is_impersonation, impersonated_by, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, targetUserID, tokenHash, deviceName, ipAddress, userAgent, absoluteExpiresAt, adminID)

	s := &models.Session{}
	err := row.Scan(&s.ID, &s.UserID, &s.TokenHash, &s.DeviceName, &s.IPAddress, &s.UserAgent, &s.LastUsedAt, &s.AbsoluteExpiresAt, &s.IsImpersonation, &s.ImpersonatedBy, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// GetUserAllData returns all data associated with a user for GDPR export
func (r *Repository) GetUserAllData(ctx context.Context, userID int64) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	// Get user record
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	data["user"] = user

	// Get sessions
	sessions, err := r.ListSessionsByUser(ctx, userID)
	if err == nil {
		data["sessions"] = sessions
	}

	// Get API tokens
	tokens, err := r.ListAPITokensByUser(ctx, userID)
	if err == nil {
		data["api_tokens"] = tokens
	}

	// Get audit logs
	logs, err := r.ListAuditLogs(ctx, &models.AuditLogFilter{UserID: &userID, Limit: 1000})
	if err == nil {
		data["audit_logs"] = logs
	}

	return data, nil
}

// CreateScanJob creates a new discovery scan job
func (r *Repository) CreateScanJob(ctx context.Context, name string, subnetIDs []int64, scheduleCron *string, createdBy int64) (*models.ScanJob, error) {
	query := `INSERT INTO scan_jobs (name, subnet_ids, schedule_cron, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, subnet_ids, schedule_cron, is_active, last_run_at, next_run_at, created_by, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, name, subnetIDs, scheduleCron, createdBy)

	j := &models.ScanJob{}
	err := row.Scan(&j.ID, &j.Name, &j.SubnetIDs, &j.ScheduleCron, &j.IsActive, &j.LastRunAt, &j.NextRunAt, &j.CreatedBy, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return j, nil
}

// GetScanJobByID retrieves a scan job by ID
func (r *Repository) GetScanJobByID(ctx context.Context, id int64) (*models.ScanJob, error) {
	query := `SELECT id, name, subnet_ids, schedule_cron, is_active, last_run_at, next_run_at, created_by, created_at, updated_at FROM scan_jobs WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	j := &models.ScanJob{}
	err := row.Scan(&j.ID, &j.Name, &j.SubnetIDs, &j.ScheduleCron, &j.IsActive, &j.LastRunAt, &j.NextRunAt, &j.CreatedBy, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return j, nil
}

// ListScanJobs returns all scan jobs
func (r *Repository) ListScanJobs(ctx context.Context) ([]*models.ScanJob, error) {
	query := `SELECT id, name, subnet_ids, schedule_cron, is_active, last_run_at, next_run_at, created_by, created_at, updated_at FROM scan_jobs ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]*models.ScanJob, 0)
	for rows.Next() {
		j := &models.ScanJob{}
		if err := rows.Scan(&j.ID, &j.Name, &j.SubnetIDs, &j.ScheduleCron, &j.IsActive, &j.LastRunAt, &j.NextRunAt, &j.CreatedBy, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

// ListActiveScanJobs returns all active scan jobs with a schedule
func (r *Repository) ListActiveScanJobs(ctx context.Context) ([]*models.ScanJob, error) {
	query := `SELECT id, name, subnet_ids, schedule_cron, is_active, last_run_at, next_run_at, created_by, created_at, updated_at FROM scan_jobs WHERE is_active = TRUE AND schedule_cron IS NOT NULL`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]*models.ScanJob, 0)
	for rows.Next() {
		j := &models.ScanJob{}
		if err := rows.Scan(&j.ID, &j.Name, &j.SubnetIDs, &j.ScheduleCron, &j.IsActive, &j.LastRunAt, &j.NextRunAt, &j.CreatedBy, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

// UpdateScanJob updates a scan job's configuration
func (r *Repository) UpdateScanJob(ctx context.Context, id int64, name string, subnetIDs []int64, scheduleCron *string, isActive bool) (*models.ScanJob, error) {
	query := `UPDATE scan_jobs SET name = $2, subnet_ids = $3, schedule_cron = $4, is_active = $5, updated_at = CURRENT_TIMESTAMP WHERE id = $1
		RETURNING id, name, subnet_ids, schedule_cron, is_active, last_run_at, next_run_at, created_by, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, id, name, subnetIDs, scheduleCron, isActive)

	j := &models.ScanJob{}
	err := row.Scan(&j.ID, &j.Name, &j.SubnetIDs, &j.ScheduleCron, &j.IsActive, &j.LastRunAt, &j.NextRunAt, &j.CreatedBy, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return j, nil
}

// UpdateScanJobRunTime updates last_run_at and next_run_at after a scan
func (r *Repository) UpdateScanJobRunTime(ctx context.Context, id int64, nextRunAt *time.Time) error {
	query := `UPDATE scan_jobs SET last_run_at = CURRENT_TIMESTAMP, next_run_at = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, nextRunAt)
	return err
}

// DeleteScanJob deletes a scan job
func (r *Repository) DeleteScanJob(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM scan_jobs WHERE id = $1`, id)
	return err
}

// CreateScanResult records the result of scanning a single IP
func (r *Repository) CreateScanResult(ctx context.Context, jobID, subnetID int64, ipAddressID *int64, ipAddress string, isAlive bool, responseTimeMs *int64) (*models.ScanResult, error) {
	query := `INSERT INTO scan_results (job_id, subnet_id, ip_address_id, ip_address, is_alive, response_time_ms)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, job_id, subnet_id, ip_address_id, ip_address, is_alive, response_time_ms, scanned_at`
	row := r.db.QueryRow(ctx, query, jobID, subnetID, ipAddressID, ipAddress, isAlive, responseTimeMs)

	sr := &models.ScanResult{}
	err := row.Scan(&sr.ID, &sr.JobID, &sr.SubnetID, &sr.IPAddressID, &sr.IPAddress, &sr.IsAlive, &sr.ResponseTimeMs, &sr.ScannedAt)
	if err != nil {
		return nil, err
	}
	return sr, nil
}

// ListScanResultsByJob returns recent scan results for a job
func (r *Repository) ListScanResultsByJob(ctx context.Context, jobID int64, limit int) ([]*models.ScanResult, error) {
	query := `SELECT id, job_id, subnet_id, ip_address_id, ip_address, is_alive, response_time_ms, scanned_at FROM scan_results WHERE job_id = $1 ORDER BY scanned_at DESC LIMIT $2`
	rows, err := r.db.Query(ctx, query, jobID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]*models.ScanResult, 0)
	for rows.Next() {
		sr := &models.ScanResult{}
		if err := rows.Scan(&sr.ID, &sr.JobID, &sr.SubnetID, &sr.IPAddressID, &sr.IPAddress, &sr.IsAlive, &sr.ResponseTimeMs, &sr.ScannedAt); err != nil {
			return nil, err
		}
		results = append(results, sr)
	}
	return results, rows.Err()
}

// ListScanResultsBySubnet returns recent scan results for a subnet
func (r *Repository) ListScanResultsBySubnet(ctx context.Context, subnetID int64, limit int) ([]*models.ScanResult, error) {
	query := `SELECT id, job_id, subnet_id, ip_address_id, ip_address, is_alive, response_time_ms, scanned_at FROM scan_results WHERE subnet_id = $1 ORDER BY scanned_at DESC LIMIT $2`
	rows, err := r.db.Query(ctx, query, subnetID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]*models.ScanResult, 0)
	for rows.Next() {
		sr := &models.ScanResult{}
		if err := rows.Scan(&sr.ID, &sr.JobID, &sr.SubnetID, &sr.IPAddressID, &sr.IPAddress, &sr.IsAlive, &sr.ResponseTimeMs, &sr.ScannedAt); err != nil {
			return nil, err
		}
		results = append(results, sr)
	}
	return results, rows.Err()
}

// Dashboard operations

// GetDashboardSummary returns aggregate IPAM counts and top utilised subnets.
func (r *Repository) GetDashboardSummary(ctx context.Context) (*models.DashboardSummary, error) {
	summary := &models.DashboardSummary{}

	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM sections`).Scan(&summary.TotalSections); err != nil {
		return nil, fmt.Errorf("count sections: %w", err)
	}
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM subnets`).Scan(&summary.TotalSubnets); err != nil {
		return nil, fmt.Errorf("count subnets: %w", err)
	}
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM ip_addresses`).Scan(&summary.TotalIPs); err != nil {
		return nil, fmt.Errorf("count ips: %w", err)
	}
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM ip_addresses WHERE status = 'assigned'`).Scan(&summary.UsedIPs); err != nil {
		return nil, fmt.Errorf("count used ips: %w", err)
	}
	if summary.TotalIPs > 0 {
		summary.UtilisationPct = float64(summary.UsedIPs) / float64(summary.TotalIPs) * 100
	}

	topQuery := `
		SELECT
			s.id,
			host(s.network_address) || '/' || s.prefix_length AS cidr,
			s.description,
			COUNT(CASE WHEN ip.status = 'assigned' THEN 1 END) AS used,
			COUNT(ip.id) AS total
		FROM subnets s
		LEFT JOIN ip_addresses ip ON ip.subnet_id = s.id
		GROUP BY s.id, s.network_address, s.prefix_length, s.description
		HAVING COUNT(ip.id) > 0
		ORDER BY
			CASE WHEN COUNT(ip.id) > 0
				THEN COUNT(CASE WHEN ip.status = 'assigned' THEN 1 END)::float / COUNT(ip.id)
				ELSE 0
			END DESC
		LIMIT 5`

	topRows, err := r.db.Query(ctx, topQuery)
	if err != nil {
		return nil, fmt.Errorf("top subnets: %w", err)
	}
	defer topRows.Close()

	summary.TopSubnets = make([]models.SubnetUtilisation, 0)
	for topRows.Next() {
		su := models.SubnetUtilisation{}
		if err := topRows.Scan(&su.ID, &su.CIDR, &su.Description, &su.Used, &su.Total); err != nil {
			return nil, err
		}
		if su.Total > 0 {
			su.UtilisationPct = float64(su.Used) / float64(su.Total) * 100
		}
		summary.TopSubnets = append(summary.TopSubnets, su)
	}
	if err := topRows.Err(); err != nil {
		return nil, err
	}

	return summary, nil
}

// GetDashboardRecentActivity returns the last 20 relevant audit log entries.
func (r *Repository) GetDashboardRecentActivity(ctx context.Context) ([]*models.DashboardActivity, error) {
	query := `
		SELECT id, action, resource_type, resource_id, user_id, username, COALESCE(resource_name, ''), created_at
		FROM audit_logs
		WHERE action IN ('ip_assigned','ip_released','subnet_created','subnet_deleted','subnet_updated')
		ORDER BY created_at DESC
		LIMIT 20`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	activities := make([]*models.DashboardActivity, 0)
	for rows.Next() {
		a := &models.DashboardActivity{}
		var createdAt time.Time
		if err := rows.Scan(&a.ID, &a.Action, &a.EntityType, &a.EntityID, &a.UserID, &a.Username, &a.Description, &createdAt); err != nil {
			return nil, err
		}
		a.CreatedAt = createdAt.Format(time.RFC3339)
		activities = append(activities, a)
	}
	return activities, rows.Err()
}

// ListSectionsPaginated returns sections with pagination.
func (r *Repository) ListSectionsPaginated(ctx context.Context, limit, offset int) ([]*models.Section, int64, error) {
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM sections`).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, name, description, created_by, created_at, updated_at FROM sections ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	sections := make([]*models.Section, 0)
	for rows.Next() {
		section := &models.Section{}
		if err := rows.Scan(&section.ID, &section.Name, &section.Description, &section.CreatedBy, &section.CreatedAt, &section.UpdatedAt); err != nil {
			return nil, 0, err
		}
		sections = append(sections, section)
	}
	return sections, total, rows.Err()
}

// ListSubnetsBySectionPaginated returns subnets for a section with pagination.
func (r *Repository) ListSubnetsBySectionPaginated(ctx context.Context, sectionID int64, limit, offset int) ([]*models.Subnet, int64, error) {
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM subnets WHERE section_id = $1`, sectionID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, section_id, host(network_address), prefix_length, description, created_at, updated_at FROM subnets WHERE section_id = $1 ORDER BY network_address LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, sectionID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	subnets := make([]*models.Subnet, 0)
	for rows.Next() {
		subnet := &models.Subnet{}
		if err := rows.Scan(&subnet.ID, &subnet.SectionID, &subnet.NetworkAddress, &subnet.PrefixLength, &subnet.Description, &subnet.CreatedAt, &subnet.UpdatedAt); err != nil {
			return nil, 0, err
		}
		subnets = append(subnets, subnet)
	}
	return subnets, total, rows.Err()
}

// ListIPAddressesBySubnetPaginated returns IP addresses for a subnet with pagination.
func (r *Repository) ListIPAddressesBySubnetPaginated(ctx context.Context, subnetID int64, limit, offset int) ([]*models.IPAddress, int64, error) {
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM ip_addresses WHERE subnet_id = $1`, subnetID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, subnet_id, address::text, hostname, status, assigned_to, created_at, updated_at FROM ip_addresses WHERE subnet_id = $1 ORDER BY address LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, subnetID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	ips := make([]*models.IPAddress, 0)
	for rows.Next() {
		ip := &models.IPAddress{}
		if err := rows.Scan(&ip.ID, &ip.SubnetID, &ip.Address, &ip.Hostname, &ip.Status, &ip.AssignedTo, &ip.CreatedAt, &ip.UpdatedAt); err != nil {
			return nil, 0, err
		}
		ips = append(ips, ip)
	}
	return ips, total, rows.Err()
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
