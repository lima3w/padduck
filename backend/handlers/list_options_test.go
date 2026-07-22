package handlers

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// parseListOptionsResult mirrors the tuple returned by parseListOptions so
// it can round-trip through JSON in the test route below.
type parseListOptionsResult struct {
	Page          int    `json:"page"`
	Limit         int    `json:"limit"`
	Sort          string `json:"sort"`
	Order         string `json:"order"`
	Query         string `json:"query"`
	Status        string `json:"status"`
	HideAvailable bool   `json:"hide_available"`
}

func parseListOptionsApp() *fiber.App {
	app := fiber.New()
	app.Get("/list-options", func(c *fiber.Ctx) error {
		page, limit, opts := parseListOptions(c)
		return c.JSON(parseListOptionsResult{
			Page: page, Limit: limit,
			Sort: opts.Sort, Order: opts.Order, Query: opts.Query,
			Status: opts.Status, HideAvailable: opts.HideAvailable,
		})
	})
	return app
}

func doParseListOptions(t *testing.T, query string) parseListOptionsResult {
	t.Helper()
	app := parseListOptionsApp()
	req := httptest.NewRequest("GET", "/list-options"+query, nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	data, _ := io.ReadAll(resp.Body)
	var out parseListOptionsResult
	assert.NoError(t, json.Unmarshal(data, &out))
	return out
}

func TestParseListOptions_Defaults(t *testing.T) {
	out := doParseListOptions(t, "")
	assert.Equal(t, 1, out.Page)
	assert.Equal(t, 25, out.Limit)
	assert.Equal(t, "asc", out.Order)
	assert.Equal(t, "", out.Sort)
	assert.Equal(t, "", out.Query)
	assert.Equal(t, "", out.Status)
	assert.False(t, out.HideAvailable)
}

func TestParseListOptions_PageBelowOneClampsToOne(t *testing.T) {
	out := doParseListOptions(t, "?page=0")
	assert.Equal(t, 1, out.Page)

	out = doParseListOptions(t, "?page=-5")
	assert.Equal(t, 1, out.Page)
}

func TestParseListOptions_LimitBelowOneClampsToTwentyFive(t *testing.T) {
	out := doParseListOptions(t, "?limit=0")
	assert.Equal(t, 25, out.Limit)

	out = doParseListOptions(t, "?limit=-1")
	assert.Equal(t, 25, out.Limit)
}

func TestParseListOptions_LimitAboveMaxClampsToMax(t *testing.T) {
	out := doParseListOptions(t, "?limit=5000")
	assert.Equal(t, maxLimit, out.Limit)
}

func TestParseListOptions_LimitAtMaxIsUnchanged(t *testing.T) {
	out := doParseListOptions(t, "?limit=1000")
	assert.Equal(t, 1000, out.Limit)
}

func TestParseListOptions_ValidPageAndLimitPassThrough(t *testing.T) {
	out := doParseListOptions(t, "?page=3&limit=50")
	assert.Equal(t, 3, out.Page)
	assert.Equal(t, 50, out.Limit)
}

func TestParseListOptions_SortAndOrderPassThrough(t *testing.T) {
	out := doParseListOptions(t, "?sort=name&order=desc")
	assert.Equal(t, "name", out.Sort)
	assert.Equal(t, "desc", out.Order)
}

func TestParseListOptions_OrderFallsBackToDirParam(t *testing.T) {
	out := doParseListOptions(t, "?dir=desc")
	assert.Equal(t, "desc", out.Order)
}

func TestParseListOptions_QueryFallsBackToSearchParam(t *testing.T) {
	out := doParseListOptions(t, "?search=printer")
	assert.Equal(t, "printer", out.Query)
}

func TestParseListOptions_QueryPrefersQOverSearch(t *testing.T) {
	out := doParseListOptions(t, "?q=switch&search=printer")
	assert.Equal(t, "switch", out.Query)
}

func TestParseListOptions_TrimsWhitespaceFromStringFields(t *testing.T) {
	out := doParseListOptions(t, "?sort=+name+&status=+active+")
	assert.Equal(t, "name", out.Sort)
	assert.Equal(t, "active", out.Status)
}

func TestParseListOptions_HideAvailableOnlyTrueWhenExactMatch(t *testing.T) {
	out := doParseListOptions(t, "?hide_available=true")
	assert.True(t, out.HideAvailable)

	out = doParseListOptions(t, "?hide_available=1")
	assert.False(t, out.HideAvailable)

	out = doParseListOptions(t, "?hide_available=TRUE")
	assert.False(t, out.HideAvailable)
}

func TestParseListOptions_NonNumericPageAndLimitFallBackToDefaults(t *testing.T) {
	out := doParseListOptions(t, "?page=abc&limit=xyz")
	assert.Equal(t, 1, out.Page)
	assert.Equal(t, 25, out.Limit)
}
