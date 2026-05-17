package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Exported error variables
// ---------------------------------------------------------------------------

func TestErrAccountLocked_NonNil(t *testing.T) {
	assert.NotNil(t, ErrAccountLocked, "ErrAccountLocked must not be nil")
}

func TestErrAccountLocked_HasMeaningfulMessage(t *testing.T) {
	assert.NotEmpty(t, ErrAccountLocked.Error(),
		"ErrAccountLocked should have a non-empty error message")
	assert.Contains(t, ErrAccountLocked.Error(), "locked",
		"ErrAccountLocked message should mention 'locked'")
}

func TestErrInvalidUnlockToken_NonNil(t *testing.T) {
	assert.NotNil(t, ErrInvalidUnlockToken, "ErrInvalidUnlockToken must not be nil")
}

func TestErrInvalidUnlockToken_HasMeaningfulMessage(t *testing.T) {
	assert.NotEmpty(t, ErrInvalidUnlockToken.Error(),
		"ErrInvalidUnlockToken should have a non-empty error message")
	assert.Contains(t, ErrInvalidUnlockToken.Error(), "token",
		"ErrInvalidUnlockToken message should mention 'token'")
}

// ---------------------------------------------------------------------------
// lockoutDuration
// ---------------------------------------------------------------------------

func TestLockoutDuration(t *testing.T) {
	cases := []struct {
		name         string
		lockoutCount int
		want         time.Duration
	}{
		// lockoutCount <= 1 → 5 minutes
		{
			name:         "count 0 returns 5 minutes",
			lockoutCount: 0,
			want:         5 * time.Minute,
		},
		{
			name:         "count 1 returns 5 minutes",
			lockoutCount: 1,
			want:         5 * time.Minute,
		},
		// lockoutCount == 2 → 15 minutes
		{
			name:         "count 2 returns 15 minutes",
			lockoutCount: 2,
			want:         15 * time.Minute,
		},
		// lockoutCount == 3 → 1 hour
		{
			name:         "count 3 returns 1 hour",
			lockoutCount: 3,
			want:         1 * time.Hour,
		},
		// lockoutCount == 4 → 4 hours
		{
			name:         "count 4 returns 4 hours",
			lockoutCount: 4,
			want:         4 * time.Hour,
		},
		// lockoutCount == 5 → 24 hours
		{
			name:         "count 5 returns 24 hours",
			lockoutCount: 5,
			want:         24 * time.Hour,
		},
		// lockoutCount >= 6 → 7 days
		{
			name:         "count 6 returns 7 days",
			lockoutCount: 6,
			want:         7 * 24 * time.Hour,
		},
		{
			name:         "count 10 returns 7 days",
			lockoutCount: 10,
			want:         7 * 24 * time.Hour,
		},
		{
			name:         "large count returns 7 days",
			lockoutCount: 100,
			want:         7 * 24 * time.Hour,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := lockoutDuration(tc.lockoutCount)
			assert.Equal(t, tc.want, got,
				"lockoutDuration(%d) should be %v", tc.lockoutCount, tc.want)
		})
	}
}

// ---------------------------------------------------------------------------
// Unexported constants — tested indirectly via lockoutDuration and
// documented expected values.
//
// maxFailedAttempts = 5
// bruteForceWindow  = 15 * time.Minute
// notifRateLimit    = 1 * time.Hour
// ipFailureThreshold = 3
//
// Because these constants are unexported, we verify them directly within the
// same package (white-box testing). Go allows package-level access in _test.go
// files that belong to the same package.
// ---------------------------------------------------------------------------

func TestUnexportedConstants_Values(t *testing.T) {
	t.Run("maxFailedAttempts is 5", func(t *testing.T) {
		assert.Equal(t, 5, maxFailedAttempts)
	})

	t.Run("bruteForceWindow is 15 minutes", func(t *testing.T) {
		assert.Equal(t, 15*time.Minute, bruteForceWindow)
	})

	t.Run("notifRateLimit is 1 hour", func(t *testing.T) {
		assert.Equal(t, 1*time.Hour, notifRateLimit)
	})

	t.Run("ipFailureThreshold is 3", func(t *testing.T) {
		assert.Equal(t, 3, ipFailureThreshold)
	})
}

// ---------------------------------------------------------------------------
// lockoutDuration as an escalating-severity proxy for the threshold constants.
// Ensures the progression 5m → 15m → 1h → 4h → 24h → 7d is maintained.
// ---------------------------------------------------------------------------

func TestLockoutDuration_Escalation(t *testing.T) {
	first := lockoutDuration(1)
	second := lockoutDuration(2)
	third := lockoutDuration(3)
	fourth := lockoutDuration(4)
	fifth := lockoutDuration(5)
	sixth := lockoutDuration(6)

	assert.Less(t, int64(first), int64(second),
		"second lockout should be longer than the first")
	assert.Less(t, int64(second), int64(third),
		"third lockout should be longer than the second")
	assert.Less(t, int64(third), int64(fourth),
		"fourth lockout should be longer than the third")
	assert.Less(t, int64(fourth), int64(fifth),
		"fifth lockout should be longer than the fourth")
	assert.Less(t, int64(fifth), int64(sixth),
		"sixth lockout should be longer than the fifth")
	assert.Equal(t, lockoutDuration(6), lockoutDuration(10),
		"all lockout counts >= 6 should return the same maximum duration (7 days)")
	assert.Equal(t, lockoutDuration(6), lockoutDuration(100),
		"very large lockout counts should still return the 7-day maximum")
}
