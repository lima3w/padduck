package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"padduck/version"
)

const telemetryInstallIDKey = "telemetry_install_id"

// featureConfigKeys lists all feature toggle config keys collected in
// feature_flags_json. Kept in sync with handlers/features.go constants.
var featureConfigKeys = []string{
	"feature_customers_enabled",
	"feature_vlans_enabled",
	"feature_vrfs_enabled",
	"feature_racks_enabled",
	"feature_locations_enabled",
	"feature_bgp_enabled",
	"feature_devices_enabled",
	"feature_nat_enabled",
	"feature_dhcp_enabled",
	"feature_circuits_enabled",
	"feature_firewall_enabled",
}

// TelemetrySnapshot is a privacy-safe aggregate payload that describes one
// install at a point in time. Field names match the padduck_analytics
// PocketBase collection schema exactly.
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

	// Locale (from admin config; empty strings omitted from payload)
	UILocale       string `json:"ui_locale,omitempty"`
	TimezoneRegion string `json:"timezone_region,omitempty"`
	CountryCode    string `json:"country_code,omitempty"`
	RegionCode     string `json:"region_code,omitempty"`

	// Object counts
	UsersTotal       int64 `json:"users_total"`
	ActiveUsers7d    int64 `json:"active_users_7d"`
	ActiveUsers30d   int64 `json:"active_users_30d"`
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

	// IPv4 subnet utilization metrics (nil when no IPv4 subnets exist)
	SubnetUtilizationAvgPct    *float64 `json:"subnet_utilization_avg_pct,omitempty"`
	SubnetUtilizationMedianPct *float64 `json:"subnet_utilization_median_pct,omitempty"`
	SubnetUtilizationP75Pct    *float64 `json:"subnet_utilization_p75_pct,omitempty"`
	SubnetUtilizationP90Pct    *float64 `json:"subnet_utilization_p90_pct,omitempty"`
	SubnetUtilizationP95Pct    *float64 `json:"subnet_utilization_p95_pct,omitempty"`
	SubnetsEmpty               int64    `json:"subnets_empty"`
	SubnetsOver50Pct           int64    `json:"subnets_over_50_pct"`
	SubnetsOver80Pct           int64    `json:"subnets_over_80_pct"`
	SubnetsOver90Pct           int64    `json:"subnets_over_90_pct"`
	SubnetsFull                int64    `json:"subnets_full"`

	// Feature flags
	SSOEnabled           bool `json:"sso_enabled"`
	LDAPEnabled          bool `json:"ldap_enabled"`
	OIDCEnabled          bool `json:"oidc_enabled"`
	SNMPDiscoveryEnabled bool `json:"snmp_discovery_enabled"`
	APIEnabled           bool `json:"api_enabled"`

	// JSON extension fields
	FeatureFlagsJSON map[string]bool `json:"feature_flags_json,omitempty"`
	ExtraMetricsJSON map[string]any  `json:"extra_metrics_json,omitempty"`
}

// TelemetryService assembles and sends TelemetrySnapshot values.
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
func (t *TelemetryService) CollectSnapshot(ctx context.Context, period string) (*TelemetrySnapshot, error) {
	installID, err := t.GetOrCreateInstallID(ctx)
	if err != nil {
		return nil, err
	}

	counts, err := t.svc.repository.GetTelemetryCounts(ctx)
	if err != nil {
		return nil, err
	}

	util, err := t.svc.repository.GetTelemetryUtilizationMetrics(ctx)
	if err != nil {
		return nil, err
	}

	ldapEnabled, oidcEnabled, samlEnabled := t.authFlags(ctx)
	snmpEnabled := t.snmpConfigured(ctx)
	apiEnabled := t.apiEnabled(ctx)

	return &TelemetrySnapshot{
		InstallID:              installID,
		SnapshotAt:             time.Now().UTC(),
		SnapshotPeriod:         period,
		TelemetrySchemaVersion: 1,
		AppVersion:             version.Version,

		Edition:        "community",
		DeploymentType: configStrDefault(t.configStr(ctx, "telemetry_deployment_type"), "unknown"),
		DeploymentMode: configStrDefault(t.configStr(ctx, "telemetry_deployment_mode"), "unknown"),
		ServerOSFamily: serverOSFamily(runtime.GOOS),
		DatabaseType:   "postgres",

		UILocale:       t.configStr(ctx, "telemetry_ui_locale"),
		TimezoneRegion: t.configStr(ctx, "telemetry_timezone_region"),
		CountryCode:    t.configStr(ctx, "telemetry_country_code"),
		RegionCode:     t.configStr(ctx, "telemetry_region_code"),

		UsersTotal:       counts.UsersTotal,
		ActiveUsers7d:    counts.ActiveUsers7d,
		ActiveUsers30d:   counts.ActiveUsers30d,
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

		SubnetUtilizationAvgPct:    roundPct(util.AvgPct),
		SubnetUtilizationMedianPct: roundPct(util.MedianPct),
		SubnetUtilizationP75Pct:    roundPct(util.P75Pct),
		SubnetUtilizationP90Pct:    roundPct(util.P90Pct),
		SubnetUtilizationP95Pct:    roundPct(util.P95Pct),
		SubnetsEmpty:               util.Empty,
		SubnetsOver50Pct:           util.Over50,
		SubnetsOver80Pct:           util.Over80,
		SubnetsOver90Pct:           util.Over90,
		SubnetsFull:                util.Full,

		SSOEnabled:           ldapEnabled || oidcEnabled || samlEnabled,
		LDAPEnabled:          ldapEnabled,
		OIDCEnabled:          oidcEnabled,
		SNMPDiscoveryEnabled: snmpEnabled,
		APIEnabled:           apiEnabled,

		FeatureFlagsJSON: t.featureFlagsJSON(ctx),
		ExtraMetricsJSON: map[string]any{
			"devices_total": counts.DevicesTotal,
		},
	}, nil
}

