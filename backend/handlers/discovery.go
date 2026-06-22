package handlers

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

type CreateScanJobRequest struct {
	Name            string  `json:"name"`
	Subnet          string  `json:"subnet,omitempty"`
	SubnetIDs       []int64 `json:"subnet_ids"`
	ScheduleCron    *string `json:"schedule_cron,omitempty"`
	AutoAddIPs      *bool   `json:"auto_add_ips,omitempty"`
	ScanType        string  `json:"scan_type,omitempty"`
	PingConcurrency int     `json:"ping_concurrency,omitempty"`
	NotifyOnChange  bool    `json:"notify_on_change,omitempty"`
	IsActive        *bool   `json:"is_active,omitempty"`
	DiscoverDNS     *bool   `json:"discover_dns,omitempty"`
	DNSOverwrite    bool    `json:"dns_overwrite,omitempty"`
}

type UpdateScanJobRequest struct {
	Name            string  `json:"name"`
	Subnet          string  `json:"subnet,omitempty"`
	SubnetIDs       []int64 `json:"subnet_ids"`
	ScheduleCron    *string `json:"schedule_cron,omitempty"`
	IsActive        bool    `json:"is_active"`
	PingConcurrency int     `json:"ping_concurrency,omitempty"`
	NotifyOnChange  bool    `json:"notify_on_change,omitempty"`
	ScanType        string  `json:"scan_type,omitempty"`
	AgentID         *int64  `json:"agent_id,omitempty"`
	AutoAddIPs      *bool   `json:"auto_add_ips,omitempty"`
	DiscoverDNS     *bool   `json:"discover_dns,omitempty"`
	DNSOverwrite    bool    `json:"dns_overwrite,omitempty"`
}

// ListScanJobs handles GET /api/v1/admin/scan-jobs
func (h *Handler) ListScanJobs(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}
	jobs, err := h.ops.Discovery.ListJobs(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing scan jobs", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(jobs)
}

// CreateScanJob handles POST /api/v1/admin/scan-jobs
func (h *Handler) CreateScanJob(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	user, _ := c.Locals("user").(*models.User)

	var req CreateScanJobRequest
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}

	// If subnet_ids is empty but a CIDR string was provided, resolve it.
	if len(req.SubnetIDs) == 0 && req.Subnet != "" {
		sn, snErr := h.service.GetRepository().GetSubnetByCIDR(c.Context(), req.Subnet)
		if snErr != nil {
			return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "subnet not found for CIDR: " + req.Subnet)
		}
		req.SubnetIDs = []int64{sn.ID}
	}

	autoAddIPs := true
	if req.AutoAddIPs != nil {
		autoAddIPs = *req.AutoAddIPs
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	discoverDNS := true
	if req.DiscoverDNS != nil {
		discoverDNS = *req.DiscoverDNS
	}
	job, err := h.ops.Discovery.CreateJob(c.Context(), req.Name, req.SubnetIDs, req.ScheduleCron, user.ID, autoAddIPs)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	job, err = h.ops.Discovery.UpdateJobFull(c.Context(), job.ID, job.Name, job.SubnetIDs, job.ScheduleCron, isActive, req.PingConcurrency, req.NotifyOnChange, req.ScanType, nil, autoAddIPs, discoverDNS, req.DNSOverwrite)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(job)
}

// GetScanJob handles GET /api/v1/admin/scan-jobs/:id
func (h *Handler) GetScanJob(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid job ID")
	}
	job, err := h.ops.Discovery.GetJob(c.Context(), int64(id))
	if err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "scan job not found")
	}
	return c.JSON(job)
}

