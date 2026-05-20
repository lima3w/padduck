package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
	"padduck/utils"
)

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UpdateUserRoleRequest struct {
	Role string `json:"role"`
}

type UserDetailResponse struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	State       string `json:"state"`
	GravatarURL string `json:"gravatar_url"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ListUsers handles GET /api/v1/users (admin only)
// Supports ?page=1&limit=25 for pagination. Without those params it returns all results.
func (h *Handler) ListUsers(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}
	if user.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "admin access required"})
	}

	page := c.QueryInt("page", 0)
	limit := c.QueryInt("limit", 0)

	if page > 0 || limit > 0 {
		if page < 1 {
			page = 1
		}
		if limit < 1 {
			limit = 25
		}
		users, total, err := h.service.ListUsersPaginated(c.Context(), page, limit)
		if err != nil {
			reqLogger(c).Error("error listing users", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
		}
		response := make([]UserDetailResponse, 0, len(users))
		for _, u := range users {
			response = append(response, UserDetailResponse{
				ID:          u.ID,
				Username:    u.Username,
				Email:       u.Email,
				Role:        u.Role,
				State:       u.State,
				GravatarURL: gravatarURL(u.Email, 80),
				CreatedAt:   u.CreatedAt.String(),
				UpdatedAt:   u.UpdatedAt.String(),
			})
		}
		return c.JSON(fiber.Map{
			"data":  response,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}

	users, err := h.service.ListAllUsers(c.Context())
	if err != nil {
		reqLogger(c).Error("error listing users", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	response := make([]UserDetailResponse, 0)
	for _, u := range users {
		response = append(response, UserDetailResponse{
			ID:          u.ID,
			Username:    u.Username,
			Email:       u.Email,
			Role:        u.Role,
			State:       u.State,
			GravatarURL: gravatarURL(u.Email, 80),
			CreatedAt:   u.CreatedAt.String(),
			UpdatedAt:   u.UpdatedAt.String(),
		})
	}

	return c.JSON(response)
}

// GetUser handles GET /api/v1/users/:id
func (h *Handler) GetUser(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user ID"})
	}

	currentUser, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found in context"})
	}

	if currentUser.ID != int64(userID) && currentUser.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "cannot view other users"})
	}

	user, err := h.service.GetUserByID(c.Context(), int64(userID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	return c.JSON(UserDetailResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		State:       user.State,
		GravatarURL: gravatarURL(user.Email, 80),
		CreatedAt:   user.CreatedAt.String(),
		UpdatedAt:   user.UpdatedAt.String(),
	})
}

// CreateUser handles POST /api/v1/users (admin only)
func (h *Handler) CreateUser(c *fiber.Ctx) error {
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}
	if currentUser.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "admin access required"})
	}

	req := new(CreateUserRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username, email, and password required"})
	}

	role := req.Role
	if role == "" {
		role = "user"
	}
	if role != "admin" && role != "user" && role != "viewer" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid role"})
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		reqLogger(c).Error("error hashing password", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create user"})
	}

	user, err := h.service.CreateUserWithPassword(c.Context(), req.Username, req.Email, hash, role)
	if err != nil {
		reqLogger(c).Error("error creating user", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create user"})
	}

	adminID, adminName := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: adminID, Username: adminName, Action: "user_created",
		ResourceType: "user", ResourceID: &user.ID, ResourceName: user.Username,
		NewValues: map[string]string{"username": user.Username, "email": user.Email, "role": user.Role},
	})

	return c.Status(fiber.StatusCreated).JSON(UserDetailResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		State:       user.State,
		GravatarURL: gravatarURL(user.Email, 80),
		CreatedAt:   user.CreatedAt.String(),
		UpdatedAt:   user.UpdatedAt.String(),
	})
}

// UpdateUserRole handles PUT /api/v1/users/:id/role (admin only)
func (h *Handler) UpdateUserRole(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user ID"})
	}

	currentUser, ok := c.Locals("user").(*models.User)
	if !ok || currentUser.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "admin access required"})
	}

	req := new(UpdateUserRoleRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Role != "admin" && req.Role != "user" && req.Role != "viewer" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid role"})
	}

	user, err := h.service.UpdateUserRole(c.Context(), int64(userID), req.Role)
	if err != nil {
		reqLogger(c).Error("error updating user role", "user_id", userID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update user"})
	}

	adminID, adminName := auditUserFromCtx(c)
	uid := int64(userID)
	h.auditLog(c, services.AuditEntry{
		UserID: adminID, Username: adminName, Action: "user_role_updated",
		ResourceType: "user", ResourceID: &uid, ResourceName: user.Username,
		NewValues: map[string]string{"role": req.Role},
	})

	return c.JSON(UserDetailResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		State:     user.State,
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
	})
}

// DeleteUser handles DELETE /api/v1/users/:id (admin only)
func (h *Handler) DeleteUser(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user ID"})
	}

	currentUser, ok := c.Locals("user").(*models.User)
	if !ok || currentUser.Role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "admin access required"})
	}

	if currentUser.ID == int64(userID) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot delete your own account"})
	}

	if err := h.service.DeleteUser(c.Context(), int64(userID)); err != nil {
		reqLogger(c).Error("error deleting user", "user_id", userID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete user"})
	}

	adminID, adminName := auditUserFromCtx(c)
	uid := int64(userID)
	h.auditLog(c, services.AuditEntry{
		UserID: adminID, Username: adminName, Action: "user_deleted",
		ResourceType: "user", ResourceID: &uid,
	})

	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateUserEmail handles PUT /api/v1/admin/users/:id/email
func (h *Handler) UpdateUserEmail(c *fiber.Ctx) error {
	admin, ok := c.Locals("user").(*models.User)
	if !ok || admin.Role != "admin" {
		return RespondError(c, fiber.StatusForbidden, ErrForbidden, "admin access required")
	}

	userID, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid user ID")
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil || req.Email == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "email is required")
	}

	if err := h.service.UpdateUserEmail(c.Context(), int64(userID), req.Email); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}

	h.auditLog(c, services.AuditEntry{
		UserID: &admin.ID, Username: admin.Username,
		Action: "user.update_email", ResourceType: "user",
		ResourceID: resourceIDPtr(int64(userID)), ResourceName: req.Email,
		Status: "success",
	})

	return c.JSON(fiber.Map{"message": "email updated"})
}

// SendPasswordResetEmail handles POST /api/v1/admin/users/:id/send-password-reset
func (h *Handler) SendPasswordResetEmail(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok || user.Role != "admin" {
		return RespondError(c, fiber.StatusForbidden, ErrForbidden, "admin access required")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid user ID")
	}

	if err := h.service.SendPasswordResetEmailByID(c.Context(), int64(id)); err != nil {
		reqLogger(c).Error("error sending password reset email", "user_id", id, "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to send password reset email")
	}

	return c.JSON(fiber.Map{"message": "Password reset email sent"})
}
