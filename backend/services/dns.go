package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"padduck/internal/pdns"
	"padduck/internal/technitium"
	"padduck/models"
)

// ZoneInfo is a provider-agnostic zone summary used by the DNS zones UI.
type ZoneInfo struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}

// RecordInfo is a provider-agnostic DNS record used by the DNS zones UI.
type RecordInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	TTL     int    `json:"ttl"`
	Content string `json:"content"`
}

// DNSService handles DNS lookups, record tracking, and PowerDNS sync.
type DNSService struct {
	svc *Service
}

// NewDNSService creates a new DNSService.
func NewDNSService(svc *Service) *DNSService {
	return &DNSService{svc: svc}
}

// pdnsClient builds a PowerDNS client from config, or returns nil if not configured.
func (d *DNSService) pdnsClient(ctx context.Context) *pdns.Client {
	enabled, _ := d.svc.Config.GetCtx(ctx, "pdns_enabled")
	if enabled != "true" {
		return nil
	}
	apiURL, _ := d.svc.Config.GetCtx(ctx, "pdns_api_url")
	apiKey, _ := d.svc.Config.GetCtx(ctx, "pdns_api_key")
	if apiURL == "" || apiKey == "" {
		return nil
	}
	return pdns.NewClient(apiURL, apiKey)
}

// CheckDNS performs a forward DNS lookup for an IP that has a dns_name set, stores
// the resolved addresses and PTR name in the database, and updates dns_last_checked.
func (d *DNSService) CheckDNS(ctx context.Context, ipID int64) error {
	ip, err := d.svc.repository.GetIPAddressByID(ctx, ipID)
	if err != nil {
		return fmt.Errorf("get ip %d: %w", ipID, err)
	}
	if ip.DNSName == nil || *ip.DNSName == "" {
		return nil
	}

	addrs, lookupErr := net.LookupHost(*ip.DNSName)
	if lookupErr != nil {
		addrs = []string{}
	}

	// Build dns_records JSON
	records, _ := json.Marshal(map[string]interface{}{
		"name":      *ip.DNSName,
		"addresses": addrs,
		"error":     errString(lookupErr),
	})

	// Best-effort PTR lookup
	ptrRecord := ""
	ptrs, ptrErr := net.LookupAddr(ip.Address)
	if ptrErr == nil && len(ptrs) > 0 {
		ptrRecord = strings.TrimSuffix(ptrs[0], ".")
	}

	return d.svc.repository.UpdateIPDNSFields(ctx, ipID, ptrRecord, json.RawMessage(records), time.Now().UTC())
}

// CheckAllDNS iterates every IP address with a dns_name set and runs CheckDNS on each.
func (d *DNSService) CheckAllDNS(ctx context.Context) {
	ips, err := d.svc.repository.ListIPAddressesWithDNSName(ctx)
	if err != nil {
		log.Printf("[dns] ListIPAddressesWithDNSName: %v", err)
		return
	}
	for _, ip := range ips {
		if err := d.CheckDNS(ctx, ip.ID); err != nil {
			log.Printf("[dns] CheckDNS ip=%d: %v", ip.ID, err)
		}
	}
}

