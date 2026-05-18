package handlers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func respondCustomerASError(c *fiber.Ctx, err error, resourceName string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, resourceName+" not found")
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return RespondError(c, fiber.StatusConflict, ErrConflict, resourceName+" already exists")
		case "23502", "23514":
			return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid "+resourceName)
		}
	}

	msg := err.Error()
	if strings.Contains(msg, "required") || strings.Contains(msg, "invalid") || strings.Contains(msg, "positive") {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, msg)
	}

	return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
}
