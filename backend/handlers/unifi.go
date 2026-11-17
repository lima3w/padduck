package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"

	"padduck/services"
)

const redactedSecret = "********"

// UniFiTestRequest carries optional unsaved connection settings. Blank values,
// and the redaction placeholder for secrets, fall back to the saved settings so
// the test-before-save flow works without re-entering stored secrets.
type UniFiTestRequest struct {
	URL      string `json:"url"`
	APIKey   string `json:"api_key"`
	Username string `json:"username"`
	Password string `json:"password"`
	Site     string `json:"site"`
	SkipTLS  bool   `json:"skip_tls"`
}

func (r UniFiTestRequest) merge(saved services.UniFiConfig) services.UniFiConfig {
	pick := func(v, fallback string) string {
		if v == "" || v == redactedSecret {
			return fallback
		}
		return v
	}
	return services.UniFiConfig{
		URL:      pick(r.URL, saved.URL),
		APIKey:   pick(r.APIKey, saved.APIKey),
		Username: pick(r.Username, saved.Username),
		Password: pick(r.Password, saved.Password),
		Site:     pick(r.Site, saved.Site),
		SkipTLS:  r.SkipTLS,
	}
}

// TestUniFiConnection handles POST /api/v1/admin/unifi/test
func (h *Handler) TestUniFiConnection(c *fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return nil
	}
	var body UniFiTestRequest
	_ = c.BodyParser(&body)
	cfg := body.merge(h.service.UniFiConfigFromSettings(c.Context()))
	if err := h.service.TestUniFiConnection(c.Context(), &cfg); err != nil {
		reqLogger(c).Error("UniFi connection test failed", "error", err)
		return RespondError(c, fiber.StatusBadGateway, ErrBadGateway, "UniFi connection failed: "+err.Error())
	}
	return c.JSON(fiber.Map{"status": "ok"})
}

// SyncUniFi handles POST /api/v1/admin/unifi/sync
func (h *Handler) SyncUniFi(c *fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return nil
	}
	res, err := h.service.SyncUniFi(c.Context())
	if err != nil {
		reqLogger(c).Error("UniFi sync failed", "error", err)
		return RespondError(c, fiber.StatusBadGateway, ErrBadGateway, err.Error())
	}
	adminID, adminName := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: adminID, Username: adminName, Action: "unifi_sync", ResourceType: "unifi",
	})
	return c.JSON(res)
}

// GetUniFiStatus handles GET /api/v1/admin/unifi/status
func (h *Handler) GetUniFiStatus(c *fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return nil
	}
	st, err := h.service.UniFiStatus(c.Context())
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(st)
}

// ListUniFiConflicts handles GET /api/v1/admin/unifi/conflicts?status=open|accepted|dismissed
func (h *Handler) ListUniFiConflicts(c *fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return nil
	}
	status := c.Query("status", "open")
	if status == "all" {
		status = ""
	} else if status != "open" && status != "accepted" && status != "dismissed" {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "status must be open, accepted, dismissed or all")
	}
	items, err := h.service.ListUniFiConflicts(c.Context(), status)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(items)
}

// ResolveUniFiConflict handles POST /api/v1/admin/unifi/conflicts/:id/resolve
// with body {"action":"accept"|"dismiss"}.
func (h *Handler) ResolveUniFiConflict(c *fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return nil
	}
	id, err := parseID(c, "id")
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid conflict ID")
	}
	var body struct {
		Action string `json:"action"`
	}
	if err := c.BodyParser(&body); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	item, err := h.service.ResolveUniFiConflict(c.Context(), id, body.Action)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RespondError(c, fiber.StatusNotFound, ErrNotFound, "conflict not found")
		}
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	adminID, adminName := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: adminID, Username: adminName, Action: "unifi_conflict_" + body.Action,
		ResourceType: "ip_address", ResourceID: &item.IPID,
	})
	return c.JSON(item)
}

// ListUniFiLinks handles GET /api/v1/unifi/links: which VLANs and devices were
// synced from UniFi, for display badges. Names and ids only; no credentials.
func (h *Handler) ListUniFiLinks(c *fiber.Ctx) error {
	items, err := h.service.ListUniFiLinks(c.Context())
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "internal server error")
	}
	return c.JSON(items)
}