// SyncIPToPDNS creates or replaces A/AAAA and PTR records for an IP in PowerDNS.
// DNS failures are logged but do not block the caller.
func (d *DNSService) SyncIPToPDNS(ctx context.Context, ip *models.IPAddress) {
	client := d.pdnsClient(ctx)
	if client == nil {
		return
	}
	if ip.DNSName == nil || *ip.DNSName == "" {
		return
	}

	defaultZone, _ := d.svc.Config.GetCtx(ctx, "pdns_default_zone")
	if defaultZone == "" {
		log.Printf("[pdns] SyncIPToPDNS: pdns_default_zone not configured")
		return
	}

	// Determine record type based on address family
	rtype := "A"
	if strings.Contains(ip.Address, ":") {
		rtype = "AAAA"
	}

	fqdn := *ip.DNSName
	if !strings.HasSuffix(fqdn, ".") {
		fqdn += "."
	}

	if err := client.CreateRecord(ctx, defaultZone, fqdn, rtype, ip.Address, 300); err != nil {
		log.Printf("[pdns] SyncIPToPDNS CreateRecord A/AAAA ip=%s dns=%s: %v", ip.Address, *ip.DNSName, err)
	}

	// PTR record
	prefixLen := 0
	if strings.Contains(ip.Address, ":") && ip.SubnetID != 0 {
		if subnet, err := d.svc.repository.GetSubnetByID(ctx, ip.SubnetID); err == nil {
			prefixLen = subnet.PrefixLength
		} else {
			log.Printf("[pdns] SyncIPToPDNS: subnet lookup for PTR ip=%s: %v", ip.Address, err)
		}
	}
	ptrZone, ptrName := buildPTR(ip.Address, prefixLen)
	if ptrZone != "" {
		ptrZones, _ := d.svc.Config.GetCtx(ctx, "pdns_ptr_zones")
		if containsZone(ptrZones, ptrZone) {
			if err := client.CreateRecord(ctx, ptrZone, ptrName, "PTR", fqdn, 300); err != nil {
				log.Printf("[pdns] SyncIPToPDNS CreateRecord PTR ip=%s: %v", ip.Address, err)
			}
		}
	}
}

// RemoveIPFromPDNS deletes A/AAAA and PTR records for an IP from PowerDNS.
// DNS failures are logged but do not block the caller.
func (d *DNSService) RemoveIPFromPDNS(ctx context.Context, ip *models.IPAddress) {
	client := d.pdnsClient(ctx)
	if client == nil {
		return
	}
	if ip.DNSName == nil || *ip.DNSName == "" {
		return
	}

	defaultZone, _ := d.svc.Config.GetCtx(ctx, "pdns_default_zone")
	if defaultZone == "" {
		return
	}

	rtype := "A"
	if strings.Contains(ip.Address, ":") {
		rtype = "AAAA"
	}

	fqdn := *ip.DNSName
	if !strings.HasSuffix(fqdn, ".") {
		fqdn += "."
	}

	if err := client.DeleteRecord(ctx, defaultZone, fqdn, rtype); err != nil {
		log.Printf("[pdns] RemoveIPFromPDNS DeleteRecord A/AAAA ip=%s dns=%s: %v", ip.Address, *ip.DNSName, err)
	}

	prefixLen := 0
	if strings.Contains(ip.Address, ":") && ip.SubnetID != 0 {
		if subnet, err := d.svc.repository.GetSubnetByID(ctx, ip.SubnetID); err == nil {
			prefixLen = subnet.PrefixLength
		} else {
			log.Printf("[pdns] RemoveIPFromPDNS: subnet lookup for PTR ip=%s: %v", ip.Address, err)
		}
	}
	ptrZone, ptrName := buildPTR(ip.Address, prefixLen)
	if ptrZone != "" {
		ptrZones, _ := d.svc.Config.GetCtx(ctx, "pdns_ptr_zones")
		if containsZone(ptrZones, ptrZone) {
			if err := client.DeleteRecord(ctx, ptrZone, ptrName, "PTR"); err != nil {
				log.Printf("[pdns] RemoveIPFromPDNS DeleteRecord PTR ip=%s: %v", ip.Address, err)
			}
		}
	}
}

// TestPDNSConnection tests connectivity to the configured PowerDNS API.
func (d *DNSService) TestPDNSConnection(ctx context.Context) error {
	client := d.pdnsClient(ctx)
	if client == nil {
		return fmt.Errorf("PowerDNS is not configured or disabled")
	}
	return client.TestConnection(ctx)
}

// technitiumClient builds a Technitium client from config, or returns nil if not configured.
func (d *DNSService) technitiumClient(ctx context.Context) *technitium.Client {
	apiURL, _ := d.svc.Config.GetCtx(ctx, "technitium_url")
	token, _ := d.svc.Config.GetCtx(ctx, "technitium_token")
	if apiURL == "" || token == "" {
		return nil
	}
	skipTLS, _ := d.svc.Config.GetCtx(ctx, "technitium_skip_tls")
	return technitium.NewClient(apiURL, token, skipTLS == "true")
}

