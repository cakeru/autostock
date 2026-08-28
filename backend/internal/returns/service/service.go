package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/batch"
	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/returns/dto"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

func (s *Service) ListForInvoice(ctx context.Context, branchID, invoiceID int64) (*dto.InvoiceReturns, error) {
	out := &dto.InvoiceReturns{Returns: []dto.ReturnResponse{}, ReturnedByItem: map[string]float64{}}

	rows, err := s.pool.Query(ctx, `
		SELECT r.id, r.invoice_id, r.refund_amount, r.refund_method, COALESCE(r.reason, ''),
		       COALESCE(u.full_name, u.username, ''), r.created_at::text
		FROM returns r LEFT JOIN users u ON u.id = r.created_by
		WHERE r.invoice_id = $1 AND r.branch_id = $2
		ORDER BY r.created_at DESC`, invoiceID, branchID)
	if err != nil {
		return nil, fmt.Errorf("list returns: %w", err)
	}
	defer rows.Close()
	byID := map[int64]*dto.ReturnResponse{}
	for rows.Next() {
		var r dto.ReturnResponse
		if err := rows.Scan(&r.ID, &r.InvoiceID, &r.RefundAmount, &r.RefundMethod, &r.Reason, &r.CreatedByName, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan return: %w", err)
		}
		r.Items = []dto.ReturnItemResp{}
		out.Returns = append(out.Returns, r)
		byID[r.ID] = &out.Returns[len(out.Returns)-1]
	}

	itemRows, err := s.pool.Query(ctx, `
		SELECT ri.return_id, ri.invoice_item_id, ri.product_id, COALESCE(ri.description, ''),
		       ri.quantity, ri.unit_price, ri.total
		FROM return_items ri JOIN returns r ON r.id = ri.return_id
		WHERE r.invoice_id = $1 AND r.branch_id = $2`, invoiceID, branchID)
	if err != nil {
		return nil, fmt.Errorf("list return items: %w", err)
	}
	defer itemRows.Close()
	for itemRows.Next() {
		var retID int64
		var it dto.ReturnItemResp
		if err := itemRows.Scan(&retID, &it.InvoiceItemID, &it.ProductID, &it.Description, &it.Quantity, &it.UnitPrice, &it.Total); err != nil {
			return nil, fmt.Errorf("scan return item: %w", err)
		}
		if r, ok := byID[retID]; ok {
			r.Items = append(r.Items, it)
		}
		out.ReturnedByItem[strconv.FormatInt(it.InvoiceItemID, 10)] += it.Quantity
	}
	return out, nil
}

