package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"padduck/models"
	"padduck/repository"
)

// samlRequestEntry holds a pending AuthnRequest ID and its expiry time.
type samlRequestEntry struct {
	expiresAt time.Time
}

// SAMLService manages SAML 2.0 Service Provider authentication.
type SAMLService struct {
	repository    *repository.Repository
	encryptionKey string

	// pendingRequests tracks outstanding AuthnRequest IDs to prevent assertion
	// replay attacks.  A sync.Map is used so no additional locking is required
	// for concurrent login flows.  Entries are removed after successful
	// assertion validation or after a 10-minute TTL.
	pendingRequests sync.Map
}

// NewSAMLService creates a new SAMLService.
func NewSAMLService(repo *repository.Repository, encryptionKey string) *SAMLService {
	return &SAMLService{repository: repo, encryptionKey: encryptionKey}
}

// GetConfig retrieves the SAML configuration from the database.
// Returns nil, nil if no configuration has been saved yet.
func (s *SAMLService) GetConfig(ctx context.Context) (*models.SAMLConfig, error) {
	return s.repository.GetSAMLConfig(ctx)
}

// SaveConfig persists the SAML configuration, auto-generating SP key/cert if empty.
func (s *SAMLService) SaveConfig(ctx context.Context, cfg *models.SAMLConfig) error {
	if cfg.SPKeyPEM == "" || cfg.SPCertPEM == "" {
		key, cert, err := generateSelfSignedCert()
		if err != nil {
			return fmt.Errorf("generating SP key/cert: %w", err)
		}
		cfg.SPKeyPEM = key
		cfg.SPCertPEM = cert
	}
	return s.repository.UpsertSAMLConfig(ctx, cfg)
}

// generateSelfSignedCert creates a new RSA 2048 key and self-signed certificate, returning PEM strings.
func generateSelfSignedCert() (keyPEM, certPEM string, err error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "Padduck SAML SP"},
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return "", "", err
	}

	keyPEM = string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))
	certPEM = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}))
	return keyPEM, certPEM, nil
}

// buildMiddleware constructs a crewjam/saml SP middleware from the stored config.
func (s *SAMLService) buildMiddleware(ctx context.Context, cfg *models.SAMLConfig) (*samlsp.Middleware, error) {
	keyPair, err := tls.X509KeyPair([]byte(cfg.SPCertPEM), []byte(cfg.SPKeyPEM))
	if err != nil {
		return nil, fmt.Errorf("loading SP key pair: %w", err)
	}
	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("parsing SP certificate: %w", err)
	}

	rootURL, err := url.Parse(cfg.EntityID)
	if err != nil {
		return nil, fmt.Errorf("parsing entity ID as URL: %w", err)
	}

	opts := samlsp.Options{
		URL:         *rootURL,
		Key:         keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate: keyPair.Leaf,
	}

	if cfg.IDPMetadataURL != "" {
		metadataURL, err := url.Parse(cfg.IDPMetadataURL)
		if err != nil {
			return nil, fmt.Errorf("parsing IdP metadata URL: %w", err)
		}
		idpMetadata, err := samlsp.FetchMetadata(ctx, http.DefaultClient, *metadataURL)
		if err != nil {
			return nil, fmt.Errorf("fetching IdP metadata: %w", err)
		}
		opts.IDPMetadata = idpMetadata
	} else if cfg.IDPMetadataXML != "" {
		idpMetadata, err := samlsp.ParseMetadata([]byte(cfg.IDPMetadataXML))
		if err != nil {
			return nil, fmt.Errorf("parsing IdP metadata XML: %w", err)
		}
		opts.IDPMetadata = idpMetadata
	} else {
		return nil, fmt.Errorf("either idp_metadata_url or idp_metadata_xml must be configured")
	}

	sp, err := samlsp.New(opts)
	if err != nil {
		return nil, fmt.Errorf("building SAML SP: %w", err)
	}
	return sp, nil
}

// GetSPMetadata returns the SP metadata XML, auto-generating key/cert if needed.
func (s *SAMLService) GetSPMetadata(ctx context.Context) ([]byte, error) {
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading SAML config: %w", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("SAML not configured")
	}
	// Auto-generate SP cert if missing
	if cfg.SPKeyPEM == "" || cfg.SPCertPEM == "" {
		key, cert, err := generateSelfSignedCert()
		if err != nil {
			return nil, err
		}
		cfg.SPKeyPEM = key
		cfg.SPCertPEM = cert
		if err := s.repository.UpsertSAMLConfig(ctx, cfg); err != nil {
			return nil, err
		}
	}

	sp, err := s.buildMiddleware(ctx, cfg)
	if err != nil {
		return nil, err
	}

	metadata := sp.ServiceProvider.Metadata()
	buf, err := marshalXML(metadata)
	if err != nil {
		return nil, fmt.Errorf("marshalling SP metadata: %w", err)
	}
	return buf, nil

}

