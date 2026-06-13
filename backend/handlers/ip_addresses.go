package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

type CreateIPAddressRequest struct {
	Address      string             `json:"address"`
	Hostname     string             `json:"hostname"`
	Status       string             `json:"status"`
	TagID        *int64             `json:"tag_id"`
	MACAddress   *string            `json:"mac_address"`
	PTRRecord    *string            `json:"ptr_record"`
	DNSName      *string            `json:"dns_name"`
	CustomFields map[string]*string `json:"custom_fields"`
}

type AssignIPAddressRequest struct {
	DeviceID   *int64  `json:"device_id"`
	TagID      *int64  `json:"tag_id"`
	MACAddress *string `json:"mac_address"`
	PTRRecord  *string `json:"ptr_record"`
}

type UpdateIPMetaRequest struct {
	Hostname     string             `json:"hostname"`
	TagID        *int64             `json:"tag_id"`
	MACAddress   *string            `json:"mac_address"`
	PTRRecord    *string            `json:"ptr_record"`
	DNSName      *string            `json:"dns_name"`
	CustomFields map[string]*string `json:"custom_fields"`
}

// CreateIPAddress handles POST /api/v1/subnets/:subnetID/ip-addresses
func (h *Handler) CreateIPAddress(c *fiber.Ctx) error {
	subnetID, err := c.ParamsInt("subnetID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid subnet ID"})
	}
	if err := h.permCheck(c, services.PermV2IPAssign, services.ResourceScope{Type: "subnet", ID: int64(subnetID)}); err != nil {
		return nil
	}

	req := new(CreateIPAddressRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	ip, err := h.service.CreateIPAddress(c.Context(), int64(subnetID), req.Address, req.Hostname, req.Status, req.TagID, req.MACAddress, req.PTRRecord, req.DNSName, req.CustomFields)
	if err != nil {
		reqLogger(c).Error("error creating IP address", "subnet_id", subnetID, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "ip_address_created",
		ResourceType: "ip_address", ResourceID: &ip.ID, ResourceName: ip.Address,
		NewValues: map[string]string{"address": ip.Address, "hostname": ip.Hostname, "status": ip.Status},
	})

	return c.Status(fiber.StatusCreated).JSON(ip)
}

// GetIPAddress handles GET /api/v1/ip-addresses/:id
func (h *Handler) GetIPAddress(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2IPRead); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid IP address ID"})
	}

	ip, err := h.service.GetIPAddress(c.Context(), int64(id))
	if err != nil {
		reqLogger(c).Error("error getting IP address", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(ip)
}

// ListIPAddresses handles GET /api/v1/subnets/:subnetID/ip-addresses
// Supports ?page=1&limit=25 for pagination.
func (h *Handler) ListIPAddresses(c *fiber.Ctx) error {
	subnetID, err := c.ParamsInt("subnetID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid subnet ID"})
	}
	if err := h.permCheck(c, services.PermV2IPList, services.ResourceScope{Type: "subnet", ID: int64(subnetID)}); err != nil {
		return nil
	}

	if c.QueryBool("full_range") {
		page, limit, _ := parseListOptions(c)
		ips, total, err := h.service.ListIPAddressesFullRange(c.Context(), int64(subnetID), page, limit)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		if ips == nil {
			ips = make([]*models.IPAddress, 0)
		}
		return c.JSON(fiber.Map{"data": ips, "total": total, "page": page, "limit": limit})
	}

	page := c.QueryInt("page", 0)
	limit := c.QueryInt("limit", 0)

	if page > 0 || limit > 0 || c.Query("sort") != "" || c.Query("q") != "" || c.Query("search") != "" || c.Query("status") != "" || c.Query("hide_available") != "" {
		page, limit, opts := parseListOptions(c)
		ips, total, err := h.service.ListIPAddressesPaginatedWithOptions(c.Context(), int64(subnetID), page, limit, opts)
		if err != nil {
			reqLogger(c).Error("error listing IP addresses", "subnet_id", subnetID, "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
		}
		if ips == nil {
			ips = make([]*models.IPAddress, 0)
		}
		return c.JSON(fiber.Map{
			"data":  ips,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}

	ips, err := h.service.ListIPAddresses(c.Context(), int64(subnetID))
	if err != nil {
		reqLogger(c).Error("error listing IP addresses", "subnet_id", subnetID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if ips == nil {
		ips = make([]*models.IPAddress, 0)
	}
	return c.JSON(ips)
}

// AssignIPAddress handles POST /api/v1/ip-addresses/:id/assign
func (h *Handler) AssignIPAddress(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2IPAssign); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid IP address ID"})
	}

	req := new(AssignIPAddressRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	ip, err := h.service.AssignIPAddress(c.Context(), int64(id), req.DeviceID)
	if err != nil {
		reqLogger(c).Error("error assigning IP address", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	auditVals := map[string]string{}
	if req.DeviceID != nil {
		auditVals["device_id"] = fmt.Sprintf("%d", *req.DeviceID)
	}
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "ip_assigned",
		ResourceType: "ip_address", ResourceID: &ip.ID, ResourceName: ip.Address,
		NewValues: auditVals,
	})

	return c.JSON(ip)
}

// ReleaseIPAddress handles POST /api/v1/ip-addresses/:id/release
func (h *Handler) ReleaseIPAddress(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2IPRelease); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid IP address ID"})
	}

	ip, err := h.service.ReleaseIPAddress(c.Context(), int64(id))
	if err != nil {
		reqLogger(c).Error("error releasing IP address", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "ip_released",
		ResourceType: "ip_address", ResourceID: &ip.ID, ResourceName: ip.Address,
	})

	return c.JSON(ip)
}

// DeleteIPAddress handles DELETE /api/v1/ip-addresses/:id
func (h *Handler) DeleteIPAddress(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2IPAssign); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid IP address ID"})
	}

	if err := h.service.DeleteIPAddress(c.Context(), int64(id)); err != nil {
		reqLogger(c).Error("error deleting IP address", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	ipid := int64(id)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "ip_address_deleted",
		ResourceType: "ip_address", ResourceID: &ipid,
	})

	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateIPMeta handles PUT /api/v1/ip-addresses/:id
