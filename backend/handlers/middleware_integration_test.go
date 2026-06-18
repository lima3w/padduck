package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"padduck/internal/testdb"
	"padduck/models"
	"padduck/repository"
	"padduck/services"
	"padduck/utils"
)

const testHandlerMFAKey = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

// testAuthHandler returns a Handler wired to a real service and database,
// plus a fixture user holding a known password.
func testAuthHandler(t *testing.T) (*Handler, *services.Service, *repository.Repository, int64) {
	t.Helper()
	pool := testdb.Connect(t, "handlers")
	testdb.Truncate(t, pool, "sessions", "api_tokens", "users")
	repo := repository.NewRepository(pool)
	svc := services.NewService(repo, testHandlerMFAKey)
	h := NewHandler(svc, svc.Ops, svc.Auth, false)

	hash, err := utils.HashPassword("middleware-password")
	require.NoError(t, err)
	u, err := repo.CreateUserWithPassword(context.Background(), "mw-user", "mw@example.com", hash, "user")
	require.NoError(t, err)
	return h, svc, repo, u.ID
}

// probeApp mounts the middleware in front of a probe route that reports what
// the middleware put in the request context.
func probeApp(mw fiber.Handler) *fiber.App {
	app := fiber.New()
	app.Use(mw)
	probe := func(c *fiber.Ctx) error {
		out := fiber.Map{"authenticated": false}
		if u, ok := c.Locals("user").(*models.User); ok && u != nil {
			out["authenticated"] = true
			out["userID"] = u.ID
		}
		if scope, ok := c.Locals("tokenScope").(string); ok {
			out["scope"] = scope
		}
		return c.JSON(out)
	}
	app.Get("/probe", probe)
	app.Post("/probe", probe)
	app.Post("/api/v1/admin/probe", probe)
	return app
}

type probeResult struct {
	Authenticated bool   `json:"authenticated"`
	UserID        int64  `json:"userID"`
	Scope         string `json:"scope"`
}

func doProbe(t *testing.T, app *fiber.App, req *http.Request) (int, probeResult) {
	t.Helper()
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var out probeResult
	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &out))
	}
	return resp.StatusCode, out
}

func sessionRequest(method, path, token string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	return req
}

func bearerRequest(method, path, token string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func TestAuthMiddleware_SessionCookie_Integration(t *testing.T) {
	h, svc, _, userID := testAuthHandler(t)
	ctx := context.Background()
	app := probeApp(h.AuthMiddleware)

	token, err := svc.Ops.Identity.CreateWebSession(ctx, userID, "10.0.0.1", "test-browser")
	require.NoError(t, err)

	// Valid session: authenticated with the right user.
	status, out := doProbe(t, app, sessionRequest("GET", "/probe", token))
	assert.Equal(t, http.StatusOK, status)
	assert.True(t, out.Authenticated)
	assert.Equal(t, userID, out.UserID)

	// Garbage cookie with no Bearer fallback: 401.
	status, _ = doProbe(t, app, sessionRequest("GET", "/probe", "tampered-token"))
	assert.Equal(t, http.StatusUnauthorized, status)

	// Revoked session: 401.
	require.NoError(t, svc.Ops.Identity.RevokeAllSessions(ctx, userID))
	status, _ = doProbe(t, app, sessionRequest("GET", "/probe", token))
	assert.Equal(t, http.StatusUnauthorized, status)
}

func TestAuthMiddleware_BearerToken_Integration(t *testing.T) {
	h, svc, _, userID := testAuthHandler(t)
	ctx := context.Background()
	app := probeApp(h.AuthMiddleware)

	raw, err := svc.Ops.Identity.GenerateAPIToken(ctx, userID, "mw-token", "admin", 30)
	require.NoError(t, err)

	status, out := doProbe(t, app, bearerRequest("GET", "/probe", raw))
	assert.Equal(t, http.StatusOK, status)
	assert.True(t, out.Authenticated)
	assert.Equal(t, userID, out.UserID)
	assert.Equal(t, "admin", out.Scope)

	// Unknown token: 401.
	status, _ = doProbe(t, app, bearerRequest("GET", "/probe", "deadbeef"))
	assert.Equal(t, http.StatusUnauthorized, status)

	// A bad cookie must fall through to a valid Bearer token, not fail hard.
	req := bearerRequest("GET", "/probe", raw)
	req.AddCookie(&http.Cookie{Name: "session", Value: "tampered"})
	status, out = doProbe(t, app, req)
	assert.Equal(t, http.StatusOK, status)
	assert.True(t, out.Authenticated)
}

func TestAuthMiddleware_ExpiredToken_Integration(t *testing.T) {
	h, _, repo, userID := testAuthHandler(t)
	ctx := context.Background()
	app := probeApp(h.AuthMiddleware)

	// Insert a token that expired an hour ago, mirroring the service hashing.
	raw := "expired-raw-token"
	hash := sha256.Sum256([]byte(raw))
	past := time.Now().UTC().Add(-time.Hour)
	_, err := repo.CreateAPITokenFull(ctx, userID, hex.EncodeToString(hash[:]), "expired", "admin", &past)
	require.NoError(t, err)

	status, _ := doProbe(t, app, bearerRequest("GET", "/probe", raw))
	assert.Equal(t, http.StatusUnauthorized, status)
}

func TestAuthMiddleware_TokenScopes_Integration(t *testing.T) {
	h, svc, _, userID := testAuthHandler(t)
	ctx := context.Background()
	app := probeApp(h.AuthMiddleware)

	readToken, err := svc.Ops.Identity.GenerateAPIToken(ctx, userID, "ro", "read", 30)
	require.NoError(t, err)
	writeToken, err := svc.Ops.Identity.GenerateAPIToken(ctx, userID, "rw", "write", 30)
	require.NoError(t, err)
	adminToken, err := svc.Ops.Identity.GenerateAPIToken(ctx, userID, "adm", "admin", 30)
	require.NoError(t, err)

	// read scope: GET allowed, mutations forbidden.
	status, _ := doProbe(t, app, bearerRequest("GET", "/probe", readToken))
	assert.Equal(t, http.StatusOK, status)
	status, _ = doProbe(t, app, bearerRequest("POST", "/probe", readToken))
	assert.Equal(t, http.StatusForbidden, status)

	// write scope: mutations allowed, admin paths forbidden.
	status, _ = doProbe(t, app, bearerRequest("POST", "/probe", writeToken))
	assert.Equal(t, http.StatusOK, status)
	status, _ = doProbe(t, app, bearerRequest("POST", "/api/v1/admin/probe", writeToken))
	assert.Equal(t, http.StatusForbidden, status)

	// admin scope: admin paths allowed.
	status, _ = doProbe(t, app, bearerRequest("POST", "/api/v1/admin/probe", adminToken))
	assert.Equal(t, http.StatusOK, status)
}

func TestAuthMiddleware_RateLimit_Integration(t *testing.T) {
	h, svc, _, userID := testAuthHandler(t)
	ctx := context.Background()
	app := probeApp(h.AuthMiddleware)

	require.NoError(t, svc.Config.SetCtx(ctx, "api_token_rate_limit_per_minute", "2"))
	t.Cleanup(func() { _ = svc.Config.SetCtx(ctx, "api_token_rate_limit_per_minute", "100") })

	raw, err := svc.Ops.Identity.GenerateAPIToken(ctx, userID, "limited", "admin", 30)
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		status, _ := doProbe(t, app, bearerRequest("GET", "/probe", raw))
		require.Equal(t, http.StatusOK, status, "request %d within the limit", i+1)
	}
	resp, err := app.Test(bearerRequest("GET", "/probe", raw))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.Equal(t, "60", resp.Header.Get("Retry-After"))
}

