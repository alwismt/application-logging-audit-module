package common

import (
	"strings"
	"time"
)

// ParseQueryTime parses RFC3339 or date-only (YYYY-MM-DD) query values.
// Date-only "from" uses start of day UTC; "to" uses end of day UTC.
func ParseQueryTime(value string, endOfDay bool) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	if t, err := time.Parse(time.RFC3339, value); err == nil {
		utc := t.UTC()
		return &utc, nil
	}

	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, err
	}
	if endOfDay {
		t = t.Add(24*time.Hour - time.Nanosecond)
	}
	utc := t.UTC()
	return &utc, nil
}
