package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"padduck/models"
)

// ---------------------------------------------------------------------------
// Helper: build a minimal Fiber app that injects a user into locals
// ---------------------------------------------------------------------------

func buildExternalAuthApp(user *models.User, method, route string, handler fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Add(method, route, func(c *fiber.Ctx) error {
		if user != nil {
			c.Locals("user", user)
		}
		return handler(c)
	})
	return app
}

// adminUser is a user with role=admin and a non-zero ID so permCheck passes.
var adminExtUser = &models.User{ID: 1, Username: "admin", Email: "admin@example.com", Role: "admin"}

// unprivExtUser has ID=0 so CheckPermission returns "permission denied" immediately
// (userID <= 0 path), which permCheck maps to 403.
var unprivExtUser = &models.User{ID: 0, Username: "user1", Email: "user@example.com", Role: "user"}

// ============================================================
// permCheck enforcement — no user / non-admin → 403
// ============================================================

func TestGetLDAPConfig_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/auth/ldap", h.GetLDAPConfig)

	req := httptest.NewRequest("GET", "/admin/auth/ldap", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetLDAPConfig_UnprivUser_Returns403(t *testing.T) {
	h := minHandler()
	app := buildExternalAuthApp(unprivExtUser, "GET", "/admin/auth/ldap", h.GetLDAPConfig)

	req := httptest.NewRequest("GET", "/admin/auth/ldap", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUpdateLDAPConfig_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/admin/auth/ldap", h.UpdateLDAPConfig)

	req := httptest.NewRequest("PUT", "/admin/auth/ldap", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestTestLDAPConnection_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/auth/ldap/test", h.TestLDAPConnection)

	req := httptest.NewRequest("POST", "/admin/auth/ldap/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestTestLDAPConnection_UnprivUser_Returns403(t *testing.T) {
	h := minHandler()
	app := buildExternalAuthApp(unprivExtUser, "POST", "/admin/auth/ldap/test", h.TestLDAPConnection)

	req := httptest.NewRequest("POST", "/admin/auth/ldap/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestListLDAPGroupMappings_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/auth/ldap/group-mappings", h.ListLDAPGroupMappings)

	req := httptest.NewRequest("GET", "/admin/auth/ldap/group-mappings", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListLDAPGroupMappings_UnprivUser_Returns403(t *testing.T) {
	h := minHandler()
	app := buildExternalAuthApp(unprivExtUser, "GET", "/admin/auth/ldap/group-mappings", h.ListLDAPGroupMappings)

	req := httptest.NewRequest("GET", "/admin/auth/ldap/group-mappings", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestCreateLDAPGroupMapping_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/admin/auth/ldap/group-mappings", h.CreateLDAPGroupMapping)

	req := httptest.NewRequest("POST", "/admin/auth/ldap/group-mappings", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateLDAPGroupMapping_UnprivUser_Returns403(t *testing.T) {
	h := minHandler()
	app := buildExternalAuthApp(unprivExtUser, "POST", "/admin/auth/ldap/group-mappings", h.CreateLDAPGroupMapping)

	req := httptest.NewRequest("POST", "/admin/auth/ldap/group-mappings", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestDeleteLDAPGroupMapping_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Delete("/admin/auth/ldap/group-mappings/:id", h.DeleteLDAPGroupMapping)

	req := httptest.NewRequest("DELETE", "/admin/auth/ldap/group-mappings/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteLDAPGroupMapping_UnprivUser_Returns403(t *testing.T) {
	h := minHandler()
	app := buildExternalAuthApp(unprivExtUser, "DELETE", "/admin/auth/ldap/group-mappings/:id", h.DeleteLDAPGroupMapping)

	req := httptest.NewRequest("DELETE", "/admin/auth/ldap/group-mappings/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetOAuth2Config_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/auth/oauth2", h.GetOAuth2Config)

	req := httptest.NewRequest("GET", "/admin/auth/oauth2", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetOAuth2Config_UnprivUser_Returns403(t *testing.T) {
	h := minHandler()
	app := buildExternalAuthApp(unprivExtUser, "GET", "/admin/auth/oauth2", h.GetOAuth2Config)

	req := httptest.NewRequest("GET", "/admin/auth/oauth2", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUpdateOAuth2Config_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/admin/auth/oauth2", h.UpdateOAuth2Config)

	req := httptest.NewRequest("PUT", "/admin/auth/oauth2", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateOAuth2Config_UnprivUser_Returns403(t *testing.T) {
	h := minHandler()
	app := buildExternalAuthApp(unprivExtUser, "PUT", "/admin/auth/oauth2", h.UpdateOAuth2Config)

	req := httptest.NewRequest("PUT", "/admin/auth/oauth2", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetSAMLConfig_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/auth/saml", h.GetSAMLConfig)

	req := httptest.NewRequest("GET", "/admin/auth/saml", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetSAMLConfig_UnprivUser_Returns403(t *testing.T) {
	h := minHandler()
	app := buildExternalAuthApp(unprivExtUser, "GET", "/admin/auth/saml", h.GetSAMLConfig)

	req := httptest.NewRequest("GET", "/admin/auth/saml", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUpdateSAMLConfig_NoUser_Returns401(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Put("/admin/auth/saml", h.UpdateSAMLConfig)

	req := httptest.NewRequest("PUT", "/admin/auth/saml", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateSAMLConfig_UnprivUser_Returns403(t *testing.T) {
	h := minHandler()
	app := buildExternalAuthApp(unprivExtUser, "PUT", "/admin/auth/saml", h.UpdateSAMLConfig)

	req := httptest.NewRequest("PUT", "/admin/auth/saml", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ============================================================
// Public endpoints — input validation (no DB needed)
// ============================================================

func TestLDAPLogin_EmptyBody_Returns400(t *testing.T) {
	// The handler validates credentials before reaching the LDAP service.
	// An empty body (no username/password) → 400 BadRequest.
	h := minHandler()
	app := fiber.New()
	app.Post("/auth/ldap/login", h.LDAPLogin)

	req := httptest.NewRequest("POST", "/auth/ldap/login", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestLDAPLogin_MissingPassword_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Post("/auth/ldap/login", h.LDAPLogin)

	req := httptest.NewRequest("POST", "/auth/ldap/login", strings.NewReader(`{"username":"user"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestOAuth2Callback_MissingParams_Returns400(t *testing.T) {
	// Missing code and state → 400 before reaching the OAuth2 service.
	h := minHandler()
	app := fiber.New()
	app.Get("/auth/oauth2/callback", h.OAuth2Callback)

	req := httptest.NewRequest("GET", "/auth/oauth2/callback", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestOAuth2Callback_MissingState_Returns400(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/auth/oauth2/callback", h.OAuth2Callback)

	req := httptest.NewRequest("GET", "/auth/oauth2/callback?code=abc", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestSAMLACS_MissingSAMLResponse_Returns400(t *testing.T) {
	// A POST to ACS with no SAMLResponse form field → 400 before reaching service.
	h := minHandler()
	app := fiber.New()
	app.Post("/auth/saml/acs", h.SAMLAssertionConsumerService)

	req := httptest.NewRequest("POST", "/auth/saml/acs", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ============================================================
// maskSecret helper
// ============================================================

func TestMaskSecret_Set(t *testing.T) {
	assert.Equal(t, "****", maskSecret(true))
}

func TestMaskSecret_NotSet(t *testing.T) {
	assert.Equal(t, "", maskSecret(false))
}
