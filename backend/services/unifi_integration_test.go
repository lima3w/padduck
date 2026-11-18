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

func fakeUniFi(t *testing.T, version string) *httptest.Server {
	t.Helper()
	env := func(w http.ResponseWriter, data any) {
		_ = json.NewEncoder(w).Encode(map[string]any{"meta": map[string]string{"rc": "ok"}, "data": data})
	}
	mux := http.NewServeMux()
	guard := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-API-KEY") != "key" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			h(w, r)
		}
	}
	mux.HandleFunc("/proxy/network/api/self/sites", guard(func(w http.ResponseWriter, _ *http.Request) {
		env(w, []map[string]any{{"name": "default"}})
	}))
	mux.HandleFunc("/proxy/network/api/s/default/rest/networkconf", guard(func(w http.ResponseWriter, _ *http.Request) {
		env(w, []map[string]any{
			{"_id": "n10", "name": "Users", "vlan_enabled": true, "vlan": "10"},
			{"_id": "n20", "name": "UniFi IoT", "vlan_enabled": true, "vlan": "20"},
			{"_id": "n1", "name": "Default"},
		})
	}))
	mux.HandleFunc("/proxy/network/api/s/default/stat/sta", guard(func(w http.ResponseWriter, _ *http.Request) {
		env(w, []map[string]any{
			{"mac": "aa:bb:cc:00:00:05", "ip": "10.0.10.5", "hostname": "new-client"},
			{"mac": "aa:bb:cc:00:00:06", "ip": "10.0.10.6", "hostname": "blank-client"},
			{"mac": "aa:bb:cc:00:00:07", "ip": "10.0.10.7", "hostname": "laptop"},
			{"mac": "AA-BB-CC-00-00-08", "ip": "10.0.10.8", "hostname": "same-mac"},
			{"mac": "aa:bb:cc:00:00:99", "ip": "172.31.99.9", "hostname": "elsewhere"},
		})
	}))
	mux.HandleFunc("/proxy/network/api/s/default/stat/device", guard(func(w http.ResponseWriter, _ *http.Request) {
		env(w, []map[string]any{
			{"mac": "11:22:33:44:55:01", "name": "SW-01", "model": "USW-24", "version": version, "type": "usw", "state": 1, "adopted": true},
			{"mac": "11:22:33:44:55:02", "name": "ap-new", "model": "U7-Pro", "version": "6.6.0", "type": "uap", "state": 0, "adopted": true,
				"uplink": map[string]any{"uplink_mac": "11:22:33:44:55:01", "uplink_remote_port": 7}},
			{"mac": "11:22:33:44:55:03", "name": "pending", "adopted": false},
		})
	}))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestSyncUniFi_Integration(t *testing.T) {
	pool := testdb.Connect(t, "services")
	testdb.Truncate(t, pool, "topology_edges", "topology_nodes", "unifi_conflicts", "unifi_links", "ip_addresses", "devices", "subnets", "vlans", "networks", "users", "configs")
	repo := repository.NewRepository(pool)
	svc := NewService(repo, testMFAKey)
	ctx := context.Background()

	u, err := repo.CreateUser(ctx, "unifi-user", "unifi@example.com")
	require.NoError(t, err)
	n, err := repo.CreateNetwork(ctx, "net", "", u.ID)
	require.NoError(t, err)
	sub, err := repo.CreateSubnet(ctx, n.ID, "10.0.10.0", 24, "users", nil, false, false)
	require.NoError(t, err)

	// VLAN 20 already exists under a different name and must keep it.
	existing, err := repo.CreateVLAN(ctx, nil, nil, nil, 20, "iot", "")
	require.NoError(t, err)

	mac := func(s string) *string { return &s }
	_, err = repo.CreateIPAddress(ctx, sub.ID, "10.0.10.6", "", "available", nil, nil, nil, nil)
	require.NoError(t, err)
	conflicted, err := repo.CreateIPAddress(ctx, sub.ID, "10.0.10.7", "printer", "assigned", nil, mac("de:ad:be:ef:00:07"), nil, nil)
	require.NoError(t, err)
	_, err = repo.CreateIPAddress(ctx, sub.ID, "10.0.10.8", "same-mac", "assigned", nil, mac("aa:bb:cc:00:00:08"), nil, nil)
	require.NoError(t, err)

	// An existing device matched by hostname (case-insensitively).
	oldFW := "6.0.0"
	dev, err := repo.CreateDevice(ctx, &repository.DeviceParams{Hostname: "sw-01", OSVersion: &oldFW, RackUnitSize: 1})
	require.NoError(t, err)

	fake := fakeUniFi(t, "7.1.0")
	for k, v := range map[string]string{"unifi_url": fake.URL, "unifi_api_key": "key"} {
		require.NoError(t, svc.Config.SetCtx(ctx, k, v))
	}

	res, err := svc.SyncUniFi(ctx)
	require.NoError(t, err)
	assert.Empty(t, res.Errors)

	// VLANs: one created, one adopted, untagged network ignored.
	assert.Equal(t, 1, res.VLANsCreated)
	assert.Equal(t, 2, res.VLANsLinked)
	v10, err := repo.FindVLANByNumber(ctx, 10)
	require.NoError(t, err)
	assert.Equal(t, "Users", v10.Name)
	v20, err := repo.GetVLANByID(ctx, existing.ID)
	require.NoError(t, err)
	assert.Equal(t, "iot", v20.Name, "existing VLAN name is never overwritten")

	links, err := repo.ListUniFiLinks(ctx)
	require.NoError(t, err)
	vlanNames := map[int64]string{}
	for _, l := range links {
		if l.VLANID != nil {
			vlanNames[*l.VLANID] = l.UniFiName
		}
	}
	assert.Equal(t, "UniFi IoT", vlanNames[existing.ID])
	assert.Equal(t, "Users", vlanNames[v10.ID])

	// Clients.
	assert.Equal(t, 1, res.ClientsCreated)
	assert.Equal(t, 1, res.ClientsUnmatched)
	created, err := repo.FindIPInSubnet(ctx, sub.ID, "10.0.10.5")
	require.NoError(t, err)
	assert.Equal(t, "new-client", created.Hostname)
	assert.Equal(t, "assigned", created.Status)
	require.NotNil(t, created.MACAddress)
	assert.Equal(t, "aa:bb:cc:00:00:05", *created.MACAddress)

	filled, err := repo.FindIPInSubnet(ctx, sub.ID, "10.0.10.6")
	require.NoError(t, err)
	assert.Equal(t, "blank-client", filled.Hostname, "blank hostname is filled in")
	assert.Equal(t, "assigned", filled.Status, "an in-use address is no longer available")
	require.NotNil(t, filled.MACAddress)

	kept, err := repo.FindIPInSubnet(ctx, sub.ID, "10.0.10.7")
	require.NoError(t, err)
	assert.Equal(t, "printer", kept.Hostname, "differing hostname is not overwritten")
	assert.Equal(t, "de:ad:be:ef:00:07", *kept.MACAddress, "differing MAC is not overwritten")

	// Two conflicts raised for 10.0.10.7; the same-MAC (dash notation) client raises none.
	assert.Equal(t, 2, res.ConflictsRaised)
	open, err := repo.ListUniFiConflicts(ctx, "open")
	require.NoError(t, err)
	require.Len(t, open, 2)
	fields := map[string]string{}
	for _, c := range open {
		assert.Equal(t, conflicted.ID, c.IPID)
		fields[c.Field] = c.UniFiValue
	}
	assert.Equal(t, "laptop", fields["hostname"])
	assert.Equal(t, "aa:bb:cc:00:00:07", fields["mac_address"])

	// Devices: one linked by hostname, one created, the un-adopted one skipped.
	assert.Equal(t, 1, res.DevicesLinked)
	assert.Equal(t, 1, res.DevicesCreated)
	linkedDev, err := repo.GetDeviceByID(ctx, dev.ID)
	require.NoError(t, err)
	require.NotNil(t, linkedDev.OSVersion)
	assert.Equal(t, "7.1.0", *linkedDev.OSVersion)
	assert.True(t, linkedDev.IsOnline)
	devs, err := repo.ListAllDevices(ctx)
	require.NoError(t, err)
	assert.Len(t, devs, 2)

	// Status is recorded.
	st, err := svc.UniFiStatus(ctx)
	require.NoError(t, err)
	assert.True(t, st.Configured)
	assert.NotNil(t, st.LastSyncedAt)
	assert.Empty(t, st.LastError)
	assert.Equal(t, 2, st.OpenConflict)
	require.NotNil(t, st.LastResult)
	assert.Equal(t, 1, st.LastResult.ClientsCreated)

	// Re-running is idempotent: no duplicates, no new conflicts, firmware updates.
	fake2 := fakeUniFi(t, "7.2.0")
	require.NoError(t, svc.Config.SetCtx(ctx, "unifi_url", fake2.URL))
	res2, err := svc.SyncUniFi(ctx)
	require.NoError(t, err)
	assert.Zero(t, res2.ClientsCreated)
	assert.Zero(t, res2.DevicesCreated)
	assert.Zero(t, res2.VLANsCreated)
	assert.Zero(t, res2.ConflictsRaised)
	assert.Equal(t, 2, res2.DevicesUpdated)
	linkedDev, err = repo.GetDeviceByID(ctx, dev.ID)
	require.NoError(t, err)
	assert.Equal(t, "7.2.0", *linkedDev.OSVersion)
	devs, err = repo.ListAllDevices(ctx)
	require.NoError(t, err)
	assert.Len(t, devs, 2)

	// Accept applies UniFi's hostname; dismiss keeps Padduck's MAC and stays quiet.
	for _, c := range open {
		action := "accept"
		if c.Field == "mac_address" {
			action = "dismiss"
		}
		_, err := svc.ResolveUniFiConflict(ctx, c.ID, action)
		require.NoError(t, err)
	}
	kept, err = repo.FindIPInSubnet(ctx, sub.ID, "10.0.10.7")
	require.NoError(t, err)
	assert.Equal(t, "laptop", kept.Hostname)
	assert.Equal(t, "de:ad:be:ef:00:07", *kept.MACAddress)

	_, err = svc.ResolveUniFiConflict(ctx, open[0].ID, "accept")
	assert.ErrorContains(t, err, "already")
	_, err = svc.ResolveUniFiConflict(ctx, open[0].ID, "bogus")
	assert.ErrorContains(t, err, "accept or dismiss")

	res3, err := svc.SyncUniFi(ctx)
	require.NoError(t, err)
	assert.Zero(t, res3.ConflictsRaised, "a dismissed difference is not raised again")
	open, err = repo.ListUniFiConflicts(ctx, "open")
	require.NoError(t, err)
	assert.Empty(t, open)

	// The AP's reported uplink becomes a suggested edge, and an admin decision survives re-sync.
	edges, err := svc.Ops.Topology.ListEdges(ctx, "suggested")
	require.NoError(t, err)
	require.Len(t, edges, 1)
	assert.Equal(t, "uplink_to", edges[0].Relationship)
	assert.Equal(t, "ap-new", edges[0].SourceLabel)
	assert.Equal(t, "sw-01", edges[0].TargetLabel)
	assert.Contains(t, *edges[0].Evidence, "port 7")
	require.NoError(t, svc.Ops.Topology.SetEdgeStatus(ctx, edges[0].ID, "dismissed"))
	_, err = svc.SyncUniFi(ctx)
	require.NoError(t, err)
	edges, err = svc.Ops.Topology.ListEdges(ctx, "suggested")
	require.NoError(t, err)
	assert.Empty(t, edges, "a dismissed suggestion is not recreated")
}