// TestTechnitiumConnection tests connectivity to the configured Technitium DNS server.
func (d *DNSService) TestTechnitiumConnection(ctx context.Context) error {
	client := d.technitiumClient(ctx)
	if client == nil {
		return fmt.Errorf("technitium DNS is not configured")
	}
	return client.TestConnection(ctx)
}

// TestTechnitiumConnectionWith tests a Technitium connection using the provided credentials,
// bypassing saved config. Used by the admin settings test-before-save flow.
func (d *DNSService) TestTechnitiumConnectionWith(ctx context.Context, apiURL, token string, skipTLS bool) error {
	if apiURL == "" || token == "" {
		return fmt.Errorf("URL and token are required")
	}
	return technitium.NewClient(apiURL, token, skipTLS).TestConnection(ctx)
}

// ListPDNSZones returns the list of zones from PowerDNS, or an empty list if not configured.
func (d *DNSService) ListPDNSZones(ctx context.Context) ([]pdns.Zone, bool, error) {
	client := d.pdnsClient(ctx)
	if client == nil {
		return nil, false, nil // not configured
	}
	zones, err := client.ListZones(ctx)
	return zones, true, err
}

// GetPDNSZone returns the full detail for a zone, including its rrsets.
func (d *DNSService) GetPDNSZone(ctx context.Context, zone string) (*pdns.ZoneDetail, error) {
	client := d.pdnsClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("PowerDNS is not configured or disabled")
	}
	return client.GetZone(ctx, zone)
}

// ListDNSZones returns a normalized zone list from whichever DNS provider is configured.
// Returns (zones, configured, err). If neither provider is configured, configured is false.
func (d *DNSService) ListDNSZones(ctx context.Context) ([]ZoneInfo, bool, error) {
	// Try PowerDNS first
	if pdnsC := d.pdnsClient(ctx); pdnsC != nil {
		raw, err := pdnsC.ListZones(ctx)
		if err != nil {
			return nil, true, err
		}
		out := make([]ZoneInfo, len(raw))
		for i, z := range raw {
			out[i] = ZoneInfo{Name: z.Name, Kind: z.Kind}
		}
		return out, true, nil
	}
	// Fall back to Technitium
	if techC := d.technitiumClient(ctx); techC != nil {
		raw, err := techC.ListZones(ctx)
		if err != nil {
			return nil, true, err
		}
		out := make([]ZoneInfo, len(raw))
		for i, z := range raw {
			out[i] = ZoneInfo{Name: z.Name, Kind: z.Type}
		}
		return out, true, nil
	}
	return nil, false, nil
}

// GetDNSZoneRecords returns normalized DNS records for a zone from whichever provider is configured.
func (d *DNSService) GetDNSZoneRecords(ctx context.Context, zone, typeFilter string) ([]RecordInfo, error) {
	// Try PowerDNS first
	if pdnsC := d.pdnsClient(ctx); pdnsC != nil {
		detail, err := pdnsC.GetZone(ctx, zone)
		if err != nil {
			return nil, err
		}
		var out []RecordInfo
		for _, rr := range detail.RRSets {
			if typeFilter != "" && rr.Type != typeFilter {
				continue
			}
			for _, rec := range rr.Records {
				if rec.Disabled {
					continue
				}
				out = append(out, RecordInfo{
					Name:    rr.Name,
					Type:    rr.Type,
					TTL:     rr.TTL,
					Content: rec.Content,
				})
			}
		}
		return out, nil
	}
	// Fall back to Technitium
	if techC := d.technitiumClient(ctx); techC != nil {
		raw, err := techC.GetZoneRecords(ctx, zone)
		if err != nil {
			return nil, err
		}
		var out []RecordInfo
		for _, rec := range raw {
			if typeFilter != "" && rec.Type != typeFilter {
				continue
			}
			out = append(out, RecordInfo{
				Name:    rec.Name,
				Type:    rec.Type,
				TTL:     rec.TTL,
				Content: rec.Content(),
			})
		}
		return out, nil
	}
	return nil, fmt.Errorf("no DNS provider configured")
}

