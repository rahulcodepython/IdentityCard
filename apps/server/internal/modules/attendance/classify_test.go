package attendance

import (
	"testing"
	"time"
)

func TestClassify(t *testing.T) {
	date := "2026-03-01"
	scheduled := "09:00"

	cases := []struct {
		name string
		scan time.Time
		want string
	}{
		{"well before", mustParse(t, "2026-03-01T08:30:00Z"), "early"},
		{"just before, within grace", mustParse(t, "2026-03-01T08:55:00Z"), "on_time"},
		{"exactly on time", mustParse(t, "2026-03-01T09:00:00Z"), "on_time"},
		{"just after, within grace", mustParse(t, "2026-03-01T09:05:00Z"), "on_time"},
		{"well after", mustParse(t, "2026-03-01T09:45:00Z"), "late"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classify(tc.scan, date, scheduled)
			if got != tc.want {
				t.Errorf("classify(%v, %s, %s) = %s, want %s", tc.scan, date, scheduled, got, tc.want)
			}
		})
	}
}

func mustParse(t *testing.T, s string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("parse %s: %v", s, err)
	}
	return parsed
}
