package service

import "testing"

func TestComputePayCostSalaryFullMonth(t *testing.T) {
	// A range covering a whole calendar month pays the full monthly salary.
	salary, commission := computePayCost("salary", 500, 0, 0, 0, 0, 31, true)
	if salary != 500 {
		t.Errorf("full month: expected salary 500, got %v", salary)
	}
	if commission != 0 {
		t.Errorf("full month: expected commission 0, got %v", commission)
	}
}

func TestComputePayCostSalaryPartialMonth(t *testing.T) {
	// A partial range is prorated on a 30-day month basis.
	salary, _ := computePayCost("salary", 500, 0, 0, 0, 0, 25, false)
	if salary != 416.67 {
		t.Errorf("partial month: expected salary 416.67, got %v", salary)
	}
}

func TestComputePayCostHourly(t *testing.T) {
	salary, _ := computePayCost("hourly", 0, 8, 0, 0, 10, 0, false)
	if salary != 80 {
		t.Errorf("hourly: expected 80, got %v", salary)
	}
}

func TestComputePayCostCommission(t *testing.T) {
	_, commission := computePayCost("commission", 0, 0, 10, 2000, 0, 0, false)
	if commission != 200 {
		t.Errorf("commission: expected 200, got %v", commission)
	}
}

func TestComputePayCostHybrid(t *testing.T) {
	salary, commission := computePayCost("hybrid", 300, 0, 5, 1000, 0, 30, true)
	if salary != 300 {
		t.Errorf("hybrid full month: expected salary 300, got %v", salary)
	}
	if commission != 50 {
		t.Errorf("hybrid: expected commission 50, got %v", commission)
	}
}

func TestIsFullMonth(t *testing.T) {
	cases := []struct {
		from, to string
		want     bool
	}{
		{"2026-08-01", "2026-08-31", true},
		{"2026-02-01", "2026-02-28", true}, // Feb 2026 has 28 days
		{"2026-08-01", "2026-08-28", false},
		{"2026-07-15", "2026-08-15", false},
		{"2026-08-02", "2026-08-31", false},
		{"not-a-date", "2026-08-31", false},
	}
	for _, c := range cases {
		if got := isFullMonth(c.from, c.to); got != c.want {
			t.Errorf("isFullMonth(%q, %q) = %v, want %v", c.from, c.to, got, c.want)
		}
	}
}