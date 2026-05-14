package services

import (
	"log"

	"ipam-next/repository"
)

type Service struct {
	repository      *repository.Repository
	encryptionKey   string
	Config          *ConfigService
	Email           *EmailService
	Registration    *RegistrationService
	MFA             *MFAService
	Audit           *AuditService
	Notification    *NotificationService
	Discovery       *DiscoveryService
	DNS             *DNSService
}

func NewService(repo *repository.Repository, mfaEncryptionKey string) *Service {
	configSvc := NewConfigService(repo)
	emailSvc := NewEmailService(configSvc)
	registrationSvc := NewRegistrationService(repo, configSvc, emailSvc)

	mfaSvc, err := NewMFAService(repo, mfaEncryptionKey)
	if err != nil {
		log.Fatalf("Failed to initialize MFA service: %v", err)
	}

	svc := &Service{
		repository:    repo,
		encryptionKey: mfaEncryptionKey,
		Config:        configSvc,
		Email:         emailSvc,
		Registration:  registrationSvc,
		MFA:           mfaSvc,
		Notification:  NewNotificationService(repo, emailSvc),
		Discovery:     NewDiscoveryService(repo),
	}
	svc.Audit = NewAuditService(svc)
	svc.DNS = NewDNSService(svc)
	return svc
}

// GetRepository returns the underlying repository
func (s *Service) GetRepository() *repository.Repository {
	return s.repository
}
