package services

import (
	"fmt"
	"net"
)

// ParseIP parses an IP address string to net.IP
func ParseIP(address string) (net.IP, error) {
	ip := net.ParseIP(address)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", address)
	}
	return ip, nil
}

// IsAvailable checks if a specific IP is in the available list
func IsIPAvailable(targetAddress string, availableIPs []string) bool {
	for _, ip := range availableIPs {
		if ip == targetAddress {
			return true
		}
	}
	return false
}

// GetFirstAvailableIP returns the first IP from a list of available IP addresses
// assumes the list is already ordered by address (from ORDER BY address in query)
func GetFirstAvailableIP(availableAddresses []string) (string, error) {
	if len(availableAddresses) == 0 {
		return "", fmt.Errorf("no available IP addresses")
	}
	return availableAddresses[0], nil
}

// IPToInt converts an IP address to a 32-bit integer (IPv4 only)
func IPToInt(ip net.IP) (uint32, error) {
	if len(ip) != 4 {
		return 0, fmt.Errorf("only IPv4 addresses supported for allocation")
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3]), nil
}

// IntToIP converts a 32-bit integer back to an IP address (IPv4 only)
func IntToIP(num uint32) net.IP {
	return net.IPv4(byte(num>>24), byte(num>>16), byte(num>>8), byte(num))
}

// NextIP returns the next IP address after the given IP
func NextIP(ip net.IP) (net.IP, error) {
	intVal, err := IPToInt(ip)
	if err != nil {
		return nil, err
	}
	return IntToIP(intVal + 1), nil
}
