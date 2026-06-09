package handlers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

// ---- Request body types ----

type SubmitSubnetRequestBody struct {
	NetworkID          int64  `json:"network_id"`
	ParentSubnetID     *int64 `json:"parent_subnet_id"`
	RequestedPrefixLen int    `json:"requested_prefix_len"`
	Purpose            string `json:"purpose"`
}

type SubmitIPRequestBody struct {
	SubnetID    int64   `json:"subnet_id"`
	RequestedIP *string `json:"requested_ip"`
	DNSName     string  `json:"dns_name"`
	Purpose     string  `json:"purpose"`
}

type ReviewRequestBody struct {
	ReviewerNote string `json:"reviewer_note"`
}

// ---- Subnet Request endpoints ----

// SubmitSubnetRequest handles POST /api/v1/requests/subnets
func (h *Handler) SubmitSubnetRequest(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2SubnetRequestSubmit); err != nil {
		return nil
	}

	currentUser, _ := c.Locals("user").(*models.User)

	req := new(SubmitSubnetRequestBody)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.NetworkID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "network_id is required"})
	}
	if req.RequestedPrefixLen <= 0 || req.RequestedPrefixLen > 32 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "requested_prefix_len must be between 1 and 32"})
	}
	if strings.TrimSpace(req.Purpose) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "purpose is required"})
	}

	sr, err := h.service.SubmitSubnetRequest(c.Context(), currentUser.ID, req.NetworkID, req.ParentSubnetID, req.RequestedPrefixLen, req.Purpose)
	if err != nil {
		reqLogger(c).Error("error submitting subnet request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "subnet_request_submitted",
		ResourceType: "subnet_request", ResourceID: &sr.ID,
		NewValues: map[string]interface{}{"network_id": req.NetworkID, "prefix_len": req.RequestedPrefixLen, "purpose": req.Purpose},
	})

	return c.Status(fiber.StatusCreated).JSON(sr)
}

// ListMySubnetRequests handles GET /api/v1/requests/subnets
func (h *Handler) ListMySubnetRequests(c *fiber.Ctx) error {
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok || currentUser == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}

	requests, err := h.service.ListMySubnetRequests(c.Context(), currentUser.ID)
	if err != nil {
		reqLogger(c).Error("error listing subnet requests", "user_id", currentUser.ID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(requests)
}

// ListAllSubnetRequests handles GET /api/v1/admin/requests/subnets
func (h *Handler) ListAllSubnetRequests(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2SubnetRequestReview); err != nil {
		return nil
	}

	requests, err := h.service.ListAllSubnetRequests(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing all subnet requests", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(requests)
}

// ApproveSubnetRequest handles POST /api/v1/admin/requests/subnets/:id/approve
func (h *Handler) ApproveSubnetRequest(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2SubnetRequestReview); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request ID"})
	}

	reviewer, _ := c.Locals("user").(*models.User)

	req := new(ReviewRequestBody)
	_ = c.BodyParser(req) // reviewer_note is optional for approval

	sr, err := h.service.ApproveSubnetRequest(c.Context(), int64(id), reviewer.ID, req.ReviewerNote)
	if err != nil {
		reqLogger(c).Error("error approving subnet request", "id", id, "error", err)
		if errors.Is(err, services.ErrNotFound) || errors.Is(err, services.ErrNotPending) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "subnet_request_approved",
		ResourceType: "subnet_request", ResourceID: &sr.ID,
	})

	return c.JSON(sr)
}

// RejectSubnetRequest handles POST /api/v1/admin/requests/subnets/:id/reject
func (h *Handler) RejectSubnetRequest(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2SubnetRequestReview); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request ID"})
	}

	reviewer, _ := c.Locals("user").(*models.User)

	req := new(ReviewRequestBody)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if strings.TrimSpace(req.ReviewerNote) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "reviewer_note is required"})
	}

	sr, err := h.service.RejectSubnetRequest(c.Context(), int64(id), reviewer.ID, req.ReviewerNote)
	if err != nil {
		reqLogger(c).Error("error rejecting subnet request", "id", id, "error", err)
		if errors.Is(err, services.ErrNotFound) || errors.Is(err, services.ErrNotPending) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "subnet_request_rejected",
		ResourceType: "subnet_request", ResourceID: &sr.ID,
	})

	return c.JSON(sr)
}

// CancelSubnetRequest handles DELETE /api/v1/requests/subnets/:id
func (h *Handler) CancelSubnetRequest(c *fiber.Ctx) error {
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok || currentUser == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request ID"})
	}

	if err := h.service.CancelSubnetRequest(c.Context(), int64(id), currentUser.ID); err != nil {
		if errors.Is(err, services.ErrNotFound) || errors.Is(err, services.ErrNotCancellable) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		reqLogger(c).Error("error cancelling subnet request", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	rid := int64(id)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "subnet_request_cancelled",
		ResourceType: "subnet_request", ResourceID: &rid,
	})

	return c.SendStatus(fiber.StatusNoContent)
}

// ---- IP Request endpoints ----

// SubmitIPRequest handles POST /api/v1/requests/ips
func (h *Handler) SubmitIPRequest(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2SubnetRequestSubmit); err != nil {
		return nil
	}

	currentUser, _ := c.Locals("user").(*models.User)

	req := new(SubmitIPRequestBody)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.SubnetID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "subnet_id is required"})
	}
	if strings.TrimSpace(req.Purpose) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "purpose is required"})
	}

	ir, err := h.service.SubmitIPRequest(c.Context(), currentUser.ID, req.SubnetID, req.RequestedIP, req.DNSName, req.Purpose)
	if err != nil {
		reqLogger(c).Error("error submitting IP request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "ip_request_submitted",
		ResourceType: "ip_request", ResourceID: &ir.ID,
		NewValues: map[string]interface{}{"subnet_id": req.SubnetID, "purpose": req.Purpose},
	})

	return c.Status(fiber.StatusCreated).JSON(ir)
}

