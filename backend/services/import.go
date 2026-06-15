package services

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"padduck/internal/export"
	"padduck/models"
	"padduck/utils"
)

// ─────────────────────────────────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────────────────────────────────

// ImportResult summarises the outcome of a bulk import.
type ImportResult struct {
	Total    int              `json:"total"`
	Imported int              `json:"imported"`
	Failed   int              `json:"failed"`
	Errors   []ImportRowError `json:"errors"`
}

// ImportRowError describes a single failed row.
type ImportRowError struct {
	Row     int    `json:"row"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// DryRunRowAction describes the action that would be taken for a single row.
type DryRunRowAction string

const (
	DryRunCreate  DryRunRowAction = "create"
	DryRunSkip    DryRunRowAction = "skip"
	DryRunWarning DryRunRowAction = "warning"
	DryRunError   DryRunRowAction = "error"
)

// DryRunRow is one row's preview outcome.
type DryRunRow struct {
	Row    int             `json:"row"`
	Action DryRunRowAction `json:"action"`
	Value  string          `json:"value"`
	Reason string          `json:"reason,omitempty"`
}

// DryRunResult summarises what would happen in a dry-run import.
type DryRunResult struct {
	DryRun   bool        `json:"dry_run"`
	Total    int         `json:"total"`
	Creates  int         `json:"creates"`
	Skips    int         `json:"skips"`
	Warnings int         `json:"warnings"`
	Errors   int         `json:"errors"`
	Rows     []DryRunRow `json:"rows"`
}

// ImportRepo lists the repository methods needed by ImportService.
// It is exported so that handler tests can build stub implementations.
type ImportRepo interface {
	// sections
	ListAllNetworks(ctx context.Context) ([]*models.Network, error)
	// subnets
	ListSubnetsBySection(ctx context.Context, networkID int64) ([]*models.Subnet, error)
	ListAllSubnets(ctx context.Context) ([]*models.Subnet, error)
	CreateSubnetWithVLAN(ctx context.Context, networkID int64, networkAddress string, prefixLength int, description string, gateway *string, autoFirst, autoLast bool, locationID *int64, nameserverID *int64, vlanID *int64) (*models.Subnet, error)
	// IPs
	ListIPAddressesBySubnet(ctx context.Context, subnetID int64) ([]*models.IPAddress, error)
	ListAllIPAddresses(ctx context.Context) ([]*models.IPAddress, error)
	CreateIPAddress(ctx context.Context, subnetID int64, address, hostname, status string, tagID *int64, macAddress, ptrRecord, dnsName *string) (*models.IPAddress, error)
	// VLANs / VRFs
	ListAllVLANs(ctx context.Context) ([]*models.VLAN, error)
	ListAllVRFs(ctx context.Context) ([]*models.VRF, error)
}

// ImportService handles CSV and third-party import operations.
type ImportService struct {
	repo ImportRepo
}

// NewImportService creates an ImportService.
func NewImportService(repo ImportRepo) *ImportService {
	return &ImportService{repo: repo}
}

type migrationBundleFile struct {
	Path        string `json:"path"`
	ContentType string `json:"content_type"`
	Bytes       int    `json:"bytes"`
	SHA256      string `json:"sha256"`
}

type migrationBundleManifest struct {
	BundleVersion string                `json:"bundle_version"`
	Source        string                `json:"source"`
	Target        string                `json:"target"`
	GeneratedAt   string                `json:"generated_at"`
	Files         []migrationBundleFile `json:"files"`
	Notes         []string              `json:"notes"`
}

// ─────────────────────────────────────────────────────────────────────────────
// ImportSubnetsCSV (#225)
// ─────────────────────────────────────────────────────────────────────────────

// ImportSubnetsCSV parses a CSV with headers:
//
//	cidr, description, section, gateway, vlan, vrf, location
//
// and creates a subnet for each valid row.
func (s *ImportService) ImportSubnetsCSV(ctx context.Context, r io.Reader) (*ImportResult, error) {
	records, err := readCSV(r)
	if err != nil {
		return nil, fmt.Errorf("parse CSV: %w", err)
	}

	result := &ImportResult{Errors: []ImportRowError{}}

	// Pre-load lookup tables once.
	sections, _ := s.repo.ListAllNetworks(ctx)
	sectionByName := indexSections(sections)

	vlans, _ := s.repo.ListAllVLANs(ctx)
	vlanByName := indexVLANs(vlans)

	for i, rec := range records {
		row := i + 2 // 1-based, header is row 1
		result.Total++

		cidr := strings.TrimSpace(rec["cidr"])
		if cidr == "" {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: row, Value: cidr, Message: "cidr is required"})
			continue
		}

		networkAddr, prefixLen, err := parseCIDR(cidr)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: row, Value: cidr, Message: "invalid cidr: " + err.Error()})
			continue
		}

		sectionName := strings.TrimSpace(rec["section"])
		networkID, ok := sectionByName[sectionName]
		if !ok {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: row, Value: cidr, Message: fmt.Sprintf("section %q not found", sectionName)})
			continue
		}

		description := strings.TrimSpace(rec["description"])

		var gateway *string
		if gw := strings.TrimSpace(rec["gateway"]); gw != "" {
			gateway = &gw
		}

		var vlanID *int64
		if vlanName := strings.TrimSpace(rec["vlan"]); vlanName != "" {
			if id, ok := vlanByName[vlanName]; ok {
				vlanID = &id
			}
		}

		_, err = s.repo.CreateSubnetWithVLAN(ctx, networkID, networkAddr, prefixLen, description, gateway, false, false, nil, nil, vlanID)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: row, Value: cidr, Message: err.Error()})
			continue
		}
		result.Imported++
	}

	return result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// ImportIPsCSV (#226)
// ─────────────────────────────────────────────────────────────────────────────

// validIPStatuses is the set of allowed status strings for IP addresses.
var validIPStatuses = map[string]bool{
	"available": true,
	"assigned":  true,
	"reserved":  true,
}

// ImportIPsCSV parses a CSV with headers:
//
//	address, hostname, status, subnet_cidr, assigned_to, mac_address
//
// and creates an IP address record for each valid row.
func (s *ImportService) ImportIPsCSV(ctx context.Context, r io.Reader) (*ImportResult, error) {
	records, err := readCSV(r)
	if err != nil {
		return nil, fmt.Errorf("parse CSV: %w", err)
	}

	result := &ImportResult{Errors: []ImportRowError{}}

	// Build subnet lookup: CIDR → id.
	subnetByCIDR, err := s.buildSubnetCIDRIndex(ctx)
	if err != nil {
		return nil, fmt.Errorf("load subnets: %w", err)
	}

	for i, rec := range records {
		row := i + 2
		result.Total++

		address := strings.TrimSpace(rec["address"])
		if address == "" {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: row, Value: address, Message: "address is required"})
			continue
		}

		subnetCIDR := strings.TrimSpace(rec["subnet_cidr"])
		subnetID, ok := subnetByCIDR[subnetCIDR]
		if !ok {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: row, Value: address, Message: fmt.Sprintf("subnet %q not found", subnetCIDR)})
			continue
		}

		status := strings.TrimSpace(rec["status"])
		if status == "" {
			status = "available"
		}
		if !validIPStatuses[status] {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: row, Value: address, Message: fmt.Sprintf("invalid status %q", status)})
			continue
		}

		hostname := strings.TrimSpace(rec["hostname"])

		var macAddress *string
		if mac := strings.TrimSpace(rec["mac_address"]); mac != "" {
			normalized, macErr := utils.NormalizeMAC(mac)
			if macErr != nil {
				result.Failed++
				result.Errors = append(result.Errors, ImportRowError{Row: row, Value: address, Message: macErr.Error()})
				continue
			}
			macAddress = &normalized
		}

		_, err = s.repo.CreateIPAddress(ctx, subnetID, address, hostname, status, nil, macAddress, nil, nil)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: row, Value: address, Message: err.Error()})
			continue
		}
		result.Imported++
	}

	return result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// DryRunSubnetsCSV / DryRunIPsCSV (#426)
// ─────────────────────────────────────────────────────────────────────────────

// DryRunSubnetsCSV validates a subnets CSV without writing to the database.
func (s *ImportService) DryRunSubnetsCSV(ctx context.Context, r io.Reader) (*DryRunResult, error) {
	records, err := readCSV(r)
	if err != nil {
		return nil, fmt.Errorf("parse CSV: %w", err)
	}

	result := &DryRunResult{DryRun: true, Rows: []DryRunRow{}}

	sections, _ := s.repo.ListAllNetworks(ctx)
	sectionByName := indexSections(sections)
	existingCIDRs, _ := s.buildSubnetCIDRIndex(ctx)

	for i, rec := range records {
		row := i + 2
		result.Total++

		cidr := strings.TrimSpace(rec["cidr"])
		if cidr == "" {
			result.Errors++
			result.Rows = append(result.Rows, DryRunRow{Row: row, Action: DryRunError, Value: cidr, Reason: "cidr is required"})
			continue
		}

		networkAddr, prefixLen, err := parseCIDR(cidr)
		if err != nil {
			result.Errors++
			result.Rows = append(result.Rows, DryRunRow{Row: row, Action: DryRunError, Value: cidr, Reason: err.Error()})
			continue
		}
		normalizedCIDR := fmt.Sprintf("%s/%d", networkAddr, prefixLen)

		// Check for duplicate.
		if _, exists := existingCIDRs[normalizedCIDR]; exists {
			result.Skips++
			result.Rows = append(result.Rows, DryRunRow{Row: row, Action: DryRunSkip, Value: cidr, Reason: "subnet already exists"})
			continue
		}

		// Check section exists if provided.
		sectionName := strings.TrimSpace(rec["section"])
		if sectionName != "" {
			if _, ok := sectionByName[sectionName]; !ok {
				result.Warnings++
				result.Rows = append(result.Rows, DryRunRow{Row: row, Action: DryRunWarning, Value: cidr, Reason: fmt.Sprintf("section %q not found; will be skipped on import", sectionName)})
				continue
			}
		}

		result.Creates++
		result.Rows = append(result.Rows, DryRunRow{Row: row, Action: DryRunCreate, Value: cidr})
	}

	return result, nil
}

// DryRunIPsCSV validates an IP addresses CSV without writing to the database.
func (s *ImportService) DryRunIPsCSV(ctx context.Context, r io.Reader) (*DryRunResult, error) {
	records, err := readCSV(r)
	if err != nil {
		return nil, fmt.Errorf("parse CSV: %w", err)
	}

	result := &DryRunResult{DryRun: true, Rows: []DryRunRow{}}

	subnetIndex, _ := s.buildSubnetCIDRIndex(ctx)

	for i, rec := range records {
		row := i + 2
		result.Total++

		address := strings.TrimSpace(rec["address"])
		if address == "" {
			result.Errors++
			result.Rows = append(result.Rows, DryRunRow{Row: row, Action: DryRunError, Value: address, Reason: "address is required"})
			continue
		}
		if net.ParseIP(address) == nil {
			result.Errors++
			result.Rows = append(result.Rows, DryRunRow{Row: row, Action: DryRunError, Value: address, Reason: "invalid IP address"})
			continue
		}

		subnetCIDR := strings.TrimSpace(rec["subnet_cidr"])
		if subnetCIDR == "" {
			result.Warnings++
			result.Rows = append(result.Rows, DryRunRow{Row: row, Action: DryRunWarning, Value: address, Reason: "subnet_cidr not provided"})
			continue
		}

		networkAddr, prefixLen, err := parseCIDR(subnetCIDR)
		if err != nil {
			result.Errors++
			result.Rows = append(result.Rows, DryRunRow{Row: row, Action: DryRunError, Value: address, Reason: fmt.Sprintf("invalid subnet_cidr: %s", err)})
			continue
		}
		normalizedCIDR := fmt.Sprintf("%s/%d", networkAddr, prefixLen)
		if _, ok := subnetIndex[normalizedCIDR]; !ok {
			result.Warnings++
			result.Rows = append(result.Rows, DryRunRow{Row: row, Action: DryRunWarning, Value: address, Reason: fmt.Sprintf("subnet %s not found", subnetCIDR)})
			continue
		}

		result.Creates++
		result.Rows = append(result.Rows, DryRunRow{Row: row, Action: DryRunCreate, Value: address})
	}

	return result, nil
}

// DryRunPHPIpamSubnetsCSV validates a PHPIpam subnets CSV without writing to the database.
func (s *ImportService) DryRunPHPIpamSubnetsCSV(ctx context.Context, r io.Reader) (*DryRunResult, error) {
	records, err := readCSV(r)
	if err != nil {
		return nil, fmt.Errorf("parse CSV: %w", err)
	}

	result := &DryRunResult{DryRun: true, Rows: []DryRunRow{}}

	sections, _ := s.repo.ListAllNetworks(ctx)
	sectionByName := indexSections(sections)
	existingCIDRs, _ := s.buildSubnetCIDRIndex(ctx)

	for i, rec := range records {
		row := i + 2
		result.Total++

		subnet := strings.TrimSpace(rec["subnet"])
		mask := strings.TrimSpace(rec["mask"])

		if subnet == "" || mask == "" {
			result.Errors++
			result.Rows = append(result.Rows, DryRunRow{Row: row, Action: DryRunError, Value: subnet, Reason: "subnet and mask are required"})
			continue
		}

		prefixLen, err := maskToPrefixLen(mask)
		if err != nil {
			result.Errors++
			result.Rows = append(result.Rows, DryRunRow{Row: row, Action: DryRunError, Value: subnet, Reason: "invalid mask: " + err.Error()})
			continue
		}
		normalizedCIDR := fmt.Sprintf("%s/%d", subnet, prefixLen)

		if _, exists := existingCIDRs[normalizedCIDR]; exists {
			result.Skips++
			result.Rows = append(result.Rows, DryRunRow{Row: row, Action: DryRunSkip, Value: subnet, Reason: "subnet already exists"})
			continue
		}

		sectionName := strings.TrimSpace(rec["sectionName"])
		if sectionName != "" {
			if _, ok := sectionByName[sectionName]; !ok {
				result.Warnings++
				result.Rows = append(result.Rows, DryRunRow{Row: row, Action: DryRunWarning, Value: subnet, Reason: fmt.Sprintf("section %q not found; will be skipped on import", sectionName)})
				continue
			}
		}

		result.Creates++
		result.Rows = append(result.Rows, DryRunRow{Row: row, Action: DryRunCreate, Value: subnet})
	}

	return result, nil
}

// DryRunPHPIpamIPsCSV validates a PHPIpam IPs CSV without writing to the database.
func (s *ImportService) DryRunPHPIpamIPsCSV(ctx context.Context, r io.Reader) (*DryRunResult, error) {
	records, err := readCSV(r)
	if err != nil {
		return nil, fmt.Errorf("parse CSV: %w", err)
	}

	result := &DryRunResult{DryRun: true, Rows: []DryRunRow{}}

	subnetIndex, _ := s.buildSubnetCIDRIndex(ctx)

	for i, rec := range records {
		row := i + 2
		result.Total++

		ip := strings.TrimSpace(rec["ip"])
		if ip == "" {
			result.Errors++
			result.Rows = append(result.Rows, DryRunRow{Row: row, Action: DryRunError, Value: ip, Reason: "ip is required"})
			continue
		}
		if net.ParseIP(ip) == nil {
			result.Errors++
			result.Rows = append(result.Rows, DryRunRow{Row: row, Action: DryRunError, Value: ip, Reason: "invalid IP address"})
			continue
		}

		subnetIP := strings.TrimSpace(rec["subnetIp"])
		subnetMask := strings.TrimSpace(rec["subnetMask"])
		if subnetIP == "" || subnetMask == "" {
			result.Warnings++
			result.Rows = append(result.Rows, DryRunRow{Row: row, Action: DryRunWarning, Value: ip, Reason: "subnetIp and subnetMask are required"})
			continue
		}

		prefixLen, err := maskToPrefixLen(subnetMask)
		if err != nil {
			result.Errors++
			result.Rows = append(result.Rows, DryRunRow{Row: row, Action: DryRunError, Value: ip, Reason: "invalid subnetMask: " + err.Error()})
			continue
		}
		cidr := fmt.Sprintf("%s/%d", subnetIP, prefixLen)
		if _, ok := subnetIndex[cidr]; !ok {
			result.Warnings++
			result.Rows = append(result.Rows, DryRunRow{Row: row, Action: DryRunWarning, Value: ip, Reason: fmt.Sprintf("subnet %s not found", cidr)})
			continue
		}

		result.Creates++
		result.Rows = append(result.Rows, DryRunRow{Row: row, Action: DryRunCreate, Value: ip})
	}

	return result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// ImportFromPHPIpam (#227)
// ─────────────────────────────────────────────────────────────────────────────

// ImportFromPHPIpam imports data exported from PHPIpam.
//
// kind = "subnets": CSV with columns subnet, mask, description, sectionName
// kind = "ips":     CSV with columns ip, hostname, description, subnetIp, subnetMask, state
func (s *ImportService) ImportFromPHPIpam(ctx context.Context, r io.Reader, kind string) (*ImportResult, error) {
	switch kind {
	case "subnets":
		return s.importPHPIpamSubnets(ctx, r)
	case "ips":
		return s.importPHPIpamIPs(ctx, r)
	default:
		return nil, fmt.Errorf("unknown kind %q: expected \"subnets\" or \"ips\"", kind)
	}
}

func (s *ImportService) importPHPIpamSubnets(ctx context.Context, r io.Reader) (*ImportResult, error) {
	records, err := readCSV(r)
	if err != nil {
		return nil, fmt.Errorf("parse CSV: %w", err)
	}

	result := &ImportResult{Errors: []ImportRowError{}}

	sections, _ := s.repo.ListAllNetworks(ctx)
	sectionByName := indexSections(sections)

	for i, rec := range records {
		row := i + 2
		result.Total++

		subnet := strings.TrimSpace(rec["subnet"])
		mask := strings.TrimSpace(rec["mask"])

		if subnet == "" || mask == "" {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: row, Value: subnet, Message: "subnet and mask are required"})
			continue
		}

		// Convert subnet+mask to CIDR prefix length.
		prefixLen, err := maskToPrefixLen(mask)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: row, Value: subnet, Message: "invalid mask: " + err.Error()})
			continue
		}

		sectionName := strings.TrimSpace(rec["sectionName"])
		networkID, ok := sectionByName[sectionName]
		if !ok {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: row, Value: subnet, Message: fmt.Sprintf("section %q not found", sectionName)})
			continue
		}

		description := strings.TrimSpace(rec["description"])

		_, err = s.repo.CreateSubnetWithVLAN(ctx, networkID, subnet, prefixLen, description, nil, false, false, nil, nil, nil)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: row, Value: subnet, Message: err.Error()})
			continue
		}
		result.Imported++
	}

	return result, nil
}

func (s *ImportService) importPHPIpamIPs(ctx context.Context, r io.Reader) (*ImportResult, error) {
	records, err := readCSV(r)
	if err != nil {
		return nil, fmt.Errorf("parse CSV: %w", err)
	}

	result := &ImportResult{Errors: []ImportRowError{}}

	subnetByCIDR, err := s.buildSubnetCIDRIndex(ctx)
	if err != nil {
		return nil, fmt.Errorf("load subnets: %w", err)
	}

	for i, rec := range records {
		row := i + 2
		result.Total++

		ip := strings.TrimSpace(rec["ip"])
		if ip == "" {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: row, Value: ip, Message: "ip is required"})
			continue
		}

		subnetIP := strings.TrimSpace(rec["subnetIp"])
		subnetMask := strings.TrimSpace(rec["subnetMask"])

		if subnetIP == "" || subnetMask == "" {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: row, Value: ip, Message: "subnetIp and subnetMask are required"})
			continue
		}

		prefixLen, err := maskToPrefixLen(subnetMask)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: row, Value: ip, Message: "invalid subnetMask: " + err.Error()})
			continue
		}

		cidr := fmt.Sprintf("%s/%d", subnetIP, prefixLen)
		subnetID, ok := subnetByCIDR[cidr]
		if !ok {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: row, Value: ip, Message: fmt.Sprintf("subnet %q not found", cidr)})
			continue
		}

		hostname := strings.TrimSpace(rec["hostname"])
		state := strings.TrimSpace(rec["state"])
		status := phpIpamStateToStatus(state)

		_, err = s.repo.CreateIPAddress(ctx, subnetID, ip, hostname, status, nil, nil, nil, nil)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: row, Value: ip, Message: err.Error()})
			continue
		}
		result.Imported++
	}

	return result, nil
}

// phpIpamStateToStatus maps PHPIpam state values to our IP status strings.
// PHPIpam states: 1=used, 2=reserved, 3=dhcp, 0/4=offline/not-used.
// Our constraint: status IN ('available', 'assigned', 'reserved').
func phpIpamStateToStatus(state string) string {
	switch strings.ToLower(state) {
	case "1", "used", "active", "3", "dhcp":
		return "assigned"
	case "2", "reserved":
		return "reserved"
	default:
		return "available"
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ExportFullData (#228)
// ─────────────────────────────────────────────────────────────────────────────

// ExportFullData exports all subnets and their IPs in the requested format.
// Returns (data, filename, contentType, error).
func (s *ImportService) ExportFullData(ctx context.Context, format string) ([]byte, string, string, error) {
	subnets, err := s.repo.ListAllSubnets(ctx)
	if err != nil {
		return nil, "", "", fmt.Errorf("list subnets: %w", err)
	}

	switch format {
	case "json":
		return s.exportFullJSON(ctx, subnets)
	default:
		return s.exportFullCSV(ctx, subnets)
	}
}

// ExportV2MigrationBundle returns a zip archive containing v1 export data and
// a manifest that v2 migration tooling can validate before import.
func (s *ImportService) ExportV2MigrationBundle(ctx context.Context) ([]byte, string, string, error) {
	subnets, err := s.repo.ListAllSubnets(ctx)
	if err != nil {
		return nil, "", "", fmt.Errorf("list subnets: %w", err)
	}

	csvData, _, _, err := s.exportFullCSV(ctx, subnets)
	if err != nil {
		return nil, "", "", err
	}
	jsonData, _, _, err := s.exportFullJSON(ctx, subnets)
	if err != nil {
		return nil, "", "", err
	}

	readme := []byte(strings.TrimSpace(`# Padduck v2 Migration Bundle

This archive was generated by the v1 pre-v2 compatibility tooling.

Files:
- manifest.json: bundle metadata, target version, checksums, and guidance.
- data/padduck-v1-export.json: structured v1 export for migration tooling.
- data/padduck-v1-export.csv: human-readable fallback export.

Validate the manifest checksums before importing this bundle into v2.
`) + "\n")

	now := time.Now().UTC()
	files := []migrationBundleFile{
		bundleFile("data/padduck-v1-export.json", "application/json", jsonData),
		bundleFile("data/padduck-v1-export.csv", "text/csv", csvData),
		bundleFile("README.md", "text/markdown", readme),
	}
	manifest := migrationBundleManifest{
		BundleVersion: "v1-pre-v2",
		Source:        "padduck-v1",
		Target:        "padduck-v2",
		GeneratedAt:   now.Format(time.RFC3339),
		Files:         files,
		Notes: []string{
			"Use manifest checksums to detect partial or modified bundle contents.",
			"JSON is the canonical migration input; CSV is included for inspection and fallback workflows.",
			"Review /api/v1/admin/compatibility/v2-warnings before importing into v2.",
		},
	}
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, "", "", fmt.Errorf("marshal manifest: %w", err)
	}
	manifestData = append(manifestData, '\n')

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	if err := addZipFile(zw, "manifest.json", manifestData); err != nil {
		return nil, "", "", err
	}
	if err := addZipFile(zw, "data/padduck-v1-export.json", jsonData); err != nil {
		return nil, "", "", err
	}
	if err := addZipFile(zw, "data/padduck-v1-export.csv", csvData); err != nil {
		return nil, "", "", err
	}
	if err := addZipFile(zw, "README.md", readme); err != nil {
		return nil, "", "", err
	}
	if err := zw.Close(); err != nil {
		return nil, "", "", fmt.Errorf("close migration bundle: %w", err)
	}

	filename := fmt.Sprintf("padduck-v2-migration-bundle-%s.zip", now.Format("20060102-150405"))
	return buf.Bytes(), filename, "application/zip", nil
}

func bundleFile(path, contentType string, data []byte) migrationBundleFile {
	sum := sha256.Sum256(data)
	return migrationBundleFile{
		Path:        path,
		ContentType: contentType,
		Bytes:       len(data),
		SHA256:      fmt.Sprintf("%x", sum[:]),
	}
}

func addZipFile(zw *zip.Writer, name string, data []byte) error {
	w, err := zw.Create(name)
	if err != nil {
		return fmt.Errorf("create bundle file %s: %w", name, err)
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write bundle file %s: %w", name, err)
	}
	return nil
}

// exportFullCSV produces two CSV sections separated by a blank line.
func (s *ImportService) exportFullCSV(ctx context.Context, subnets []*models.Subnet) ([]byte, string, string, error) {
	subnetHeaders := []string{"cidr", "description", "network_id", "gateway", "vlan_id"}
	subnetRows := make([]map[string]string, 0, len(subnets))
	for _, sub := range subnets {
		row := map[string]string{
			"cidr":        fmt.Sprintf("%s/%d", sub.NetworkAddress, sub.PrefixLength),
			"description": sub.Description,
			"network_id":  strconv.FormatInt(sub.NetworkID, 10),
			"gateway":     strPtrVal(sub.Gateway),
			"vlan_id":     int64PtrVal(sub.VLANID),
		}
		subnetRows = append(subnetRows, row)
	}

	subnetCSV, err := export.GenerateCSV(subnetHeaders, subnetRows)
	if err != nil {
		return nil, "", "", fmt.Errorf("generate subnet CSV: %w", err)
	}

	allIPs, err := s.repo.ListAllIPAddresses(ctx)
	if err != nil {
		return nil, "", "", fmt.Errorf("fetch IP addresses: %w", err)
	}
	ipsBySubnet := make(map[int64][]*models.IPAddress, len(subnets))
	for _, ip := range allIPs {
		ipsBySubnet[ip.SubnetID] = append(ipsBySubnet[ip.SubnetID], ip)
	}

	ipHeaders := []string{"address", "hostname", "status", "subnet_cidr", "mac_address"}
	ipRows := make([]map[string]string, 0)
	for _, sub := range subnets {
		cidr := fmt.Sprintf("%s/%d", sub.NetworkAddress, sub.PrefixLength)
		for _, ip := range ipsBySubnet[sub.ID] {
			ipRows = append(ipRows, map[string]string{
				"address":     ip.Address,
				"hostname":    ip.Hostname,
				"status":      ip.Status,
				"subnet_cidr": cidr,
				"mac_address": strPtrVal(ip.MACAddress),
			})
		}
	}

	ipCSV, err := export.GenerateCSV(ipHeaders, ipRows)
	if err != nil {
		return nil, "", "", fmt.Errorf("generate IP CSV: %w", err)
	}

	// Combine: subnets section, blank line, IPs section.
	data := make([]byte, 0, len(subnetCSV)+len(ipCSV)+2)
	data = append(data, subnetCSV...)
	data = append(data, '\n')
	data = append(data, ipCSV...)

	ts := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("ipam-export-%s.csv", ts)
	return data, filename, "text/csv", nil
}

// exportFullJSON produces a JSON payload with "subnets" and "ip_addresses" arrays.
func (s *ImportService) exportFullJSON(ctx context.Context, subnets []*models.Subnet) ([]byte, string, string, error) {
	type ipRow struct {
		Address    string  `json:"address"`
		Hostname   string  `json:"hostname"`
		Status     string  `json:"status"`
		SubnetCIDR string  `json:"subnet_cidr"`
		MACAddress *string `json:"mac_address,omitempty"`
	}

	type payload struct {
		Subnets     []*models.Subnet `json:"subnets"`
		IPAddresses []ipRow          `json:"ip_addresses"`
	}

	allIPs, err := s.repo.ListAllIPAddresses(ctx)
	if err != nil {
		return nil, "", "", fmt.Errorf("fetch IP addresses: %w", err)
	}
	ipsBySubnet := make(map[int64][]*models.IPAddress, len(subnets))
	for _, ip := range allIPs {
		ipsBySubnet[ip.SubnetID] = append(ipsBySubnet[ip.SubnetID], ip)
	}

	ips := make([]ipRow, 0, len(allIPs))
	for _, sub := range subnets {
		cidr := fmt.Sprintf("%s/%d", sub.NetworkAddress, sub.PrefixLength)
		for _, ip := range ipsBySubnet[sub.ID] {
			ips = append(ips, ipRow{
				Address:    ip.Address,
				Hostname:   ip.Hostname,
				Status:     ip.Status,
				SubnetCIDR: cidr,
				MACAddress: ip.MACAddress,
			})
		}
	}

	data, err := json.Marshal(payload{Subnets: subnets, IPAddresses: ips})
	if err != nil {
		return nil, "", "", fmt.Errorf("marshal JSON: %w", err)
	}

	ts := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("ipam-export-%s.json", ts)
	return data, filename, "application/json", nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// readCSV reads all rows of a CSV reader and returns them as header-keyed maps.
// The first row is treated as the header.
func readCSV(r io.Reader) ([]map[string]string, error) {
	cr := csv.NewReader(r)
	cr.TrimLeadingSpace = true

	headers, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	for i, h := range headers {
		headers[i] = strings.TrimSpace(h)
	}

	var records []map[string]string
	for {
		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}
		rec := make(map[string]string, len(headers))
		for i, h := range headers {
			if i < len(row) {
				rec[h] = row[i]
			}
		}
		records = append(records, rec)
	}
	return records, nil
}

// parseCIDR splits a CIDR string into network address and prefix length.
func parseCIDR(cidr string) (string, int, error) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", 0, err
	}
	ones, _ := ipNet.Mask.Size()
	// Use the provided IP rather than the network address so callers can
	// pass either 10.0.0.0/24 or 10.0.0.1/24.
	_ = ip
	return ipNet.IP.String(), ones, nil
}

// maskToPrefixLen converts a dotted-decimal mask to a prefix length.
// It also accepts a plain integer string (e.g. "24").
func maskToPrefixLen(mask string) (int, error) {
	// Try as integer first.
	if n, err := strconv.Atoi(mask); err == nil {
		if n < 0 || n > 128 {
			return 0, fmt.Errorf("prefix length %d out of range", n)
		}
		return n, nil
	}
	// Try dotted-decimal.
	ip := net.ParseIP(mask)
	if ip == nil {
		return 0, fmt.Errorf("cannot parse mask %q", mask)
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return 0, fmt.Errorf("non-IPv4 mask %q", mask)
	}
	ones, _ := net.IPMask(ip4).Size()
	if ones == 0 && ip4[0] != 0 {
		return 0, fmt.Errorf("invalid mask %q", mask)
	}
	return ones, nil
}

// buildSubnetCIDRIndex returns a map from CIDR string to subnet ID.
func (s *ImportService) buildSubnetCIDRIndex(ctx context.Context) (map[string]int64, error) {
	subnets, err := s.repo.ListAllSubnets(ctx)
	if err != nil {
		return nil, err
	}
	idx := make(map[string]int64, len(subnets))
	for _, sub := range subnets {
		cidr := fmt.Sprintf("%s/%d", sub.NetworkAddress, sub.PrefixLength)
		idx[cidr] = sub.ID
	}
	return idx, nil
}

// indexSections returns a name→ID map.
func indexSections(sections []*models.Network) map[string]int64 {
	idx := make(map[string]int64, len(sections))
	for _, s := range sections {
		idx[s.Name] = s.ID
	}
	return idx
}

// indexVLANs returns a name→ID map.
func indexVLANs(vlans []*models.VLAN) map[string]int64 {
	idx := make(map[string]int64, len(vlans))
	for _, v := range vlans {
		idx[v.Name] = v.ID
	}
	return idx
}

// strPtrVal dereferences a string pointer, returning "" for nil.
func strPtrVal(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// int64PtrVal formats a *int64 as a string, returning "" for nil.
func int64PtrVal(p *int64) string {
	if p == nil {
		return ""
	}
	return strconv.FormatInt(*p, 10)
}
