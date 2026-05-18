package handlers

import (
	"context"
	"crypto/rand"
	"errors"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
)

// errResponseWritten is a sentinel returned by permCheck after it has already
// written the error response. Callers must return nil (not this error) so Fiber
// does not invoke the default error handler on top of the written response.
var errResponseWritten = errors.New("response written")

type Handler struct {
	service      *services.Service
	tokenLimiter *tokenRateLimiter
	isProduction bool
	csrfSecret   []byte // per-process CSRF signing key, generated at startup
}

func NewHandler(service *services.Service, isProduction bool) *Handler {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		panic("handlers: failed to generate CSRF secret: " + err.Error())
	}
	return &Handler{
		service:      service,
		tokenLimiter: newTokenRateLimiter(),
		isProduction: isProduction,
		csrfSecret:   secret,
	}
}

// StartTokenLimiterCleanup starts the background cleanup goroutine for the
// token rate limiter. It should be called once after NewHandler, passing the
// application context so that the goroutine stops on shutdown.
func (h *Handler) StartTokenLimiterCleanup(ctx context.Context) {
	h.tokenLimiter.StartCleanup(ctx)
}

// permCheck verifies the authenticated user has the given permission.
// On denial it writes the error response and returns errResponseWritten (non-nil).
// Callers must do: if err := h.permCheck(...); err != nil { return nil }
func (h *Handler) permCheck(c *fiber.Ctx, permission string, scopes ...services.ResourceScope) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		_ = RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "not authenticated")
		return errResponseWritten
	}
	if err := h.service.CheckPermission(c.Context(), user.ID, permission, scopes...); err != nil {
		_ = RespondError(c, fiber.StatusForbidden, ErrForbidden, "permission denied")
		return errResponseWritten
	}
	return nil
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	// Add rate limiting middleware (100 requests per minute per IP)
	app.Use(h.RateLimitMiddleware(100, 1*time.Minute))

	// Attach a unique request ID to every request
	app.Use(RequestIDMiddleware)

	// Add logging middleware
	app.Use(loggingMiddleware)

	// Grafana SimpleJSON datasource routes (v1.14.0 #236) — Bearer token auth, no CSRF
	grafana := app.Group("/api/grafana")
	grafana.Use(h.AuthMiddleware)
	grafana.Get("/", h.GrafanaHealth)
	grafana.Post("/search", h.RequireBearerAuth, h.GrafanaSearch)
	grafana.Post("/query", h.RequireBearerAuth, h.GrafanaQuery)

	// API v1 routes
	api := app.Group("/api/v1")

	// Auth routes (public - no authentication required)
	auth := api.Group("/auth")
	auth.Post("/login", h.Login)
	auth.Post("/register", h.Register)
	auth.Get("/verify-email", h.VerifyEmail)
	auth.Post("/resend-verification", h.ResendVerification)
	auth.Post("/request-password-reset", h.RequestPasswordReset)
	auth.Post("/reset-password", h.ResetPassword)
	auth.Post("/unlock", h.RequestUnlock)
	auth.Get("/unlock", h.VerifyUnlock)
	auth.Post("/verify-mfa", h.VerifyMFA)
	h.RegisterExternalAuthPublicRoutes(auth)

	// CSRF token endpoint
	api.Get("/csrf-token", h.GetCSRFToken)

	// Protected routes (require authentication)
	protected := api.Group("")
	protected.Use(h.AuthMiddleware)
	protected.Use(h.CSRFMiddleware)
	protected.Get("/features", h.GetFeatures)
	protected.Get("/search", h.GlobalSearch)

	// User profile endpoints (protected)
	me := protected.Group("/auth/me")
	me.Get("", h.GetCurrentUser)
	me.Post("/logout", h.Logout)
	me.Post("/tokens", h.GenerateTokenForMe)
	me.Get("/tokens", h.ListMyTokens)
	me.Post("/tokens/:id/rotate", h.RotateToken)
	me.Post("/tokens/:id/extend", h.ExtendToken)
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
	user.Get("/notification-preferences", h.GetNotificationPreferences)
	user.Put("/notification-preferences", h.UpdateNotificationPreferences)

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
	subnet.Get("/:subnetID/next-available", h.GetNextAvailableIP)

	// IP Addresses collection routes (nested under subnets, kept for compatibility)
	ipAddresses := subnets.Group("/:subnetID/ip-addresses")
	ipAddresses.Get("", h.ListIPAddresses)
	ipAddresses.Post("", h.CreateIPAddress)
	ipAddresses.Post("/allocate", h.AllocateIPAddress)

	// IP Addresses resource routes (top-level)
	ipAddress := protected.Group("/ip-addresses")
	ipAddress.Get("/:id", h.GetIPAddress)
	ipAddress.Put("/:id", h.UpdateIPMeta)
	ipAddress.Post("/:id/assign", h.AssignIPAddress)
	ipAddress.Post("/:id/release", h.ReleaseIPAddress)
	ipAddress.Post("/:id/assign-with-lease", h.AssignIPAddressWithLease)
	ipAddress.Get("/:id/lease-status", h.IsIPLeaseExpired)
	ipAddress.Post("/:id/release-expired", h.ReleaseExpiredLease)
	ipAddress.Delete("/:id", h.DeleteIPAddress)
	ipAddress.Post("/search/:subnetID", h.SearchIPAddresses)

	// VRFs routes
	vrfs := protected.Group("/vrfs", h.requireFeature(featureVrfs))
	vrfs.Get("", h.ListVRFs)
	vrfs.Post("", h.CreateVRF)
	vrfs.Get("/:id", h.GetVRF)
	vrfs.Put("/:id", h.UpdateVRF)
	vrfs.Delete("/:id", h.DeleteVRF)
	vrfs.Get("/:vrfID/vlans", h.ListVLANsByVRF)

	// VLANs routes (top-level)
	vlans := protected.Group("/vlans", h.requireFeature(featureVlans))
	vlans.Get("", h.ListVLANs)
	vlans.Post("", h.CreateVLAN)
	vlans.Get("/:id", h.GetVLAN)
	vlans.Put("/:id", h.UpdateVLAN)
	vlans.Delete("/:id", h.DeleteVLAN)
	vlans.Get("/:id/subnets", h.GetVLANSubnets)
	vlans.Post("/:id/subnets", h.AssignSubnetToVLAN)
	vlans.Delete("/:id/subnets/:subnetID", h.RemoveSubnetFromVLAN)

	// VLAN Domains routes (v1.8.0 #206)
	vlanDomains := protected.Group("/vlan-domains", h.requireFeature(featureVlans))
	vlanDomains.Get("", h.ListVLANDomains)
	vlanDomains.Post("", h.CreateVLANDomain)
	vlanDomains.Get("/:id", h.GetVLANDomain)
	vlanDomains.Put("/:id", h.UpdateVLANDomain)
	vlanDomains.Delete("/:id", h.DeleteVLANDomain)

	// VLAN Groups routes (v1.8.0 #207)
	vlanGroups := protected.Group("/vlan-groups", h.requireFeature(featureVlans))
	vlanGroups.Get("", h.ListVLANGroups)
	vlanGroups.Post("", h.CreateVLANGroup)
	vlanGroups.Get("/:id", h.GetVLANGroup)
	vlanGroups.Put("/:id", h.UpdateVLANGroup)
	vlanGroups.Delete("/:id", h.DeleteVLANGroup)

	// Admin routes (protected + admin role required)
	admin := protected.Group("/admin")
	admin.Get("/config", h.GetConfig)
	admin.Put("/config", h.UpdateConfig)
	admin.Post("/config/test-email", h.TestSMTP)
	admin.Get("/updates/check", h.CheckForUpdates)
	admin.Get("/approvals", h.ListPendingApprovals)
	admin.Post("/approvals/:id/approve", h.ApproveUser)
	admin.Post("/approvals/:id/reject", h.RejectUser)
	admin.Get("/users", h.ListUsers)
	admin.Post("/users/:id/unlock", h.AdminUnlockUser)
	admin.Post("/users/:id/send-password-reset", h.SendPasswordResetEmail)
	admin.Put("/users/:id/email", h.UpdateUserEmail)
	admin.Get("/notification-stats", h.GetNotificationStats)
	admin.Get("/webhooks", h.ListWebhookEndpoints)
	admin.Post("/webhooks", h.CreateWebhookEndpoint)
	admin.Put("/webhooks/:id", h.UpdateWebhookEndpoint)
	admin.Delete("/webhooks/:id", h.DeleteWebhookEndpoint)
	admin.Get("/webhooks/deliveries", h.ListWebhookDeliveries)
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

	// User suspension (v0.8.14 #168)
	admin.Post("/users/:id/suspend", h.SuspendUser)
	admin.Post("/users/:id/unsuspend", h.UnsuspendUser)

	// User impersonation (v0.8.14 #167)
	admin.Post("/users/:id/impersonate", h.ImpersonateUser)

	// Bulk user operations (v0.8.14 #169)
	admin.Post("/users/bulk-suspend", h.BulkSuspendUsers)
	admin.Post("/users/bulk-activate", h.BulkActivateUsers)
	admin.Post("/users/bulk-delete", h.BulkDeleteUsers)
	admin.Post("/users/bulk-import", h.BulkImportUsers)

	// GDPR admin (v0.8.14 #170)
	admin.Post("/users/:id/gdpr-delete", h.GDPRDeleteUser)

	// Discovery / scan jobs (v0.9.0) + advanced scanning (v1.9.0)
	admin.Get("/scan-jobs", h.ListScanJobs)
	admin.Post("/scan-jobs", h.CreateScanJob)
	admin.Get("/scan-jobs/:id", h.GetScanJob)
	admin.Put("/scan-jobs/:id", h.UpdateScanJob)
	admin.Delete("/scan-jobs/:id", h.DeleteScanJob)
	admin.Post("/scan-jobs/:id/run", h.RunScanJobNow)
	admin.Get("/scan-jobs/:id/results", h.GetScanJobResults)
	// Scan history (#211)
	admin.Get("/scan-jobs/:id/history", h.GetScanJobHistory)
	admin.Get("/scan-jobs/:id/history/:run_id", h.GetScanRunDetail)
	// Scan agents (#212)
	admin.Get("/scan-agents", h.ListScanAgents)
	admin.Post("/scan-agents", h.CreateScanAgent)
	admin.Get("/scan-agents/:id", h.GetScanAgent)
	admin.Post("/scan-agents/:id/rotate-token", h.RotateScanAgentToken)
	admin.Delete("/scan-agents/:id", h.DeleteScanAgent)

	// Agent API routes (#212) — authenticated via Bearer token
	scanAgent := api.Group("/scan-agent")
	scanAgent.Use(h.AgentAuthMiddleware)
	scanAgent.Get("/jobs", h.AgentGetJobs)
	scanAgent.Post("/results", h.AgentPostResults)
	scanAgent.Post("/heartbeat", h.AgentHeartbeat)

	// Custom fields admin CRUD (v1.4.0)
	admin.Get("/custom-fields", h.ListCustomFieldDefinitions)
	admin.Post("/custom-fields", h.CreateCustomFieldDefinition)
	admin.Put("/custom-fields/reorder", h.ReorderCustomFieldDefinitions)
	admin.Get("/custom-fields/:id", h.GetCustomFieldDefinition)
	admin.Put("/custom-fields/:id", h.UpdateCustomFieldDefinition)
	admin.Delete("/custom-fields/:id", h.DeleteCustomFieldDefinition)

	// Subnet scan results (v0.9.0)
	subnet.Get("/:id/scan-results", h.GetSubnetScanResults)

	// Dashboard (v1.1.0 #174)
	dashboard := protected.Group("/dashboard")
	dashboard.Get("/summary", h.GetDashboardSummary)
	dashboard.Get("/recent-activity", h.GetDashboardRecentActivity)

	// Subnet tree (v1.1.0 #177)
	sections.Get("/:id/subnets/tree", h.GetSubnetTree)

	// Subnet overlap report (v1.2.0 #181)
	admin.Get("/subnets/overlap-report", h.GetOverlapReport)

	// IP Tags (v1.2.0 #179)
	tags := protected.Group("/tags")
	tags.Get("", h.ListTags)
	tags.Post("", h.CreateTag)
	tags.Put("/:id", h.UpdateTag)
	tags.Delete("/:id", h.DeleteTag)

	// Devices (v1.3.0)
	deviceTypes := protected.Group("/device-types", h.requireFeature(featureDevices))
	deviceTypes.Get("", h.ListDeviceTypes)

	devices := protected.Group("/devices", h.requireFeature(featureDevices))
	devices.Get("", h.ListDevices)
	devices.Post("", h.CreateDevice)
	devices.Post("/search", h.SearchDevices)
	devices.Get("/:id", h.GetDevice)
	devices.Put("/:id", h.UpdateDevice)
	devices.Delete("/:id", h.DeleteDevice)
	devices.Get("/:id/ip-addresses", h.ListDeviceIPAddresses)
	devices.Post("/:id/ip-addresses/:ip_id/associate", h.AssociateIPToDevice)
	devices.Delete("/:id/ip-addresses/:ip_id", h.UnlinkIPFromDevice)
	devices.Get("/:id/interfaces", h.ListDeviceInterfaces)
	devices.Post("/:id/interfaces", h.CreateDeviceInterface)
	devices.Put("/:id/interfaces/:if_id", h.UpdateDeviceInterface)
	devices.Delete("/:id/interfaces/:if_id", h.DeleteDeviceInterface)
	devices.Get("/:id/snmp-credentials", h.GetDeviceSNMPCredentials)

	// Racks (v1.5.0 #195)
	racks := protected.Group("/racks", h.requireFeature(featureRacks))
	racks.Get("", h.ListRacks)
	racks.Post("", h.CreateRack)
	racks.Get("/:id", h.GetRack)
	racks.Put("/:id", h.UpdateRack)
	racks.Delete("/:id", h.DeleteRack)
	racks.Get("/:id/devices", h.ListDevicesInRack)

	// Locations (v1.5.0 #194)
	locations := protected.Group("/locations", h.requireFeature(featureLocations))
	locations.Get("", h.ListLocations)
	locations.Get("/tree", h.GetLocationTree)
	locations.Post("", h.CreateLocation)
	locations.Get("/:id", h.GetLocation)
	locations.Put("/:id", h.UpdateLocation)
	locations.Delete("/:id", h.DeleteLocation)

	// Nameservers (v1.6.0 #198)
	nameservers := protected.Group("/nameservers")
	nameservers.Get("", h.ListNameservers)
	nameservers.Post("", h.CreateNameserver)
	nameservers.Get("/:id", h.GetNameserver)
	nameservers.Put("/:id", h.UpdateNameserver)
	nameservers.Delete("/:id", h.DeleteNameserver)

	// DNS admin endpoints (v1.6.0 #199, #200)
	admin.Post("/dns/check-all", h.CheckAllDNS)
	admin.Post("/dns/test", h.TestPowerDNSConnection)
	admin.Post("/dns/technitium/test", h.TestTechnitiumConnection)

	// DNS zone browser (v1.6.0 #201)
	dns := protected.Group("/dns")
	dns.Get("/zones", h.ListDNSZones)
	dns.Get("/zones/:zone/records", h.GetDNSZoneRecords)

	// Request workflows — user endpoints (v1.7.0 #202 #203)
	requests := protected.Group("/requests")
	requests.Post("/subnets", h.SubmitSubnetRequest)
	requests.Get("/subnets", h.ListMySubnetRequests)
	requests.Delete("/subnets/:id", h.CancelSubnetRequest)
	requests.Post("/ips", h.SubmitIPRequest)
	requests.Get("/ips", h.ListMyIPRequests)
	requests.Delete("/ips/:id", h.CancelIPRequest)
	// Comments (#204)
	requests.Get("/:type/:id/comments", h.ListRequestComments)
	requests.Post("/:type/:id/comments", h.AddRequestComment)

	// Request workflows — admin endpoints (v1.7.0 #202 #203 #205)
	admin.Get("/requests/subnets", h.ListAllSubnetRequests)
	admin.Post("/requests/subnets/:id/approve", h.ApproveSubnetRequest)
	admin.Post("/requests/subnets/:id/reject", h.RejectSubnetRequest)
	admin.Get("/requests/ips", h.ListAllIPRequests)
	admin.Post("/requests/ips/:id/approve", h.ApproveIPRequest)
	admin.Post("/requests/ips/:id/reject", h.RejectIPRequest)
	admin.Get("/requests/pending-count", h.GetPendingRequestCount)

	// VLAN usage report (v1.8.0 #209)
	admin.Get("/vlans/usage-report", h.requireFeature(featureVlans), h.GetVLANUsageReport)

	// Network tools (v1.10.0 #216 #217)
	admin.Post("/subnets/:id/split", h.SplitSubnet)
	admin.Post("/subnets/merge", h.MergeSubnets)
	admin.Post("/subnets/:id/resize", h.ResizeSubnet)

	// IPv6 delegations (v1.10.0 #218)
	subnet.Get("/:id/delegations", h.ListDelegations)
	subnet.Post("/:id/delegations", h.CreateDelegation)
	delegations := protected.Group("/delegations")
	delegations.Put("/:id", h.UpdateDelegation)
	delegations.Delete("/:id", h.DeleteDelegation)

	// Network topology (v1.10.0 #219)
	sections.Get("/:id/topology", h.GetSectionTopology)

	// Reporting & Analytics (v1.11.0 #220 #221 #222 #223 #224)
	// Utilisation history
	subnet.Get("/:id/utilisation/history", h.GetSubnetUtilisationHistory)
	admin.Get("/reports/utilization-trends", h.GetUtilisationTrends)
	// Threshold alerts
	admin.Get("/reports/subnets-near-capacity", h.GetSubnetsNearCapacity)
	// Scheduled reports
	admin.Get("/reports/scheduled", h.ListScheduledReports)
	admin.Post("/reports/scheduled", h.CreateScheduledReport)
	admin.Get("/reports/scheduled/:id", h.GetScheduledReport)
	admin.Put("/reports/scheduled/:id", h.UpdateScheduledReport)
	admin.Delete("/reports/scheduled/:id", h.DeleteScheduledReport)
	admin.Post("/reports/scheduled/:id/run", h.RunScheduledReportNow)
	// Export endpoints
	admin.Get("/reports/export/subnets", h.ExportSubnets)
	admin.Get("/reports/export/ips", h.ExportIPs)
	admin.Get("/reports/export/inactive-ips", h.ExportInactiveIPs)
	admin.Get("/reports/export/sections", h.ExportSectionsCSV)
	admin.Get("/reports/export/devices", h.ExportDevicesCSV)
	admin.Get("/reports/export/vlans", h.ExportVLANsCSV)
	admin.Get("/reports/export/vrfs", h.ExportVRFsCSV)
	// Inactive IP reclamation
	admin.Get("/reports/inactive-ips", h.GetInactiveIPs)
	admin.Post("/ip-addresses/bulk-release", h.BulkReleaseIPs)

	// Import & Export (v1.12.0 #225 #226 #227 #228)
	admin.Post("/import/subnets", h.ImportSubnetsCSV)
	admin.Post("/import/ips", h.ImportIPsCSV)
	admin.Post("/import/phpipam", h.ImportFromPHPIpam)
	admin.Get("/export/full", h.ExportFullData)

	// LDAP & SSO (v1.13.0 #229 #230 #231 #232)
	h.RegisterExternalAuthAdminRoutes(admin)

	// GDPR user self-service (v0.8.14 #170)
	me.Get("/export", h.ExportMyData)
	me.Post("/deletion-request", h.RequestDeletion)

	// Privacy policy (v0.8.14 #171)
	me.Post("/accept-privacy", h.AcceptPrivacyPolicy)
	api.Get("/privacy-policy/version", h.GetPrivacyPolicyVersion)

	// Customers (v1.14.0 #234)
	customers := protected.Group("/customers", h.requireFeature(featureCustomers))
	customers.Get("", h.ListCustomers)
	customers.Post("", h.CreateCustomer)
	customers.Get("/:id", h.GetCustomer)
	customers.Put("/:id", h.UpdateCustomer)
	customers.Delete("/:id", h.DeleteCustomer)

	// BGP Autonomous Systems (v1.14.0 #235)
	asSystems := protected.Group("/autonomous-systems", h.requireFeature(featureBgp))
	asSystems.Get("", h.ListAutonomousSystems)
	asSystems.Post("", h.CreateAutonomousSystem)
	asSystems.Get("/:id", h.GetAutonomousSystem)
	asSystems.Put("/:id", h.UpdateAutonomousSystem)
	asSystems.Delete("/:id", h.DeleteAutonomousSystem)

	log.Println("Routes registered successfully")
}

func loggingMiddleware(c *fiber.Ctx) error {
	log.Printf("%s %s", c.Method(), c.Path())
	return c.Next()
}
