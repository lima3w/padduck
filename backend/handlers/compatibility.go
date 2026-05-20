package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"ipam-next/services"
)

type compatibilityWarning struct {
	ID              string   `json:"id"`
	Area            string   `json:"area"`
	Severity        string   `json:"severity"`
	Summary         string   `json:"summary"`
	Detail          string   `json:"detail"`
	V1Surface       string   `json:"v1_surface"`
	V2Change        string   `json:"v2_change"`
	RecommendedWork string   `json:"recommended_work"`
	DocsURL         string   `json:"docs_url,omitempty"`
	APIs            []string `json:"apis,omitempty"`
	Fields          []string `json:"fields,omitempty"`
	Workflows       []string `json:"workflows,omitempty"`
}

func v2CompatibilityWarnings() []compatibilityWarning {
	return []compatibilityWarning{
		{
			ID:              "api-ip-address-nesting",
			Area:            "api",
			Severity:        "warning",
			Summary:         "Nested IP address routes remain available for v1 compatibility.",
			Detail:          "Clients should prefer top-level IP address endpoints before upgrading to v2.",
			V1Surface:       "/api/v1/sections/{sectionID}/subnets/{subnetID}/ip-addresses",
			V2Change:        "v2 will standardize IP address collection operations under /api/v2/ip-addresses with explicit subnet filters.",
			RecommendedWork: "Update clients to call top-level IP address endpoints and pass subnet IDs explicitly.",
			DocsURL:         "/docs/api-contract.md",
			APIs: []string{
				"GET /api/v1/sections/{sectionID}/subnets/{subnetID}/ip-addresses",
				"POST /api/v1/sections/{sectionID}/subnets/{subnetID}/ip-addresses",
			},
		},
		{
			ID:              "api-british-utilisation-spelling",
			Area:            "api",
			Severity:        "info",
			Summary:         "Utilisation history uses British spelling in v1.",
			Detail:          "The v1 route is retained, but v2 API naming will use utilization consistently.",
			V1Surface:       "/api/v1/subnets/{id}/utilisation/history",
			V2Change:        "v2 will expose utilization history with American spelling and compatibility aliases may be time-limited.",
			RecommendedWork: "Prefer generated clients or route constants so spelling aliases can be migrated centrally.",
			DocsURL:         "/docs/api-contract.md",
			APIs:            []string{"GET /api/v1/subnets/{id}/utilisation/history"},
		},
		{
			ID:              "workflow-automation-idempotency",
			Area:            "workflow",
			Severity:        "warning",
			Summary:         "Automation writes should send idempotency keys.",
			Detail:          "v1 accepts automation write requests without an Idempotency-Key header, but retry-safe workflows need one.",
			V1Surface:       "/api/v1/automation/* write endpoints",
			V2Change:        "v2 may require Idempotency-Key on automation allocation, reservation, release, DNS update, and device registration requests.",
			RecommendedWork: "Add stable Idempotency-Key values to external automation workflow retries.",
			DocsURL:         "/docs/api-client-examples.md",
			Workflows:       []string{"automation allocation", "automation DNS update", "automation device registration"},
		},
		{
			ID:              "field-legacy-role",
			Area:            "field",
			Severity:        "warning",
			Summary:         "Legacy user role fields are compatibility fallbacks.",
			Detail:          "v1 still exposes the user role column while custom roles and scoped permissions are available.",
			V1Surface:       "users.role",
			V2Change:        "v2 will emphasize role assignments and permission scopes over the legacy role column.",
			RecommendedWork: "Move integrations to role assignment and permission endpoints instead of depending only on users.role.",
			DocsURL:         "/docs/user-guide.md",
			Fields:          []string{"users.role"},
		},
		{
			ID:              "workflow-feature-flags",
			Area:            "workflow",
			Severity:        "info",
			Summary:         "Optional modules can be disabled by administrators.",
			Detail:          "v1 returns 404 for disabled feature modules. v2 clients should discover feature state before rendering module workflows.",
			V1Surface:       "/api/v1/features and feature-gated resources",
			V2Change:        "v2 will keep explicit feature discovery as the supported compatibility contract.",
			RecommendedWork: "Call /api/v1/features at startup and hide disabled module actions before making writes.",
			DocsURL:         "/docs/user-guide.md",
			APIs:            []string{"GET /api/v1/features"},
		},
	}
}

func compatibilitySummary(warnings []compatibilityWarning) fiber.Map {
	bySeverity := fiber.Map{}
	byArea := fiber.Map{}
	for _, warning := range warnings {
		bySeverity[warning.Severity] = asInt(bySeverity[warning.Severity]) + 1
		byArea[warning.Area] = asInt(byArea[warning.Area]) + 1
	}
	return fiber.Map{
		"total":       len(warnings),
		"by_severity": bySeverity,
		"by_area":     byArea,
	}
}

func asInt(v any) int {
	if n, ok := v.(int); ok {
		return n
	}
	return 0
}

// GetV2CompatibilityWarnings handles GET /api/v1/admin/compatibility/v2-warnings.
func (h *Handler) GetV2CompatibilityWarnings(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2AdminRead); err != nil {
		return nil
	}

	warnings := v2CompatibilityWarnings()
	return c.JSON(fiber.Map{
		"target_version": "v2.0",
		"generated_at":   time.Now().UTC().Format(time.RFC3339),
		"summary":        compatibilitySummary(warnings),
		"warnings":       warnings,
	})
}
