package handlers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/repository"
	"ipam-next/services"
)

// ListDeviceTypes handles GET /api/v1/device-types
func (h *Handler) ListDeviceTypes(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2DeviceRead); err != nil {
		return nil
	}

	types, err := h.service.ListDeviceTypes(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing device types", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(types)
}

// ListDevices handles GET /api/v1/devices
// Supports ?page=1&limit=25 for pagination. Without those params it returns all results.
func (h *Handler) ListDevices(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2DeviceRead); err != nil {
		return nil
	}

	page := c.QueryInt("page", 0)
	limit := c.QueryInt("limit", 0)

	if page > 0 || limit > 0 || c.Query("sort") != "" || c.Query("q") != "" || c.Query("search") != "" {
		page, limit, opts := parseListOptions(c)
		offset := (page - 1) * limit
		devices, total, err := h.service.ListDevicesWithOptions(c.Context(), limit, offset, opts)
		if err != nil {
			reqLogger(c).Error("error listing devices", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
		}
		if devices == nil {
			devices = make([]*models.Device, 0)
		}
		return c.JSON(fiber.Map{
			"data":  devices,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}

	devices, err := h.service.ListAllDevices(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing devices", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if devices == nil {
		devices = make([]*models.Device, 0)
	}
	return c.JSON(devices)
}

// CreateDevice handles POST /api/v1/devices
func (h *Handler) CreateDevice(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2DeviceWrite); err != nil {
		return nil
	}

	req := new(repository.DeviceParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	req.Hostname = strings.TrimSpace(req.Hostname)
	if req.Hostname == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "hostname is required"})
	}

	device, err := h.service.CreateDevice(c.Context(), req)
	if err != nil {
		reqLogger(c).Error("error creating device", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(device)
}

// GetDevice handles GET /api/v1/devices/:id
func (h *Handler) GetDevice(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2DeviceRead); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid device ID"})
	}

	device, err := h.service.GetDevice(c.Context(), int64(id))
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "device not found"})
		}
		reqLogger(c).Error("error getting device", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(device)
}

// UpdateDevice handles PUT /api/v1/devices/:id
func (h *Handler) UpdateDevice(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2DeviceWrite); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid device ID"})
	}

	req := new(repository.DeviceParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	req.Hostname = strings.TrimSpace(req.Hostname)
	if req.Hostname == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "hostname is required"})
	}

	device, err := h.service.UpdateDevice(c.Context(), int64(id), req)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "device not found"})
		}
		reqLogger(c).Error("error updating device", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(device)
}

// DeleteDevice handles DELETE /api/v1/devices/:id
func (h *Handler) DeleteDevice(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2DeviceDelete); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid device ID"})
	}

	if err := h.service.DeleteDevice(c.Context(), int64(id)); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "device not found"})
		}
		reqLogger(c).Error("error deleting device", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetDeviceSNMPCredentials handles GET /api/v1/devices/:id/snmp-credentials
func (h *Handler) GetDeviceSNMPCredentials(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2DeviceAdmin); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid device ID"})
	}

	creds, err := h.service.GetDeviceSNMPCredentials(c.Context(), int64(id))
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "device not found"})
		}
		reqLogger(c).Error("error getting SNMP credentials", "device_id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(creds)
}