// marshalXML serialises a SAML EntityDescriptor to XML bytes.
func marshalXML(v *saml.EntityDescriptor) ([]byte, error) {
	return xml.MarshalIndent(v, "", "  ")
}

// GetLoginURL returns the SP-initiated AuthnRequest redirect URL and stores
// the AuthnRequest ID so it can be validated during assertion processing to
// prevent replay attacks.
func (s *SAMLService) GetLoginURL(ctx context.Context, relayState string) (string, error) {
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("loading SAML config: %w", err)
	}
	if cfg == nil || !cfg.Enabled {
		return "", fmt.Errorf("SAML authentication is not enabled")
	}

	sp, err := s.buildMiddleware(ctx, cfg)
	if err != nil {
		return "", err
	}

	authnReq, err := sp.ServiceProvider.MakeAuthenticationRequest(
		sp.ServiceProvider.GetSSOBindingLocation(saml.HTTPRedirectBinding),
		saml.HTTPRedirectBinding,
		saml.HTTPPostBinding,
	)
	if err != nil {
		return "", fmt.Errorf("building SAML AuthnRequest: %w", err)
	}

	// Store the request ID with a 10-minute TTL for InResponseTo validation.
	s.pendingRequests.Store(authnReq.ID, samlRequestEntry{expiresAt: time.Now().Add(10 * time.Minute)})

	// Prune expired entries opportunistically to prevent unbounded growth.
	s.pruneExpiredRequests()

	redirectURL, err := authnReq.Redirect(relayState, &sp.ServiceProvider)
	if err != nil {
		return "", fmt.Errorf("building SAML redirect URL: %w", err)
	}
	return redirectURL.String(), nil
}

// pruneExpiredRequests removes expired AuthnRequest IDs from the pending map.
func (s *SAMLService) pruneExpiredRequests() {
	now := time.Now()
	s.pendingRequests.Range(func(key, value interface{}) bool {
		if entry, ok := value.(samlRequestEntry); ok && now.After(entry.expiresAt) {
			s.pendingRequests.Delete(key)
		}
		return true
	})
}

// pendingRequestIDs returns all non-expired AuthnRequest IDs currently stored.
func (s *SAMLService) pendingRequestIDs() []string {
	now := time.Now()
	var ids []string
	s.pendingRequests.Range(func(key, value interface{}) bool {
		if entry, ok := value.(samlRequestEntry); ok && now.Before(entry.expiresAt) {
			ids = append(ids, key.(string))
		}
		return true
	})
	return ids
}

// ProcessAssertion validates a SAML response and finds or creates a local user.
// samlResponse is the base64-encoded SAMLResponse POST parameter.
func (s *SAMLService) ProcessAssertion(ctx context.Context, samlResponse, acsURL string) (*models.User, error) {
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading SAML config: %w", err)
	}
	if cfg == nil || !cfg.Enabled {
		return nil, fmt.Errorf("SAML authentication is not enabled")
	}

	sp, err := s.buildMiddleware(ctx, cfg)
	if err != nil {
		return nil, err
	}

	rawXML, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		return nil, fmt.Errorf("decoding SAML response: %w", err)
	}

	parsedURL, err := url.Parse(acsURL)
	if err != nil || parsedURL.String() == "" {
		parsedURL = &url.URL{}
	}

	// Pass the stored AuthnRequest IDs so the library can validate InResponseTo,
	// preventing assertion replay attacks.
	requestIDs := s.pendingRequestIDs()
	assertion, err := sp.ServiceProvider.ParseXMLResponse(rawXML, requestIDs, *parsedURL)
	if err != nil {
		return nil, fmt.Errorf("invalid SAML assertion: %w", err)
	}

	// Remove the consumed request ID so it cannot be replayed.
	if assertion.Subject != nil {
		for _, sc := range assertion.Subject.SubjectConfirmations {
			if sc.SubjectConfirmationData != nil && sc.SubjectConfirmationData.InResponseTo != "" {
				s.pendingRequests.Delete(sc.SubjectConfirmationData.InResponseTo)
			}
		}
	}

	// Extract NameID as the external identifier
	nameID := assertion.Subject.NameID.Value
	if nameID == "" {
		return nil, fmt.Errorf("SAML assertion missing NameID")
	}

	// Extract email and username from attributes
	email := nameID
	username := nameID
	for _, stmt := range assertion.AttributeStatements {
		for _, attr := range stmt.Attributes {
			switch attr.Name {
			case "email", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress":
				if len(attr.Values) > 0 {
					email = attr.Values[0].Value
				}
			case "displayName", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name",
				"uid", "sAMAccountName":
				if len(attr.Values) > 0 {
					username = attr.Values[0].Value
				}
			}
		}
	}

	// Find or create local user
	user, err := s.repository.FindUserByExternalAuth(ctx, "saml", nameID)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if user == nil {
		user, err = s.repository.CreateExternalUser(ctx, username, email, "saml", nameID)
		if err != nil {
			return nil, fmt.Errorf("creating local user: %w", err)
		}
	}
	return user, nil
}
