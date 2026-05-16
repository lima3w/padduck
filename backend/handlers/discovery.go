package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
)

type CreateScanJobRequest struct {
	Name         string  `json:"name"`
	SubnetIDs    []int64 `json:"subnet_ids"`
	ScheduleCron *string `json:"schedule_cron,omitempty"`
}

type UpdateScanJobRequest struct {
	Name            string  `json:"name"`
	SubnetIDs       []int64 `json:"subnet_ids"`
	ScheduleCron    *string `json:"schedule_cron,omitempty"`
	IsActive        bool    `json:"is_active"`
	PingConcurrency int     `json:"ping_concurrency,omitempty"`
	NotifyOnChange  bool    `json:"notify_on_change,omitempty"`
	ScanType        string  `json:"scan_type,omitempty"`
	AgentID         *int64  `json:"agent_id,omitempty"`
}

// ListScanJobs handles GET /api/v1/admin/scan-jobs
func (h *Handler) ListScanJobs(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}
	jobs, err := h.service.Discovery.ListJobs(c.Context())
	if err != nil {
		log.Printf("Error listing scan jobs: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(jobs)
}

// CreateScanJob handles POST /api/v1/admin/scan-jobs
func (h *Handler) CreateScanJob(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	user, _ := c.Locals("user").(*models.User)

	var req CreateScanJobRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	job, err := h.service.Discovery.CreateJob(c.Context(), req.Name, req.SubnetIDs, req.ScheduleCron, user.ID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(job)
}

// GetScanJob handles GET /api/v1/admin/scan-jobs/:id
func (h *Handler) GetScanJob(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid job ID"})
	}
	job, err := h.service.Discovery.GetJob(c.Context(), int64(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "scan job not found"})
	}
	return c.JSON(job)
}

// UpdateScanJob handles PUT /api/v1/admin/scan-jobs/:id
func (h *Handler) UpdateScanJob(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid job ID"})
	}
	var req UpdateScanJobRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	job, err := h.service.Discovery.UpdateJobFull(c.Context(), int64(id), req.Name, req.SubnetIDs, req.ScheduleCron, req.IsActive, req.PingConcurrency, req.NotifyOnChange, req.ScanType, req.AgentID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(job)
}

// DeleteScanJob handles DELETE /api/v1/admin/scan-jobs/:id
func (h *Handler) DeleteScanJob(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid job ID"})
	}
	if err := h.service.Discovery.DeleteJob(c.Context(), int64(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete scan job"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// RunScanJobNow handles POST /api/v1/admin/scan-jobs/:id/run
func (h *Handler) RunScanJobNow(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid job ID"})
	}
	job, err := h.service.Discovery.GetJob(c.Context(), int64(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "scan job not found"})
	}
	go func() {
		if err := h.service.Discovery.RunJob(c.Context(), job); err != nil {
			log.Printf("Scan job %d run error: %v", id, err)
		}
	}()
	return c.JSON(fiber.Map{"message": "scan job started"})
}

// GetScanJobResults handles GET /api/v1/admin/scan-jobs/:id/results
func (h *Handler) GetScanJobResults(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid job ID"})
	}
	limit := c.QueryInt("limit", 100)
	results, err := h.service.Discovery.ListResults(c.Context(), int64(id), limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(results)
}

// GetSubnetScanResults handles GET /api/v1/subnets/:id/scan-results
func (h *Handler) GetSubnetScanResults(c *fiber.Ctx) error {
	if err := h.permCheck(c, "subnets:read"); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid subnet ID"})
	}
	limit := c.QueryInt("limit", 100)
	results, err := h.service.Discovery.ListSubnetResults(c.Context(), int64(id), limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(results)
}
