package handlers

import (
	"log"

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
	// Add logging middleware
	app.Use(loggingMiddleware)

	// API v1 routes
	api := app.Group("/api/v1")

	// Sections routes
	sections := api.Group("/sections")
	sections.Get("", h.ListSections)
	sections.Post("", h.CreateSection)
	sections.Get("/:id", h.GetSection)
	sections.Put("/:id", h.UpdateSection)
	sections.Delete("/:id", h.DeleteSection)

	log.Println("Routes registered successfully")
}

func loggingMiddleware(c *fiber.Ctx) error {
	log.Printf("%s %s", c.Method(), c.Path())
	return c.Next()
}