func (h *Handler) UpdateIPMeta(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2IPAssign); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid IP address ID"})
	}

	req := new(UpdateIPMetaRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	ip, err := h.service.UpdateIPAddressMeta(c.Context(), int64(id), req.Hostname, req.TagID, req.MACAddress, req.PTRRecord, req.DNSName, req.CustomFields)
	if err != nil {
		reqLogger(c).Error("error updating IP meta", "id", id, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(ip)
}

// QuickCreateIPAddress handles POST /api/v1/ip-addresses/quick-create
// Finds the most-specific subnet for the given address and creates a new IP record in it.
func (h *Handler) QuickCreateIPAddress(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2IPAssign); err != nil {
		return nil
	}
	var req struct {
		Address string `json:"address"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Address == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "address is required"})
	}
	ip, err := h.service.QuickCreateIPAddress(c.Context(), req.Address)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(ip)
}

// AllocateIPAddress handles POST /api/v1/subnets/:subnetID/ip-addresses/allocate
func (h *Handler) AllocateIPAddress(c *fiber.Ctx) error {
	subnetID, err := c.ParamsInt("subnetID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid subnet ID"})
	}
	if err := h.permCheck(c, services.PermV2IPAssign, services.ResourceScope{Type: "subnet", ID: int64(subnetID)}); err != nil {
		return nil
	}

	type AllocateRequest struct {
		DeviceID *int64 `json:"device_id"`
	}

	req := new(AllocateRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	ip, err := h.service.AllocateIPAddress(c.Context(), int64(subnetID), req.DeviceID)
	if err != nil {
		reqLogger(c).Error("error allocating IP address", "subnet_id", subnetID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	auditVals := map[string]string{}
	if req.DeviceID != nil {
		auditVals["device_id"] = fmt.Sprintf("%d", *req.DeviceID)
	}
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "ip_allocated",
		ResourceType: "ip_address", ResourceID: &ip.ID, ResourceName: ip.Address,
		NewValues: auditVals,
	})

	return c.Status(fiber.StatusCreated).JSON(ip)
}

// GetSubnetUtilization handles GET /api/v1/subnets/:subnetID/utilization
func (h *Handler) GetSubnetUtilization(c *fiber.Ctx) error {
	subnetID, err := c.ParamsInt("subnetID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid subnet ID"})
	}
	if err := h.permCheck(c, services.PermV2SubnetRead, services.ResourceScope{Type: "subnet", ID: int64(subnetID)}); err != nil {
		return nil
	}

	utilization, err := h.service.GetSubnetUtilization(c.Context(), int64(subnetID))
	if err != nil {
		reqLogger(c).Error("error getting subnet utilization", "subnet_id", subnetID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(utilization)
}

type AssignWithLeaseRequest struct {
	DeviceID          *int64 `json:"device_id"`
	LeaseDurationDays int    `json:"lease_duration_days"`
}

// AssignIPAddressWithLease handles POST /api/v1/ip-addresses/:id/assign-with-lease
func (h *Handler) AssignIPAddressWithLease(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid IP address ID"})
	}
	if err := h.permCheck(c, services.PermV2IPAssign); err != nil {
		return nil
	}

	req := new(AssignWithLeaseRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	ip, err := h.service.AssignIPAddressWithLease(c.Context(), int64(id), req.DeviceID, req.LeaseDurationDays)
	if err != nil {
		reqLogger(c).Error("error assigning IP address with lease", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(ip)
}

// IsIPLeaseExpired handles GET /api/v1/ip-addresses/:id/lease-status
func (h *Handler) IsIPLeaseExpired(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid IP address ID"})
	}
	if err := h.permCheck(c, services.PermV2IPRead); err != nil {
		return nil
	}

	expired, err := h.service.IsIPLeaseExpired(c.Context(), int64(id))
	if err != nil {
		reqLogger(c).Error("error checking IP lease status", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(fiber.Map{"expired": expired})
}

// ReleaseExpiredLease handles POST /api/v1/ip-addresses/:id/release-expired
func (h *Handler) ReleaseExpiredLease(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid IP address ID"})
	}
	if err := h.permCheck(c, services.PermV2IPRelease); err != nil {
		return nil
	}

	ip, err := h.service.ReleaseExpiredLease(c.Context(), int64(id))
	if err != nil {
		reqLogger(c).Error("error releasing expired lease", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(ip)
}

// GetNextAvailableIP handles GET /api/v1/subnets/:subnetID/next-available
// Returns the next free IP address in a subnet without allocating it.
// Useful for automation (n8n, Zapier, Make) to preview before acting.
func (h *Handler) GetNextAvailableIP(c *fiber.Ctx) error {
	subnetID, err := c.ParamsInt("subnetID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid subnet ID"})
	}
	if err := h.permCheck(c, services.PermV2IPList, services.ResourceScope{Type: "subnet", ID: int64(subnetID)}); err != nil {
		return nil
	}
	ip, err := h.service.FindNextAvailableIP(c.Context(), int64(subnetID))
	if err != nil {
		reqLogger(c).Error("error finding next available IP", "subnet_id", subnetID, "error", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "no available IP addresses in subnet"})
	}
	return c.JSON(ip)
}
