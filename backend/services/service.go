package services

import (
	"log"

	"ipam-next/repository"
)

type Service struct {
	repository   *repository.Repository
	Config       *ConfigService
	Email        *EmailService
	Registration *RegistrationService
	MFA          *MFAService
}

func NewService(repo *repository.Repository, mfaEncryptionKey string) *Service {
	configSvc := NewConfigService(repo)
	emailSvc := NewEmailService(configSvc)
	registrationSvc := NewRegistrationService(repo, configSvc, emailSvc)

	mfaSvc, err := NewMFAService(repo, mfaEncryptionKey)
	if err != nil {
		log.Fatalf("Failed to initialize MFA service: %v", err)
	}

	return &Service{
		repository:   repo,
		Config:       configSvc,
		Email:        emailSvc,
		Registration: registrationSvc,
		MFA:          mfaSvc,
	}
}

// GetRepository returns the underlying repository
func (s *Service) GetRepository() *repository.Repository {
	return s.repository
}
