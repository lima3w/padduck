package handlers

import (
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// Both handlers are intentionally unauthenticated (see agent_binary.go) and
// depend only on the local filesystem, so their full behavior — including
// success paths — is testable without a DB. Paths are relative to the
// package's working directory during `go test`, so each test manages its
// own setup/teardown under ./data/agent and never runs in parallel with the
// others in this file to avoid clobbering shared fixture state.

func removeAgentDataDir(t *testing.T) {
	t.Helper()
	_ = os.RemoveAll("./data")
}

func writeAgentBinary(t *testing.T, content string) {
	t.Helper()
	assert.NoError(t, os.MkdirAll(filepath.Dir(agentBinaryPath), 0o755))
	assert.NoError(t, os.WriteFile(agentBinaryPath, []byte(content), 0o644))
}

func writeAgentVersion(t *testing.T, content string) {
	t.Helper()
	assert.NoError(t, os.MkdirAll(filepath.Dir(agentVersionPath), 0o755))
	assert.NoError(t, os.WriteFile(agentVersionPath, []byte(content), 0o644))
}

func TestDownloadAgentBinary_FileMissing_Returns404(t *testing.T) {
	removeAgentDataDir(t)
	t.Cleanup(func() { removeAgentDataDir(t) })

	h := minHandler()
	app := fiber.New()
	app.Get("/agent/download", h.DownloadAgentBinary)

	req := httptest.NewRequest("GET", "/agent/download", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestDownloadAgentBinary_FileExists_ReturnsFile(t *testing.T) {
	removeAgentDataDir(t)
	writeAgentBinary(t, "fake-binary-contents")
	t.Cleanup(func() { removeAgentDataDir(t) })

	h := minHandler()
	app := fiber.New()
	app.Get("/agent/download", h.DownloadAgentBinary)

	req := httptest.NewRequest("GET", "/agent/download", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Disposition"), `filename="padduck-agent"`)
	assert.Equal(t, "application/octet-stream", resp.Header.Get("Content-Type"))

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "fake-binary-contents", string(body))
}

func TestGetAgentBinaryVersion_NothingPresent_ReturnsUnavailableEmptyVersion(t *testing.T) {
	removeAgentDataDir(t)
	t.Cleanup(func() { removeAgentDataDir(t) })

	h := minHandler()
	app := fiber.New()
	app.Get("/agent/version", h.GetAgentBinaryVersion)

	req := httptest.NewRequest("GET", "/agent/version", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body := parseRolesResponse(t, resp.Body)
	assert.Equal(t, false, body["available"])
	assert.Equal(t, "", body["version"])
}

func TestGetAgentBinaryVersion_VersionFileWithoutBinary_ReportsUnavailableAndBlanksVersion(t *testing.T) {
	removeAgentDataDir(t)
	writeAgentVersion(t, "v1.2.3")
	t.Cleanup(func() { removeAgentDataDir(t) })

	h := minHandler()
	app := fiber.New()
	app.Get("/agent/version", h.GetAgentBinaryVersion)

	req := httptest.NewRequest("GET", "/agent/version", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)

	body := parseRolesResponse(t, resp.Body)
	// The handler blanks the version whenever the binary itself is missing,
	// even if a version file is present — this is intentional, not a bug.
	assert.Equal(t, false, body["available"])
	assert.Equal(t, "", body["version"])
}

func TestGetAgentBinaryVersion_BinaryAndVersionPresent_ReturnsBoth(t *testing.T) {
	removeAgentDataDir(t)
	writeAgentBinary(t, "fake-binary-contents")
	writeAgentVersion(t, " v1.2.3 \n")
	t.Cleanup(func() { removeAgentDataDir(t) })

	h := minHandler()
	app := fiber.New()
	app.Get("/agent/version", h.GetAgentBinaryVersion)

	req := httptest.NewRequest("GET", "/agent/version", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)

	body := parseRolesResponse(t, resp.Body)
	assert.Equal(t, true, body["available"])
	assert.Equal(t, "v1.2.3", body["version"])
}

func TestGetAgentBinaryVersion_BinaryPresentNoVersionFile_ReturnsUnknownVersion(t *testing.T) {
	removeAgentDataDir(t)
	writeAgentBinary(t, "fake-binary-contents")
	t.Cleanup(func() { removeAgentDataDir(t) })

	h := minHandler()
	app := fiber.New()
	app.Get("/agent/version", h.GetAgentBinaryVersion)

	req := httptest.NewRequest("GET", "/agent/version", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)

	body := parseRolesResponse(t, resp.Body)
	assert.Equal(t, true, body["available"])
	assert.Equal(t, "unknown", body["version"])
}

func TestGetAgentBinaryVersion_EmptyVersionFile_ReturnsUnknownVersion(t *testing.T) {
	removeAgentDataDir(t)
	writeAgentBinary(t, "fake-binary-contents")
	writeAgentVersion(t, "   \n")
	t.Cleanup(func() { removeAgentDataDir(t) })

	h := minHandler()
	app := fiber.New()
	app.Get("/agent/version", h.GetAgentBinaryVersion)

	req := httptest.NewRequest("GET", "/agent/version", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)

	body := parseRolesResponse(t, resp.Body)
	assert.Equal(t, true, body["available"])
	assert.Equal(t, "unknown", body["version"])
}
