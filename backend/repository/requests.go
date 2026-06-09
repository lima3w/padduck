package repository

import (
	"context"
	"fmt"

	"padduck/models"
)

// ---- Subnet Requests ----

// CreateSubnetRequest inserts a new subnet request.
func (r *Repository) CreateSubnetRequest(ctx context.Context, requesterID, networkID int64, parentSubnetID *int64, prefixLen int, purpose string) (*models.SubnetRequest, error) {
	query := `
		INSERT INTO subnet_requests (requester_id, network_id, parent_subnet_id, requested_prefix_len, purpose)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, requester_id, network_id, parent_subnet_id, requested_prefix_len, purpose, status, reviewer_id, reviewer_note, subnet_id, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, requesterID, networkID, parentSubnetID, prefixLen, purpose)
	return scanSubnetRequest(row)
}

// GetSubnetRequestByID returns a subnet request by ID, joining requester/reviewer usernames.
func (r *Repository) GetSubnetRequestByID(ctx context.Context, id int64) (*models.SubnetRequest, error) {
	query := `
		SELECT sr.id, sr.requester_id, COALESCE(ru.username,''), sr.network_id, sr.parent_subnet_id,
		       sr.requested_prefix_len, sr.purpose, sr.status, sr.reviewer_id, COALESCE(rv.username,''),
		       sr.reviewer_note, sr.subnet_id, sr.created_at, sr.updated_at
		FROM subnet_requests sr
		LEFT JOIN users ru ON ru.id = sr.requester_id
		LEFT JOIN users rv ON rv.id = sr.reviewer_id
		WHERE sr.id = $1`
	row := r.db.QueryRow(ctx, query, id)
	return scanSubnetRequestFull(row)
}

// ListSubnetRequestsByRequester returns all subnet requests for a specific requester.
func (r *Repository) ListSubnetRequestsByRequester(ctx context.Context, requesterID int64) ([]*models.SubnetRequest, error) {
	query := `
		SELECT sr.id, sr.requester_id, COALESCE(ru.username,''), sr.network_id, sr.parent_subnet_id,
		       sr.requested_prefix_len, sr.purpose, sr.status, sr.reviewer_id, COALESCE(rv.username,''),
		       sr.reviewer_note, sr.subnet_id, sr.created_at, sr.updated_at
		FROM subnet_requests sr
		LEFT JOIN users ru ON ru.id = sr.requester_id
		LEFT JOIN users rv ON rv.id = sr.reviewer_id
		WHERE sr.requester_id = $1
		ORDER BY sr.created_at DESC`
	return r.querySubnetRequests(ctx, query, requesterID)
}

// ListAllSubnetRequests returns all subnet requests.
func (r *Repository) ListAllSubnetRequests(ctx context.Context) ([]*models.SubnetRequest, error) {
	query := `
		SELECT sr.id, sr.requester_id, COALESCE(ru.username,''), sr.network_id, sr.parent_subnet_id,
		       sr.requested_prefix_len, sr.purpose, sr.status, sr.reviewer_id, COALESCE(rv.username,''),
		       sr.reviewer_note, sr.subnet_id, sr.created_at, sr.updated_at
		FROM subnet_requests sr
		LEFT JOIN users ru ON ru.id = sr.requester_id
		LEFT JOIN users rv ON rv.id = sr.reviewer_id
		ORDER BY sr.created_at DESC`
	return r.querySubnetRequests(ctx, query)
}

// ApproveSubnetRequest sets a subnet request to approved and links the created subnet.
func (r *Repository) ApproveSubnetRequest(ctx context.Context, id, reviewerID, subnetID int64, reviewerNote string) (*models.SubnetRequest, error) {
	query := `
		UPDATE subnet_requests
		SET status = 'approved', reviewer_id = $2, reviewer_note = $3, subnet_id = $4, updated_at = NOW()
		WHERE id = $1 AND status = 'pending'
		RETURNING id`
	var updatedID int64
	err := r.db.QueryRow(ctx, query, id, reviewerID, reviewerNote, subnetID).Scan(&updatedID)
	if err != nil {
		return nil, fmt.Errorf("subnet request not found or not pending")
	}
	return r.GetSubnetRequestByID(ctx, updatedID)
}

// RejectSubnetRequest sets a subnet request to rejected.
func (r *Repository) RejectSubnetRequest(ctx context.Context, id, reviewerID int64, reviewerNote string) (*models.SubnetRequest, error) {
	query := `
		UPDATE subnet_requests
		SET status = 'rejected', reviewer_id = $2, reviewer_note = $3, updated_at = NOW()
		WHERE id = $1 AND status = 'pending'
		RETURNING id`
	var updatedID int64
	err := r.db.QueryRow(ctx, query, id, reviewerID, reviewerNote).Scan(&updatedID)
	if err != nil {
		return nil, fmt.Errorf("subnet request not found or not pending")
	}
	return r.GetSubnetRequestByID(ctx, updatedID)
}

// CancelSubnetRequest cancels a pending subnet request (only by requester).
func (r *Repository) CancelSubnetRequest(ctx context.Context, id, requesterID int64) error {
	query := `
		UPDATE subnet_requests
		SET status = 'cancelled', updated_at = NOW()
		WHERE id = $1 AND requester_id = $2 AND status = 'pending'`
	result, err := r.db.Exec(ctx, query, id, requesterID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("subnet request not found or not cancellable")
	}
	return nil
}

// CountPendingSubnetRequests returns the count of pending subnet requests.
func (r *Repository) CountPendingSubnetRequests(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM subnet_requests WHERE status = 'pending'`).Scan(&count)
	return count, err
}

