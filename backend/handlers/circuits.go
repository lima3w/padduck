package handlers

import (
	"github.com/gofiber/fiber/v2"
	"ipam-next/repository"
	"ipam-next/services"
)

func (h *Handler) ListCircuitProviders(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2CircuitList); err != nil {
		return nil
	}
	items, err := h.service.ListCircuitProviders(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(items)
}

func (h *Handler) CreateCircuitProvider(c *fiber.Ctx) error {
	req := new(repository.CircuitProviderParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.permCheck(c, services.PermV2CircuitWrite); err != nil {
		return nil
	}
	item, err := h.service.CreateCircuitProvider(c.Context(), req)
	if err != nil {
		return respondCustomerASError(c, err, "circuit provider")
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handler) UpdateCircuitProvider(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid circuit provider ID"})
	}
	req := new(repository.CircuitProviderParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.permCheck(c, services.PermV2CircuitWrite); err != nil {
		return nil
	}
	item, err := h.service.UpdateCircuitProvider(c.Context(), id, req)
	if err != nil {
		return respondCustomerASError(c, err, "circuit provider")
	}
	return c.JSON(item)
}

func (h *Handler) DeleteCircuitProvider(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid circuit provider ID"})
	}
	if err := h.permCheck(c, services.PermV2CircuitDelete); err != nil {
		return nil
	}
	if err := h.service.DeleteCircuitProvider(c.Context(), id); err != nil {
		return respondCustomerASError(c, err, "circuit provider")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListPhysicalCircuits(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2CircuitList); err != nil {
		return nil
	}
	items, err := h.service.ListPhysicalCircuits(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(items)
}

func (h *Handler) CreatePhysicalCircuit(c *fiber.Ctx) error {
	req := new(repository.PhysicalCircuitParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.permCheck(c, services.PermV2CircuitWrite); err != nil {
		return nil
	}
	item, err := h.service.CreatePhysicalCircuit(c.Context(), req)
	if err != nil {
		return respondCustomerASError(c, err, "physical circuit")
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handler) UpdatePhysicalCircuit(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid physical circuit ID"})
	}
	req := new(repository.PhysicalCircuitParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.permCheck(c, services.PermV2CircuitWrite); err != nil {
		return nil
	}
	item, err := h.service.UpdatePhysicalCircuit(c.Context(), id, req)
	if err != nil {
		return respondCustomerASError(c, err, "physical circuit")
	}
	return c.JSON(item)
}

func (h *Handler) DeletePhysicalCircuit(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid physical circuit ID"})
	}
	if err := h.permCheck(c, services.PermV2CircuitDelete); err != nil {
		return nil
	}
	if err := h.service.DeletePhysicalCircuit(c.Context(), id); err != nil {
		return respondCustomerASError(c, err, "physical circuit")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListLogicalCircuits(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2CircuitList); err != nil {
		return nil
	}
	items, err := h.service.ListLogicalCircuits(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(items)
}

func (h *Handler) CreateLogicalCircuit(c *fiber.Ctx) error {
	req := new(repository.LogicalCircuitParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.permCheck(c, services.PermV2CircuitWrite); err != nil {
		return nil
	}
	item, err := h.service.CreateLogicalCircuit(c.Context(), req)
	if err != nil {
		return respondCustomerASError(c, err, "logical circuit")
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handler) UpdateLogicalCircuit(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid logical circuit ID"})
	}
	req := new(repository.LogicalCircuitParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.permCheck(c, services.PermV2CircuitWrite); err != nil {
		return nil
	}
	item, err := h.service.UpdateLogicalCircuit(c.Context(), id, req)
	if err != nil {
		return respondCustomerASError(c, err, "logical circuit")
	}
	return c.JSON(item)
}

func (h *Handler) DeleteLogicalCircuit(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid logical circuit ID"})
	}
	if err := h.permCheck(c, services.PermV2CircuitDelete); err != nil {
		return nil
	}
	if err := h.service.DeleteLogicalCircuit(c.Context(), id); err != nil {
		return respondCustomerASError(c, err, "logical circuit")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