// SyncIPToTechnitium creates an A/AAAA record in Technitium for the IP's dns_name
// and, if a reverse zone is configured, a PTR record too.
// Requires technitium_default_zone to be configured. DNS failures are logged but do not block the caller.
func (d *DNSService) SyncIPToTechnitium(ctx context.Context, ip *models.IPAddress) {
	client := d.technitiumClient(ctx)
	if client == nil {
		return
	}
	if ip.DNSName == nil || *ip.DNSName == "" {
		return
	}
	zone, _ := d.svc.Config.GetCtx(ctx, "technitium_default_zone")
	if zone == "" {
		log.Printf("[technitium] SyncIPToTechnitium: technitium_default_zone not configured")
		return
	}
	fqdn := *ip.DNSName
	if !strings.HasSuffix(fqdn, ".") {
		fqdn += "."
	}
	if err := client.AddRecord(ctx, zone, fqdn, ip.Address); err != nil {
		log.Printf("[technitium] SyncIPToTechnitium AddRecord ip=%s dns=%s: %v", ip.Address, fqdn, err)
	}
	// PTR record — reuse buildPTR helper with subnet prefix for IPv6
	prefixLen := 0
	if strings.Contains(ip.Address, ":") && ip.SubnetID != 0 {
		if subnet, err := d.svc.repository.GetSubnetByID(ctx, ip.SubnetID); err == nil {
			prefixLen = subnet.PrefixLength
		}
	}
	ptrZone, ptrName := buildPTR(ip.Address, prefixLen)
	if ptrZone != "" {
		ptrZones, _ := d.svc.Config.GetCtx(ctx, "technitium_ptr_zones")
		if containsZone(ptrZones, ptrZone) {
			if err := client.AddPTRRecord(ctx, ptrZone, ptrName, fqdn); err != nil {
				log.Printf("[technitium] SyncIPToTechnitium AddPTRRecord ip=%s: %v", ip.Address, err)
			}
		}
	}
}

// RemoveIPFromTechnitium deletes the A/AAAA and PTR records for an IP from Technitium.
// DNS failures are logged but do not block the caller.
func (d *DNSService) RemoveIPFromTechnitium(ctx context.Context, ip *models.IPAddress) {
	client := d.technitiumClient(ctx)
	if client == nil {
		return
	}
	if ip.DNSName == nil || *ip.DNSName == "" {
		return
	}
	zone, _ := d.svc.Config.GetCtx(ctx, "technitium_default_zone")
	if zone == "" {
		return
	}
	if err := client.DeleteRecord(ctx, zone, *ip.DNSName, ip.Address); err != nil {
		log.Printf("[technitium] RemoveIPFromTechnitium DeleteRecord ip=%s dns=%s: %v", ip.Address, *ip.DNSName, err)
	}
	// PTR record
	prefixLen := 0
	if strings.Contains(ip.Address, ":") && ip.SubnetID != 0 {
		if subnet, err := d.svc.repository.GetSubnetByID(ctx, ip.SubnetID); err == nil {
			prefixLen = subnet.PrefixLength
		}
	}
	ptrZone, ptrName := buildPTR(ip.Address, prefixLen)
	if ptrZone != "" {
		ptrZones, _ := d.svc.Config.GetCtx(ctx, "technitium_ptr_zones")
		if containsZone(ptrZones, ptrZone) {
			if err := client.DeletePTRRecord(ctx, ptrZone, ptrName); err != nil {
				log.Printf("[technitium] RemoveIPFromTechnitium DeletePTRRecord ip=%s: %v", ip.Address, err)
			}
		}
	}
}

// SyncIPToDNS syncs an IP's dns_name to whichever DNS provider is configured.
func (d *DNSService) SyncIPToDNS(ctx context.Context, ip *models.IPAddress) {
	if d.pdnsClient(ctx) != nil {
		d.SyncIPToPDNS(ctx, ip)
		return
	}
	d.SyncIPToTechnitium(ctx, ip)
}

