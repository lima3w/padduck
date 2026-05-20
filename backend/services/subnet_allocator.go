package services

import (
	"encoding/binary"
	"fmt"
	"net"

	"padduck/models"
)

// findFirstFreeBlock finds the first network address that fits /<prefixLen>
// within the given parent network, avoiding any existing subnets.
func findFirstFreeBlock(parentAddr string, parentPrefix, prefixLen int, existing []*models.Subnet) (string, error) {
	if prefixLen < parentPrefix {
		return "", fmt.Errorf("requested prefix /%d is larger than parent /%d", prefixLen, parentPrefix)
	}
	if prefixLen > 32 {
		return "", fmt.Errorf("invalid prefix length %d", prefixLen)
	}

	parentCIDR := fmt.Sprintf("%s/%d", parentAddr, parentPrefix)
	_, parentNet, err := net.ParseCIDR(parentCIDR)
	if err != nil {
		return "", fmt.Errorf("invalid parent CIDR %s: %w", parentCIDR, err)
	}

	// Block size in number of IPs
	blockSize := uint32(1) << uint(32-prefixLen)
	parentStart := ipToUint32(parentNet.IP.To4())
	parentEnd := parentStart + (uint32(1) << uint(32-parentPrefix))

	// Collect occupied ranges from existing subnets
	type occupied struct {
		start, end uint32
	}
	var occupied_ranges []occupied
	for _, sub := range existing {
		cidr := fmt.Sprintf("%s/%d", sub.NetworkAddress, sub.PrefixLength)
		_, sn, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if !parentNet.Contains(sn.IP) {
			continue
		}
		s := ipToUint32(sn.IP.To4())
		sz := uint32(1) << uint(32-sub.PrefixLength)
		occupied_ranges = append(occupied_ranges, struct{ start, end uint32 }{s, s + sz})
	}

	// Try each aligned block within the parent
	for candidate := parentStart; candidate+blockSize <= parentEnd; candidate += blockSize {
		// Check alignment
		if candidate%blockSize != 0 {
			candidate = (candidate/blockSize+1)*blockSize - blockSize
			continue
		}
		// Check overlap with occupied
		overlap := false
		for _, o := range occupied_ranges {
			if candidate < o.end && candidate+blockSize > o.start {
				overlap = true
				break
			}
		}
		if !overlap {
			return uint32ToIP(candidate).String(), nil
		}
	}
	return "", fmt.Errorf("no free /%d block available in %s", prefixLen, parentCIDR)
}

func ipToUint32(ip net.IP) uint32 {
	if len(ip) == 4 {
		return binary.BigEndian.Uint32(ip)
	}
	return 0
}

func uint32ToIP(n uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, n)
	return ip
}