// ListMyIPRequests handles GET /api/v1/requests/ips
func (h *Handler) ListMyIPRequests(c *fiber.Ctx) error {
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok || currentUser == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}

	requests, err := h.service.ListMyIPRequests(c.Context(), currentUser.ID)
	if err != nil {
		reqLogger(c).Error("error listing IP requests", "user_id", currentUser.ID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(requests)
}

// ListAllIPRequests handles GET /api/v1/admin/requests/ips
func (h *Handler) ListAllIPRequests(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2SubnetRequestReview); err != nil {
		return nil
	}

	requests, err := h.service.ListAllIPRequests(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing all IP requests", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(requests)
}

// ApproveIPRequest handles POST /api/v1/admin/requests/ips/:id/approve
func (h *Handler) ApproveIPRequest(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2SubnetRequestReview); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request ID"})
	}

	reviewer, _ := c.Locals("user").(*models.User)

	req := new(ReviewRequestBody)
	_ = c.BodyParser(req) // reviewer_note optional for approval

	ir, err := h.service.ApproveIPRequest(c.Context(), int64(id), reviewer.ID, req.ReviewerNote)
	if err != nil {
		reqLogger(c).Error("error approving IP request", "id", id, "error", err)
		var takenErr *services.IPAlreadyTakenError
		if errors.As(err, &takenErr) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": takenErr.Error()})
		}
		if errors.Is(err, services.ErrNotFound) || errors.Is(err, services.ErrNotPending) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "ip_request_approved",
		ResourceType: "ip_request", ResourceID: &ir.ID,
	})

	return c.JSON(ir)
}

// RejectIPRequest handles POST /api/v1/admin/requests/ips/:id/reject
func (h *Handler) RejectIPRequest(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2SubnetRequestReview); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request ID"})
	}

	reviewer, _ := c.Locals("user").(*models.User)

	req := new(ReviewRequestBody)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if strings.TrimSpace(req.ReviewerNote) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "reviewer_note is required"})
	}

	ir, err := h.service.RejectIPRequest(c.Context(), int64(id), reviewer.ID, req.ReviewerNote)
	if err != nil {
		reqLogger(c).Error("error rejecting IP request", "id", id, "error", err)
		if errors.Is(err, services.ErrNotFound) || errors.Is(err, services.ErrNotPending) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "ip_request_rejected",
		ResourceType: "ip_request", ResourceID: &ir.ID,
	})

	return c.JSON(ir)
}

// CancelIPRequest handles DELETE /api/v1/requests/ips/:id
func (h *Handler) CancelIPRequest(c *fiber.Ctx) error {
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok || currentUser == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request ID"})
	}

	if err := h.service.CancelIPRequest(c.Context(), int64(id), currentUser.ID); err != nil {
		if errors.Is(err, services.ErrNotFound) || errors.Is(err, services.ErrNotCancellable) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		reqLogger(c).Error("error cancelling IP request", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	rid := int64(id)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "ip_request_cancelled",
		ResourceType: "ip_request", ResourceID: &rid,
	})

	return c.SendStatus(fiber.StatusNoContent)
}

// ---- Request Comments endpoints ----

// ListRequestComments handles GET /api/v1/requests/:type/:id/comments
func (h *Handler) ListRequestComments(c *fiber.Ctx) error {
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok || currentUser == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}

	reqType := c.Params("type")
	if reqType != "subnet" && reqType != "ip" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "type must be 'subnet' or 'ip'"})
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request ID"})
	}

	if currentUser.Role != "admin" {
		ownerID, err := h.service.GetRequestOwner(c.Context(), reqType, int64(id))
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "request not found"})
		}
		if ownerID != currentUser.ID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "access denied"})
		}
	}

	comments, err := h.service.ListRequestComments(c.Context(), reqType, int64(id))
	if err != nil {
		reqLogger(c).Error("error listing request comments", "type", reqType, "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(comments)
}

// AddRequestComment handles POST /api/v1/requests/:type/:id/comments
func (h *Handler) AddRequestComment(c *fiber.Ctx) error {
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok || currentUser == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}

	reqType := c.Params("type")
	if reqType != "subnet" && reqType != "ip" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "type must be 'subnet' or 'ip'"})
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request ID"})
	}

	if currentUser.Role != "admin" {
		ownerID, err := h.service.GetRequestOwner(c.Context(), reqType, int64(id))
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "request not found"})
		}
		if ownerID != currentUser.ID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "access denied"})
		}
	}

	var body struct {
		Body string `json:"body"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if strings.TrimSpace(body.Body) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "body is required"})
	}

	comment, err := h.service.AddRequestComment(c.Context(), reqType, int64(id), currentUser.ID, body.Body)
	if err != nil {
		reqLogger(c).Error("error adding request comment", "type", reqType, "id", id, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(comment)
}

// ---- Dashboard/admin endpoints ----

// GetPendingRequestCount handles GET /api/v1/admin/requests/pending-count
func (h *Handler) GetPendingRequestCount(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2SubnetRequestReview); err != nil {
		return nil
	}

	subnetCount, ipCount, err := h.service.GetPendingRequestCounts(c.Context())
	if err != nil {
		reqLogger(c).Error("error getting pending request counts", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(fiber.Map{
		"pending_subnet_requests": subnetCount,
		"pending_ip_requests":     ipCount,
		"total":                   subnetCount + ipCount,
	})
}