func TestSyncUniFi_NotConfiguredAndBadCredentials(t *testing.T) {
	pool := testdb.Connect(t, "services")
	testdb.Truncate(t, pool, "configs")
	svc := NewService(repository.NewRepository(pool), testMFAKey)
	ctx := context.Background()

	_, err := svc.SyncUniFi(ctx)
	assert.ErrorContains(t, err, "not configured")

	fake := fakeUniFi(t, "1")
	require.NoError(t, svc.Config.SetCtx(ctx, "unifi_url", fake.URL))
	require.NoError(t, svc.Config.SetCtx(ctx, "unifi_api_key", "wrong"))
	_, err = svc.SyncUniFi(ctx)
	assert.ErrorContains(t, err, "rejected the credentials")
	st, err := svc.UniFiStatus(ctx)
	require.NoError(t, err)
	assert.Contains(t, st.LastError, "rejected the credentials")

	assert.NoError(t, svc.TestUniFiConnection(ctx, &UniFiConfig{URL: fake.URL, APIKey: "key"}))
	assert.ErrorContains(t, svc.TestUniFiConnection(ctx, &UniFiConfig{URL: fake.URL}), "required")
}

func TestNormalizeMAC(t *testing.T) {
	assert.Equal(t, "aa:bb:cc:dd:ee:ff", normalizeMAC("AA-BB-CC-DD-EE-FF"))
	assert.Equal(t, "aa:bb:cc:dd:ee:ff", normalizeMAC(" aa:bb:cc:dd:ee:ff "))
	assert.Empty(t, normalizeMAC("nonsense"))
}