func (s *Service) Create(ctx context.Context, branchID, userID int64, req *dto.CreateReturnRequest) (*dto.ReturnResponse, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var customerID *int64
	var invStatus string
	var invTotal float64
	if err := tx.QueryRow(ctx,
		`SELECT customer_id, status, total_usd FROM invoices WHERE id = $1 AND branch_id = $2 FOR UPDATE`,
		req.InvoiceID, branchID).Scan(&customerID, &invStatus, &invTotal); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get invoice: %w", err)
	}
	if invStatus == "voided" {
		return nil, &domain.AppError{Code: "VOIDED", Message: "Cannot return items on a voided invoice", Status: 400}
	}

	type line struct {
		invoiceItemID int64
		productID     *int64
		description   string
		quantity      float64
		unitPrice     float64
		total         float64
	}
	var lines []line
	var refundTotal float64

	for _, in := range req.Items {
		var productID *int64
		var desc string
		var qtySold, unitPrice float64
		if err := tx.QueryRow(ctx, `
			SELECT product_id, description, quantity, unit_price_usd
			FROM invoice_items WHERE id = $1 AND invoice_id = $2`,
			in.InvoiceItemID, req.InvoiceID).Scan(&productID, &desc, &qtySold, &unitPrice); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, &domain.AppError{Code: "INVALID_ITEM", Message: "Item is not on this invoice", Status: 400}
			}
			return nil, fmt.Errorf("get invoice item: %w", err)
		}
		var already float64
		if err := tx.QueryRow(ctx, `SELECT COALESCE(SUM(quantity), 0) FROM return_items WHERE invoice_item_id = $1`, in.InvoiceItemID).Scan(&already); err != nil {
			return nil, fmt.Errorf("returned qty: %w", err)
		}
		if in.Quantity > qtySold-already+0.001 {
			return nil, &domain.AppError{Code: "OVER_RETURN", Message: "Cannot return more than was sold", Status: 400}
		}
		lineTotal := round2(in.Quantity * unitPrice)
		refundTotal += lineTotal
		lines = append(lines, line{in.InvoiceItemID, productID, desc, in.Quantity, unitPrice, lineTotal})
	}
	refundTotal = round2(refundTotal)

	var returnID int64
	if err := tx.QueryRow(ctx, `
		INSERT INTO returns (branch_id, invoice_id, customer_id, refund_amount, refund_method, reason, created_by)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7) RETURNING id`,
		branchID, req.InvoiceID, customerID, refundTotal, req.RefundMethod, req.Reason, userID).Scan(&returnID); err != nil {
		return nil, fmt.Errorf("create return: %w", err)
	}

	for _, l := range lines {
		if _, err := tx.Exec(ctx, `
			INSERT INTO return_items (return_id, invoice_item_id, product_id, description, quantity, unit_price, total)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			returnID, l.invoiceItemID, l.productID, l.description, l.quantity, l.unitPrice, l.total); err != nil {
			return nil, fmt.Errorf("add return item: %w", err)
		}
		// Restock product lines back into inventory as a costed batch.
		if l.productID != nil {
			qty := round2(l.quantity)
			if qty > 0 {
				var buyPrice float64
				_ = tx.QueryRow(ctx, `SELECT buy_price FROM products WHERE id = $1`, *l.productID).Scan(&buyPrice)
				if _, err := tx.Exec(ctx, `UPDATE products SET stock_quantity = stock_quantity + $1 WHERE id = $2`, qty, *l.productID); err != nil {
					return nil, fmt.Errorf("restock: %w", err)
				}
				batchID, err := batch.Create(ctx, tx, branchID, *l.productID, qty, buyPrice, nil, 0, "Customer return", "", "", &userID, "", "")
				if err != nil {
					return nil, err
				}
				if err := batch.RecordMovement(ctx, tx, branchID, *l.productID, qty, "return", "return", &returnID, &batchID, &userID); err != nil {
					return nil, err
				}
			}
		}
	}

	// Store-credit refunds become a held deposit the customer can spend later.
	if req.RefundMethod == "store_credit" && customerID != nil {
		if _, err := tx.Exec(ctx, `
			INSERT INTO deposits (branch_id, customer_id, amount, note, created_by, return_id)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			branchID, *customerID, refundTotal, "Store credit from return", userID, returnID); err != nil {
			return nil, fmt.Errorf("store credit: %w", err)
		}
	}

	// If everything has now been returned, flag the invoice as refunded.
	var totalReturned float64
	if err := tx.QueryRow(ctx, `SELECT COALESCE(SUM(refund_amount), 0) FROM returns WHERE invoice_id = $1`, req.InvoiceID).Scan(&totalReturned); err != nil {
		return nil, fmt.Errorf("sum returns: %w", err)
	}
	if totalReturned >= invTotal-0.001 {
		if _, err := tx.Exec(ctx, `UPDATE invoices SET payment_status = 'refunded', updated_at = NOW() WHERE id = $1`, req.InvoiceID); err != nil {
			return nil, fmt.Errorf("flag refunded: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	list, err := s.ListForInvoice(ctx, branchID, req.InvoiceID)
	if err != nil {
		return nil, err
	}
	for i := range list.Returns {
		if list.Returns[i].ID == returnID {
			return &list.Returns[i], nil
		}
	}
	return nil, nil
}

// Undo reverses a return: restores the invoice's refund state, removes the
// store-credit deposit it created, and takes the restocked inventory back out.
// It refuses when the returned stock has already been sold again (its batch
// was consumed) — reversing that would corrupt the FIFO ledger.
func (s *Service) Undo(ctx context.Context, branchID, userID, returnID int64) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var invoiceID int64
	if err := tx.QueryRow(ctx, `SELECT invoice_id FROM returns WHERE id = $1 AND branch_id = $2 FOR UPDATE`, returnID, branchID).
		Scan(&invoiceID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("get return: %w", err)
	}

	var invStatus string
	if err := tx.QueryRow(ctx, `SELECT status FROM invoices WHERE id = $1 FOR UPDATE`, invoiceID).Scan(&invStatus); err != nil {
		return fmt.Errorf("get invoice: %w", err)
	}
	if invStatus == "voided" {
		return &domain.AppError{Code: "VOIDED", Message: "Cannot undo a return on a voided invoice", Status: 400}
	}

	// Reverse the restock: only safe while the return's batch is untouched
	// (quantity_remaining == quantity_received). Otherwise the stock has
	// already been sold again and the ledger can't be cleanly unwound.
	rows, err := tx.Query(ctx, `
		SELECT id, product_id, quantity_change, batch_id FROM stock_movements
		WHERE reference_type = 'return' AND reference_id = $1 AND quantity_change > 0`, returnID)
	if err != nil {
		return fmt.Errorf("list return movements: %w", err)
	}
	type mv struct {
		id        int64
		productID int64
		qty       float64
		batchID   *int64
	}
	var movements []mv
	for rows.Next() {
		var m mv
		if err := rows.Scan(&m.id, &m.productID, &m.qty, &m.batchID); err != nil {
			rows.Close()
			return fmt.Errorf("scan movement: %w", err)
		}
		movements = append(movements, m)
	}
	rows.Close()

	for _, m := range movements {
		if m.batchID != nil {
			var received, remaining float64
			if err := tx.QueryRow(ctx, `SELECT quantity_received, quantity_remaining FROM batches WHERE id = $1 FOR UPDATE`, *m.batchID).
				Scan(&received, &remaining); err != nil {
				return fmt.Errorf("get batch: %w", err)
			}
			if remaining != received {
				return &domain.AppError{
					Code:    "RETURN_STOCK_SOLD",
					Message: "This return's restocked items have already been sold again — undo isn't possible. Adjust stock manually instead.",
					Status:  400,
				}
			}
		}
	}
	for _, m := range movements {
		if _, err := tx.Exec(ctx, `DELETE FROM stock_movements WHERE id = $1`, m.id); err != nil {
			return fmt.Errorf("delete movement: %w", err)
		}
		if m.batchID != nil {
			if _, err := tx.Exec(ctx, `DELETE FROM batches WHERE id = $1`, *m.batchID); err != nil {
				return fmt.Errorf("delete batch: %w", err)
			}
		}
		if _, err := tx.Exec(ctx, `UPDATE products SET stock_quantity = stock_quantity - $1 WHERE id = $2 AND branch_id = $3`,
			m.qty, m.productID, branchID); err != nil {
			return fmt.Errorf("deduct restock: %w", err)
		}
	}

	if _, err := tx.Exec(ctx, `DELETE FROM deposits WHERE return_id = $1`, returnID); err != nil {
		return fmt.Errorf("delete deposit: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM return_items WHERE return_id = $1`, returnID); err != nil {
		return fmt.Errorf("delete return items: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM returns WHERE id = $1`, returnID); err != nil {
		return fmt.Errorf("delete return: %w", err)
	}

	// Recompute the invoice's refunded flag from the remaining returns.
	var totalReturned, invTotal, paid float64
	if err := tx.QueryRow(ctx, `SELECT COALESCE(SUM(refund_amount), 0) FROM returns WHERE invoice_id = $1`, invoiceID).Scan(&totalReturned); err != nil {
		return fmt.Errorf("sum returns: %w", err)
	}
	if err := tx.QueryRow(ctx, `SELECT total_usd, paid_amount FROM invoices WHERE id = $1`, invoiceID).Scan(&invTotal, &paid); err != nil {
		return fmt.Errorf("get invoice totals: %w", err)
	}
	if totalReturned < invTotal-0.001 {
		paymentStatus := "unpaid"
		if paid >= invTotal-0.001 {
			paymentStatus = "paid"
		} else if paid > 0 {
			paymentStatus = "partial"
		}
		if _, err := tx.Exec(ctx, `UPDATE invoices SET payment_status = $1, updated_at = NOW() WHERE id = $2`, paymentStatus, invoiceID); err != nil {
			return fmt.Errorf("restore payment status: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
