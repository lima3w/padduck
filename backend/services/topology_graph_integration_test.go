package services

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"padduck/internal/testdb"
	"padduck/repository"
)

func TestTopologyGraphSync_Integration(t *testing.T) {
	pool := testdb.Connect(t, "services")
	testdb.Truncate(t, pool, "topology_edges", "topology_nodes", "ip_addresses", "devices", "subnets", "vlans", "locations", "networks", "users")
	repo := repository.NewRepository(pool)
	svc := NewService(repo, testMFAKey)
	topo := svc.Ops.Topology
	ctx := context.Background()

	u, err := repo.CreateUser(ctx, "topo-user", "topo@example.com")
	require.NoError(t, err)
	n, err := repo.CreateNetwork(ctx, "net", "", u.ID)
	require.NoError(t, err)
	parent, err := repo.CreateSubnet(ctx, n.ID, "10.1.0.0", 16, "campus", nil, false, false)
	require.NoError(t, err)
	child, err := repo.CreateSubnet(ctx, n.ID, "10.1.1.0", 24, "users", nil, false, false)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `UPDATE subnets SET parent_subnet_id=$1 WHERE id=$2`, parent.ID, child.ID)
	require.NoError(t, err)
	vlan, err := repo.CreateVLAN(ctx, nil, nil, nil, 10, "Users", "")
	require.NoError(t, err)
	_, err = repo.AssignSubnetToVLAN(ctx, child.ID, &vlan.ID)
	require.NoError(t, err)
	dev, err := repo.CreateDevice(ctx, &repository.DeviceParams{Hostname: "core-sw", RackUnitSize: 1})
	require.NoError(t, err)
	other, err := repo.CreateDevice(ctx, &repository.DeviceParams{Hostname: "edge-fw", RackUnitSize: 1})
	require.NoError(t, err)
	ip, err := repo.CreateIPAddress(ctx, child.ID, "10.1.1.2", "core-sw", "assigned", nil, nil, nil, nil)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `UPDATE ip_addresses SET device_id=$1 WHERE id=$2`, dev.ID, ip.ID)
	require.NoError(t, err)

	require.NoError(t, topo.SyncGraph(ctx))

	g, err := topo.GetFullGraph(ctx, 100)
	require.NoError(t, err)
	assert.Len(t, g.Nodes, 5, "two subnets, one vlan, two devices")
	rels := map[string]bool{}
	for _, e := range g.Edges {
		rels[e.Data["source"].(string)+" "+e.Data["relationship"].(string)+" "+e.Data["target"].(string)] = true
	}
	assert.True(t, rels["subnet_"+itoa(parent.ID)+" contains subnet_"+itoa(child.ID)])
	assert.True(t, rels["subnet_"+itoa(child.ID)+" member_of vlan_"+itoa(vlan.ID)])
	assert.True(t, rels["device_"+itoa(dev.ID)+" member_of subnet_"+itoa(child.ID)])

	// Labels are attached for the UI.
	byID := map[string]string{}
	for _, nd := range g.Nodes {
		byID[nd.Data["id"].(string)], _ = nd.Data["label"].(string)
	}
	assert.Equal(t, "10.1.1.0/24", byID["subnet_"+itoa(child.ID)])
	assert.Equal(t, "core-sw", byID["device_"+itoa(dev.ID)])
	assert.Equal(t, "VLAN 10", byID["vlan_"+itoa(vlan.ID)])

	// Traversal and path-finding work over the synced graph.
	nb, err := topo.GetNeighbors(ctx, nil, "device", dev.ID, 2)
	require.NoError(t, err)
	assert.Len(t, nb.Nodes, 4, "device, its subnet, the parent subnet and the vlan; not the unconnected device")
	path, err := topo.GetPath(ctx, "device", dev.ID, "vlan", vlan.ID)
	require.NoError(t, err)
	assert.Len(t, path.Nodes, 3)

	// A manually confirmed edge survives re-sync; a suggested one can be confirmed or dismissed.
	require.NoError(t, repo.SuggestTopologyEdge(ctx, "device", dev.ID, "device", other.ID, "uplink_to", 0.9, "UniFi uplink", "UniFi uplink"))
	sugg, err := topo.ListEdges(ctx, "suggested")
	require.NoError(t, err)
	require.Len(t, sugg, 1)
	require.NoError(t, topo.SyncGraph(ctx))
	sugg, err = topo.ListEdges(ctx, "suggested")
	require.NoError(t, err)
	require.Len(t, sugg, 1, "sync leaves suggestions alone")

	require.NoError(t, topo.SetEdgeStatus(ctx, sugg[0].ID, "confirmed"))
	nb, err = topo.GetNeighbors(ctx, nil, "device", other.ID, 1)
	require.NoError(t, err)
	assert.Len(t, nb.Nodes, 2)
	require.NoError(t, topo.SetEdgeStatus(ctx, sugg[0].ID, "dismissed"))
	nb, err = topo.GetNeighbors(ctx, nil, "device", other.ID, 1)
	require.NoError(t, err)
	assert.Len(t, nb.Nodes, 1, "dismissed edges are not traversed")
	assert.Error(t, topo.SetEdgeStatus(ctx, sugg[0].ID, "suggested"))

	// Derived edges follow inventory changes, and deleted resources leave the graph.
	_, err = repo.AssignSubnetToVLAN(ctx, child.ID, nil)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `DELETE FROM devices WHERE id=$1`, other.ID)
	require.NoError(t, err)
	require.NoError(t, topo.SyncGraph(ctx))
	g, err = topo.GetFullGraph(ctx, 100)
	require.NoError(t, err)
	assert.Len(t, g.Nodes, 4)
	for _, e := range g.Edges {
		assert.NotContains(t, e.Data["target"], "vlan_")
	}
}

func itoa(i int64) string { return strconv.FormatInt(i, 10) }
