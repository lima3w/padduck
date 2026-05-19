package handlers

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// TestErrorCodeConstants verifies the string values of all error code constants.
func TestErrorCodeConstants(t *testing.T) {
	assert.Equal(t, ErrorCode("BAD_REQUEST"), ErrBadRequest)
	assert.Equal(t, ErrorCode("UNAUTHORIZED"), ErrUnauthorized)
	assert.Equal(t, ErrorCode("FORBIDDEN"), ErrForbidden)
	assert.Equal(t, ErrorCode("NOT_FOUND"), ErrNotFound)
	assert.Equal(t, ErrorCode("CONFLICT"), ErrConflict)
	assert.Equal(t, ErrorCode("VALIDATION_ERROR"), ErrValidation)
	assert.Equal(t, ErrorCode("INTERNAL_SERVER_ERROR"), ErrInternalServer)
	assert.Equal(t, ErrorCode("SERVICE_UNAVAILABLE"), ErrServiceUnavailable)
}

// TestErrorCodeStringValues verifies the raw string values.
func TestErrorCodeStringValues(t *testing.T) {
	assert.Equal(t, "BAD_REQUEST", string(ErrBadRequest))
	assert.Equal(t, "UNAUTHORIZED", string(ErrUnauthorized))
	assert.Equal(t, "FORBIDDEN", string(ErrForbidden))
	assert.Equal(t, "NOT_FOUND", string(ErrNotFound))
	assert.Equal(t, "CONFLICT", string(ErrConflict))
	assert.Equal(t, "VALIDATION_ERROR", string(ErrValidation))
	assert.Equal(t, "INTERNAL_SERVER_ERROR", string(ErrInternalServer))
	assert.Equal(t, "SERVICE_UNAVAILABLE", string(ErrServiceUnavailable))
}

func TestRespondValidationError(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return RespondValidationError(c, "validation failed", []ValidationField{
			{Field: "name", Message: "name is required"},
		})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

	body := parseErrorResponse(t, resp.Body)
	assert.Equal(t, "validation failed", body["error"])
	assert.Equal(t, "VALIDATION_ERROR", body["code"])
	fields := body["fields"].([]interface{})
	assert.Len(t, fields, 1)
}

// parseErrorResponse is a helper that reads a response body and unmarshals it.
func parseErrorResponse(t *testing.T, body io.Reader) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	data, err := io.ReadAll(body)
	assert.NoError(t, err)
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)
	return result
}

// TestRespondError_400 verifies a 400 Bad Request response.
func TestRespondError_400(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid input")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	body := parseErrorResponse(t, resp.Body)
	assert.Equal(t, "invalid input", body["error"])
	assert.Equal(t, "BAD_REQUEST", body["code"])
}

// TestRespondError_500 verifies a 500 Internal Server Error response.
func TestRespondError_500(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "something went wrong")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	body := parseErrorResponse(t, resp.Body)
	assert.Equal(t, "something went wrong", body["error"])
	assert.Equal(t, "INTERNAL_SERVER_ERROR", body["code"])
}

// TestRespondError_WithDetails verifies that details are included in the response.
func TestRespondError_WithDetails(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "validation failed", "field 'name' is required")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	body := parseErrorResponse(t, resp.Body)
	assert.Equal(t, "validation failed", body["error"])
	assert.Equal(t, "BAD_REQUEST", body["code"])
	assert.Equal(t, "field 'name' is required", body["details"])
}

// TestRespondError_WithoutDetails verifies that details are absent when not provided.
func TestRespondError_WithoutDetails(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "resource not found")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	body := parseErrorResponse(t, resp.Body)
	assert.Equal(t, "resource not found", body["error"])
	assert.Equal(t, "NOT_FOUND", body["code"])
	// details should be absent (omitempty) or empty string
	details, exists := body["details"]
	if exists {
		assert.Equal(t, "", details)
	}
}

// TestRespondError_HandlerHelpers verifies that handler method wrappers call RespondError.
func TestRespondError_HandlerHelpers(t *testing.T) {
	h := NewHandler(nil, false)

	tests := []struct {
		name       string
		fn         func(c *fiber.Ctx) error
		statusCode int
		code       string
	}{
		{"BadRequest", func(c *fiber.Ctx) error { return h.StatusBadRequest(c, "bad") }, 400, "BAD_REQUEST"},
		{"Unauthorized", func(c *fiber.Ctx) error { return h.StatusUnauthorized(c, "unauth") }, 401, "UNAUTHORIZED"},
		{"Forbidden", func(c *fiber.Ctx) error { return h.StatusForbidden(c, "forbidden") }, 403, "FORBIDDEN"},
		{"NotFound", func(c *fiber.Ctx) error { return h.StatusNotFound(c, "not found") }, 404, "NOT_FOUND"},
		{"Conflict", func(c *fiber.Ctx) error { return h.StatusConflict(c, "conflict") }, 409, "CONFLICT"},
		{"InternalServerError", func(c *fiber.Ctx) error { return h.StatusInternalServerError(c, "server err") }, 500, "INTERNAL_SERVER_ERROR"},
		{"ServiceUnavailable", func(c *fiber.Ctx) error { return h.StatusServiceUnavailable(c, "unavailable") }, 503, "SERVICE_UNAVAILABLE"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/test", tt.fn)
			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.statusCode, resp.StatusCode)
			body := parseErrorResponse(t, resp.Body)
			assert.Equal(t, tt.code, body["code"])
		})
	}
}

// TestRespondError_AllStatusCodes exercises each helper via RespondError directly.
func TestRespondError_AllCodes(t *testing.T) {
	tests := []struct {
		statusCode int
		code       ErrorCode
		codeStr    string
	}{
		{401, ErrUnauthorized, "UNAUTHORIZED"},
		{403, ErrForbidden, "FORBIDDEN"},
		{409, ErrConflict, "CONFLICT"},
		{503, ErrServiceUnavailable, "SERVICE_UNAVAILABLE"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.codeStr, func(t *testing.T) {
			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				return RespondError(c, tt.statusCode, tt.code, "test message")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.statusCode, resp.StatusCode)

			body := parseErrorResponse(t, resp.Body)
			assert.Equal(t, tt.codeStr, body["code"])
		})
	}
}
