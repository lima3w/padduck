package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

// ListPrivacyVersions handles GET /api/v1/admin/privacy/versions
func (h *Handler) ListPrivacyVersions(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	versions, err := h.service.GetRepository().ListPrivacyVersions(c.Context())
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to list privacy versions", err.Error())
	}
	if versions == nil {
		versions = []*models.PrivacyPolicyVersion{}
	}
	return c.JSON(fiber.Map{"versions": versions, "total": len(versions)})
}

// CreatePrivacyVersion handles POST /api/v1/admin/privacy/versions
// Body: {"version":"2.0","effective_date":"2026-06-01","summary":"..."}
func (h *Handler) CreatePrivacyVersion(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	var req struct {
		Version       string  `json:"version"`
		EffectiveDate string  `json:"effective_date"`
		Summary       *string `json:"summary"`
	}
	if err := c.BodyParser(&req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if req.Version == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "version is required")
	}
	if req.EffectiveDate == "" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "effective_date is required")
	}
	v, err := h.service.GetRepository().CreatePrivacyVersion(c.Context(), req.Version, req.EffectiveDate, req.Summary)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to create privacy version", err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"version": v})
}

// GetConsentReport handles GET /api/v1/admin/privacy/consent-report
func (h *Handler) GetConsentReport(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	statuses, err := h.service.GetRepository().ListUserConsentStatus(c.Context())
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to get consent report", err.Error())
	}
	if statuses == nil {
		statuses = []*models.UserConsentStatus{}
	}
	noConsent := 0
	for _, s := range statuses {
		if !s.HasConsent {
			noConsent++
		}
	}
	return c.JSON(fiber.Map{"users": statuses, "total": len(statuses), "no_consent_count": noConsent})
}
