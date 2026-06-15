package services

import (
	"context"
	"errors"
	"runtime"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"padduck/version"
)

const telemetryInstallIDKey = "telemetry_install_id"

// TelemetrySnapshot is a privacy-safe aggregate payload that describes one
// install at a point in time. Field names match the padduck_analytics
// PocketBase collection schema exactly.
//
// Fields that require additional infrastructure (active user counts,
// utilization percentiles, locale config, opt-in sender) are omitted from
// this first increment and will be added before the sender is wired up.
type TelemetrySnapshot struct {
	InstallID              string    `json:"install_id"`
	SnapshotAt             time.Time `json:"snapshot_at"`
	SnapshotPeriod         string    `json:"snapshot_period"`
	TelemetrySchemaVersion int       `json:"telemetry_schema_version"`
	AppVersion             string    `json:"app_version"`

	// System metadata
	Edition        string `json:"edition"`
	DeploymentType string `json:"deployment_type"`
	DeploymentMode string `json:"deployment_mode"`
	ServerOSFamily string `json:"server_os_family"`
	DatabaseType   string `json:"database_type"`

	// Object counts
	UsersTotal       int64 `json:"users_total"`
	CustomersTotal   int64 `json:"customers_total"`
	LocationsTotal   int64 `json:"locations_total"`
	VLANsTotal       int64 `json:"vlans_total"`
	SubnetsTotal     int64 `json:"subnets_total"`
	IPv4SubnetsTotal int64 `json:"ipv4_subnets_total"`
	IPv6SubnetsTotal int64 `json:"ipv6_subnets_total"`

	// IPv4 subnet size buckets
	IPv4Subnets29to32 int64 `json:"ipv4_subnets_29_to_32"`
	IPv4Subnets25to28 int64 `json:"ipv4_subnets_25_to_28"`
	IPv4Subnets24     int64 `json:"ipv4_subnets_24"`
	IPv4Subnets16to23 int64 `json:"ipv4_subnets_16_to_23"`
	IPv4Subnets8to15  int64 `json:"ipv4_subnets_8_to_15"`

	// Feature flags
	SSOEnabled           bool `json:"sso_enabled"`
	LDAPEnabled          bool `json:"ldap_enabled"`
	OIDCEnabled          bool `json:"oidc_enabled"`
	SNMPDiscoveryEnabled bool `json:"snmp_discovery_enabled"`
	APIEnabled           bool `json:"api_enabled"`
}

// TelemetryService assembles TelemetrySnapshot values from live app state.
// It does not transmit anything — the sender will be added in a later increment.
type TelemetryService struct {
	svc *Service
}

func newTelemetryService(svc *Service) *TelemetryService {
	return &TelemetryService{svc: svc}
}

// GetOrCreateInstallID returns the stable per-install UUID, generating and
// persisting it on first call. The ID is stored in the configs table and
// never derived from any identifying host or network property.
func (t *TelemetryService) GetOrCreateInstallID(ctx context.Context) (string, error) {
	val, err := t.svc.Config.GetCtx(ctx, telemetryInstallIDKey)
	if err == nil && val != "" {
		return val, nil
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}
	id := uuid.New().String()
	if err := t.svc.Config.SetCtx(ctx, telemetryInstallIDKey, id); err != nil {
		return "", err
	}
	return id, nil
}

// CollectSnapshot gathers all available telemetry fields and returns a
// populated TelemetrySnapshot. No data is sent anywhere by this method.
func (t *TelemetryService) CollectSnapshot(ctx context.Context) (*TelemetrySnapshot, error) {
	installID, err := t.GetOrCreateInstallID(ctx)
	if err != nil {
		return nil, err
	}

	counts, err := t.svc.repository.GetTelemetryCounts(ctx)
	if err != nil {
		return nil, err
	}

	ldapEnabled, oidcEnabled, samlEnabled := t.authFlags(ctx)
	snmpEnabled := t.snmpConfigured(ctx)
	apiEnabled := t.apiEnabled(ctx)

	return &TelemetrySnapshot{
		InstallID:              installID,
		SnapshotAt:             time.Now().UTC(),
		SnapshotPeriod:         "manual",
		TelemetrySchemaVersion: 1,
		AppVersion:             version.Version,

		Edition:        "unknown",
		DeploymentType: "unknown",
		DeploymentMode: "unknown",
		ServerOSFamily: serverOSFamily(runtime.GOOS),
		DatabaseType:   "postgres",

		UsersTotal:       counts.UsersTotal,
		CustomersTotal:   counts.CustomersTotal,
		LocationsTotal:   counts.LocationsTotal,
		VLANsTotal:       counts.VLANsTotal,
		SubnetsTotal:     counts.SubnetsTotal,
		IPv4SubnetsTotal: counts.IPv4SubnetsTotal,
		IPv6SubnetsTotal: counts.IPv6SubnetsTotal,

		IPv4Subnets29to32: counts.IPv4Subnets29to32,
		IPv4Subnets25to28: counts.IPv4Subnets25to28,
		IPv4Subnets24:     counts.IPv4Subnets24,
		IPv4Subnets16to23: counts.IPv4Subnets16to23,
		IPv4Subnets8to15:  counts.IPv4Subnets8to15,

		SSOEnabled:           ldapEnabled || oidcEnabled || samlEnabled,
		LDAPEnabled:          ldapEnabled,
		OIDCEnabled:          oidcEnabled,
		SNMPDiscoveryEnabled: snmpEnabled,
		APIEnabled:           apiEnabled,
	}, nil
}

// authFlags returns whether LDAP, OIDC, and SAML are each enabled.
// A missing or unconfigured row is treated as disabled, not an error.
func (t *TelemetryService) authFlags(ctx context.Context) (ldap, oidc, saml bool) {
	if cfg, err := t.svc.LDAP.GetConfig(ctx); err == nil && cfg != nil {
		ldap = cfg.Enabled
	}
	if cfg, err := t.svc.OAuth2.GetConfig(ctx); err == nil && cfg != nil {
		oidc = cfg.Enabled
	}
	if cfg, err := t.svc.SAML.GetConfig(ctx); err == nil && cfg != nil {
		saml = cfg.Enabled
	}
	return
}

// snmpConfigured returns true if a global SNMP community string is set,
// indicating that SNMP discovery is configured.
func (t *TelemetryService) snmpConfigured(ctx context.Context) bool {
	v, err := t.svc.Config.GetCtx(ctx, "scanner_snmp_community")
	return err == nil && v != ""
}

// apiEnabled returns true if anonymous (unauthenticated) API access is enabled.
func (t *TelemetryService) apiEnabled(ctx context.Context) bool {
	v, err := t.svc.Config.GetCtx(ctx, "anonymous_api_enabled")
	return err == nil && v == "true"
}

// serverOSFamily maps a runtime.GOOS value to the allowed telemetry values.
func serverOSFamily(goos string) string {
	switch goos {
	case "linux":
		return "linux"
	case "windows":
		return "windows"
	case "darwin":
		return "macos"
	case "freebsd", "openbsd", "netbsd", "dragonfly":
		return "bsd"
	default:
		return "unknown"
	}
}
