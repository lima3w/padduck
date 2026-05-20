package handlers

import (
	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/repository"
	"ipam-next/services"
)

func parseID(c *fiber.Ctx, name string) (int64, error) {
	id, err := c.ParamsInt(name)
	return int64(id), err
}

func (h *Handler) ListNATRules(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2NATList); err != nil {
		return nil
	}
	items, err := h.service.ListNATRules(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing NAT rules", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if items == nil {
		items = []*models.NATRule{}
	}
	return c.JSON(items)
}

func (h *Handler) GetNATRule(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2NATRead); err != nil {
		return nil
	}
	id, err := parseID(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid NAT rule ID"})
	}
	item, err := h.service.GetNATRule(c.Context(), id)
	if err != nil {
		return respondCustomerASError(c, err, "NAT rule")
	}
	return c.JSON(item)
}

func (h *Handler) CreateNATRule(c *fiber.Ctx) error {
	req := new(repository.NATRuleParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.permCheck(c, services.PermV2NATWrite); err != nil {
		return nil
	}
	item, err := h.service.CreateNATRule(c.Context(), req)
	if err != nil {
		return respondCustomerASError(c, err, "NAT rule")
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handler) UpdateNATRule(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid NAT rule ID"})
	}
	req := new(repository.NATRuleParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.permCheck(c, services.PermV2NATWrite); err != nil {
		return nil
	}
	item, err := h.service.UpdateNATRule(c.Context(), id, req)
	if err != nil {
		return respondCustomerASError(c, err, "NAT rule")
	}
	return c.JSON(item)
}

func (h *Handler) DeleteNATRule(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid NAT rule ID"})
	}
	if err := h.permCheck(c, services.PermV2NATDelete); err != nil {
		return nil
	}
	if err := h.service.DeleteNATRule(c.Context(), id); err != nil {
		return respondCustomerASError(c, err, "NAT rule")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListDHCPServers(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2DHCPList); err != nil {
		return nil
	}
	items, err := h.service.ListDHCPServers(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if items == nil {
		items = []*models.DHCPServer{}
	}
	return c.JSON(items)
}

func (h *Handler) CreateDHCPServer(c *fiber.Ctx) error {
	req := new(repository.DHCPServerParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.permCheck(c, services.PermV2DHCPWrite); err != nil {
		return nil
	}
	item, err := h.service.CreateDHCPServer(c.Context(), req)
	if err != nil {
		return respondCustomerASError(c, err, "DHCP server")
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handler) UpdateDHCPServer(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid DHCP server ID"})
	}
	req := new(repository.DHCPServerParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.permCheck(c, services.PermV2DHCPWrite); err != nil {
		return nil
	}
	item, err := h.service.UpdateDHCPServer(c.Context(), id, req)
	if err != nil {
		return respondCustomerASError(c, err, "DHCP server")
	}
	return c.JSON(item)
}

func (h *Handler) DeleteDHCPServer(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid DHCP server ID"})
	}
	if err := h.permCheck(c, services.PermV2DHCPDelete); err != nil {
		return nil
	}
	if err := h.service.DeleteDHCPServer(c.Context(), id); err != nil {
		return respondCustomerASError(c, err, "DHCP server")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListDHCPLeases(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2DHCPList); err != nil {
		return nil
	}
	items, err := h.service.ListDHCPLeases(c.Context(), int64(c.QueryInt("server_id", 0)))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if items == nil {
		items = []*models.DHCPLease{}
	}
	return c.JSON(items)
}

func (h *Handler) CreateDHCPLease(c *fiber.Ctx) error {
	req := new(repository.DHCPLeaseParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.permCheck(c, services.PermV2DHCPWrite); err != nil {
		return nil
	}
	item, err := h.service.CreateDHCPLease(c.Context(), req)
	if err != nil {
		return respondCustomerASError(c, err, "DHCP lease")
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handler) UpdateDHCPLease(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid DHCP lease ID"})
	}
	req := new(repository.DHCPLeaseParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.permCheck(c, services.PermV2DHCPWrite); err != nil {
		return nil
	}
	item, err := h.service.UpdateDHCPLease(c.Context(), id, req)
	if err != nil {
		return respondCustomerASError(c, err, "DHCP lease")
	}
	return c.JSON(item)
}

func (h *Handler) DeleteDHCPLease(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid DHCP lease ID"})
	}
	if err := h.permCheck(c, services.PermV2DHCPDelete); err != nil {
		return nil
	}
	if err := h.service.DeleteDHCPLease(c.Context(), id); err != nil {
		return respondCustomerASError(c, err, "DHCP lease")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
