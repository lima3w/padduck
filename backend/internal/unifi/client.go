// Package unifi provides a minimal client for the UniFi Network Application
// API. It supports UniFi OS API keys and the classic username/password
// cookie login (UniFi OS consoles and standalone controllers).
package unifi

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Site is a UniFi site.
type Site struct {
	Name string `json:"name"` // internal id, e.g. "default"
	Desc string `json:"desc"`
}

// flexString decodes a JSON string or number into a string. UniFi controllers
// differ on whether a network's "vlan" is "10" or 10.
type flexString string

func (f *flexString) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		*f = flexString(s)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(b, &n); err != nil {
		return fmt.Errorf("expected a string or number: %w", err)
	}
	*f = flexString(n.String())
	return nil
}

// Network is a UniFi network (LAN/VLAN).
type Network struct {
	ID          string     `json:"_id"`
	Name        string     `json:"name"`
	Purpose     string     `json:"purpose"`
	VLANEnabled bool       `json:"vlan_enabled"`
	VLAN        flexString `json:"vlan"`
	IPSubnet    string     `json:"ip_subnet"`
	Enabled     bool       `json:"enabled"`
}

// VLANNumber returns the 802.1Q tag, or 0 when the network is untagged.
func (n Network) VLANNumber() int {
	if !n.VLANEnabled {
		return 0
	}
	var v int
	if _, err := fmt.Sscanf(string(n.VLAN), "%d", &v); err != nil || v < 1 || v > 4094 {
		return 0
	}
	return v
}

// Client is a connected (active) UniFi client.
type Client struct {
	MAC       string `json:"mac"`
	IP        string `json:"ip"`
	Hostname  string `json:"hostname"`
	Name      string `json:"name"`
	NetworkID string `json:"network_id"`
	Network   string `json:"network"`
	IsWired   bool   `json:"is_wired"`
	LastSeen  int64  `json:"last_seen"`
}

// DisplayName prefers the admin-assigned alias, then the reported hostname.
func (c Client) DisplayName() string {
	if c.Name != "" {
		return c.Name
	}
	return c.Hostname
}

// Device is a UniFi network device (AP, switch, gateway).
type Device struct {
	ID      string `json:"_id"`
	MAC     string `json:"mac"`
	Name    string `json:"name"`
	Model   string `json:"model"`
	Version string `json:"version"`
	Type    string `json:"type"`
	IP      string `json:"ip"`
	State   int    `json:"state"` // 1 = connected
	Adopted bool   `json:"adopted"`
}

// Online reports whether the controller sees the device as connected.
func (d Device) Online() bool { return d.State == 1 }

// DisplayName falls back to the MAC when the device has no name.
func (d Device) DisplayName() string {
	if d.Name != "" {
		return d.Name
	}
	return d.MAC
}

// Config holds connection settings. APIKey takes precedence; otherwise
// Username/Password are used for the classic cookie login.
type Config struct {
	BaseURL  string
	APIKey   string
	Username string
	Password string
	Site     string
	SkipTLS  bool
}

// API is a UniFi API client.
type API struct {
	cfg        Config
	httpClient *http.Client

	mu     sync.Mutex
	prefix string // "/proxy/network" on UniFi OS, "" on a standalone controller
	authed bool
}

// New returns a client. Site defaults to "default".
func New(cfg Config) *API {
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	if cfg.Site == "" {
		cfg.Site = "default"
	}
	transport := http.DefaultTransport
	if cfg.SkipTLS {
		slog.Warn("UniFi TLS certificate verification is disabled — do not use in production", "url", cfg.BaseURL)
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // #nosec G402 -- explicit admin UniFi setting.
		}
	}
	jar, _ := cookiejar.New(nil)
	return &API{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout:   20 * time.Second,
			Transport: transport,
			Jar:       jar,
		},
	}
}

