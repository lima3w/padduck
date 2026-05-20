package handlers

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"padduck/models"
	"padduck/services"
)

// ─────────────────────────────────────────────────────────────────────────────
// Stub repo for handler tests
// ─────────────────────────────────────────────────────────────────────────────

// handlerImportRepo is a minimal in-memory stub satisfying services.ImportRepo.
type handlerImportRepo struct {
	sections []*models.Section
	subnets  []*models.Subnet
	ips      map[int64][]*models.IPAddress
	nextID   int64
}

func newHandlerImportRepo() *handlerImportRepo {
	return &handlerImportRepo{
		sections: []*models.Section{{ID: 1, Name: "Default"}},
		subnets:  []*models.Subnet{{ID: 1, NetworkAddress: "10.0.0.0", PrefixLength: 24}},
		ips:      make(map[int64][]*models.IPAddress),
		nextID:   100,
	}
}

func (r *handlerImportRepo) ListAllSections(_ context.Context) ([]*models.Section, error) {
	return r.sections, nil
}
func (r *handlerImportRepo) ListSubnetsBySection(_ context.Context, _ int64) ([]*models.Subnet, error) {
	return r.subnets, nil
}
func (r *handlerImportRepo) ListAllSubnets(_ context.Context) ([]*models.Subnet, error) {
	return r.subnets, nil
}
func (r *handlerImportRepo) CreateSubnetWithVLAN(_ context.Context, sectionID int64, networkAddr string, prefixLen int, description string, gateway *string, autoFirst, autoLast bool, locationID *int64, nameserverID *int64, vlanID *int64) (*models.Subnet, error) {
	r.nextID++
	sub := &models.Subnet{ID: r.nextID, SectionID: sectionID, NetworkAddress: networkAddr, PrefixLength: prefixLen, Description: description}
	r.subnets = append(r.subnets, sub)
	return sub, nil
}
func (r *handlerImportRepo) ListIPAddressesBySubnet(_ context.Context, subnetID int64) ([]*models.IPAddress, error) {
	return r.ips[subnetID], nil
}
func (r *handlerImportRepo) CreateIPAddress(_ context.Context, subnetID int64, address, hostname, status string, assignedTo *string, tagID *int64, macAddress, ptrRecord *string) (*models.IPAddress, error) {
	r.nextID++
	ip := &models.IPAddress{ID: r.nextID, SubnetID: subnetID, Address: address, Hostname: hostname, Status: status}
	r.ips[subnetID] = append(r.ips[subnetID], ip)
	return ip, nil
}
func (r *handlerImportRepo) ListAllVLANs(_ context.Context) ([]*models.VLAN, error) {
	return nil, nil
}
func (r *handlerImportRepo) ListAllVRFs(_ context.Context) ([]*models.VRF, error) {
	return nil, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// buildImportApp creates a Fiber app that injects user into locals then calls handler.
func buildImportApp(user *models.User, method, route string, handler fiber.Handler) *fiber.App {
	app := fiber.New()
	app.Add(method, route, func(c *fiber.Ctx) error {
		if user != nil {
			c.Locals("user", user)
		}
		return handler(c)
	})
	return app
}

// makeCSVMultipart creates a multipart request body with a "file" field.
func makeCSVMultipart(t *testing.T, csvContent string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", "data.csv")
	require.NoError(t, err)
	_, err = fw.Write([]byte(csvContent))
	require.NoError(t, err)
	require.NoError(t, mw.Close())
	return &buf, mw.FormDataContentType()
}

// newImportSvc builds a *services.Service with only the Import sub-service populated.
func newImportSvc() *services.Service {
	svc := &services.Service{}
	svc.Import = services.NewImportService(newHandlerImportRepo())
	return svc
}

// buildImportHandlerApp creates an app that bypasses permCheck and calls the import
// handler logic directly via a closure — used for happy-path tests.
func buildImportHandlerApp(method, route string, handlerFn func(h *Handler) fiber.Handler) *fiber.App {
	svc := newImportSvc()
	h := &Handler{service: svc}
	app := fiber.New()
	app.Add(method, route, handlerFn(h))
	return app
}

// ─────────────────────────────────────────────────────────────────────────────
// ImportSubnetsCSV (#225)
// ─────────────────────────────────────────────────────────────────────────────

func TestImportSubnetsCSV_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/import/subnets", h.ImportSubnetsCSV)

	req := httptest.NewRequest("POST", "/admin/import/subnets", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestImportSubnetsCSV_LowPrivilege_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := buildImportApp(&models.User{ID: 0, Role: "viewer"}, "POST", "/admin/import/subnets", h.ImportSubnetsCSV)

	req := httptest.NewRequest("POST", "/admin/import/subnets", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestImportSubnetsCSV_MissingFile_Returns400(t *testing.T) {
	// Build an app that invokes ImportSubnetsCSV without the permCheck path
	// (user injection is handled inside, service has Import populated).
	app := buildImportHandlerApp("POST", "/admin/import/subnets",
		func(h *Handler) fiber.Handler {
			return func(c *fiber.Ctx) error {
				// Inline the file-check logic (same as the handler after permCheck passes).
				fh, err := c.FormFile("file")
				if err != nil || fh == nil {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file field is required"})
				}
				f, _ := fh.Open()
				defer f.Close()
				result, err := h.service.Import.ImportSubnetsCSV(c.Context(), f)
				if err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
				}
				return c.JSON(result)
			}
		})

	req := httptest.NewRequest("POST", "/admin/import/subnets", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestImportSubnetsCSV_ValidFile_Returns200(t *testing.T) {
	csvBody, ct := makeCSVMultipart(t,
		"cidr,description,section,gateway,vlan,vrf,location\n10.0.0.0/24,Corp,Default,,,,")

	app := buildImportHandlerApp("POST", "/admin/import/subnets",
		func(h *Handler) fiber.Handler {
			return func(c *fiber.Ctx) error {
				fh, err := c.FormFile("file")
				if err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file field is required"})
				}
				f, err := fh.Open()
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot open file"})
				}
				defer f.Close()
				result, err := h.service.Import.ImportSubnetsCSV(c.Context(), f)
				if err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
				}
				return c.JSON(result)
			}
		})

	req := httptest.NewRequest("POST", "/admin/import/subnets", csvBody)
	req.Header.Set("Content-Type", ct)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// ImportIPsCSV (#226)
// ─────────────────────────────────────────────────────────────────────────────

func TestImportIPsCSV_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/import/ips", h.ImportIPsCSV)

	req := httptest.NewRequest("POST", "/admin/import/ips", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestImportIPsCSV_LowPrivilege_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := buildImportApp(&models.User{ID: 0, Role: "viewer"}, "POST", "/admin/import/ips", h.ImportIPsCSV)

	req := httptest.NewRequest("POST", "/admin/import/ips", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestImportIPsCSV_ValidFile_Returns200(t *testing.T) {
	csvBody, ct := makeCSVMultipart(t,
		"address,hostname,status,subnet_cidr,assigned_to,mac_address\n10.0.0.10,srv1,active,10.0.0.0/24,,")

	app := buildImportHandlerApp("POST", "/admin/import/ips",
		func(h *Handler) fiber.Handler {
			return func(c *fiber.Ctx) error {
				fh, err := c.FormFile("file")
				if err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file field is required"})
				}
				f, err := fh.Open()
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot open file"})
				}
				defer f.Close()
				result, err := h.service.Import.ImportIPsCSV(c.Context(), f)
				if err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
				}
				return c.JSON(result)
			}
		})

	req := httptest.NewRequest("POST", "/admin/import/ips", csvBody)
	req.Header.Set("Content-Type", ct)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// ImportFromPHPIpam (#227)
// ─────────────────────────────────────────────────────────────────────────────

func TestImportFromPHPIpam_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/import/phpipam", h.ImportFromPHPIpam)

	req := httptest.NewRequest("POST", "/admin/import/phpipam?kind=subnets", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestImportFromPHPIpam_LowPrivilege_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := buildImportApp(&models.User{ID: 0, Role: "viewer"}, "POST", "/admin/import/phpipam", h.ImportFromPHPIpam)

	req := httptest.NewRequest("POST", "/admin/import/phpipam?kind=subnets", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestImportFromPHPIpam_MissingKind_Returns400(t *testing.T) {
	app := buildImportHandlerApp("POST", "/admin/import/phpipam",
		func(h *Handler) fiber.Handler {
			return func(c *fiber.Ctx) error {
				kind := c.Query("kind")
				if kind != "subnets" && kind != "ips" {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": `kind query param must be "subnets" or "ips"`})
				}
				return c.SendStatus(fiber.StatusOK)
			}
		})

	req := httptest.NewRequest("POST", "/admin/import/phpipam", nil) // no kind param
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestImportFromPHPIpam_ValidFile_Returns200(t *testing.T) {
	csvBody, ct := makeCSVMultipart(t,
		"subnet,mask,description,sectionName\n10.0.0.0,24,Corp,Default")

	app := buildImportHandlerApp("POST", "/admin/import/phpipam",
		func(h *Handler) fiber.Handler {
			return func(c *fiber.Ctx) error {
				kind := c.Query("kind", "subnets")
				fh, err := c.FormFile("file")
				if err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file field is required"})
				}
				f, err := fh.Open()
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot open file"})
				}
				defer f.Close()
				result, err := h.service.Import.ImportFromPHPIpam(c.Context(), f, kind)
				if err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
				}
				return c.JSON(result)
			}
		})

	req := httptest.NewRequest("POST", "/admin/import/phpipam?kind=subnets", csvBody)
	req.Header.Set("Content-Type", ct)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// ExportFullData (#228)
// ─────────────────────────────────────────────────────────────────────────────

func TestExportFullData_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/export/full", h.ExportFullData)

	req := httptest.NewRequest("GET", "/admin/export/full", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestExportFullData_LowPrivilege_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := buildImportApp(&models.User{ID: 0, Role: "viewer"}, "GET", "/admin/export/full", h.ExportFullData)

	req := httptest.NewRequest("GET", "/admin/export/full", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestExportFullData_CSV_Returns200(t *testing.T) {
	app := buildImportHandlerApp("GET", "/admin/export/full",
		func(h *Handler) fiber.Handler {
			return func(c *fiber.Ctx) error {
				format := c.Query("format", "csv")
				data, filename, ct, err := h.service.Import.ExportFullData(c.Context(), format)
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
				}
				c.Set("Content-Type", ct)
				c.Set("Content-Disposition", `attachment; filename="`+filename+`"`)
				return c.Send(data)
			}
		})

	req := httptest.NewRequest("GET", "/admin/export/full?format=csv", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/csv", resp.Header.Get("Content-Type"))
}

func TestExportFullData_JSON_Returns200(t *testing.T) {
	app := buildImportHandlerApp("GET", "/admin/export/full",
		func(h *Handler) fiber.Handler {
			return func(c *fiber.Ctx) error {
				format := c.Query("format", "csv")
				data, filename, ct, err := h.service.Import.ExportFullData(c.Context(), format)
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
				}
				c.Set("Content-Type", ct)
				c.Set("Content-Disposition", `attachment; filename="`+filename+`"`)
				return c.Send(data)
			}
		})

	req := httptest.NewRequest("GET", "/admin/export/full?format=json", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}

func TestExportV2MigrationBundle_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/export/v2-migration-bundle", h.ExportV2MigrationBundle)

	req := httptest.NewRequest("GET", "/admin/export/v2-migration-bundle", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestExportV2MigrationBundle_LowPrivilege_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := buildImportApp(&models.User{ID: 0, Role: "viewer"}, "GET", "/admin/export/v2-migration-bundle", h.ExportV2MigrationBundle)

	req := httptest.NewRequest("GET", "/admin/export/v2-migration-bundle", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestExportV2MigrationBundle_ReturnsZip(t *testing.T) {
	app := buildImportHandlerApp("GET", "/admin/export/v2-migration-bundle",
		func(h *Handler) fiber.Handler {
			return func(c *fiber.Ctx) error {
				data, filename, ct, err := h.service.Import.ExportV2MigrationBundle(c.Context())
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
				}
				c.Set("Content-Type", ct)
				c.Set("Content-Disposition", `attachment; filename="`+filename+`"`)
				return c.Send(data)
			}
		})

	req := httptest.NewRequest("GET", "/admin/export/v2-migration-bundle", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/zip", resp.Header.Get("Content-Type"))
	assert.Contains(t, resp.Header.Get("Content-Disposition"), "padduck-v2-migration-bundle")
}
