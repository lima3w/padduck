package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"ipam-next/models"
)

// permUser returns a user with ID=0 so CheckPermission returns "permission denied"
// without touching the nil service repository — gives a clean 403.
func permUser() *models.User { return &models.User{ID: 0, Role: "user"} }

func deviceApp(h *Handler, method, path string, handler fiber.Handler) *fiber.App {
	app := fiber.New()
	switch method {
	case "GET":
		app.Get(path, handler)
	case "POST":
		app.Post(path, handler)
	case "PUT":
		app.Put(path, handler)
	case "DELETE":
		app.Delete(path, handler)
	}
	return app
}

func deviceAppAs(h *Handler, method, path string, handler fiber.Handler, u *models.User) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", u)
		return c.Next()
	})
	switch method {
	case "GET":
		app.Get(path, handler)
	case "POST":
		app.Post(path, handler)
	case "PUT":
		app.Put(path, handler)
	case "DELETE":
		app.Delete(path, handler)
	}
	return app
}

// ---------------------------------------------------------------------------
// ListDeviceTypes — GET /api/v1/device-types
// ---------------------------------------------------------------------------

func TestListDeviceTypes_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "GET", "/device-types", h.ListDeviceTypes)
	resp, err := app.Test(httptest.NewRequest("GET", "/device-types", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListDeviceTypes_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "GET", "/device-types", h.ListDeviceTypes, permUser())
	resp, err := app.Test(httptest.NewRequest("GET", "/device-types", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ListDevices — GET /api/v1/devices
// ---------------------------------------------------------------------------

func TestListDevices_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "GET", "/devices", h.ListDevices)
	resp, err := app.Test(httptest.NewRequest("GET", "/devices", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListDevices_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "GET", "/devices", h.ListDevices, permUser())
	resp, err := app.Test(httptest.NewRequest("GET", "/devices", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CreateDevice — POST /api/v1/devices
// ---------------------------------------------------------------------------

func TestCreateDevice_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "POST", "/devices", h.CreateDevice)
	req := httptest.NewRequest("POST", "/devices", strings.NewReader(`{"hostname":"router1"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateDevice_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "POST", "/devices", h.CreateDevice, permUser())
	req := httptest.NewRequest("POST", "/devices", strings.NewReader(`{"hostname":"router1"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetDevice — GET /api/v1/devices/:id
// ---------------------------------------------------------------------------

func TestGetDevice_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "GET", "/devices/:id", h.GetDevice)
	resp, err := app.Test(httptest.NewRequest("GET", "/devices/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetDevice_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "GET", "/devices/:id", h.GetDevice, permUser())
	resp, err := app.Test(httptest.NewRequest("GET", "/devices/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdateDevice — PUT /api/v1/devices/:id
// ---------------------------------------------------------------------------

func TestUpdateDevice_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "PUT", "/devices/:id", h.UpdateDevice)
	req := httptest.NewRequest("PUT", "/devices/1", strings.NewReader(`{"hostname":"router1"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateDevice_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "PUT", "/devices/:id", h.UpdateDevice, permUser())
	req := httptest.NewRequest("PUT", "/devices/1", strings.NewReader(`{"hostname":"router1"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DeleteDevice — DELETE /api/v1/devices/:id
// ---------------------------------------------------------------------------

func TestDeleteDevice_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "DELETE", "/devices/:id", h.DeleteDevice)
	resp, err := app.Test(httptest.NewRequest("DELETE", "/devices/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteDevice_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "DELETE", "/devices/:id", h.DeleteDevice, permUser())
	resp, err := app.Test(httptest.NewRequest("DELETE", "/devices/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GetDeviceSNMPCredentials — GET /api/v1/devices/:id/snmp-credentials
// ---------------------------------------------------------------------------

func TestGetDeviceSNMPCredentials_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "GET", "/devices/:id/snmp-credentials", h.GetDeviceSNMPCredentials)
	resp, err := app.Test(httptest.NewRequest("GET", "/devices/1/snmp-credentials", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetDeviceSNMPCredentials_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "GET", "/devices/:id/snmp-credentials", h.GetDeviceSNMPCredentials, permUser())
	resp, err := app.Test(httptest.NewRequest("GET", "/devices/1/snmp-credentials", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ListDeviceIPAddresses — GET /api/v1/devices/:id/ip-addresses
// ---------------------------------------------------------------------------

func TestListDeviceIPAddresses_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "GET", "/devices/:id/ip-addresses", h.ListDeviceIPAddresses)
	resp, err := app.Test(httptest.NewRequest("GET", "/devices/1/ip-addresses", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListDeviceIPAddresses_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "GET", "/devices/:id/ip-addresses", h.ListDeviceIPAddresses, permUser())
	resp, err := app.Test(httptest.NewRequest("GET", "/devices/1/ip-addresses", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// SearchDevices — POST /api/v1/devices/search
// ---------------------------------------------------------------------------

func TestSearchDevices_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "POST", "/devices/search", h.SearchDevices)
	req := httptest.NewRequest("POST", "/devices/search", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestSearchDevices_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "POST", "/devices/search", h.SearchDevices, permUser())
	req := httptest.NewRequest("POST", "/devices/search", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ListDeviceInterfaces — GET /api/v1/devices/:id/interfaces
// ---------------------------------------------------------------------------

func TestListDeviceInterfaces_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "GET", "/devices/:id/interfaces", h.ListDeviceInterfaces)
	resp, err := app.Test(httptest.NewRequest("GET", "/devices/1/interfaces", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListDeviceInterfaces_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "GET", "/devices/:id/interfaces", h.ListDeviceInterfaces, permUser())
	resp, err := app.Test(httptest.NewRequest("GET", "/devices/1/interfaces", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// CreateDeviceInterface — POST /api/v1/devices/:id/interfaces
// ---------------------------------------------------------------------------

func TestCreateDeviceInterface_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "POST", "/devices/:id/interfaces", h.CreateDeviceInterface)
	req := httptest.NewRequest("POST", "/devices/1/interfaces", strings.NewReader(`{"name":"eth0"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateDeviceInterface_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "POST", "/devices/:id/interfaces", h.CreateDeviceInterface, permUser())
	req := httptest.NewRequest("POST", "/devices/1/interfaces", strings.NewReader(`{"name":"eth0"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UpdateDeviceInterface — PUT /api/v1/devices/:id/interfaces/:if_id
// ---------------------------------------------------------------------------

func TestUpdateDeviceInterface_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "PUT", "/devices/:id/interfaces/:if_id", h.UpdateDeviceInterface)
	req := httptest.NewRequest("PUT", "/devices/1/interfaces/2", strings.NewReader(`{"name":"eth0"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateDeviceInterface_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "PUT", "/devices/:id/interfaces/:if_id", h.UpdateDeviceInterface, permUser())
	req := httptest.NewRequest("PUT", "/devices/1/interfaces/2", strings.NewReader(`{"name":"eth0"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DeleteDeviceInterface — DELETE /api/v1/devices/:id/interfaces/:if_id
// ---------------------------------------------------------------------------

func TestDeleteDeviceInterface_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "DELETE", "/devices/:id/interfaces/:if_id", h.DeleteDeviceInterface)
	resp, err := app.Test(httptest.NewRequest("DELETE", "/devices/1/interfaces/2", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteDeviceInterface_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "DELETE", "/devices/:id/interfaces/:if_id", h.DeleteDeviceInterface, permUser())
	resp, err := app.Test(httptest.NewRequest("DELETE", "/devices/1/interfaces/2", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// AssociateIPToDevice — POST /api/v1/devices/:id/ip-addresses/:ip_id/associate
// ---------------------------------------------------------------------------

func TestAssociateIPToDevice_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "POST", "/devices/:id/ip-addresses/:ip_id/associate", h.AssociateIPToDevice)
	req := httptest.NewRequest("POST", "/devices/1/ip-addresses/2/associate", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAssociateIPToDevice_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "POST", "/devices/:id/ip-addresses/:ip_id/associate", h.AssociateIPToDevice, permUser())
	req := httptest.NewRequest("POST", "/devices/1/ip-addresses/2/associate", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UnlinkIPFromDevice — DELETE /api/v1/devices/:id/ip-addresses/:ip_id
// ---------------------------------------------------------------------------

func TestUnlinkIPFromDevice_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := deviceApp(h, "DELETE", "/devices/:id/ip-addresses/:ip_id", h.UnlinkIPFromDevice)
	resp, err := app.Test(httptest.NewRequest("DELETE", "/devices/1/ip-addresses/2", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUnlinkIPFromDevice_NoPermission_Returns403(t *testing.T) {
	h := &Handler{}
	app := deviceAppAs(h, "DELETE", "/devices/:id/ip-addresses/:ip_id", h.UnlinkIPFromDevice, permUser())
	resp, err := app.Test(httptest.NewRequest("DELETE", "/devices/1/ip-addresses/2", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
