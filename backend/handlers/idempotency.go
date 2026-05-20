package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"padduck/models"
)

const idempotencyTTL = 24 * time.Hour

type idempotencyStore struct {
	mu      sync.Mutex
	entries map[string]idempotencyEntry
	now     func() time.Time
}

type idempotencyEntry struct {
	Fingerprint string
	StatusCode  int
	ContentType string
	Body        []byte
	ExpiresAt   time.Time
}

func newIdempotencyStore() *idempotencyStore {
	return &idempotencyStore{
		entries: make(map[string]idempotencyEntry),
		now:     time.Now,
	}
}

func (s *idempotencyStore) get(key, fingerprint string) (idempotencyEntry, bool, bool) {
	if s == nil {
		return idempotencyEntry{}, false, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	for entryKey, entry := range s.entries {
		if now.After(entry.ExpiresAt) {
			delete(s.entries, entryKey)
		}
	}

	entry, ok := s.entries[key]
	if !ok {
		return idempotencyEntry{}, false, false
	}
	if entry.Fingerprint != fingerprint {
		return idempotencyEntry{}, true, true
	}
	return entry, true, false
}

func (s *idempotencyStore) set(key, fingerprint string, statusCode int, contentType string, body []byte) {
	if s == nil || statusCode < 200 || statusCode >= 300 {
		return
	}
	s.mu.Lock()
	s.entries[key] = idempotencyEntry{
		Fingerprint: fingerprint,
		StatusCode:  statusCode,
		ContentType: contentType,
		Body:        append([]byte(nil), body...),
		ExpiresAt:   s.now().Add(idempotencyTTL),
	}
	s.mu.Unlock()
}

func (h *Handler) IdempotencyMiddleware(c *fiber.Ctx) error {
	key := strings.TrimSpace(c.Get("Idempotency-Key"))
	if key == "" {
		key = idempotencyKeyFromBody(c.Body())
	}
	if key == "" {
		return c.Next()
	}
	if h.idempotency == nil {
		h.idempotency = newIdempotencyStore()
	}

	scope := idempotencyScope(c, key)
	fingerprint := idempotencyFingerprint(c)
	entry, found, conflict := h.idempotency.get(scope, fingerprint)
	if conflict {
		return RespondError(c, fiber.StatusConflict, ErrConflict, "idempotency key was already used with a different request")
	}
	if found {
		c.Set("X-Idempotent-Replay", "true")
		if entry.ContentType != "" {
			c.Set(fiber.HeaderContentType, entry.ContentType)
		}
		return c.Status(entry.StatusCode).Send(entry.Body)
	}

	if err := c.Next(); err != nil {
		return err
	}
	h.idempotency.set(scope, fingerprint, c.Response().StatusCode(), string(c.Response().Header.ContentType()), c.Response().Body())
	return nil
}

func idempotencyScope(c *fiber.Ctx, key string) string {
	userID := "anonymous"
	if user := currentUserID(c); user != "" {
		userID = user
	}
	return fmt.Sprintf("%s:%s:%s:%s", userID, c.Method(), c.Route().Path, key)
}

func currentUserID(c *fiber.Ctx) string {
	user, ok := c.Locals("user").(*models.User)
	if ok && user != nil {
		return fmt.Sprintf("%d", user.ID)
	}
	if u := c.Locals("user"); u != nil {
		return fmt.Sprintf("%v", u)
	}
	return ""
}

func idempotencyFingerprint(c *fiber.Ctx) string {
	sum := sha256.Sum256(append([]byte(c.Method()+" "+c.Route().Path+"\n"), c.Body()...))
	return hex.EncodeToString(sum[:])
}

func idempotencyKeyFromBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var payload struct {
		IdempotencyKey string `json:"idempotency_key"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	return strings.TrimSpace(payload.IdempotencyKey)
}
