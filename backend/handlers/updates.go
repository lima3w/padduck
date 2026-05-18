package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/version"
)

type releaseInfo struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
}

// CheckForUpdates handles GET /api/v1/admin/updates/check.
func (h *Handler) CheckForUpdates(c *fiber.Ctx) error {
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok || currentUser.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "admin access required"})
	}

	enabled, _ := h.service.Config.GetCtx(c.Context(), "update_check_enabled")
	if enabled == "false" {
		return c.JSON(fiber.Map{
			"enabled":        false,
			"currentVersion": version.Version,
			"currentCommit":  version.Commit,
			"message":        "update checks are disabled",
		})
	}

	url, _ := h.service.Config.GetCtx(c.Context(), "update_check_url")
	if strings.TrimSpace(url) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "update_check_url is not configured"})
	}
	token, _ := h.service.Config.GetCtx(c.Context(), "update_check_token")

	req, err := http.NewRequestWithContext(c.Context(), http.MethodGet, url, nil) // #nosec G107 -- admin-configured update endpoint.
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid update_check_url"})
	}
	req.Header.Set("Accept", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "update check failed"})
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": fmt.Sprintf("update source returned %d", resp.StatusCode)})
	}

	var rel releaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "invalid update response"})
	}
	latest := firstNonEmpty(rel.TagName, rel.Name)
	if latest == "" {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "update response did not include a version"})
	}

	return c.JSON(fiber.Map{
		"enabled":         true,
		"currentVersion":  version.Version,
		"currentCommit":   version.Commit,
		"buildDate":       version.BuildDate,
		"latestVersion":   latest,
		"latestName":      rel.Name,
		"releaseUrl":      rel.HTMLURL,
		"publishedAt":     rel.PublishedAt,
		"updateAvailable": compareVersions(version.Version, latest) < 0,
	})
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func compareVersions(current, latest string) int {
	cur := versionParts(current)
	next := versionParts(latest)
	maxLen := len(cur)
	if len(next) > maxLen {
		maxLen = len(next)
	}
	for i := 0; i < maxLen; i++ {
		a, b := 0, 0
		if i < len(cur) {
			a = cur[i]
		}
		if i < len(next) {
			b = next[i]
		}
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
	}
	return 0
}

func versionParts(v string) []int {
	v = strings.TrimPrefix(strings.TrimSpace(strings.ToLower(v)), "v")
	if idx := strings.IndexAny(v, "-+"); idx >= 0 {
		v = v[:idx]
	}
	raw := strings.Split(v, ".")
	parts := make([]int, 0, len(raw))
	for _, part := range raw {
		if part == "" {
			parts = append(parts, 0)
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			return []int{0}
		}
		parts = append(parts, n)
	}
	return parts
}
