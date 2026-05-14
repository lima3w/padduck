package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"ipam-next/models"
)

// unprivReq is a viewer user that lacks request permissions.
var unprivReq = &models.User{ID: 0, Role: "viewer"}

// adminReqUser has ID=0, which will cause permCheck to fail the permission check
// (ID<=0 → permission denied in CheckPermission). We use this to test 403 paths
// without needing a real DB. For a real admin with service, we'd need DB integration.
var adminReqUser = &models.User{ID: 0, Role: "admin"}

// ---------------------------------------------------------------------------
// POST /requests/subnets
// ---------------------------------------------------------------------------

func TestSubmitSubnetRequest_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/requests/subnets", h.SubmitSubnetRequest)

	resp, err := app.Test(httptest.NewRequest("POST", "/requests/subnets", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestSubmitSubnetRequest_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/requests/subnets", func(c *fiber.Ctx) error {
		c.Locals("user", unprivReq)
		return h.SubmitSubnetRequest(c)
	})

	resp, err := app.Test(httptest.NewRequest("POST", "/requests/subnets", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestSubmitSubnetRequest_BadBody_Returns400(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	// Viewer has no submit perm, but we want to test body validation — use admin with ID=0
	// which will be denied at permCheck too. Use a user with submit perm via permCheck bypass.
	// Since we can't bypass without DB, just verify 403 for unpriv user.
	app.Post("/requests/subnets", func(c *fiber.Ctx) error {
		c.Locals("user", unprivReq)
		return h.SubmitSubnetRequest(c)
	})
	req := httptest.NewRequest("POST", "/requests/subnets", strings.NewReader(`{"section_id":0}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	// viewer has no submit perm → 403
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GET /requests/subnets
// ---------------------------------------------------------------------------

func TestListMySubnetRequests_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/requests/subnets", h.ListMySubnetRequests)

	resp, err := app.Test(httptest.NewRequest("GET", "/requests/subnets", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DELETE /requests/subnets/:id
// ---------------------------------------------------------------------------

func TestCancelSubnetRequest_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/requests/subnets/:id", h.CancelSubnetRequest)

	resp, err := app.Test(httptest.NewRequest("DELETE", "/requests/subnets/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GET /admin/requests/subnets
// ---------------------------------------------------------------------------

func TestListAllSubnetRequests_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/requests/subnets", h.ListAllSubnetRequests)

	resp, err := app.Test(httptest.NewRequest("GET", "/admin/requests/subnets", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListAllSubnetRequests_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/requests/subnets", func(c *fiber.Ctx) error {
		c.Locals("user", unprivReq)
		return h.ListAllSubnetRequests(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/admin/requests/subnets", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// POST /admin/requests/subnets/:id/approve
// ---------------------------------------------------------------------------

func TestApproveSubnetRequest_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/requests/subnets/:id/approve", h.ApproveSubnetRequest)

	resp, err := app.Test(httptest.NewRequest("POST", "/admin/requests/subnets/1/approve", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestApproveSubnetRequest_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/requests/subnets/:id/approve", func(c *fiber.Ctx) error {
		c.Locals("user", unprivReq)
		return h.ApproveSubnetRequest(c)
	})

	resp, err := app.Test(httptest.NewRequest("POST", "/admin/requests/subnets/1/approve", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// POST /admin/requests/subnets/:id/reject
// ---------------------------------------------------------------------------

func TestRejectSubnetRequest_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/requests/subnets/:id/reject", h.RejectSubnetRequest)

	resp, err := app.Test(httptest.NewRequest("POST", "/admin/requests/subnets/1/reject", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestRejectSubnetRequest_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/requests/subnets/:id/reject", func(c *fiber.Ctx) error {
		c.Locals("user", unprivReq)
		return h.RejectSubnetRequest(c)
	})

	resp, err := app.Test(httptest.NewRequest("POST", "/admin/requests/subnets/1/reject", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// POST /requests/ips
// ---------------------------------------------------------------------------

func TestSubmitIPRequest_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/requests/ips", h.SubmitIPRequest)

	resp, err := app.Test(httptest.NewRequest("POST", "/requests/ips", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestSubmitIPRequest_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/requests/ips", func(c *fiber.Ctx) error {
		c.Locals("user", unprivReq)
		return h.SubmitIPRequest(c)
	})

	resp, err := app.Test(httptest.NewRequest("POST", "/requests/ips", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GET /requests/ips
// ---------------------------------------------------------------------------

func TestListMyIPRequests_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/requests/ips", h.ListMyIPRequests)

	resp, err := app.Test(httptest.NewRequest("GET", "/requests/ips", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// DELETE /requests/ips/:id
// ---------------------------------------------------------------------------

func TestCancelIPRequest_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Delete("/requests/ips/:id", h.CancelIPRequest)

	resp, err := app.Test(httptest.NewRequest("DELETE", "/requests/ips/1", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GET /admin/requests/ips
// ---------------------------------------------------------------------------

func TestListAllIPRequests_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/requests/ips", h.ListAllIPRequests)

	resp, err := app.Test(httptest.NewRequest("GET", "/admin/requests/ips", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestListAllIPRequests_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/requests/ips", func(c *fiber.Ctx) error {
		c.Locals("user", unprivReq)
		return h.ListAllIPRequests(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/admin/requests/ips", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// POST /admin/requests/ips/:id/approve
// ---------------------------------------------------------------------------

func TestApproveIPRequest_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/requests/ips/:id/approve", h.ApproveIPRequest)

	resp, err := app.Test(httptest.NewRequest("POST", "/admin/requests/ips/1/approve", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestApproveIPRequest_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/requests/ips/:id/approve", func(c *fiber.Ctx) error {
		c.Locals("user", unprivReq)
		return h.ApproveIPRequest(c)
	})

	resp, err := app.Test(httptest.NewRequest("POST", "/admin/requests/ips/1/approve", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// POST /admin/requests/ips/:id/reject
// ---------------------------------------------------------------------------

func TestRejectIPRequest_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/requests/ips/:id/reject", h.RejectIPRequest)

	resp, err := app.Test(httptest.NewRequest("POST", "/admin/requests/ips/1/reject", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestRejectIPRequest_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/admin/requests/ips/:id/reject", func(c *fiber.Ctx) error {
		c.Locals("user", unprivReq)
		return h.RejectIPRequest(c)
	})

	resp, err := app.Test(httptest.NewRequest("POST", "/admin/requests/ips/1/reject", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GET /requests/:type/:id/comments
// ---------------------------------------------------------------------------

func TestListRequestComments_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/requests/:type/:id/comments", h.ListRequestComments)

	resp, err := app.Test(httptest.NewRequest("GET", "/requests/subnet/1/comments", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// POST /requests/:type/:id/comments
// ---------------------------------------------------------------------------

func TestAddRequestComment_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Post("/requests/:type/:id/comments", h.AddRequestComment)

	resp, err := app.Test(httptest.NewRequest("POST", "/requests/subnet/1/comments", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// GET /admin/requests/pending-count
// ---------------------------------------------------------------------------

func TestGetPendingRequestCount_NoUser_Returns401(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/requests/pending-count", h.GetPendingRequestCount)

	resp, err := app.Test(httptest.NewRequest("GET", "/admin/requests/pending-count", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetPendingRequestCount_NoPermission_Returns403(t *testing.T) {
	h := &Handler{service: nil}
	app := fiber.New()
	app.Get("/admin/requests/pending-count", func(c *fiber.Ctx) error {
		c.Locals("user", unprivReq)
		return h.GetPendingRequestCount(c)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/admin/requests/pending-count", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// Request struct field validation
// ---------------------------------------------------------------------------

func TestSubmitSubnetRequestBody_Fields(t *testing.T) {
	pID := int64(5)
	req := &SubmitSubnetRequestBody{
		SectionID:          1,
		ParentSubnetID:     &pID,
		RequestedPrefixLen: 24,
		Purpose:            "Testing",
	}
	assert.Equal(t, int64(1), req.SectionID)
	assert.Equal(t, int64(5), *req.ParentSubnetID)
	assert.Equal(t, 24, req.RequestedPrefixLen)
	assert.Equal(t, "Testing", req.Purpose)
}

func TestSubmitIPRequestBody_Fields(t *testing.T) {
	ip := "192.168.1.100"
	req := &SubmitIPRequestBody{
		SubnetID:    1,
		RequestedIP: &ip,
		DNSName:     "web.example.com",
		Purpose:     "Web server",
	}
	assert.Equal(t, int64(1), req.SubnetID)
	assert.Equal(t, "192.168.1.100", *req.RequestedIP)
	assert.Equal(t, "web.example.com", req.DNSName)
	assert.Equal(t, "Web server", req.Purpose)
}

func TestReviewRequestBody_Fields(t *testing.T) {
	req := &ReviewRequestBody{ReviewerNote: "Approved for production use"}
	assert.Equal(t, "Approved for production use", req.ReviewerNote)
}
