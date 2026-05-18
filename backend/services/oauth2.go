package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"ipam-next/models"
	"ipam-next/repository"
)

// decodeJSON decodes JSON from r into v.
func decodeJSON(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

// OAuth2Service manages OAuth2 / OIDC authentication.
type OAuth2Service struct {
	repository    *repository.Repository
	encryptionKey string
}

// NewOAuth2Service creates a new OAuth2Service.
func NewOAuth2Service(repo *repository.Repository, encryptionKey string) *OAuth2Service {
	return &OAuth2Service{repository: repo, encryptionKey: encryptionKey}
}

// GetConfig retrieves the OAuth2 configuration from the database.
// Returns nil, nil if no configuration has been saved yet.
func (s *OAuth2Service) GetConfig(ctx context.Context) (*models.OAuth2Config, error) {
	return s.repository.GetOAuth2Config(ctx)
}

// SaveConfig encrypts the client secret and persists the OAuth2 configuration.
func (s *OAuth2Service) SaveConfig(ctx context.Context, cfg *models.OAuth2Config) error {
	if len(cfg.ClientSecretEnc) > 0 {
		enc, err := EncryptBytesWithKey(s.encryptionKey, cfg.ClientSecretEnc)
		if err != nil {
			return fmt.Errorf("encrypting client secret: %w", err)
		}
		cfg.ClientSecretEnc = enc
	}
	return s.repository.UpsertOAuth2Config(ctx, cfg)
}

// clientSecret decrypts the stored client secret.
func (s *OAuth2Service) clientSecret(cfg *models.OAuth2Config) (string, error) {
	if len(cfg.ClientSecretEnc) == 0 {
		return "", nil
	}
	pt, err := DecryptBytesWithKey(s.encryptionKey, cfg.ClientSecretEnc)
	if err != nil {
		return "", fmt.Errorf("decrypting client secret: %w", err)
	}
	return string(pt), nil
}

// buildOAuth2Config constructs a golang.org/x/oauth2.Config from the stored settings.
func (s *OAuth2Service) buildOAuth2Config(ctx context.Context, cfg *models.OAuth2Config) (*oauth2.Config, *gooidc.Provider, error) {
	secret, err := s.clientSecret(cfg)
	if err != nil {
		return nil, nil, err
	}

	scopes := strings.Fields(strings.ReplaceAll(cfg.Scopes, ",", " "))
	if len(scopes) == 0 {
		scopes = []string{"openid", "email", "profile"}
	}

	var provider *gooidc.Provider
	authURL := cfg.AuthorizationURL
	tokenURL := cfg.TokenURL

	// If a discovery URL is configured, use OIDC auto-discovery.
	if cfg.DiscoveryURL != "" {
		p, err := gooidc.NewProvider(ctx, cfg.DiscoveryURL)
		if err != nil {
			return nil, nil, fmt.Errorf("OIDC discovery failed: %w", err)
		}
		provider = p
		authURL = p.Endpoint().AuthURL
		tokenURL = p.Endpoint().TokenURL
	}

	oc := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: secret,
		RedirectURL:  cfg.RedirectURI,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
	}
	return oc, provider, nil
}

// generateState creates a cryptographically random state string.
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GetAuthURL builds the OAuth2 authorization URL and persists the state token.
func (s *OAuth2Service) GetAuthURL(ctx context.Context, redirectURI string) (string, string, error) {
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return "", "", fmt.Errorf("loading OAuth2 config: %w", err)
	}
	if cfg == nil || !cfg.Enabled {
		return "", "", fmt.Errorf("OAuth2 authentication is not enabled")
	}

	oc, _, err := s.buildOAuth2Config(ctx, cfg)
	if err != nil {
		return "", "", err
	}

	state, err := generateState()
	if err != nil {
		return "", "", fmt.Errorf("generating state: %w", err)
	}

	if err := s.repository.SaveOAuth2State(ctx, state, redirectURI, time.Now().Add(10*time.Minute)); err != nil {
		return "", "", fmt.Errorf("saving state: %w", err)
	}

	url := oc.AuthCodeURL(state, oauth2.AccessTypeOnline)
	return url, state, nil
}

