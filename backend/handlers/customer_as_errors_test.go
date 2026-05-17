package handlers

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestRespondCustomerASError_StatusMapping(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"not_found", pgx.ErrNoRows, fiber.StatusNotFound},
		{"duplicate", &pgconn.PgError{Code: "23505"}, fiber.StatusConflict},
		{"check_constraint", &pgconn.PgError{Code: "23514"}, fiber.StatusBadRequest},
		{"validation", errors.New("ASN must be a positive integer"), fiber.StatusBadRequest},
		{"unknown", errors.New("connection failed"), fiber.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c *fiber.Ctx) error {
				return respondCustomerASError(c, tc.err, "customer")
			})

			resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
			assert.NoError(t, err)
			assert.Equal(t, tc.wantStatus, resp.StatusCode)
		})
	}
}
