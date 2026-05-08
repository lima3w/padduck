package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
)

type CreateIPAddressRequest struct {
	Address      string             `json:"address"`
	Hostname     string             `json:"hostname"`
	Status       string             `json:"status"`
	TagID        *int64             `json:"tag_id"`
	MACAddress   *string            `json:"mac_address"`
	PTRRecord    *string            `json:"ptr_record"`
	CustomFields map[string]*string `json:"custom_fields"`
}

type AssignIPAddressRequest struct {
	AssignedTo string  `json:"assigned_to"`
	TagID      *int64  `json:"tag_id"`
	MACAddress *string `json:"mac_address"`
	PTRRecord  *string `json:"ptr_record"`
}

type UpdateIPMetaRequest struct {
	TagID        *int64             `json:"tag_id"`
	MACAddress   *string            `json:"mac_address"`
	PTRRecord    *string            `json:"ptr_record"`
	CustomFields map[string]*string `json:"custom_fields"`
}

// CreateIPAddress handles POST /api/v1/subnets/:subnetID/ip-addresses
func (h *Handler) CreateIPAddress(c *fiber.Ctx) error {
	subnetID, err := c.ParamsInt("subnetID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid subnet ID"})
	}
	if err := h.permCheck(c, services.PermV2IPAssign, services.ResourceScope{Type: "subnet", ID: int64(subnetID)}); err != nil {
		return err
	}

	req := new(CreateIPAddressRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	ip, err := h.service.CreateIPAddress(c.Context(), int64(subnetID), req.Address, req.Hostname, req.Status, req.TagID, req.MACAddress, req.PTRRecord, req.CustomFields)
	if err != nil {
		log.Printf("Error creating IP address: %v", err)
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
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid IP address ID"})
	}
	if err := h.permCheck(c, services.PermV2IPRead); err != nil {
		return nil
	}

	ip, err := h.service.GetIPAddress(c.Context(), int64(id))
	if err != nil {
		log.Printf("Error getting IP address %d: %v", id, err)
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
		return err
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
		ips, total, err := h.service.ListIPAddressesPaginated(c.Context(), int64(subnetID), page, limit)
		if err != nil {
			log.Printf("Error listing IP addresses for subnet %d: %v", subnetID, err)
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
		log.Printf("Error listing IP addresses for subnet %d: %v", subnetID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	if ips == nil {
		ips = make([]*models.IPAddress, 0)
	}
	return c.JSON(ips)
}

// AssignIPAddress handles POST /api/v1/ip-addresses/:id/assign
func (h *Handler) AssignIPAddress(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid IP address ID"})
	}
	if err := h.permCheck(c, services.PermV2IPAssign); err != nil {
		return nil
	}

	req := new(AssignIPAddressRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	ip, err := h.service.AssignIPAddress(c.Context(), int64(id), req.AssignedTo)
	if err != nil {
		log.Printf("Error assigning IP address %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "ip_assigned",
		ResourceType: "ip_address", ResourceID: &ip.ID, ResourceName: ip.Address,
		NewValues: map[string]string{"assigned_to": req.AssignedTo},
	})

	return c.JSON(ip)
}

// ReleaseIPAddress handles POST /api/v1/ip-addresses/:id/release
func (h *Handler) ReleaseIPAddress(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid IP address ID"})
	}
	if err := h.permCheck(c, services.PermV2IPRelease); err != nil {
		return nil
	}

	ip, err := h.service.ReleaseIPAddress(c.Context(), int64(id))
	if err != nil {
		log.Printf("Error releasing IP address %d: %v", id, err)
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
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid IP address ID"})
	}
	if err := h.permCheck(c, services.PermV2IPAssign); err != nil {
		return nil
	}

	if err := h.service.DeleteIPAddress(c.Context(), int64(id)); err != nil {
		log.Printf("Error deleting IP address %d: %v", id, err)
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
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid IP address ID"})
	}
	if err := h.permCheck(c, services.PermV2IPAssign); err != nil {
		return nil
	}

	req := new(UpdateIPMetaRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	ip, err := h.service.UpdateIPAddressMeta(c.Context(), int64(id), req.TagID, req.MACAddress, req.PTRRecord, req.CustomFields)
	if err != nil {
		log.Printf("Error updating IP meta %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(ip)
}

// AllocateIPAddress handles POST /api/v1/subnets/:subnetID/ip-addresses/allocate
func (h *Handler) AllocateIPAddress(c *fiber.Ctx) error {
	subnetID, err := c.ParamsInt("subnetID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid subnet ID"})
	}
	if err := h.permCheck(c, services.PermV2IPAssign, services.ResourceScope{Type: "subnet", ID: int64(subnetID)}); err != nil {
		return err
	}

	type AllocateRequest struct {
		AssignedTo string `json:"assigned_to"`
	}

	req := new(AllocateRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	ip, err := h.service.AllocateIPAddress(c.Context(), int64(subnetID), req.AssignedTo)
	if err != nil {
		log.Printf("Error allocating IP address: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "ip_allocated",
		ResourceType: "ip_address", ResourceID: &ip.ID, ResourceName: ip.Address,
		NewValues: map[string]string{"assigned_to": req.AssignedTo},
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
		return err
	}

	utilization, err := h.service.GetSubnetUtilization(c.Context(), int64(subnetID))
	if err != nil {
		log.Printf("Error getting subnet utilization %d: %v", subnetID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(utilization)
}

type AssignWithLeaseRequest struct {
	AssignedTo       string `json:"assigned_to"`
	LeaseDurationDays int    `json:"lease_duration_days"`
}

// AssignIPAddressWithLease handles POST /api/v1/ip-addresses/:id/assign-with-lease
func (h *Handler) AssignIPAddressWithLease(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid IP address ID"})
	}

	req := new(AssignWithLeaseRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	ip, err := h.service.AssignIPAddressWithLease(c.Context(), int64(id), req.AssignedTo, req.LeaseDurationDays)
	if err != nil {
		log.Printf("Error assigning IP address with lease %d: %v", id, err)
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

	expired, err := h.service.IsIPLeaseExpired(c.Context(), int64(id))
	if err != nil {
		log.Printf("Error checking IP lease status %d: %v", id, err)
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

	ip, err := h.service.ReleaseExpiredLease(c.Context(), int64(id))
	if err != nil {
		log.Printf("Error releasing expired lease %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(ip)
}
