package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"padduck/internal/testdb"
	"padduck/repository"
)

func testNetworkModulesService(t *testing.T) *NetworkModulesService {
	t.Helper()
	pool := testdb.Connect(t, "services")
	testdb.Truncate(t, pool,
		"nat_rules", "firewall_zone_mappings", "firewall_zones",
		"dhcp_leases", "dhcp_servers",
		"logical_circuits", "physical_circuits", "circuit_providers")
	return NewNetworkModulesService(repository.NewRepository(pool))
}

func TestDefaultString(t *testing.T) {
	assert.Equal(t, "fallback", defaultString("", "fallback"))
	assert.Equal(t, "fallback", defaultString("   ", "fallback"))
	assert.Equal(t, "value", defaultString("value", "fallback"))
}

func TestValidateCIDRLike(t *testing.T) {
	assert.NoError(t, validateCIDRLike("10.0.0.1", "field"))
	assert.NoError(t, validateCIDRLike("10.0.0.0/24", "field"))
	assert.NoError(t, validateCIDRLike("2001:db8::1", "field"))
	assert.ErrorContains(t, validateCIDRLike("not-an-ip", "internal CIDR"), "internal CIDR")
	assert.ErrorContains(t, validateCIDRLike("10.0.0.0/99", "field"), "must be an IP address or CIDR")
}

func TestNATRuleCRUD_Integration(t *testing.T) {
	nm := testNetworkModulesService(t)
	ctx := context.Background()

	// Validation failures.
	_, err := nm.CreateNATRule(ctx, &repository.NATRuleParams{InternalCIDR: "10.0.0.0/24", ExternalCIDR: "203.0.113.0/24"})
	assert.ErrorContains(t, err, "name is required")
	_, err = nm.CreateNATRule(ctx, &repository.NATRuleParams{Name: "bad", InternalCIDR: "nope", ExternalCIDR: "203.0.113.0/24"})
	assert.ErrorContains(t, err, "internal CIDR")
	_, err = nm.CreateNATRule(ctx, &repository.NATRuleParams{Name: "bad", InternalCIDR: "10.0.0.0/24", ExternalCIDR: "nope"})
	assert.ErrorContains(t, err, "external CIDR")

	// Create applies the documented defaults.
	rule, err := nm.CreateNATRule(ctx, &repository.NATRuleParams{
		Name: "office-snat", InternalCIDR: "10.0.0.0/24", ExternalCIDR: "203.0.113.10",
	})
	require.NoError(t, err)
	assert.Equal(t, "static", rule.Type)
	assert.Equal(t, "any", rule.Protocol)
	assert.Equal(t, "active", rule.Status)

	got, err := nm.GetNATRule(ctx, rule.ID)
	require.NoError(t, err)
	assert.Equal(t, "office-snat", got.Name)

	rules, err := nm.ListNATRules(ctx)
	require.NoError(t, err)
	assert.Len(t, rules, 1)

	updated, err := nm.UpdateNATRule(ctx, rule.ID, &repository.NATRuleParams{
		Name: "office-snat-v2", Type: "dynamic", InternalCIDR: "10.0.0.0/24", ExternalCIDR: "203.0.113.11",
	})
	require.NoError(t, err)
	assert.Equal(t, "office-snat-v2", updated.Name)
	assert.Equal(t, "dynamic", updated.Type)

	require.NoError(t, nm.DeleteNATRule(ctx, rule.ID))
	_, err = nm.GetNATRule(ctx, rule.ID)
	assert.Error(t, err)
}

