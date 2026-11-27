package services

import (
	"bytes"
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"padduck/internal/testdb"
	"padduck/repository"
)

func TestParseSecretRef(t *testing.T) {
	for in, want := range map[string]string{
		"secret://core-snmp": "core-snmp",
		"secret://a.b_c-1":   "a.b_c-1",
		"secret://":          "",
		"secret://bad name":  "",
		"secret://../etc":    "",
		"public":             "",
		"SECRET://upper":     "",
		"secret://-leading":  "",
	} {
		name, ok := ParseSecretRef(in)
		assert.Equal(t, want != "", ok, in)
		assert.Equal(t, want, name, in)
	}
}

func TestSecretsSealOpen(t *testing.T) {
	s, err := NewSecretsService(nil, nil, testMFAKey)
	require.NoError(t, err)

	d1, v1, err := s.seal("hunter2")
	require.NoError(t, err)
	assert.False(t, bytes.Contains(v1, []byte("hunter2")), "ciphertext must not contain plaintext")
	got, err := s.open(d1, v1)
	require.NoError(t, err)
	assert.Equal(t, "hunter2", got)

	// A fresh data key is used per seal, so the same value encrypts differently.
	d2, v2, err := s.seal("hunter2")
	require.NoError(t, err)
	assert.NotEqual(t, d1, d2)
	assert.NotEqual(t, v1, v2)

	// A different application key cannot open the data key.
	other, err := NewSecretsService(nil, nil, "ff"+testMFAKey[2:])
	require.NoError(t, err)
	_, err = other.open(d1, v1)
	assert.Error(t, err)

	_, _, err = s.seal("")
	assert.Error(t, err)
	_, err = NewSecretsService(nil, nil, "short")
	assert.Error(t, err)
}

func testSecretsService(t *testing.T) (*SecretsService, *pgxpool.Pool) {
	t.Helper()
	pool := testdb.Connect(t, "services")
	testdb.Truncate(t, pool, "secret_versions", "secrets")
	repo := repository.NewRepository(pool)
	s, err := NewSecretsService(repo, nil, testMFAKey)
	require.NoError(t, err)
	return s, pool
}

func TestSecretsLifecycle(t *testing.T) {
	s, pool := testSecretsService(t)
	ctx := context.Background()

	sec, err := s.Create(ctx, "core-snmp", "core switches", "v1-value", nil)
	require.NoError(t, err)
	assert.Equal(t, 1, sec.Version)
	assert.True(t, sec.Exists)

	_, err = s.Create(ctx, "core-snmp", "", "x", nil)
	assert.Error(t, err, "duplicate name in the same org")
	_, err = s.Create(ctx, "bad name", "", "x", nil)
	assert.Error(t, err)

	// The database never holds the plaintext.
	var n int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM secret_versions WHERE position('\x76312d76616c7565'::bytea in encrypted_value) > 0`).Scan(&n))
	assert.Zero(t, n)

	got, err := s.Resolve(ctx, "core-snmp", "test")
	require.NoError(t, err)
	assert.Equal(t, "v1-value", got)

	// Rotation adds a version; the name now resolves to the new value.
	rot, err := s.Rotate(ctx, sec.ID, "v2-value", nil)
	require.NoError(t, err)
	assert.Equal(t, 2, rot.Version)
	assert.NotNil(t, rot.RotatedAt)
	got, err = s.Resolve(ctx, "core-snmp", "test")
	require.NoError(t, err)
	assert.Equal(t, "v2-value", got)

	versions, err := s.Versions(ctx, sec.ID)
	require.NoError(t, err)
	require.Len(t, versions, 2)
	assert.True(t, versions[0].Current)
	assert.Equal(t, 2, versions[0].Version)

	// The current version cannot be deleted; an old one can.
	assert.Error(t, s.DeleteVersion(ctx, sec.ID, 2))
	require.NoError(t, s.DeleteVersion(ctx, sec.ID, 1))
	assert.ErrorIs(t, s.DeleteVersion(ctx, sec.ID, 1), pgx.ErrNoRows)

	// A deleted number is never reused.
	rot, err = s.Rotate(ctx, sec.ID, "v3-value", nil)
	require.NoError(t, err)
	assert.Equal(t, 3, rot.Version)

	// ResolveRef passes plain values through untouched.
	plain, err := s.ResolveRef(ctx, "public", "test")
	require.NoError(t, err)
	assert.Equal(t, "public", plain)
	viaRef, err := s.ResolveRef(ctx, "secret://core-snmp", "test")
	require.NoError(t, err)
	assert.Equal(t, "v3-value", viaRef)

	list, err := s.List(ctx)
	require.NoError(t, err)
	require.Len(t, list, 1)

	require.NoError(t, s.Delete(ctx, sec.ID))
	_, err = s.Resolve(ctx, "core-snmp", "test")
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.ErrorIs(t, s.Delete(ctx, sec.ID), pgx.ErrNoRows)
}
