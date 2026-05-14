package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ipam-next/models"
	"ipam-next/repository"
)

// ---------------------------------------------------------------------------
// CreateLocation — input validation (no DB needed)
// ---------------------------------------------------------------------------

func TestCreateLocation_EmptyName_ReturnsError(t *testing.T) {
	svc := &Service{}
	_, err := svc.CreateLocation(context.Background(), &repository.LocationParams{Name: "", Type: "site"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestCreateLocation_DefaultsTypeToOther(t *testing.T) {
	// Type defaulting happens before the repo call; confirm it mutates req.
	svc := &Service{}
	req := &repository.LocationParams{Name: "HQ"} // Type intentionally empty
	// We expect an error from the nil repo, not from validation.
	// If Type were not defaulted first, the "other" default would never be set.
	req.Type = "" // trigger the default branch
	// Call will panic on nil repo after validation passes; recover it.
	func() {
		defer func() { recover() }()
		_, _ = svc.CreateLocation(context.Background(), req)
	}()
	assert.Equal(t, "other", req.Type)
}

// ---------------------------------------------------------------------------
// GetLocation — input validation
// ---------------------------------------------------------------------------

func TestGetLocation_InvalidID_ReturnsError(t *testing.T) {
	svc := &Service{}
	_, err := svc.GetLocation(context.Background(), 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid location ID")

	_, err = svc.GetLocation(context.Background(), -5)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid location ID")
}

// ---------------------------------------------------------------------------
// UpdateLocation — input validation
// ---------------------------------------------------------------------------

func TestUpdateLocation_InvalidID_ReturnsError(t *testing.T) {
	svc := &Service{}
	_, err := svc.UpdateLocation(context.Background(), 0, &repository.LocationParams{Name: "X"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid location ID")
}

func TestUpdateLocation_EmptyName_ReturnsError(t *testing.T) {
	svc := &Service{}
	_, err := svc.UpdateLocation(context.Background(), 1, &repository.LocationParams{Name: ""})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

// ---------------------------------------------------------------------------
// DeleteLocation — input validation
// ---------------------------------------------------------------------------

func TestDeleteLocation_InvalidID_ReturnsError(t *testing.T) {
	svc := &Service{}
	err := svc.DeleteLocation(context.Background(), 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid location ID")
}

// ---------------------------------------------------------------------------
// buildLocationTree — pure function, no DB
// ---------------------------------------------------------------------------

func TestBuildLocationTree_EmptyList_ReturnsEmptySlice(t *testing.T) {
	result := buildLocationTree(nil)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)

	result = buildLocationTree([]*models.Location{})
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestBuildLocationTree_FlatList_AllRoots(t *testing.T) {
	locs := []*models.Location{
		{ID: 1, Name: "Site A"},
		{ID: 2, Name: "Site B"},
	}
	result := buildLocationTree(locs)
	assert.Len(t, result, 2)
	assert.Len(t, result[0].Children, 0)
	assert.Len(t, result[1].Children, 0)
}

func TestBuildLocationTree_NestedLocations(t *testing.T) {
	parentID := int64(1)
	grandparentID := int64(2)
	locs := []*models.Location{
		{ID: 1, Name: "Site"},
		{ID: 2, ParentID: &parentID, Name: "Building"},
		{ID: 3, ParentID: &grandparentID, Name: "Floor"},
	}
	result := buildLocationTree(locs)

	require.Len(t, result, 1)
	assert.Equal(t, "Site", result[0].Name)

	require.Len(t, result[0].Children, 1)
	assert.Equal(t, "Building", result[0].Children[0].Name)

	require.Len(t, result[0].Children[0].Children, 1)
	assert.Equal(t, "Floor", result[0].Children[0].Children[0].Name)
}

func TestBuildLocationTree_OrphanedNode_DroppedFromTree(t *testing.T) {
	missingParent := int64(99)
	locs := []*models.Location{
		{ID: 1, Name: "Site"},
		{ID: 2, ParentID: &missingParent, Name: "Orphan"},
	}
	result := buildLocationTree(locs)
	// Only the root should appear; orphan's parent doesn't exist in the map.
	assert.Len(t, result, 1)
	assert.Equal(t, "Site", result[0].Name)
}