func TestFirewallZonesAndMappings_Integration(t *testing.T) {
	nm := testNetworkModulesService(t)
	ctx := context.Background()

	_, err := nm.CreateFirewallZone(ctx, &repository.FirewallZoneParams{})
	assert.ErrorContains(t, err, "name is required")

	zone, err := nm.CreateFirewallZone(ctx, &repository.FirewallZoneParams{Name: "dmz"})
	require.NoError(t, err)
	assert.Equal(t, "#2563eb", zone.Color, "default color applied")
	assert.Equal(t, "active", zone.Status)

	// Mapping validation table.
	for _, tc := range []struct {
		name string
		req  repository.FirewallZoneMappingParams
		want string
	}{
		{"missing zone", repository.FirewallZoneMappingParams{ObjectType: "cidr", CIDR: "10.0.0.0/24"}, "zone is required"},
		{"missing object type", repository.FirewallZoneMappingParams{ZoneID: zone.ID}, "object type is required"},
		{"cidr type without cidr", repository.FirewallZoneMappingParams{ZoneID: zone.ID, ObjectType: "cidr"}, "CIDR is required"},
		{"no object id or cidr", repository.FirewallZoneMappingParams{ZoneID: zone.ID, ObjectType: "subnet"}, "object ID or CIDR is required"},
		{"invalid cidr", repository.FirewallZoneMappingParams{ZoneID: zone.ID, ObjectType: "cidr", CIDR: "10.0.0.0/99"}, "CIDR must be valid"},
	} {
		req := tc.req
		_, err := nm.CreateFirewallZoneMapping(ctx, &req)
		assert.ErrorContains(t, err, tc.want, tc.name)
	}

	mapping, err := nm.CreateFirewallZoneMapping(ctx, &repository.FirewallZoneMappingParams{
		ZoneID: zone.ID, ObjectType: "cidr", CIDR: "10.20.0.0/16",
	})
	require.NoError(t, err)
	assert.Equal(t, "both", mapping.Direction, "default direction applied")

	mappings, err := nm.ListFirewallZoneMappings(ctx, zone.ID)
	require.NoError(t, err)
	assert.Len(t, mappings, 1)

	updatedMapping, err := nm.UpdateFirewallZoneMapping(ctx, mapping.ID, &repository.FirewallZoneMappingParams{
		ZoneID: zone.ID, ObjectType: "cidr", CIDR: "10.30.0.0/16", Direction: "inbound",
	})
	require.NoError(t, err)
	assert.Equal(t, "inbound", updatedMapping.Direction)

	require.NoError(t, nm.DeleteFirewallZoneMapping(ctx, mapping.ID))

	updatedZone, err := nm.UpdateFirewallZone(ctx, zone.ID, &repository.FirewallZoneParams{Name: "dmz-renamed", Color: "#000000"})
	require.NoError(t, err)
	assert.Equal(t, "dmz-renamed", updatedZone.Name)

	require.NoError(t, nm.DeleteFirewallZone(ctx, zone.ID))
	_, err = nm.GetFirewallZone(ctx, zone.ID)
	assert.Error(t, err)
}

func TestDHCPServersAndLeases_Integration(t *testing.T) {
	nm := testNetworkModulesService(t)
	ctx := context.Background()

	_, err := nm.CreateDHCPServer(ctx, &repository.DHCPServerParams{Address: "10.0.0.2"})
	assert.ErrorContains(t, err, "name is required")
	_, err = nm.CreateDHCPServer(ctx, &repository.DHCPServerParams{Name: "dhcp-01", Address: "not-an-ip"})
	assert.ErrorContains(t, err, "must be an IP address")

	server, err := nm.CreateDHCPServer(ctx, &repository.DHCPServerParams{Name: "dhcp-01", Address: "10.0.0.2"})
	require.NoError(t, err)
	assert.Equal(t, "active", server.Status)

	// Lease validation.
	_, err = nm.CreateDHCPLease(ctx, &repository.DHCPLeaseParams{IPAddress: "10.0.0.50", MACAddress: "aa:bb:cc:dd:ee:ff"})
	assert.ErrorContains(t, err, "server is required")
	_, err = nm.CreateDHCPLease(ctx, &repository.DHCPLeaseParams{ServerID: server.ID, IPAddress: "bad", MACAddress: "aa:bb:cc:dd:ee:ff"})
	assert.ErrorContains(t, err, "IP address must be valid")
	_, err = nm.CreateDHCPLease(ctx, &repository.DHCPLeaseParams{ServerID: server.ID, IPAddress: "10.0.0.50"})
	assert.ErrorContains(t, err, "MAC address is required")

	lease, err := nm.CreateDHCPLease(ctx, &repository.DHCPLeaseParams{
		ServerID: server.ID, IPAddress: "10.0.0.50", MACAddress: "aa:bb:cc:dd:ee:ff", Hostname: "printer",
	})
	require.NoError(t, err)
	assert.Equal(t, "active", lease.State)

	leases, err := nm.ListDHCPLeases(ctx, server.ID)
	require.NoError(t, err)
	assert.Len(t, leases, 1)

	updatedLease, err := nm.UpdateDHCPLease(ctx, lease.ID, &repository.DHCPLeaseParams{
		ServerID: server.ID, IPAddress: "10.0.0.51", MACAddress: "aa:bb:cc:dd:ee:ff", State: "expired",
	})
	require.NoError(t, err)
	assert.Equal(t, "expired", updatedLease.State)

	require.NoError(t, nm.DeleteDHCPLease(ctx, lease.ID))

	updatedServer, err := nm.UpdateDHCPServer(ctx, server.ID, &repository.DHCPServerParams{Name: "dhcp-01b", Address: "10.0.0.3"})
	require.NoError(t, err)
	assert.Equal(t, "dhcp-01b", updatedServer.Name)

	require.NoError(t, nm.DeleteDHCPServer(ctx, server.ID))
	servers, err := nm.ListDHCPServers(ctx)
	require.NoError(t, err)
	assert.Empty(t, servers)
}

