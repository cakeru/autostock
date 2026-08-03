package service

import (
	"context"
	"testing"

	"github.com/cakeru/autostock/internal/testutil"
)

// Regression: ListVehicles used to 500 on customers with vehicles because the
// SELECT included distance_unit but the Scan didn't read it (column/scan
// mismatch). Keeps the whole customer row loadable.
func TestListVehiclesScansDistanceUnit(t *testing.T) {
	pool := testutil.ConnectDB(t)
	branchID := testutil.SeedBranch(t, pool)
	customerID := testutil.SeedCustomer(t, pool, branchID)
	testutil.SeedVehicle(t, pool, customerID)

	s := NewService(pool)
	vehicles, err := s.ListVehicles(context.Background(), customerID)
	if err != nil {
		t.Fatalf("ListVehicles returned error for a customer with a vehicle: %v", err)
	}
	if len(vehicles) != 1 {
		t.Fatalf("expected 1 vehicle, got %d", len(vehicles))
	}
	if vehicles[0].DistanceUnit != "km" {
		t.Errorf("expected distance_unit 'km', got %q", vehicles[0].DistanceUnit)
	}
}
