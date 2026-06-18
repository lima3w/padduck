package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

func resourceIDPtr(id int64) *int64 {
	if id == 0 {
		return nil
	}
	return &id
}

// SuspendUser handles POST /api/v1/admin/users/:id/suspend
func (h *Handler) SuspendUser(c *fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return nil
	}
	admin := c.Locals("user").(*models.User)

	userID, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid user ID")
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.BodyParser(&req); err != nil || req.Reason == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "reason is required")
	}

	if err := h.service.SuspendUser(c.Context(), int64(userID), admin.ID, req.Reason); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}

	h.auditLog(c, services.AuditEntry{
		UserID: &admin.ID, Username: admin.Username,
		Action: "user.suspend", ResourceType: "user",
		ResourceID: resourceIDPtr(int64(userID)), ResourceName: fmt.Sprintf("user:%d", userID),
		Status: "success",
	})
	return c.JSON(fiber.Map{"message": "user suspended"})
}

// UnsuspendUser handles POST /api/v1/admin/users/:id/unsuspend
func (h *Handler) UnsuspendUser(c *fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return nil
	}
	admin := c.Locals("user").(*models.User)

	userID, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid user ID")
	}

	if err := h.service.UnsuspendUser(c.Context(), int64(userID)); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}

	h.auditLog(c, services.AuditEntry{
		UserID: &admin.ID, Username: admin.Username,
		Action: "user.unsuspend", ResourceType: "user",
		ResourceID: resourceIDPtr(int64(userID)), ResourceName: fmt.Sprintf("user:%d", userID),
		Status: "success",
	})
	return c.JSON(fiber.Map{"message": "user unsuspended"})
}

// ImpersonateUser handles POST /api/v1/admin/users/:id/impersonate
// Returns a session token that authenticates as the target user
func (h *Handler) ImpersonateUser(c *fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return nil
	}
	admin := c.Locals("user").(*models.User)

	targetID, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid user ID")
	}

	var req struct {
		MFACode string `json:"mfa_code"`
	}
	_ = c.BodyParser(&req)

	if h.auth.MFA.IsMFAEnabled(c.Context(), admin.ID) {
		if req.MFACode == "" {
			return RespondError(c, fiber.StatusForbidden, ErrForbidden, "MFA code required for impersonation")
		}
		if !h.auth.MFA.ValidateTOTPCode(c.Context(), admin.ID, req.MFACode) {
			return RespondError(c, fiber.StatusForbidden, ErrForbidden, "invalid MFA code")
		}
	}

	token, err := h.service.StartImpersonation(c.Context(), int64(targetID), admin.ID, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}

	h.setSessionCookie(c, token)

	h.auditLog(c, services.AuditEntry{
		UserID: &admin.ID, Username: admin.Username,
		Action: "user.impersonate", ResourceType: "user",
		ResourceID: resourceIDPtr(int64(targetID)), ResourceName: fmt.Sprintf("user:%d", targetID),
		Status: "success",
	})

	return c.JSON(fiber.Map{
		"token":   token,
		"message": "impersonation session created (expires in 1 hour)",
	})
}

// BulkSuspendUsers handles POST /api/v1/admin/users/bulk-suspend
func (h *Handler) BulkSuspendUsers(c *fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return nil
	}
	admin := c.Locals("user").(*models.User)

	var req struct {
		UserIDs []int64 `json:"user_ids"`
		Reason  string  `json:"reason"`
	}
	if err := c.BodyParser(&req); err != nil || len(req.UserIDs) == 0 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "user_ids required")
	}
	if req.Reason == "" {
		req.Reason = "bulk suspension"
	}

	count, err := h.service.BulkSuspendUsers(c.Context(), req.UserIDs, admin.ID, req.Reason)
	if err != nil {
		reqLogger(c).Error("bulk suspend error", "error", err)
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(fiber.Map{"suspended": count})
}

// BulkActivateUsers handles POST /api/v1/admin/users/bulk-activate
func (h *Handler) BulkActivateUsers(c *fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return nil
	}

	var req struct {
		UserIDs []int64 `json:"user_ids"`
	}
	if err := c.BodyParser(&req); err != nil || len(req.UserIDs) == 0 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "user_ids required")
	}

	count, err := h.service.BulkActivateUsers(c.Context(), req.UserIDs)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(fiber.Map{"activated": count})
}

// BulkDeleteUsers handles POST /api/v1/admin/users/bulk-delete
func (h *Handler) BulkDeleteUsers(c *fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return nil
	}
	admin := c.Locals("user").(*models.User)

	var req struct {
		UserIDs []int64 `json:"user_ids"`
	}
	if err := c.BodyParser(&req); err != nil || len(req.UserIDs) == 0 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "user_ids required")
	}

	// Filter out the admin's own ID to prevent self-deletion
	filtered := make([]int64, 0, len(req.UserIDs))
	for _, id := range req.UserIDs {
		if id == admin.ID {
			continue // skip self-deletion
		}
		filtered = append(filtered, id)
	}
	req.UserIDs = filtered

	if len(req.UserIDs) == 0 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "no valid user IDs to delete")
	}

	count, err := h.service.BulkDeleteUsers(c.Context(), req.UserIDs)
	if err != nil {
		if err.Error() == "cannot delete all admins" {
			return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
		}
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}

	h.auditLog(c, services.AuditEntry{
		UserID: &admin.ID, Username: admin.Username,
		Action: "user.bulk_delete", ResourceType: "user",
		ResourceName: fmt.Sprintf("%d users", len(req.UserIDs)),
		Status:       "success",
	})

	return c.JSON(fiber.Map{"deleted": count})
}