func TestCircuits_Integration(t *testing.T) {
	nm := testNetworkModulesService(t)
	ctx := context.Background()

	provider, err := nm.CreateCircuitProvider(ctx, &repository.CircuitProviderParams{Name: "FiberCo", AccountNo: "AC-1"})
	require.NoError(t, err)

	physical, err := nm.CreatePhysicalCircuit(ctx, &repository.PhysicalCircuitParams{
		ProviderID: provider.ID, CircuitID: "FC-100", Name: "uplink-a", Type: "fiber", Status: "active",
	})
	require.NoError(t, err)

	logical, err := nm.CreateLogicalCircuit(ctx, &repository.LogicalCircuitParams{
		PhysicalCircuitID: &physical.ID, Name: "transit-vlan", ServiceID: "SVC-9", Type: "transit", Status: "active",
	})
	require.NoError(t, err)

	physicals, err := nm.ListPhysicalCircuits(ctx)
	require.NoError(t, err)
	assert.Len(t, physicals, 1)
	logicals, err := nm.ListLogicalCircuits(ctx)
	require.NoError(t, err)
	assert.Len(t, logicals, 1)

	updatedPhysical, err := nm.UpdatePhysicalCircuit(ctx, physical.ID, &repository.PhysicalCircuitParams{
		ProviderID: provider.ID, CircuitID: "FC-100", Name: "uplink-a1", Type: "fiber", Status: "retired",
	})
	require.NoError(t, err)
	assert.Equal(t, "uplink-a1", updatedPhysical.Name)

	updatedLogical, err := nm.UpdateLogicalCircuit(ctx, logical.ID, &repository.LogicalCircuitParams{
		PhysicalCircuitID: &physical.ID, Name: "transit-vlan-2", ServiceID: "SVC-9", Type: "transit", Status: "active",
	})
	require.NoError(t, err)
	assert.Equal(t, "transit-vlan-2", updatedLogical.Name)

	updatedProvider, err := nm.UpdateCircuitProvider(ctx, provider.ID, &repository.CircuitProviderParams{Name: "FiberCo Inc"})
	require.NoError(t, err)
	assert.Equal(t, "FiberCo Inc", updatedProvider.Name)

	require.NoError(t, nm.DeleteLogicalCircuit(ctx, logical.ID))
	require.NoError(t, nm.DeletePhysicalCircuit(ctx, physical.ID))
	require.NoError(t, nm.DeleteCircuitProvider(ctx, provider.ID))

	providers, err := nm.ListCircuitProviders(ctx)
	require.NoError(t, err)
	assert.Empty(t, providers)
}
