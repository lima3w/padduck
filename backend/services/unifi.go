package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"padduck/internal/unifi"
	"padduck/models"
	"padduck/repository"
)

// UniFiConfig is the connection configuration for the UniFi integration.
type UniFiConfig struct {
	URL      string `json:"url"`
	APIKey   string `json:"api_key"`
	Username string `json:"username"`
	Password string `json:"password"`
	Site     string `json:"site"`
	SkipTLS  bool   `json:"skip_tls"`
}

// Complete reports whether there is enough to attempt a connection.
func (c UniFiConfig) Complete() bool {
	return c.URL != "" && (c.APIKey != "" || (c.Username != "" && c.Password != ""))
}

// UniFiSyncResult summarizes one sync run.
type UniFiSyncResult struct {
	VLANsCreated     int      `json:"vlans_created"`
	VLANsLinked      int      `json:"vlans_linked"`
	ClientsCreated   int      `json:"clients_created"`
	ClientsUpdated   int      `json:"clients_updated"`
	ClientsUnmatched int      `json:"clients_unmatched"`
	DevicesCreated   int      `json:"devices_created"`
	DevicesLinked    int      `json:"devices_linked"`
	DevicesUpdated   int      `json:"devices_updated"`
	ConflictsRaised  int      `json:"conflicts_raised"`
	Errors           []string `json:"errors,omitempty"`
}

// UniFiStatus is the outcome of the last sync, for the settings page.
type UniFiStatus struct {
	Configured   bool             `json:"configured"`
	LastSyncedAt *time.Time       `json:"last_synced_at,omitempty"`
	LastError    string           `json:"last_error"`
	LastResult   *UniFiSyncResult `json:"last_result,omitempty"`
	OpenConflict int              `json:"open_conflicts"`
}

const (
	unifiSyncedAtKey = "unifi_last_synced_at"
	unifiErrorKey    = "unifi_last_sync_error"
	unifiResultKey   = "unifi_last_sync_result"
	maxSyncErrors    = 20
)

// unifiSyncMu serializes sync runs (scheduled and manual).
var unifiSyncMu sync.Mutex

// UniFiConfigFromSettings reads the saved UniFi connection settings.
func (s *Service) UniFiConfigFromSettings(ctx context.Context) UniFiConfig {
	get := func(k string) string { v, _ := s.Config.GetCtx(ctx, k); return v }
	return UniFiConfig{
		URL:      get("unifi_url"),
		APIKey:   get("unifi_api_key"),
		Username: get("unifi_username"),
		Password: get("unifi_password"),
		Site:     get("unifi_site"),
		SkipTLS:  get("unifi_skip_tls") == "true",
	}
}

func newUniFiAPI(c UniFiConfig) *unifi.API {
	return unifi.New(unifi.Config{BaseURL: c.URL, APIKey: c.APIKey, Username: c.Username, Password: c.Password, Site: c.Site, SkipTLS: c.SkipTLS})
}

// TestUniFiConnection checks the given connection settings (or, when cfg is
// nil, the saved ones): authentication plus that the configured site exists.
func (s *Service) TestUniFiConnection(ctx context.Context, cfg *UniFiConfig) error {
	c := s.UniFiConfigFromSettings(ctx)
	if cfg != nil {
		c = *cfg
	}
	if !c.Complete() {
		return fmt.Errorf("a URL and either an API key or a username and password are required")
	}
	return newUniFiAPI(c).TestConnection(ctx)
}

