package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"padduck/services"
)

func (h *Handler) ListJobs(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}
	return c.JSON(fiber.Map{"jobs": h.service.Jobs.List()})
}

func (h *Handler) GetJob(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid job ID")
	}
	job, ok := h.service.Jobs.Get(id)
	if !ok {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "job not found")
	}
	return c.JSON(job)
}

func (h *Handler) CancelJob(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid job ID")
	}
	job, err := h.service.Jobs.Cancel(id)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	return c.JSON(job)
}

func (h *Handler) RetryJob(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid job ID")
	}
	job, err := h.service.Jobs.Retry(id)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	return c.Status(fiber.StatusAccepted).JSON(job)
}
