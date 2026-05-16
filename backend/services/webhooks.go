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
	"net/http"
	"net/url"
	"strings"
	"time"

	"ipam-next/models"
	"ipam-next/repository"
)

const maxWebhookRetries = 5

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
	EventType    string      `json:"event_type"`
	Action       string      `json:"action"`
	ResourceType string      `json:"resource_type"`
	ResourceID   *int64      `json:"resource_id,omitempty"`
	ResourceName string      `json:"resource_name,omitempty"`
	UserID       *int64      `json:"user_id,omitempty"`
	Username     string      `json:"username,omitempty"`
	Status       string      `json:"status"`
	OldValues    interface{} `json:"old_values,omitempty"`
	NewValues    interface{} `json:"new_values,omitempty"`
	OccurredAt   time.Time   `json:"occurred_at"`
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.URL, bytes.NewBufferString(delivery.Payload))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ipam-next-webhooks/1.0")
	req.Header.Set("X-IPAM-Event", delivery.EventType)
	req.Header.Set("X-IPAM-Delivery", fmt.Sprintf("%d", delivery.ID))
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
	cleanEvents := make([]string, 0, len(endpoint.Events))
	for _, event := range endpoint.Events {
		event = strings.TrimSpace(event)
		if event != "" {
			cleanEvents = append(cleanEvents, event)
		}
	}
	endpoint.Events = cleanEvents
	return nil
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

func signWebhookPayload(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
