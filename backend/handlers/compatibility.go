package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"padduck/config"
	"padduck/services"
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

type compatibilityCheck struct {
	ID              string   `json:"id"`
	Area            string   `json:"area"`
	Status          string   `json:"status"`
	Summary         string   `json:"summary"`
	Detail          string   `json:"detail"`
	RecommendedWork string   `json:"recommended_work,omitempty"`
	Signals         []string `json:"signals,omitempty"`
}

type deprecationReportItem struct {
	ID              string   `json:"id"`
	Area            string   `json:"area"`
	Severity        string   `json:"severity"`
	Summary         string   `json:"summary"`
	Detail          string   `json:"detail"`
	V1Surface       string   `json:"v1_surface"`
	V2Change        string   `json:"v2_change"`
	RecommendedWork string   `json:"recommended_work"`
	Impacted        []string `json:"impacted,omitempty"`
	DocsURL         string   `json:"docs_url,omitempty"`
}

func v2CompatibilityWarnings() []compatibilityWarning {
	return []compatibilityWarning{
		{
			ID:              "api-ip-address-nesting",
			Area:            "api",
			Severity:        "warning",
			Summary:         "Nested IP address routes remain available for v1 compatibility.",
			Detail:          "Clients should prefer top-level IP address endpoints before upgrading to v2.",
			V1Surface:       "/api/v1/networks/{networkID}/subnets/{subnetID}/ip-addresses",
			V2Change:        "v2 will standardize IP address collection operations under /api/v2/ip-addresses with explicit subnet filters.",
			RecommendedWork: "Update clients to call top-level IP address endpoints and pass subnet IDs explicitly.",
			DocsURL:         "/docs/api-contract.md",
			APIs: []string{
				"GET /api/v1/networks/{networkID}/subnets/{subnetID}/ip-addresses",
				"POST /api/v1/networks/{networkID}/subnets/{subnetID}/ip-addresses",
			},
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

func compatibilityWarningSummary(warnings []compatibilityWarning) fiber.Map {
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

func compatibilityCheckSummary(checks []compatibilityCheck) fiber.Map {
	byStatus := fiber.Map{}
	byArea := fiber.Map{}
	for _, check := range checks {
		byStatus[check.Status] = asInt(byStatus[check.Status]) + 1
		byArea[check.Area] = asInt(byArea[check.Area]) + 1
	}
	return fiber.Map{
		"total":     len(checks),
		"by_status": byStatus,
		"by_area":   byArea,
		"ready":     asInt(byStatus["fail"]) == 0,
	}
}

func compatibilityDeprecationReport(warnings []compatibilityWarning) []deprecationReportItem {
	report := make([]deprecationReportItem, 0, len(warnings))
	for _, warning := range warnings {
		impacted := append([]string{}, warning.APIs...)
		impacted = append(impacted, warning.Fields...)
		impacted = append(impacted, warning.Workflows...)
		report = append(report, deprecationReportItem{
			ID:              warning.ID,
			Area:            warning.Area,
			Severity:        warning.Severity,
			Summary:         warning.Summary,
			Detail:          warning.Detail,
			V1Surface:       warning.V1Surface,
			V2Change:        warning.V2Change,
			RecommendedWork: warning.RecommendedWork,
			Impacted:        impacted,
			DocsURL:         warning.DocsURL,
		})
	}
	return report
}

func asInt(v any) int {
	if n, ok := v.(int); ok {
		return n
	}
	return 0
}

func (h *Handler) v2MigrationReadinessChecks(ctx context.Context) []compatibilityCheck {
	checks := []compatibilityCheck{}

	requiredTables := []string{
		"schema_migrations",
		"configs",
		"networks",
		"subnets",
		"ip_addresses",
		"roles",
		"role_permissions",
		"user_roles",
		"api_tokens",
		"webhook_endpoints",
		"custom_fields",
	}
	missingTables := []string{}
	for _, table := range requiredTables {
		ok, err := h.compatTableExists(ctx, table)
		if err != nil || !ok {
			missingTables = append(missingTables, table)
		}
	}
	schemaStatus := "pass"
	schemaDetail := "Core v1 schema tables needed by migration tooling are present."
	schemaWork := ""
	if len(missingTables) > 0 {
		schemaStatus = "fail"
		schemaDetail = "Missing required schema tables: " + strings.Join(missingTables, ", ") + "."
		schemaWork = "Run pending v1 database migrations before exporting or importing a v2 bundle."
	}
	checks = append(checks, compatibilityCheck{
		ID:              "schema-core-tables",
		Area:            "schema",
		Status:          schemaStatus,
		Summary:         "Core schema tables",
		Detail:          schemaDetail,
		RecommendedWork: schemaWork,
		Signals:         []string{fmt.Sprintf("required_tables=%d", len(requiredTables)), fmt.Sprintf("missing_tables=%d", len(missingTables))},
	})

	configCount := h.compatCount(ctx, "SELECT COUNT(*) FROM configs")
	appURL, _ := h.compatConfig(ctx, "app_url")
	configSignals := []string{fmt.Sprintf("config_entries=%d", configCount), "app_url_configured=false"}
	configStatus := "pass"
	configDetail := "Required operational configuration is present."
	configWork := ""
	if strings.TrimSpace(appURL) != "" {
		configSignals[1] = "app_url_configured=true"
	} else {
		configStatus = "warn"
		configDetail = "Application URL is not configured; exported references may need manual adjustment."
		configWork = "Set app_url to the externally reachable v1 URL before preparing the migration bundle."
	}
	if h.isProduction && !config.HasPersistentMFAKey() {
		configStatus = "fail"
		configDetail = "Production MFA encryption key is not present in the runtime environment or persistent key file."
		configWork = "Set MFA_ENCRYPTION_KEY or restart the backend so it can create data/mfa-encryption-key."
	}
	checks = append(checks, compatibilityCheck{
		ID:              "config-runtime",
		Area:            "config",
		Status:          configStatus,
		Summary:         "Runtime configuration",
		Detail:          configDetail,
		RecommendedWork: configWork,
		Signals:         configSignals,
	})

	pdnsEnabled, _ := h.compatConfig(ctx, "pdns_enabled")
	pdnsURL, _ := h.compatConfig(ctx, "pdns_api_url")
	technitiumURL, _ := h.compatConfig(ctx, "technitium_url")
	integrationStatus := "pass"
	integrationDetail := "No incomplete enabled integration configuration was detected."
	integrationWork := ""
	if pdnsEnabled == "true" && strings.TrimSpace(pdnsURL) == "" {
		integrationStatus = "warn"
		integrationDetail = "PowerDNS is enabled but no API URL is configured."
		integrationWork = "Complete PowerDNS settings or disable the integration before migration."
	}
	checks = append(checks, compatibilityCheck{
		ID:              "integrations-configured",
		Area:            "integrations",
		Status:          integrationStatus,
		Summary:         "Integration configuration",
		Detail:          integrationDetail,
		RecommendedWork: integrationWork,
		Signals: []string{
			fmt.Sprintf("pdns_enabled=%t", pdnsEnabled == "true"),
			fmt.Sprintf("pdns_url_configured=%t", strings.TrimSpace(pdnsURL) != ""),
			fmt.Sprintf("technitium_configured=%t", strings.TrimSpace(technitiumURL) != ""),
			"update_check_configured=true",
		},
	})

	customFieldCount := h.compatCount(ctx, "SELECT COUNT(*) FROM custom_fields")
	checks = append(checks, compatibilityCheck{
		ID:      "custom-fields-exportable",
		Area:    "custom_fields",
		Status:  "pass",
		Summary: "Custom field inventory",
		Detail:  "Custom field definitions can be included in migration planning.",
		Signals: []string{fmt.Sprintf("custom_fields=%d", customFieldCount)},
	})

	customRoleCount := h.compatCount(ctx, "SELECT COUNT(*) FROM roles WHERE is_system = false")
	unassignedUserCount := h.compatCount(ctx, "SELECT COUNT(*) FROM users u WHERE NOT EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = u.id)")
	roleStatus := "pass"
	roleDetail := "Role assignments are ready for migration review."
	roleWork := ""
	if unassignedUserCount > 0 {
		roleStatus = "warn"
		roleDetail = "Some users only have legacy role-column access and no explicit role assignment."
		roleWork = "Assign explicit RBAC roles to all active users before v2 migration."
	}
	checks = append(checks, compatibilityCheck{
		ID:              "roles-explicit-assignments",
		Area:            "roles",
		Status:          roleStatus,
		Summary:         "Role assignments",
		Detail:          roleDetail,
		RecommendedWork: roleWork,
		Signals:         []string{fmt.Sprintf("custom_roles=%d", customRoleCount), fmt.Sprintf("users_without_role_assignments=%d", unassignedUserCount)},
	})

	tokenCount := h.compatCount(ctx, "SELECT COUNT(*) FROM api_tokens")
	tokensWithoutExpiry := h.compatCount(ctx, "SELECT COUNT(*) FROM api_tokens WHERE expires_at IS NULL")
	tokenStatus := "pass"
	tokenDetail := "API access records have explicit expiry metadata."
	tokenWork := ""
	if tokensWithoutExpiry > 0 {
		tokenStatus = "warn"
		tokenDetail = "Some API access records have no expiration date."
		tokenWork = "Rotate or extend API access records with an explicit expiry before migration."
	}
	checks = append(checks, compatibilityCheck{
		ID:              "tokens-expiring",
		Area:            "tokens",
		Status:          tokenStatus,
		Summary:         "API token lifecycle",
		Detail:          tokenDetail,
		RecommendedWork: tokenWork,
		Signals:         []string{fmt.Sprintf("api_tokens=%d", tokenCount), fmt.Sprintf("tokens_without_expiry=%d", tokensWithoutExpiry)},
	})

	activeWebhooks := h.compatCount(ctx, "SELECT COUNT(*) FROM webhook_endpoints WHERE is_active = true")
	unfilteredWebhooks := h.compatCount(ctx, "SELECT COUNT(*) FROM webhook_endpoints WHERE is_active = true AND (events IS NULL OR cardinality(events) = 0)")
	webhookStatus := "pass"
	webhookDetail := "Active webhook subscriptions have explicit event filters."
	webhookWork := ""
	if unfilteredWebhooks > 0 {
		webhookStatus = "warn"
		webhookDetail = "Some active webhook subscriptions have no explicit event filter."
		webhookWork = "Review webhook subscriptions and record the intended v2 event schemas."
	}
	checks = append(checks, compatibilityCheck{
		ID:              "webhooks-explicit-events",
		Area:            "webhooks",
		Status:          webhookStatus,
		Summary:         "Webhook subscriptions",
		Detail:          webhookDetail,
		RecommendedWork: webhookWork,
		Signals:         []string{fmt.Sprintf("active_webhooks=%d", activeWebhooks), fmt.Sprintf("unfiltered_webhooks=%d", unfilteredWebhooks)},
	})

	return checks
}

func (h *Handler) compatTableExists(ctx context.Context, table string) (bool, error) {
	var exists bool
	err := h.service.GetRepository().GetPool().QueryRow(ctx, "SELECT to_regclass($1) IS NOT NULL", table).Scan(&exists)
	return exists, err
}

func (h *Handler) compatCount(ctx context.Context, query string, args ...any) int {
	var count int
	if err := h.service.GetRepository().GetPool().QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0
	}
	return count
}

func (h *Handler) compatConfig(ctx context.Context, key string) (string, bool) {
	value, err := h.service.Config.GetCtx(ctx, key)
	if err != nil {
		return "", false
	}
	return value, true
}

// GetV2CompatibilityWarnings handles GET /api/v1/admin/compatibility/v2-warnings.
func (h *Handler) GetV2CompatibilityWarnings(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}

	warnings := v2CompatibilityWarnings()
	return c.JSON(fiber.Map{
		"target_version": "v2.0",
		"generated_at":   time.Now().UTC().Format(time.RFC3339),
		"summary":        compatibilityWarningSummary(warnings),
		"warnings":       warnings,
	})
}

// GetV2MigrationReadiness handles GET /api/v1/admin/compatibility/v2-readiness.
func (h *Handler) GetV2MigrationReadiness(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}

	checks := h.v2MigrationReadinessChecks(c.Context())
	return c.JSON(fiber.Map{
		"target_version": "v2.0",
		"generated_at":   time.Now().UTC().Format(time.RFC3339),
		"summary":        compatibilityCheckSummary(checks),
		"checks":         checks,
	})
}

// GetV2DeprecationReport handles GET /api/v1/admin/compatibility/deprecations.
func (h *Handler) GetV2DeprecationReport(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminRead) {
		return nil
	}

	warnings := v2CompatibilityWarnings()
	return c.JSON(fiber.Map{
		"target_version": "v2.0",
		"generated_at":   time.Now().UTC().Format(time.RFC3339),
		"summary":        compatibilityWarningSummary(warnings),
		"deprecations":   compatibilityDeprecationReport(warnings),
	})
}
