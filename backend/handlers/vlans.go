package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
)

type CreateVLANRequest struct {
	VRFID       *int64 `json:"vrf_id"`
	VlanID      int    `json:"vlan_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateVLANRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *Handler) ListVLANs(c *fiber.Ctx) error {
	vlans, err := h.service.ListVLANs(c.Context())
	if err != nil {
		log.Printf("Error listing VLANs: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	if vlans == nil {
		vlans = make([]*models.VLAN, 0)
	}

	return c.JSON(vlans)
}

func (h *Handler) GetVLAN(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VLAN ID"})
	}

	vlan, err := h.service.GetVLAN(c.Context(), int64(id))
	if err != nil {
		log.Printf("Error getting VLAN %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(vlan)
}

func (h *Handler) CreateVLAN(c *fiber.Ctx) error {
	req := new(CreateVLANRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	vlan, err := h.service.CreateVLAN(c.Context(), req.VRFID, req.VlanID, req.Name, req.Description)
	if err != nil {
		log.Printf("Error creating VLAN: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(vlan)
}

func (h *Handler) UpdateVLAN(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VLAN ID"})
	}

	req := new(UpdateVLANRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	vlan, err := h.service.UpdateVLAN(c.Context(), int64(id), req.Name, req.Description)
	if err != nil {
		log.Printf("Error updating VLAN %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(vlan)
}

func (h *Handler) DeleteVLAN(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VLAN ID"})
	}

	if err := h.service.DeleteVLAN(c.Context(), int64(id)); err != nil {
		log.Printf("Error deleting VLAN %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListVLANsByVRF(c *fiber.Ctx) error {
	vrfID, err := c.ParamsInt("vrfID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid VRF ID"})
	}

	vlans, err := h.service.ListVLANsByVRF(c.Context(), int64(vrfID))
	if err != nil {
		log.Printf("Error listing VLANs for VRF %d: %v", vrfID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	if vlans == nil {
		vlans = make([]*models.VLAN, 0)
	}

	return c.JSON(vlans)
}
