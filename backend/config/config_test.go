package config

import (
	"encoding/hex"
	"os"
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

func TestLoadMFAEncryptionKeyCreatesPersistentProductionKey(t *testing.T) {
	tmp := t.TempDir()
	previous, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(previous))
	})

	key := loadMFAEncryptionKey("", true)
	decoded, err := hex.DecodeString(key)
	require.NoError(t, err)
	assert.Len(t, decoded, 32)
	assert.False(t, isWeakKey(decoded))

	data, err := os.ReadFile(PersistentMFAKeyPath())
	require.NoError(t, err)
	assert.Equal(t, key+"\n", string(data))

	info, err := os.Stat(PersistentMFAKeyPath())
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestLoadMFAEncryptionKeyReusesPersistentProductionKey(t *testing.T) {
	tmp := t.TempDir()
	previous, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(previous))
	})

	first := loadMFAEncryptionKey("", true)
	second := loadMFAEncryptionKey("", true)
	assert.Equal(t, first, second)
}
