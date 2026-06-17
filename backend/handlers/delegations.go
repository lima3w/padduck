package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

// createDelegationRequest is the request body for creating an IPv6 delegation
type createDelegationRequest struct {
	DelegatedPrefix        string     `json:"delegated_prefix"`
	DelegatedToDeviceID    *int64     `json:"delegated_to_device_id"`
	DelegatedToDescription *string    `json:"delegated_to_description"`
	ValidLifetimeSec       *int       `json:"valid_lifetime_sec"`
	PreferredLifetimeSec   *int       `json:"preferred_lifetime_sec"`
	ExpiresAt              *time.Time `json:"expires_at"`
}

// updateDelegationRequest is the request body for updating an IPv6 delegation
type updateDelegationRequest struct {
	DelegatedPrefix        string     `json:"delegated_prefix"`
	DelegatedToDeviceID    *int64     `json:"delegated_to_device_id"`
	DelegatedToDescription *string    `json:"delegated_to_description"`
	ValidLifetimeSec       *int       `json:"valid_lifetime_sec"`
	PreferredLifetimeSec   *int       `json:"preferred_lifetime_sec"`
	ExpiresAt              *time.Time `json:"expires_at"`
}

// ListDelegations handles GET /api/v1/subnets/:id/delegations
func (h *Handler) ListDelegations(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2SubnetRead) {
		return nil
	}

	subnetID, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid subnet ID")
	}

	delegations, err := h.service.ListDelegations(c.Context(), int64(subnetID))
	if err != nil {
		reqLogger(c).Error("error listing delegations", "subnet_id", subnetID, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	if delegations == nil {
		delegations = make([]*models.IPv6Delegation, 0)
	}
	return c.JSON(delegations)
}

// CreateDelegation handles POST /api/v1/subnets/:id/delegations
func (h *Handler) CreateDelegation(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}

	subnetID, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid subnet ID")
	}

	req := new(createDelegationRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if req.DelegatedPrefix == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "delegated_prefix is required")
	}

	d := &models.IPv6Delegation{
		ParentSubnetID:         int64(subnetID),
		DelegatedPrefix:        req.DelegatedPrefix,
		DelegatedToDeviceID:    req.DelegatedToDeviceID,
		DelegatedToDescription: req.DelegatedToDescription,
		ValidLifetimeSec:       req.ValidLifetimeSec,
		PreferredLifetimeSec:   req.PreferredLifetimeSec,
		ExpiresAt:              req.ExpiresAt,
	}

	result, err := h.service.CreateDelegation(c.Context(), d)
	if err != nil {
		reqLogger(c).Error("error creating delegation", "error", err)
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "ipv6_delegation.created",
		ResourceType: "ipv6_delegation", ResourceID: &result.ID,
		ResourceName: result.DelegatedPrefix,
	})

	return c.Status(fiber.StatusCreated).JSON(result)
}

// UpdateDelegation handles PUT /api/v1/delegations/:id
func (h *Handler) UpdateDelegation(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid delegation ID")
	}

	req := new(updateDelegationRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if req.DelegatedPrefix == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "delegated_prefix is required")
	}

	d := &models.IPv6Delegation{
		DelegatedPrefix:        req.DelegatedPrefix,
		DelegatedToDeviceID:    req.DelegatedToDeviceID,
		DelegatedToDescription: req.DelegatedToDescription,
		ValidLifetimeSec:       req.ValidLifetimeSec,
		PreferredLifetimeSec:   req.PreferredLifetimeSec,
		ExpiresAt:              req.ExpiresAt,
	}

	result, err := h.service.UpdateDelegation(c.Context(), int64(id), d)
	if err != nil {
		reqLogger(c).Error("error updating delegation", "id", id, "error", err)
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "ipv6_delegation.updated",
		ResourceType: "ipv6_delegation", ResourceID: &result.ID,
		ResourceName: result.DelegatedPrefix,
	})

	return c.JSON(result)
}

// DeleteDelegation handles DELETE /api/v1/delegations/:id
func (h *Handler) DeleteDelegation(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid delegation ID")
	}

	if err := h.service.DeleteDelegation(c.Context(), int64(id)); err != nil {
		reqLogger(c).Error("error deleting delegation", "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	uid, uname := auditUserFromCtx(c)
	did := int64(id)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "ipv6_delegation.deleted",
		ResourceType: "ipv6_delegation", ResourceID: &did,
	})

	return c.SendStatus(fiber.StatusNoContent)
}

// GetNetworkTopology handles GET /api/v1/networks/:id/topology
func (h *Handler) GetNetworkTopology(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2SubnetRead) {
		return nil
	}

	networkID, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid section ID")
	}

	topology, err := h.service.GetRepository().GetNetworkTopology(c.Context(), int64(networkID))
	if err != nil {
		reqLogger(c).Error("error getting section topology", "network_id", networkID, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	return c.JSON(topology)
}
