package handlers

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
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
	Description         string             `json:"description"`
	Gateway             *string            `json:"gateway"`
	AutoReserveFirst    bool               `json:"auto_reserve_first"`
	AutoReserveLast     bool               `json:"auto_reserve_last"`
	LocationID          *int64             `json:"location_id"`
	NameserverID        *int64             `json:"nameserver_id"`
	VLANID              *int64             `json:"vlan_id"`
	CustomFields        map[string]*string `json:"custom_fields"`
	TechnitiumScopeName string             `json:"technitium_scope_name"`
}

// CreateSubnet handles POST /api/v1/networks/:networkID/subnets
func (h *Handler) CreateSubnet(c *fiber.Ctx) error {
	networkID, err := c.ParamsInt("networkID")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid section ID")
	}
	if !h.requirePerm(c, services.PermV2SubnetWrite, services.ResourceScope{Type: "section", ID: int64(networkID)}) {
		return nil
	}

	req := new(CreateSubnetRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	subnet, err := h.service.CreateSubnet(c.Context(), int64(networkID), req.NetworkAddress, req.PrefixLength, req.Description, req.Gateway, req.AutoReserveFirst, req.AutoReserveLast, req.LocationID, req.NameserverID, req.VLANID, req.CustomFields)
	if err != nil {
		var overlapErr *services.SubnetOverlapError
		if errors.As(err, &overlapErr) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": overlapErr.Error(), "conflicting_cidr": overlapErr.ConflictingCIDR})
		}
		reqLogger(c).Error("error creating subnet", "network_id", networkID, "error", err)
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}

	uid, uname := auditUserFromCtx(c)
	cidr := fmt.Sprintf("%s/%d", subnet.NetworkAddress, subnet.PrefixLength)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "subnet_created",
		ResourceType: "subnet", ResourceID: &subnet.ID, ResourceName: cidr,
		NewValues: map[string]interface{}{"cidr": cidr, "description": subnet.Description, "network_id": subnet.NetworkID},
	})

	return c.Status(fiber.StatusCreated).JSON(subnet)
}

// GetSubnet handles GET /api/v1/subnets/:id
func (h *Handler) GetSubnet(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2SubnetRead) {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid subnet ID")
	}

	subnet, err := h.service.GetSubnet(c.Context(), int64(id))
	if err != nil {
		reqLogger(c).Error("error getting subnet", "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	return c.JSON(subnet)
}

// ListSubnets handles GET /api/v1/networks/:networkID/subnets
// Supports ?page=1&limit=25 for pagination.
func (h *Handler) ListSubnets(c *fiber.Ctx) error {
	networkID, err := c.ParamsInt("networkID")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid section ID")
	}
	if !h.requirePerm(c, services.PermV2SubnetList, services.ResourceScope{Type: "section", ID: int64(networkID)}) {
		return nil
	}

	page, limit, opts := parseListOptions(c)
	if c.Query("page") != "" || c.Query("limit") != "" || opts.Sort != "" || opts.Query != "" {
		subnets, total, err := h.service.ListSubnetsPaginatedWithOptions(c.Context(), int64(networkID), page, limit, opts)
		if err != nil {
			reqLogger(c).Error("error listing subnets", "network_id", networkID, "error", err)
			return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
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

	subnets, err := h.service.ListSubnets(c.Context(), int64(networkID))
	if err != nil {
		reqLogger(c).Error("error listing subnets", "network_id", networkID, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
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
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid subnet ID")
	}
	if !h.requirePerm(c, services.PermV2SubnetWrite, services.ResourceScope{Type: "subnet", ID: int64(id)}) {
		return nil
	}

	req := new(UpdateSubnetRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	oldSubnet, _ := h.service.GetSubnet(c.Context(), int64(id))

	subnet, err := h.service.UpdateSubnet(c.Context(), int64(id), req.Description, req.Gateway, req.AutoReserveFirst, req.AutoReserveLast, req.LocationID, req.NameserverID, req.VLANID, req.CustomFields, req.TechnitiumScopeName)
	if err != nil {
		reqLogger(c).Error("error updating subnet", "id", id, "error", err)
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}

	uid, uname := auditUserFromCtx(c)
	cidr := fmt.Sprintf("%s/%d", subnet.NetworkAddress, subnet.PrefixLength)
	var oldVals interface{}
	if oldSubnet != nil {
		oldVals = map[string]interface{}{
			"description": oldSubnet.Description,
			"gateway":     oldSubnet.Gateway,
			"location_id": oldSubnet.LocationID,
			"vlan_id":     oldSubnet.VLANID,
		}
	}
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "subnet_updated",
		ResourceType: "subnet", ResourceID: &subnet.ID, ResourceName: cidr,
		OldValues: oldVals,
		NewValues: map[string]interface{}{
			"description": req.Description,
			"gateway":     req.Gateway,
			"location_id": req.LocationID,
			"vlan_id":     req.VLANID,
		},
	})

	return c.JSON(subnet)
}

// GetOverlapReport handles GET /api/v1/admin/subnets/overlap-report
func (h *Handler) GetOverlapReport(c *fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return nil
	}

	pairs, err := h.service.OverlapReport(c.Context())
	if err != nil {
		reqLogger(c).Error("error generating overlap report", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
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
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid subnet ID")
	}
	if !h.requirePerm(c, services.PermV2SubnetDelete, services.ResourceScope{Type: "subnet", ID: int64(id)}) {
		return nil
	}

	if err := h.service.DeleteSubnet(c.Context(), int64(id)); err != nil {
		reqLogger(c).Error("error deleting subnet", "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	uid, uname := auditUserFromCtx(c)
	sid := int64(id)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "subnet_deleted",
		ResourceType: "subnet", ResourceID: &sid,
	})

	return c.SendStatus(fiber.StatusNoContent)
}
