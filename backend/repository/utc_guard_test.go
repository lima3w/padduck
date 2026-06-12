package repository

// Guard against the TIMESTAMP/UTC skew bug class found four times during the
// v1.31.25 test campaign (broken MFA login, inoperative account lockout,
// skewed token expiry — all on non-UTC hosts): the schema uses TIMESTAMP
// without time zone, and pgx stores a time.Time's wall-clock digits as-is
// while reading them back as UTC. Any local-time value written to the
// database or used as a SQL parameter is therefore wrong by the host's UTC
// offset.
//
// Convention: in this package, every time.Now() must immediately be
// converted with .UTC(). Repository code always talks to the database, so
// there is no legitimate local-time use here. Service-layer code that passes
// times into repository calls must apply the same rule at the call site.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestRepositoryTimeNowIsAlwaysUTC(t *testing.T) {
	bare := regexp.MustCompile(`time\.Now\(\)`)
	utc := regexp.MustCompile(`time\.Now\(\)\.UTC\(\)`)

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(filepath.Clean(name))
		if err != nil {
			t.Fatal(err)
		}
		for i, line := range strings.Split(string(src), "\n") {
			if bare.MatchString(line) && !utc.MatchString(line) {
				t.Errorf("%s:%d: bare time.Now() in repository code — use time.Now().UTC() "+
					"(TIMESTAMP columns store wall-clock digits and read back as UTC; "+
					"local time skews by the host's UTC offset)", name, i+1)
			}
		}
	}
}
