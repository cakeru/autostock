// Package batch holds intake-lot logic shared by the inventory and invoice
// services: creating batches, recording ledger movements, and consuming stock
// oldest-first (FIFO) so outflows are traceable to the batch they came from.
package batch

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Create inserts a new intake batch and returns its id. invoiceNumber and
// invoiceImage record the supplier invoice this receive came with (optional).
func Create(ctx context.Context, tx pgx.Tx, branchID, productID int64, qty float64, unitCost float64, supplierID *int64, amountPaid float64, supplier, dot, notes string, receivedBy *int64, invoiceNumber, invoiceImage string) (int64, error) {
	var id int64
	err := tx.QueryRow(ctx, `
		INSERT INTO batches (branch_id, product_id, supplier_id, supplier, dot_code, unit_cost,
		                     quantity_received, quantity_remaining, amount_paid, notes, received_by,
		                     invoice_number, invoice_image)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), $6, $7, $7, $8, NULLIF($9, ''), $10,
		        NULLIF($11, ''), NULLIF($12, ''))
		RETURNING id`,
		branchID, productID, supplierID, supplier, dot, unitCost, qty, amountPaid, notes, receivedBy, invoiceNumber, invoiceImage).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create batch: %w", err)
	}
	return id, nil
}

// RecordMovement writes one ledger row. refID/batchID/recordedBy are nullable.
func RecordMovement(ctx context.Context, tx pgx.Tx, branchID, productID int64, change float64, reason, refType string, refID, batchID, recordedBy *int64) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO stock_movements (branch_id, product_id, quantity_change, reason,
		                             reference_type, reference_id, batch_id, recorded_by)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7, $8)`,
		branchID, productID, change, reason, refType, refID, batchID, recordedBy)
	if err != nil {
		return fmt.Errorf("record movement: %w", err)
	}
	return nil
}

// ConsumeFIFO deducts qty from the product's batches oldest-first, writing a
// movement per batch consumed. Any shortfall beyond what batches hold (possible
// only from pre-batch legacy data) is recorded as one unbatched movement so the
// ledger still balances and the product-level stock stays authoritative.
func ConsumeFIFO(ctx context.Context, tx pgx.Tx, branchID, productID int64, qty float64, reason, refType string, refID, recordedBy *int64) error {
	rows, err := tx.Query(ctx, `
		SELECT id, quantity_remaining FROM batches
		WHERE product_id = $1 AND branch_id = $2 AND quantity_remaining > 0
		ORDER BY received_at, id
		FOR UPDATE`, productID, branchID)
	if err != nil {
		return fmt.Errorf("lock batches: %w", err)
	}
	type lot struct {
		id  int64
		rem float64
	}
	var lots []lot
	for rows.Next() {
		var l lot
		if err := rows.Scan(&l.id, &l.rem); err != nil {
			rows.Close()
			return fmt.Errorf("scan batch: %w", err)
		}
		lots = append(lots, l)
	}
	rows.Close()

	remaining := qty
	for _, l := range lots {
		if remaining <= 0 {
			break
		}
		take := l.rem
		if take > remaining {
			take = remaining
		}
		if _, err := tx.Exec(ctx, `UPDATE batches SET quantity_remaining = quantity_remaining - $1 WHERE id = $2`, take, l.id); err != nil {
			return fmt.Errorf("decrement batch: %w", err)
		}
		bid := l.id
		if err := RecordMovement(ctx, tx, branchID, productID, -take, reason, refType, refID, &bid, recordedBy); err != nil {
			return err
		}
		remaining -= take
	}
	if remaining > 0 {
		if err := RecordMovement(ctx, tx, branchID, productID, -remaining, reason, refType, refID, nil, recordedBy); err != nil {
			return err
		}
	}
	return nil
}