func TestOptionalAuthMiddleware_Integration(t *testing.T) {
	h, svc, _, userID := testAuthHandler(t)
	ctx := context.Background()
	app := probeApp(h.OptionalAuthMiddleware)

	// No credentials: request continues anonymously.
	status, out := doProbe(t, app, httptest.NewRequest("GET", "/probe", nil))
	assert.Equal(t, http.StatusOK, status)
	assert.False(t, out.Authenticated)

	// Invalid cookie: anonymous, not an error.
	status, out = doProbe(t, app, sessionRequest("GET", "/probe", "garbage"))
	assert.Equal(t, http.StatusOK, status)
	assert.False(t, out.Authenticated)

	// Invalid bearer: anonymous, not an error.
	status, out = doProbe(t, app, bearerRequest("GET", "/probe", "garbage"))
	assert.Equal(t, http.StatusOK, status)
	assert.False(t, out.Authenticated)

	// Malformed Authorization header: anonymous, not an error.
	req := httptest.NewRequest("GET", "/probe", nil)
	req.Header.Set("Authorization", "Basic abc")
	status, out = doProbe(t, app, req)
	assert.Equal(t, http.StatusOK, status)
	assert.False(t, out.Authenticated)

	// Valid session cookie: identified.
	token, err := svc.Ops.Identity.CreateWebSession(ctx, userID, "10.0.0.1", "test-browser")
	require.NoError(t, err)
	status, out = doProbe(t, app, sessionRequest("GET", "/probe", token))
	assert.Equal(t, http.StatusOK, status)
	assert.True(t, out.Authenticated)
	assert.Equal(t, userID, out.UserID)

	// Valid bearer: identified.
	raw, err := svc.Ops.Identity.GenerateAPIToken(ctx, userID, "opt", "read", 30)
	require.NoError(t, err)
	status, out = doProbe(t, app, bearerRequest("GET", "/probe", raw))
	assert.Equal(t, http.StatusOK, status)
	assert.True(t, out.Authenticated)
}

func TestAnonymousAPIMiddleware_Integration(t *testing.T) {
	h, svc, _, userID := testAuthHandler(t)
	ctx := context.Background()
	app := probeApp(h.AnonymousAPIMiddleware)

	// Disabled (default): unauthenticated requests are rejected.
	require.NoError(t, svc.Config.SetCtx(ctx, "anonymous_api_enabled", "false"))
	status, _ := doProbe(t, app, httptest.NewRequest("GET", "/probe", nil))
	assert.Equal(t, http.StatusUnauthorized, status)

	// Enabled: unauthenticated requests pass through anonymously.
	require.NoError(t, svc.Config.SetCtx(ctx, "anonymous_api_enabled", "true"))
	t.Cleanup(func() { _ = svc.Config.SetCtx(ctx, "anonymous_api_enabled", "false") })
	status, out := doProbe(t, app, httptest.NewRequest("GET", "/probe", nil))
	assert.Equal(t, http.StatusOK, status)
	assert.False(t, out.Authenticated)

	// Enabled with a valid session: still identifies the user.
	token, err := svc.Ops.Identity.CreateWebSession(ctx, userID, "10.0.0.1", "test-browser")
	require.NoError(t, err)
	status, out = doProbe(t, app, sessionRequest("GET", "/probe", token))
	assert.Equal(t, http.StatusOK, status)
	assert.True(t, out.Authenticated)
	assert.Equal(t, userID, out.UserID)
}
