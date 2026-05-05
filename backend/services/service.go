package services

import (
	"ipam-next/repository"
)

type Service struct {
	repository   *repository.Repository
	Config       *ConfigService
	Email        *EmailService
	Registration *RegistrationService
}

func NewService(repo *repository.Repository) *Service {
	configSvc := NewConfigService(repo)
	emailSvc := NewEmailService(configSvc)
	registrationSvc := NewRegistrationService(repo, configSvc, emailSvc)

	return &Service{
		repository:   repo,
		Config:       configSvc,
		Email:        emailSvc,
		Registration: registrationSvc,
	}
}

// GetRepository returns the underlying repository
func (s *Service) GetRepository() *repository.Repository {
	return s.repository
}
