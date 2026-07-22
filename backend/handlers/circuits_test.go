package handlers

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// The three circuit resources (providers, physical circuits, logical
// circuits) share an identical handler shape:
//   - List:   requirePerm only, no params — guard branches only.
//   - Create: body parsed BEFORE requirePerm — invalid-body is reachable
//     with no auth, then guard branches with a valid body.
//   - Update: :id then body parsed BEFORE requirePerm — invalid-ID and
//     (valid-ID) invalid-body are reachable with no auth, then guard
//     branches with valid ID+body.
//   - Delete: :id parsed BEFORE requirePerm — invalid-ID is reachable with
//     no auth, then guard branches with a valid ID.
//
// testCircuitCRUDGuards drives that shared battery once per resource so the
// three (near-identical) handler groups don't need triplicated test code.

func testCircuitCRUDGuards(t *testing.T, resource, basePath string, list, create, update, del fiber.Handler) {
	t.Helper()
	itemPath := basePath + "/:id"

	t.Run(resource+"/List_NoUser_Returns401", func(t *testing.T) {
		app := fiber.New()
		app.Get(basePath, list)
		req := httptest.NewRequest("GET", basePath, nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run(resource+"/List_NoPermission_Returns403", func(t *testing.T) {
		app := fiber.New()
		app.Get(basePath, func(c *fiber.Ctx) error {
			c.Locals("user", permUser())
			return list(c)
		})
		req := httptest.NewRequest("GET", basePath, nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})

	t.Run(resource+"/Create_InvalidBody_Returns400", func(t *testing.T) {
		app := fiber.New()
		app.Post(basePath, create)
		req := httptest.NewRequest("POST", basePath, strings.NewReader("{not valid json"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		data, _ := io.ReadAll(resp.Body)
		var body ErrorResponse
		assert.NoError(t, json.Unmarshal(data, &body))
		assert.Equal(t, string(ErrBadRequest), body.Code)
	})

	t.Run(resource+"/Create_ValidBodyNoUser_Returns401", func(t *testing.T) {
		app := fiber.New()
		app.Post(basePath, create)
		req := httptest.NewRequest("POST", basePath, strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run(resource+"/Create_ValidBodyNoPermission_Returns403", func(t *testing.T) {
		app := fiber.New()
		app.Post(basePath, func(c *fiber.Ctx) error {
			c.Locals("user", permUser())
			return create(c)
		})
		req := httptest.NewRequest("POST", basePath, strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})

	t.Run(resource+"/Update_InvalidID_Returns400", func(t *testing.T) {
		app := fiber.New()
		app.Put(itemPath, update)
		req := httptest.NewRequest("PUT", basePath+"/not-a-number", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run(resource+"/Update_ValidIDInvalidBody_Returns400", func(t *testing.T) {
		app := fiber.New()
		app.Put(itemPath, update)
		req := httptest.NewRequest("PUT", basePath+"/1", strings.NewReader("{not valid json"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run(resource+"/Update_ValidIDValidBodyNoUser_Returns401", func(t *testing.T) {
		app := fiber.New()
		app.Put(itemPath, update)
		req := httptest.NewRequest("PUT", basePath+"/1", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run(resource+"/Update_ValidIDValidBodyNoPermission_Returns403", func(t *testing.T) {
		app := fiber.New()
		app.Put(itemPath, func(c *fiber.Ctx) error {
			c.Locals("user", permUser())
			return update(c)
		})
		req := httptest.NewRequest("PUT", basePath+"/1", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})

	t.Run(resource+"/Delete_InvalidID_Returns400", func(t *testing.T) {
		app := fiber.New()
		app.Delete(itemPath, del)
		req := httptest.NewRequest("DELETE", basePath+"/not-a-number", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run(resource+"/Delete_ValidIDNoUser_Returns401", func(t *testing.T) {
		app := fiber.New()
		app.Delete(itemPath, del)
		req := httptest.NewRequest("DELETE", basePath+"/1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run(resource+"/Delete_ValidIDNoPermission_Returns403", func(t *testing.T) {
		app := fiber.New()
		app.Delete(itemPath, func(c *fiber.Ctx) error {
			c.Locals("user", permUser())
			return del(c)
		})
		req := httptest.NewRequest("DELETE", basePath+"/1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})
}

func TestCircuitProviderCRUD_Guards(t *testing.T) {
	h := minHandler()
	testCircuitCRUDGuards(t, "CircuitProvider", "/circuit-providers",
		h.ListCircuitProviders, h.CreateCircuitProvider, h.UpdateCircuitProvider, h.DeleteCircuitProvider)
}

func TestPhysicalCircuitCRUD_Guards(t *testing.T) {
	h := minHandler()
	testCircuitCRUDGuards(t, "PhysicalCircuit", "/physical-circuits",
		h.ListPhysicalCircuits, h.CreatePhysicalCircuit, h.UpdatePhysicalCircuit, h.DeletePhysicalCircuit)
}

func TestLogicalCircuitCRUD_Guards(t *testing.T) {
	h := minHandler()
	testCircuitCRUDGuards(t, "LogicalCircuit", "/logical-circuits",
		h.ListLogicalCircuits, h.CreateLogicalCircuit, h.UpdateLogicalCircuit, h.DeleteLogicalCircuit)
}
