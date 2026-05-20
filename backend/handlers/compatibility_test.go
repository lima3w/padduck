package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestV2CompatibilityWarningsCatalog(t *testing.T) {
	warnings := v2CompatibilityWarnings()

	assert.NotEmpty(t, warnings)
	ids := map[string]bool{}
	areas := map[string]bool{}
	for _, warning := range warnings {
		assert.NotEmpty(t, warning.ID)
		assert.False(t, ids[warning.ID], "duplicate warning id %s", warning.ID)
		ids[warning.ID] = true
		assert.NotEmpty(t, warning.Area)
		assert.NotEmpty(t, warning.Severity)
		assert.NotEmpty(t, warning.Summary)
		assert.NotEmpty(t, warning.V1Surface)
		assert.NotEmpty(t, warning.V2Change)
		assert.NotEmpty(t, warning.RecommendedWork)
		areas[warning.Area] = true
	}

	assert.True(t, areas["api"])
	assert.True(t, areas["field"])
	assert.True(t, areas["workflow"])
}

func TestCompatibilitySummary(t *testing.T) {
	warnings := v2CompatibilityWarnings()
	summary := compatibilityWarningSummary(warnings)

	assert.Equal(t, len(warnings), summary["total"])
	assert.NotEmpty(t, summary["by_severity"])
	assert.NotEmpty(t, summary["by_area"])
}

func TestCompatibilityDeprecationReport(t *testing.T) {
	warnings := v2CompatibilityWarnings()
	report := compatibilityDeprecationReport(warnings)

	assert.Len(t, report, len(warnings))
	assert.Equal(t, warnings[0].ID, report[0].ID)
	assert.NotEmpty(t, report[0].Impacted)
	assert.Equal(t, warnings[0].RecommendedWork, report[0].RecommendedWork)
}

func TestCompatibilityCheckSummary(t *testing.T) {
	checks := []compatibilityCheck{
		{Area: "schema", Status: "pass"},
		{Area: "config", Status: "warn"},
		{Area: "tokens", Status: "fail"},
	}

	summary := compatibilityCheckSummary(checks)

	assert.Equal(t, len(checks), summary["total"])
	assert.Equal(t, false, summary["ready"])
	assert.NotEmpty(t, summary["by_status"])
	assert.NotEmpty(t, summary["by_area"])
}