// SyncUniFi pulls networks, clients and devices from UniFi into Padduck.
//
//   - Networks with a VLAN tag are matched to Padduck VLANs by VLAN number
//     (created when missing) and linked so the UniFi network name can be shown.
//   - Clients become IP records in the subnet that contains their address.
//     Blank hostname/MAC fields are filled in; a populated field that differs
//     is never overwritten, but raised as a conflict for review. Clients outside
//     every tracked subnet are counted, not created.
//   - UniFi devices are linked to the matching device (by earlier link, else by
//     hostname) or created, and their status and firmware are kept current.
func (s *Service) SyncUniFi(ctx context.Context) (*UniFiSyncResult, error) {
	if !unifiSyncMu.TryLock() {
		return nil, fmt.Errorf("a UniFi sync is already running")
	}
	defer unifiSyncMu.Unlock()

	cfg := s.UniFiConfigFromSettings(ctx)
	if !cfg.Complete() {
		return nil, fmt.Errorf("UniFi is not configured")
	}

	res := &UniFiSyncResult{}
	err := s.runUniFiSync(ctx, newUniFiAPI(cfg), res)
	s.recordUniFiSync(ctx, res, err)
	return res, err
}

func (res *UniFiSyncResult) addError(format string, args ...any) {
	if len(res.Errors) < maxSyncErrors {
		res.Errors = append(res.Errors, fmt.Sprintf(format, args...))
	}
}

func (s *Service) runUniFiSync(ctx context.Context, api *unifi.API, res *UniFiSyncResult) error {
	networks, err := api.ListNetworks(ctx)
	if err != nil {
		return fmt.Errorf("listing networks: %w", err)
	}
	clients, err := api.ListClients(ctx)
	if err != nil {
		return fmt.Errorf("listing clients: %w", err)
	}
	devices, err := api.ListDevices(ctx)
	if err != nil {
		return fmt.Errorf("listing devices: %w", err)
	}

	s.syncUniFiNetworks(ctx, networks, res)
	s.syncUniFiClients(ctx, clients, res)
	s.syncUniFiDevices(ctx, devices, res)
	return nil
}

func (s *Service) syncUniFiNetworks(ctx context.Context, networks []unifi.Network, res *UniFiSyncResult) {
	for _, n := range networks {
		num := n.VLANNumber()
		if num == 0 {
			continue // untagged networks have no VLAN to reconcile
		}
		vlan, err := s.repository.FindVLANByNumber(ctx, num)
		if err != nil {
			vlan, err = s.repository.CreateVLAN(ctx, nil, nil, nil, num, n.Name, "Imported from UniFi")
			if err != nil {
				res.addError("creating VLAN %d (%s): %v", num, n.Name, err)
				continue
			}
			res.VLANsCreated++
		}
		if err := s.repository.UpsertUniFiVLANLink(ctx, vlan.ID, n.ID, n.Name); err != nil {
			res.addError("linking VLAN %d: %v", num, err)
			continue
		}
		res.VLANsLinked++
	}
}

// normalizeMAC returns aa:bb:cc:dd:ee:ff, or "" when the input is not a MAC.
func normalizeMAC(s string) string {
	hw, err := net.ParseMAC(strings.TrimSpace(s))
	if err != nil {
		return ""
	}
	return strings.ToLower(hw.String())
}

func (s *Service) syncUniFiClients(ctx context.Context, clients []unifi.Client, res *UniFiSyncResult) {
	now := time.Now()
	for _, c := range clients {
		mac := normalizeMAC(c.MAC)
		if net.ParseIP(c.IP) == nil || mac == "" {
			continue
		}
		sub, err := s.repository.FindSubnetContaining(ctx, c.IP)
		if err != nil {
			res.ClientsUnmatched++
			continue
		}
		name := c.DisplayName()
		ip, err := s.repository.FindIPInSubnet(ctx, sub.ID, c.IP)
		if err != nil {
			if _, err := s.repository.CreateIPAddress(ctx, sub.ID, c.IP, name, "assigned", nil, &mac, nil, nil); err != nil {
				res.addError("creating IP %s: %v", c.IP, err)
				continue
			}
			res.ClientsCreated++
			continue
		}
		if s.reconcileUniFiClientIP(ctx, ip, name, mac, res) {
			res.ClientsUpdated++
		}
		_ = s.repository.UpdateLastSeen(ctx, ip.ID, now)
	}
}

