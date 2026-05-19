package repository

import (
	"context"
	"encoding/json"

	"ipam-next/models"
)

// GetDeviceFingerprint returns the fingerprint for a device, or nil if none exists.
func (r *Repository) GetDeviceFingerprint(ctx context.Context, deviceID int64) (*models.DeviceFingerprint, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, device_id, open_ports, os_guess, vendor_guess, confidence_score, evidence, last_updated_at
		FROM device_fingerprints WHERE device_id = $1`, deviceID)
	return scanFingerprint(row)
}

// UpsertDeviceFingerprint creates or updates the fingerprint for a device.
func (r *Repository) UpsertDeviceFingerprint(ctx context.Context, deviceID int64, openPorts []int, osGuess, vendorGuess *string, confidenceScore float64, evidence []string) (*models.DeviceFingerprint, error) {
	portsJSON, _ := json.Marshal(openPorts)
	evidenceJSON, _ := json.Marshal(evidence)
	row := r.db.QueryRow(ctx, `
		INSERT INTO device_fingerprints (device_id, open_ports, os_guess, vendor_guess, confidence_score, evidence, last_updated_at)
		VALUES ($1, $2::jsonb, $3, $4, $5, $6::jsonb, now())
		ON CONFLICT (device_id) DO UPDATE SET
			open_ports = EXCLUDED.open_ports,
			os_guess = EXCLUDED.os_guess,
			vendor_guess = EXCLUDED.vendor_guess,
			confidence_score = EXCLUDED.confidence_score,
			evidence = EXCLUDED.evidence,
			last_updated_at = now()
		RETURNING id, device_id, open_ports, os_guess, vendor_guess, confidence_score, evidence, last_updated_at`,
		deviceID, string(portsJSON), osGuess, vendorGuess, confidenceScore, string(evidenceJSON))
	return scanFingerprint(row)
}

func scanFingerprint(row interface{ Scan(dest ...any) error }) (*models.DeviceFingerprint, error) {
	f := &models.DeviceFingerprint{}
	var portsJSON, evidenceJSON []byte
	if err := row.Scan(&f.ID, &f.DeviceID, &portsJSON, &f.OSGuess, &f.VendorGuess, &f.ConfidenceScore, &evidenceJSON, &f.LastUpdatedAt); err != nil {
		return nil, err
	}
	if portsJSON != nil {
		_ = json.Unmarshal(portsJSON, &f.OpenPorts)
	}
	if evidenceJSON != nil {
		_ = json.Unmarshal(evidenceJSON, &f.Evidence)
	}
	if f.OpenPorts == nil {
		f.OpenPorts = []int{}
	}
	if f.Evidence == nil {
		f.Evidence = []string{}
	}
	return f, nil
}
