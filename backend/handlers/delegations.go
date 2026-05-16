package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
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
	if err := h.permCheck(c, services.PermV2SubnetRead); err != nil {
		return nil
	}

	subnetID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid subnet ID"})
	}

	delegations, err := h.service.ListDelegations(c.Context(), int64(subnetID))
	if err != nil {
		log.Printf("ListDelegations error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if delegations == nil {
		delegations = make([]*models.IPv6Delegation, 0)
	}
	return c.JSON(delegations)
}

// CreateDelegation handles POST /api/v1/subnets/:id/delegations
func (h *Handler) CreateDelegation(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	subnetID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid subnet ID"})
	}

	req := new(createDelegationRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.DelegatedPrefix == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "delegated_prefix is required"})
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
		log.Printf("CreateDelegation error: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
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
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid delegation ID"})
	}

	req := new(updateDelegationRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.DelegatedPrefix == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "delegated_prefix is required"})
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
		log.Printf("UpdateDelegation error: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
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
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid delegation ID"})
	}

	if err := h.service.DeleteDelegation(c.Context(), int64(id)); err != nil {
		log.Printf("DeleteDelegation error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	did := int64(id)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "ipv6_delegation.deleted",
		ResourceType: "ipv6_delegation", ResourceID: &did,
	})

	return c.SendStatus(fiber.StatusNoContent)
}

// GetSectionTopology handles GET /api/v1/sections/:id/topology
func (h *Handler) GetSectionTopology(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2SubnetRead); err != nil {
		return nil
	}

	sectionID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid section ID"})
	}

	topology, err := h.service.GetRepository().GetSectionTopology(c.Context(), int64(sectionID))
	if err != nil {
		log.Printf("GetSectionTopology error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(topology)
}