// reconcileUniFiClientIP fills blanks and raises conflicts for differences.
// It reports whether the record was changed.
func (s *Service) reconcileUniFiClientIP(ctx context.Context, ip *models.IPAddress, name, mac string, res *UniFiSyncResult) bool {
	changed := false

	if name != "" {
		switch {
		case ip.Hostname == "":
			if err := s.repository.SetIPHostname(ctx, ip.ID, name); err != nil {
				res.addError("updating hostname for %s: %v", ip.Address, err)
			} else {
				changed = true
			}
		case !strings.EqualFold(ip.Hostname, name):
			s.raiseUniFiConflict(ctx, ip, "hostname", ip.Hostname, name, res)
		}
	}

	if ip.MACAddress == nil || *ip.MACAddress == "" {
		if err := s.repository.SetIPMAC(ctx, ip.ID, mac); err != nil {
			res.addError("updating MAC for %s: %v", ip.Address, err)
		} else {
			changed = true
		}
	} else if normalizeMAC(*ip.MACAddress) != mac {
		s.raiseUniFiConflict(ctx, ip, "mac_address", *ip.MACAddress, mac, res)
	}

	// An address UniFi sees a client on is in use, whatever the record says.
	if ip.Status == "available" {
		if _, err := s.repository.UpdateIPAddressStatus(ctx, ip.ID, "assigned", ip.DeviceID); err != nil {
			res.addError("marking %s assigned: %v", ip.Address, err)
		} else {
			changed = true
		}
	}
	return changed
}

func (s *Service) raiseUniFiConflict(ctx context.Context, ip *models.IPAddress, field, padduck, unifiVal string, res *UniFiSyncResult) {
	created, err := s.repository.CreateUniFiConflict(ctx, ip.ID, field, padduck, unifiVal)
	if err != nil {
		res.addError("recording %s conflict for %s: %v", field, ip.Address, err)
		return
	}
	if created {
		res.ConflictsRaised++
	}
}

func (s *Service) syncUniFiDevices(ctx context.Context, devices []unifi.Device, res *UniFiSyncResult) {
	for _, d := range devices {
		mac := normalizeMAC(d.MAC)
		if mac == "" || !d.Adopted {
			continue
		}
		name := d.DisplayName()
		id, linked, err := s.repository.FindDeviceIDByUniFiID(ctx, mac)
		if err != nil {
			res.addError("looking up device %s: %v", name, err)
			continue
		}
		if !linked {
			id, linked, err = s.repository.FindUnlinkedDeviceIDByHostname(ctx, name)
			if err != nil {
				res.addError("looking up device %s: %v", name, err)
				continue
			}
			if linked {
				res.DevicesLinked++
			}
		} else {
			res.DevicesUpdated++
		}
		if !linked {
			vendor, model, version := "Ubiquiti", d.Model, d.Version
			dev, err := s.repository.CreateDevice(ctx, &repository.DeviceParams{
				Hostname: name, Vendor: &vendor, Model: &model, OSVersion: &version, RackUnitSize: 1,
			})
			if err != nil {
				res.addError("creating device %s: %v", name, err)
				continue
			}
			id = dev.ID
			res.DevicesCreated++
		}
		if err := s.repository.UpdateDeviceFromUniFi(ctx, id, "Ubiquiti", d.Model, d.Version, d.Online()); err != nil {
			res.addError("updating device %s: %v", name, err)
			continue
		}
		details := map[string]any{"mac": mac, "model": d.Model, "firmware": d.Version, "type": d.Type, "ip": d.IP, "online": d.Online()}
		if err := s.repository.UpsertUniFiDeviceLink(ctx, id, mac, name, details); err != nil {
			res.addError("linking device %s: %v", name, err)
		}
	}
}

func (s *Service) recordUniFiSync(ctx context.Context, res *UniFiSyncResult, syncErr error) {
	_ = s.Config.SetCtx(ctx, unifiSyncedAtKey, time.Now().UTC().Format(time.RFC3339))
	msg := ""
	if syncErr != nil {
		msg = syncErr.Error()
	}
	_ = s.Config.SetCtx(ctx, unifiErrorKey, msg)
	if raw, err := json.Marshal(res); err == nil {
		_ = s.Config.SetCtx(ctx, unifiResultKey, string(raw))
	}
}

