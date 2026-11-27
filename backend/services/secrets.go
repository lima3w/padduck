package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"

	"padduck/models"
	"padduck/repository"
	"padduck/utils"
)

// SecretRefPrefix marks a configuration value that refers to a stored secret
// by name instead of holding the credential inline, e.g. "secret://core-snmp".
const SecretRefPrefix = "secret://"

// secretNameRE keeps names usable inside a secret:// reference.
var secretNameRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

const maxSecretValueLen = 16 * 1024

// SecretsService stores integration credentials encrypted at rest. Every value
// is encrypted with its own random data key (DEK); the DEK is encrypted with the
// application key (KEK, the MFA/encryption key) and stored beside the value.
type SecretsService struct {
	repo  *repository.Repository
	audit *AuditService
	kek   []byte
	// keyHex is the same key, hex-encoded, for decrypting legacy inline credentials.
	keyHex string
}

// NewSecretsService builds the service from the hex-encoded 32-byte application key.
func NewSecretsService(repo *repository.Repository, audit *AuditService, keyHex string) (*SecretsService, error) {
	kek, err := hex.DecodeString(keyHex)
	if err != nil || len(kek) != 32 {
		return nil, fmt.Errorf("secrets: encryption key must be a 64-character hex string (32 bytes)")
	}
	return &SecretsService{repo: repo, audit: audit, kek: kek, keyHex: keyHex}, nil
}

// ParseSecretRef returns the secret name when v is a secret:// reference.
func ParseSecretRef(v string) (name string, ok bool) {
	if !strings.HasPrefix(v, SecretRefPrefix) {
		return "", false
	}
	name = strings.TrimPrefix(v, SecretRefPrefix)
	if !secretNameRE.MatchString(name) {
		return "", false
	}
	return name, true
}

// seal encrypts value under a fresh DEK and returns (encrypted DEK, encrypted value).
func (s *SecretsService) seal(value string) (encDEK, encValue []byte, err error) {
	if value == "" {
		return nil, nil, fmt.Errorf("secret value is required")
	}
	if len(value) > maxSecretValueLen {
		return nil, nil, fmt.Errorf("secret value must be at most %d bytes", maxSecretValueLen)
	}
	dek := make([]byte, 32)
	if _, err := rand.Read(dek); err != nil {
		return nil, nil, err
	}
	if encValue, err = encryptBytes(dek, []byte(value)); err != nil {
		return nil, nil, err
	}
	if encDEK, err = encryptBytes(s.kek, dek); err != nil {
		return nil, nil, err
	}
	return encDEK, encValue, nil
}

func (s *SecretsService) open(encDEK, encValue []byte) (string, error) {
	dek, err := decryptBytes(s.kek, encDEK)
	if err != nil {
		return "", fmt.Errorf("decrypting data key: %w", err)
	}
	pt, err := decryptBytes(dek, encValue)
	if err != nil {
		return "", fmt.Errorf("decrypting secret value: %w", err)
	}
	return string(pt), nil
}

// Create stores a new secret at version 1.
func (s *SecretsService) Create(ctx context.Context, name, description, value string, createdBy *int64) (*models.Secret, error) {
	if !secretNameRE.MatchString(name) {
		return nil, fmt.Errorf("invalid secret name: use letters, digits, '.', '_' or '-' (max 128 characters)")
	}
	encDEK, encValue, err := s.seal(value)
	if err != nil {
		return nil, err
	}
	return s.repo.CreateSecret(ctx, name, description, createdBy, encDEK, encValue)
}

// Get returns secret metadata.
func (s *SecretsService) Get(ctx context.Context, id int64) (*models.Secret, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid secret ID")
	}
	return s.repo.GetSecret(ctx, id)
}

// List returns metadata for all secrets.
func (s *SecretsService) List(ctx context.Context) ([]*models.Secret, error) {
	return s.repo.ListSecrets(ctx)
}

// Rotate stores value as a new version and makes it current. References by
// name keep working and resolve to the new value.
func (s *SecretsService) Rotate(ctx context.Context, id int64, value string, createdBy *int64) (*models.Secret, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid secret ID")
	}
	encDEK, encValue, err := s.seal(value)
	if err != nil {
		return nil, err
	}
	return s.repo.RotateSecret(ctx, id, createdBy, encDEK, encValue)
}

// Versions lists the stored versions of a secret, newest first.
func (s *SecretsService) Versions(ctx context.Context, id int64) ([]*models.SecretVersion, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid secret ID")
	}
	return s.repo.ListSecretVersions(ctx, id)
}

