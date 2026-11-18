package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"padduck/repository"
	"padduck/services"
	"padduck/utils"
)

// orgUser creates an admin user that belongs to a fresh organization and
// returns its session token plus the organization ID.
func orgUser(t *testing.T, svc *services.Service, repo *repository.Repository, label string) (token string, orgID, userID int64) {
	t.Helper()
	ctx := context.Background()
	slug := fmt.Sprintf("%s-%d", label, time.Now().UnixNano())
	org, err := repo.CreateOrganization(ctx, "Org "+slug, slug)
	require.NoError(t, err)
	hash, err := utils.HashPassword("iso-password")
	require.NoError(t, err)
	u, err := repo.CreateUserWithPassword(ctx, "iso-"+slug, slug+"@example.com", hash, "admin")
	require.NoError(t, err)
	_, err = repo.GetPool().Exec(ctx, `UPDATE users SET organization_id = $1 WHERE id = $2`, org.ID, u.ID)
	require.NoError(t, err)
	token, err = svc.Ops.Identity.CreateWebSession(ctx, u.ID, "10.0.0.1", "iso-browser")
	require.NoError(t, err)
	return token, org.ID, u.ID
}

// TestHTTPOrgIsolation_Integration drives the real middleware and handlers with
// two users in different organizations.
func TestHTTPOrgIsolation_Integration(t *testing.T) {
	h, svc, repo, _ := testAuthHandler(t)
	tokenA, orgA, userA := orgUser(t, svc, repo, "http-a")
	tokenB, _, _ := orgUser(t, svc, repo, "http-b")
	ctxA := repository.WithOrg(context.Background(), orgA)

	net, err := repo.CreateNetwork(ctxA, "http-net-a", "", userA)
	require.NoError(t, err)
	sub, err := repo.CreateSubnet(ctxA, net.ID, "10.88.0.0", 24, "http-sub-a", nil, false, false)
	require.NoError(t, err)

	app := fiber.New()
	app.Use(h.AuthMiddleware)
	app.Get("/networks", h.ListNetworks)
	app.Get("/networks/:id", h.GetNetwork)
	app.Get("/subnets/:id", h.GetSubnet)
	app.Delete("/subnets/:id", h.DeleteSubnet)

	do := func(method, path, token string) (int, string) {
		resp, err := app.Test(sessionRequest(method, path, token))
		require.NoError(t, err)
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(body)
	}
	_ = httptest.NewRequest // keep the import local to this file's helpers

	subPath := fmt.Sprintf("/subnets/%d", sub.ID)
	netPath := fmt.Sprintf("/networks/%d", net.ID)

	// The owner reads its data.
	status, body := do("GET", subPath, tokenA)
	assert.Equal(t, http.StatusOK, status, body)
	status, body = do("GET", netPath, tokenA)
	assert.Equal(t, http.StatusOK, status, body)

	// Another organization's admin does not.
	status, body = do("GET", subPath, tokenB)
	assert.NotEqual(t, http.StatusOK, status, "cross-org subnet read leaked: %s", body)
	assert.NotContains(t, body, "http-sub-a")
	status, body = do("GET", netPath, tokenB)
	assert.NotEqual(t, http.StatusOK, status, "cross-org network read leaked: %s", body)
	assert.NotContains(t, body, "http-net-a")

	// Lists only contain the caller's own rows.
	_, body = do("GET", "/networks", tokenB)
	assert.NotContains(t, body, "http-net-a")
	_, body = do("GET", "/networks", tokenA)
	assert.Contains(t, body, "http-net-a")
	var decoded any
	assert.NoError(t, json.Unmarshal([]byte(body), &decoded))

	// A cross-org delete must not remove the row.
	_, _ = do("DELETE", subPath, tokenB)
	status, body = do("GET", subPath, tokenA)
	assert.Equal(t, http.StatusOK, status, "subnet was deleted across orgs: %s", body)
}
