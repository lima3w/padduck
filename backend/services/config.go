package services

import (
	"context"

	"ipam-next/models"
	"ipam-next/repository"
)

type ConfigService struct {
	repository *repository.Repository
}

func NewConfigService(repo *repository.Repository) *ConfigService {
	return &ConfigService{repository: repo}
}

func (s *ConfigService) Get(key string) (string, error) {
	cfg, err := s.repository.GetConfig(context.Background(), key)
	if err != nil {
		return "", err
	}
	return cfg.Value, nil
}

func (s *ConfigService) Set(key, value string) error {
	return s.repository.SetConfig(context.Background(), key, value)
}

func (s *ConfigService) List() ([]*models.Config, error) {
	return s.repository.ListConfigs(context.Background())
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
