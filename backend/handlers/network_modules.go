package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/repository"
	"padduck/services"
)

func parseID(c *fiber.Ctx, name string) (int64, error) {
	id, err := c.ParamsInt(name)
	return int64(id), err
}

func (h *Handler) ListNATRules(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2NATList) {
		return nil
	}
	items, err := h.ops.NetworkModules.ListNATRules(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing NAT rules", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	if items == nil {
		items = []*models.NATRule{}
	}
	return c.JSON(items)
}

func (h *Handler) GetNATRule(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2NATRead) {
		return nil
	}
	id, err := parseID(c, "id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid NAT rule ID")
	}
	item, err := h.ops.NetworkModules.GetNATRule(c.Context(), id)
	if err != nil {
		return respondCustomerASError(c, err, "NAT rule")
	}
	return c.JSON(item)
}

func (h *Handler) CreateNATRule(c *fiber.Ctx) error {
	req := new(repository.NATRuleParams)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if !h.requirePerm(c, services.PermV2NATWrite) {
		return nil
	}
	item, err := h.ops.NetworkModules.CreateNATRule(c.Context(), req)
	if err != nil {
		return respondCustomerASError(c, err, "NAT rule")
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handler) UpdateNATRule(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid NAT rule ID")
	}
	req := new(repository.NATRuleParams)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if !h.requirePerm(c, services.PermV2NATWrite) {
		return nil
	}
	item, err := h.ops.NetworkModules.UpdateNATRule(c.Context(), id, req)
	if err != nil {
		return respondCustomerASError(c, err, "NAT rule")
	}
	return c.JSON(item)
}

func (h *Handler) DeleteNATRule(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid NAT rule ID")
	}
	if !h.requirePerm(c, services.PermV2NATDelete) {
		return nil
	}
	if err := h.ops.NetworkModules.DeleteNATRule(c.Context(), id); err != nil {
		return respondCustomerASError(c, err, "NAT rule")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListFirewallZones(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2FirewallList) {
		return nil
	}
	items, err := h.ops.NetworkModules.ListFirewallZones(c.Context())
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	if items == nil {
		items = []*models.FirewallZone{}
	}
	return c.JSON(items)
}

func (h *Handler) GetFirewallZone(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2FirewallRead) {
		return nil
	}
	id, err := parseID(c, "id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid firewall zone ID")
	}
	item, err := h.ops.NetworkModules.GetFirewallZone(c.Context(), id)
	if err != nil {
		return respondCustomerASError(c, err, "firewall zone")
	}
	return c.JSON(item)
}

func (h *Handler) CreateFirewallZone(c *fiber.Ctx) error {
	req := new(repository.FirewallZoneParams)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if !h.requirePerm(c, services.PermV2FirewallWrite) {
		return nil
	}
	item, err := h.ops.NetworkModules.CreateFirewallZone(c.Context(), req)
	if err != nil {
		return respondCustomerASError(c, err, "firewall zone")
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handler) UpdateFirewallZone(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid firewall zone ID")
	}
	req := new(repository.FirewallZoneParams)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if !h.requirePerm(c, services.PermV2FirewallWrite) {
		return nil
	}
	item, err := h.ops.NetworkModules.UpdateFirewallZone(c.Context(), id, req)
	if err != nil {
		return respondCustomerASError(c, err, "firewall zone")
	}
	return c.JSON(item)
}

func (h *Handler) DeleteFirewallZone(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid firewall zone ID")
	}
	if !h.requirePerm(c, services.PermV2FirewallDelete) {
		return nil
	}
	if err := h.ops.NetworkModules.DeleteFirewallZone(c.Context(), id); err != nil {
		return respondCustomerASError(c, err, "firewall zone")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListFirewallZoneMappings(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2FirewallList) {
		return nil
	}
	items, err := h.ops.NetworkModules.ListFirewallZoneMappings(c.Context(), int64(c.QueryInt("zone_id", 0)))
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	if items == nil {
		items = []*models.FirewallZoneMapping{}
	}
	return c.JSON(items)
}

