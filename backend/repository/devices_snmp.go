package repository

import (
	"context"

	"padduck/models"
)

// GetDeviceSNMP returns the raw (encrypted) SNMP credentials for a device.
func (r *Repository) GetDeviceSNMP(ctx context.Context, id int64) (*models.DeviceSNMP, error) {
	query := `
		SELECT id, snmp_community, COALESCE(snmp_version, 'v2c'), snmp_v3_user,
		       snmp_v3_auth_proto, snmp_v3_auth_pass, snmp_v3_priv_proto, snmp_v3_priv_pass
		FROM devices WHERE id=$1`
	row := r.db.QueryRow(ctx, query, id)
	creds := &models.DeviceSNMP{}
	err := row.Scan(
		&creds.DeviceID, &creds.SNMPCommunity, &creds.SNMPVersion,
		&creds.SNMPV3User, &creds.SNMPV3AuthProto, &creds.SNMPV3AuthPass,
		&creds.SNMPV3PrivProto, &creds.SNMPV3PrivPass,
	)
	if err != nil {
		return nil, err
	}
	return creds, nil
}

// GetDeviceSNMPByIPID returns the raw (encrypted) SNMP credentials for the
// device linked to the given IP address ID. Returns nil, nil when the IP has
// no associated device.
func (r *Repository) GetDeviceSNMPByIPID(ctx context.Context, ipID int64) (*models.DeviceSNMP, error) {
	query := `
		SELECT d.id, d.snmp_community, COALESCE(d.snmp_version, 'v2c'), d.snmp_v3_user,
		       d.snmp_v3_auth_proto, d.snmp_v3_auth_pass, d.snmp_v3_priv_proto, d.snmp_v3_priv_pass
		FROM ip_addresses ip
		JOIN devices d ON d.id = ip.device_id
		WHERE ip.id = $1`
	row := r.db.QueryRow(ctx, query, ipID)
	creds := &models.DeviceSNMP{}
	err := row.Scan(
		&creds.DeviceID, &creds.SNMPCommunity, &creds.SNMPVersion,
		&creds.SNMPV3User, &creds.SNMPV3AuthProto, &creds.SNMPV3AuthPass,
		&creds.SNMPV3PrivProto, &creds.SNMPV3PrivPass,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	return creds, nil
}
