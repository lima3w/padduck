package services

import (
	"context"
	"fmt"
	"time"
)

const (
	DefaultSessionTimeout = 24 * time.Hour
)

// UpdateLastLogin updates the user's last login timestamp
func (s *Service) UpdateLastLogin(ctx context.Context, userID int64) error {
	return s.repository.UpdateLastLogin(ctx, userID)
}

// IsSessionExpired checks if a session has expired
func (s *Service) IsSessionExpired(lastLoginAt *time.Time) bool {
	if lastLoginAt == nil {
		return false
	}

	expiresAt := lastLoginAt.Add(DefaultSessionTimeout)
	return time.Now().After(expiresAt)
}

// CheckSessionValid validates if a session is still valid
func (s *Service) CheckSessionValid(ctx context.Context, userID int64) (bool, error) {
	user, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("user not found")
	}

	if s.IsSessionExpired(user.LastLoginAt) {
		return false, fmt.Errorf("session expired")
	}

	return true, nil
}
