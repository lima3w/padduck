package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// The network-module resources (NAT rules, firewall zones, firewall zone
// mappings, DHCP servers, DHCP leases) share an identical handler shape:
//   - List:   requirePerm only, no params — guard branches only.
//   - Get:    (where present) requirePerm THEN :id — guard branches only,
//     since the guard blocks before the ID is parsed.
//   - Create: body parsed BEFORE requirePerm — invalid-body is reachable
//     with no auth, then guard branches with a valid body.
//   - Update: :id then body parsed BEFORE requirePerm — invalid-ID and
//     (valid-ID) invalid-body are reachable with no auth, then guard
//     branches with valid ID+body.
//   - Delete: :id parsed BEFORE requirePerm — invalid-ID is reachable with
//     no auth, then guard branches with a valid ID.
//
// testNetworkModuleCRUDGuards drives that shared battery once per resource
// so the five (near-identical) handler groups don't need repeated code.

type networkModuleHandlers struct {
	list, create, update, del fiber.Handler
	get                       fiber.Handler // nil if the resource has no single-item GET
}

func testNetworkModuleCRUDGuards(t *testing.T, resource, basePath string, hh networkModuleHandlers) {
	t.Helper()
	itemPath := basePath + "/:id"

	t.Run(resource+"/List_NoUser_Returns401", func(t *testing.T) {
		app := fiber.New()
		app.Get(basePath, hh.list)
		req := httptest.NewRequest("GET", basePath, nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run(resource+"/List_NoPermission_Returns403", func(t *testing.T) {
		app := fiber.New()
		app.Get(basePath, func(c *fiber.Ctx) error {
			c.Locals("user", permUser())
			return hh.list(c)
		})
		req := httptest.NewRequest("GET", basePath, nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})

	if hh.get != nil {
		t.Run(resource+"/Get_NoUser_Returns401", func(t *testing.T) {
			app := fiber.New()
			app.Get(itemPath, hh.get)
			req := httptest.NewRequest("GET", basePath+"/1", nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
		})

		t.Run(resource+"/Get_NoPermission_Returns403", func(t *testing.T) {
			app := fiber.New()
			app.Get(itemPath, func(c *fiber.Ctx) error {
				c.Locals("user", permUser())
				return hh.get(c)
			})
			req := httptest.NewRequest("GET", basePath+"/1", nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
		})
	}

	t.Run(resource+"/Create_InvalidBody_Returns400", func(t *testing.T) {
		app := fiber.New()
		app.Post(basePath, hh.create)
		req := httptest.NewRequest("POST", basePath, strings.NewReader("{not valid json"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run(resource+"/Create_ValidBodyNoUser_Returns401", func(t *testing.T) {
		app := fiber.New()
		app.Post(basePath, hh.create)
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
			return hh.create(c)
		})
		req := httptest.NewRequest("POST", basePath, strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})

	t.Run(resource+"/Update_InvalidID_Returns400", func(t *testing.T) {
		app := fiber.New()
		app.Put(itemPath, hh.update)
		req := httptest.NewRequest("PUT", basePath+"/not-a-number", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run(resource+"/Update_ValidIDInvalidBody_Returns400", func(t *testing.T) {
		app := fiber.New()
		app.Put(itemPath, hh.update)
		req := httptest.NewRequest("PUT", basePath+"/1", strings.NewReader("{not valid json"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run(resource+"/Update_ValidIDValidBodyNoUser_Returns401", func(t *testing.T) {
		app := fiber.New()
		app.Put(itemPath, hh.update)
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
			return hh.update(c)
		})
		req := httptest.NewRequest("PUT", basePath+"/1", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})

	t.Run(resource+"/Delete_InvalidID_Returns400", func(t *testing.T) {
		app := fiber.New()
		app.Delete(itemPath, hh.del)
		req := httptest.NewRequest("DELETE", basePath+"/not-a-number", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run(resource+"/Delete_ValidIDNoUser_Returns401", func(t *testing.T) {
		app := fiber.New()
		app.Delete(itemPath, hh.del)
		req := httptest.NewRequest("DELETE", basePath+"/1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run(resource+"/Delete_ValidIDNoPermission_Returns403", func(t *testing.T) {
		app := fiber.New()
		app.Delete(itemPath, func(c *fiber.Ctx) error {
			c.Locals("user", permUser())
			return hh.del(c)
		})
		req := httptest.NewRequest("DELETE", basePath+"/1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})
}

func TestNATRuleCRUD_Guards(t *testing.T) {
	h := minHandler()
	testNetworkModuleCRUDGuards(t, "NATRule", "/nat-rules", networkModuleHandlers{
		list: h.ListNATRules, get: h.GetNATRule, create: h.CreateNATRule, update: h.UpdateNATRule, del: h.DeleteNATRule,
	})
}

func TestFirewallZoneCRUD_Guards(t *testing.T) {
	h := minHandler()
	testNetworkModuleCRUDGuards(t, "FirewallZone", "/firewall-zones", networkModuleHandlers{
		list: h.ListFirewallZones, get: h.GetFirewallZone, create: h.CreateFirewallZone, update: h.UpdateFirewallZone, del: h.DeleteFirewallZone,
	})
}

func TestFirewallZoneMappingCRUD_Guards(t *testing.T) {
	h := minHandler()
	testNetworkModuleCRUDGuards(t, "FirewallZoneMapping", "/firewall-zone-mappings", networkModuleHandlers{
		list: h.ListFirewallZoneMappings, create: h.CreateFirewallZoneMapping, update: h.UpdateFirewallZoneMapping, del: h.DeleteFirewallZoneMapping,
	})
}

func TestDHCPServerCRUD_Guards(t *testing.T) {
	h := minHandler()
	testNetworkModuleCRUDGuards(t, "DHCPServer", "/dhcp-servers", networkModuleHandlers{
		list: h.ListDHCPServers, create: h.CreateDHCPServer, update: h.UpdateDHCPServer, del: h.DeleteDHCPServer,
	})
}

func TestDHCPLeaseCRUD_Guards(t *testing.T) {
	h := minHandler()
	testNetworkModuleCRUDGuards(t, "DHCPLease", "/dhcp-leases", networkModuleHandlers{
		list: h.ListDHCPLeases, create: h.CreateDHCPLease, update: h.UpdateDHCPLease, del: h.DeleteDHCPLease,
	})
}
