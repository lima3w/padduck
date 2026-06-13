package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"padduck/models"
)

func createTestCFDef(t *testing.T, r *Repository, entityType, name, fieldType string, order int) *models.CustomFieldDefinition {
	t.Helper()
	def, err := r.CreateCustomFieldDefinition(context.Background(), &CustomFieldDefinitionParams{
		EntityType:   entityType,
		Name:         name,
		Label:        name,
		FieldType:    fieldType,
		DisplayOrder: order,
		IsSearchable: true,
	})
	require.NoError(t, err)
	return def
}

func TestCustomFieldDefinitionCRUD_Integration(t *testing.T) {
	r := testRepo(t)
	ctx := context.Background()

	def := createTestCFDef(t, r, "device", "warranty_url", "url", 1)

	got, err := r.GetCustomFieldDefinition(ctx, def.ID)
	require.NoError(t, err)
	assert.Equal(t, "warranty_url", got.Name)
	assert.Equal(t, "url", got.FieldType)

	_, err = r.UpdateCustomFieldDefinition(ctx, def.ID, &CustomFieldDefinitionParams{
		EntityType: "device",
		Name:       "warranty_url",
		Label:      "Warranty Link",
		FieldType:  "url",
	})
	require.NoError(t, err)
	got, err = r.GetCustomFieldDefinition(ctx, def.ID)
	require.NoError(t, err)
	assert.Equal(t, "Warranty Link", got.Label)

	require.NoError(t, r.DeleteCustomFieldDefinition(ctx, def.ID))
	_, err = r.GetCustomFieldDefinition(ctx, def.ID)
	assert.Error(t, err, "deleted definition must not be retrievable")
}

func TestCustomFieldDefinitions_ListAndReorder_Integration(t *testing.T) {
	r := testRepo(t)
	ctx := context.Background()

	a := createTestCFDef(t, r, "device", "field_a", "text", 1)
	b := createTestCFDef(t, r, "device", "field_b", "text", 2)
	c := createTestCFDef(t, r, "device", "field_c", "text", 3)
	// A different entity type must not appear in the device list.
	createTestCFDef(t, r, "subnet", "subnet_only", "text", 1)

	defs, err := r.ListCustomFieldDefinitions(ctx, "device")
	require.NoError(t, err)
	require.Len(t, defs, 3)
	assert.Equal(t, []string{"field_a", "field_b", "field_c"},
		[]string{defs[0].Name, defs[1].Name, defs[2].Name})

	// Reverse the order.
	require.NoError(t, r.ReorderCustomFieldDefinitions(ctx, []int64{c.ID, b.ID, a.ID}))
	defs, err = r.ListCustomFieldDefinitions(ctx, "device")
	require.NoError(t, err)
	require.Len(t, defs, 3)
	assert.Equal(t, []string{"field_c", "field_b", "field_a"},
		[]string{defs[0].Name, defs[1].Name, defs[2].Name})
}

func TestCustomFieldValues_SetGetUpsert_Integration(t *testing.T) {
	r := testRepo(t)
	ctx := context.Background()

	def := createTestCFDef(t, r, "device", "rack_notes", "text", 1)
	deviceID := createTestDevice(t, r, "cf-host", "V", "M")
	defs := []*models.CustomFieldDefinition{def}

	require.NoError(t, r.SetCustomFieldValues(ctx, "device", deviceID, defs,
		map[string]*string{"rack_notes": strPtr("front of rack 12")}))

	values, err := r.GetCustomFieldValues(ctx, "device", deviceID)
	require.NoError(t, err)
	require.Contains(t, values, "rack_notes")
	require.NotNil(t, values["rack_notes"])
	assert.Equal(t, "front of rack 12", *values["rack_notes"])

	// Upsert path: setting again must update, not duplicate.
	require.NoError(t, r.SetCustomFieldValues(ctx, "device", deviceID, defs,
		map[string]*string{"rack_notes": strPtr("moved to rack 14")}))
	values, err = r.GetCustomFieldValues(ctx, "device", deviceID)
	require.NoError(t, err)
	require.NotNil(t, values["rack_notes"])
	assert.Equal(t, "moved to rack 14", *values["rack_notes"])

	// Values for names without a matching definition are ignored.
	require.NoError(t, r.SetCustomFieldValues(ctx, "device", deviceID, defs,
		map[string]*string{"unknown_field": strPtr("x")}))
	values, err = r.GetCustomFieldValues(ctx, "device", deviceID)
	require.NoError(t, err)
	assert.NotContains(t, values, "unknown_field")
}

