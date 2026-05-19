package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ipam-next/models"
	"ipam-next/repository"
)

const (
	maxWebhookRetries         = 5
	WebhookEventSchemaVersion = "2026-05-19"
	webhookUserAgent          = "ipam-next-webhooks/1.0"
)

// WebhookService manages outbound webhook endpoints and deliveries.
type WebhookService struct {
	repo   *repository.Repository
	client *http.Client
}

func NewWebhookService(repo *repository.Repository) *WebhookService {
	return &WebhookService{
		repo: repo,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type WebhookEvent struct {
	SchemaVersion string      `json:"schema_version"`
	EventType     string      `json:"event_type"`
	Action        string      `json:"action"`
	ResourceType  string      `json:"resource_type"`
	ResourceID    *int64      `json:"resource_id,omitempty"`
	ResourceName  string      `json:"resource_name,omitempty"`
	UserID        *int64      `json:"user_id,omitempty"`
	Username      string      `json:"username,omitempty"`
	Status        string      `json:"status"`
	OldValues     interface{} `json:"old_values,omitempty"`
	NewValues     interface{} `json:"new_values,omitempty"`
	OccurredAt    time.Time   `json:"occurred_at"`
}

func (w *WebhookService) ListEndpoints(ctx context.Context) ([]*models.WebhookEndpoint, error) {
	return w.repo.ListWebhookEndpoints(ctx)
}

func (w *WebhookService) ListDeliveries(ctx context.Context, limit int) ([]*models.WebhookDelivery, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return w.repo.ListWebhookDeliveries(ctx, limit)
}

func (w *WebhookService) ListFailureGroups(ctx context.Context, limit int) ([]*models.WebhookFailureGroup, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return w.repo.ListWebhookFailureGroups(ctx, limit)
}

func (w *WebhookService) GetDelivery(ctx context.Context, id int64) (*models.WebhookDelivery, error) {
	return w.repo.GetWebhookDelivery(ctx, id)
}

func (w *WebhookService) ReplayDelivery(ctx context.Context, id int64) (*models.WebhookDelivery, error) {
	return w.repo.ReplayWebhookDelivery(ctx, id)
}

func (w *WebhookService) CreateEndpoint(ctx context.Context, endpoint *models.WebhookEndpoint) (*models.WebhookEndpoint, error) {
	if err := validateWebhookEndpoint(endpoint); err != nil {
		return nil, err
	}
	return w.repo.CreateWebhookEndpoint(ctx, endpoint)
}

func (w *WebhookService) UpdateEndpoint(ctx context.Context, endpoint *models.WebhookEndpoint) (*models.WebhookEndpoint, error) {
	if err := validateWebhookEndpoint(endpoint); err != nil {
		return nil, err
	}
	if endpoint.Secret == "" {
		current, err := w.repo.GetWebhookEndpoint(ctx, endpoint.ID)
		if err != nil {
			return nil, err
		}
		endpoint.Secret = current.Secret
	}
	return w.repo.UpdateWebhookEndpoint(ctx, endpoint)
}

func (w *WebhookService) DeleteEndpoint(ctx context.Context, id int64) error {
	return w.repo.DeleteWebhookEndpoint(ctx, id)
}

func (w *WebhookService) Queue(ctx context.Context, event WebhookEvent) {
	if event.Status != "" && event.Status != "success" {
		return
	}
	if event.EventType == "" {
		event.EventType = strings.Trim(event.ResourceType+"."+event.Action, ".")
	}
	if event.SchemaVersion == "" {
		event.SchemaVersion = WebhookEventSchemaVersion
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now().UTC()
	}

	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("webhook: failed to marshal event %s: %v", event.EventType, err)
		return
	}

	endpoints, err := w.repo.ListActiveWebhookEndpoints(ctx)
	if err != nil {
		log.Printf("webhook: failed to load endpoints: %v", err)
		return
	}
	for _, endpoint := range endpoints {
		if !endpointAllowsEvent(endpoint, event.EventType) {
			continue
		}
		if !endpointAllowsResource(endpoint, event) {
			continue
		}
		if _, err := w.repo.CreateWebhookDelivery(ctx, endpoint.ID, event.EventType, string(payload)); err != nil {
			log.Printf("webhook: failed to queue delivery endpoint=%d event=%s: %v", endpoint.ID, event.EventType, err)
		}
	}
}

func (w *WebhookService) ProcessQueue(ctx context.Context) {
	deliveries, err := w.repo.GetPendingWebhookDeliveries(ctx, 50)
	if err != nil {
		log.Printf("webhook: failed to load pending deliveries: %v", err)
		return
	}
	for _, delivery := range deliveries {
		if ctx.Err() != nil {
			return
		}
		endpoint, err := w.repo.GetWebhookEndpoint(ctx, delivery.EndpointID)
		if err != nil {
			msg := fmt.Sprintf("endpoint unavailable: %v", err)
			_ = w.repo.MarkWebhookFailed(ctx, delivery.ID, msg, delivery.RetryCount+1, nil, nil)
			continue
		}
		if !endpoint.IsActive {
			_ = w.repo.MarkWebhookFailed(ctx, delivery.ID, "endpoint inactive", delivery.RetryCount+1, nil, nil)
			continue
		}
		statusCode, err := w.deliver(ctx, endpoint, delivery)
		if err == nil && statusCode >= 200 && statusCode < 300 {
			_ = w.repo.MarkWebhookDelivered(ctx, delivery.ID, statusCode)
			continue
		}

		errMsg := "non-2xx response"
		if err != nil {
			errMsg = err.Error()
		}
		newRetryCount := delivery.RetryCount + 1
		var nextRetryAt *time.Time
		if newRetryCount < maxWebhookRetries {
			t := time.Now().Add(time.Duration(newRetryCount*newRetryCount) * time.Minute)
			nextRetryAt = &t
		}
		var statusPtr *int
		if statusCode > 0 {
			statusPtr = &statusCode
		}
		if markErr := w.repo.MarkWebhookFailed(ctx, delivery.ID, errMsg, newRetryCount, nextRetryAt, statusPtr); markErr != nil {
			log.Printf("webhook: failed to mark delivery failed id=%d: %v", delivery.ID, markErr)
		}
	}
}

func (w *WebhookService) StartWorker(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				w.ProcessQueue(ctx)
			}
		}
	}()
}

