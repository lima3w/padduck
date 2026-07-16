package config

import (
	"encoding/hex"
	"net/url"
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

func clearDBEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{"DATABASE_URL", "POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB", "POSTGRES_HOST", "POSTGRES_PORT"} {
		orig, had := os.LookupEnv(key)
		require.NoError(t, os.Unsetenv(key))
		t.Cleanup(func() {
			if had {
				os.Setenv(key, orig)
			} else {
				os.Unsetenv(key)
			}
		})
	}
}

func TestBuildDatabaseURLPrefersExplicitOverride(t *testing.T) {
	clearDBEnv(t)
	os.Setenv("DATABASE_URL", "postgres://custom:conn@example.com:5432/mydb")
	os.Setenv("POSTGRES_PASSWORD", "should-be-ignored")

	assert.Equal(t, "postgres://custom:conn@example.com:5432/mydb", buildDatabaseURL())
}

func TestBuildDatabaseURLFallsBackToDefaultWhenNothingSet(t *testing.T) {
	clearDBEnv(t)

	assert.Equal(t, "postgres://ipam:ipam@localhost:5432/ipam", buildDatabaseURL())
}

func TestBuildDatabaseURLEscapesSpecialCharactersInPassword(t *testing.T) {
	clearDBEnv(t)
	os.Setenv("POSTGRES_PASSWORD", "p@ss:w/ord#1?%")

	got := buildDatabaseURL()

	u, err := url.Parse(got)
	require.NoError(t, err)
	password, ok := u.User.Password()
	require.True(t, ok)
	assert.Equal(t, "p@ss:w/ord#1?%", password)
	assert.Equal(t, "padduck", u.User.Username())
	assert.Equal(t, "db:5432", u.Host)
	assert.Equal(t, "/padduck", u.Path)
}

func TestBuildDatabaseURLUsesPostgresPartsWhenProvided(t *testing.T) {
	clearDBEnv(t)
	os.Setenv("POSTGRES_USER", "customuser")
	os.Setenv("POSTGRES_PASSWORD", "secret")
	os.Setenv("POSTGRES_DB", "customdb")
	os.Setenv("POSTGRES_HOST", "otherhost")
	os.Setenv("POSTGRES_PORT", "5433")

	assert.Equal(t, "postgres://customuser:secret@otherhost:5433/customdb", buildDatabaseURL())
}
