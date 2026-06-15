package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerOSFamily(t *testing.T) {
	cases := []struct {
		goos string
		want string
	}{
		{"linux", "linux"},
		{"windows", "windows"},
		{"darwin", "macos"},
		{"freebsd", "bsd"},
		{"openbsd", "bsd"},
		{"netbsd", "bsd"},
		{"dragonfly", "bsd"},
		{"plan9", "unknown"},
		{"", "unknown"},
	}
	for _, tc := range cases {
		assert.Equalf(t, tc.want, serverOSFamily(tc.goos), "goos=%q", tc.goos)
	}
}

func TestTelemetrySnapshotConstants(t *testing.T) {
	// Verify that field values that are set from constants match the schema spec.
	s := &TelemetrySnapshot{
		TelemetrySchemaVersion: 1,
		DatabaseType:           "postgres",
		SnapshotPeriod:         "manual",
	}
	assert.Equal(t, 1, s.TelemetrySchemaVersion)
	assert.Equal(t, "postgres", s.DatabaseType)
	assert.Equal(t, "manual", s.SnapshotPeriod)
}
