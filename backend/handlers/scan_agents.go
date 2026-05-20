package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

// ListScanAgents handles GET /api/v1/admin/scan-agents
func (h *Handler) ListScanAgents(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}
	agents, err := h.service.Discovery.ListAgents(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing scan agents", "error", err)
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

// agentSubnetInfo is the per-subnet payload delivered to a scan agent.
type agentSubnetInfo struct {
	ID   int64  `json:"id"`
	CIDR string `json:"cidr"`
}

// agentJobResponse enriches a ScanJob with the CIDRs the agent needs to scan.
type agentJobResponse struct {
	ID              int64             `json:"id"`
	Name            string            `json:"name"`
	Subnets         []agentSubnetInfo `json:"subnets"`
	PingConcurrency int               `json:"ping_concurrency"`
	ScanType        string            `json:"scan_type"`
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

	// Enrich each job with subnet CIDRs so the agent can scan without extra round-trips.
	resp := make([]agentJobResponse, 0, len(jobs))
	for _, job := range jobs {
		r := agentJobResponse{
			ID:              job.ID,
			Name:            job.Name,
			PingConcurrency: job.PingConcurrency,
			ScanType:        job.ScanType,
		}
		for _, sid := range job.SubnetIDs {
			subnet, err := h.service.GetSubnet(c.Context(), sid)
			if err != nil {
				reqLogger(c).Warn("subnet not found for agent job", "subnet_id", sid, "job_id", job.ID, "error", err)
				continue
			}
			r.Subnets = append(r.Subnets, agentSubnetInfo{
				ID:   subnet.ID,
				CIDR: fmt.Sprintf("%s/%d", subnet.NetworkAddress, subnet.PrefixLength),
			})
		}
		resp = append(resp, r)
	}
	return c.JSON(resp)
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
	var req struct {
		Version      *string  `json:"version"`
		Capabilities []string `json:"capabilities"`
		Status       string   `json:"status"`
		LastError    *string  `json:"last_error"`
	}
	// parse body (ok if body is empty — all fields optional)
	_ = c.BodyParser(&req)
	if req.Status == "" {
		req.Status = "healthy"
	}
	if err := h.service.Discovery.HeartbeatAgent(c.Context(), agent.ID, req.Version, req.Capabilities, req.Status, req.LastError); err != nil {
		reqLogger(c).Error("agent heartbeat error", "agent_id", agent.ID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(fiber.Map{"message": "ok"})
}

// GetAgentHealthSummary handles GET /api/v1/admin/scan-agents/health
// Returns all agents with computed health status considering last_seen staleness.
func (h *Handler) GetAgentHealthSummary(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}
	agents, err := h.service.Discovery.ListAgents(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	// Compute effective status: if last_seen is >5 min ago, mark offline
	now := time.Now()
	for _, a := range agents {
		if a.LastSeen != nil && now.Sub(*a.LastSeen) > 5*time.Minute && a.Status != "offline" {
			a.Status = "offline"
		}
	}
	return c.JSON(agents)
}
