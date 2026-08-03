package service

import (
	"testing"
	"time"
)

func TestValidateCron(t *testing.T) {
	valid := []string{"0 2 * * *", "*/5 * * * *", "0 0 * * 0", "15 10 * * 1-5"}
	for _, expr := range valid {
		if err := validateCron(expr); err != nil {
			t.Errorf("validateCron(%q) = %v, want nil", expr, err)
		}
	}
	invalid := []string{"", "not a cron", "60 * * * *", "0 2 * *", "0 24 * * *"}
	for _, expr := range invalid {
		if err := validateCron(expr); err == nil {
			t.Errorf("validateCron(%q) = nil, want error", expr)
		}
	}
}

func TestNextRun(t *testing.T) {
	base := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
	n := nextRun("0 2 * * *", base)
	if n == nil {
		t.Fatal("nextRun returned nil")
	}
	want := time.Date(2026, 8, 4, 2, 0, 0, 0, time.UTC)
	if !n.Equal(want) {
		t.Errorf("nextRun = %v, want %v", n, want)
	}

	hourly := nextRun("0 */6 * * *", base)
	if hourly == nil || !hourly.Equal(time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)) {
		t.Errorf("nextRun hourly = %v, want 12:00", hourly)
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Nightly backup":       "nightly-backup",
		"  Weekly  (Sunday)  ": "weekly-sunday",
		"!!!":                  "backup",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}
