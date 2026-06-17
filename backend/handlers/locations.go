package handlers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/repository"
	"padduck/services"
)

// ListLocations handles GET /api/v1/locations
// Supports ?page=1&limit=25 for pagination. Without those params it returns all results.
func (h *Handler) ListLocations(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2LocationList) {
		return nil
	}

	page, limit, _ := parseListOptions(c)
	if c.Query("page") != "" || c.Query("limit") != "" {
		locs, total, err := h.service.ListLocationsPaginated(c.Context(), page, limit)
		if err != nil {
			reqLogger(c).Error("error listing locations", "error", err)
			return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
		}
		if locs == nil {
			locs = make([]*models.Location, 0)
		}
		return c.JSON(fiber.Map{
			"data":  locs,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}

	locs, err := h.service.ListLocations(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing locations", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	if locs == nil {
		locs = make([]*models.Location, 0)
	}
	return c.JSON(locs)
}

// GetLocationTree handles GET /api/v1/locations/tree
func (h *Handler) GetLocationTree(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2LocationList) {
		return nil
	}
	tree, err := h.service.GetLocationTree(c.Context())
	if err != nil {
		reqLogger(c).Error("error getting location tree", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(tree)
}

// CreateLocation handles POST /api/v1/locations
func (h *Handler) CreateLocation(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2LocationWrite) {
		return nil
	}
	req := new(repository.LocationParams)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "location name is required")
	}
	if len(req.Name) > 255 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "location name must be 255 characters or fewer")
	}
	loc, err := h.service.CreateLocation(c.Context(), req)
	if err != nil {
		reqLogger(c).Error("error creating location", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.Status(fiber.StatusCreated).JSON(loc)
}

// GetLocation handles GET /api/v1/locations/:id
func (h *Handler) GetLocation(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2LocationRead) {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid location ID")
	}
	loc, err := h.service.GetLocation(c.Context(), int64(id))
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return RespondError(c, fiber.StatusNotFound, ErrNotFound, "location not found")
		}
		reqLogger(c).Error("error getting location", "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(loc)
}

// UpdateLocation handles PUT /api/v1/locations/:id
func (h *Handler) UpdateLocation(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2LocationWrite) {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid location ID")
	}
	req := new(repository.LocationParams)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "location name is required")
	}
	if len(req.Name) > 255 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "location name must be 255 characters or fewer")
	}
	loc, err := h.service.UpdateLocation(c.Context(), int64(id), req)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return RespondError(c, fiber.StatusNotFound, ErrNotFound, "location not found")
		}
		reqLogger(c).Error("error updating location", "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(loc)
}

// DeleteLocation handles DELETE /api/v1/locations/:id
func (h *Handler) DeleteLocation(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2LocationDelete) {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid location ID")
	}
	if err := h.service.DeleteLocation(c.Context(), int64(id)); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return RespondError(c, fiber.StatusNotFound, ErrNotFound, "location not found")
		}
		reqLogger(c).Error("error deleting location", "id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
