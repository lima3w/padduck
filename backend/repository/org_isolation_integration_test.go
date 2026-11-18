package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestOrg(t *testing.T, r *Repository, label string) int64 {
	t.Helper()
	slug := fmt.Sprintf("%s-%d", label, time.Now().UnixNano())
	org, err := r.CreateOrganization(context.Background(), "Org "+slug, slug)
	require.NoError(t, err)
	return org.ID
}

// isMissing reports whether err means "no such row" (what a cross-org lookup returns).
func isMissing(err error) bool {
	return err != nil && (errors.Is(err, pgx.ErrNoRows) || strings.Contains(err.Error(), "not found"))
}

// TestOrgIsolation_Integration proves that one organization cannot read, change
// or delete another's rows, cannot point its own rows at them, and can reuse
// names and VLAN numbers the other organization already has.
func TestOrgIsolation_Integration(t *testing.T) {
	r := testRepo(t)
	base := context.Background()
	orgA, orgB := newTestOrg(t, r, "iso-a"), newTestOrg(t, r, "iso-b")
	a, b := WithOrg(base, orgA), WithOrg(base, orgB)
	uid := createTestUser(t, r)

	// --- Organization A builds an inventory.
	netA, err := r.CreateNetwork(a, "net-a", "", uid)
	require.NoError(t, err)
	subA, err := r.CreateSubnet(a, netA.ID, "10.60.0.0", 24, "sub-a", nil, false, false)
	require.NoError(t, err)
	ipA, err := r.CreateIPAddress(a, subA.ID, "10.60.0.5", "host-a", "assigned", nil, nil, nil, nil)
	require.NoError(t, err)
	vrfA, err := r.CreateVRF(a, "vrf-shared-name", "65000:1", "")
	require.NoError(t, err)
	vlanA, err := r.CreateVLAN(a, nil, nil, nil, 100, "vlan-a", "")
	require.NoError(t, err)
	locA, err := r.CreateLocation(a, &LocationParams{Name: "loc-a", Type: "site", Status: "active"})
	require.NoError(t, err)
	rackA, err := r.CreateRack(a, &RackParams{LocationID: &locA.ID, Name: "rack-a", SizeU: 42})
	require.NoError(t, err)
	custA, err := r.CreateCustomer(a, "cust-shared-name", "", "", "", "")
	require.NoError(t, err)
	nsA, err := r.CreateNameserver(a, &NameserverParams{Name: "ns-a", Server1: "10.0.0.53"})
	require.NoError(t, err)
	devA, err := r.CreateDevice(a, &DeviceParams{Hostname: "dev-a"})
	require.NoError(t, err)
	tagA, err := r.CreateIPTag(a, "tag-shared-name", "#ff0000", nil)
	require.NoError(t, err)

	// --- Organization B sees none of it.
	t.Run("reads are invisible across orgs", func(t *testing.T) {
		_, err := r.GetNetworkByID(b, netA.ID)
		assert.True(t, isMissing(err), "network: %v", err)
		_, err = r.GetSubnetByID(b, subA.ID)
		assert.True(t, isMissing(err), "subnet: %v", err)
		_, err = r.GetIPAddressByID(b, ipA.ID)
		assert.True(t, isMissing(err), "ip: %v", err)
		_, err = r.GetVRFByID(b, vrfA.ID)
		assert.True(t, isMissing(err), "vrf: %v", err)
		_, err = r.GetVLANByID(b, vlanA.ID)
		assert.True(t, isMissing(err), "vlan: %v", err)
		_, err = r.GetLocationByID(b, locA.ID)
		assert.True(t, isMissing(err), "location: %v", err)
		_, err = r.GetRackByID(b, rackA.ID)
		assert.True(t, isMissing(err), "rack: %v", err)
		_, err = r.GetCustomerByID(b, custA.ID)
		assert.True(t, isMissing(err), "customer: %v", err)
		_, err = r.GetNameserverByID(b, nsA.ID)
		assert.True(t, isMissing(err), "nameserver: %v", err)
		_, err = r.GetDeviceByID(b, devA.ID)
		assert.True(t, isMissing(err), "device: %v", err)
		_, err = r.GetIPTagByID(b, tagA.ID)
		assert.True(t, isMissing(err), "tag: %v", err)
		_, err = r.GetSubnetByCIDR(b, "10.60.0.0/24")
		assert.True(t, isMissing(err), "subnet by cidr: %v", err)
		_, err = r.FindSubnetForIP(b, "10.60.0.5")
		assert.True(t, isMissing(err), "subnet for ip: %v", err)
	})

	t.Run("lists exclude other orgs", func(t *testing.T) {
		nets, err := r.ListAllNetworks(b)
		require.NoError(t, err)
		assert.Empty(t, nets)
		subs, err := r.ListSubnetsBySection(b, netA.ID)
		require.NoError(t, err)
		assert.Empty(t, subs)
		ips, err := r.ListAllIPAddresses(b)
		require.NoError(t, err)
		assert.Empty(t, ips)
		ipsBySubnet, err := r.ListIPAddressesBySubnet(b, subA.ID)
		require.NoError(t, err)
		assert.Empty(t, ipsBySubnet)
		vlans, err := r.ListAllVLANs(b)
		require.NoError(t, err)
		assert.Empty(t, vlans)
		locs, err := r.ListLocations(b)
		require.NoError(t, err)
		assert.Empty(t, locs)
		racks, err := r.ListRacks(b, nil)
		require.NoError(t, err)
		assert.Empty(t, racks)
		nss, err := r.ListNameservers(b)
		require.NoError(t, err)
		assert.Empty(t, nss)
		devs, total, err := r.ListDevicesWithOptions(b, ListOptions{Limit: 50})
		require.NoError(t, err)
		assert.Empty(t, devs)
		assert.Zero(t, total)
		custs, total, err := r.ListCustomersPaginated(b, 50, 0)
		require.NoError(t, err)
		assert.Empty(t, custs)
		assert.Zero(t, total)
		vrfs, total, err := r.ListVRFsPaginated(b, 50, 0)
		require.NoError(t, err)
		assert.Empty(t, vrfs)
		assert.Zero(t, total)

		tags, err := r.ListIPTags(b)
		require.NoError(t, err)
		for _, tag := range tags {
			assert.NotEqual(t, tagA.ID, tag.ID, "org A's tag leaked into org B's list")
		}
		// Built-in tags stay visible to everyone.
		assert.NotEmpty(t, tags)

		found, err := r.GlobalSearchSubnets(b, "10.60", 10)
		require.NoError(t, err)
		assert.Empty(t, found)
		foundIPs, err := r.SearchIPAddressesGlobal(b, "10.60", 10)
		require.NoError(t, err)
		assert.Empty(t, foundIPs)
	})

	t.Run("dashboard counts are per org", func(t *testing.T) {
		sumB, err := r.GetDashboardSummary(b, &orgB)
		require.NoError(t, err)
		assert.Zero(t, sumB.TotalSubnets)
		assert.Zero(t, sumB.TotalNetworks)
		sumA, err := r.GetDashboardSummary(a, &orgA)
		require.NoError(t, err)
		assert.Equal(t, int64(1), int64(sumA.TotalSubnets))
	})

	t.Run("writes cannot reach other orgs", func(t *testing.T) {
		_, err := r.UpdateNetwork(b, netA.ID, "hijacked", "")
		assert.True(t, isMissing(err), "update network: %v", err)
		_, err = r.UpdateCustomer(b, custA.ID, "hijacked", "", "", "", "")
		assert.True(t, isMissing(err), "update customer: %v", err)
		_, err = r.UpdateVRF(b, vrfA.ID, "hijacked", "", "")
		assert.True(t, isMissing(err), "update vrf: %v", err)

		// Deletes are silent no-ops across orgs; the row must survive.
		_ = r.DeleteNetwork(b, netA.ID)
		_ = r.DeleteSubnet(b, subA.ID)
		_ = r.DeleteVLAN(b, vlanA.ID)
		_ = r.DeleteDevice(b, devA.ID)
		_ = r.DeleteCustomer(b, custA.ID)
		_ = r.DeleteVRF(b, vrfA.ID)
		_ = r.DeleteLocation(b, locA.ID)
		_ = r.DeleteRack(b, rackA.ID)
		_ = r.DeleteNameserver(b, nsA.ID)
		deleted, err := r.BulkDeleteIPAddresses(b, []int64{ipA.ID})
		require.NoError(t, err)
		assert.Empty(t, deleted)

		for name, get := range map[string]func() error{
			"network":    func() error { _, e := r.GetNetworkByID(a, netA.ID); return e },
			"subnet":     func() error { _, e := r.GetSubnetByID(a, subA.ID); return e },
			"ip":         func() error { _, e := r.GetIPAddressByID(a, ipA.ID); return e },
			"vlan":       func() error { _, e := r.GetVLANByID(a, vlanA.ID); return e },
			"device":     func() error { _, e := r.GetDeviceByID(a, devA.ID); return e },
			"customer":   func() error { _, e := r.GetCustomerByID(a, custA.ID); return e },
			"vrf":        func() error { _, e := r.GetVRFByID(a, vrfA.ID); return e },
			"location":   func() error { _, e := r.GetLocationByID(a, locA.ID); return e },
			"rack":       func() error { _, e := r.GetRackByID(a, rackA.ID); return e },
			"nameserver": func() error { _, e := r.GetNameserverByID(a, nsA.ID); return e },
		} {
			assert.NoError(t, get(), "%s must survive a cross-org delete", name)
		}

		// Org A's own data is unchanged.
		n, err := r.GetNetworkByID(a, netA.ID)
		require.NoError(t, err)
		assert.Equal(t, "net-a", n.Name)
	})

	t.Run("rows cannot point at other orgs", func(t *testing.T) {
		_, err := r.CreateSubnet(b, netA.ID, "10.61.0.0", 24, "x", nil, false, false)
		assert.Error(t, err, "subnet in another org's network")
		_, err = r.CreateIPAddress(b, subA.ID, "10.60.0.9", "x", "available", nil, nil, nil, nil)
		assert.Error(t, err, "IP in another org's subnet")
		_, err = r.CreateRack(b, &RackParams{LocationID: &locA.ID, Name: "x", SizeU: 42})
		assert.Error(t, err, "rack in another org's location")
		_, err = r.CreateDevice(b, &DeviceParams{Hostname: "x", LocationID: &locA.ID})
		assert.Error(t, err, "device in another org's location")
		_, err = r.CreateDevice(b, &DeviceParams{Hostname: "x2", RackID: &rackA.ID})
		assert.Error(t, err, "device in another org's rack")
		_, err = r.CreateVLAN(b, &vrfA.ID, nil, nil, 200, "x", "")
		assert.Error(t, err, "VLAN in another org's VRF")
		_, err = r.CreateIPAddress(b, subA.ID, "10.60.0.10", "x", "available", &tagA.ID, nil, nil, nil)
		assert.Error(t, err)

		netB, err := r.CreateNetwork(b, "net-b", "", uid)
		require.NoError(t, err)
		subB, err := r.CreateSubnet(b, netB.ID, "10.62.0.0", 24, "sub-b", &nsA.Server1, false, false)
		require.NoError(t, err)
		_, err = r.UpdateSubnetWithLocation(b, subB.ID, "d", nil, false, false, &locA.ID)
		assert.Error(t, err, "subnet pointed at another org's location")
		_, err = r.UpdateSubnetWithVLAN(b, subB.ID, "d", nil, false, false, nil, &nsA.ID, nil)
		assert.Error(t, err, "subnet pointed at another org's nameserver")
		_, err = r.AssignSubnetToVLAN(b, subB.ID, &vlanA.ID)
		assert.Error(t, err, "subnet pointed at another org's VLAN")
	})

	t.Run("names and VLAN numbers are unique per org, not globally", func(t *testing.T) {
		_, err := r.CreateVRF(b, "vrf-shared-name", "65000:2", "")
		assert.NoError(t, err)
		_, err = r.CreateCustomer(b, "cust-shared-name", "", "", "", "")
		assert.NoError(t, err)
		_, err = r.CreateVLAN(b, nil, nil, nil, 100, "vlan-b", "")
		assert.NoError(t, err)
		_, err = r.CreateIPTag(b, "tag-shared-name", "#00ff00", nil)
		assert.NoError(t, err)

		// ...but still unique inside one org.
		_, err = r.CreateVRF(a, "vrf-shared-name", "65000:3", "")
		assert.Error(t, err)
		_, err = r.CreateVLAN(a, nil, nil, nil, 100, "again", "")
		assert.Error(t, err)
	})

	t.Run("system context sees every org", func(t *testing.T) {
		nets, err := r.ListAllNetworks(base)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(nets), 2)
		_, err = r.GetSubnetByID(base, subA.ID)
		assert.NoError(t, err)
	})
}

func TestExpandOrg(t *testing.T) {
	q, args := expandOrg(WithOrg(context.Background(), 7),
		"SELECT 1 FROM t WHERE id = $1 AND @org(t.organization_id) AND @org(t.organization_id)", []any{42})
	assert.Equal(t,
		"SELECT 1 FROM t WHERE id = $1 AND (t.organization_id = $2::bigint OR $2::bigint IS NULL) AND (t.organization_id = $3::bigint OR $3::bigint IS NULL)", q)
	assert.Equal(t, []any{42, int64(7), int64(7)}, args)

	q, args = expandOrg(context.Background(), "INSERT INTO t (a, organization_id) VALUES ($1, @orgval)", []any{"x"})
	assert.Equal(t, "INSERT INTO t (a, organization_id) VALUES ($1, $2)", q)
	assert.Equal(t, []any{"x", int64(1)}, args, "system inserts land in the default org")

	_, args = expandOrg(context.Background(), "SELECT 1 WHERE @org(c)", nil)
	assert.Equal(t, []any{nil}, args, "system reads are unscoped")
}