// BulkImportUsers handles POST /api/v1/admin/users/bulk-import
// Accepts multipart/form-data with a "file" field (CSV) and optional "default_password"
func (h *Handler) BulkImportUsers(c *fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return nil
	}
	admin := c.Locals("user").(*models.User)

	file, err := c.FormFile("file")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "file field required")
	}

	const maxBulkImportSize = 5 * 1024 * 1024 // 5 MB
	if file.Size > maxBulkImportSize {
		return RespondError(c, fiber.StatusRequestEntityTooLarge, ErrPayloadTooLarge, "file too large (max 5 MB)")
	}

	f, err := file.Open()
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to open file")
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.TrimLeadingSpace = true
	rows, err := reader.ReadAll()
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid CSV: " + err.Error())
	}

	if len(rows) < 2 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "CSV must have header row and at least one data row")
	}

	// Parse header to find column indices
	header := rows[0]
	colIdx := map[string]int{"username": -1, "email": -1, "role": -1}
	for i, col := range header {
		key := strings.ToLower(strings.TrimSpace(col))
		if _, ok := colIdx[key]; ok {
			colIdx[key] = i
		}
	}
	if colIdx["username"] < 0 || colIdx["email"] < 0 {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "CSV must have 'username' and 'email' columns")
	}

	records := make([]services.BulkUserImportRecord, 0, len(rows)-1)
	for _, row := range rows[1:] {
		if len(row) == 0 {
			continue
		}
		rec := services.BulkUserImportRecord{}
		if colIdx["username"] < len(row) {
			rec.Username = strings.TrimSpace(row[colIdx["username"]])
		}
		if colIdx["email"] < len(row) {
			rec.Email = strings.TrimSpace(row[colIdx["email"]])
		}
		if colIdx["role"] >= 0 && colIdx["role"] < len(row) {
			rec.Role = strings.TrimSpace(row[colIdx["role"]])
		}
		records = append(records, rec)
	}

	results, err := h.service.BulkImportUsers(c.Context(), records)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "import failed")
	}

	h.auditLog(c, services.AuditEntry{
		UserID: &admin.ID, Username: admin.Username,
		Action: "user.bulk_import", ResourceType: "user",
		ResourceName: fmt.Sprintf("%d users", len(records)),
		Status:       "success",
	})

	return c.JSON(fiber.Map{"results": results})
}

// ExportMyData handles GET /api/v1/auth/me/export (GDPR data export)
func (h *Handler) ExportMyData(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "not authenticated")
	}

	data, err := h.service.ExportUserData(c.Context(), user.ID)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to export data")
	}

	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"user_%d_export_%s.json\"", user.ID, time.Now().Format("20060102")))
	c.Set("Content-Type", "application/json")

	enc, _ := json.MarshalIndent(data, "", "  ")
	return c.Send(enc)
}

// RequestDeletion handles POST /api/v1/auth/me/deletion-request
func (h *Handler) RequestDeletion(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "not authenticated")
	}

	if err := h.service.RequestAccountDeletion(c.Context(), user.ID); err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to submit deletion request")
	}

	h.auditLog(c, services.AuditEntry{
		UserID: &user.ID, Username: user.Username,
		Action: "user.deletion_requested", ResourceType: "user",
		ResourceID: &user.ID, ResourceName: user.Username,
		Status: "success",
	})

	return c.JSON(fiber.Map{"message": "deletion request submitted; an admin will process it shortly"})
}

// GDPRDeleteUser handles POST /api/v1/admin/users/:id/gdpr-delete
func (h *Handler) GDPRDeleteUser(c *fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return nil
	}
	admin := c.Locals("user").(*models.User)

	userID, err := c.ParamsInt("id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid user ID")
	}

	if err := h.service.GDPRDeleteUser(c.Context(), int64(userID)); err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to anonymize user")
	}

	h.auditLog(c, services.AuditEntry{
		UserID: &admin.ID, Username: admin.Username,
		Action: "user.gdpr_delete", ResourceType: "user",
		ResourceID: resourceIDPtr(int64(userID)), ResourceName: fmt.Sprintf("user:%d", userID),
		Status: "success",
	})

	return c.JSON(fiber.Map{"message": "user data anonymized"})
}

// AcceptPrivacyPolicy handles POST /api/v1/auth/me/accept-privacy
func (h *Handler) AcceptPrivacyPolicy(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return RespondError(c, fiber.StatusUnauthorized, ErrUnauthorized, "not authenticated")
	}

	if err := h.service.AcceptPrivacyPolicy(c.Context(), user.ID); err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to record consent")
	}

	return c.JSON(fiber.Map{"message": "privacy policy accepted"})
}

// GetPrivacyPolicyVersion handles GET /api/v1/privacy-policy/version
func (h *Handler) GetPrivacyPolicyVersion(c *fiber.Ctx) error {
	version := h.service.GetPrivacyPolicyVersion(c.Context())
	return c.JSON(fiber.Map{"version": version})
}