// Exchange exchanges an authorization code for a user, finding or creating a local account.
func (s *OAuth2Service) Exchange(ctx context.Context, code, state string) (*models.User, error) {
	// Validate and consume the state
	_, err := s.repository.ConsumeOAuth2State(ctx, state)
	if err != nil {
		return nil, fmt.Errorf("invalid OAuth2 state: %w", err)
	}

	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading OAuth2 config: %w", err)
	}
	if cfg == nil || !cfg.Enabled {
		return nil, fmt.Errorf("OAuth2 authentication is not enabled")
	}

	oc, oidcProvider, err := s.buildOAuth2Config(ctx, cfg)
	if err != nil {
		return nil, err
	}

	token, err := oc.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	var sub, email, preferredUsername string

	// Try OIDC ID token first
	if oidcProvider != nil {
		rawIDToken, ok := token.Extra("id_token").(string)
		if ok {
			verifier := oidcProvider.Verifier(&gooidc.Config{ClientID: cfg.ClientID})
			idToken, err := verifier.Verify(ctx, rawIDToken)
			if err == nil {
				var claims struct {
					Sub               string `json:"sub"`
					Email             string `json:"email"`
					EmailVerified     *bool  `json:"email_verified"`
					PreferredUsername string `json:"preferred_username"`
					Name              string `json:"name"`
				}
				if err := idToken.Claims(&claims); err == nil {
					sub = claims.Sub
					// Only trust the email if email_verified is absent (provider
					// guarantees verification) or explicitly true.
					if claims.EmailVerified == nil || *claims.EmailVerified {
						email = claims.Email
					}
					preferredUsername = claims.PreferredUsername
					if preferredUsername == "" {
						preferredUsername = claims.Name
					}
				}
			}
		}
	}

	// Fall back to userinfo endpoint if OIDC claims were not obtained
	if sub == "" && cfg.UserinfoURL != "" {
		httpClient := oc.Client(ctx, token)
		resp, err := httpClient.Get(cfg.UserinfoURL)
		if err == nil {
			defer resp.Body.Close()
			var info struct {
				Sub               string `json:"sub"`
				Email             string `json:"email"`
				EmailVerified     *bool  `json:"email_verified"`
				PreferredUsername string `json:"preferred_username"`
				Name              string `json:"name"`
			}
			if decErr := decodeJSON(resp.Body, &info); decErr == nil {
				sub = info.Sub
				// Only trust the email if email_verified is absent or explicitly true.
				if info.EmailVerified == nil || *info.EmailVerified {
					email = info.Email
				}
				preferredUsername = info.PreferredUsername
				if preferredUsername == "" {
					preferredUsername = info.Name
				}
			}
		}
	}

	if sub == "" {
		return nil, fmt.Errorf("could not determine user identity from OAuth2 provider")
	}
	// Require a verified email address.  Storing a fabricated "@oauth2.local"
	// address (or an unverified one) would create a false identity that could
	// be exploited for privilege escalation if email-based lookups are ever
	// introduced.  Fail loudly instead so the administrator can configure the
	// OAuth2 provider to include a verified email claim.
	if email == "" {
		return nil, fmt.Errorf("OAuth2 provider did not return a verified email address")
	}
	if preferredUsername == "" {
		preferredUsername = email
	}

	// Find or create local user
	user, err := s.repository.FindUserByExternalAuth(ctx, "oauth2", sub)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if user == nil {
		user, err = s.repository.CreateExternalUser(ctx, preferredUsername, email, "oauth2", sub)
		if err != nil {
			return nil, fmt.Errorf("creating local user: %w", err)
		}
	}
	return user, nil
}