func TestSearchSubnetsWithCustomFields_Integration(t *testing.T) {
	r := testRepo(t)
	ctx := context.Background()

	networkID := createTestNetwork(t, r)
	prodID := createTestSubnet(t, r, networkID, "10.10.0.0", 24)
	stageID := createTestSubnet(t, r, networkID, "10.20.0.0", 24)

	envDef := createTestCFDef(t, r, "subnet", "environment", "dropdown", 1) // non-text: exact match
	urlDef := createTestCFDef(t, r, "subnet", "runbook", "url", 2)         // text-like: ILIKE match
	defs := []*models.CustomFieldDefinition{envDef, urlDef}

	require.NoError(t, r.SetCustomFieldValues(ctx, "subnet", prodID, defs, map[string]*string{
		"environment": strPtr("production"),
		"runbook":     strPtr("https://wiki.example.com/prod-network"),
	}))
	require.NoError(t, r.SetCustomFieldValues(ctx, "subnet", stageID, defs, map[string]*string{
		"environment": strPtr("staging"),
		"runbook":     strPtr("https://wiki.example.com/staging-network"),
	}))

	// Exact match on a non-text field type.
	subnets, err := r.SearchSubnetsWithCustomFields(ctx, networkID, "", 10, 0,
		map[string]string{"environment": "production"})
	require.NoError(t, err)
	require.Len(t, subnets, 1)
	assert.Equal(t, prodID, subnets[0].ID)

	// Partial value on a non-text type must NOT match (exact compare).
	subnets, err = r.SearchSubnetsWithCustomFields(ctx, networkID, "", 10, 0,
		map[string]string{"environment": "prod"})
	require.NoError(t, err)
	assert.Empty(t, subnets)

	// ILIKE partial match on a text-like field type.
	subnets, err = r.SearchSubnetsWithCustomFields(ctx, networkID, "", 10, 0,
		map[string]string{"runbook": "STAGING"})
	require.NoError(t, err)
	require.Len(t, subnets, 1)
	assert.Equal(t, stageID, subnets[0].ID)

	// Multiple filters AND together.
	subnets, err = r.SearchSubnetsWithCustomFields(ctx, networkID, "", 10, 0,
		map[string]string{"environment": "production", "runbook": "staging"})
	require.NoError(t, err)
	assert.Empty(t, subnets)

	// Unknown filter names are ignored rather than failing the query.
	subnets, err = r.SearchSubnetsWithCustomFields(ctx, networkID, "", 10, 0,
		map[string]string{"nonexistent": "x"})
	require.NoError(t, err)
	assert.Len(t, subnets, 2)

	// Filter values are parameterized: an injection attempt is just a non-match.
	subnets, err = r.SearchSubnetsWithCustomFields(ctx, networkID, "", 10, 0,
		map[string]string{"environment": `' OR '1'='1`})
	require.NoError(t, err)
	assert.Empty(t, subnets)
}

func TestSearchIPAddressesWithCustomFields_Integration(t *testing.T) {
	r := testRepo(t)
	ctx := context.Background()

	networkID := createTestNetwork(t, r)
	subnetID := createTestSubnet(t, r, networkID, "192.168.50.0", 24)

	web, err := r.CreateIPAddress(ctx, subnetID, "192.168.50.10", "web-01", "assigned", nil, nil, nil, nil)
	require.NoError(t, err)
	db, err := r.CreateIPAddress(ctx, subnetID, "192.168.50.20", "db-01", "assigned", nil, nil, nil, nil)
	require.NoError(t, err)

	roleDef := createTestCFDef(t, r, "ip_address", "service_role", "dropdown", 1)
	defs := []*models.CustomFieldDefinition{roleDef}
	require.NoError(t, r.SetCustomFieldValues(ctx, "ip_address", web.ID, defs, map[string]*string{"service_role": strPtr("frontend")}))
	require.NoError(t, r.SetCustomFieldValues(ctx, "ip_address", db.ID, defs, map[string]*string{"service_role": strPtr("database")}))

	// Custom field filter narrows to the matching IP.
	ips, err := r.SearchIPAddressesWithCustomFields(ctx, subnetID, "", "", 10, 0, IPSearchFilter{},
		map[string]string{"service_role": "frontend"})
	require.NoError(t, err)
	require.Len(t, ips, 1)
	assert.Equal(t, "web-01", ips[0].Hostname)

	// Combined with the base query filter.
	ips, err = r.SearchIPAddressesWithCustomFields(ctx, subnetID, "db-01", "", 10, 0, IPSearchFilter{},
		map[string]string{"service_role": "database"})
	require.NoError(t, err)
	require.Len(t, ips, 1)

	// Base query and CF filter that contradict each other return nothing.
	ips, err = r.SearchIPAddressesWithCustomFields(ctx, subnetID, "web-01", "", 10, 0, IPSearchFilter{},
		map[string]string{"service_role": "database"})
	require.NoError(t, err)
	assert.Empty(t, ips)
}

