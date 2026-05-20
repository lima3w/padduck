package config

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateMFAKeyGeneratesEphemeralDevelopmentKey(t *testing.T) {
	key := validateMFAKey("", false)
	decoded, err := hex.DecodeString(key)
	require.NoError(t, err)
	assert.Len(t, decoded, 32)
	assert.False(t, isWeakKey(decoded))
}

func TestValidateMFAKeyRejectsWeakDevelopmentKey(t *testing.T) {
	key := validateMFAKey("0000000000000000000000000000000000000000000000000000000000000000", false)
	decoded, err := hex.DecodeString(key)
	require.NoError(t, err)
	assert.Len(t, decoded, 32)
	assert.False(t, isWeakKey(decoded))
}