// DeleteVersion removes one old version. The current version cannot be removed.
func (s *SecretsService) DeleteVersion(ctx context.Context, id int64, version int) error {
	if id <= 0 || version <= 0 {
		return fmt.Errorf("invalid secret ID or version")
	}
	return s.repo.DeleteSecretVersion(ctx, id, version)
}

// Delete removes a secret and every version.
func (s *SecretsService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid secret ID")
	}
	return s.repo.DeleteSecret(ctx, id)
}

// Resolve returns the plaintext of the named secret's current version. It is for
// server-side use only (no API returns it) and every access is audit-logged.
// purpose says why the value was read (for example "snmp_scan").
func (s *SecretsService) Resolve(ctx context.Context, name, purpose string) (string, error) {
	sec, encDEK, encValue, err := s.repo.GetCurrentSecretCiphertext(ctx, name)
	if err != nil {
		return "", fmt.Errorf("resolving secret %q: %w", name, err)
	}
	val, err := s.open(encDEK, encValue)
	if err != nil {
		return "", err
	}
	if s.audit != nil {
		orgID := sec.OrganizationID
		s.audit.Log(ctx, AuditEntry{
			OrgID: &orgID, Action: "secret_accessed", ResourceType: "secret",
			ResourceID: &sec.ID, ResourceName: sec.Name,
			NewValues: map[string]any{"purpose": purpose, "version": sec.Version},
		})
	}
	return val, nil
}

// ResolveRef returns v unchanged unless it is a secret:// reference, in which
// case it returns the referenced secret's current value.
func (s *SecretsService) ResolveRef(ctx context.Context, v, purpose string) (string, error) {
	name, ok := ParseSecretRef(v)
	if !ok {
		return v, nil
	}
	return s.Resolve(ctx, name, purpose)
}

// ValidateRef checks a value that is about to be stored in a credential field.
// Plain values pass; a secret:// value must name a secret that exists.
func (s *SecretsService) ValidateRef(ctx context.Context, v string) error {
	if !strings.HasPrefix(v, SecretRefPrefix) {
		return nil
	}
	name, ok := ParseSecretRef(v)
	if !ok {
		return fmt.Errorf("invalid secret reference %q: use secret://<name> with letters, digits, '.', '_' or '-'", v)
	}
	if _, err := s.repo.GetSecretByName(ctx, name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("invalid secret reference: secret %q does not exist", name)
		}
		return err
	}
	return nil
}

// revealCredential returns the plaintext of a stored credential field: a
// secret:// reference is resolved through the secrets store, anything else is
// decrypted with decrypt (the legacy inline encryption).
func revealCredential(ctx context.Context, sec *SecretsService, stored, purpose string, decrypt func(string) (string, error)) (string, error) {
	if strings.HasPrefix(stored, SecretRefPrefix) {
		if sec == nil {
			return "", fmt.Errorf("secrets store unavailable")
		}
		return sec.ResolveRef(ctx, stored, purpose)
	}
	return decrypt(stored)
}

// MigrateDeviceSNMPCommunities moves every device's inline (encrypted) SNMP
// community string into a named secret "snmp-community-<device id>" and points
// the device at it with secret://snmp-community-<device id>. It is idempotent
// and safe to run on every start: devices already pointing at a secret are
// skipped, and a failure on one device leaves it unchanged and is retried on the
// next start. It returns how many devices were migrated.
func (s *SecretsService) MigrateDeviceSNMPCommunities(ctx context.Context) (int, error) {
	legacy, err := s.repo.ListDevicesWithInlineCommunity(ctx)
	if err != nil {
		return 0, err
	}
	migrated := 0
	for _, d := range legacy {
		plain, err := utils.DecryptString(d.Stored, s.keyHex)
		if err != nil || plain == "" {
			slog.Warn("secrets: leaving device SNMP community unmigrated (cannot decrypt)", "device_id", d.DeviceID)
			continue
		}
		encDEK, encValue, err := s.seal(plain)
		if err != nil {
			slog.Warn("secrets: SNMP community migration failed", "device_id", d.DeviceID, "error", err)
			continue
		}
		name := fmt.Sprintf("snmp-community-%d", d.DeviceID)
		ok, err := s.repo.MigrateDeviceCommunityToSecret(ctx, d, name, encDEK, encValue)
		if err != nil {
			slog.Warn("secrets: SNMP community migration failed", "device_id", d.DeviceID, "error", err)
			continue
		}
		if ok {
			migrated++
			if s.audit != nil {
				orgID := d.OrgID
				s.audit.Log(ctx, AuditEntry{
					OrgID: &orgID, Username: "system", Action: "secret_created",
					ResourceType: "secret", ResourceName: name,
					NewValues: map[string]any{"name": name, "source": "device_snmp_migration", "device_id": d.DeviceID},
				})
			}
		}
	}
	return migrated, nil
}
