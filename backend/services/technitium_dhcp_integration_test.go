package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"padduck/internal/testdb"
	"padduck/repository"
)

// fakeTechnitium serves the two DHCP endpoints the sync uses. Scope "bad"
// answers with an API error so partial failures can be exercised.
func fakeTechnitium(t *testing.T) *httptest.Server {
	t.Helper()
	ok := func(w http.ResponseWriter, resp any) {
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "response": resp})
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/dhcp/scopes/list", func(w http.ResponseWriter, _ *http.Request) {
		ok(w, map[string]any{"scopes": []map[string]any{{"name": "vlan10"}, {"name": "flat"}, {"name": "bad"}}})
	})
	mux.HandleFunc("/api/dhcp/leases/list", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("scopeName") {
		case "vlan10":
			ok(w, map[string]any{"leases": []map[string]any{
				{"address": "192.0.2.10", "hardwareAddress": "aa:aa:aa:aa:aa:01", "hostName": "a", "type": "Dynamic"},
			}})
		case "flat":
			ok(w, map[string]any{"leases": []map[string]any{
				{"address": "198.51.100.5", "hardwareAddress": "aa:aa:aa:aa:aa:02", "hostName": "b", "type": "Reserved"},
			}})
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "error", "errorMessage": "access denied"})
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestSyncTechnitiumLeases_Integration(t *testing.T) {
	pool := testdb.Connect(t, "services")
	testdb.Truncate(t, pool, "dhcp_leases", "dhcp_servers", "ip_addresses", "subnets", "vlans", "networks", "users", "configs")
	repo := repository.NewRepository(pool)
	svc := NewService(repo, testMFAKey)
	ctx := context.Background()

	u, err := repo.CreateUser(ctx, "dhcp-user", "dhcp@example.com")
	require.NoError(t, err)
	n, err := repo.CreateNetwork(ctx, "dhcp-net", "fixture", u.ID)
	require.NoError(t, err)
	vlan, err := repo.CreateVLAN(ctx, nil, nil, nil, 10, "users", "")
	require.NoError(t, err)
	// 192.0.2.0/24 is on a VLAN; 198.51.100.0/24 is not tracked at all.
	sub, err := repo.CreateSubnetWithVLAN(ctx, n.ID, "192.0.2.0", 24, "vlan10", nil, false, false, nil, nil, &vlan.ID)
	require.NoError(t, err)

	fake := fakeTechnitium(t)
	require.NoError(t, svc.Config.SetCtx(ctx, "technitium_url", fake.URL))
	require.NoError(t, svc.Config.SetCtx(ctx, "technitium_token", "tok"))

	count, err := svc.SyncTechnitiumLeases(ctx)
	// The "bad" scope fails but the others still sync, and the error is reported.
	assert.ErrorContains(t, err, "bad")
	assert.Equal(t, 2, count)

	server, err := repo.GetOrCreateTechnitiumDHCPServer(ctx, fake.URL)
	require.NoError(t, err)
	assert.True(t, server.AutoSync, "auto-sync defaults to on")
	require.NotNil(t, server.LastSyncedAt)
	assert.Contains(t, server.LastSyncError, "bad")
	require.NotNil(t, server.LastSyncCount)
	assert.Equal(t, 2, *server.LastSyncCount)

	leases, err := repo.ListDHCPLeasesFiltered(ctx, repository.DHCPLeaseFilter{ServerID: server.ID})
	require.NoError(t, err)
	require.Len(t, leases, 2)
	byIP := map[string]int{}
	for i, l := range leases {
		byIP[l.IPAddress] = i
	}

	// Lease inside a VLAN-tagged subnet picks up the subnet and its VLAN.
	l1 := leases[byIP["192.0.2.10"]]
	require.NotNil(t, l1.SubnetID)
	assert.Equal(t, sub.ID, *l1.SubnetID)
	require.NotNil(t, l1.VLANID)
	assert.Equal(t, vlan.ID, *l1.VLANID)
	require.NotNil(t, l1.VLANName)
	assert.Equal(t, "10 users", *l1.VLANName)
	assert.Equal(t, "active", l1.State)

	// Lease outside any known subnet still syncs, without a subnet or VLAN.
	l2 := leases[byIP["198.51.100.5"]]
	assert.Nil(t, l2.SubnetID)
	assert.Nil(t, l2.VLANID)
	assert.Equal(t, "reserved", l2.State)

	// Filters by subnet, VLAN, and a non-matching VLAN.
	got, err := repo.ListDHCPLeasesFiltered(ctx, repository.DHCPLeaseFilter{SubnetID: sub.ID})
	require.NoError(t, err)
	assert.Len(t, got, 1)
	got, err = repo.ListDHCPLeasesFiltered(ctx, repository.DHCPLeaseFilter{VLANID: vlan.ID})
	require.NoError(t, err)
	assert.Len(t, got, 1)
	got, err = repo.ListDHCPLeasesFiltered(ctx, repository.DHCPLeaseFilter{VLANID: vlan.ID + 1000})
	require.NoError(t, err)
	assert.Empty(t, got)

	// A server-level VLAN is the fallback for leases outside any VLAN subnet.
	off := false
	_, err = repo.UpdateDHCPServer(ctx, server.ID, &repository.DHCPServerParams{
		Name: server.Name, Address: server.Address, Vendor: server.Vendor, Status: server.Status,
		VLANID: &vlan.ID, AutoSync: &off,
	})
	require.NoError(t, err)

	// Scheduled sync honors the per-server opt-out; manual sync does not.
	_, skipped, err := svc.syncTechnitiumLeases(ctx, false)
	require.NoError(t, err)
	assert.True(t, skipped)

	_, _ = svc.SyncTechnitiumLeases(ctx)
	got, err = repo.ListDHCPLeasesFiltered(ctx, repository.DHCPLeaseFilter{VLANID: vlan.ID})
	require.NoError(t, err)
	assert.Len(t, got, 2, "server VLAN applies to the lease with no VLAN subnet")
}

func TestTechnitiumSyncInterval(t *testing.T) {
	pool := testdb.Connect(t, "services")
	testdb.Truncate(t, pool, "configs")
	svc := NewService(repository.NewRepository(pool), testMFAKey)
	ctx := context.Background()

	assert.Equal(t, 15*60*1e9, float64(svc.technitiumSyncInterval(ctx)), "default is 15 minutes")
	require.NoError(t, svc.Config.SetCtx(ctx, "technitium_dhcp_sync_interval_minutes", "0"))
	assert.Zero(t, svc.technitiumSyncInterval(ctx), "0 disables the scheduled sync")
	require.NoError(t, svc.Config.SetCtx(ctx, "technitium_dhcp_sync_interval_minutes", "5"))
	assert.Equal(t, 5*60*1e9, float64(svc.technitiumSyncInterval(ctx)))
	require.NoError(t, svc.Config.SetCtx(ctx, "technitium_dhcp_sync_interval_minutes", "junk"))
	assert.Equal(t, 15*60*1e9, float64(svc.technitiumSyncInterval(ctx)), "bad value falls back to default")
}
