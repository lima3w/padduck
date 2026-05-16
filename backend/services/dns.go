package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"ipam-next/internal/pdns"
	"ipam-next/internal/technitium"
	"ipam-next/models"
)

// DNSService handles DNS lookups, record tracking, and PowerDNS sync.
type DNSService struct {
	svc *Service
}

// NewDNSService creates a new DNSService.
func NewDNSService(svc *Service) *DNSService {
	return &DNSService{svc: svc}
}

// pdnsClient builds a PowerDNS client from config, or returns nil if not configured.
func (d *DNSService) pdnsClient() *pdns.Client {
	enabled, _ := d.svc.Config.Get("pdns_enabled")
	if enabled != "true" {
		return nil
	}
	apiURL, _ := d.svc.Config.Get("pdns_api_url")
	apiKey, _ := d.svc.Config.Get("pdns_api_key")
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

	return d.svc.repository.UpdateIPDNSFields(ctx, ipID, ptrRecord, json.RawMessage(records), time.Now())
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
	client := d.pdnsClient()
	if client == nil {
		return
	}
	if ip.DNSName == nil || *ip.DNSName == "" {
		return
	}

	defaultZone, _ := d.svc.Config.Get("pdns_default_zone")
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
	ptrZone, ptrName := buildPTR(ip.Address)
	if ptrZone != "" {
		ptrZones, _ := d.svc.Config.Get("pdns_ptr_zones")
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
	client := d.pdnsClient()
	if client == nil {
		return
	}
	if ip.DNSName == nil || *ip.DNSName == "" {
		return
	}

	defaultZone, _ := d.svc.Config.Get("pdns_default_zone")
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

	ptrZone, ptrName := buildPTR(ip.Address)
	if ptrZone != "" {
		ptrZones, _ := d.svc.Config.Get("pdns_ptr_zones")
		if containsZone(ptrZones, ptrZone) {
			if err := client.DeleteRecord(ctx, ptrZone, ptrName, "PTR"); err != nil {
				log.Printf("[pdns] RemoveIPFromPDNS DeleteRecord PTR ip=%s: %v", ip.Address, err)
			}
		}
	}
}

// TestPDNSConnection tests connectivity to the configured PowerDNS API.
func (d *DNSService) TestPDNSConnection(ctx context.Context) error {
	client := d.pdnsClient()
	if client == nil {
		return fmt.Errorf("PowerDNS is not configured or disabled")
	}
	return client.TestConnection(ctx)
}

// technitiumClient builds a Technitium client from config, or returns nil if not configured.
func (d *DNSService) technitiumClient() *technitium.Client {
	apiURL, _ := d.svc.Config.Get("technitium_url")
	token, _ := d.svc.Config.Get("technitium_token")
	if apiURL == "" || token == "" {
		return nil
	}
	return technitium.NewClient(apiURL, token)
}

// TestTechnitiumConnection tests connectivity to the configured Technitium DNS server.
func (d *DNSService) TestTechnitiumConnection(ctx context.Context) error {
	client := d.technitiumClient()
	if client == nil {
		return fmt.Errorf("Technitium DNS is not configured")
	}
	return client.TestConnection(ctx)
}

// ListPDNSZones returns the list of zones from PowerDNS, or an empty list if not configured.
func (d *DNSService) ListPDNSZones(ctx context.Context) ([]pdns.Zone, bool, error) {
	client := d.pdnsClient()
	if client == nil {
		return nil, false, nil // not configured
	}
	zones, err := client.ListZones(ctx)
	return zones, true, err
}

// GetPDNSZone returns the full detail for a zone, including its rrsets.
func (d *DNSService) GetPDNSZone(ctx context.Context, zone string) (*pdns.ZoneDetail, error) {
	client := d.pdnsClient()
	if client == nil {
		return nil, fmt.Errorf("PowerDNS is not configured or disabled")
	}
	return client.GetZone(ctx, zone)
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

// buildPTR returns the PTR zone and name for an IPv4 address (e.g. "0.0.168.192.in-addr.arpa." and "1.0.168.192.in-addr.arpa.").
// Returns empty strings for non-IPv4 addresses.
func buildPTR(address string) (zone, name string) {
	ip := net.ParseIP(address)
	if ip == nil {
		return "", ""
	}
	ip4 := ip.To4()
	if ip4 == nil {
		// IPv6 PTR not implemented
		return "", ""
	}
	// e.g. 192.168.1.5 → zone=0.168.192.in-addr.arpa., name=5.1.168.192.in-addr.arpa.
	parts := strings.Split(ip4.String(), ".")
	zone = fmt.Sprintf("%s.%s.%s.in-addr.arpa.", parts[2], parts[1], parts[0])
	name = fmt.Sprintf("%s.%s.%s.%s.in-addr.arpa.", parts[3], parts[2], parts[1], parts[0])
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
