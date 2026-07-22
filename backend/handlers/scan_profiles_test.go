package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"padduck/models"
)

// ---------------------------------------------------------------------------
// redactScanProfile / redactScanProfiles — pure functions, no service
// dependency, fully testable.
// ---------------------------------------------------------------------------

func strPtrScanProfile(s string) *string { return &s }

func TestRedactScanProfile_Nil(t *testing.T) {
	assert.Nil(t, redactScanProfile(nil))
}

func TestRedactScanProfile_NoCommunitySet(t *testing.T) {
	p := &models.ScanProfile{ID: 1, Name: "default"}
	got := redactScanProfile(p)
	assert.Same(t, p, got)
	assert.Nil(t, got.SNMPCommunity)
}

func TestRedactScanProfile_EmptyCommunity(t *testing.T) {
	p := &models.ScanProfile{ID: 1, Name: "default", SNMPCommunity: strPtrScanProfile("")}
	got := redactScanProfile(p)
	assert.Same(t, p, got)
	assert.Equal(t, "", *got.SNMPCommunity)
}

func TestRedactScanProfile_CommunitySet(t *testing.T) {
	p := &models.ScanProfile{ID: 1, Name: "default", SNMPCommunity: strPtrScanProfile("public")}
	got := redactScanProfile(p)
	// The original must not be mutated — redaction returns a clone.
	assert.NotSame(t, p, got)
	assert.Equal(t, "public", *p.SNMPCommunity)
	assert.Equal(t, "***", *got.SNMPCommunity)
	assert.Equal(t, p.ID, got.ID)
	assert.Equal(t, p.Name, got.Name)
}

func TestRedactScanProfiles_Empty(t *testing.T) {
	got := redactScanProfiles(nil)
	assert.NotNil(t, got)
	assert.Empty(t, got)
}

func TestRedactScanProfiles_Multiple(t *testing.T) {
	profiles := []*models.ScanProfile{
		{ID: 1, Name: "a", SNMPCommunity: strPtrScanProfile("public")},
		{ID: 2, Name: "b"},
	}
	got := redactScanProfiles(profiles)
	assert.Len(t, got, 2)
	assert.Equal(t, "***", *got[0].SNMPCommunity)
	assert.Nil(t, got[1].SNMPCommunity)
}

// ---------------------------------------------------------------------------
// All seven route handlers check requirePerm before touching params/body,
// so the deeper validation/business-logic branches are only reachable by a
// permitted user — which requires a live repo (see plan). Only the auth
// guard branches are testable here without a DB.
// ---------------------------------------------------------------------------

func TestListScanProfiles_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/scan-profiles", h.ListScanProfiles)

	req := httptest.NewRequest("GET", "/admin/scan-profiles", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListScanProfiles_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/scan-profiles", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.ListScanProfiles(c)
	})

	req := httptest.NewRequest("GET", "/admin/scan-profiles", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestCreateScanProfile_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/scan-profiles", h.CreateScanProfile)

	req := httptest.NewRequest("POST", "/admin/scan-profiles", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateScanProfile_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/scan-profiles", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.CreateScanProfile(c)
	})

	req := httptest.NewRequest("POST", "/admin/scan-profiles", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetScanProfile_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/scan-profiles/:id", h.GetScanProfile)

	req := httptest.NewRequest("GET", "/admin/scan-profiles/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetScanProfile_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/scan-profiles/:id", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.GetScanProfile(c)
	})

	req := httptest.NewRequest("GET", "/admin/scan-profiles/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUpdateScanProfile_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/admin/scan-profiles/:id", h.UpdateScanProfile)

	req := httptest.NewRequest("PUT", "/admin/scan-profiles/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateScanProfile_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/admin/scan-profiles/:id", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.UpdateScanProfile(c)
	})

	req := httptest.NewRequest("PUT", "/admin/scan-profiles/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestDeleteScanProfile_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Delete("/admin/scan-profiles/:id", h.DeleteScanProfile)

	req := httptest.NewRequest("DELETE", "/admin/scan-profiles/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteScanProfile_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Delete("/admin/scan-profiles/:id", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.DeleteScanProfile(c)
	})

	req := httptest.NewRequest("DELETE", "/admin/scan-profiles/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetSubnetScanProfile_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/subnets/:id/scan-profile", h.GetSubnetScanProfile)

	req := httptest.NewRequest("GET", "/admin/subnets/1/scan-profile", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetSubnetScanProfile_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/subnets/:id/scan-profile", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.GetSubnetScanProfile(c)
	})

	req := httptest.NewRequest("GET", "/admin/subnets/1/scan-profile", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestSetSubnetScanProfile_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/admin/subnets/:id/scan-profile", h.SetSubnetScanProfile)

	req := httptest.NewRequest("PUT", "/admin/subnets/1/scan-profile", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestSetSubnetScanProfile_NoPermission_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/admin/subnets/:id/scan-profile", func(c *fiber.Ctx) error {
		c.Locals("user", permUser())
		return h.SetSubnetScanProfile(c)
	})

	req := httptest.NewRequest("PUT", "/admin/subnets/1/scan-profile", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
