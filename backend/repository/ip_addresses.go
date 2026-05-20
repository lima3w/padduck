package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"padduck/models"
)

// IP Address operations

// ipSelectCols is the column list for ip_addresses JOINed with ip_tags
const ipSelectCols = `ip.id, ip.subnet_id, ip.address::text, ip.hostname, ip.status, ip.assigned_to,
	ip.tag_id, t.id, t.name, t.colour, t.description, t.is_system, t.created_at,
	ip.last_seen, ip.mac_address, ip.ptr_record,
	ip.dns_name, ip.dns_records::text, ip.dns_last_checked,
	ip.port_open,
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
	var portOpenRaw []byte

	err := row.Scan(
		&ip.ID, &ip.SubnetID, &ip.Address, &ip.Hostname, &ip.Status, &ip.AssignedTo,
		&tagID, &tagIDInner, &tagName, &tagColour, &tagDesc, &tagIsSystem, &tagCreatedAt,
		&ip.LastSeen, &ip.MACAddress, &ip.PTRRecord,
		&ip.DNSName, &ip.DNSRecords, &ip.DNSLastChecked,
		&portOpenRaw,
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
	if len(portOpenRaw) > 0 {
		if err2 := json.Unmarshal(portOpenRaw, &ip.PortOpen); err2 != nil {
			ip.PortOpen = nil
		}
	}
	return ip, nil
}

func (r *Repository) CreateIPAddress(ctx context.Context, subnetID int64, address, hostname string, status string, assignedTo *string, tagID *int64, macAddress, ptrRecord *string) (*models.IPAddress, error) {
	query := `WITH ins AS (
		INSERT INTO ip_addresses (subnet_id, address, hostname, status, assigned_to, tag_id, mac_address, ptr_record)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
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

// GetSubnetUtilizationCounts returns total, available, assigned, and reserved IP counts for a subnet in a single query.
func (r *Repository) GetSubnetUtilizationCounts(ctx context.Context, subnetID int64) (total, available, assigned, reserved int64, err error) {
	err = r.db.QueryRow(ctx, `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = 'available') AS available,
			COUNT(*) FILTER (WHERE status = 'assigned') AS assigned,
			COUNT(*) FILTER (WHERE status = 'reserved') AS reserved
		FROM ip_addresses WHERE subnet_id = $1
	`, subnetID).Scan(&total, &available, &assigned, &reserved)
	return
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

// UpdateIPDNSFields stores the result of a DNS check on an IP address.
func (r *Repository) UpdateIPDNSFields(ctx context.Context, ipID int64, ptrRecord string, dnsRecords json.RawMessage, lastChecked time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE ip_addresses SET ptr_record=$2, dns_records=$3, dns_last_checked=$4, updated_at=now() WHERE id=$1`,
		ipID, ptrRecord, string(dnsRecords), lastChecked,
	)
	return err
}

// ListIPAddressesWithDNSName returns all IP addresses that have a dns_name set.
func (r *Repository) ListIPAddressesWithDNSName(ctx context.Context) ([]*models.IPAddress, error) {
	query := `SELECT ` + ipSelectCols + ` ` + ipFromJoin + ` WHERE ip.dns_name IS NOT NULL AND ip.dns_name != '' ORDER BY ip.id`
	rows, err := r.db.Query(ctx, query)
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

// UpdateIPPortScan stores the result of a port scan on an IP address.
func (r *Repository) UpdateIPPortScan(ctx context.Context, ipID int64, ports map[string]bool) error {
	data, err := json.Marshal(ports)
	if err != nil {
		return fmt.Errorf("marshal port_open: %w", err)
	}
	_, err = r.db.Exec(ctx,
		`UPDATE ip_addresses SET port_open=$2, updated_at=now() WHERE id=$1`,
		ipID, string(data),
	)
	return err
}
