package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
)

type Handler struct {
	service *services.Service
}

func NewHandler(service *services.Service) *Handler {
	return &Handler{service: service}
}

// permCheck verifies the authenticated user has the given permission (with optional resource scopes).
// Returns a Fiber error response if denied, nil if allowed.
func (h *Handler) permCheck(c *fiber.Ctx, permission string, scopes ...services.ResourceScope) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "not authenticated")
	}
	if err := h.service.CheckPermission(c.Context(), user.ID, permission, scopes...); err != nil {
		return RespondError(c, fiber.StatusForbidden, ErrForbidden, "permission denied")
	}
	return nil
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
	auth.Post("/login", h.Login)
	auth.Post("/register", h.Register)
	auth.Get("/verify-email", h.VerifyEmail)
	auth.Post("/resend-verification", h.ResendVerification)
	auth.Post("/tokens/:userID", h.GenerateToken)
	auth.Get("/tokens/:userID", h.ListTokens)
	auth.Delete("/tokens/:tokenID", h.RevokeToken)
	auth.Post("/request-password-reset", h.RequestPasswordReset)
	auth.Post("/reset-password", h.ResetPassword)
	auth.Post("/unlock", h.RequestUnlock)
	auth.Get("/unlock", h.VerifyUnlock)

	// CSRF token endpoint
	api.Get("/csrf-token", h.GetCSRFToken)

	// Protected routes (require authentication)
	protected := api.Group("")
	protected.Use(h.AuthMiddleware)
	protected.Use(h.CSRFMiddleware)

	// MFA verification (public — called before session is issued)
	auth.Post("/verify-mfa", h.VerifyMFA)

	// User profile endpoints (protected)
	me := protected.Group("/auth/me")
	me.Get("", h.GetCurrentUser)
	me.Post("/logout", h.Logout)
	me.Post("/tokens", h.GenerateTokenForMe)
	me.Get("/tokens", h.ListMyTokens)
	me.Get("/sessions", h.ListMySessions)
	me.Delete("/sessions", h.LogoutAllDevices)
	me.Delete("/sessions/:sessionID", h.RevokeMySession)
	me.Get("/mfa", h.GetMFAStatus)
	me.Post("/mfa/setup", h.SetupTOTP)
	me.Post("/mfa/confirm", h.ConfirmTOTP)
	me.Delete("/mfa", h.DisableTOTP)
	me.Post("/mfa/backup-codes", h.RegenerateBackupCodes)

	// Security / audit endpoints
	user := protected.Group("/user")
	user.Get("/login-history", h.GetLoginHistory)

	// User management endpoints (protected)
	users := protected.Group("/users")
	users.Get("", h.ListUsers)
	users.Get("/:id", h.GetUser)
	users.Post("", h.CreateUser)
	users.Put("/:id/role", h.UpdateUserRole)
	users.Delete("/:id", h.DeleteUser)

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
	subnet.Get("/:subnetID/ip-addresses", h.ListIPAddresses)
	subnet.Post("/:subnetID/ip-addresses", h.CreateIPAddress)
	subnet.Post("/:subnetID/ip-addresses/allocate", h.AllocateIPAddress)

	// IP Addresses collection routes (nested under subnets, kept for compatibility)
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

	// Admin routes (protected + admin role required)
	admin := protected.Group("/admin")
	admin.Get("/config", h.GetConfig)
	admin.Put("/config", h.UpdateConfig)
	admin.Post("/config/test-email", h.TestSMTP)
	admin.Get("/approvals", h.ListPendingApprovals)
	admin.Post("/approvals/:id/approve", h.ApproveUser)
	admin.Post("/approvals/:id/reject", h.RejectUser)
	admin.Post("/users/:id/unlock", h.AdminUnlockUser)
	admin.Post("/users/:id/send-password-reset", h.SendPasswordResetEmail)
	admin.Get("/audit-logs", h.GetAuditLogs)
	admin.Get("/audit-logs/export", h.ExportAuditLogs)
	admin.Post("/audit-logs/purge", h.PurgeAuditLogs)

	// Role management
	admin.Get("/roles", h.ListRoles)
	admin.Post("/roles", h.CreateRole)
	admin.Get("/roles/:id", h.GetRole)
	admin.Put("/roles/:id", h.UpdateRole)
	admin.Delete("/roles/:id", h.DeleteRole)
	admin.Post("/roles/:id/permissions", h.AddPermissionToRole)
	admin.Delete("/roles/:id/permissions/:perm_id", h.RemovePermissionFromRole)
	admin.Get("/permissions", h.ListAvailablePermissions)

	// User-role assignment
	admin.Get("/users/:id/roles", h.GetUserRoles)
	admin.Post("/users/:id/roles", h.AssignRoleToUser)
	admin.Delete("/users/:id/roles/:role_id", h.RemoveRoleFromUser)

	log.Println("Routes registered successfully")
}

func loggingMiddleware(c *fiber.Ctx) error {
	log.Printf("%s %s", c.Method(), c.Path())
	return c.Next()
}
