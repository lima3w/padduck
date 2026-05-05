package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
)

type GenerateTokenRequest struct {
	TokenName string `json:"token_name"`
}

type GenerateTokenResponse struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}

type ListTokensResponse struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	CreatedAt  string `json:"created_at"`
	LastUsedAt string `json:"last_used_at"`
}

type UserResponse struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// GenerateToken handles POST /api/v1/auth/tokens
func (h *Handler) GenerateToken(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("userID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user ID"})
	}

	req := new(GenerateTokenRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	token, err := h.service.GenerateAPIToken(c.Context(), int64(userID), req.TokenName)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(GenerateTokenResponse{
		Token: token,
		Name:  req.TokenName,
	})
}

// ListTokens handles GET /api/v1/auth/tokens/:userID
func (h *Handler) ListTokens(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("userID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user ID"})
	}

	tokens, err := h.service.ListUserTokens(c.Context(), int64(userID))
	if err != nil {
		log.Printf("Error listing tokens: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	// Convert to response format (omit token hash)
	response := make([]ListTokensResponse, 0)
	if tokens != nil {
		for _, token := range tokens {
			lastUsed := ""
			if token.LastUsedAt != nil {
				lastUsed = token.LastUsedAt.String()
			}
			response = append(response, ListTokensResponse{
				ID:         token.ID,
				Name:       token.Name,
				CreatedAt:  token.CreatedAt.String(),
				LastUsedAt: lastUsed,
			})
		}
	}

	return c.JSON(response)
}

// RevokeToken handles DELETE /api/v1/auth/tokens/:tokenID
func (h *Handler) RevokeToken(c *fiber.Ctx) error {
	tokenID, err := c.ParamsInt("tokenID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid token ID"})
	}

	if err := h.service.RevokeAPIToken(c.Context(), int64(tokenID)); err != nil {
		log.Printf("Error revoking token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetCurrentUser handles GET /api/v1/auth/me
func (h *Handler) GetCurrentUser(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found in context"})
	}

	return c.JSON(UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
	})
}

// GenerateTokenForMe handles POST /api/v1/auth/me/tokens
func (h *Handler) GenerateTokenForMe(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user ID not found in context"})
	}

	req := new(GenerateTokenRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	token, err := h.service.GenerateAPIToken(c.Context(), userID, req.TokenName)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(GenerateTokenResponse{
		Token: token,
		Name:  req.TokenName,
	})
}

// ListMyTokens handles GET /api/v1/auth/me/tokens
func (h *Handler) ListMyTokens(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user ID not found in context"})
	}

	tokens, err := h.service.ListUserTokens(c.Context(), userID)
	if err != nil {
		log.Printf("Error listing tokens: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	response := make([]ListTokensResponse, 0)
	if tokens != nil {
		for _, token := range tokens {
			lastUsed := ""
			if token.LastUsedAt != nil {
				lastUsed = token.LastUsedAt.String()
			}
			response = append(response, ListTokensResponse{
				ID:         token.ID,
				Name:       token.Name,
				CreatedAt:  token.CreatedAt.String(),
				LastUsedAt: lastUsed,
			})
		}
	}

	return c.JSON(response)
}
