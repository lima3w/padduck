package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

func redactScanProfile(profile *models.ScanProfile) *models.ScanProfile {
	if profile == nil || profile.SNMPCommunity == nil || *profile.SNMPCommunity == "" {
		return profile
	}
	clone := *profile
	redacted := "***"
	clone.SNMPCommunity = &redacted
	return &clone
}

func redactScanProfiles(profiles []*models.ScanProfile) []*models.ScanProfile {
	out := make([]*models.ScanProfile, 0, len(profiles))
	for _, profile := range profiles {
		out = append(out, redactScanProfile(profile))
	}
	return out
}

// ListScanProfiles handles GET /api/v1/admin/scan-profiles
func (h *Handler) ListScanProfiles(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}
	profiles, err := h.service.Discovery.ListScanProfiles(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing scan profiles", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(redactScanProfiles(profiles))
}

// CreateScanProfile handles POST /api/v1/admin/scan-profiles
func (h *Handler) CreateScanProfile(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	var req struct {
		Name            string  `json:"name"`
		Description     *string `json:"description"`
		ScanType        string  `json:"scan_type"`
		PingConcurrency int     `json:"ping_concurrency"`
		TCPPorts        *string `json:"tcp_ports"`
		DNSLookup       bool    `json:"dns_lookup"`
		SNMPCommunity   *string `json:"snmp_community"`
		SNMPVersion     string  `json:"snmp_version"`
	}
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	profile, err := h.service.Discovery.CreateScanProfile(c.Context(), req.Name, req.ScanType, req.Description, req.PingConcurrency, req.TCPPorts, req.DNSLookup, req.SNMPCommunity, req.SNMPVersion)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(redactScanProfile(profile))
}

// GetScanProfile handles GET /api/v1/admin/scan-profiles/:id
func (h *Handler) GetScanProfile(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid profile ID")
	}
	profile, err := h.service.Discovery.GetScanProfileByID(c.Context(), int64(id))
	if err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "scan profile not found")
	}
	return c.JSON(redactScanProfile(profile))
}

// UpdateScanProfile handles PUT /api/v1/admin/scan-profiles/:id
func (h *Handler) UpdateScanProfile(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid profile ID")
	}
	var req struct {
		Name            string  `json:"name"`
		Description     *string `json:"description"`
		ScanType        string  `json:"scan_type"`
		PingConcurrency int     `json:"ping_concurrency"`
		TCPPorts        *string `json:"tcp_ports"`
		DNSLookup       bool    `json:"dns_lookup"`
		SNMPCommunity   *string `json:"snmp_community"`
		SNMPVersion     string  `json:"snmp_version"`
	}
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	profile, err := h.service.Discovery.UpdateScanProfile(c.Context(), int64(id), req.Name, req.ScanType, req.Description, req.PingConcurrency, req.TCPPorts, req.DNSLookup, req.SNMPCommunity, req.SNMPVersion)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	return c.JSON(redactScanProfile(profile))
}

// DeleteScanProfile handles DELETE /api/v1/admin/scan-profiles/:id
func (h *Handler) DeleteScanProfile(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid profile ID")
	}
	if err := h.service.Discovery.DeleteScanProfile(c.Context(), int64(id)); err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to delete scan profile")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// GetSubnetScanProfile handles GET /api/v1/admin/subnets/:id/scan-profile
func (h *Handler) GetSubnetScanProfile(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}
	subnetID, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid subnet ID")
	}
	subnet, err := h.service.GetSubnet(c.Context(), int64(subnetID))
	if err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "subnet not found")
	}
	if subnet.ScanProfileID == nil {
		return c.JSON(fiber.Map{"profile": nil})
	}
	profile, err := h.service.Discovery.GetScanProfileByID(c.Context(), *subnet.ScanProfileID)
	if err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "scan profile not found")
	}
	return c.JSON(fiber.Map{"profile": redactScanProfile(profile)})
}

// SetSubnetScanProfile handles PUT /api/v1/admin/subnets/:id/scan-profile
func (h *Handler) SetSubnetScanProfile(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	subnetID, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid subnet ID")
	}
	var req struct {
		ProfileID *int64 `json:"profile_id"` // null to clear
	}
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if err := h.service.Discovery.SetSubnetScanProfile(c.Context(), int64(subnetID), req.ProfileID); err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to update subnet scan profile")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