// UpdateScanJob handles PUT /api/v1/admin/scan-jobs/:id
func (h *Handler) UpdateScanJob(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid job ID")
	}
	var req UpdateScanJobRequest
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if len(req.SubnetIDs) == 0 && req.Subnet != "" {
		sn, snErr := h.service.GetRepository().GetSubnetByCIDR(c.Context(), req.Subnet)
		if snErr != nil {
			return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "subnet not found for CIDR: " + req.Subnet)
		}
		req.SubnetIDs = []int64{sn.ID}
	}
	autoAddIPsUpdate := true
	if req.AutoAddIPs != nil {
		autoAddIPsUpdate = *req.AutoAddIPs
	}
	discoverDNSUpdate := true
	if req.DiscoverDNS != nil {
		discoverDNSUpdate = *req.DiscoverDNS
	}
	job, err := h.ops.Discovery.UpdateJobFull(c.Context(), int64(id), req.Name, req.SubnetIDs, req.ScheduleCron, req.IsActive, req.PingConcurrency, req.NotifyOnChange, req.ScanType, req.AgentID, autoAddIPsUpdate, discoverDNSUpdate, req.DNSOverwrite)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	return c.JSON(job)
}

// DeleteScanJob handles DELETE /api/v1/admin/scan-jobs/:id
func (h *Handler) DeleteScanJob(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid job ID")
	}
	if err := h.ops.Discovery.DeleteJob(c.Context(), int64(id)); err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to delete scan job")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// RunScanJobNow handles POST /api/v1/admin/scan-jobs/:id/run
func (h *Handler) RunScanJobNow(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid job ID")
	}
	job, err := h.ops.Discovery.GetJob(c.Context(), int64(id))
	if err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "scan job not found")
	}
	bgJob := h.ops.Jobs.Enqueue("scan", "Run scan job "+job.Name, fiber.Map{"scan_job_id": job.ID}, 1, func(ctx context.Context, reporter *services.JobReporter) (interface{}, error) {
		reporter.Progress(0, 1, "running scan")
		if err := h.ops.Discovery.RunJob(ctx, job); err != nil {
			slog.Error("scan job run error", "job_id", id, "error", err)
			return nil, err
		}
		reporter.Progress(1, 1, "scan complete")
		return fiber.Map{"scan_job_id": job.ID}, nil
	})
	return c.Status(fiber.StatusAccepted).JSON(bgJob)
}

// GetScanJobStatus handles GET /api/v1/admin/scan-jobs/:id/status
func (h *Handler) GetScanJobStatus(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid job ID")
	}
	return c.JSON(fiber.Map{"running": h.ops.Discovery.IsRunning(int64(id))})
}

// GetScanJobResults handles GET /api/v1/admin/scan-jobs/:id/results
func (h *Handler) GetScanJobResults(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid job ID")
	}
	limit := c.QueryInt("limit", 100)
	results, err := h.ops.Discovery.ListResults(c.Context(), int64(id), limit)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(results)
}

// GetSubnetScanResults handles GET /api/v1/subnets/:id/scan-results
func (h *Handler) GetSubnetScanResults(c *fiber.Ctx) error {
	if !h.requirePerm(c, "subnets:read") {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid subnet ID")
	}
	limit := c.QueryInt("limit", 100)
	results, err := h.ops.Discovery.ListSubnetResults(c.Context(), int64(id), limit)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(results)
}

// GetObservedState handles GET /api/v1/discovery/observed?resource_type=ip_address&resource_id=42
func (h *Handler) GetObservedState(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	resourceType := c.Query("resource_type", "ip_address")
	resourceID := c.QueryInt("resource_id")
	if resourceID <= 0 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "resource_id query param required")
	}
	state, err := h.ops.Discovery.GetObservedState(c.Context(), resourceType, int64(resourceID))
	if err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "observed state not found")
	}
	return c.JSON(state)
}

// ListUnregisteredHosts handles GET /api/v1/discovery/unregistered
// Returns IPs seen by the scanner that do not match any authoritative ip_addresses record.
func (h *Handler) ListUnregisteredHosts(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	hosts, err := h.ops.Discovery.ListUnregisteredHosts(c.Context(), orgIDFromCtx(c))
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	if hosts == nil {
		hosts = []*models.ObservedState{}
	}
	return c.JSON(hosts)
}
