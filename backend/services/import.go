package services

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"ipam-next/internal/export"
	"ipam-next/models"
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

// ImportRepo lists the repository methods needed by ImportService.
// It is exported so that handler tests can build stub implementations.
type ImportRepo interface {
	// sections
	ListAllSections(ctx context.Context) ([]*models.Section, error)
	// subnets
	ListSubnetsBySection(ctx context.Context, sectionID int64) ([]*models.Subnet, error)
	ListAllSubnets(ctx context.Context) ([]*models.Subnet, error)
	CreateSubnetWithVLAN(ctx context.Context, sectionID int64, networkAddress string, prefixLength int, description string, gateway *string, autoFirst, autoLast bool, locationID *int64, nameserverID *int64, vlanID *int64) (*models.Subnet, error)
	// IPs
	ListIPAddressesBySubnet(ctx context.Context, subnetID int64) ([]*models.IPAddress, error)
	CreateIPAddress(ctx context.Context, subnetID int64, address, hostname, status string, assignedTo *string, tagID *int64, macAddress, ptrRecord *string) (*models.IPAddress, error)
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
	sections, _ := s.repo.ListAllSections(ctx)
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
		sectionID, ok := sectionByName[sectionName]
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

		_, err = s.repo.CreateSubnetWithVLAN(ctx, sectionID, networkAddr, prefixLen, description, gateway, false, false, nil, nil, vlanID)
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
	"active":   true,
	"reserved": true,
	"dhcp":     true,
	"inactive": true,
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
			status = "active"
		}
		if !validIPStatuses[status] {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: row, Value: address, Message: fmt.Sprintf("invalid status %q", status)})
			continue
		}

		hostname := strings.TrimSpace(rec["hostname"])

		var assignedTo *string
		if at := strings.TrimSpace(rec["assigned_to"]); at != "" {
			assignedTo = &at
		}

		var macAddress *string
		if mac := strings.TrimSpace(rec["mac_address"]); mac != "" {
			macAddress = &mac
		}

		_, err = s.repo.CreateIPAddress(ctx, subnetID, address, hostname, status, assignedTo, nil, macAddress, nil)
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

	sections, _ := s.repo.ListAllSections(ctx)
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
		sectionID, ok := sectionByName[sectionName]
		if !ok {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: row, Value: subnet, Message: fmt.Sprintf("section %q not found", sectionName)})
			continue
		}

		description := strings.TrimSpace(rec["description"])

		_, err = s.repo.CreateSubnetWithVLAN(ctx, sectionID, subnet, prefixLen, description, nil, false, false, nil, nil, nil)
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
func phpIpamStateToStatus(state string) string {
	switch strings.ToLower(state) {
	case "1", "used", "active":
		return "active"
	case "2", "reserved":
		return "reserved"
	case "3", "dhcp":
		return "dhcp"
	default:
		return "inactive"
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

// exportFullCSV produces two CSV sections separated by a blank line.
func (s *ImportService) exportFullCSV(ctx context.Context, subnets []*models.Subnet) ([]byte, string, string, error) {
	subnetHeaders := []string{"cidr", "description", "section_id", "gateway", "vlan_id"}
	subnetRows := make([]map[string]string, 0, len(subnets))
	for _, sub := range subnets {
		row := map[string]string{
			"cidr":        fmt.Sprintf("%s/%d", sub.NetworkAddress, sub.PrefixLength),
			"description": sub.Description,
			"section_id":  strconv.FormatInt(sub.SectionID, 10),
			"gateway":     strPtrVal(sub.Gateway),
			"vlan_id":     int64PtrVal(sub.VLANID),
		}
		subnetRows = append(subnetRows, row)
	}

	subnetCSV, err := export.GenerateCSV(subnetHeaders, subnetRows)
	if err != nil {
		return nil, "", "", fmt.Errorf("generate subnet CSV: %w", err)
	}

	ipHeaders := []string{"address", "hostname", "status", "subnet_cidr", "assigned_to", "mac_address"}
	ipRows := make([]map[string]string, 0)
	for _, sub := range subnets {
		ips, err := s.repo.ListIPAddressesBySubnet(ctx, sub.ID)
		if err != nil {
			continue
		}
		cidr := fmt.Sprintf("%s/%d", sub.NetworkAddress, sub.PrefixLength)
		for _, ip := range ips {
			ipRows = append(ipRows, map[string]string{
				"address":     ip.Address,
				"hostname":    ip.Hostname,
				"status":      ip.Status,
				"subnet_cidr": cidr,
				"assigned_to": strPtrVal(ip.AssignedTo),
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
		AssignedTo *string `json:"assigned_to,omitempty"`
		MACAddress *string `json:"mac_address,omitempty"`
	}

	type payload struct {
		Subnets     []*models.Subnet `json:"subnets"`
		IPAddresses []ipRow          `json:"ip_addresses"`
	}

	ips := make([]ipRow, 0)
	for _, sub := range subnets {
		subIPs, err := s.repo.ListIPAddressesBySubnet(ctx, sub.ID)
		if err != nil {
			continue
		}
		cidr := fmt.Sprintf("%s/%d", sub.NetworkAddress, sub.PrefixLength)
		for _, ip := range subIPs {
			ips = append(ips, ipRow{
				Address:    ip.Address,
				Hostname:   ip.Hostname,
				Status:     ip.Status,
				SubnetCIDR: cidr,
				AssignedTo: ip.AssignedTo,
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
func indexSections(sections []*models.Section) map[string]int64 {
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
