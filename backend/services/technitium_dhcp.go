package services

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
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
// Returns the number of leases processed.
func (s *Service) SyncTechnitiumLeases(ctx context.Context) (int, error) {
	client := s.technitiumDHCPClient(ctx)
	if client == nil {
		return 0, fmt.Errorf("technitium is not configured")
	}

	apiURL, _ := s.Config.GetCtx(ctx, "technitium_url")
	server, err := s.repository.GetOrCreateTechnitiumDHCPServer(ctx, apiURL)
	if err != nil {
		return 0, fmt.Errorf("getting DHCP server record: %w", err)
	}

	scopes, err := client.ListDHCPScopes(ctx)
	if err != nil {
		return 0, fmt.Errorf("listing scopes: %w", err)
	}

	total := 0
	for _, scope := range scopes {
		leases, err := client.ListDHCPLeases(ctx, scope.Name)
		if err != nil {
			// Log and continue — one bad scope shouldn't abort the whole sync.
			continue
		}
		for _, lease := range leases {
			state := "active"
			if lease.LeaseType == "Reserved" {
				state = "reserved"
			} else if !lease.LeaseExpires.IsZero() && lease.LeaseExpires.Before(time.Now()) {
				state = "expired"
			}

			// Best-effort subnet and IP linkage.
			var subnetID *int64
			var ipID *int64
			if sub, err := s.repository.FindSubnetContaining(ctx, lease.IPAddress); err == nil {
				id := sub.ID
				subnetID = &id
			}
			if ip, err := s.repository.FindIPByAddress(ctx, lease.IPAddress); err == nil {
				id := ip.ID
				ipID = &id
			}

			var startsAt, endsAt *string
			if !lease.LeaseObtained.IsZero() {
				s := lease.LeaseObtained.UTC().Format(time.RFC3339)
				startsAt = &s
			}
			if !lease.LeaseExpires.IsZero() {
				e := lease.LeaseExpires.UTC().Format(time.RFC3339)
				endsAt = &e
			}

			_, err := s.repository.UpsertDHCPLease(ctx, &repository.DHCPLeaseParams{
				ServerID:   server.ID,
				IPAddress:  lease.IPAddress,
				MACAddress: lease.HardwareAddress,
				Hostname:   lease.HostName,
				SubnetID:   subnetID,
				IPID:       ipID,
				StartsAt:   startsAt,
				EndsAt:     endsAt,
				State:      state,
			})
			if err != nil {
				continue
			}
			total++
		}
	}
	return total, nil
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

	subnet, err := s.CreateSubnet(ctx, networkID, networkAddr, prefixLen, scope.Name, gw, false, false, nil, nil, nil, nil)
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
