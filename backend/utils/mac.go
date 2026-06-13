package utils

import (
	"fmt"
	"strings"
)

// NormalizeMAC accepts a MAC address in any common notation (colon, dash, dot separators,
// or unseparated) and returns it in lowercase colon-separated form (aa:bb:cc:dd:ee:ff).
// Returns an error if the input is not a valid 48-bit MAC address.
func NormalizeMAC(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", nil
	}

	// Strip all recognized separators and whitespace, then validate.
	stripped := strings.NewReplacer(":", "", "-", "", ".", "").Replace(s)
	stripped = strings.ToLower(stripped)

	if len(stripped) != 12 {
		return "", fmt.Errorf("invalid MAC address: %q", s)
	}
	for _, c := range stripped {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return "", fmt.Errorf("invalid MAC address: %q", s)
		}
	}

	return fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		stripped[0:2], stripped[2:4], stripped[4:6],
		stripped[6:8], stripped[8:10], stripped[10:12],
	), nil
}
