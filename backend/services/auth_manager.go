package services

// AuthManager bundles identity and authentication sub-services.
type AuthManager struct {
	Email        *EmailService
	Registration *RegistrationService
	MFA          *MFAService
	Notification *NotificationService
	LDAP         *LDAPService
	OAuth2       *OAuth2Service
	SAML         *SAMLService
}
