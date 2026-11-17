package services

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"padduck/internal/technitium"
	"padduck/models"
	"padduck/repository"
)

// technitiumDHCPClient builds a Technitium client from global config, returning nil if unconfigured.
func (s *Service) technitiumDHCPClient(ctx context.Context) *technitium.Client {
	apiURL, _ := s.Config.GetCtx(ctx, "technitium_url")
	token, _ := s.Config.GetCtx(ctx, "technitium_token")
	if apiURL == "" || token == "" {
		return nil
	}
	skipTLS, _ := s.Config.GetCtx(ctx, "technitium_skip_tls")
	return technitium.NewClient(apiURL, token, skipTLS == "true")
}

// ListTechnitiumDHCPScopes returns all DHCP scopes from the configured Technitium server.
func (s *Service) ListTechnitiumDHCPScopes(ctx context.Context) ([]technitium.DHCPScope, error) {
	client := s.technitiumDHCPClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("technitium is not configured")
	}
	return client.ListDHCPScopes(ctx)
}

// SyncTechnitiumLeases pulls all DHCP leases from Technitium and upserts them into dhcp_leases.
// It is the manual ("sync now") entry point and ignores the server's auto-sync
// opt-out. Returns the number of leases processed.
func (s *Service) SyncTechnitiumLeases(ctx context.Context) (int, error) {
	n, _, err := s.syncTechnitiumLeases(ctx, true)
	return n, err
}

// syncTechnitiumLeases does the sync. When force is false and the Technitium
// server record has auto-sync turned off, it does nothing and reports skipped.
// The outcome (time, count, error) is recorded on the server row.
func (s *Service) syncTechnitiumLeases(ctx context.Context, force bool) (count int, skipped bool, err error) {
	client := s.technitiumDHCPClient(ctx)
	if client == nil {
		return 0, false, fmt.Errorf("technitium is not configured")
	}

	apiURL, _ := s.Config.GetCtx(ctx, "technitium_url")
	server, err := s.repository.GetOrCreateTechnitiumDHCPServer(ctx, apiURL)
	if err != nil {
		return 0, false, fmt.Errorf("getting DHCP server record: %w", err)
	}
	if !force && !server.AutoSync {
		return 0, true, nil
	}

	defer func() {
		msg := ""
		if err != nil {
			msg = err.Error()
		}
		_ = s.repository.RecordDHCPServerSync(ctx, server.ID, count, msg)
	}()

	scopes, err := client.ListDHCPScopes(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("listing scopes: %w", err)
	}

	var failed []string
	for _, scope := range scopes {
		leases, lerr := client.ListDHCPLeases(ctx, scope.Name)
		if lerr != nil {
			// One bad scope shouldn't abort the whole sync; report it at the end.
			failed = append(failed, scope.Name)
			continue
		}
		for _, lease := range leases {
			params := s.technitiumLeaseParams(ctx, server, lease)
			if _, uerr := s.repository.UpsertDHCPLease(ctx, params); uerr != nil {
				continue
			}
			count++
		}
	}
	if len(failed) > 0 {
		err = fmt.Errorf("failed to read leases for scope(s): %s", strings.Join(failed, ", "))
	}
	return count, false, err
}

// technitiumLeaseParams maps a Technitium lease onto a dhcp_leases row,
// linking it to a matching subnet and IP record where one exists. The lease's
// VLAN comes from its subnet, falling back to the server's VLAN, so mixed
// environments (some VLANs on Technitium, some not) work without configuration.
func (s *Service) technitiumLeaseParams(ctx context.Context, server *models.DHCPServer, lease technitium.DHCPLease) *repository.DHCPLeaseParams {
	state := "active"
	if lease.LeaseType == "Reserved" {
		state = "reserved"
	} else if !lease.LeaseExpires.IsZero() && lease.LeaseExpires.Before(time.Now()) {
		state = "expired"
	}

	var subnetID, ipID, vlanID *int64
	if sub, err := s.repository.FindSubnetContaining(ctx, lease.IPAddress); err == nil {
		id := sub.ID
		subnetID = &id
		vlanID = sub.VLANID
	}
	if vlanID == nil {
		vlanID = server.VLANID
	}
	if ip, err := s.repository.FindIPByAddress(ctx, lease.IPAddress); err == nil {
		id := ip.ID
		ipID = &id
	}

	var startsAt, endsAt *string
	if !lease.LeaseObtained.IsZero() {
		v := lease.LeaseObtained.UTC().Format(time.RFC3339)
		startsAt = &v
	}
	if !lease.LeaseExpires.IsZero() {
		v := lease.LeaseExpires.UTC().Format(time.RFC3339)
		endsAt = &v
	}

	return &repository.DHCPLeaseParams{
		ServerID:   server.ID,
		IPAddress:  lease.IPAddress,
		MACAddress: lease.HardwareAddress,
		Hostname:   lease.HostName,
		SubnetID:   subnetID,
		IPID:       ipID,
		VLANID:     vlanID,
		StartsAt:   startsAt,
		EndsAt:     endsAt,
		State:      state,
	}
}

// defaultTechnitiumSyncMinutes is the scheduled lease-sync interval when
// technitium_dhcp_sync_interval_minutes is unset. A value of 0 disables it.
const defaultTechnitiumSyncMinutes = 15

// technitiumSyncInterval reads the configured sync interval (0 = disabled).
func (s *Service) technitiumSyncInterval(ctx context.Context) time.Duration {
	v, _ := s.Config.GetCtx(ctx, "technitium_dhcp_sync_interval_minutes")
	mins := defaultTechnitiumSyncMinutes
	if v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			n = defaultTechnitiumSyncMinutes
		}
		mins = n
	}
	return time.Duration(mins) * time.Minute
}