func (w *WebhookService) deliver(ctx context.Context, endpoint *models.WebhookEndpoint, delivery *models.WebhookDelivery) (int, error) {
	if isPrivateURL(endpoint.URL) {
		return 0, fmt.Errorf("webhook endpoint URL resolves to a private or reserved address")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.URL, bytes.NewBufferString(delivery.Payload))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", webhookUserAgent)
	req.Header.Set("X-IPAM-Event", delivery.EventType)
	req.Header.Set("X-IPAM-Delivery", fmt.Sprintf("%d", delivery.ID))
	req.Header.Set("X-IPAM-Event-Schema-Version", WebhookEventSchemaVersion)
	if endpoint.Secret != "" {
		req.Header.Set("X-IPAM-Signature-256", signWebhookPayload(endpoint.Secret, []byte(delivery.Payload)))
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func SampleWebhookEventPayload(eventType string) WebhookEvent {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	resourceID := int64(1001)
	userID := int64(42)
	switch eventType {
	case "device.created":
		return WebhookEvent{
			SchemaVersion: WebhookEventSchemaVersion,
			EventType:     "device.created",
			Action:        "created",
			ResourceType:  "device",
			ResourceID:    &resourceID,
			ResourceName:  "core-router-01",
			UserID:        &userID,
			Username:      "admin",
			Status:        "success",
			NewValues: map[string]any{
				"hostname": "core-router-01",
				"vendor":   "Juniper",
				"model":    "MX204",
			},
			OccurredAt: now,
		}
	default:
		return WebhookEvent{
			SchemaVersion: WebhookEventSchemaVersion,
			EventType:     "ip_address.assigned",
			Action:        "assigned",
			ResourceType:  "ip_address",
			ResourceID:    &resourceID,
			ResourceName:  "10.0.0.10",
			UserID:        &userID,
			Username:      "admin",
			Status:        "success",
			NewValues: map[string]any{
				"address":     "10.0.0.10",
				"subnet_cidr": "10.0.0.0/24",
				"assigned_to": "ops",
			},
			OccurredAt: now,
		}
	}
}

// isPrivateURL returns true if the URL resolves to a private, loopback, or link-local address.
// Hosts that fail to resolve are also considered unsafe (fail-closed).
func isPrivateURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return true // treat unparseable URLs as unsafe
	}
	host := u.Hostname()
	ips, err := net.LookupHost(host)
	if err != nil || len(ips) == 0 {
		return true
	}
	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			return true
		}
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return true
		}
	}
	return false
}

