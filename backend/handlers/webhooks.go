package handlers

import (
	"context"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"padduck/models"
	"padduck/services"
)

var webhookPrivateRanges []*net.IPNet

func init() {
	for _, cidr := range []string{
		"127.0.0.0/8",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16",
		"::1/128",
		"fc00::/7",
		"fe80::/10",
	} {
		_, network, _ := net.ParseCIDR(cidr)
		webhookPrivateRanges = append(webhookPrivateRanges, network)
	}
}

func isPrivateIP(ip net.IP) bool {
	for _, network := range webhookPrivateRanges {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

type webhookEndpointRequest struct {
	Name             string            `json:"name"`
	URL              string            `json:"url"`
	Secret           string            `json:"secret"`
	Events           []string          `json:"events"`
	ObjectTypes      []string          `json:"object_types"`
	TagFilters       []string          `json:"tag_filters"`
	FilterConditions map[string]string `json:"filter_conditions"`
	IsActive         *bool             `json:"is_active"`
}

type webhookEndpointResponse struct {
	ID               int64             `json:"id"`
	Name             string            `json:"name"`
	URL              string            `json:"url"`
	Events           []string          `json:"events"`
	ObjectTypes      []string          `json:"object_types"`
	TagFilters       []string          `json:"tag_filters"`
	FilterConditions map[string]string `json:"filter_conditions,omitempty"`
	IsActive         bool              `json:"is_active"`
	CreatedBy        *int64            `json:"created_by,omitempty"`
	CreatedAt        string            `json:"created_at"`
	UpdatedAt        string            `json:"updated_at"`
}

func (h *Handler) GetWebhookSamplePayload(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	eventType := c.Query("event_type", "ip_address.assigned")
	return c.JSON(fiber.Map{
		"schema_version": services.WebhookEventSchemaVersion,
		"payload":        services.SampleWebhookEventPayload(eventType),
		"headers": fiber.Map{
			"Content-Type":                "application/json",
			"User-Agent":                  "padduck-webhooks/1.0",
			"X-IPAM-Event":                eventType,
			"X-IPAM-Event-Schema-Version": services.WebhookEventSchemaVersion,
			"X-IPAM-Signature-256":        "sha256=<hex-hmac-sha256>",
		},
	})
}

func (h *Handler) ListWebhookEndpoints(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	endpoints, err := h.ops.Webhooks.ListEndpoints(c.Context(), orgIDFromCtx(c))
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to load webhook endpoints")
	}
	out := make([]webhookEndpointResponse, 0, len(endpoints))
	for _, endpoint := range endpoints {
		out = append(out, formatWebhookEndpoint(endpoint))
	}
	return c.JSON(out)
}

func (h *Handler) CreateWebhookEndpoint(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	req := new(webhookEndpointRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if fields := validateWebhookEndpointRequest(req); len(fields) > 0 {
		return RespondValidationError(c, "validation failed", fields)
	}
	if err := h.ops.OrgSettings.CheckQuota(c.Context(), orgIDFromCtx(c), "webhooks"); err != nil {
		return RespondError(c, fiber.StatusUnprocessableEntity, ErrQuotaExceeded, err.Error())
	}
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	createdBy, username := auditUserFromCtx(c)
	endpoint, err := h.ops.Webhooks.CreateEndpoint(c.Context(), &models.WebhookEndpoint{
		OrganizationID:   orgIDFromCtx(c),
		Name:             req.Name,
		URL:              req.URL,
		Secret:           req.Secret,
		Events:           req.Events,
		ObjectTypes:      req.ObjectTypes,
		TagFilters:       req.TagFilters,
		FilterConditions: req.FilterConditions,
		IsActive:         active,
		CreatedBy:        createdBy,
	})
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	h.auditLog(c, services.AuditEntry{
		UserID: createdBy, Username: username, Action: "created",
		ResourceType: "webhook_endpoint", ResourceID: &endpoint.ID, ResourceName: endpoint.Name,
		NewValues: map[string]interface{}{"name": endpoint.Name, "url": endpoint.URL, "events": endpoint.Events, "object_types": endpoint.ObjectTypes, "tag_filters": endpoint.TagFilters, "is_active": endpoint.IsActive},
	})
	return c.Status(fiber.StatusCreated).JSON(formatWebhookEndpoint(endpoint))
}

func (h *Handler) UpdateWebhookEndpoint(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid webhook endpoint ID")
	}
	req := new(webhookEndpointRequest)
	if err := c.BodyParser(req); err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid request body")
	}
	if fields := validateWebhookEndpointRequest(req); len(fields) > 0 {
		return RespondValidationError(c, "validation failed", fields)
	}
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	endpoint, err := h.ops.Webhooks.UpdateEndpoint(c.Context(), &models.WebhookEndpoint{
		ID:               id,
		Name:             req.Name,
		URL:              req.URL,
		Secret:           req.Secret,
		Events:           req.Events,
		ObjectTypes:      req.ObjectTypes,
		TagFilters:       req.TagFilters,
		FilterConditions: req.FilterConditions,
		IsActive:         active,
	})
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, err.Error())
	}
	uid, username := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: username, Action: "updated",
		ResourceType: "webhook_endpoint", ResourceID: &endpoint.ID, ResourceName: endpoint.Name,
		NewValues: map[string]interface{}{"name": endpoint.Name, "url": endpoint.URL, "events": endpoint.Events, "object_types": endpoint.ObjectTypes, "tag_filters": endpoint.TagFilters, "is_active": endpoint.IsActive},
	})
	return c.JSON(formatWebhookEndpoint(endpoint))
}

