package services

import (
	"context"
	"log"

	"padduck/models"
	"padduck/repository"
)

type Service struct {
	repository    *repository.Repository
	encryptionKey string
	Config        *ConfigService
	Audit         *AuditService
	Auth          *AuthManager
	Ops           *OpsManager
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

	// Webhooks must be created before AuditService (Audit queues webhook events).
	webhookSvc := NewWebhookService(repo)
	auditSvc := NewAuditService(repo, configSvc, webhookSvc)

	notificationSvc := NewNotificationService(repo, emailSvc)

	svc := &Service{
		repository:    repo,
		encryptionKey: mfaEncryptionKey,
		Config:        configSvc,
		Audit:         auditSvc,
		Auth: &AuthManager{
			Email:        emailSvc,
			Registration: registrationSvc,
			MFA:          mfaSvc,
			Notification: notificationSvc,
			LDAP:         ldapSvc,
			OAuth2:       oauth2Svc,
			SAML:         samlSvc,
		},
	}
	dnsSvc := NewDNSService(configSvc, repo)
	identitySvc := NewIdentityService(repo, configSvc, emailSvc, mfaSvc, notificationSvc)
	infraSvc := NewInfrastructureService(repo, mfaEncryptionKey)
	svc.Ops = &OpsManager{
		Discovery:      NewDiscoveryService(repo, configSvc, mfaEncryptionKey),
		Reports:        NewReportsService(repo, configSvc, emailSvc, auditSvc),
		Import:         NewImportService(repo),
		Jobs:           NewJobService(),
		Webhooks:       webhookSvc,
		Topology:       NewTopologyService(repo),
		DNS:            dnsSvc,
		Automation:     NewAutomationService(repo, svc),
		Telemetry:      newTelemetryService(configSvc, repo, ldapSvc, oauth2Svc, samlSvc),
		NetworkModules: NewNetworkModulesService(repo),
		IPAM:           NewIPAMService(repo, configSvc, dnsSvc),
		Identity:       identitySvc,
		Infrastructure: infraSvc,
		Customers:      NewCustomerService(repo),
	}
	return svc
}

// GetRepository returns the underlying repository
func (s *Service) GetRepository() *repository.Repository {
	return s.repository
}

// AllocateIPAddress forwards to IPAMService to satisfy the automationIPAM interface.
func (s *Service) AllocateIPAddress(ctx context.Context, subnetID int64, deviceID *int64) (*models.IPAddress, error) {
	return s.Ops.IPAM.AllocateIPAddress(ctx, subnetID, deviceID)
}

// CreateIPAddress forwards to IPAMService to satisfy the automationIPAM interface.
func (s *Service) CreateIPAddress(ctx context.Context, subnetID int64, address, hostname, status string, tagID *int64, macAddress, ptrRecord, dnsName *string, customFields ...map[string]*string) (*models.IPAddress, error) {
	return s.Ops.IPAM.CreateIPAddress(ctx, subnetID, address, hostname, status, tagID, macAddress, ptrRecord, dnsName, customFields...)
}

// ReleaseIPAddress forwards to IPAMService to satisfy the automationIPAM interface.
func (s *Service) ReleaseIPAddress(ctx context.Context, id int64) (*models.IPAddress, error) {
	return s.Ops.IPAM.ReleaseIPAddress(ctx, id)
}

// CreateDevice forwards to InfrastructureService to satisfy the automationIPAM interface.
func (s *Service) CreateDevice(ctx context.Context, req *DeviceCreateRequest) (*models.Device, error) {
	return s.Ops.Infrastructure.CreateDevice(ctx, req)
}

// InitAdminPassword forwards to IdentityService (called from main.go startup).
func (s *Service) InitAdminPassword(ctx context.Context, password string) (bool, error) {
	return s.Ops.Identity.InitAdminPassword(ctx, password)
}

// ForceResetAdminPassword forwards to IdentityService (called from main.go startup).
func (s *Service) ForceResetAdminPassword(ctx context.Context, password string) error {
	return s.Ops.Identity.ForceResetAdminPassword(ctx, password)
}
