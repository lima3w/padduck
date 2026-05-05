package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
)

type SearchRequest struct {
	Query  string `json:"query"`
	Limit  int64  `json:"limit"`
	Offset int64  `json:"offset"`
	Status string `json:"status"`
}

// SearchSections handles POST /api/v1/sections/search
func (h *Handler) SearchSections(c *fiber.Ctx) error {
	req := new(SearchRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	sections, err := h.service.SearchSections(c.Context(), req.Query, req.Limit, req.Offset)
	if err != nil {
		log.Printf("Error searching sections: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Return empty array if nil
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

	subnets, err := h.service.SearchSubnets(c.Context(), int64(sectionID), req.Query, req.Limit, req.Offset)
	if err != nil {
		log.Printf("Error searching subnets: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Return empty array if nil
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

	req := new(SearchRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	ips, err := h.service.SearchIPAddresses(c.Context(), int64(subnetID), req.Query, req.Status, req.Limit, req.Offset)
	if err != nil {
		log.Printf("Error searching IP addresses: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Return empty array if nil
	if ips == nil {
		ips = make([]*models.IPAddress, 0)
	}

	return c.JSON(ips)
}
