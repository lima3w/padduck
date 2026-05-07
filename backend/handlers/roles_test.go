package handlers

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"ipam-next/models"
	"ipam-next/services"
)

// buildRolesApp creates a minimal Fiber app that injects the given user into
// locals before invoking the target handler. Pass nil user to test unauthenticated.
func buildRolesApp(user *models.User, route string, method string, handler fiber.Handler) *fiber.App {
	h := &Handler{service: nil}
	app := fiber.New(fiber.Config{
		// Convert panics to 500 so tests don't crash.
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
	_ = h
	return app
}

func parseRolesResponse(t *testing.T, body io.Reader) map[string]interface{} {
	t.Helper()
	data, err := io.ReadAll(body)
	assert.NoError(t, err)
	var result map[string]interface{}
	_ = json.Unmarshal(data, &result)
	return result
}

// ---------------------------------------------------------------------------
// Request struct validation
// ---------------------------------------------------------------------------

func TestCreateRoleRequest_Fields(t *testing.T) {
	req := &CreateRoleRequest{Name: "network-admin", Description: "Manages network resources"}
	assert.Equal(t, "network-admin", req.Name)
	assert.Equal(t, "Manages network resources", req.Description)
}

func TestUpdateRoleRequest_Fields(t *testing.T) {
	req := &UpdateRoleRequest{Name: "updated-role", Description: "Updated desc"}
	assert.Equal(t, "updated-role", req.Name)
	assert.Equal(t, "Updated desc", req.Description)
}

func TestAddPermissionRequest_Fields(t *testing.T) {
	rt := "subnet"
	rid := int64(5)
	req := &AddPermissionRequest{
		Permission:   services.PermV2SubnetRead,
		ResourceType: &rt,
		ResourceID:   &rid,
	}
	assert.Equal(t, services.PermV2SubnetRead, req.Permission)
	assert.Equal(t, "subnet", *req.ResourceType)
	assert.Equal(t, int64(5), *req.ResourceID)
}

func TestAssignRoleRequest_Fields(t *testing.T) {
	req := &AssignRoleRequest{RoleID: 3}
	assert.Equal(t, int64(3), req.RoleID)
}

// ---------------------------------------------------------------------------
// ListRoles — auth enforcement
// ---------------------------------------------------------------------------

func TestListRoles_NoUser_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/roles", h.ListRoles)

	req := httptest.NewRequest("GET", "/admin/roles", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestListRoles_NonAdmin_Returns403(t *testing.T) {
	for _, role := range []string{"user", "viewer", "operator"} {
		t.Run(role, func(t *testing.T) {
			h := &Handler{service: nil}
			app := fiber.New()
			app.Get("/admin/roles", func(c *fiber.Ctx) error {
				c.Locals("user", &models.User{Role: role})
				return h.ListRoles(c)
			})
			req := httptest.NewRequest("GET", "/admin/roles", nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
			body := parseRolesResponse(t, resp.Body)
			assert.Equal(t, string(ErrForbidden), body["code"])
		})
	}
}

// ---------------------------------------------------------------------------
// GetRole — auth enforcement and param validation
// ---------------------------------------------------------------------------

func TestGetRole_NoUser_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/roles/:id", h.GetRole)

	req := httptest.NewRequest("GET", "/admin/roles/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetRole_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/roles/:id", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.GetRole(c)
	})
	req := httptest.NewRequest("GET", "/admin/roles/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetRole_AdminInvalidID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/roles/:id", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "admin"})
		return h.GetRole(c)
	})
	req := httptest.NewRequest("GET", "/admin/roles/notanumber", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CreateRole — auth enforcement
// ---------------------------------------------------------------------------

func TestCreateRole_NoUser_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/roles", h.CreateRole)

	req := httptest.NewRequest("POST", "/admin/roles", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestCreateRole_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/roles", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "viewer"})
		return h.CreateRole(c)
	})
	req := httptest.NewRequest("POST", "/admin/roles", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdateRole — auth enforcement and param validation
// ---------------------------------------------------------------------------

func TestUpdateRole_NoUser_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/admin/roles/:id", h.UpdateRole)

	req := httptest.NewRequest("PUT", "/admin/roles/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUpdateRole_AdminInvalidID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Put("/admin/roles/:id", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "admin"})
		return h.UpdateRole(c)
	})
	req := httptest.NewRequest("PUT", "/admin/roles/bad", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DeleteRole — auth enforcement and param validation
// ---------------------------------------------------------------------------

func TestDeleteRole_NoUser_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/admin/roles/:id", h.DeleteRole)

	req := httptest.NewRequest("DELETE", "/admin/roles/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestDeleteRole_AdminInvalidID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/admin/roles/:id", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "admin"})
		return h.DeleteRole(c)
	})
	req := httptest.NewRequest("DELETE", "/admin/roles/xyz", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// AddPermissionToRole — auth enforcement and param validation
// ---------------------------------------------------------------------------

func TestAddPermissionToRole_NoUser_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/roles/:id/permissions", h.AddPermissionToRole)

	req := httptest.NewRequest("POST", "/admin/roles/1/permissions", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestAddPermissionToRole_AdminInvalidID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/roles/:id/permissions", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "admin"})
		return h.AddPermissionToRole(c)
	})
	req := httptest.NewRequest("POST", "/admin/roles/bad/permissions", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// RemovePermissionFromRole — auth enforcement and param validation
// ---------------------------------------------------------------------------

func TestRemovePermissionFromRole_NoUser_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/admin/roles/:id/permissions/:perm_id", h.RemovePermissionFromRole)

	req := httptest.NewRequest("DELETE", "/admin/roles/1/permissions/2", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestRemovePermissionFromRole_AdminInvalidPermID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/admin/roles/:id/permissions/:perm_id", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "admin"})
		return h.RemovePermissionFromRole(c)
	})
	req := httptest.NewRequest("DELETE", "/admin/roles/1/permissions/bad", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ListAvailablePermissions — full happy path (no DB needed)
// ---------------------------------------------------------------------------

func TestListAvailablePermissions_NoUser_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/permissions", h.ListAvailablePermissions)

	req := httptest.NewRequest("GET", "/admin/permissions", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestListAvailablePermissions_Admin_Returns200WithAllPerms(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/permissions", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "admin"})
		return h.ListAvailablePermissions(c)
	})
	req := httptest.NewRequest("GET", "/admin/permissions", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	data, _ := io.ReadAll(resp.Body)
	var perms []string
	assert.NoError(t, json.Unmarshal(data, &perms))
	assert.Equal(t, len(services.AllPermissions), len(perms))
	for _, expected := range services.AllPermissions {
		assert.Contains(t, perms, expected)
	}
}

// ---------------------------------------------------------------------------
// GetUserRoles — auth enforcement and param validation
// ---------------------------------------------------------------------------

func TestGetUserRoles_NoUser_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/users/:id/roles", h.GetUserRoles)

	req := httptest.NewRequest("GET", "/admin/users/1/roles", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGetUserRoles_AdminInvalidID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/users/:id/roles", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "admin"})
		return h.GetUserRoles(c)
	})
	req := httptest.NewRequest("GET", "/admin/users/bad/roles", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// AssignRoleToUser — auth enforcement and param validation
// ---------------------------------------------------------------------------

func TestAssignRoleToUser_NoUser_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/users/:id/roles", h.AssignRoleToUser)

	req := httptest.NewRequest("POST", "/admin/users/1/roles", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestAssignRoleToUser_AdminInvalidID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/users/:id/roles", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "admin"})
		return h.AssignRoleToUser(c)
	})
	req := httptest.NewRequest("POST", "/admin/users/bad/roles", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// RemoveRoleFromUser — auth enforcement and param validation
// ---------------------------------------------------------------------------

func TestRemoveRoleFromUser_NoUser_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/admin/users/:id/roles/:role_id", h.RemoveRoleFromUser)

	req := httptest.NewRequest("DELETE", "/admin/users/1/roles/2", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestRemoveRoleFromUser_AdminInvalidUserID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/admin/users/:id/roles/:role_id", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "admin"})
		return h.RemoveRoleFromUser(c)
	})
	req := httptest.NewRequest("DELETE", "/admin/users/bad/roles/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestRemoveRoleFromUser_AdminInvalidRoleID_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/admin/users/:id/roles/:role_id", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "admin"})
		return h.RemoveRoleFromUser(c)
	})
	req := httptest.NewRequest("DELETE", "/admin/users/1/roles/bad", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