// RemoveIPFromDNS removes an IP's dns_name from whichever DNS provider is configured.
func (d *DNSService) RemoveIPFromDNS(ctx context.Context, ip *models.IPAddress) {
	if d.pdnsClient(ctx) != nil {
		d.RemoveIPFromPDNS(ctx, ip)
		return
	}
	d.RemoveIPFromTechnitium(ctx, ip)
}

// SyncDNSZoneIPs reads all A/AAAA records from the configured DNS provider and
// optionally creates IPAM records for newly-discovered addresses (auto-add) or
// removes IPAM records that are no longer present in DNS (auto-remove).
// IPs managed by this function use "dns-sync" as the hostname prefix so they
// can be identified for auto-remove.
func (d *DNSService) SyncDNSZoneIPs(ctx context.Context) error {
	autoAdd, _ := d.svc.Config.GetCtx(ctx, "dns_auto_add_ips_enabled")
	autoRemove, _ := d.svc.Config.GetCtx(ctx, "dns_auto_remove_ips_enabled")
	if autoAdd != "true" && autoRemove != "true" {
		return nil
	}

	// Collect all A/AAAA records from the active DNS provider.
	zones, configured, err := d.ListDNSZones(ctx)
	if err != nil {
		return fmt.Errorf("SyncDNSZoneIPs: list zones: %w", err)
	}
	if !configured {
		return nil
	}

	// dnsIPs maps IP address string → DNS name (first record wins).
	dnsIPs := make(map[string]string)
	for _, zone := range zones {
		records, recErr := d.GetDNSZoneRecords(ctx, zone.Name, "")
		if recErr != nil {
			log.Printf("[dns-sync] list records for zone %s: %v", zone.Name, recErr)
			continue
		}
		for _, rec := range records {
			if rec.Type != "A" && rec.Type != "AAAA" {
				continue
			}
			if rec.Content == "" {
				continue
			}
			if _, exists := dnsIPs[rec.Content]; !exists {
				dnsIPs[rec.Content] = strings.TrimSuffix(rec.Name, ".")
			}
		}
	}

	// Load all subnets so we can find a home for each DNS IP.
	subnets, err := d.svc.repository.ListAllSubnets(ctx)
	if err != nil {
		return fmt.Errorf("SyncDNSZoneIPs: list subnets: %w", err)
	}

	// Build parsed subnet list for CIDR containment checks.
	type parsedSubnet struct {
		id      int64
		network *net.IPNet
	}
	var parsed []parsedSubnet
	for _, s := range subnets {
		cidr := fmt.Sprintf("%s/%d", s.NetworkAddress, s.PrefixLength)
		_, ipnet, parseErr := net.ParseCIDR(cidr)
		if parseErr != nil {
			continue
		}
		parsed = append(parsed, parsedSubnet{id: s.ID, network: ipnet})
	}

	findSubnetForIP := func(ipStr string) int64 {
		candidate := net.ParseIP(ipStr)
		if candidate == nil {
			return 0
		}
		var bestID int64
		var bestBits int = -1
		for _, ps := range parsed {
			if ps.network.Contains(candidate) {
				ones, _ := ps.network.Mask.Size()
				if ones > bestBits {
					bestBits = ones
					bestID = ps.id
				}
			}
		}
		return bestID
	}

	// Auto-add: for each DNS IP not already in IPAM, create it.
	if autoAdd == "true" {
		for ipStr, dnsName := range dnsIPs {
			subnetID := findSubnetForIP(ipStr)
			if subnetID == 0 {
				continue // no matching subnet
			}
			existing, lookupErr := d.svc.repository.GetIPAddressBySubnetAndAddress(ctx, subnetID, ipStr)
			if lookupErr == nil && existing != nil {
				continue // already exists
			}
			hn := dnsName
			_, createErr := d.svc.repository.CreateIPAddress(ctx, subnetID, ipStr, hn, "active", nil, nil, nil, nil, nil)
			if createErr != nil {
				log.Printf("[dns-sync] create IP %s in subnet %d: %v", ipStr, subnetID, createErr)
			} else {
				log.Printf("[dns-sync] added IP %s (dns=%s) to subnet %d", ipStr, dnsName, subnetID)
			}
		}
	}

	// Auto-remove: find IPAM IPs with hostname prefix "dns-sync:" that no longer appear in DNS.
	if autoRemove == "true" {
		for _, s := range subnets {
			ips, listErr := d.svc.repository.ListIPAddressesBySubnet(ctx, s.ID)
			if listErr != nil {
				log.Printf("[dns-sync] list IPs for subnet %d: %v", s.ID, listErr)
				continue
			}
			for _, ip := range ips {
				if !strings.HasPrefix(ip.Hostname, "dns-sync:") {
					continue
				}
				if _, stillInDNS := dnsIPs[ip.Address]; !stillInDNS {
					if delErr := d.svc.repository.DeleteIPAddress(ctx, ip.ID); delErr != nil {
						log.Printf("[dns-sync] remove IP %s (id=%d): %v", ip.Address, ip.ID, delErr)
					} else {
						log.Printf("[dns-sync] removed IP %s (no longer in DNS)", ip.Address)
					}
				}
			}
		}
	}

	return nil
}

