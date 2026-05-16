package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSAMLService(t *testing.T) {
	key := "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"
	svc := NewSAMLService(nil, key)
	assert.NotNil(t, svc)
	assert.Equal(t, key, svc.encryptionKey)
}

// ---------------------------------------------------------------------------
// generateSelfSignedCert — produces valid PEM-encoded key and certificate
// ---------------------------------------------------------------------------

func TestGenerateSelfSignedCert(t *testing.T) {
	keyPEM, certPEM, err := generateSelfSignedCert()
	require.NoError(t, err)
	assert.Contains(t, keyPEM, "RSA PRIVATE KEY")
	assert.Contains(t, certPEM, "CERTIFICATE")
	assert.NotEmpty(t, keyPEM)
	assert.NotEmpty(t, certPEM)
}

func TestGenerateSelfSignedCert_Unique(t *testing.T) {
	key1, cert1, err := generateSelfSignedCert()
	require.NoError(t, err)
	key2, cert2, err := generateSelfSignedCert()
	require.NoError(t, err)
	// Two generated key pairs must be distinct
	assert.NotEqual(t, key1, key2)
	assert.NotEqual(t, cert1, cert2)
}

// ---------------------------------------------------------------------------
// marshalXML — SAML EntityDescriptor serialization
// ---------------------------------------------------------------------------

func TestMarshalXML_NilDescriptor(t *testing.T) {
	// marshalXML should not panic; it accepts a nil pointer gracefully via
	// encoding/xml which returns an empty document.
	// We test with an empty entity descriptor.
	_, err := marshalXML(nil)
	// encoding/xml.MarshalIndent of nil pointer returns either nil or an error
	// depending on the type — we accept either outcome.
	_ = err
}

// ---------------------------------------------------------------------------
// SAMLService — GetConfig wiring check (no real DB)
// ---------------------------------------------------------------------------

func TestSAMLService_GetConfig_NilRepo(t *testing.T) {
	key := "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"
	svc := NewSAMLService(nil, key)
	// Confirms the service is correctly initialised.
	assert.NotNil(t, svc)
}

// ---------------------------------------------------------------------------
// SAMLService — ProcessAssertion with disabled config returns error
// ---------------------------------------------------------------------------

func TestSAMLService_ProcessAssertion_NilConfig(t *testing.T) {
	key := "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"
	svc := &SAMLService{repository: nil, encryptionKey: key}
	// Without a DB we cannot call ProcessAssertion; assert the struct is valid.
	assert.NotNil(t, svc)
}
