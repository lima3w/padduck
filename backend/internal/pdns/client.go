// Package pdns provides a minimal PowerDNS Authoritative API client.
package pdns

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Zone represents a PowerDNS zone.
type Zone struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Kind   string `json:"kind"`
	URL    string `json:"url"`
	Serial uint32 `json:"serial"`
}

// RRSet is a resource record set inside a zone.
type RRSet struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	TTL        int      `json:"ttl"`
	ChangeType string   `json:"changetype,omitempty"`
	Records    []Record `json:"records"`
}

// Record is a single DNS record value.
type Record struct {
	Content  string `json:"content"`
	Disabled bool   `json:"disabled"`
}

// ZoneDetail includes the full rrsets for a zone.
type ZoneDetail struct {
	Zone
	RRSets []RRSet `json:"rrsets"`
}

// Client is a PowerDNS API client.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient returns a new PowerDNS API Client.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	return c.httpClient.Do(req)
}

func checkStatus(resp *http.Response, want int) error {
	if resp.StatusCode == want {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
}

// TestConnection pings the PowerDNS API server.
func (c *Client) TestConnection(ctx context.Context) error {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/servers/localhost", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkStatus(resp, http.StatusOK)
}

// ListZones returns all zones on the localhost server.
func (c *Client) ListZones(ctx context.Context) ([]Zone, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/servers/localhost/zones", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := checkStatus(resp, http.StatusOK); err != nil {
		return nil, err
	}
	var zones []Zone
	if err := json.NewDecoder(resp.Body).Decode(&zones); err != nil {
		return nil, fmt.Errorf("decoding zones: %w", err)
	}
	return zones, nil
}

// GetZone returns full detail (including rrsets) for a zone.
func (c *Client) GetZone(ctx context.Context, zone string) (*ZoneDetail, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/servers/localhost/zones/"+zone, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := checkStatus(resp, http.StatusOK); err != nil {
		return nil, err
	}
	var zd ZoneDetail
	if err := json.NewDecoder(resp.Body).Decode(&zd); err != nil {
		return nil, fmt.Errorf("decoding zone detail: %w", err)
	}
	return &zd, nil
}

// CreateRecord creates (patches) an rrset in the given zone.
func (c *Client) CreateRecord(ctx context.Context, zone, name, rtype, content string, ttl int) error {
	payload := map[string]interface{}{
		"rrsets": []RRSet{
			{
				Name:       name,
				Type:       rtype,
				TTL:        ttl,
				ChangeType: "REPLACE",
				Records: []Record{
					{Content: content, Disabled: false},
				},
			},
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := c.do(ctx, http.MethodPatch, "/api/v1/servers/localhost/zones/"+zone, strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkStatus(resp, http.StatusNoContent)
}

// DeleteRecord deletes an rrset from the given zone.
func (c *Client) DeleteRecord(ctx context.Context, zone, name, rtype string) error {
	payload := map[string]interface{}{
		"rrsets": []RRSet{
			{
				Name:       name,
				Type:       rtype,
				ChangeType: "DELETE",
			},
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := c.do(ctx, http.MethodPatch, "/api/v1/servers/localhost/zones/"+zone, strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkStatus(resp, http.StatusNoContent)
}
