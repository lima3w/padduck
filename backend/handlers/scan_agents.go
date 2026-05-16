package handlers

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
)

// ListScanAgents handles GET /api/v1/admin/scan-agents
func (h *Handler) ListScanAgents(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}
	agents, err := h.service.Discovery.ListAgents(c.Context())
	if err != nil {
		log.Printf("Error listing scan agents: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(agents)
}

// CreateScanAgent handles POST /api/v1/admin/scan-agents
func (h *Handler) CreateScanAgent(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	agent, rawToken, err := h.service.Discovery.CreateAgent(c.Context(), req.Name)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"agent": agent,
		"token": rawToken,
	})
}

// GetScanAgent handles GET /api/v1/admin/scan-agents/:id
func (h *Handler) GetScanAgent(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid agent ID"})
	}
	agent, err := h.service.Discovery.GetAgent(c.Context(), int64(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "scan agent not found"})
	}
	return c.JSON(agent)
}

// RotateScanAgentToken handles POST /api/v1/admin/scan-agents/:id/rotate-token
func (h *Handler) RotateScanAgentToken(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid agent ID"})
	}
	agent, rawToken, err := h.service.Discovery.RotateToken(c.Context(), int64(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to rotate token"})
	}
	return c.JSON(fiber.Map{
		"agent": agent,
		"token": rawToken,
	})
}

// DeleteScanAgent handles DELETE /api/v1/admin/scan-agents/:id
func (h *Handler) DeleteScanAgent(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminWrite); err != nil {
		return nil
	}
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid agent ID"})
	}
	if err := h.service.Discovery.DeleteAgent(c.Context(), int64(id)); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "scan agent not found"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Agent API endpoints (authenticated via Bearer token)
// ---------------------------------------------------------------------------

// agentFromContext extracts the authenticated agent from the Fiber context.
func agentFromContext(c *fiber.Ctx) (*models.ScanAgent, bool) {
	a, ok := c.Locals("scan_agent").(*models.ScanAgent)
	return a, ok
}

// AgentAuthMiddleware authenticates requests from scan agents via Bearer token.
func (h *Handler) AgentAuthMiddleware(c *fiber.Ctx) error {
	auth := c.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing or invalid authorization"})
	}
	rawToken := strings.TrimPrefix(auth, "Bearer ")
	agent, err := h.service.Discovery.AuthenticateAgent(c.Context(), rawToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or inactive agent token"})
	}
	c.Locals("scan_agent", agent)
	return c.Next()
}

// AgentGetJobs handles GET /api/v1/scan-agent/jobs
func (h *Handler) AgentGetJobs(c *fiber.Ctx) error {
	agent, ok := agentFromContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}
	jobs, err := h.service.Discovery.GetJobsForAgent(c.Context(), agent.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(jobs)
}

// AgentPostResults handles POST /api/v1/scan-agent/results
func (h *Handler) AgentPostResults(c *fiber.Ctx) error {
	agent, ok := agentFromContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}
	var req struct {
		JobID   int64                      `json:"job_id"`
		Results []services.AgentScanResult `json:"results"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.JobID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "job_id is required"})
	}
	if err := h.service.Discovery.AcceptAgentResults(c.Context(), agent.ID, req.JobID, req.Results); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "results accepted"})
}

// AgentHeartbeat handles POST /api/v1/scan-agent/heartbeat
func (h *Handler) AgentHeartbeat(c *fiber.Ctx) error {
	agent, ok := agentFromContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}
	if err := h.service.Discovery.HeartbeatAgent(c.Context(), agent.ID); err != nil {
		log.Printf("agent heartbeat error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(fiber.Map{"message": "ok"})
}
