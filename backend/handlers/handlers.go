package handlers

import (
	"github.com/gofiber/fiber/v2"
	"ipam-next/services"
)

type Handler struct {
	service *services.Service
}

func NewHandler(service *services.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	// All HTTP endpoints will be registered here
	// Example: app.Get("/api/v1/sections", h.ListSections)
}