// ensureAuth logs in once for the cookie modes and fixes the API path prefix.
func (a *API) ensureAuth(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.authed {
		return nil
	}
	if a.cfg.APIKey != "" {
		a.prefix = "/proxy/network"
		a.authed = true
		return nil
	}
	if a.cfg.Username == "" || a.cfg.Password == "" {
		return fmt.Errorf("an API key or a username and password are required")
	}
	body, err := json.Marshal(map[string]string{"username": a.cfg.Username, "password": a.cfg.Password})
	if err != nil {
		return err
	}
	// UniFi OS console first, then the standalone controller login.
	for _, c := range []struct{ path, prefix string }{
		{"/api/auth/login", "/proxy/network"},
		{"/api/login", ""},
	} {
		status, err := a.postLogin(ctx, c.path, body)
		if err != nil {
			return err
		}
		switch {
		case status == http.StatusOK:
			a.prefix = c.prefix
			a.authed = true
			return nil
		case status == http.StatusNotFound:
			continue
		case status == http.StatusBadRequest || status == http.StatusUnauthorized || status == http.StatusForbidden:
			return fmt.Errorf("login rejected (HTTP %d): check the username and password", status)
		default:
			return fmt.Errorf("login failed (HTTP %d)", status)
		}
	}
	return fmt.Errorf("no UniFi login endpoint found at %s", a.cfg.BaseURL)
}

func (a *API) postLogin(ctx context.Context, path string, body []byte) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.cfg.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	return resp.StatusCode, nil
}

// list GETs a classic-API path and decodes the {"meta":..,"data":[..]} envelope.
func (a *API) list(ctx context.Context, path string, dst any) error {
	if err := a.ensureAuth(ctx); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.cfg.BaseURL+a.prefix+path, nil)
	if err != nil {
		return err
	}
	if a.cfg.APIKey != "" {
		req.Header.Set("X-API-KEY", a.cfg.APIKey)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("UniFi rejected the credentials (HTTP %d)", resp.StatusCode)
	default:
		return fmt.Errorf("unexpected HTTP status %d from %s", resp.StatusCode, path)
	}
	var env struct {
		Meta struct {
			RC  string `json:"rc"`
			Msg string `json:"msg"`
		} `json:"meta"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("parsing response from %s: %w", path, err)
	}
	if env.Meta.RC != "" && env.Meta.RC != "ok" {
		return fmt.Errorf("UniFi API error: %s", env.Meta.Msg)
	}
	return json.Unmarshal(env.Data, dst)
}

func (a *API) sitePath(suffix string) string {
	return "/api/s/" + url.PathEscape(a.cfg.Site) + suffix
}

// ListSites returns the sites visible to the credentials.
func (a *API) ListSites(ctx context.Context) ([]Site, error) {
	var out []Site
	err := a.list(ctx, "/api/self/sites", &out)
	return out, err
}

// ListNetworks returns the configured site's networks.
func (a *API) ListNetworks(ctx context.Context) ([]Network, error) {
	var out []Network
	err := a.list(ctx, a.sitePath("/rest/networkconf"), &out)
	return out, err
}

// ListClients returns the site's currently connected clients.
func (a *API) ListClients(ctx context.Context) ([]Client, error) {
	var out []Client
	err := a.list(ctx, a.sitePath("/stat/sta"), &out)
	return out, err
}

// ListDevices returns the site's UniFi devices.
func (a *API) ListDevices(ctx context.Context) ([]Device, error) {
	var out []Device
	err := a.list(ctx, a.sitePath("/stat/device"), &out)
	return out, err
}

// TestConnection verifies authentication and that the configured site exists.
func (a *API) TestConnection(ctx context.Context) error {
	sites, err := a.ListSites(ctx)
	if err != nil {
		return err
	}
	for _, s := range sites {
		if s.Name == a.cfg.Site {
			return nil
		}
	}
	names := make([]string, 0, len(sites))
	for _, s := range sites {
		names = append(names, s.Name)
	}
	return fmt.Errorf("site %q not found (available: %s)", a.cfg.Site, strings.Join(names, ", "))
}