func validateWebhookEndpointRequest(req *webhookEndpointRequest) []ValidationField {
	fields := make([]ValidationField, 0)
	name := strings.TrimSpace(req.Name)
	if name == "" {
		fields = append(fields, ValidationField{Field: "name", Message: "name is required"})
	} else if len(name) > 255 {
		fields = append(fields, ValidationField{Field: "name", Message: "name must be 255 characters or fewer"})
	}
	rawURL := strings.TrimSpace(req.URL)
	if rawURL == "" {
		fields = append(fields, ValidationField{Field: "url", Message: "url is required"})
		return fields
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		fields = append(fields, ValidationField{Field: "url", Message: "url must use http or https"})
		return fields
	}
	addrs, err := net.LookupHost(parsed.Hostname())
	if err != nil {
		fields = append(fields, ValidationField{Field: "url", Message: "url hostname could not be resolved"})
		return fields
	}
	for _, addr := range addrs {
		if ip := net.ParseIP(addr); ip != nil && isPrivateIP(ip) {
			fields = append(fields, ValidationField{Field: "url", Message: "url must not resolve to a private or loopback address"})
			return fields
		}
	}
	return fields
}

func (h *Handler) DeleteWebhookEndpoint(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid webhook endpoint ID")
	}
	if err := h.ops.Webhooks.DeleteEndpoint(c.Context(), id); err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "webhook endpoint not found")
	}
	uid, username := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: username, Action: "deleted",
		ResourceType: "webhook_endpoint", ResourceID: &id,
	})
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListWebhookDeliveries(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	limit := c.QueryInt("limit", 50)
	deliveries, err := h.ops.Webhooks.ListDeliveries(c.Context(), limit)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to load webhook deliveries")
	}
	return c.JSON(deliveries)
}

func (h *Handler) ListWebhookFailureGroups(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	limit := c.QueryInt("limit", 50)
	groups, err := h.ops.Webhooks.ListFailureGroups(c.Context(), limit)
	if err != nil {
		return RespondError(c, fiber.StatusInternalServerError, ErrInternalServer, "failed to load webhook failure groups")
	}
	return c.JSON(groups)
}

func (h *Handler) GetWebhookDelivery(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid webhook delivery ID")
	}
	delivery, err := h.ops.Webhooks.GetDelivery(c.Context(), id)
	if err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "webhook delivery not found")
	}
	return c.JSON(delivery)
}

func (h *Handler) ReplayWebhookDelivery(c *fiber.Ctx) error {
	if !h.requirePerm(c, services.PermV2AdminWrite) {
		return nil
	}
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, ErrBadRequest, "invalid webhook delivery ID")
	}
	if c.QueryBool("async") {
		job := h.ops.Jobs.Enqueue("webhook_replay", "Replay webhook delivery", fiber.Map{"delivery_id": id}, 2, func(ctx context.Context, reporter *services.JobReporter) (interface{}, error) {
			reporter.Progress(0, 1, "replaying webhook delivery")
			delivery, err := h.ops.Webhooks.ReplayDelivery(ctx, id)
			reporter.Progress(1, 1, "webhook replay complete")
			return delivery, err
		})
		return c.Status(fiber.StatusAccepted).JSON(job)
	}
	delivery, err := h.ops.Webhooks.ReplayDelivery(c.Context(), id)
	if err != nil {
		return RespondError(c, fiber.StatusNotFound, ErrNotFound, "webhook delivery not found")
	}
	uid, username := auditUserFromCtx(c)
	h.auditLog(c, services.AuditEntry{
		UserID: uid, Username: username, Action: "webhook_replayed",
		ResourceType: "webhook_delivery", ResourceID: &delivery.ID, ResourceName: delivery.EventType,
	})
	return c.JSON(delivery)
}

func formatWebhookEndpoint(endpoint *models.WebhookEndpoint) webhookEndpointResponse {
	return webhookEndpointResponse{
		ID:               endpoint.ID,
		Name:             endpoint.Name,
		URL:              endpoint.URL,
		Events:           endpoint.Events,
		ObjectTypes:      endpoint.ObjectTypes,
		TagFilters:       endpoint.TagFilters,
		FilterConditions: endpoint.FilterConditions,
		IsActive:         endpoint.IsActive,
		CreatedBy:        endpoint.CreatedBy,
		CreatedAt:        endpoint.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        endpoint.UpdatedAt.Format(time.RFC3339),
	}
}
