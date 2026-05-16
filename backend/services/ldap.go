package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"strings"

	ldap "github.com/go-ldap/ldap/v3"
	"ipam-next/models"
	"ipam-next/repository"
)

// LDAPService manages LDAP / Active Directory authentication.
type LDAPService struct {
	repository    *repository.Repository
	encryptionKey string
}

// NewLDAPService creates a new LDAPService.
func NewLDAPService(repo *repository.Repository, encryptionKey string) *LDAPService {
	return &LDAPService{repository: repo, encryptionKey: encryptionKey}
}

// GetConfig retrieves the LDAP configuration from the database.
// Returns nil, nil if no configuration has been saved yet.
func (s *LDAPService) GetConfig(ctx context.Context) (*models.LDAPConfig, error) {
	return s.repository.GetLDAPConfig(ctx)
}

// SaveConfig encrypts the bind password and persists the LDAP configuration.
func (s *LDAPService) SaveConfig(ctx context.Context, cfg *models.LDAPConfig) error {
	if len(cfg.BindPasswordEnc) > 0 {
		enc, err := EncryptBytesWithKey(s.encryptionKey, cfg.BindPasswordEnc)
		if err != nil {
			return fmt.Errorf("encrypting bind password: %w", err)
		}
		cfg.BindPasswordEnc = enc
	}
	return s.repository.UpsertLDAPConfig(ctx, cfg)
}

// bindPassword decrypts the stored bind password from the config.
func (s *LDAPService) bindPassword(cfg *models.LDAPConfig) (string, error) {
	if len(cfg.BindPasswordEnc) == 0 {
		return "", nil
	}
	pt, err := DecryptBytesWithKey(s.encryptionKey, cfg.BindPasswordEnc)
	if err != nil {
		return "", fmt.Errorf("decrypting bind password: %w", err)
	}
	return string(pt), nil
}

// dial opens an LDAP connection according to the TLS mode in cfg.
func (s *LDAPService) dial(cfg *models.LDAPConfig) (*ldap.Conn, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	if cfg.TLSSkipVerify {
		slog.Warn("LDAP TLS certificate verification is disabled — do not use in production", "host", cfg.Host)
	}
	tlsCfg := &tls.Config{InsecureSkipVerify: cfg.TLSSkipVerify} //nolint:gosec

	switch cfg.TLSMode {
	case "tls":
		return ldap.DialTLS("tcp", addr, tlsCfg)
	default:
		conn, err := ldap.DialURL(fmt.Sprintf("ldap://%s", addr))
		if err != nil {
			return nil, err
		}
		if cfg.TLSMode == "starttls" {
			if err := conn.StartTLS(tlsCfg); err != nil {
				conn.Close()
				return nil, err
			}
		}
		return conn, nil
	}
}

// TestConnection opens and binds an LDAP connection to verify the config is correct.
// Returns nil on success.
func (s *LDAPService) TestConnection(ctx context.Context) error {
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("loading LDAP config: %w", err)
	}
	if cfg == nil {
		return fmt.Errorf("LDAP not configured")
	}

	conn, err := s.dial(cfg)
	if err != nil {
		return fmt.Errorf("connecting to LDAP server: %w", err)
	}
	defer conn.Close()

	bindPw, err := s.bindPassword(cfg)
	if err != nil {
		return err
	}

	if cfg.BindDN != "" {
		if err := conn.Bind(cfg.BindDN, bindPw); err != nil {
			return fmt.Errorf("LDAP bind failed: %w", err)
		}
	} else {
		if err := conn.UnauthenticatedBind(""); err != nil {
			return fmt.Errorf("LDAP anonymous bind failed: %w", err)
		}
	}
	return nil
}

// Authenticate verifies username+password against LDAP, then finds or creates a local user.
func (s *LDAPService) Authenticate(ctx context.Context, username, password string) (*models.User, error) {
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading LDAP config: %w", err)
	}
	if cfg == nil || !cfg.Enabled {
		return nil, fmt.Errorf("LDAP authentication is not enabled")
	}

	// Open service account connection for searching
	conn, err := s.dial(cfg)
	if err != nil {
		return nil, fmt.Errorf("connecting to LDAP server: %w", err)
	}
	defer conn.Close()

	bindPw, err := s.bindPassword(cfg)
	if err != nil {
		return nil, err
	}

	// Bind with service account to search for user
	if cfg.BindDN != "" {
		if err := conn.Bind(cfg.BindDN, bindPw); err != nil {
			return nil, fmt.Errorf("LDAP service bind failed: %w", err)
		}
	}

	// Search for the user
	filter := fmt.Sprintf(cfg.UserFilter, ldap.EscapeFilter(username))
	searchReq := ldap.NewSearchRequest(
		cfg.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{"dn", cfg.UsernameAttr, cfg.EmailAttr, "memberOf"},
		nil,
	)
	sr, err := conn.Search(searchReq)
	if err != nil {
		return nil, fmt.Errorf("LDAP search failed: %w", err)
	}
	if len(sr.Entries) == 0 {
		return nil, fmt.Errorf("user not found in LDAP")
	}
	entry := sr.Entries[0]
	userDN := entry.DN

	// Verify password by binding as the user
	if err := conn.Bind(userDN, password); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Extract attributes
	ldapEmail := entry.GetAttributeValue(cfg.EmailAttr)
	ldapUsername := entry.GetAttributeValue(cfg.UsernameAttr)
	if ldapUsername == "" {
		ldapUsername = username
	}
	if ldapEmail == "" {
		ldapEmail = ldapUsername + "@ldap.local"
	}

	// Find or create local user
	user, err := s.repository.FindUserByExternalAuth(ctx, "ldap", userDN)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if user == nil {
		user, err = s.repository.CreateExternalUser(ctx, ldapUsername, ldapEmail, "ldap", userDN)
		if err != nil {
			return nil, fmt.Errorf("creating local user: %w", err)
		}
	}

	// Sync group memberships
	groups := entry.GetAttributeValues("memberOf")
	if err := s.SyncGroups(ctx, user.ID, groups); err != nil {
		// Non-fatal — log but don't block login
		_ = err
	}

	return user, nil
}

// SyncGroups maps LDAP group DNs to local roles and assigns them to the user.
func (s *LDAPService) SyncGroups(ctx context.Context, userID int64, ldapGroups []string) error {
	mappings, err := s.repository.GetLDAPGroupMappings(ctx)
	if err != nil {
		return fmt.Errorf("loading group mappings: %w", err)
	}

	groupSet := make(map[string]bool, len(ldapGroups))
	for _, g := range ldapGroups {
		groupSet[strings.ToLower(g)] = true
	}

	for _, m := range mappings {
		if groupSet[strings.ToLower(m.LDAPGroupDN)] {
			// Best-effort role assignment; ignore duplicates
			_ = s.repository.AssignRoleToUser(ctx, userID, m.RoleID)
		}
	}
	return nil
}
