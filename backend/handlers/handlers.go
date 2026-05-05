package handlers

import (
	"log"
	"time"

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
	// Add rate limiting middleware (100 requests per minute per IP)
	app.Use(h.RateLimitMiddleware(100, 1*time.Minute))

	// Add logging middleware
	app.Use(loggingMiddleware)

	// API v1 routes
	api := app.Group("/api/v1")

	// Auth routes (public - no authentication required)
	auth := api.Group("/auth")
	auth.Post("/tokens/:userID", h.GenerateToken)
	auth.Get("/tokens/:userID", h.ListTokens)
	auth.Delete("/tokens/:tokenID", h.RevokeToken)

	// Protected routes (require authentication)
	protected := api.Group("")
	protected.Use(h.AuthMiddleware)

	// User profile endpoints (protected)
	me := protected.Group("/auth/me")
	me.Get("", h.GetCurrentUser)
	me.Post("/tokens", h.GenerateTokenForMe)
	me.Get("/tokens", h.ListMyTokens)

	// Sections routes
	sections := protected.Group("/sections")
	sections.Get("", h.ListSections)
	sections.Post("", h.CreateSection)
	sections.Post("/search", h.SearchSections)
	sections.Get("/:id", h.GetSection)
	sections.Put("/:id", h.UpdateSection)
	sections.Delete("/:id", h.DeleteSection)

	// Subnets collection routes (nested under sections)
	subnets := sections.Group("/:sectionID/subnets")
	subnets.Get("", h.ListSubnets)
	subnets.Post("", h.CreateSubnet)

	// Subnets resource routes (top-level)
	subnet := protected.Group("/subnets")
	subnet.Get("/:id", h.GetSubnet)
	subnet.Put("/:id", h.UpdateSubnet)
	subnet.Delete("/:id", h.DeleteSubnet)
	subnet.Get("/:subnetID/utilization", h.GetSubnetUtilization)
	subnet.Post("/search/:sectionID", h.SearchSubnets)

	// IP Addresses collection routes (nested under subnets)
	ipAddresses := subnets.Group("/:subnetID/ip-addresses")
	ipAddresses.Get("", h.ListIPAddresses)
	ipAddresses.Post("", h.CreateIPAddress)
	ipAddresses.Post("/allocate", h.AllocateIPAddress)

	// IP Addresses resource routes (top-level)
	ipAddress := protected.Group("/ip-addresses")
	ipAddress.Get("/:id", h.GetIPAddress)
	ipAddress.Post("/:id/assign", h.AssignIPAddress)
	ipAddress.Post("/:id/release", h.ReleaseIPAddress)
	ipAddress.Post("/:id/assign-with-lease", h.AssignIPAddressWithLease)
	ipAddress.Get("/:id/lease-status", h.IsIPLeaseExpired)
	ipAddress.Post("/:id/release-expired", h.ReleaseExpiredLease)
	ipAddress.Delete("/:id", h.DeleteIPAddress)
	ipAddress.Post("/search/:subnetID", h.SearchIPAddresses)

	// VRFs routes
	vrfs := protected.Group("/vrfs")
	vrfs.Get("", h.ListVRFs)
	vrfs.Post("", h.CreateVRF)
	vrfs.Get("/:id", h.GetVRF)
	vrfs.Put("/:id", h.UpdateVRF)
	vrfs.Delete("/:id", h.DeleteVRF)
	vrfs.Get("/:vrfID/vlans", h.ListVLANsByVRF)

	// VLANs routes (top-level)
	vlans := protected.Group("/vlans")
	vlans.Get("", h.ListVLANs)
	vlans.Post("", h.CreateVLAN)
	vlans.Get("/:id", h.GetVLAN)
	vlans.Put("/:id", h.UpdateVLAN)
	vlans.Delete("/:id", h.DeleteVLAN)

	log.Println("Routes registered successfully")
}

func loggingMiddleware(c *fiber.Ctx) error {
	log.Printf("%s %s", c.Method(), c.Path())
	return c.Next()
}
