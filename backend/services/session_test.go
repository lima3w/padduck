package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// parseDeviceName
// ---------------------------------------------------------------------------

func TestParseDeviceName(t *testing.T) {
	cases := []struct {
		name      string
		userAgent string
		want      string
	}{
		// Empty string
		{
			name:      "empty string returns Unknown Device",
			userAgent: "",
			want:      "Unknown Device",
		},
		// Known device patterns
		{
			name:      "iPhone UA returns iPhone",
			userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15",
			want:      "iPhone",
		},
		{
			name:      "iPad UA returns iPad",
			userAgent: "Mozilla/5.0 (iPad; CPU OS 16_0 like Mac OS X) AppleWebKit/605.1.15",
			want:      "iPad",
		},
		{
			name:      "Android Mobile returns Android Phone",
			userAgent: "Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 Mobile Safari/537.36",
			want:      "Android Phone",
		},
		{
			name:      "Android without Mobile returns Android Tablet",
			userAgent: "Mozilla/5.0 (Linux; Android 13; Nexus 10) AppleWebKit/537.36 Safari/537.36",
			want:      "Android Tablet",
		},
		{
			name:      "Macintosh UA returns Mac",
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 13_0) AppleWebKit/537.36 Chrome/115",
			want:      "Mac",
		},
		{
			name:      "Windows UA returns Windows PC",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/115",
			want:      "Windows PC",
		},
		{
			name:      "Linux UA returns Linux",
			userAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/115",
			want:      "Linux",
		},
		{
			name:      "curl UA returns curl",
			userAgent: "curl/7.88.1",
			want:      "curl",
		},
		// Fallback: unknown UA shorter than 50 chars is returned as-is
		{
			name:      "unknown short UA returned as-is",
			userAgent: "MyCustomClient/1.0",
			want:      "MyCustomClient/1.0",
		},
		// Fallback: unknown UA exactly 50 chars is returned as-is (boundary)
		{
			name:      "unknown UA exactly 50 chars returned as-is",
			userAgent: "12345678901234567890123456789012345678901234567890", // 50 chars
			want:      "12345678901234567890123456789012345678901234567890",
		},
		// Fallback: unknown UA longer than 50 chars is truncated to 50 chars
		{
			name:      "unknown UA longer than 50 chars truncated to 50",
			userAgent: "123456789012345678901234567890123456789012345678901", // 51 chars
			want:      "12345678901234567890123456789012345678901234567890", // first 50
		},
		{
			name:      "unknown UA much longer than 50 chars truncated to 50",
			userAgent: "UnknownBrowser/99.0 (OperatingSystem; Architecture; Build/XYZ) AppleWebKit/537.36",
			want:      "UnknownBrowser/99.0 (OperatingSystem; Architecture",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseDeviceName(tc.userAgent)
			assert.Equal(t, tc.want, got)
		})
	}
}

// Verify the truncation boundary: result is always <= 50 chars for long unknown UAs.
func TestParseDeviceName_TruncationLength(t *testing.T) {
	longUA := "SomeCompletelyUnknownBrowserProduct/1.0 (compatible; extra tokens; more tokens; even more)"
	result := parseDeviceName(longUA)
	assert.LessOrEqual(t, len(result), 50,
		"parseDeviceName should return at most 50 chars for an unknown long UA")
	assert.Equal(t, longUA[:50], result)
}

// ---------------------------------------------------------------------------
// IsSessionExpired
// ---------------------------------------------------------------------------

func TestIsSessionExpired(t *testing.T) {
	svc := &IdentityService{} // zero-value; IsSessionExpired does not use any IdentityService fields

	t.Run("nil lastLoginAt returns false", func(t *testing.T) {
		assert.False(t, svc.IsSessionExpired(nil))
	})

	t.Run("timestamp 25 hours ago is expired", func(t *testing.T) {
		past := time.Now().Add(-25 * time.Hour)
		assert.True(t, svc.IsSessionExpired(&past))
	})

	t.Run("timestamp 23 hours ago is not expired", func(t *testing.T) {
		past := time.Now().Add(-23 * time.Hour)
		assert.False(t, svc.IsSessionExpired(&past))
	})

	// Use a small margin (2 seconds) to avoid flakiness at the exact boundary.
	t.Run("timestamp just before 24-hour boundary is not expired", func(t *testing.T) {
		justBefore := time.Now().Add(-DefaultSessionTimeout + 2*time.Second)
		assert.False(t, svc.IsSessionExpired(&justBefore))
	})

	t.Run("timestamp just after 24-hour boundary is expired", func(t *testing.T) {
		justAfter := time.Now().Add(-DefaultSessionTimeout - 2*time.Second)
		assert.True(t, svc.IsSessionExpired(&justAfter))
	})

	t.Run("future timestamp is not expired", func(t *testing.T) {
		future := time.Now().Add(1 * time.Hour)
		assert.False(t, svc.IsSessionExpired(&future))
	})

	t.Run("exactly now is not expired", func(t *testing.T) {
		now := time.Now()
		assert.False(t, svc.IsSessionExpired(&now))
	})
}

// ---------------------------------------------------------------------------
// DefaultSessionTimeout constant
// ---------------------------------------------------------------------------

func TestDefaultSessionTimeout_Value(t *testing.T) {
	assert.Equal(t, 24*time.Hour, DefaultSessionTimeout,
		"DefaultSessionTimeout should be 24 hours")
}
