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

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	return s.repository.GetUserByID(ctx, id)
}

// ListAllUsers returns all users
func (s *Service) ListAllUsers(ctx context.Context) ([]*models.User, error) {
	return s.repository.ListAllUsers(ctx)
}

// CreateUserWithPassword creates a new user with password hash
func (s *Service) CreateUserWithPassword(ctx context.Context, username, email, passwordHash, role string) (*models.User, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}

	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return nil, fmt.Errorf("invalid email format: %w", err)
	}

	if role != "admin" && role != "user" && role != "viewer" {
		return nil, fmt.Errorf("invalid role")
	}

	return s.repository.CreateUserWithPassword(ctx, username, email, passwordHash, role)
}

// UpdateUserRole updates a user's role
func (s *Service) UpdateUserRole(ctx context.Context, userID int64, role string) (*models.User, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	if role != "admin" && role != "user" && role != "viewer" {
		return nil, fmt.Errorf("invalid role")
	}

	return s.repository.UpdateUserRole(ctx, userID, role)
}

// DeleteUser deletes a user
func (s *Service) DeleteUser(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}

	return s.repository.DeleteUser(ctx, userID)
}
