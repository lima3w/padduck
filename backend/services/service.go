package services

import (
	"log"
	"time"

	"padduck/models"
	"padduck/repository"
)

type Service struct {
	repository    *repository.Repository
	encryptionKey string
	Config        *ConfigService
	Email         *EmailService
	Registration  *RegistrationService
	MFA           *MFAService
	Audit         *AuditService
	Notification  *NotificationService
	LDAP          *LDAPService
	OAuth2        *OAuth2Service
	SAML          *SAMLService
	Ops           *OpsManager

	dashboardSummaryCache  *ttlCache[*models.DashboardSummary]
	dashboardActivityCache *ttlCache[[]*models.DashboardActivity]
}

func NewService(repo *repository.Repository, mfaEncryptionKey string) *Service {
	configSvc := NewConfigService(repo)
	emailSvc := NewEmailService(configSvc)
	registrationSvc := NewRegistrationService(repo, configSvc, emailSvc)

	mfaSvc, err := NewMFAService(repo, mfaEncryptionKey)
	if err != nil {
		log.Fatalf("Failed to initialize MFA service: %v", err)
	}

	ldapSvc := NewLDAPService(repo, mfaEncryptionKey)
	oauth2Svc := NewOAuth2Service(repo, mfaEncryptionKey)
	samlSvc := NewSAMLService(repo, mfaEncryptionKey)

	svc := &Service{
		repository:             repo,
		encryptionKey:          mfaEncryptionKey,
		Config:                 configSvc,
		Email:                  emailSvc,
		Registration:           registrationSvc,
		MFA:                    mfaSvc,
		Notification:           NewNotificationService(repo, emailSvc),
		LDAP:                   ldapSvc,
		OAuth2:                 oauth2Svc,
		SAML:                   samlSvc,
		dashboardSummaryCache:  newTTLCache[*models.DashboardSummary](30 * time.Second),
		dashboardActivityCache: newTTLCache[[]*models.DashboardActivity](15 * time.Second),
	}
	svc.Audit = NewAuditService(svc)
	svc.Ops = &OpsManager{
		Discovery:  NewDiscoveryService(repo, configSvc, mfaEncryptionKey),
		Reports:    NewReportsService(repo, configSvc, emailSvc, svc.Audit),
		Import:     NewImportService(repo),
		Jobs:       NewJobService(),
		Webhooks:   NewWebhookService(repo),
		Topology:   NewTopologyService(repo),
		DNS:        NewDNSService(configSvc, repo),
		Automation: NewAutomationService(repo, svc),
		Telemetry:  newTelemetryService(configSvc, repo, ldapSvc, oauth2Svc, samlSvc),
	}
	return svc
}

// GetRepository returns the underlying repository
func (s *Service) GetRepository() *repository.Repository {
	return s.repository
}
