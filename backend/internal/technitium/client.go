// Package technitium provides a minimal Technitium DNS Server API client.
package technitium

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DHCPScope represents a Technitium DHCP scope.
type DHCPScope struct {
	Name             string `json:"name"`
	StartingAddress  string `json:"startingAddress"`
	EndingAddress    string `json:"endingAddress"`
	SubnetMask       string `json:"subnetMask"`
	RouterAddress    string `json:"routerAddress"`
	Enabled          bool   `json:"enabled"`
	LeaseTimeDays    int    `json:"leaseTimeDays"`
	LeaseTimeHours   int    `json:"leaseTimeHours"`
	LeaseTimeMinutes int    `json:"leaseTimeMinutes"`
}

// DHCPLease represents a single DHCP lease returned by Technitium.
type DHCPLease struct {
	ClientIdentifier string    `json:"clientIdentifier"`
	IPAddress        string    `json:"address"`
	HardwareAddress  string    `json:"hardwareAddress"`
	HostName         string    `json:"hostName"`
	LeaseType        string    `json:"type"`
	LeaseObtained    time.Time `json:"leaseObtained"`
	LeaseExpires     time.Time `json:"leaseExpires"`
}

// Zone represents a Technitium DNS zone.
type Zone struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Disabled bool   `json:"disabled"`
	Serial   uint32 `json:"serial"`
}

// Record represents a single DNS record in a Technitium zone.
type Record struct {
	Name  string          `json:"name"`
	Type  string          `json:"type"`
	TTL   int             `json:"ttl"`
	RData json.RawMessage `json:"rData"`
}

// Content extracts a human-readable value from the RData object.
// Technitium returns rData as a typed object; we pull the most relevant field per record type.
func (r *Record) Content() string {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(r.RData, &m); err != nil {
		return string(r.RData)
	}
	pick := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := m[k]; ok {
				var s string
				if err := json.Unmarshal(v, &s); err == nil {
					return s
				}
			}
		}
		return ""
	}
	switch r.Type {
	case "A", "AAAA":
		return pick("ipAddress")
	case "PTR":
		return pick("ptrdName")
	case "CNAME":
		return pick("cname")
	case "MX":
		return pick("exchange")
	case "NS":
		return pick("nsDomainName")
	case "TXT":
		return pick("text")
	case "SOA":
		return pick("primaryNameServer")
	default:
		return pick("ipAddress", "cname", "exchange", "nsDomainName", "ptrdName", "text")
	}
}

// Client is a Technitium DNS Server API client.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient returns a new Technitium API Client.
func NewClient(baseURL, token string, skipTLS bool) *Client {
	transport := http.DefaultTransport
	if skipTLS {
		slog.Warn("Technitium TLS certificate verification is disabled — do not use in production", "url", baseURL)
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // #nosec G402 -- explicit admin Technitium setting.
		}
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
		},
	}
}

