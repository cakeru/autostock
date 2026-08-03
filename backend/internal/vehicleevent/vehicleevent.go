// Package vehicleevent records vehicle_service_events — the log invoice.Create
// writes to whenever a sale includes a tire or oil-flagged product, so the
// vehicle profile's due-for-service estimate has real history to project
// from instead of guessing off a single most-recent sale. Mirrors the
// batch/telegram packages' style: plain functions taking a tx, so any service
// can log without depending on a vehicle-service instance.
package vehicleevent

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// LogEvent records one detected (or manually entered) service event.
// Idempotent per (vehicle, event_type, invoice): a retry or double-processing
// of the same invoice silently no-ops via the partial unique index, so voiding
// and re-issuing a mis-keyed invoice can't double-count a service. The rating
// that applied at the time is recorded: lifeKm (distance), lifeDays, lifeMonths.
func LogEvent(ctx context.Context, tx pgx.Tx, branchID, vehicleID int64, eventType string, mileage *int, occurredAt time.Time, invoiceID, serviceJobID *int64, productName string, lifeKm, lifeDays, lifeMonths *int, createdBy *int64) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO vehicle_service_events (branch_id, vehicle_id, event_type, mileage, occurred_at, invoice_id, service_job_id, product_name, life_km, life_days, life_months, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), $9, $10, $11, $12)
		ON CONFLICT (vehicle_id, event_type, invoice_id, product_name) WHERE invoice_id IS NOT NULL DO NOTHING`,
		branchID, vehicleID, eventType, mileage, occurredAt, invoiceID, serviceJobID, productName, lifeKm, lifeDays, lifeMonths, createdBy)
	if err != nil {
		return fmt.Errorf("log vehicle service event: %w", err)
	}
	return nil
}
