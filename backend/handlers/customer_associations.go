package handlers

import (
	"github.com/gofiber/fiber/v2"
	"ipam-next/repository"
	"ipam-next/services"
)

func (h *Handler) ListCustomerAssociations(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2CustomerRead); err != nil {
		return nil
	}
	customerID := int64(c.QueryInt("customer_id", 0))
	if paramID, err := c.ParamsInt("id"); err == nil && paramID > 0 {
		customerID = int64(paramID)
	}
	items, err := h.service.ListCustomerAssociations(c.Context(), customerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(items)
}

func (h *Handler) CreateCustomerAssociation(c *fiber.Ctx) error {
	req := new(repository.CustomerAssociationParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.permCheck(c, services.PermV2CustomerWrite); err != nil {
		return nil
	}
	item, err := h.service.CreateCustomerAssociation(c.Context(), req)
	if err != nil {
		return respondCustomerASError(c, err, "customer association")
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handler) DeleteCustomerAssociation(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid customer association ID"})
	}
	if err := h.permCheck(c, services.PermV2CustomerDelete); err != nil {
		return nil
	}
	if err := h.service.DeleteCustomerAssociation(c.Context(), id); err != nil {
		return respondCustomerASError(c, err, "customer association")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