func (h *Handler) CreateFirewallZoneMapping(c *fiber.Ctx) error {
	req := new(repository.FirewallZoneMappingParams)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if !h.requirePerm(c, services.PermV2FirewallWrite) {
		return nil
	}
	item, err := h.ops.NetworkModules.CreateFirewallZoneMapping(c.Context(), req)
	if err != nil {
		return respondCustomerASError(c, err, "firewall zone mapping")
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handler) UpdateFirewallZoneMapping(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid firewall mapping ID")
	}
	req := new(repository.FirewallZoneMappingParams)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if !h.requirePerm(c, services.PermV2FirewallWrite) {
		return nil
	}
	item, err := h.ops.NetworkModules.UpdateFirewallZoneMapping(c.Context(), id, req)
	if err != nil {
		return respondCustomerASError(c, err, "firewall zone mapping")
	}
	return c.JSON(item)
}

func (h *Handler) DeleteFirewallZoneMapping(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid firewall mapping ID")
	}
	if !h.requirePerm(c, services.PermV2FirewallDelete) {
		return nil
	}
	if err := h.ops.NetworkModules.DeleteFirewallZoneMapping(c.Context(), id); err != nil {
		return respondCustomerASError(c, err, "firewall zone mapping")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListDHCPServers(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2DHCPList) {
		return nil
	}
	items, err := h.ops.NetworkModules.ListDHCPServers(c.Context())
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	if items == nil {
		items = []*models.DHCPServer{}
	}
	return c.JSON(items)
}

func (h *Handler) CreateDHCPServer(c *fiber.Ctx) error {
	req := new(repository.DHCPServerParams)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if !h.requirePerm(c, services.PermV2DHCPWrite) {
		return nil
	}
	item, err := h.ops.NetworkModules.CreateDHCPServer(c.Context(), req)
	if err != nil {
		return respondCustomerASError(c, err, "DHCP server")
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handler) UpdateDHCPServer(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid DHCP server ID")
	}
	req := new(repository.DHCPServerParams)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if !h.requirePerm(c, services.PermV2DHCPWrite) {
		return nil
	}
	item, err := h.ops.NetworkModules.UpdateDHCPServer(c.Context(), id, req)
	if err != nil {
		return respondCustomerASError(c, err, "DHCP server")
	}
	return c.JSON(item)
}

func (h *Handler) DeleteDHCPServer(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid DHCP server ID")
	}
	if !h.requirePerm(c, services.PermV2DHCPDelete) {
		return nil
	}
	if err := h.ops.NetworkModules.DeleteDHCPServer(c.Context(), id); err != nil {
		return respondCustomerASError(c, err, "DHCP server")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListDHCPLeases(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2DHCPList) {
		return nil
	}
	items, err := h.ops.NetworkModules.ListDHCPLeases(c.Context(), int64(c.QueryInt("server_id", 0)))
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	if items == nil {
		items = []*models.DHCPLease{}
	}
	return c.JSON(items)
}

func (h *Handler) CreateDHCPLease(c *fiber.Ctx) error {
	req := new(repository.DHCPLeaseParams)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if !h.requirePerm(c, services.PermV2DHCPWrite) {
		return nil
	}
	item, err := h.ops.NetworkModules.CreateDHCPLease(c.Context(), req)
	if err != nil {
		return respondCustomerASError(c, err, "DHCP lease")
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handler) UpdateDHCPLease(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid DHCP lease ID")
	}
	req := new(repository.DHCPLeaseParams)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if !h.requirePerm(c, services.PermV2DHCPWrite) {
		return nil
	}
	item, err := h.ops.NetworkModules.UpdateDHCPLease(c.Context(), id, req)
	if err != nil {
		return respondCustomerASError(c, err, "DHCP lease")
	}
	return c.JSON(item)
}

func (h *Handler) DeleteDHCPLease(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid DHCP lease ID")
	}
	if !h.requirePerm(c, services.PermV2DHCPDelete) {
		return nil
	}
	if err := h.ops.NetworkModules.DeleteDHCPLease(c.Context(), id); err != nil {
		return respondCustomerASError(c, err, "DHCP lease")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
