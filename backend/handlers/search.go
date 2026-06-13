package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
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

// GlobalSearch handles GET /api/v1/search?q=...
func (h *Handler) GlobalSearch(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2NetworkList); err != nil {
		return nil
	}
	q := c.Query("q")
	result, err := h.service.GlobalSearch(c.Context(), q, 5)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, err.Error())
	}
	return c.JSON(result)
}

// SearchNetworks handles POST /api/v1/networks/search
func (h *Handler) SearchNetworks(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2NetworkList); err != nil {
		return nil
	}

	req := new(SearchRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	sections, err := h.service.SearchNetworks(c.Context(), req.Query, req.Limit, req.Offset)
	if err != nil {
		reqLogger(c).Error("error searching sections", "error", err)
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}

	if sections == nil {
		sections = make([]*models.Network, 0)
	}

	return c.JSON(sections)
}

// SearchSubnets handles POST /api/v1/subnets/search/:networkID
func (h *Handler) SearchSubnets(c *fiber.Ctx) error {
	networkID, err := c.ParamsInt("networkID")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid section ID")
	}
	if err := h.permCheck(c, services.PermV2SubnetList, services.ResourceScope{Type: "section", ID: int64(networkID)}); err != nil {
		return nil
	}

	req := new(SearchRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	subnets, err := h.service.SearchSubnets(c.Context(), int64(networkID), req.Query, req.Limit, req.Offset, req.CustomFields)
	if err != nil {
		reqLogger(c).Error("error searching subnets", "error", err)
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
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
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid subnet ID")
	}
	if err := h.permCheck(c, services.PermV2IPList, services.ResourceScope{Type: "subnet", ID: int64(subnetID)}); err != nil {
		return nil
	}

	req := new(IPSearchRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
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
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}

	if ips == nil {
		ips = make([]*models.IPAddress, 0)
	}

	return c.JSON(ips)
}

// SearchIPAddressesGlobal handles GET /api/v1/ip-addresses/search?q=...
// Returns up to 20 IP addresses matching the query across all subnets.
func (h *Handler) SearchIPAddressesGlobal(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2IPList); err != nil {
		return nil
	}
	q := c.Query("q")
	ips, err := h.service.SearchIPAddressesGlobal(c.Context(), q)
	if err != nil {
		reqLogger(c).Error("error searching IP addresses globally", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "search failed")
	}
	if ips == nil {
		ips = make([]*models.IPAddress, 0)
	}
	return c.JSON(ips)
}