// SendNow collects and sends a snapshot immediately, bypassing the opt-in
// check. Used by the admin test endpoint.
func (t *TelemetryService) SendNow(ctx context.Context) error {
	return t.doSend(ctx, "manual")
}

// SendSnapshot sends a scheduled snapshot. Silently skips if the opt-in
// flag is not set or the connection is not configured.
func (t *TelemetryService) SendSnapshot(ctx context.Context, period string) error {
	if enabled, _ := t.svc.Config.GetCtx(ctx, "telemetry_enabled"); enabled != "true" {
		return nil
	}
	return t.doSend(ctx, period)
}

// doSend collects the snapshot and POSTs it to PocketBase.
func (t *TelemetryService) doSend(ctx context.Context, period string) error {
	pbURL, _ := t.svc.Config.GetCtx(ctx, "telemetry_pocketbase_url")
	pbToken, _ := t.svc.Config.GetCtx(ctx, "telemetry_pocketbase_token")
	if pbURL == "" || pbToken == "" {
		return fmt.Errorf("telemetry_pocketbase_url and telemetry_pocketbase_token must be configured")
	}

	snapshot, err := t.CollectSnapshot(ctx, period)
	if err != nil {
		return fmt.Errorf("collect snapshot: %w", err)
	}

	body, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}

	endpoint := strings.TrimRight(pbURL, "/") + "/api/collections/padduck_analytics/records"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+pbToken)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("pocketbase returned %d: %s", resp.StatusCode, b)
	}
	return nil
}

// StartTelemetryJob launches a background goroutine that sends scheduled
// snapshots. The period is read from config at startup (default: daily).
func (t *TelemetryService) StartTelemetryJob(ctx context.Context) {
	go func() {
		period := "daily"
		interval := 24 * time.Hour
		if cfg, _ := t.svc.Config.GetCtx(ctx, "telemetry_snapshot_period"); cfg == "weekly" {
			period = "weekly"
			interval = 7 * 24 * time.Hour
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := t.SendSnapshot(ctx, period); err != nil {
					slog.Warn("telemetry: snapshot send failed", "error", err)
				}
			}
		}
	}()
}

// authFlags returns whether LDAP, OIDC, and SAML are each enabled.
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

func (t *TelemetryService) snmpConfigured(ctx context.Context) bool {
	v, err := t.svc.Config.GetCtx(ctx, "scanner_snmp_community")
	return err == nil && v != ""
}

func (t *TelemetryService) apiEnabled(ctx context.Context) bool {
	v, err := t.svc.Config.GetCtx(ctx, "anonymous_api_enabled")
	return err == nil && v == "true"
}

func (t *TelemetryService) configStr(ctx context.Context, key string) string {
	v, _ := t.svc.Config.GetCtx(ctx, key)
	return v
}

func configStrDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func (t *TelemetryService) featureFlagsJSON(ctx context.Context) map[string]bool {
	flags := make(map[string]bool, len(featureConfigKeys))
	for _, k := range featureConfigKeys {
		v, err := t.svc.Config.GetCtx(ctx, k)
		flags[k] = err != nil || v != "false"
	}
	return flags
}

func roundPct(p *float64) *float64 {
	if p == nil {
		return nil
	}
	v := math.Round(*p*100) / 100
	return &v
}

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
