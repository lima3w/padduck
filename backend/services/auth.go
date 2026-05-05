package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"ipam-next/models"
)

const (
	TokenLength = 32
)

// GenerateAPIToken creates a new API token for a user
func (s *Service) GenerateAPIToken(ctx context.Context, userID int64, tokenName string) (token string, err error) {
	if userID <= 0 {
		return "", fmt.Errorf("invalid user ID")
	}
	if tokenName == "" {
		return "", fmt.Errorf("token name is required")
	}

	// Generate random token
	tokenBytes := make([]byte, TokenLength)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}
	token = hex.EncodeToString(tokenBytes)

	// Hash token for storage
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	// Store token hash in database
	_, err = s.repository.CreateAPIToken(ctx, userID, tokenHash, tokenName)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ValidateAPIToken checks if a token is valid and returns the user
func (s *Service) ValidateAPIToken(ctx context.Context, token string) (*models.User, error) {
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	// Hash the provided token
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	// Look up token in database
	apiToken, err := s.repository.GetAPITokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	// Check if token has expired
	if apiToken.ExpiresAt != nil && apiToken.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token has expired")
	}

	// Update last used timestamp
	_ = s.repository.UpdateAPITokenLastUsed(ctx, apiToken.ID)

	// Get user information
	user, err := s.repository.GetUserByID(ctx, apiToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// RevokeAPIToken deletes a token
func (s *Service) RevokeAPIToken(ctx context.Context, tokenID int64) error {
	if tokenID <= 0 {
		return fmt.Errorf("invalid token ID")
	}

	return s.repository.DeleteAPIToken(ctx, tokenID)
}

// ListUserTokens returns all tokens for a user
func (s *Service) ListUserTokens(ctx context.Context, userID int64) ([]*models.APIToken, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	return s.repository.ListAPITokensByUser(ctx, userID)
}
