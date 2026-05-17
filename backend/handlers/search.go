package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
)

type SearchRequest struct {
	Query        string            `json:"query"`
	Limit        int64             `json:"limit"`
	Offset       int64             `json:"offset"`
	Status       string            `json:"status"`
	CustomFields map[string]string `json:"custom_fields"`
}

type IPSearchRequest struct {
	Query          string            `json:"query"`
	Limit          int64             `json:"limit"`
	Offset         int64             `json:"offset"`
	Status         string            `json:"status"`
	TagID          *int64            `json:"tag_id"`
	MACAddress     string            `json:"mac_address"`
	PTRRecord      string            `json:"ptr_record"`
	IsAssigned     *bool             `json:"is_assigned"`
	LastSeenAfter  *time.Time        `json:"last_seen_after"`
	LastSeenBefore *time.Time        `json:"last_seen_before"`
	CustomFields   map[string]string `json:"custom_fields"`
}

// SearchSections handles POST /api/v1/sections/search
func (h *Handler) SearchSections(c *fiber.Ctx) error {
	req := new(SearchRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	sections, err := h.service.SearchSections(c.Context(), req.Query, req.Limit, req.Offset)
	if err != nil {
		reqLogger(c).Error("error searching sections", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if sections == nil {
		sections = make([]*models.Section, 0)
	}

	return c.JSON(sections)
}

// SearchSubnets handles POST /api/v1/subnets/search/:sectionID
func (h *Handler) SearchSubnets(c *fiber.Ctx) error {
	sectionID, err := c.ParamsInt("sectionID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid section ID"})
	}

	req := new(SearchRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	subnets, err := h.service.SearchSubnets(c.Context(), int64(sectionID), req.Query, req.Limit, req.Offset, req.CustomFields)
	if err != nil {
		reqLogger(c).Error("error searching subnets", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if subnets == nil {
		subnets = make([]*models.Subnet, 0)
	}

	return c.JSON(subnets)
}

// SearchIPAddresses handles POST /api/v1/ip-addresses/search/:subnetID
func (h *Handler) SearchIPAddresses(c *fiber.Ctx) error {
	subnetID, err := c.ParamsInt("subnetID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid subnet ID"})
	}

	req := new(IPSearchRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	opts := services.IPSearchOptions{
		TagID:          req.TagID,
		MACAddress:     req.MACAddress,
		PTRRecord:      req.PTRRecord,
		IsAssigned:     req.IsAssigned,
		LastSeenAfter:  req.LastSeenAfter,
		LastSeenBefore: req.LastSeenBefore,
		CustomFields:   req.CustomFields,
	}

	ips, err := h.service.SearchIPAddresses(c.Context(), int64(subnetID), req.Query, req.Status, req.Limit, req.Offset, opts)
	if err != nil {
		reqLogger(c).Error("error searching IP addresses", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if ips == nil {
		ips = make([]*models.IPAddress, 0)
	}

	return c.JSON(ips)
}
