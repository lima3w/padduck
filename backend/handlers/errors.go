package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

// ErrorResponse is the standard error response format
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	ErrBadRequest       ErrorCode = "BAD_REQUEST"
	ErrUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrForbidden        ErrorCode = "FORBIDDEN"
	ErrNotFound         ErrorCode = "NOT_FOUND"
	ErrConflict         ErrorCode = "CONFLICT"
	ErrInternalServer   ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
)

// RespondError sends a standardized error response
func RespondError(c *fiber.Ctx, statusCode int, code ErrorCode, message string, details ...string) error {
	logMessage := message
	if len(details) > 0 {
		logMessage = logMessage + ": " + details[0]
	}

	// Log errors for debugging (except 4xx client errors in some cases)
	if statusCode >= 500 || statusCode == 401 || statusCode == 403 {
		log.Printf("[%d %s] %s", statusCode, code, logMessage)
	}

	detailStr := ""
	if len(details) > 0 {
		detailStr = details[0]
	}

	return c.Status(statusCode).JSON(ErrorResponse{
		Error:   message,
		Code:    string(code),
		Details: detailStr,
	})
}

// StatusBadRequest returns a 400 error
func (h *Handler) StatusBadRequest(c *fiber.Ctx, message string, details ...string) error {
	return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, message, details...)
}

// StatusUnauthorized returns a 401 error
func (h *Handler) StatusUnauthorized(c *fiber.Ctx, message string, details ...string) error {
	return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, message, details...)
}

// StatusForbidden returns a 403 error
func (h *Handler) StatusForbidden(c *fiber.Ctx, message string, details ...string) error {
	return RespondError(c, fiber.StatusForbidden, ErrForbidden, message, details...)
}

// StatusNotFound returns a 404 error
func (h *Handler) StatusNotFound(c *fiber.Ctx, message string, details ...string) error {
	return RespondError(c, fiber.StatusNotFound, ErrNotFound, message, details...)
}

// StatusConflict returns a 409 error
func (h *Handler) StatusConflict(c *fiber.Ctx, message string, details ...string) error {
	return RespondError(c, fiber.StatusConflict, ErrConflict, message, details...)
}

// StatusInternalServerError returns a 500 error
func (h *Handler) StatusInternalServerError(c *fiber.Ctx, message string, details ...string) error {
	return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, message, details...)
}

// StatusServiceUnavailable returns a 503 error
func (h *Handler) StatusServiceUnavailable(c *fiber.Ctx, message string, details ...string) error {
	return RespondError(c, fiber.StatusServiceUnavailable, ErrServiceUnavailable, message, details...)
}
