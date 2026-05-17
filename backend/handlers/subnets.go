package handlers

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
)

type CreateSubnetRequest struct {
	NetworkAddress   string             `json:"network_address"`
	PrefixLength     int                `json:"prefix_length"`
	Description      string             `json:"description"`
	Gateway          *string            `json:"gateway"`
	AutoReserveFirst bool               `json:"auto_reserve_first"`
	AutoReserveLast  bool               `json:"auto_reserve_last"`
	LocationID       *int64             `json:"location_id"`
	NameserverID     *int64             `json:"nameserver_id"`
	VLANID           *int64             `json:"vlan_id"`
	CustomFields     map[string]*string `json:"custom_fields"`
}

type UpdateSubnetRequest struct {
	Description      string             `json:"description"`
	Gateway          *string            `json:"gateway"`
	AutoReserveFirst bool               `json:"auto_reserve_first"`
	AutoReserveLast  bool               `json:"auto_reserve_last"`
	LocationID       *int64             `json:"location_id"`
	NameserverID     *int64             `json:"nameserver_id"`
	VLANID           *int64             `json:"vlan_id"`
	CustomFields     map[string]*string `json:"custom_fields"`
}

// CreateSubnet handles POST /api/v1/sections/:sectionID/subnets
func (h *Handler) CreateSubnet(c *fiber.Ctx) error {
	sectionID, err := c.ParamsInt("sectionID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid section ID"})
	}
	if err := h.permCheck(c, services.PermV2SubnetWrite, services.ResourceScope{Type: "section", ID: int64(sectionID)}); err != nil {
		return nil
	}

	req := new(CreateSubnetRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	subnet, err := h.service.CreateSubnet(c.Context(), int64(sectionID), req.NetworkAddress, req.PrefixLength, req.Description, req.Gateway, req.AutoReserveFirst, req.AutoReserveLast, req.LocationID, req.NameserverID, req.VLANID, req.CustomFields)
	if err != nil {
		var overlapErr *services.SubnetOverlapError
		if errors.As(err, &overlapErr) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": overlapErr.Error(), "conflicting_cidr": overlapErr.ConflictingCIDR})
		}
		reqLogger(c).Error("error creating subnet", "section_id", sectionID, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	uid, uname := auditUserFromCtx(c)
	cidr := fmt.Sprintf("%s/%d", subnet.NetworkAddress, subnet.PrefixLength)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "subnet_created",
		ResourceType: "subnet", ResourceID: &subnet.ID, ResourceName: cidr,
		NewValues: map[string]interface{}{"cidr": cidr, "description": subnet.Description, "section_id": subnet.SectionID},
	})

	return c.Status(fiber.StatusCreated).JSON(subnet)
}

// GetSubnet handles GET /api/v1/subnets/:id
func (h *Handler) GetSubnet(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid subnet ID"})
	}
	if err := h.permCheck(c, services.PermV2SubnetRead); err != nil {
		return nil
	}

	subnet, err := h.service.GetSubnet(c.Context(), int64(id))
	if err != nil {
		reqLogger(c).Error("error getting subnet", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(subnet)
}

// ListSubnets handles GET /api/v1/sections/:sectionID/subnets
// Supports ?page=1&limit=25 for pagination.
func (h *Handler) ListSubnets(c *fiber.Ctx) error {
	sectionID, err := c.ParamsInt("sectionID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid section ID"})
	}
	if err := h.permCheck(c, services.PermV2SubnetList, services.ResourceScope{Type: "section", ID: int64(sectionID)}); err != nil {
		return nil
	}

	page := c.QueryInt("page", 0)
	limit := c.QueryInt("limit", 0)

	if page > 0 || limit > 0 {
		if page < 1 {
			page = 1
		}
		if limit < 1 {
			limit = 25
		}
		subnets, total, err := h.service.ListSubnetsPaginated(c.Context(), int64(sectionID), page, limit)
		if err != nil {
			reqLogger(c).Error("error listing subnets", "section_id", sectionID, "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
		}
		if subnets == nil {
			subnets = make([]*models.Subnet, 0)
		}
		return c.JSON(fiber.Map{
			"data":  subnets,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}

	subnets, err := h.service.ListSubnets(c.Context(), int64(sectionID))
	if err != nil {
		reqLogger(c).Error("error listing subnets", "section_id", sectionID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if subnets == nil {
		subnets = make([]*models.Subnet, 0)
	}
	return c.JSON(subnets)
}

// UpdateSubnet handles PUT /api/v1/subnets/:id
func (h *Handler) UpdateSubnet(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid subnet ID"})
	}
	if err := h.permCheck(c, services.PermV2SubnetWrite, services.ResourceScope{Type: "subnet", ID: int64(id)}); err != nil {
		return nil
	}

	req := new(UpdateSubnetRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	subnet, err := h.service.UpdateSubnet(c.Context(), int64(id), req.Description, req.Gateway, req.AutoReserveFirst, req.AutoReserveLast, req.LocationID, req.NameserverID, req.VLANID, req.CustomFields)
	if err != nil {
		reqLogger(c).Error("error updating subnet", "id", id, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	uid, uname := auditUserFromCtx(c)
	cidr := fmt.Sprintf("%s/%d", subnet.NetworkAddress, subnet.PrefixLength)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "subnet_updated",
		ResourceType: "subnet", ResourceID: &subnet.ID, ResourceName: cidr,
		NewValues: map[string]string{"description": req.Description},
	})

	return c.JSON(subnet)
}

// GetOverlapReport handles GET /api/v1/admin/subnets/overlap-report
func (h *Handler) GetOverlapReport(c *fiber.Ctx) error {
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok || currentUser.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "admin access required"})
	}

	pairs, err := h.service.OverlapReport(c.Context())
	if err != nil {
		reqLogger(c).Error("error generating overlap report", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	if pairs == nil {
		pairs = make([]*services.OverlapPair, 0)
	}

	return c.JSON(fiber.Map{"overlaps": pairs})
}

// DeleteSubnet handles DELETE /api/v1/subnets/:id
func (h *Handler) DeleteSubnet(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid subnet ID"})
	}
	if err := h.permCheck(c, services.PermV2SubnetDelete, services.ResourceScope{Type: "subnet", ID: int64(id)}); err != nil {
		return nil
	}

	if err := h.service.DeleteSubnet(c.Context(), int64(id)); err != nil {
		reqLogger(c).Error("error deleting subnet", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	sid := int64(id)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "subnet_deleted",
		ResourceType: "subnet", ResourceID: &sid,
	})

	return c.SendStatus(fiber.StatusNoContent)
}
