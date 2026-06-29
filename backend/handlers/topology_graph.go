package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"padduck/services"
)

// GetTopologyGraph handles GET /api/v1/topology/graph
// Query params: root_type, root_id (required); depth (optional, default 3)
func (h *Handler) GetTopologyGraph(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2SubnetRead) {
		return nil
	}

	rootType := c.Query("root_type")
	rootIDStr := c.Query("root_id")
	if rootType == "" || rootIDStr == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "root_type and root_id are required")
	}
	rootID, err := strconv.ParseInt(rootIDStr, 10, 64)
	if err != nil || rootID <= 0 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid root_id")
	}

	depth := 3
	if d := c.Query("depth"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 10 {
			depth = parsed
		}
	}

	graph, err := h.ops.Topology.GetNeighbors(c.Context(), orgIDFromCtx(c), rootType, rootID, depth)
	if err != nil {
		reqLogger(c).Error("topology graph error", "root_type", rootType, "root_id", rootID, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(graph)
}

// GetTopologyPath handles GET /api/v1/topology/path
// Query params: from_type, from_id, to_type, to_id (all required)
func (h *Handler) GetTopologyPath(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2SubnetRead) {
		return nil
	}

	fromType := c.Query("from_type")
	fromIDStr := c.Query("from_id")
	toType := c.Query("to_type")
	toIDStr := c.Query("to_id")
	if fromType == "" || fromIDStr == "" || toType == "" || toIDStr == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "from_type, from_id, to_type, and to_id are required")
	}

	fromID, err := strconv.ParseInt(fromIDStr, 10, 64)
	if err != nil || fromID <= 0 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid from_id")
	}
	toID, err := strconv.ParseInt(toIDStr, 10, 64)
	if err != nil || toID <= 0 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid to_id")
	}

	graph, err := h.ops.Topology.GetPath(c.Context(), fromType, fromID, toType, toID)
	if err != nil {
		reqLogger(c).Error("topology path error", "from", fromType, "from_id", fromID, "to", toType, "to_id", toID, "error", err)
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, err.Error())
	}
	return c.JSON(graph)
}
