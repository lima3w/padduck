package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"padduck/internal/testdb"
	"padduck/repository"
	"padduck/services"
	"padduck/utils"
)

// testFullApp builds the real application (RegisterRoutes, all middleware
// including CSRF) on a scratch database with one password-holding user.
func testFullApp(t *testing.T) (*fiber.App, *services.Service, *repository.Repository, int64) {
	t.Helper()
	pool := testdb.Connect(t, "handlers")
	testdb.Truncate(t, pool,
		"sessions", "api_tokens",
		"login_attempts", "account_lockouts", "security_notifications", "users")
	repo := repository.NewRepository(pool)
	svc := services.NewService(repo, testHandlerMFAKey)
	h := NewHandler(svc, svc.Ops, false)
	app := fiber.New()
	h.RegisterRoutes(app)

	hash, err := utils.HashPassword("flow-password-123")
	require.NoError(t, err)
	u, err := repo.CreateUserWithPassword(context.Background(), "flow-user", "flow@example.com", hash, "user")
	require.NoError(t, err)
	return app, svc, repo, u.ID
}

func postJSON(t *testing.T, app *fiber.App, path string, body string, cookies ...*http.Cookie) *http.Response {
	t.Helper()
	req := httptest.NewRequest("POST", path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	resp, err := app.Test(req, 60000)
	require.NoError(t, err)
	return resp
}

func decodeJSON(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()
	var out map[string]any
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &out), "body: %s", body)
	return out
}

func cookieFromResponse(t *testing.T, resp *http.Response, name string) *http.Cookie {
	t.Helper()
	for _, c := range resp.Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func loginAs(t *testing.T, app *fiber.App, username, password string) *http.Cookie {
	t.Helper()
	resp := postJSON(t, app, "/api/v1/auth/login",
		fmt.Sprintf(`{"username":%q,"password":%q}`, username, password))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	cookie := cookieFromResponse(t, resp, "session")
	require.NotNil(t, cookie, "login must set a session cookie")
	resp.Body.Close()
	return cookie
}

// csrfPair fetches a CSRF cookie+header pair from the token endpoint.
func csrfPair(t *testing.T, app *fiber.App) (*http.Cookie, string) {
	t.Helper()
	resp, err := app.Test(httptest.NewRequest("GET", "/api/v1/csrf-token", nil))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body := decodeJSON(t, resp)
	token, _ := body["csrfToken"].(string)
	if token == "" {
		token, _ = body["csrf_token"].(string)
	}
	require.NotEmpty(t, token, "csrf endpoint must return a token: %v", body)
	cookie := cookieFromResponse(t, resp, CSRFCookieName)
	require.NotNil(t, cookie)
	return cookie, token
}

func TestLoginSessionCSRFFlow_E2E(t *testing.T) {
	app, _, _, userID := testFullApp(t)

	session := loginAs(t, app, "flow-user", "flow-password-123")

	// The session cookie authenticates /auth/me.
	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	req.AddCookie(session)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	me := decodeJSON(t, resp)
	assert.EqualValues(t, userID, me["id"])

	// A cookie-authenticated mutation without a CSRF token is rejected.
	resp = postJSON(t, app, "/api/v1/auth/me/logout", "", session)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "logout without CSRF must fail")
	resp.Body.Close()

	// With the CSRF cookie + header pair it succeeds.
	csrfCookie, csrfToken := csrfPair(t, app)
	req = httptest.NewRequest("POST", "/api/v1/auth/me/logout", nil)
	req.AddCookie(session)
	req.AddCookie(csrfCookie)
	req.Header.Set(CSRFHeaderName, csrfToken)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Contains(t, []int{http.StatusOK, http.StatusNoContent}, resp.StatusCode)
	resp.Body.Close()

	// The session is gone after logout.
	req = httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	req.AddCookie(session)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
}

func TestLogin_FailureMapping_E2E(t *testing.T) {
	app, _, repo, userID := testFullApp(t)
	ctx := context.Background()

	readBody := func(resp *http.Response) string {
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		return string(b)
	}

	// Wrong password and unknown user: same status, byte-identical body, so
	// the response can never confirm whether the account exists.
	respWrong := postJSON(t, app, "/api/v1/auth/login", `{"username":"flow-user","password":"wrong"}`)
	assert.Equal(t, http.StatusUnauthorized, respWrong.StatusCode)
	bodyWrong := readBody(respWrong)

	respGhost := postJSON(t, app, "/api/v1/auth/login", `{"username":"no-such-user","password":"wrong"}`)
	assert.Equal(t, http.StatusUnauthorized, respGhost.StatusCode)
	assert.Equal(t, bodyWrong, readBody(respGhost), "unknown-user and wrong-password responses must be identical")

	// Missing fields.
	resp := postJSON(t, app, "/api/v1/auth/login", `{"username":"flow-user"}`)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()

	// Disabled account: 403 only with the correct password.
	require.NoError(t, repo.UpdateUserState(ctx, userID, "disabled"))
	resp = postJSON(t, app, "/api/v1/auth/login", `{"username":"flow-user","password":"flow-password-123"}`)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
	resp = postJSON(t, app, "/api/v1/auth/login", `{"username":"flow-user","password":"wrong"}`)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
	require.NoError(t, repo.UpdateUserState(ctx, userID, "active"))
}

