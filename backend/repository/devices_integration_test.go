package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestDevice(t *testing.T, r *Repository, hostname, vendor, model string) int64 {
	t.Helper()
	d, err := r.CreateDevice(context.Background(), &DeviceParams{
		Hostname: hostname,
		Vendor:   strPtr(vendor),
		Model:    strPtr(model),
	})
	require.NoError(t, err)
	return d.ID
}

func TestDeviceCRUD_Integration(t *testing.T) {
	r := testRepo(t)
	ctx := context.Background()

	created, err := r.CreateDevice(ctx, &DeviceParams{
		Hostname:    "core-sw-01",
		Description: strPtr("core switch"),
		Vendor:      strPtr("Juniper"),
		Model:       strPtr("EX4300"),
	})
	require.NoError(t, err)
	require.NotZero(t, created.ID)

	got, err := r.GetDeviceByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "core-sw-01", got.Hostname)
	require.NotNil(t, got.Vendor)
	assert.Equal(t, "Juniper", *got.Vendor)

	_, err = r.UpdateDevice(ctx, created.ID, &DeviceParams{
		Hostname: "core-sw-01-renamed",
		Vendor:   strPtr("Juniper"),
		Model:    strPtr("EX4400"),
	})
	require.NoError(t, err)
	got, err = r.GetDeviceByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "core-sw-01-renamed", got.Hostname)
	require.NotNil(t, got.Model)
	assert.Equal(t, "EX4400", *got.Model)

	require.NoError(t, r.DeleteDevice(ctx, created.ID))
	_, err = r.GetDeviceByID(ctx, created.ID)
	assert.Error(t, err, "deleted device must not be retrievable")
}

func TestListDevicesWithOptions_QueryFilter_Integration(t *testing.T) {
	r := testRepo(t)
	ctx := context.Background()

	createTestDevice(t, r, "edge-router-01", "Cisco", "ASR1001")
	createTestDevice(t, r, "core-switch-01", "Arista", "7050X")
	createTestDevice(t, r, "backup-nas-01", "Synology", "RS1221")

	// Match by hostname substring (case-insensitive).
	devices, total, err := r.ListDevicesWithOptions(ctx, ListOptions{Query: "EDGE-ROUTER", Limit: 10})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, devices, 1)
	assert.Equal(t, "edge-router-01", devices[0].Hostname)

	// Match by vendor.
	_, total, err = r.ListDevicesWithOptions(ctx, ListOptions{Query: "arista", Limit: 10})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)

	// Match by model.
	_, total, err = r.ListDevicesWithOptions(ctx, ListOptions{Query: "RS1221", Limit: 10})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)

	// No match.
	devices, total, err = r.ListDevicesWithOptions(ctx, ListOptions{Query: "does-not-exist", Limit: 10})
	require.NoError(t, err)
	assert.EqualValues(t, 0, total)
	assert.Empty(t, devices)
}

func TestListDevicesWithOptions_SortAndOrder_Integration(t *testing.T) {
	r := testRepo(t)
	ctx := context.Background()

	createTestDevice(t, r, "bravo", "V2", "M2")
	createTestDevice(t, r, "alpha", "V1", "M1")
	createTestDevice(t, r, "charlie", "V3", "M3")

	devices, _, err := r.ListDevicesWithOptions(ctx, ListOptions{Sort: "hostname", Order: "asc", Limit: 10})
	require.NoError(t, err)
	require.Len(t, devices, 3)
	assert.Equal(t, []string{"alpha", "bravo", "charlie"},
		[]string{devices[0].Hostname, devices[1].Hostname, devices[2].Hostname})

	devices, _, err = r.ListDevicesWithOptions(ctx, ListOptions{Sort: "hostname", Order: "desc", Limit: 10})
	require.NoError(t, err)
	require.Len(t, devices, 3)
	assert.Equal(t, "charlie", devices[0].Hostname)

	// A sort column outside the allowlist must fall back, not be interpolated.
	devices, _, err = r.ListDevicesWithOptions(ctx, ListOptions{Sort: "1; DROP TABLE devices; --", Order: "asc", Limit: 10})
	require.NoError(t, err)
	require.Len(t, devices, 3)
	assert.Equal(t, "alpha", devices[0].Hostname, "bogus sort key must fall back to hostname")
}

func TestListDevicesWithOptions_Pagination_Integration(t *testing.T) {
	r := testRepo(t)
	ctx := context.Background()

	for i := 1; i <= 5; i++ {
		createTestDevice(t, r, fmt.Sprintf("dev-%02d", i), "V", "M")
	}

	// First page.
	devices, total, err := r.ListDevicesWithOptions(ctx, ListOptions{Sort: "hostname", Order: "asc", Limit: 2, Offset: 0})
	require.NoError(t, err)
	assert.EqualValues(t, 5, total)
	require.Len(t, devices, 2)
	assert.Equal(t, "dev-01", devices[0].Hostname)

	// Last page is partial.
	devices, total, err = r.ListDevicesWithOptions(ctx, ListOptions{Sort: "hostname", Order: "asc", Limit: 2, Offset: 4})
	require.NoError(t, err)
	assert.EqualValues(t, 5, total)
	require.Len(t, devices, 1)
	assert.Equal(t, "dev-05", devices[0].Hostname)

	// Offset past the end: empty page but correct total.
	devices, total, err = r.ListDevicesWithOptions(ctx, ListOptions{Sort: "hostname", Order: "asc", Limit: 2, Offset: 100})
	require.NoError(t, err)
	assert.EqualValues(t, 5, total)
	assert.Empty(t, devices)

	// Limit larger than the table.
	devices, _, err = r.ListDevicesWithOptions(ctx, ListOptions{Sort: "hostname", Order: "asc", Limit: 50, Offset: 0})
	require.NoError(t, err)
	assert.Len(t, devices, 5)
}

func TestListDevicesWithOptions_InjectionAttempt_Integration(t *testing.T) {
	r := testRepo(t)
	ctx := context.Background()

	createTestDevice(t, r, "victim", "V", "M")

	// The query value must be parameterized: no error, no match, table intact.
	devices, total, err := r.ListDevicesWithOptions(ctx, ListOptions{Query: `'; DROP TABLE devices; --`, Limit: 10})
	require.NoError(t, err)
	assert.EqualValues(t, 0, total)
	assert.Empty(t, devices)

	devices, total, err = r.ListDevicesWithOptions(ctx, ListOptions{Limit: 10})
	require.NoError(t, err)
	assert.EqualValues(t, 1, total, "devices table must survive an injection attempt")
	require.Len(t, devices, 1)
	assert.Equal(t, "victim", devices[0].Hostname)
}
