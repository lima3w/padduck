package services

import (
	"context"
	"fmt"
	"net/mail"

	"ipam-next/models"
)

// CreateUser creates a new user with validation
func (s *Service) CreateUser(ctx context.Context, username, email string) (*models.User, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}

	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return nil, fmt.Errorf("invalid email format: %w", err)
	}

	return s.repository.CreateUser(ctx, username, email)
}

// GetUser retrieves a user by ID
func (s *Service) GetUser(ctx context.Context, id int64) (*models.User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	return s.repository.GetUserByID(ctx, id)
}
