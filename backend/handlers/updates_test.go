package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"padduck/models"
)

// CheckForUpdates is gated by requireAdmin; immediately after the guard it
// touches h.service.Config with a nil service, so only the guard branches
// are testable here without a DB. The pure helper functions below have no
// service/repo dependency and are fully testable.

func TestCheckForUpdates_NoUser_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/updates/check", h.CheckForUpdates)

	req := httptest.NewRequest("GET", "/admin/updates/check", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestCheckForUpdates_NonAdmin_Returns403(t *testing.T) {
	h := minHandler()
	app := fiber.New()
	app.Get("/admin/updates/check", func(c *fiber.Ctx) error {
		c.Locals("user", &models.User{Role: "user"})
		return h.CheckForUpdates(c)
	})

	req := httptest.NewRequest("GET", "/admin/updates/check", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestFirstNonEmpty(t *testing.T) {
	assert.Equal(t, "a", firstNonEmpty("a", "b"))
	assert.Equal(t, "b", firstNonEmpty("", "b"))
	assert.Equal(t, "b", firstNonEmpty("   ", "b"))
	assert.Equal(t, "", firstNonEmpty())
	assert.Equal(t, "", firstNonEmpty("", "  "))
	assert.Equal(t, "c", firstNonEmpty("", "", "c"))
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		name    string
		current string
		latest  string
		want    int
	}{
		{"equal", "1.2.3", "1.2.3", 0},
		{"equal with v prefix", "v1.2.3", "1.2.3", 0},
		{"current older patch", "1.2.3", "1.2.4", -1},
		{"current newer patch", "1.2.4", "1.2.3", 1},
		{"current older minor", "1.2.3", "1.3.0", -1},
		{"current older major", "1.2.3", "2.0.0", -1},
		{"different length, current shorter", "1.2", "1.2.1", -1},
		{"different length, current longer", "1.2.1", "1.2", 1},
		{"dev/pre-release suffix ignored on latest", "1.2.3", "1.2.4-rc1", -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, compareVersions(tc.current, tc.latest))
		})
	}
}

func TestVersionParts(t *testing.T) {
	assert.Equal(t, []int{1, 2, 3}, versionParts("1.2.3"))
	assert.Equal(t, []int{1, 2, 3}, versionParts("v1.2.3"))
	assert.Equal(t, []int{1, 2, 3}, versionParts(" V1.2.3 "))
	assert.Equal(t, []int{1, 2, 0}, versionParts("1.2.0-rc1"))
	assert.Equal(t, []int{1, 2, 0}, versionParts("1.2.0+build5"))
	assert.Equal(t, []int{0}, versionParts("not-a-version"))
	assert.Equal(t, []int{1, 0, 3}, versionParts("1..3"))
}
