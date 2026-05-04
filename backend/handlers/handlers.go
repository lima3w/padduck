package handlers

import "github.com/gofiber/fiber/v2"

type Handler struct {
	// services will be injected here
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	// All HTTP endpoints will be registered here
}