// StartTechnitiumDHCPSync launches a background job that periodically syncs
// Technitium leases. It is a no-op while Technitium is unconfigured, the
// interval is 0, or the server record has auto-sync turned off. The interval is
// re-read each minute so admin changes apply without a restart.
func (s *Service) StartTechnitiumDHCPSync(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		var last time.Time
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				interval := s.technitiumSyncInterval(ctx)
				if interval == 0 || time.Since(last) < interval {
					continue
				}
				if s.technitiumDHCPClient(ctx) == nil {
					continue
				}
				last = time.Now()
				if _, _, err := s.syncTechnitiumLeases(ctx, false); err != nil {
					log.Printf("technitium dhcp sync: %v", err)
				}
			}
		}
	}()
}

// ImportTechnitiumScope imports a Technitium DHCP scope as a new subnet in the given network.
// It links the scope name to the new subnet and syncs its leases.
func (s *Service) ImportTechnitiumScope(ctx context.Context, scopeName string, networkID int64) (*models.Subnet, error) {
	client := s.technitiumDHCPClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("technitium is not configured")
	}

	scopes, err := client.ListDHCPScopes(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing scopes: %w", err)
	}
	var scope *technitium.DHCPScope
	for i := range scopes {
		if scopes[i].Name == scopeName {
			scope = &scopes[i]
			break
		}
	}
	if scope == nil {
		return nil, fmt.Errorf("scope %q not found", scopeName)
	}

	networkAddr, prefixLen, err := scopeToCIDR(scope.StartingAddress, scope.SubnetMask)
	if err != nil {
		return nil, fmt.Errorf("computing CIDR for scope %q: %w", scopeName, err)
	}

	gateway := scope.RouterAddress
	var gw *string
	if gateway != "" {
		gw = &gateway
	}

	subnet, err := s.Ops.IPAM.CreateSubnet(ctx, networkID, networkAddr, prefixLen, scope.Name, gw, false, false, nil, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("creating subnet: %w", err)
	}

	// Link the scope name to the new subnet.
	if err := s.repository.SetSubnetTechnitiumScope(ctx, subnet.ID, scopeName); err != nil {
		return subnet, nil // non-fatal
	}
	subnet.TechnitiumScopeName = scopeName
	return subnet, nil
}

// PushDHCPReservation creates a DHCP reservation in Technitium for the given IP address.
// The IP must have a MAC address and its subnet must have a Technitium scope name configured.
func (s *Service) PushDHCPReservation(ctx context.Context, ipID int64) error {
	client := s.technitiumDHCPClient(ctx)
	if client == nil {
		return fmt.Errorf("technitium is not configured")
	}

	ip, err := s.repository.GetIPAddressByID(ctx, ipID)
	if err != nil {
		return fmt.Errorf("IP address not found")
	}
	if ip.MACAddress == nil || *ip.MACAddress == "" {
		return fmt.Errorf("IP address has no MAC address configured")
	}

	subnet, err := s.repository.GetSubnetByID(ctx, ip.SubnetID)
	if err != nil {
		return fmt.Errorf("subnet not found")
	}
	if subnet.TechnitiumScopeName == "" {
		return fmt.Errorf("subnet has no Technitium DHCP scope configured")
	}

	return client.AddDHCPReservation(ctx, subnet.TechnitiumScopeName, ip.Address, *ip.MACAddress, ip.Hostname)
}

// RemoveDHCPReservation deletes a DHCP reservation from Technitium for the given IP address.
func (s *Service) RemoveDHCPReservation(ctx context.Context, ipID int64) error {
	client := s.technitiumDHCPClient(ctx)
	if client == nil {
		return fmt.Errorf("technitium is not configured")
	}

	ip, err := s.repository.GetIPAddressByID(ctx, ipID)
	if err != nil {
		return fmt.Errorf("IP address not found")
	}

	subnet, err := s.repository.GetSubnetByID(ctx, ip.SubnetID)
	if err != nil {
		return fmt.Errorf("subnet not found")
	}
	if subnet.TechnitiumScopeName == "" {
		return fmt.Errorf("subnet has no Technitium DHCP scope configured")
	}

	return client.RemoveDHCPReservation(ctx, subnet.TechnitiumScopeName, ip.Address)
}

// scopeToCIDR converts a starting address and subnet mask to a network address and prefix length.
func scopeToCIDR(startingAddress, subnetMask string) (string, int, error) {
	ip := net.ParseIP(startingAddress)
	if ip == nil {
		return "", 0, fmt.Errorf("invalid starting address %q", startingAddress)
	}
	mask := net.ParseIP(subnetMask)
	if mask == nil {
		return "", 0, fmt.Errorf("invalid subnet mask %q", subnetMask)
	}
	ip4 := ip.To4()
	mask4 := mask.To4()
	if ip4 == nil || mask4 == nil {
		return "", 0, fmt.Errorf("only IPv4 scopes are supported")
	}
	networkIP := make(net.IP, 4)
	for i := range 4 {
		networkIP[i] = ip4[i] & mask4[i]
	}
	ones := bits(binary.BigEndian.Uint32(mask4))
	return networkIP.String(), ones, nil
}

// bits counts the number of set bits in a uint32 (population count).
func bits(n uint32) int {
	count := 0
	for n > 0 {
		count += int(n & 1)
		n >>= 1
	}
	return count
}
