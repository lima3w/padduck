package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

// ListDriftItems handles GET /api/v1/admin/drift
// Query params: ?status=open (default open; pass empty string for all)
func (h *Handler) ListDriftItems(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	status := c.Query("status", "open")
	items, err := h.ops.Discovery.ListDriftItems(c.Context(), orgIDFromCtx(c), status)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	if items == nil {
		items = []*models.DriftItem{}
	}
	return c.JSON(items)
}

// GetDriftItem handles GET /api/v1/admin/drift/:id
func (h *Handler) GetDriftItem(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid drift item ID")
	}
	item, err := h.ops.Discovery.GetDriftItem(c.Context(), id)
	if err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "drift item not found")
	}
	return c.JSON(item)
}

// AcceptDrift handles POST /api/v1/admin/drift/:id/accept
// Applies all field diffs to the authoritative record and marks the item accepted.
func (h *Handler) AcceptDrift(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid drift item ID")
	}
	item, err := h.ops.Discovery.GetDriftItem(c.Context(), id)
	if err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "drift item not found")
	}
	if item.Status != "open" {
		return RespondError(c, fiber.StatusConflict, ErrConflict, "drift item is already resolved")
	}

	repo := h.service.GetRepository()

	if item.ResourceType == "ip_address" {
		ip, err := repo.GetIPAddressByID(c.Context(), item.ResourceID)
		if err != nil {
			return RespondError(c, fiber.StatusNotFound, ErrNotFound, "authoritative IP record not found")
		}
		// Apply each field diff to the authoritative record.
		for _, diff := range item.FieldDiffs {
			switch diff.Field {
			case "hostname":
				ip.Hostname = diff.Observed
			case "mac_address":
				mac := diff.Observed
				ip.MACAddress = &mac
			}
		}
		if _, err := repo.UpdateIPAddressFull(c.Context(), ip.ID, ip.Hostname, ip.TagID, ip.MACAddress, ip.PTRRecord, ip.DNSName); err != nil {
			reqLogger(c).Error("accept drift: update IP failed", "ip_id", ip.ID, "error", err)
			return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to update authoritative record")
		}
	}

	uid, uname := auditUserFromCtx(c)
	resolvedBy := uid
	if err := h.ops.Discovery.ResolveDriftItem(c.Context(), id, "accepted", resolvedBy); err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to resolve drift item")
	}
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "drift_accepted",
		ResourceType: item.ResourceType, ResourceID: &item.ResourceID,
	})
	return c.JSON(fiber.Map{"status": "accepted"})
}

// DismissDrift handles POST /api/v1/admin/drift/:id/dismiss
// Marks the drift item dismissed without modifying the authoritative record.
func (h *Handler) DismissDrift(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid drift item ID")
	}
	item, err := h.ops.Discovery.GetDriftItem(c.Context(), id)
	if err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "drift item not found")
	}
	if item.Status != "open" {
		return RespondError(c, fiber.StatusConflict, ErrConflict, "drift item is already resolved")
	}
	uid, uname := auditUserFromCtx(c)
	if err := h.ops.Discovery.ResolveDriftItem(c.Context(), id, "dismissed", uid); err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to dismiss drift item")
	}
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "drift_dismissed",
		ResourceType: item.ResourceType, ResourceID: &item.ResourceID,
	})
	return c.JSON(fiber.Map{"status": "dismissed"})
}

// EscalateDrift handles POST /api/v1/admin/drift/:id/escalate
// Creates a discovery_conflicts record for investigation and marks the item escalated.
func (h *Handler) EscalateDrift(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid drift item ID")
	}

	var body struct {
		Note string `json:"note"`
	}
	_ = c.BodyParser(&body)

	item, err := h.ops.Discovery.GetDriftItem(c.Context(), id)
	if err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "drift item not found")
	}
	if item.Status != "open" {
		return RespondError(c, fiber.StatusConflict, ErrConflict, "drift item is already resolved")
	}

	// Create a discovery conflict entry for each diverged field.
	for _, diff := range item.FieldDiffs {
		authVal := diff.Authoritative
		note := body.Note
		var notePtr *string
		if note != "" {
			notePtr = &note
		} else {
			notePtr = &authVal
		}
		_, _ = h.ops.Discovery.CreateDiscoveryConflict(c.Context(),
			item.ResourceID, diff.Field, diff.Observed, notePtr, 1.0, "drift")
	}

	uid, uname := auditUserFromCtx(c)
	if err := h.ops.Discovery.ResolveDriftItem(c.Context(), id, "escalated", uid); err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to escalate drift item")
	}
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: uname, Action: "drift_escalated",
		ResourceType: item.ResourceType, ResourceID: &item.ResourceID,
	})
	return c.JSON(fiber.Map{"status": "escalated"})
}