func TestLogin_Lockout_E2E(t *testing.T) {
	app, _, _, _ := testFullApp(t)

	for i := 0; i < 5; i++ {
		resp := postJSON(t, app, "/api/v1/auth/login", `{"username":"flow-user","password":"wrong"}`)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		resp.Body.Close()
	}

	// Locked: even the correct password gets 429 with the lockout message.
	resp := postJSON(t, app, "/api/v1/auth/login", `{"username":"flow-user","password":"flow-password-123"}`)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	body := decodeJSON(t, resp)
	assert.Contains(t, body["error"], "temporarily locked")

	// A wrong password while locked still reads as a generic 401.
	resp = postJSON(t, app, "/api/v1/auth/login", `{"username":"flow-user","password":"wrong"}`)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
}

func TestLogin_MFAFlow_E2E(t *testing.T) {
	app, svc, _, userID := testFullApp(t)
	ctx := context.Background()

	// Enroll the user in MFA directly through the service.
	secret, _, err := svc.MFA.SetupTOTP(ctx, userID, "flow-user", "flow@example.com")
	require.NoError(t, err)
	code, err := totp.GenerateCode(secret, time.Now())
	require.NoError(t, err)
	_, err = svc.MFA.ConfirmTOTP(ctx, userID, code)
	require.NoError(t, err)

	// Password login returns a challenge, no session cookie.
	resp := postJSON(t, app, "/api/v1/auth/login", `{"username":"flow-user","password":"flow-password-123"}`)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Nil(t, cookieFromResponse(t, resp, "session"))
	body := decodeJSON(t, resp)
	assert.Equal(t, true, body["mfa_required"])
	challenge, _ := body["mfa_challenge"].(string)
	require.NotEmpty(t, challenge)

	// A wrong code is rejected.
	resp = postJSON(t, app, "/api/v1/auth/verify-mfa",
		fmt.Sprintf(`{"mfa_challenge":%q,"code":"000000"}`, challenge))
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	// The right code completes login with a session cookie.
	code, err = totp.GenerateCode(secret, time.Now())
	require.NoError(t, err)
	resp = postJSON(t, app, "/api/v1/auth/verify-mfa",
		fmt.Sprintf(`{"mfa_challenge":%q,"code":%q}`, challenge, code))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotNil(t, cookieFromResponse(t, resp, "session"))
	resp.Body.Close()
}

func TestAvatarEndpoints_E2E(t *testing.T) {
	app, _, _, _ := testFullApp(t)

	session := loginAs(t, app, "flow-user", "flow-password-123")
	csrfCookie, csrfToken := csrfPair(t, app)

	authedReq := func(method, path string, body io.Reader) *http.Request {
		req := httptest.NewRequest(method, path, body)
		req.AddCookie(session)
		req.AddCookie(csrfCookie)
		req.Header.Set(CSRFHeaderName, csrfToken)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		return req
	}

	// No custom avatar: redirect to Gravatar.
	req := httptest.NewRequest("GET", "/api/v1/auth/me/avatar", nil)
	req.AddCookie(session)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Location"), "gravatar.com")
	resp.Body.Close()

	// Upload a real PNG.
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 8, 8))))
	dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
	resp, err = app.Test(authedReq("PUT", "/api/v1/auth/me/avatar",
		strings.NewReader(fmt.Sprintf(`{"source":"custom","data":%q}`, dataURL))), 60000)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Non-image payloads are rejected (the #77 validation).
	junk := "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte("not an image"))
	resp, err = app.Test(authedReq("PUT", "/api/v1/auth/me/avatar",
		strings.NewReader(fmt.Sprintf(`{"source":"custom","data":%q}`, junk))), 60000)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()

	// The stored avatar is served with the real content type and nosniff.
	req = httptest.NewRequest("GET", "/api/v1/auth/me/avatar", nil)
	req.AddCookie(session)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "image/png", resp.Header.Get("Content-Type"))
	assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
	resp.Body.Close()
}