// UniFiStatus returns the saved sync status and open conflict count.
func (s *Service) UniFiStatus(ctx context.Context) (*UniFiStatus, error) {
	st := &UniFiStatus{Configured: s.UniFiConfigFromSettings(ctx).Complete()}
	if v, _ := s.Config.GetCtx(ctx, unifiSyncedAtKey); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			st.LastSyncedAt = &t
		}
	}
	st.LastError, _ = s.Config.GetCtx(ctx, unifiErrorKey)
	if v, _ := s.Config.GetCtx(ctx, unifiResultKey); v != "" {
		r := &UniFiSyncResult{}
		if json.Unmarshal([]byte(v), r) == nil {
			st.LastResult = r
		}
	}
	open, err := s.repository.ListUniFiConflicts(ctx, "open")
	if err != nil {
		return nil, err
	}
	st.OpenConflict = len(open)
	return st, nil
}

// ListUniFiConflicts lists conflicts by status ("" = all).
func (s *Service) ListUniFiConflicts(ctx context.Context, status string) ([]*models.UniFiConflict, error) {
	return s.repository.ListUniFiConflicts(ctx, status)
}

// ListUniFiLinks lists VLAN and device links for display.
func (s *Service) ListUniFiLinks(ctx context.Context) ([]*models.UniFiLink, error) {
	return s.repository.ListUniFiLinks(ctx)
}

// ResolveUniFiConflict accepts (applies UniFi's value to the IP record) or
// dismisses (keeps Padduck's value) an open conflict.
func (s *Service) ResolveUniFiConflict(ctx context.Context, id int64, action string) (*models.UniFiConflict, error) {
	if action != "accept" && action != "dismiss" {
		return nil, fmt.Errorf("action must be accept or dismiss")
	}
	c, err := s.repository.GetUniFiConflict(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("conflict not found")
	}
	if c.Status != "open" {
		return nil, fmt.Errorf("conflict is already %s", c.Status)
	}
	if action == "accept" {
		switch c.Field {
		case "hostname":
			err = s.repository.SetIPHostname(ctx, c.IPID, c.UniFiValue)
		case "mac_address":
			err = s.repository.SetIPMAC(ctx, c.IPID, c.UniFiValue)
		}
		if err != nil {
			return nil, fmt.Errorf("applying UniFi value: %w", err)
		}
	}
	status := "dismissed"
	if action == "accept" {
		status = "accepted"
	}
	if err := s.repository.ResolveUniFiConflict(ctx, id, status); err != nil {
		return nil, fmt.Errorf("conflict is no longer open")
	}
	return s.repository.GetUniFiConflict(ctx, id)
}

// unifiSyncInterval reads the scheduled sync interval. Unlike the DHCP sync it
// defaults to off: a UniFi sync creates VLANs, IP records and devices, so it
// should be an explicit opt-in after a first manual run.
func (s *Service) unifiSyncInterval(ctx context.Context) time.Duration {
	v, _ := s.Config.GetCtx(ctx, "unifi_sync_interval_minutes")
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || n < 0 {
		return 0
	}
	return time.Duration(n) * time.Minute
}

// StartUniFiSync launches the optional scheduled sync. The interval is re-read
// each minute so admin changes apply without a restart.
func (s *Service) StartUniFiSync(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		var last time.Time
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				interval := s.unifiSyncInterval(ctx)
				if interval == 0 || time.Since(last) < interval {
					continue
				}
				if !s.UniFiConfigFromSettings(ctx).Complete() {
					continue
				}
				last = time.Now()
				if _, err := s.SyncUniFi(ctx); err != nil {
					log.Printf("unifi sync: %v", err)
				}
			}
		}
	}()
}
