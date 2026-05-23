package handlers

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const agentBinaryPath = "./data/agent/padduck-agent"
const agentVersionPath = "./data/agent/version.txt"

// DownloadAgentBinary serves the pre-built agent binary from ./data/agent/padduck-agent.
// No authentication is required — the binary itself contains no secrets.
// Callers embed the agent token in their environment at runtime.
//
// GET /api/v1/agent/download
func (h *Handler) DownloadAgentBinary(c *fiber.Ctx) error {
	if _, err := os.Stat(agentBinaryPath); os.IsNotExist(err) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "agent binary not available — place the padduck-agent binary at ./data/agent/padduck-agent on the server",
		})
	}

	c.Set("Content-Disposition", `attachment; filename="padduck-agent"`)
	c.Set("Content-Type", "application/octet-stream")
	return c.SendFile(agentBinaryPath)
}

// GetAgentBinaryVersion returns the version of the hosted agent binary.
//
// GET /api/v1/agent/version
func (h *Handler) GetAgentBinaryVersion(c *fiber.Ctx) error {
	version := "unknown"

	if data, err := os.ReadFile(agentVersionPath); err == nil {
		v := strings.TrimSpace(string(data))
		if v != "" {
			version = v
		}
	}

	available := true
	if _, err := os.Stat(agentBinaryPath); os.IsNotExist(err) {
		available = false
		version = ""
	}

	// Also expose the absolute path for debugging (relative to CWD)
	absPath, _ := filepath.Abs(agentBinaryPath)
	_ = absPath

	return c.JSON(fiber.Map{
		"available": available,
		"version":   version,
	})
}
