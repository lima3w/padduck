package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

// V2ListNetworks handles GET /api/v2/networks.
//
// Always paginates and returns the standard v2 envelope:
//
//	{ "data": [...], "meta": { "page": 1, "limit": 25, "total": 10 } }
func (h *Handler) V2ListNetworks(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2NetworkRead) {
		return nil
	}

	page, limit, opts := parseListOptions(c)
	networks, total, err := h.ops.IPAM.ListNetworksPaginatedWithOptions(c.Context(), page, limit, opts)
	if err != nil {
		reqLogger(c).Error("v2: error listing networks", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	if networks == nil {
		networks = make([]*models.Network, 0)
	}

	return c.JSON(V2List(networks, V2Meta{Page: page, Limit: limit, Total: total}))
}
