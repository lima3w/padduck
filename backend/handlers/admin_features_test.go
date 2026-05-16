package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// adminOnly wires a handler that requires admin, testing no-user and non-admin cases.
func adminOnlyApp(h *Handler, method, path string, handler fiber.Handler) *fiber.App {
	app := fiber.New()
	switch method {
	case "GET":
		app.Get(path, handler)
	case "POST":
		app.Post(path, handler)
	case "DELETE":
		app.Delete(path, handler)
	}
	return app
}

func adminOnlyAppAs(h *Handler, method, path string, handler fiber.Handler, u interface{}) *fiber.App {
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
	case "DELETE":
		app.Delete(path, handler)
	}
	return app
}

func jsonReq(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// ---------------------------------------------------------------------------
// SuspendUser — POST /admin/users/:id/suspend
// ---------------------------------------------------------------------------

func TestSuspendUser_NoUser_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/admin/users/:id/suspend", h.SuspendUser)
	resp, err := app.Test(jsonReq("POST", "/admin/users/1/suspend", `{"reason":"test"}`))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestSuspendUser_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{}
	app := adminOnlyAppAs(h, "POST", "/admin/users/:id/suspend", h.SuspendUser, nonAdminUser)
	resp, err := app.Test(jsonReq("POST", "/admin/users/1/suspend", `{"reason":"test"}`))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestSuspendUser_BadID_Returns400(t *testing.T) {
	h := &Handler{}
	app := adminOnlyAppAs(h, "POST", "/admin/users/:id/suspend", h.SuspendUser, adminUser)
	resp, err := app.Test(jsonReq("POST", "/admin/users/notanumber/suspend", `{"reason":"test"}`))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestSuspendUser_MissingReason_Returns400(t *testing.T) {
	h := &Handler{}
	app := adminOnlyAppAs(h, "POST", "/admin/users/:id/suspend", h.SuspendUser, adminUser)
	resp, err := app.Test(jsonReq("POST", "/admin/users/1/suspend", `{}`))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// UnsuspendUser — POST /admin/users/:id/unsuspend
// ---------------------------------------------------------------------------

func TestUnsuspendUser_NoUser_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/admin/users/:id/unsuspend", h.UnsuspendUser)
	resp, err := app.Test(jsonReq("POST", "/admin/users/1/unsuspend", ``))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUnsuspendUser_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{}
	app := adminOnlyAppAs(h, "POST", "/admin/users/:id/unsuspend", h.UnsuspendUser, nonAdminUser)
	resp, err := app.Test(jsonReq("POST", "/admin/users/1/unsuspend", ``))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestUnsuspendUser_BadID_Returns400(t *testing.T) {
	h := &Handler{}
	app := adminOnlyAppAs(h, "POST", "/admin/users/:id/unsuspend", h.UnsuspendUser, adminUser)
	resp, err := app.Test(jsonReq("POST", "/admin/users/notanumber/unsuspend", ``))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ImpersonateUser — POST /admin/users/:id/impersonate
// ---------------------------------------------------------------------------

func TestImpersonateUser_NoUser_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/admin/users/:id/impersonate", h.ImpersonateUser)
	resp, err := app.Test(jsonReq("POST", "/admin/users/1/impersonate", ``))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestImpersonateUser_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{}
	app := adminOnlyAppAs(h, "POST", "/admin/users/:id/impersonate", h.ImpersonateUser, nonAdminUser)
	resp, err := app.Test(jsonReq("POST", "/admin/users/1/impersonate", ``))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestImpersonateUser_BadID_Returns400(t *testing.T) {
	h := &Handler{}
	app := adminOnlyAppAs(h, "POST", "/admin/users/:id/impersonate", h.ImpersonateUser, adminUser)
	resp, err := app.Test(jsonReq("POST", "/admin/users/notanumber/impersonate", ``))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// BulkSuspendUsers — POST /admin/users/bulk-suspend
// ---------------------------------------------------------------------------

func TestBulkSuspendUsers_NoUser_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/admin/users/bulk-suspend", h.BulkSuspendUsers)
	resp, err := app.Test(jsonReq("POST", "/admin/users/bulk-suspend", `{"user_ids":[1]}`))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestBulkSuspendUsers_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{}
	app := adminOnlyAppAs(h, "POST", "/admin/users/bulk-suspend", h.BulkSuspendUsers, nonAdminUser)
	resp, err := app.Test(jsonReq("POST", "/admin/users/bulk-suspend", `{"user_ids":[1]}`))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestBulkSuspendUsers_EmptyIDs_Returns400(t *testing.T) {
	h := &Handler{}
	app := adminOnlyAppAs(h, "POST", "/admin/users/bulk-suspend", h.BulkSuspendUsers, adminUser)
	resp, err := app.Test(jsonReq("POST", "/admin/users/bulk-suspend", `{"user_ids":[]}`))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// BulkActivateUsers — POST /admin/users/bulk-activate
// ---------------------------------------------------------------------------

func TestBulkActivateUsers_NoUser_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/admin/users/bulk-activate", h.BulkActivateUsers)
	resp, err := app.Test(jsonReq("POST", "/admin/users/bulk-activate", `{"user_ids":[1]}`))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestBulkActivateUsers_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{}
	app := adminOnlyAppAs(h, "POST", "/admin/users/bulk-activate", h.BulkActivateUsers, nonAdminUser)
	resp, err := app.Test(jsonReq("POST", "/admin/users/bulk-activate", `{"user_ids":[1]}`))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestBulkActivateUsers_EmptyIDs_Returns400(t *testing.T) {
	h := &Handler{}
	app := adminOnlyAppAs(h, "POST", "/admin/users/bulk-activate", h.BulkActivateUsers, adminUser)
	resp, err := app.Test(jsonReq("POST", "/admin/users/bulk-activate", `{}`))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// BulkDeleteUsers — POST /admin/users/bulk-delete
// ---------------------------------------------------------------------------

func TestBulkDeleteUsers_NoUser_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/admin/users/bulk-delete", h.BulkDeleteUsers)
	resp, err := app.Test(jsonReq("POST", "/admin/users/bulk-delete", `{"user_ids":[1]}`))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestBulkDeleteUsers_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{}
	app := adminOnlyAppAs(h, "POST", "/admin/users/bulk-delete", h.BulkDeleteUsers, nonAdminUser)
	resp, err := app.Test(jsonReq("POST", "/admin/users/bulk-delete", `{"user_ids":[1]}`))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestBulkDeleteUsers_EmptyIDs_Returns400(t *testing.T) {
	h := &Handler{}
	app := adminOnlyAppAs(h, "POST", "/admin/users/bulk-delete", h.BulkDeleteUsers, adminUser)
	resp, err := app.Test(jsonReq("POST", "/admin/users/bulk-delete", `{"user_ids":[]}`))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// BulkImportUsers — POST /admin/users/bulk-import (multipart)
// ---------------------------------------------------------------------------

func TestBulkImportUsers_NoUser_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/admin/users/bulk-import", h.BulkImportUsers)
	resp, err := app.Test(httptest.NewRequest("POST", "/admin/users/bulk-import", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestBulkImportUsers_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{}
	app := adminOnlyAppAs(h, "POST", "/admin/users/bulk-import", h.BulkImportUsers, nonAdminUser)
	resp, err := app.Test(httptest.NewRequest("POST", "/admin/users/bulk-import", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestBulkImportUsers_NoFile_Returns400(t *testing.T) {
	h := &Handler{}
	app := adminOnlyAppAs(h, "POST", "/admin/users/bulk-import", h.BulkImportUsers, adminUser)
	resp, err := app.Test(httptest.NewRequest("POST", "/admin/users/bulk-import", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestBulkImportUsers_CSVMissingColumns_Returns400(t *testing.T) {
	h := &Handler{}
	app := adminOnlyAppAs(h, "POST", "/admin/users/bulk-import", h.BulkImportUsers, adminUser)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, _ := w.CreateFormFile("file", "users.csv")
	_, _ = fw.Write([]byte("role\nadmin\n")) // no username or email columns
	w.Close()

	req := httptest.NewRequest("POST", "/admin/users/bulk-import", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// ExportMyData — GET /auth/me/export
// ---------------------------------------------------------------------------

func TestExportMyData_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := simpleApp(h, "GET", "/auth/me/export", h.ExportMyData, nil)
	resp, err := app.Test(httptest.NewRequest("GET", "/auth/me/export", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// RequestDeletion — POST /auth/me/deletion-request
// ---------------------------------------------------------------------------

func TestRequestDeletion_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := simpleApp(h, "POST", "/auth/me/deletion-request", h.RequestDeletion, nil)
	resp, err := app.Test(httptest.NewRequest("POST", "/auth/me/deletion-request", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GDPRDeleteUser — POST /admin/users/:id/gdpr-delete
// ---------------------------------------------------------------------------

func TestGDPRDeleteUser_NoUser_Returns403(t *testing.T) {
	h := &Handler{}
	app := fiber.New()
	app.Post("/admin/users/:id/gdpr-delete", h.GDPRDeleteUser)
	resp, err := app.Test(jsonReq("POST", "/admin/users/1/gdpr-delete", ``))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGDPRDeleteUser_NonAdmin_Returns403(t *testing.T) {
	h := &Handler{}
	app := adminOnlyAppAs(h, "POST", "/admin/users/:id/gdpr-delete", h.GDPRDeleteUser, nonAdminUser)
	resp, err := app.Test(jsonReq("POST", "/admin/users/1/gdpr-delete", ``))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestGDPRDeleteUser_BadID_Returns400(t *testing.T) {
	h := &Handler{}
	app := adminOnlyAppAs(h, "POST", "/admin/users/:id/gdpr-delete", h.GDPRDeleteUser, adminUser)
	resp, err := app.Test(jsonReq("POST", "/admin/users/notanumber/gdpr-delete", ``))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// AcceptPrivacyPolicy — POST /auth/me/accept-privacy
// ---------------------------------------------------------------------------

func TestAcceptPrivacyPolicy_NoUser_Returns401(t *testing.T) {
	h := &Handler{}
	app := simpleApp(h, "POST", "/auth/me/accept-privacy", h.AcceptPrivacyPolicy, nil)
	resp, err := app.Test(httptest.NewRequest("POST", "/auth/me/accept-privacy", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
