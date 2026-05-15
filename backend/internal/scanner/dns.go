package scanner

import (
	"context"
	"net"
	"strings"
	"time"
)

// DNSResult holds the outcome of a reverse DNS lookup for a single IP.
type DNSResult struct {
	PTR            string
	FwdRevMismatch bool
}

// ResolveHostname performs a PTR lookup for ip, then verifies the result
// with a forward lookup. Returns an empty DNSResult if no PTR exists or
// the lookup times out.
func ResolveHostname(ip string, timeout time.Duration) DNSResult {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resolver := net.DefaultResolver

	names, err := resolver.LookupAddr(ctx, ip)
	if err != nil || len(names) == 0 {
		return DNSResult{}
	}

	ptr := strings.TrimSuffix(names[0], ".")

	// Forward lookup to verify the PTR resolves back to the same IP.
	addrs, err := resolver.LookupHost(ctx, ptr)
	if err != nil {
		return DNSResult{PTR: ptr, FwdRevMismatch: true}
	}
	for _, addr := range addrs {
		if addr == ip {
			return DNSResult{PTR: ptr, FwdRevMismatch: false}
		}
	}
	return DNSResult{PTR: ptr, FwdRevMismatch: true}
}
