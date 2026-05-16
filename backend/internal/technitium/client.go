// Package technitium provides a minimal Technitium DNS Server API client.
package technitium

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Zone represents a Technitium DNS zone.
type Zone struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Disabled bool   `json:"disabled"`
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
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
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

// get performs a GET request to the given path with the token appended.
func (c *Client) get(ctx context.Context, path string, params url.Values) (*http.Response, error) {
	if params == nil {
		params = url.Values{}
	}
	params.Set("token", c.token)
	fullURL := c.baseURL + path + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}
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
		Status   string `json:"status"`
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
	params.Set("domain", zone)  // required; listZone=true returns all records, not just the apex
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

// AddRecord adds a DNS A record in the specified zone.
func (c *Client) AddRecord(ctx context.Context, zone, fqdn, ipAddress string) error {
	params := url.Values{}
	params.Set("zone", zone)
	params.Set("domain", fqdn)
	params.Set("type", "A")
	params.Set("ipAddress", ipAddress)
	resp, err := c.get(ctx, "/api/zones/records/add", params)
	if err != nil {
		return err
	}
	return checkResponse(resp)
}

// DeleteRecord deletes a DNS A record from the specified zone.
func (c *Client) DeleteRecord(ctx context.Context, zone, fqdn string) error {
	params := url.Values{}
	params.Set("zone", zone)
	params.Set("domain", fqdn)
	params.Set("type", "A")
	resp, err := c.get(ctx, "/api/zones/records/delete", params)
	if err != nil {
		return err
	}
	return checkResponse(resp)
}