// ListDeviceIPAddresses handles GET /api/v1/devices/:id/ip-addresses
func (h *Handler) ListDeviceIPAddresses(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2DeviceRead); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid device ID"})
	}

	ips, err := h.service.ListDeviceIPAddresses(c.Context(), int64(id))
	if err != nil {
		reqLogger(c).Error("error listing IP addresses for device", "device_id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(ips)
}

type associateIPRequest struct {
	InterfaceName *string `json:"interface_name"`
	IsPrimary     bool    `json:"is_primary"`
}

// AssociateIPToDevice handles POST /api/v1/devices/:id/ip-addresses/:ip_id/associate
func (h *Handler) AssociateIPToDevice(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2DeviceWrite); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid device ID"})
	}

	ipID, err := c.ParamsInt("ip_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid IP address ID"})
	}

	req := new(associateIPRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.service.AssociateIPToDevice(c.Context(), int64(id), int64(ipID), req.InterfaceName, req.IsPrimary); err != nil {
		reqLogger(c).Error("error associating IP to device", "device_id", id, "ip_id", ipID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UnlinkIPFromDevice handles DELETE /api/v1/devices/:id/ip-addresses/:ip_id
func (h *Handler) UnlinkIPFromDevice(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2DeviceWrite); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid device ID"})
	}

	ipID, err := c.ParamsInt("ip_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid IP address ID"})
	}

	if err := h.service.UnlinkIPFromDevice(c.Context(), int64(id), int64(ipID)); err != nil {
		if errors.Is(err, services.ErrNotAssociated) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		reqLogger(c).Error("error unlinking IP from device", "device_id", id, "ip_id", ipID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListDeviceInterfaces handles GET /api/v1/devices/:id/interfaces
func (h *Handler) ListDeviceInterfaces(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2DeviceRead); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid device ID"})
	}

	ifaces, err := h.service.ListDeviceInterfaces(c.Context(), int64(id))
	if err != nil {
		reqLogger(c).Error("error listing interfaces for device", "device_id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(ifaces)
}

// CreateDeviceInterface handles POST /api/v1/devices/:id/interfaces
func (h *Handler) CreateDeviceInterface(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2DeviceWrite); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid device ID"})
	}

	req := new(repository.DeviceInterfaceParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "interface name is required"})
	}

	iface, err := h.service.CreateDeviceInterface(c.Context(), int64(id), req)
	if err != nil {
		reqLogger(c).Error("error creating interface for device", "device_id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(iface)
}

// UpdateDeviceInterface handles PUT /api/v1/devices/:id/interfaces/:if_id
func (h *Handler) UpdateDeviceInterface(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2DeviceWrite); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid device ID"})
	}

	ifID, err := c.ParamsInt("if_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid interface ID"})
	}

	req := new(repository.DeviceInterfaceParams)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "interface name is required"})
	}

	iface, err := h.service.UpdateDeviceInterface(c.Context(), int64(id), int64(ifID), req)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "interface not found"})
		}
		reqLogger(c).Error("error updating interface on device", "device_id", id, "if_id", ifID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(iface)
}

// DeleteDeviceInterface handles DELETE /api/v1/devices/:id/interfaces/:if_id
func (h *Handler) DeleteDeviceInterface(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2DeviceDelete); err != nil {
		return nil
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid device ID"})
	}

	ifID, err := c.ParamsInt("if_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid interface ID"})
	}

	if err := h.service.DeleteDeviceInterface(c.Context(), int64(id), int64(ifID)); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "interface not found"})
		}
		reqLogger(c).Error("error deleting interface on device", "device_id", id, "if_id", ifID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

type deviceSearchRequest struct {
	Query        string            `json:"query"`
	TypeID       *int64            `json:"type_id"`
	SectionID    *int64            `json:"section_id"`
	Vendor       *string           `json:"vendor"`
	IsOnline     *bool             `json:"is_online"`
	VLANID       *int64            `json:"vlan_id"`
	CustomFields map[string]string `json:"custom_fields"`
}

// SearchDevices handles POST /api/v1/devices/search
func (h *Handler) SearchDevices(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2DeviceRead); err != nil {
		return nil
	}

	req := new(deviceSearchRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	filter := &repository.DeviceSearchFilter{
		Query:     req.Query,
		TypeID:    req.TypeID,
		SectionID: req.SectionID,
		Vendor:    req.Vendor,
		IsOnline:  req.IsOnline,
		VLANID:    req.VLANID,
	}

	devices, err := h.service.SearchDevices(c.Context(), filter, req.CustomFields)
	if err != nil {
		reqLogger(c).Error("error searching devices", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(devices)
}
