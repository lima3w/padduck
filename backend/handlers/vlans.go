package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

type CreateVLANRequest struct {
	VRFID       *int64 `json:"vrf_id"`
	DomainID    *int64 `json:"domain_id"`
	GroupID     *int64 `json:"group_id"`
	VlanID      int    `json:"vlan_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateVLANRequest struct {
	DomainID    *int64 `json:"domain_id"`
	GroupID     *int64 `json:"group_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AssignSubnetToVLANRequest struct {
	SubnetID int64 `json:"subnet_id"`
}

type CreateVLANDomainRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type UpdateVLANDomainRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type CreateVLANGroupRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Colour      *string `json:"colour"`
}

type UpdateVLANGroupRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Colour      *string `json:"colour"`
}

// ListVLANs handles GET /api/v1/vlans
// Supports ?page=1&limit=25 for pagination. Without those params it returns all results.
func (h *Handler) ListVLANs(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2VLANList); err != nil {
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
		vlans, total, err := h.service.ListVLANsPaginated(c.Context(), page, limit)
		if err != nil {
			reqLogger(c).Error("error listing VLANs", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
		}
		if vlans == nil {
			vlans = make([]*models.VLAN, 0)
		}
		return c.JSON(fiber.Map{
			"data":  vlans,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}

	vlans, err := h.service.ListVLANs(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing VLANs", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	if vlans == nil {
		vlans = make([]*models.VLAN, 0)
	}

	return c.JSON(vlans)
}

func (h *Handler) GetVLAN(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2VLANRead); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VLAN ID"})
	}

	vlan, err := h.service.GetVLAN(c.Context(), int64(id))
	if err != nil {
		reqLogger(c).Error("error getting VLAN", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(vlan)
}

func (h *Handler) CreateVLAN(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2VLANWrite); err != nil {
		return nil
	}
	req := new(CreateVLANRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	vlan, err := h.service.CreateVLAN(c.Context(), req.VRFID, req.DomainID, req.GroupID, req.VlanID, req.Name, req.Description)
	if err != nil {
		reqLogger(c).Error("error creating VLAN", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "vlan_created",
		ResourceType: "vlan", ResourceID: &vlan.ID, ResourceName: fmt.Sprintf("%s (ID %d)", vlan.Name, vlan.VlanID),
		NewValues: map[string]interface{}{"vlan_id": vlan.VlanID, "name": vlan.Name},
	})

	return c.Status(fiber.StatusCreated).JSON(vlan)
}

func (h *Handler) UpdateVLAN(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VLAN ID"})
	}
	if err := h.permCheck(c, services.PermV2VLANWrite, services.ResourceScope{Type: "vlan", ID: int64(id)}); err != nil {
		return nil
	}

	req := new(UpdateVLANRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	vlan, err := h.service.UpdateVLAN(c.Context(), int64(id), req.DomainID, req.GroupID, req.Name, req.Description)
	if err != nil {
		reqLogger(c).Error("error updating VLAN", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "vlan_updated",
		ResourceType: "vlan", ResourceID: &vlan.ID, ResourceName: vlan.Name,
		NewValues: map[string]string{"name": req.Name, "description": req.Description},
	})

	return c.JSON(vlan)
}

func (h *Handler) DeleteVLAN(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VLAN ID"})
	}
	if err := h.permCheck(c, services.PermV2VLANDelete, services.ResourceScope{Type: "vlan", ID: int64(id)}); err != nil {
		return nil
	}

	if err := h.service.DeleteVLAN(c.Context(), int64(id)); err != nil {
		reqLogger(c).Error("error deleting VLAN", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	vid := int64(id)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "vlan_deleted",
		ResourceType: "vlan", ResourceID: &vid,
	})

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) GetVLANSubnets(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2VLANRead); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VLAN ID"})
	}

	subnets, err := h.service.GetVLANSubnets(c.Context(), int64(id))
	if err != nil {
		reqLogger(c).Error("error getting VLAN subnets", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	if subnets == nil {
		subnets = make([]*models.Subnet, 0)
	}

	return c.JSON(subnets)
}

func (h *Handler) AssignSubnetToVLAN(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VLAN ID"})
	}
	if err := h.permCheck(c, services.PermV2VLANWrite, services.ResourceScope{Type: "vlan", ID: int64(id)}); err != nil {
		return nil
	}

	req := new(AssignSubnetToVLANRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.permCheck(c, services.PermV2SubnetWrite, services.ResourceScope{Type: "subnet", ID: req.SubnetID}); err != nil {
		return nil
	}

	subnet, err := h.service.AssignSubnetToVLAN(c.Context(), int64(id), req.SubnetID)
	if err != nil {
		reqLogger(c).Error("error assigning subnet to VLAN", "subnet_id", req.SubnetID, "vlan_id", id, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	uid, uname := auditUserFromCtx(c)
	cidr := fmt.Sprintf("%s/%d", subnet.NetworkAddress, subnet.PrefixLength)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "vlan_subnet_assigned",
		ResourceType: "subnet", ResourceID: &subnet.ID, ResourceName: cidr,
		NewValues: map[string]interface{}{"vlan_id": id},
	})

	return c.JSON(subnet)
}

func (h *Handler) RemoveSubnetFromVLAN(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VLAN ID"})
	}
	subnetID, err := c.ParamsInt("subnetID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid subnet ID"})
	}
	if err := h.permCheck(c, services.PermV2VLANWrite, services.ResourceScope{Type: "vlan", ID: int64(id)}); err != nil {
		return nil
	}
	if err := h.permCheck(c, services.PermV2SubnetWrite, services.ResourceScope{Type: "subnet", ID: int64(subnetID)}); err != nil {
		return nil
	}

	subnet, err := h.service.RemoveSubnetFromVLAN(c.Context(), int64(id), int64(subnetID))
	if err != nil {
		reqLogger(c).Error("error removing subnet from VLAN", "subnet_id", subnetID, "vlan_id", id, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	uid, uname := auditUserFromCtx(c)
	cidr := fmt.Sprintf("%s/%d", subnet.NetworkAddress, subnet.PrefixLength)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "vlan_subnet_removed",
		ResourceType: "subnet", ResourceID: &subnet.ID, ResourceName: cidr,
		OldValues: map[string]interface{}{"vlan_id": id},
	})

	return c.JSON(subnet)
}

