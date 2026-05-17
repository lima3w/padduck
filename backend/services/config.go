package services

import (
	"context"
	"fmt"

	"ipam-next/models"
	"ipam-next/repository"
)

type ConfigService struct {
	repository *repository.Repository
}

func NewConfigService(repo *repository.Repository) *ConfigService {
	return &ConfigService{repository: repo}
}

// GetCtx retrieves a config value using the provided context, propagating
// cancellation and tracing to the underlying DB query.
func (s *ConfigService) GetCtx(ctx context.Context, key string) (string, error) {
	if s.repository == nil {
		return "", fmt.Errorf("config: no repository")
	}
	cfg, err := s.repository.GetConfig(ctx, key)
	if err != nil {
		return "", err
	}
	return cfg.Value, nil
}

// Get retrieves a config value. Callers that have a context should use GetCtx.
func (s *ConfigService) Get(key string) (string, error) {
	return s.GetCtx(context.Background(), key)
}

// SetCtx persists a config value using the provided context.
func (s *ConfigService) SetCtx(ctx context.Context, key, value string) error {
	if s.repository == nil {
		return fmt.Errorf("config: no repository")
	}
	return s.repository.SetConfig(ctx, key, value)
}

// Set persists a config value. Callers that have a context should use SetCtx.
func (s *ConfigService) Set(key, value string) error {
	return s.SetCtx(context.Background(), key, value)
}

// SetMultiple applies all key-value pairs atomically. If any write fails, none are persisted.
func (s *ConfigService) SetMultiple(pairs map[string]string) error {
	if s.repository == nil {
		return fmt.Errorf("config: no repository")
	}
	return s.repository.SetConfigMultiple(context.Background(), pairs)
}

// ListCtx returns all config entries using the provided context.
func (s *ConfigService) ListCtx(ctx context.Context) ([]*models.Config, error) {
	if s.repository == nil {
		return nil, fmt.Errorf("config: no repository")
	}
	return s.repository.ListConfigs(ctx)
}

// List returns all config entries. Callers that have a context should use ListCtx.
func (s *ConfigService) List() ([]*models.Config, error) {
	return s.ListCtx(context.Background())
}

func (s *ConfigService) IsRegistrationEnabled() bool {
	val, err := s.Get("registration_enabled")
	if err != nil {
		return true // default open
	}
	return val != "false"
}

func (s *ConfigService) IsSMTPConfigured() bool {
	host, err := s.Get("smtp_host")
	return err == nil && host != ""
}

func (s *ConfigService) IsEmailVerificationRequired() bool {
	if !s.IsSMTPConfigured() {
		return false
	}
	val, err := s.Get("require_email_verification")
	if err != nil {
		return false
	}
	return val == "true"
}

func (s *ConfigService) IsAdminApprovalRequired() bool {
	val, err := s.Get("require_admin_approval")
	if err != nil {
		return false
	}
	return val == "true"
}
