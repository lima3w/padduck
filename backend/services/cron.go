package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// matchesCron reports whether t matches a simple 5-field cron expression
// (minute hour dom month dow). Only literal values and "*" wildcards are
// supported — no ranges, steps, or lists.
func matchesCron(cron string, t time.Time) bool {
	parts := strings.Fields(cron)
	if len(parts) != 5 {
		return false
	}
	checks := []struct {
		field string
		value int
	}{
		{parts[0], t.Minute()},
		{parts[1], t.Hour()},
		{parts[2], t.Day()},
		{parts[3], int(t.Month())},
		{parts[4], int(t.Weekday())},
	}
	for _, c := range checks {
		if c.field == "*" {
			continue
		}
		v, err := strconv.Atoi(c.field)
		if err != nil || v != c.value {
			return false
		}
	}
	return true
}

// validateCron returns an error if cron does not contain exactly 5 fields.
func validateCron(cron string) error {
	if len(strings.Fields(cron)) != 5 {
		return fmt.Errorf("cron must have 5 fields (min hour dom month dow), got %d", len(strings.Fields(cron)))
	}
	return nil
}