func (h *Handler) ListVLANsByVRF(c *fiber.Ctx) error {
	vrfID, err := c.ParamsInt("vrfID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VRF ID"})
	}

	vlans, err := h.service.ListVLANsByVRF(c.Context(), int64(vrfID))
	if err != nil {
		reqLogger(c).Error("error listing VLANs by VRF", "vrf_id", vrfID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	if vlans == nil {
		vlans = make([]*models.VLAN, 0)
	}

	return c.JSON(vlans)
}

// VLAN Domain handlers

func (h *Handler) ListVLANDomains(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2VLANDomainList); err != nil {
		return nil
	}
	domains, err := h.service.ListVLANDomains(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing VLAN domains", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if domains == nil {
		domains = make([]*models.VLANDomain, 0)
	}
	return c.JSON(domains)
}

func (h *Handler) GetVLANDomain(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2VLANDomainRead); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VLAN domain ID"})
	}
	domain, err := h.service.GetVLANDomain(c.Context(), int64(id))
	if err != nil {
		reqLogger(c).Error("error getting VLAN domain", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(domain)
}

func (h *Handler) CreateVLANDomain(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2VLANDomainWrite); err != nil {
		return nil
	}
	req := new(CreateVLANDomainRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	domain, err := h.service.CreateVLANDomain(c.Context(), req.Name, req.Description)
	if err != nil {
		reqLogger(c).Error("error creating VLAN domain", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "vlan_domain_created",
		ResourceType: "vlan_domain", ResourceID: &domain.ID, ResourceName: domain.Name,
		NewValues: map[string]interface{}{"name": domain.Name},
	})

	return c.Status(fiber.StatusCreated).JSON(domain)
}

func (h *Handler) UpdateVLANDomain(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2VLANDomainWrite); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VLAN domain ID"})
	}
	req := new(UpdateVLANDomainRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	domain, err := h.service.UpdateVLANDomain(c.Context(), int64(id), req.Name, req.Description)
	if err != nil {
		reqLogger(c).Error("error updating VLAN domain", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "vlan_domain_updated",
		ResourceType: "vlan_domain", ResourceID: &domain.ID, ResourceName: domain.Name,
		NewValues: map[string]interface{}{"name": req.Name},
	})

	return c.JSON(domain)
}

func (h *Handler) DeleteVLANDomain(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2VLANDomainDelete); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VLAN domain ID"})
	}

	if err := h.service.DeleteVLANDomain(c.Context(), int64(id)); err != nil {
		reqLogger(c).Error("error deleting VLAN domain", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	did := int64(id)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "vlan_domain_deleted",
		ResourceType: "vlan_domain", ResourceID: &did,
	})

	return c.SendStatus(fiber.StatusNoContent)
}

// VLAN Group handlers

func (h *Handler) ListVLANGroups(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2VLANGroupList); err != nil {
		return nil
	}
	groups, err := h.service.ListVLANGroups(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing VLAN groups", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if groups == nil {
		groups = make([]*models.VLANGroup, 0)
	}
	return c.JSON(groups)
}

func (h *Handler) GetVLANGroup(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2VLANGroupRead); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VLAN group ID"})
	}
	group, err := h.service.GetVLANGroup(c.Context(), int64(id))
	if err != nil {
		reqLogger(c).Error("error getting VLAN group", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(group)
}

func (h *Handler) CreateVLANGroup(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2VLANGroupWrite); err != nil {
		return nil
	}
	req := new(CreateVLANGroupRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	group, err := h.service.CreateVLANGroup(c.Context(), req.Name, req.Description, req.Colour)
	if err != nil {
		reqLogger(c).Error("error creating VLAN group", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "vlan_group_created",
		ResourceType: "vlan_group", ResourceID: &group.ID, ResourceName: group.Name,
		NewValues: map[string]interface{}{"name": group.Name},
	})

	return c.Status(fiber.StatusCreated).JSON(group)
}

func (h *Handler) UpdateVLANGroup(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2VLANGroupWrite); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VLAN group ID"})
	}
	req := new(UpdateVLANGroupRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	group, err := h.service.UpdateVLANGroup(c.Context(), int64(id), req.Name, req.Description, req.Colour)
	if err != nil {
		reqLogger(c).Error("error updating VLAN group", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "vlan_group_updated",
		ResourceType: "vlan_group", ResourceID: &group.ID, ResourceName: group.Name,
		NewValues: map[string]interface{}{"name": req.Name},
	})

	return c.JSON(group)
}

func (h *Handler) GetVLANUsageReport(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2VLANRead); err != nil {
		return nil
	}

	report, err := h.service.GetVLANUsageReport(c.Context())
	if err != nil {
		reqLogger(c).Error("error generating VLAN usage report", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(report)
}

func (h *Handler) DeleteVLANGroup(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2VLANGroupDelete); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VLAN group ID"})
	}

	if err := h.service.DeleteVLANGroup(c.Context(), int64(id)); err != nil {
		reqLogger(c).Error("error deleting VLAN group", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	gid := int64(id)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "vlan_group_deleted",
		ResourceType: "vlan_group", ResourceID: &gid,
	})

	return c.SendStatus(fiber.StatusNoContent)
}
