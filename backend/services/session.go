package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"padduck/models"
)

const (
	DefaultIdleTimeoutMinutes   = 60
	DefaultAbsoluteTimeoutHours = 168 // 7 days
	sessionTokenLength          = 32
)

// parseDeviceName extracts a human-readable device name from a user agent string.
func parseDeviceName(userAgent string) string {
	ua := userAgent
	if ua == "" {
		return "Unknown Device"
	}

	switch {
	case strings.Contains(ua, "iPhone"):
		return "iPhone"
	case strings.Contains(ua, "iPad"):
		return "iPad"
	case strings.Contains(ua, "Android"):
		if strings.Contains(ua, "Mobile") {
			return "Android Phone"
		}
		return "Android Tablet"
	case strings.Contains(ua, "Macintosh") || strings.Contains(ua, "Mac OS X"):
		return "Mac"
	case strings.Contains(ua, "Windows"):
		return "Windows PC"
	case strings.Contains(ua, "Linux"):
		return "Linux"
	case strings.Contains(ua, "curl"):
		return "curl"
	default:
		// Return first 50 chars of UA as fallback
		if len(ua) > 50 {
			return ua[:50]
		}
		return ua
	}
}

func (s *Service) sessionIdleTimeout(ctx context.Context) time.Duration {
	if val, err := s.Config.GetCtx(ctx, "session_idle_timeout_minutes"); err == nil && val != "" {
		var mins int
		if _, err := fmt.Sscanf(val, "%d", &mins); err == nil && mins > 0 {
			return time.Duration(mins) * time.Minute
		}
	}
	return DefaultIdleTimeoutMinutes * time.Minute
}

func (s *Service) sessionAbsoluteTimeout(ctx context.Context) time.Duration {
	if val, err := s.Config.GetCtx(ctx, "session_absolute_timeout_hours"); err == nil && val != "" {
		var hrs int
		if _, err := fmt.Sscanf(val, "%d", &hrs); err == nil && hrs > 0 {
			return time.Duration(hrs) * time.Hour
		}
	}
	return DefaultAbsoluteTimeoutHours * time.Hour
}

// CreateWebSession creates a new authenticated session and returns the raw token.
func (s *Service) CreateWebSession(ctx context.Context, userID int64, ipAddress, userAgent string) (string, error) {
	if userID <= 0 {
		return "", fmt.Errorf("invalid user ID")
	}

	tokenBytes := make([]byte, sessionTokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}
	rawToken := hex.EncodeToString(tokenBytes)

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	deviceName := parseDeviceName(userAgent)
	absoluteExpiry := time.Now().Add(s.sessionAbsoluteTimeout(ctx))

	_, err := s.repository.CreateSession(ctx, userID, tokenHash, deviceName, ipAddress, userAgent, absoluteExpiry)
	if err != nil {
		return "", fmt.Errorf("failed to store session: %w", err)
	}

	return rawToken, nil
}

// ValidateSession validates a session token and returns the user.
// It enforces both idle timeout and absolute timeout and updates last_used_at.
func (s *Service) ValidateSession(ctx context.Context, rawToken string) (*models.User, *models.Session, error) {
	if rawToken == "" {
		return nil, nil, fmt.Errorf("token is required")
	}

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	session, err := s.repository.GetSessionByHash(ctx, tokenHash)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid session")
	}

	now := time.Now()

	// Absolute timeout
	if now.After(session.AbsoluteExpiresAt) {
		_ = s.repository.DeleteSession(ctx, session.ID)
		return nil, nil, fmt.Errorf("session expired")
	}

	// Idle timeout
	idleTimeout := s.sessionIdleTimeout(ctx)
	if now.After(session.LastUsedAt.Add(idleTimeout)) {
		_ = s.repository.DeleteSession(ctx, session.ID)
		return nil, nil, fmt.Errorf("session expired due to inactivity")
	}

	// Refresh last_used_at (sliding idle timeout)
	_ = s.repository.UpdateSessionLastUsed(ctx, session.ID)

	user, err := s.repository.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found")
	}

	if user.State == "suspended" || user.State == "deleted" {
		_ = s.repository.DeleteSession(ctx, session.ID)
		return nil, nil, fmt.Errorf("account is %s", user.State)
	}

	return user, session, nil
}

// RevokeSession revokes a specific session by raw token.
func (s *Service) RevokeSession(ctx context.Context, userID int64, rawToken string) error {
	if rawToken == "" {
		return fmt.Errorf("token is required")
	}

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	session, err := s.repository.GetSessionByHash(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("session not found")
	}

	if session.UserID != userID {
		return fmt.Errorf("session does not belong to this user")
	}

	return s.repository.DeleteSession(ctx, session.ID)
}

// RevokeSessionByID revokes a session by its ID, verifying ownership.
func (s *Service) RevokeSessionByID(ctx context.Context, userID, sessionID int64) error {
	if sessionID <= 0 {
		return fmt.Errorf("invalid session ID")
	}

	sessions, err := s.repository.ListSessionsByUser(ctx, userID)
	if err != nil {
		return err
	}

	for _, sess := range sessions {
		if sess.ID == sessionID {
			return s.repository.DeleteSession(ctx, sessionID)
		}
	}

	return fmt.Errorf("session not found")
}

// RevokeAllSessions revokes all sessions for a user (logout all devices).
func (s *Service) RevokeAllSessions(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}
	return s.repository.DeleteAllUserSessions(ctx, userID)
}

// ListUserSessions returns all active sessions for a user.
func (s *Service) ListUserSessions(ctx context.Context, userID int64) ([]*models.Session, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}
	return s.repository.ListSessionsByUser(ctx, userID)
}

// UpdateLastLogin updates the user's last login timestamp
func (s *Service) UpdateLastLogin(ctx context.Context, userID int64) error {
	return s.repository.UpdateLastLogin(ctx, userID)
}

// IsSessionExpired is kept for backwards compatibility with API token auth.
func (s *Service) IsSessionExpired(lastLoginAt *time.Time) bool {
	if lastLoginAt == nil {
		return false
	}
	expiresAt := lastLoginAt.Add(DefaultSessionTimeout)
	return time.Now().After(expiresAt)
}

const DefaultSessionTimeout = 24 * time.Hour