func (r *Repository) querySubnetRequests(ctx context.Context, query string, args ...interface{}) ([]*models.SubnetRequest, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*models.SubnetRequest, 0)
	for rows.Next() {
		sr, err := scanSubnetRequestFull(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, sr)
	}
	return result, rows.Err()
}

type subnetRequestScanner interface {
	Scan(dest ...interface{}) error
}

func scanSubnetRequest(s subnetRequestScanner) (*models.SubnetRequest, error) {
	sr := &models.SubnetRequest{}
	err := s.Scan(
		&sr.ID, &sr.RequesterID, &sr.NetworkID, &sr.ParentSubnetID,
		&sr.RequestedPrefixLen, &sr.Purpose, &sr.Status,
		&sr.ReviewerID, &sr.ReviewerNote, &sr.SubnetID,
		&sr.CreatedAt, &sr.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return sr, nil
}

func scanSubnetRequestFull(s subnetRequestScanner) (*models.SubnetRequest, error) {
	sr := &models.SubnetRequest{}
	err := s.Scan(
		&sr.ID, &sr.RequesterID, &sr.RequesterUsername, &sr.NetworkID, &sr.ParentSubnetID,
		&sr.RequestedPrefixLen, &sr.Purpose, &sr.Status,
		&sr.ReviewerID, &sr.ReviewerUsername,
		&sr.ReviewerNote, &sr.SubnetID,
		&sr.CreatedAt, &sr.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return sr, nil
}

// ---- IP Requests ----

// CreateIPRequest inserts a new IP request.
func (r *Repository) CreateIPRequest(ctx context.Context, requesterID, subnetID int64, requestedIP *string, dnsName, purpose string) (*models.IPRequest, error) {
	query := `
		INSERT INTO ip_requests (requester_id, subnet_id, requested_ip, dns_name, purpose)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, requester_id, subnet_id, requested_ip, dns_name, purpose, status, reviewer_id, reviewer_note, ip_address_id, created_at, updated_at`
	row := r.db.QueryRow(ctx, query, requesterID, subnetID, requestedIP, dnsName, purpose)
	return scanIPRequest(row)
}

// GetIPRequestByID returns an IP request by ID, joining requester/reviewer usernames.
func (r *Repository) GetIPRequestByID(ctx context.Context, id int64) (*models.IPRequest, error) {
	query := `
		SELECT ir.id, ir.requester_id, COALESCE(ru.username,''), ir.subnet_id, ir.requested_ip,
		       ir.dns_name, ir.purpose, ir.status, ir.reviewer_id, COALESCE(rv.username,''),
		       ir.reviewer_note, ir.ip_address_id, ir.created_at, ir.updated_at
		FROM ip_requests ir
		LEFT JOIN users ru ON ru.id = ir.requester_id
		LEFT JOIN users rv ON rv.id = ir.reviewer_id
		WHERE ir.id = $1`
	row := r.db.QueryRow(ctx, query, id)
	return scanIPRequestFull(row)
}

// ListIPRequestsByRequester returns all IP requests for a specific requester.
func (r *Repository) ListIPRequestsByRequester(ctx context.Context, requesterID int64) ([]*models.IPRequest, error) {
	query := `
		SELECT ir.id, ir.requester_id, COALESCE(ru.username,''), ir.subnet_id, ir.requested_ip,
		       ir.dns_name, ir.purpose, ir.status, ir.reviewer_id, COALESCE(rv.username,''),
		       ir.reviewer_note, ir.ip_address_id, ir.created_at, ir.updated_at
		FROM ip_requests ir
		LEFT JOIN users ru ON ru.id = ir.requester_id
		LEFT JOIN users rv ON rv.id = ir.reviewer_id
		WHERE ir.requester_id = $1
		ORDER BY ir.created_at DESC`
	return r.queryIPRequests(ctx, query, requesterID)
}

// ListAllIPRequests returns all IP requests.
func (r *Repository) ListAllIPRequests(ctx context.Context) ([]*models.IPRequest, error) {
	query := `
		SELECT ir.id, ir.requester_id, COALESCE(ru.username,''), ir.subnet_id, ir.requested_ip,
		       ir.dns_name, ir.purpose, ir.status, ir.reviewer_id, COALESCE(rv.username,''),
		       ir.reviewer_note, ir.ip_address_id, ir.created_at, ir.updated_at
		FROM ip_requests ir
		LEFT JOIN users ru ON ru.id = ir.requester_id
		LEFT JOIN users rv ON rv.id = ir.reviewer_id
		ORDER BY ir.created_at DESC`
	return r.queryIPRequests(ctx, query)
}

// ApproveIPRequest sets an IP request to approved and links the assigned IP address.
func (r *Repository) ApproveIPRequest(ctx context.Context, id, reviewerID, ipAddressID int64, reviewerNote string) (*models.IPRequest, error) {
	query := `
		UPDATE ip_requests
		SET status = 'approved', reviewer_id = $2, reviewer_note = $3, ip_address_id = $4, updated_at = NOW()
		WHERE id = $1 AND status = 'pending'
		RETURNING id`
	var updatedID int64
	err := r.db.QueryRow(ctx, query, id, reviewerID, reviewerNote, ipAddressID).Scan(&updatedID)
	if err != nil {
		return nil, fmt.Errorf("ip request not found or not pending")
	}
	return r.GetIPRequestByID(ctx, updatedID)
}

// RejectIPRequest sets an IP request to rejected.
func (r *Repository) RejectIPRequest(ctx context.Context, id, reviewerID int64, reviewerNote string) (*models.IPRequest, error) {
	query := `
		UPDATE ip_requests
		SET status = 'rejected', reviewer_id = $2, reviewer_note = $3, updated_at = NOW()
		WHERE id = $1 AND status = 'pending'
		RETURNING id`
	var updatedID int64
	err := r.db.QueryRow(ctx, query, id, reviewerID, reviewerNote).Scan(&updatedID)
	if err != nil {
		return nil, fmt.Errorf("ip request not found or not pending")
	}
	return r.GetIPRequestByID(ctx, updatedID)
}

// CancelIPRequest cancels a pending IP request (only by requester).
func (r *Repository) CancelIPRequest(ctx context.Context, id, requesterID int64) error {
	query := `
		UPDATE ip_requests
		SET status = 'cancelled', updated_at = NOW()
		WHERE id = $1 AND requester_id = $2 AND status = 'pending'`
	result, err := r.db.Exec(ctx, query, id, requesterID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("ip request not found or not cancellable")
	}
	return nil
}

// CountPendingIPRequests returns the count of pending IP requests.
func (r *Repository) CountPendingIPRequests(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM ip_requests WHERE status = 'pending'`).Scan(&count)
	return count, err
}

// GetIPAddressBySubnetAndAddress finds an IP address in a subnet by its address string.
func (r *Repository) GetIPAddressBySubnetAndAddress(ctx context.Context, subnetID int64, address string) (*models.IPAddress, error) {
	query := `SELECT ` + ipSelectCols + ` ` + ipFromJoin + ` WHERE ip.subnet_id = $1 AND ip.address = $2::inet`
	row := r.db.QueryRow(ctx, query, subnetID, address)
	return scanIP(row)
}

func (r *Repository) queryIPRequests(ctx context.Context, query string, args ...interface{}) ([]*models.IPRequest, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*models.IPRequest, 0)
	for rows.Next() {
		ir, err := scanIPRequestFull(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, ir)
	}
	return result, rows.Err()
}

type ipRequestScanner interface {
	Scan(dest ...interface{}) error
}

func scanIPRequest(s ipRequestScanner) (*models.IPRequest, error) {
	ir := &models.IPRequest{}
	err := s.Scan(
		&ir.ID, &ir.RequesterID, &ir.SubnetID, &ir.RequestedIP,
		&ir.DNSName, &ir.Purpose, &ir.Status,
		&ir.ReviewerID, &ir.ReviewerNote, &ir.IPAddressID,
		&ir.CreatedAt, &ir.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return ir, nil
}

func scanIPRequestFull(s ipRequestScanner) (*models.IPRequest, error) {
	ir := &models.IPRequest{}
	err := s.Scan(
		&ir.ID, &ir.RequesterID, &ir.RequesterUsername, &ir.SubnetID, &ir.RequestedIP,
		&ir.DNSName, &ir.Purpose, &ir.Status,
		&ir.ReviewerID, &ir.ReviewerUsername,
		&ir.ReviewerNote, &ir.IPAddressID,
		&ir.CreatedAt, &ir.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return ir, nil
}

// ---- Request Comments ----

// CreateRequestComment inserts a new comment on a request.
func (r *Repository) CreateRequestComment(ctx context.Context, requestType string, requestID, authorID int64, body string) (*models.RequestComment, error) {
	query := `
		INSERT INTO request_comments (request_type, request_id, author_id, body)
		VALUES ($1::request_comment_type, $2, $3, $4)
		RETURNING id, request_type, request_id, author_id, body, created_at`
	row := r.db.QueryRow(ctx, query, requestType, requestID, authorID, body)

	c := &models.RequestComment{}
	err := row.Scan(&c.ID, &c.RequestType, &c.RequestID, &c.AuthorID, &c.Body, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	// Fetch author username
	user, err := r.GetUserByID(ctx, authorID)
	if err == nil {
		c.AuthorUsername = user.Username
	}
	return c, nil
}

// ListRequestComments returns all comments for a given request type and ID.
func (r *Repository) ListRequestComments(ctx context.Context, requestType string, requestID int64) ([]*models.RequestComment, error) {
	query := `
		SELECT rc.id, rc.request_type, rc.request_id, rc.author_id, COALESCE(u.username,''), rc.body, rc.created_at
		FROM request_comments rc
		LEFT JOIN users u ON u.id = rc.author_id
		WHERE rc.request_type = $1::request_comment_type AND rc.request_id = $2
		ORDER BY rc.created_at ASC`
	rows, err := r.db.Query(ctx, query, requestType, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*models.RequestComment, 0)
	for rows.Next() {
		c := &models.RequestComment{}
		if err := rows.Scan(&c.ID, &c.RequestType, &c.RequestID, &c.AuthorID, &c.AuthorUsername, &c.Body, &c.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

// UpdateIPDNSName sets the dns_name field on an ip_address record.
func (r *Repository) UpdateIPDNSName(ctx context.Context, ipID int64, dnsName string) error {
	_, err := r.db.Exec(ctx, `UPDATE ip_addresses SET dns_name = $2, updated_at = NOW() WHERE id = $1`, ipID, dnsName)
	return err
}

// GetUsersByRole returns all users with the given legacy role.
func (r *Repository) GetUsersByRole(ctx context.Context, role string) ([]*models.User, error) {
	query := `SELECT id, username, email, password_hash, role, state, last_login_at, suspended_at, suspended_by, suspension_reason, privacy_accepted_at, privacy_accepted_version, deletion_requested_at, anonymized_at, created_at, updated_at FROM users WHERE role = $1 AND state = 'active'`
	rows, err := r.db.Query(ctx, query, role)
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