// resolvesToPrivateIP returns true only if the URL's host is an IP literal or resolves
// via DNS to a private, loopback, or link-local address. Unlike isPrivateURL, DNS
// resolution failures are not treated as a block — only confirmed private addresses are.
func resolvesToPrivateIP(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := u.Hostname()
	// If the host is an IP literal, check it directly.
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
	}
	// For hostnames, only block if DNS resolves and every address is private.
	ips, err := net.LookupHost(host)
	if err != nil || len(ips) == 0 {
		return false // can't confirm — allow creation
	}
	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return true
		}
	}
	return false
}

func validateWebhookEndpoint(endpoint *models.WebhookEndpoint) error {
	endpoint.Name = strings.TrimSpace(endpoint.Name)
	endpoint.URL = strings.TrimSpace(endpoint.URL)
	if endpoint.Name == "" {
		return fmt.Errorf("name is required")
	}
	if endpoint.URL == "" {
		return fmt.Errorf("url is required")
	}
	u, err := url.Parse(endpoint.URL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("url must be absolute")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("url must use http or https")
	}
	if resolvesToPrivateIP(endpoint.URL) {
		return fmt.Errorf("url must not resolve to a private or reserved address")
	}
	cleanEvents := make([]string, 0, len(endpoint.Events))
	for _, event := range endpoint.Events {
		event = strings.TrimSpace(event)
		if event != "" {
			cleanEvents = append(cleanEvents, event)
		}
	}
	endpoint.Events = cleanEvents
	endpoint.ObjectTypes = cleanStringList(endpoint.ObjectTypes)
	endpoint.TagFilters = cleanStringList(endpoint.TagFilters)
	if endpoint.FilterConditions == nil {
		endpoint.FilterConditions = map[string]string{}
	}
	return nil
}

func cleanStringList(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func endpointAllowsEvent(endpoint *models.WebhookEndpoint, eventType string) bool {
	if len(endpoint.Events) == 0 {
		return true
	}
	for _, allowed := range endpoint.Events {
		if allowed == "*" || allowed == eventType {
			return true
		}
		if strings.HasSuffix(allowed, ".*") && strings.HasPrefix(eventType, strings.TrimSuffix(allowed, "*")) {
			return true
		}
	}
	return false
}

func endpointAllowsResource(endpoint *models.WebhookEndpoint, event WebhookEvent) bool {
	if len(endpoint.ObjectTypes) > 0 {
		matched := false
		for _, objectType := range endpoint.ObjectTypes {
			if objectType == "*" || objectType == event.ResourceType {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if len(endpoint.FilterConditions) == 0 && len(endpoint.TagFilters) == 0 {
		return true
	}
	values := eventValues(event)
	for key, expected := range endpoint.FilterConditions {
		if expected == "" || expected == "*" {
			continue
		}
		actual, ok := values[key]
		if !ok {
			return false
		}
		if strings.HasSuffix(expected, "*") {
			if !strings.HasPrefix(actual, strings.TrimSuffix(expected, "*")) {
				return false
			}
			continue
		}
		if actual != expected {
			return false
		}
	}
	if len(endpoint.TagFilters) > 0 {
		tag := values["tag"]
		tagID := values["tag_id"]
		for _, allowed := range endpoint.TagFilters {
			if allowed == "*" || allowed == tag || allowed == tagID {
				return true
			}
		}
		return false
	}
	return true
}

func eventValues(event WebhookEvent) map[string]string {
	values := map[string]string{
		"event_type":    event.EventType,
		"action":        event.Action,
		"resource_type": event.ResourceType,
		"resource_name": event.ResourceName,
		"username":      event.Username,
	}
	if event.ResourceID != nil {
		values["resource_id"] = fmt.Sprintf("%d", *event.ResourceID)
	}
	if event.UserID != nil {
		values["user_id"] = fmt.Sprintf("%d", *event.UserID)
	}
	mergeEventValues(values, event.NewValues)
	return values
}

func mergeEventValues(values map[string]string, source interface{}) {
	switch typed := source.(type) {
	case map[string]string:
		for k, v := range typed {
			values[k] = v
		}
	case map[string]interface{}:
		for k, v := range typed {
			values[k] = fmt.Sprint(v)
		}
	}
}

func signWebhookPayload(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
