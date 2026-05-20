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
	Discovery     *DiscoveryService
	DNS           *DNSService
	Reports       *ReportsService
	Import        *ImportService
	Webhooks      *WebhookService
	Automation    *AutomationService
	LDAP          *LDAPService
	OAuth2        *OAuth2Service
	SAML          *SAMLService
	Topology      *TopologyService
	Jobs          *JobService

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

	svc := &Service{
		repository:             repo,
		encryptionKey:          mfaEncryptionKey,
		Config:                 configSvc,
		Email:                  emailSvc,
		Registration:           registrationSvc,
		MFA:                    mfaSvc,
		Notification:           NewNotificationService(repo, emailSvc),
		Discovery:              NewDiscoveryService(repo, configSvc, mfaEncryptionKey),
		LDAP:                   NewLDAPService(repo, mfaEncryptionKey),
		OAuth2:                 NewOAuth2Service(repo, mfaEncryptionKey),
		SAML:                   NewSAMLService(repo, mfaEncryptionKey),
		Jobs:                   NewJobService(),
		dashboardSummaryCache:  newTTLCache[*models.DashboardSummary](30 * time.Second),
		dashboardActivityCache: newTTLCache[[]*models.DashboardActivity](15 * time.Second),
	}
	svc.Audit = NewAuditService(svc)
	svc.DNS = NewDNSService(svc)
	svc.Reports = NewReportsService(repo, configSvc, emailSvc, svc.Audit)
	svc.Import = NewImportService(repo)
	svc.Webhooks = NewWebhookService(repo)
	svc.Automation = NewAutomationService(svc)
	svc.Topology = NewTopologyService(repo)
	return svc
}

// GetRepository returns the underlying repository
func (s *Service) GetRepository() *repository.Repository {
	return s.repository
}
