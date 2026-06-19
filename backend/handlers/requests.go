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
	if !h.requirePerm(c, services.PermV2SubnetRequestSubmit) {
		return nil
	}

	currentUser, _ := c.Locals("user").(*models.User)

	req := new(SubmitSubnetRequestBody)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if req.NetworkID <= 0 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "network_id is required")
	}
	if req.RequestedPrefixLen <= 0 || req.RequestedPrefixLen > 32 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "requested_prefix_len must be between 1 and 32")
	}
	if strings.TrimSpace(req.Purpose) == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "purpose is required")
	}

	sr, err := h.ops.Workflow.SubmitSubnetRequest(c.Context(), currentUser.ID, req.NetworkID, req.ParentSubnetID, req.RequestedPrefixLen, req.Purpose)
	if err != nil {
		reqLogger(c).Error("error submitting subnet request", "error", err)
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
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
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "not authenticated")
	}

	requests, err := h.ops.Workflow.ListMySubnetRequests(c.Context(), currentUser.ID)
	if err != nil {
		reqLogger(c).Error("error listing subnet requests", "user_id", currentUser.ID, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(requests)
}

// ListAllSubnetRequests handles GET /api/v1/admin/requests/subnets
func (h *Handler) ListAllSubnetRequests(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2SubnetRequestReview) {
		return nil
	}

	requests, err := h.ops.Workflow.ListAllSubnetRequests(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing all subnet requests", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(requests)
}

// ApproveSubnetRequest handles POST /api/v1/admin/requests/subnets/:id/approve
func (h *Handler) ApproveSubnetRequest(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2SubnetRequestReview) {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request ID")
	}

	reviewer, _ := c.Locals("user").(*models.User)

	req := new(ReviewRequestBody)
	_ = c.BodyParser(req) // reviewer_note is optional for approval

	sr, err := h.ops.Workflow.ApproveSubnetRequest(c.Context(), int64(id), reviewer.ID, req.ReviewerNote)
	if err != nil {
		reqLogger(c).Error("error approving subnet request", "id", id, "error", err)
		if errors.Is(err, services.ErrNotFound) || errors.Is(err, services.ErrNotPending) {
			return RespondError(c, fiber.StatusNotFound, ErrNotFound, err.Error())
		}
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
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
	if !h.requirePerm(c, services.PermV2SubnetRequestReview) {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request ID")
	}

	reviewer, _ := c.Locals("user").(*models.User)

	req := new(ReviewRequestBody)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if strings.TrimSpace(req.ReviewerNote) == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "reviewer_note is required")
	}

	sr, err := h.ops.Workflow.RejectSubnetRequest(c.Context(), int64(id), reviewer.ID, req.ReviewerNote)
	if err != nil {
		reqLogger(c).Error("error rejecting subnet request", "id", id, "error", err)
		if errors.Is(err, services.ErrNotFound) || errors.Is(err, services.ErrNotPending) {
			return RespondError(c, fiber.StatusNotFound, ErrNotFound, err.Error())
		}
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
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
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "not authenticated")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request ID")
	}

	if err := h.ops.Workflow.CancelSubnetRequest(c.Context(), int64(id), currentUser.ID); err != nil {
		if errors.Is(err, services.ErrNotFound) || errors.Is(err, services.ErrNotCancellable) {
			return RespondError(c, fiber.StatusNotFound, ErrNotFound, err.Error())
		}
		reqLogger(c).Error("error cancelling subnet request", "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
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
	if !h.requirePerm(c, services.PermV2SubnetRequestSubmit) {
		return nil
	}

	currentUser, _ := c.Locals("user").(*models.User)

	req := new(SubmitIPRequestBody)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if req.SubnetID <= 0 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "subnet_id is required")
	}
	if strings.TrimSpace(req.Purpose) == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "purpose is required")
	}

	ir, err := h.ops.Workflow.SubmitIPRequest(c.Context(), currentUser.ID, req.SubnetID, req.RequestedIP, req.DNSName, req.Purpose)
	if err != nil {
		reqLogger(c).Error("error submitting IP request", "error", err)
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
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
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "not authenticated")
	}

	requests, err := h.ops.Workflow.ListMyIPRequests(c.Context(), currentUser.ID)
	if err != nil {
		reqLogger(c).Error("error listing IP requests", "user_id", currentUser.ID, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(requests)
}

// ListAllIPRequests handles GET /api/v1/admin/requests/ips
func (h *Handler) ListAllIPRequests(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2SubnetRequestReview) {
		return nil
	}

	requests, err := h.ops.Workflow.ListAllIPRequests(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing all IP requests", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(requests)
}

// ApproveIPRequest handles POST /api/v1/admin/requests/ips/:id/approve
func (h *Handler) ApproveIPRequest(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2SubnetRequestReview) {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request ID")
	}

	reviewer, _ := c.Locals("user").(*models.User)

	req := new(ReviewRequestBody)
	_ = c.BodyParser(req) // reviewer_note optional for approval

	ir, err := h.ops.Workflow.ApproveIPRequest(c.Context(), int64(id), reviewer.ID, req.ReviewerNote)
	if err != nil {
		reqLogger(c).Error("error approving IP request", "id", id, "error", err)
		var takenErr *services.IPAlreadyTakenError
		if errors.As(err, &takenErr) {
			return RespondError(c, fiber.StatusConflict, ErrConflict, takenErr.Error())
		}
		if errors.Is(err, services.ErrNotFound) || errors.Is(err, services.ErrNotPending) {
			return RespondError(c, fiber.StatusNotFound, ErrNotFound, err.Error())
		}
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
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
	if !h.requirePerm(c, services.PermV2SubnetRequestReview) {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request ID")
	}

	reviewer, _ := c.Locals("user").(*models.User)

	req := new(ReviewRequestBody)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if strings.TrimSpace(req.ReviewerNote) == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "reviewer_note is required")
	}

	ir, err := h.ops.Workflow.RejectIPRequest(c.Context(), int64(id), reviewer.ID, req.ReviewerNote)
	if err != nil {
		reqLogger(c).Error("error rejecting IP request", "id", id, "error", err)
		if errors.Is(err, services.ErrNotFound) || errors.Is(err, services.ErrNotPending) {
			return RespondError(c, fiber.StatusNotFound, ErrNotFound, err.Error())
		}
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
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
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "not authenticated")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request ID")
	}

	if err := h.ops.Workflow.CancelIPRequest(c.Context(), int64(id), currentUser.ID); err != nil {
		if errors.Is(err, services.ErrNotFound) || errors.Is(err, services.ErrNotCancellable) {
			return RespondError(c, fiber.StatusNotFound, ErrNotFound, err.Error())
		}
		reqLogger(c).Error("error cancelling IP request", "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
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
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "not authenticated")
	}

	reqType := c.Params("type")
	if reqType != "subnet" && reqType != "ip" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "type must be 'subnet' or 'ip'")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request ID")
	}

	if currentUser.Role != "admin" {
		ownerID, err := h.ops.Workflow.GetRequestOwner(c.Context(), reqType, int64(id))
		if err != nil {
			return RespondError(c, fiber.StatusNotFound, ErrNotFound, "request not found")
		}
		if ownerID != currentUser.ID {
			return RespondError(c, fiber.StatusForbidden, ErrForbidden, "access denied")
		}
	}

	comments, err := h.ops.Workflow.ListRequestComments(c.Context(), reqType, int64(id))
	if err != nil {
		reqLogger(c).Error("error listing request comments", "type", reqType, "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(comments)
}

// AddRequestComment handles POST /api/v1/requests/:type/:id/comments
func (h *Handler) AddRequestComment(c *fiber.Ctx) error {
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok || currentUser == nil {
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "not authenticated")
	}

	reqType := c.Params("type")
	if reqType != "subnet" && reqType != "ip" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "type must be 'subnet' or 'ip'")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request ID")
	}

	if currentUser.Role != "admin" {
		ownerID, err := h.ops.Workflow.GetRequestOwner(c.Context(), reqType, int64(id))
		if err != nil {
			return RespondError(c, fiber.StatusNotFound, ErrNotFound, "request not found")
		}
		if ownerID != currentUser.ID {
			return RespondError(c, fiber.StatusForbidden, ErrForbidden, "access denied")
		}
	}

	var body struct {
		Body string `json:"body"`
	}
	if err := c.BodyParser(&body); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if strings.TrimSpace(body.Body) == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "body is required")
	}

	comment, err := h.ops.Workflow.AddRequestComment(c.Context(), reqType, int64(id), currentUser.ID, body.Body)
	if err != nil {
		reqLogger(c).Error("error adding request comment", "type", reqType, "id", id, "error", err)
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(comment)
}

// ---- Dashboard/admin endpoints ----

// GetPendingRequestCount handles GET /api/v1/admin/requests/pending-count
func (h *Handler) GetPendingRequestCount(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2SubnetRequestReview) {
		return nil
	}

	subnetCount, ipCount, err := h.ops.Workflow.GetPendingRequestCounts(c.Context())
	if err != nil {
		reqLogger(c).Error("error getting pending request counts", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	return c.JSON(fiber.Map{
		"pending_subnet_requests": subnetCount,
		"pending_ip_requests":     ipCount,
		"total":                   subnetCount + ipCount,
	})
}
