package handlers

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
)

type CreateSubnetRequest struct {
	NetworkAddress string `json:"network_address"`
	PrefixLength   int    `json:"prefix_length"`
	Description    string `json:"description"`
}

type UpdateSubnetRequest struct {
	Description string `json:"description"`
}

// CreateSubnet handles POST /api/v1/sections/:sectionID/subnets
func (h *Handler) CreateSubnet(c *fiber.Ctx) error {
	sectionID, err := c.ParamsInt("sectionID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid section ID"})
	}
	if err := h.permCheck(c, services.PermV2SubnetWrite, services.ResourceScope{Type: "section", ID: int64(sectionID)}); err != nil {
		return err
	}

	req := new(CreateSubnetRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	subnet, err := h.service.CreateSubnet(c.Context(), int64(sectionID), req.NetworkAddress, req.PrefixLength, req.Description)
	if err != nil {
		log.Printf("Error creating subnet: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
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
		return err
	}

	subnet, err := h.service.GetSubnet(c.Context(), int64(id))
	if err != nil {
		log.Printf("Error getting subnet %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(subnet)
}

// ListSubnets handles GET /api/v1/sections/:sectionID/subnets
func (h *Handler) ListSubnets(c *fiber.Ctx) error {
	sectionID, err := c.ParamsInt("sectionID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid section ID"})
	}
	if err := h.permCheck(c, services.PermV2SubnetList, services.ResourceScope{Type: "section", ID: int64(sectionID)}); err != nil {
		return err
	}

	subnets, err := h.service.ListSubnets(c.Context(), int64(sectionID))
	if err != nil {
		log.Printf("Error listing subnets for section %d: %v", sectionID, err)
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
		return err
	}

	req := new(UpdateSubnetRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	subnet, err := h.service.UpdateSubnet(c.Context(), int64(id), req.Description)
	if err != nil {
		log.Printf("Error updating subnet %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
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

// DeleteSubnet handles DELETE /api/v1/subnets/:id
func (h *Handler) DeleteSubnet(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid subnet ID"})
	}
	if err := h.permCheck(c, services.PermV2SubnetDelete, services.ResourceScope{Type: "subnet", ID: int64(id)}); err != nil {
		return err
	}

	if err := h.service.DeleteSubnet(c.Context(), int64(id)); err != nil {
		log.Printf("Error deleting subnet %d: %v", id, err)
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
