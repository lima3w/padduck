package handlers

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
	"ipam-next/models"
	"ipam-next/services"
)

type SessionResponse struct {
	ID                int64  `json:"id"`
	DeviceName        string `json:"device_name"`
	IPAddress         string `json:"ip_address"`
	LastUsedAt        string `json:"last_used_at"`
	AbsoluteExpiresAt string `json:"absolute_expires_at"`
	CreatedAt         string `json:"created_at"`
}

func isAccountLocked(err error) bool {
	return errors.Is(err, services.ErrAccountLocked)
}

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
	Role      string `json:"role"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// GenerateToken handles POST /api/v1/auth/tokens
func (h *Handler) GenerateToken(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("userID")
	if err != nil {
		return h.StatusBadRequest(c, "Invalid user ID")
	}

	req := new(GenerateTokenRequest)
	if err := c.BodyParser(req); err != nil {
		return h.StatusBadRequest(c, "Invalid request body")
	}

	token, err := h.service.GenerateAPIToken(c.Context(), int64(userID), req.TokenName)
	if err != nil {
		return h.StatusInternalServerError(c, "Failed to generate token", err.Error())
	}

	uid := int64(userID)
	h.auditLog(c, services.AuditEntry{
		UserID: &uid, Action: "token_created",
		ResourceType: "api_token", ResourceName: req.TokenName,
	})

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

	uid, uname := auditUserFromCtx(c)
	tid := int64(tokenID)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "token_revoked",
		ResourceType: "api_token", ResourceID: &tid,
	})

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
		Role:      user.Role,
		State:     user.State,
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

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "token_created",
		ResourceType: "api_token", ResourceName: req.TokenName,
	})

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

// Login handles POST /api/v1/auth/login
func (h *Handler) Login(c *fiber.Ctx) error {
	req := new(LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username and password required"})
	}

	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")

	result, err := h.service.AuthenticateUser(c.Context(), req.Username, req.Password, ipAddress, userAgent)
	if err != nil {
		log.Printf("Authentication error for user %s: %v", req.Username, err)
		switch {
		case err == services.ErrEmailNotVerified, err == services.ErrPendingApproval,
			err == services.ErrAccountRejected, err == services.ErrAccountDisabled:
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		case isAccountLocked(err):
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "account temporarily locked due to too many failed login attempts"})
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid username or password"})
	}

	// MFA required — return challenge token instead of session
	if result.MFARequired {
		return c.JSON(fiber.Map{
			"mfa_required":  true,
			"mfa_challenge": result.MFAChallenge,
		})
	}

	user := result.User

	// Update last login time
	if err := h.service.UpdateLastLogin(c.Context(), user.ID); err != nil {
		log.Printf("Error updating last login: %v", err)
	}

	token, err := h.service.CreateWebSession(c.Context(), user.ID, c.IP(), c.Get("User-Agent"))
	if err != nil {
		log.Printf("Error creating session for user %d: %v", user.ID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create session"})
	}

	uid := user.ID
	h.auditLog(c, services.AuditEntry{
		UserID: &uid, Username: user.Username, Action: "login",
		ResourceType: "session", Status: "success",
	})

	return c.JSON(LoginResponse{
		Token: token,
		User: UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			State:     user.State,
			CreatedAt: user.CreatedAt.String(),
			UpdatedAt: user.UpdatedAt.String(),
		},
	})
}

// Logout handles POST /api/v1/auth/logout
func (h *Handler) Logout(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user ID not found in context"})
	}

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing authorization header"})
	}

	var token string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	if err := h.service.RevokeSession(c.Context(), userID, token); err != nil {
		log.Printf("Error revoking session for user %d: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to logout"})
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "logout", ResourceType: "session",
	})

	return c.SendStatus(fiber.StatusNoContent)
}

// ListMySessions handles GET /api/v1/auth/me/sessions
func (h *Handler) ListMySessions(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user ID not found in context"})
	}

	sessions, err := h.service.ListUserSessions(c.Context(), userID)
	if err != nil {
		log.Printf("Error listing sessions for user %d: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	response := make([]SessionResponse, 0, len(sessions))
	for _, s := range sessions {
		response = append(response, SessionResponse{
			ID:                s.ID,
			DeviceName:        s.DeviceName,
			IPAddress:         s.IPAddress,
			LastUsedAt:        s.LastUsedAt.String(),
			AbsoluteExpiresAt: s.AbsoluteExpiresAt.String(),
			CreatedAt:         s.CreatedAt.String(),
		})
	}

	return c.JSON(response)
}

// RevokeMySession handles DELETE /api/v1/auth/me/sessions/:sessionID
func (h *Handler) RevokeMySession(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user ID not found in context"})
	}

	sessionID, err := c.ParamsInt("sessionID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid session ID"})
	}

	if err := h.service.RevokeSessionByID(c.Context(), userID, int64(sessionID)); err != nil {
		log.Printf("Error revoking session %d for user %d: %v", sessionID, userID, err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "session not found"})
	}

	uid, uname := auditUserFromCtx(c)
	sid := int64(sessionID)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "session_revoked",
		ResourceType: "session", ResourceID: &sid,
	})

	return c.SendStatus(fiber.StatusNoContent)
}

// LogoutAllDevices handles DELETE /api/v1/auth/me/sessions
func (h *Handler) LogoutAllDevices(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user ID not found in context"})
	}

	if err := h.service.RevokeAllSessions(c.Context(), userID); err != nil {
		log.Printf("Error revoking all sessions for user %d: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	uid, uname := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "logout_all_devices", ResourceType: "session",
	})

	return c.SendStatus(fiber.StatusNoContent)
}
