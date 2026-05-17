package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// ConfigService — nil repository paths
// ---------------------------------------------------------------------------

func TestConfigService_Get_NilRepository_ReturnsError(t *testing.T) {
	svc := NewConfigService(nil)
	_, err := svc.Get("any_key")
	assert.Error(t, err, "Get with nil repository must return an error")
	assert.Contains(t, err.Error(), "config")
}

func TestConfigService_Set_NilRepository_ReturnsError(t *testing.T) {
	svc := NewConfigService(nil)
	err := svc.Set("any_key", "value")
	assert.Error(t, err, "Set with nil repository must return an error")
	assert.Contains(t, err.Error(), "config")
}

func TestConfigService_List_NilRepository_ReturnsError(t *testing.T) {
	svc := NewConfigService(nil)
	list, err := svc.List()
	assert.Error(t, err, "List with nil repository must return an error")
	assert.Nil(t, list, "returned slice must be nil on error")
}

// ---------------------------------------------------------------------------
// ConfigService — IsRegistrationEnabled defaults
// ---------------------------------------------------------------------------

// With a nil repository Get returns an error, so IsRegistrationEnabled falls
// through to the default-open path and returns true.
func TestConfigService_IsRegistrationEnabled_NilRepository_ReturnsTrue(t *testing.T) {
	svc := NewConfigService(nil)
	enabled := svc.IsRegistrationEnabled()
	assert.True(t, enabled, "registration must be enabled by default when repository is unavailable")
}

// ---------------------------------------------------------------------------
// ConfigService — IsSMTPConfigured defaults
// ---------------------------------------------------------------------------

// With a nil repository Get returns an error, so IsSMTPConfigured cannot
// retrieve smtp_host and returns false.
func TestConfigService_IsSMTPConfigured_NilRepository_ReturnsFalse(t *testing.T) {
	svc := NewConfigService(nil)
	configured := svc.IsSMTPConfigured()
	assert.False(t, configured, "SMTP must be reported as not configured when repository is unavailable")
}

// ---------------------------------------------------------------------------
// ConfigService — IsEmailVerificationRequired defaults
// ---------------------------------------------------------------------------

func TestConfigService_IsEmailVerificationRequired_NilRepository_ReturnsFalse(t *testing.T) {
	svc := NewConfigService(nil)
	required := svc.IsEmailVerificationRequired()
	assert.False(t, required, "email verification must not be required when SMTP is not configured")
}

// ---------------------------------------------------------------------------
// ConfigService — IsAdminApprovalRequired defaults
// ---------------------------------------------------------------------------

func TestConfigService_IsAdminApprovalRequired_NilRepository_ReturnsFalse(t *testing.T) {
	svc := NewConfigService(nil)
	required := svc.IsAdminApprovalRequired()
	assert.False(t, required, "admin approval must not be required when repository is unavailable")
}