// AutoSyncDNSZoneIPs is a periodic wrapper around SyncDNSZoneIPs that logs errors
// but does not propagate them. Intended to be called from a background scheduler.
func (d *DNSService) AutoSyncDNSZoneIPs(ctx context.Context) {
	if err := d.SyncDNSZoneIPs(ctx); err != nil {
		log.Printf("[dns-sync] AutoSyncDNSZoneIPs: %v", err)
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// buildPTR returns the PTR zone and record name for an IP address.
// For IPv4 the zone is fixed at the /24 boundary (e.g. "1.168.192.in-addr.arpa.").
// For IPv6, prefixLen is rounded down to a nibble boundary to compute the zone
// (e.g. prefixLen=48 → 12-nibble zone "0.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa.").
// Returns empty strings for unparseable addresses.
func buildPTR(address string, prefixLen int) (zone, name string) {
	ip := net.ParseIP(address)
	if ip == nil {
		return "", ""
	}
	if ip4 := ip.To4(); ip4 != nil {
		parts := strings.Split(ip4.String(), ".")
		zone = fmt.Sprintf("%s.%s.%s.in-addr.arpa.", parts[2], parts[1], parts[0])
		name = fmt.Sprintf("%s.%s.%s.%s.in-addr.arpa.", parts[3], parts[2], parts[1], parts[0])
		return zone, name
	}
	// IPv6: expand to 32 hex nibbles, reverse for PTR name and zone.
	ip6 := ip.To16()
	if ip6 == nil {
		return "", ""
	}
	nibbles := make([]byte, 32)
	const hexChars = "0123456789abcdef"
	for i := 0; i < 16; i++ {
		nibbles[i*2] = hexChars[ip6[i]>>4]
		nibbles[i*2+1] = hexChars[ip6[i]&0xf]
	}
	// Full reversed name (nibbles most-significant last).
	nameNibbles := make([]string, 32)
	for i := 0; i < 32; i++ {
		nameNibbles[i] = string(nibbles[31-i])
	}
	name = strings.Join(nameNibbles, ".") + ".ip6.arpa."
	// Zone uses the most-significant prefixLen/4 nibbles (min 1, max 32).
	nibbleCount := prefixLen / 4
	if nibbleCount < 1 {
		nibbleCount = 1
	}
	if nibbleCount > 32 {
		nibbleCount = 32
	}
	zoneNibbles := make([]string, nibbleCount)
	for i := 0; i < nibbleCount; i++ {
		zoneNibbles[i] = string(nibbles[nibbleCount-1-i])
	}
	zone = strings.Join(zoneNibbles, ".") + ".ip6.arpa."
	return zone, name
}

// containsZone reports whether the comma-separated ptrZones string contains the target zone.
func containsZone(ptrZones, target string) bool {
	for _, z := range strings.Split(ptrZones, ",") {
		if strings.TrimSpace(z) == target {
			return true
		}
	}
	return false
}