// get performs a GET request to the given path, authenticating via Authorization header.
func (c *Client) get(ctx context.Context, path string, params url.Values) (*http.Response, error) {
	var rawQuery string
	if len(params) > 0 {
		rawQuery = "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path+rawQuery, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	return c.httpClient.Do(req)
}

// checkResponse reads the JSON response and returns an error if the status is not "ok".
func checkResponse(resp *http.Response) error {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var envelope struct {
		Status       string `json:"status"`
		ErrorMessage string `json:"errorMessage"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}
	if envelope.Status != "ok" {
		if envelope.ErrorMessage != "" {
			return fmt.Errorf("technitium API error: %s", envelope.ErrorMessage)
		}
		return fmt.Errorf("technitium API returned status %q", envelope.Status)
	}
	return nil
}

// checkResponseDecode reads the JSON response, checks for ok status, and decodes into dst.
func checkResponseDecode(resp *http.Response, dst interface{}) error {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var envelope struct {
		Status       string          `json:"status"`
		ErrorMessage string          `json:"errorMessage"`
		Response     json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}
	if envelope.Status != "ok" {
		if envelope.ErrorMessage != "" {
			return fmt.Errorf("technitium API error: %s", envelope.ErrorMessage)
		}
		return fmt.Errorf("technitium API returned status %q", envelope.Status)
	}
	if dst != nil && len(envelope.Response) > 0 {
		if err := json.Unmarshal(envelope.Response, dst); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}
	return nil
}

// TestConnection checks that the Technitium API is reachable and the token is valid.
// It calls /api/zones/list which is available on all supported Technitium versions.
func (c *Client) TestConnection(ctx context.Context) error {
	resp, err := c.get(ctx, "/api/zones/list", nil)
	if err != nil {
		return err
	}
	return checkResponse(resp)
}

// ListZones returns all zones from the Technitium DNS server.
func (c *Client) ListZones(ctx context.Context) ([]Zone, error) {
	resp, err := c.get(ctx, "/api/zones/list", nil)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Zones []Zone `json:"zones"`
	}
	if err := checkResponseDecode(resp, &payload); err != nil {
		return nil, err
	}
	return payload.Zones, nil
}

// GetZoneRecords returns all DNS records for the specified zone.
func (c *Client) GetZoneRecords(ctx context.Context, zone string) ([]Record, error) {
	params := url.Values{}
	params.Set("zone", zone)
	params.Set("domain", zone) // required; listZone=true returns all records, not just the apex
	params.Set("listZone", "true")
	resp, err := c.get(ctx, "/api/zones/records/get", params)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Records []Record `json:"records"`
	}
	if err := checkResponseDecode(resp, &payload); err != nil {
		return nil, err
	}
	return payload.Records, nil
}

// AddRecord adds an A or AAAA record in the specified zone.
// The record type is inferred from ipAddress: an address containing ":" is
// treated as IPv6 (AAAA); otherwise IPv4 (A).
func (c *Client) AddRecord(ctx context.Context, zone, fqdn, ipAddress string) error {
	rtype := "A"
	if strings.Contains(ipAddress, ":") {
		rtype = "AAAA"
	}
	params := url.Values{}
	params.Set("zone", zone)
	params.Set("domain", fqdn)
	params.Set("type", rtype)
	params.Set("ipAddress", ipAddress)
	resp, err := c.get(ctx, "/api/zones/records/add", params)
	if err != nil {
		return err
	}
	return checkResponse(resp)
}

// DeleteRecord deletes an A or AAAA record from the specified zone.
// The record type is inferred from the presence of ":" in ipAddress.
// Pass an empty ipAddress to delete an A record (backwards-compatible default).
func (c *Client) DeleteRecord(ctx context.Context, zone, fqdn, ipAddress string) error {
	rtype := "A"
	if strings.Contains(ipAddress, ":") {
		rtype = "AAAA"
	}
	params := url.Values{}
	params.Set("zone", zone)
	params.Set("domain", fqdn)
	params.Set("type", rtype)
	resp, err := c.get(ctx, "/api/zones/records/delete", params)
	if err != nil {
		return err
	}
	return checkResponse(resp)
}

// AddPTRRecord adds a PTR record in the specified reverse zone.
func (c *Client) AddPTRRecord(ctx context.Context, zone, ptrName, fqdn string) error {
	params := url.Values{}
	params.Set("zone", zone)
	params.Set("domain", ptrName)
	params.Set("type", "PTR")
	params.Set("ptrName", fqdn)
	resp, err := c.get(ctx, "/api/zones/records/add", params)
	if err != nil {
		return err
	}
	return checkResponse(resp)
}

// DeletePTRRecord deletes a PTR record from the specified reverse zone.
func (c *Client) DeletePTRRecord(ctx context.Context, zone, ptrName string) error {
	params := url.Values{}
	params.Set("zone", zone)
	params.Set("domain", ptrName)
	params.Set("type", "PTR")
	resp, err := c.get(ctx, "/api/zones/records/delete", params)
	if err != nil {
		return err
	}
	return checkResponse(resp)
}

// ListDHCPScopes returns all DHCP scopes configured on the server.
func (c *Client) ListDHCPScopes(ctx context.Context) ([]DHCPScope, error) {
	resp, err := c.get(ctx, "/api/dhcp/scopes/list", nil)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Scopes []DHCPScope `json:"scopes"`
	}
	if err := checkResponseDecode(resp, &payload); err != nil {
		return nil, err
	}
	return payload.Scopes, nil
}

// ListDHCPLeases returns all leases for a given scope. Pass an empty scopeName to list all.
func (c *Client) ListDHCPLeases(ctx context.Context, scopeName string) ([]DHCPLease, error) {
	params := url.Values{}
	if scopeName != "" {
		params.Set("scopeName", scopeName)
	}
	resp, err := c.get(ctx, "/api/dhcp/leases/list", params)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Leases []DHCPLease `json:"leases"`
	}
	if err := checkResponseDecode(resp, &payload); err != nil {
		return nil, err
	}
	return payload.Leases, nil
}

// AddDHCPReservation creates a static DHCP reservation in the given scope.
func (c *Client) AddDHCPReservation(ctx context.Context, scopeName, ipAddress, macAddress, hostname string) error {
	params := url.Values{}
	params.Set("name", scopeName)
	params.Set("ipAddress", ipAddress)
	params.Set("hardwareAddress", macAddress)
	params.Set("hostName", hostname)
	resp, err := c.get(ctx, "/api/dhcp/scopes/addReservation", params)
	if err != nil {
		return err
	}
	return checkResponse(resp)
}

// RemoveDHCPReservation removes a static DHCP reservation from the given scope.
func (c *Client) RemoveDHCPReservation(ctx context.Context, scopeName, ipAddress string) error {
	params := url.Values{}
	params.Set("name", scopeName)
	params.Set("ipAddress", ipAddress)
	resp, err := c.get(ctx, "/api/dhcp/scopes/removeReservation", params)
	if err != nil {
		return err
	}
	return checkResponse(resp)
}
