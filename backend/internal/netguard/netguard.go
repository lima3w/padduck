package netguard

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

const defaultTimeout = 10 * time.Second

// NewHTTPClient returns an HTTP client that refuses redirects and outbound
// connections to loopback, link-local, private, multicast, or unspecified IPs.
func NewHTTPClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	dialer := &net.Dialer{Timeout: timeout}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		addrs, err := safeHostAddrs(ctx, host)
		if err != nil {
			return nil, err
		}
		var lastErr error
		for _, addr := range addrs {
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(addr.String(), port))
			if err == nil {
				return conn, nil
			}
			lastErr = err
		}
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, fmt.Errorf("host could not be dialed")
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return ValidateURL(req.Context(), req.URL.String())
		},
	}
}

// ValidateURL ensures the URL is absolute, uses HTTP(S), has no embedded
// credentials, and does not resolve to unsafe addresses.
func ValidateURL(ctx context.Context, rawURL string) error {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("url must be absolute")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("url must use http or https")
	}
	if u.User != nil {
		return fmt.Errorf("url must not include credentials")
	}
	return ValidateHost(ctx, u.Hostname())
}

// ValidateHost fails closed when a host cannot be resolved or any resolved
// address is unsafe for server-side outbound requests.
func ValidateHost(ctx context.Context, host string) error {
	_, err := safeHostAddrs(ctx, host)
	return err
}

func safeHostAddrs(ctx context.Context, host string) ([]netip.Addr, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil, fmt.Errorf("url host is required")
	}
	if ip, err := netip.ParseAddr(host); err == nil {
		if !IsSafeOutboundAddr(ip) {
			return nil, fmt.Errorf("host resolves to a private or reserved address")
		}
		return []netip.Addr{ip}, nil
	}

	resolver := net.DefaultResolver
	addrs, err := resolver.LookupNetIP(ctx, "ip", host)
	if err != nil || len(addrs) == 0 {
		return nil, fmt.Errorf("host could not be resolved")
	}
	for _, addr := range addrs {
		if !IsSafeOutboundAddr(addr) {
			return nil, fmt.Errorf("host resolves to a private or reserved address")
		}
	}
	return addrs, nil
}

// IsSafeOutboundAddr returns true only for public unicast addresses.
func IsSafeOutboundAddr(addr netip.Addr) bool {
	if !addr.IsValid() {
		return false
	}
	if addr.Is4In6() {
		addr = addr.Unmap()
	}
	return addr.IsGlobalUnicast() &&
		!addr.IsPrivate() &&
		!addr.IsLoopback() &&
		!addr.IsLinkLocalUnicast() &&
		!addr.IsLinkLocalMulticast() &&
		!addr.IsMulticast() &&
		!addr.IsUnspecified()
}
