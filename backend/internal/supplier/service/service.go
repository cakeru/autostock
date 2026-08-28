package service

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/supplier/dto"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

const selectSupplier = `
	SELECT s.id, s.name, COALESCE(s.phone,''), COALESCE(s.email,''), COALESCE(s.address,''),
	       COALESCE(s.notes,''), s.is_active, s.created_at::text,
	       COALESCE(SUM(b.quantity_received * b.unit_cost), 0),
	       COALESCE(SUM(b.quantity_received * b.unit_cost - b.amount_paid), 0),
	       COUNT(b.id)
	FROM suppliers s
	LEFT JOIN batches b ON b.supplier_id = s.id
`

func scanSupplier(row pgx.Row) (*dto.SupplierResponse, error) {
	var s dto.SupplierResponse
	if err := row.Scan(&s.ID, &s.Name, &s.Phone, &s.Email, &s.Address, &s.Notes,
		&s.IsActive, &s.CreatedAt, &s.TotalPurchased, &s.Outstanding, &s.PurchaseCount); err != nil {
		return nil, err
	}
	s.TotalPurchased = round2(s.TotalPurchased)
	s.Outstanding = round2(s.Outstanding)
	return &s, nil
}

func (s *Service) List(ctx context.Context, branchID int64) ([]dto.SupplierResponse, error) {
	rows, err := s.pool.Query(ctx, selectSupplier+`
		WHERE s.branch_id = $1 AND s.is_active
		GROUP BY s.id ORDER BY s.name`, branchID)
	if err != nil {
		return nil, fmt.Errorf("list suppliers: %w", err)
	}
	defer rows.Close()

	out := []dto.SupplierResponse{}
	for rows.Next() {
		sup, err := scanSupplier(rows)
		if err != nil {
			return nil, fmt.Errorf("scan supplier: %w", err)
		}
		out = append(out, *sup)
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, branchID, id int64) (*dto.SupplierResponse, error) {
	sup, err := scanSupplier(s.pool.QueryRow(ctx, selectSupplier+`
		WHERE s.branch_id = $1 AND s.id = $2 GROUP BY s.id`, branchID, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get supplier: %w", err)
	}
	return sup, nil
}

func (s *Service) Create(ctx context.Context, branchID int64, req *dto.CreateSupplierRequest) (*dto.SupplierResponse, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO suppliers (branch_id, name, phone, email, address, notes)
		VALUES ($1, $2, NULLIF($3,''), NULLIF($4,''), NULLIF($5,''), NULLIF($6,''))
		RETURNING id`,
		branchID, req.Name, req.Phone, req.Email, req.Address, req.Notes).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create supplier: %w", err)
	}
	return s.Get(ctx, branchID, id)
}

func (s *Service) Update(ctx context.Context, branchID, id int64, req *dto.UpdateSupplierRequest) (*dto.SupplierResponse, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE suppliers SET
			name = COALESCE($1, name),
			phone = COALESCE($2, phone),
			email = COALESCE($3, email),
			address = COALESCE($4, address),
			notes = COALESCE($5, notes),
			updated_at = NOW()
		WHERE id = $6 AND branch_id = $7`,
		req.Name, req.Phone, req.Email, req.Address, req.Notes, id, branchID)
	if err != nil {
		return nil, fmt.Errorf("update supplier: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}
	return s.Get(ctx, branchID, id)
}

func (s *Service) Deactivate(ctx context.Context, branchID, id int64) error {
	tag, err := s.pool.Exec(ctx, `UPDATE suppliers SET is_active = false, updated_at = NOW() WHERE id = $1 AND branch_id = $2`, id, branchID)
	if err != nil {
		return fmt.Errorf("deactivate supplier: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Service) Purchases(ctx context.Context, branchID, supplierID int64) ([]dto.PurchaseItem, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT b.id, b.product_id, COALESCE(p.name, ''), b.quantity_received, b.unit_cost,
		       b.quantity_received * b.unit_cost, b.amount_paid,
		       b.quantity_received * b.unit_cost - b.amount_paid, COALESCE(b.dot_code, ''), b.received_at::text
		FROM batches b LEFT JOIN products p ON p.id = b.product_id
		WHERE b.supplier_id = $1 AND b.branch_id = $2
		ORDER BY b.received_at DESC`, supplierID, branchID)
	if err != nil {
		return nil, fmt.Errorf("supplier purchases: %w", err)
	}
	defer rows.Close()

	out := []dto.PurchaseItem{}
	batchIDs := []int64{}
	for rows.Next() {
		var it dto.PurchaseItem
		if err := rows.Scan(&it.BatchID, &it.ProductID, &it.ProductName, &it.Quantity, &it.UnitCost,
			&it.TotalCost, &it.AmountPaid, &it.Owed, &it.DOTCode, &it.ReceivedAt); err != nil {
			return nil, fmt.Errorf("scan purchase: %w", err)
		}
		it.Invoices = []dto.BatchInvoice{}
		out = append(out, it)
		batchIDs = append(batchIDs, it.BatchID)
	}
	rows.Close()

	if len(batchIDs) > 0 {
		invRows, err := s.pool.Query(ctx, `
			SELECT id, batch_id, COALESCE(invoice_number, ''), COALESCE(invoice_image, ''), amount, amount_paid
			FROM batch_invoices WHERE batch_id = ANY($1) ORDER BY id`, batchIDs)
		if err != nil {
			return nil, fmt.Errorf("batch invoices: %w", err)
		}
		byBatch := map[int64][]dto.BatchInvoice{}
		for invRows.Next() {
			var inv dto.BatchInvoice
			var batchID int64
			if err := invRows.Scan(&inv.ID, &batchID, &inv.InvoiceNumber, &inv.InvoiceImage, &inv.Amount, &inv.AmountPaid); err != nil {
				invRows.Close()
				return nil, fmt.Errorf("scan batch invoice: %w", err)
			}
			inv.Owed = round2(inv.Amount - inv.AmountPaid)
			byBatch[batchID] = append(byBatch[batchID], inv)
		}
		invRows.Close()
		for i := range out {
			if invs, ok := byBatch[out[i].BatchID]; ok {
				out[i].Invoices = invs
			}
		}
	}
	return out, nil
}

// Pay settles the selected invoices in full (each invoice's outstanding), plus
// any selected whole purchases that have no invoices recorded. Batch-level
// amount_paid is kept in sync so supplier totals stay correct.
func (s *Service) Pay(ctx context.Context, branchID, supplierID int64, invoiceIDs, batchIDs []int64) (*dto.SupplierResponse, error) {
	if len(invoiceIDs) == 0 && len(batchIDs) == 0 {
		return nil, &domain.AppError{Code: "INVALID_REQUEST", Message: "Select at least one invoice or purchase to pay", Status: 400}
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM suppliers WHERE id=$1 AND branch_id=$2)`, supplierID, branchID).Scan(&exists); err != nil || !exists {
		return nil, domain.ErrNotFound
	}

	paid := 0.0

	if len(invoiceIDs) > 0 {
		rows, err := tx.Query(ctx, `
			SELECT bi.id, bi.batch_id, bi.amount - bi.amount_paid
			FROM batch_invoices bi
			JOIN batches b ON b.id = bi.batch_id
			WHERE bi.id = ANY($1) AND b.supplier_id = $2 AND b.branch_id = $3
			  AND (bi.amount - bi.amount_paid) > 0
			FOR UPDATE OF bi`, invoiceIDs, supplierID, branchID)
		if err != nil {
			return nil, fmt.Errorf("owed invoices: %w", err)
		}
		type owedInvoice struct {
			id, batchID int64
			owed        float64
		}
		var invoices []owedInvoice
		for rows.Next() {
			var o owedInvoice
			if err := rows.Scan(&o.id, &o.batchID, &o.owed); err != nil {
				rows.Close()
				return nil, fmt.Errorf("scan owed invoice: %w", err)
			}
			invoices = append(invoices, o)
		}
		rows.Close()
		for _, o := range invoices {
			if _, err := tx.Exec(ctx, `UPDATE batch_invoices SET amount_paid = amount_paid + $1 WHERE id = $2`, round2(o.owed), o.id); err != nil {
				return nil, fmt.Errorf("apply invoice payment: %w", err)
			}
			if _, err := tx.Exec(ctx, `UPDATE batches SET amount_paid = amount_paid + $1 WHERE id = $2`, round2(o.owed), o.batchID); err != nil {
				return nil, fmt.Errorf("apply batch payment: %w", err)
			}
			paid += o.owed
		}
	}

	if len(batchIDs) > 0 {
		// Whole-purchase payment only applies to batches with no invoices —
		// invoiced purchases are paid invoice by invoice.
		rows, err := tx.Query(ctx, `
			SELECT id, quantity_received * unit_cost - amount_paid
			FROM batches b
			WHERE b.id = ANY($1) AND b.supplier_id = $2 AND b.branch_id = $3
			  AND (quantity_received * unit_cost - amount_paid) > 0
			  AND NOT EXISTS (SELECT 1 FROM batch_invoices bi WHERE bi.batch_id = b.id)
			FOR UPDATE`, batchIDs, supplierID, branchID)
		if err != nil {
			return nil, fmt.Errorf("owed batches: %w", err)
		}
		type owedBatch struct {
			id   int64
			owed float64
		}
		var batches []owedBatch
		for rows.Next() {
			var o owedBatch
			if err := rows.Scan(&o.id, &o.owed); err != nil {
				rows.Close()
				return nil, fmt.Errorf("scan owed batch: %w", err)
			}
			batches = append(batches, o)
		}
		rows.Close()
		for _, o := range batches {
			if _, err := tx.Exec(ctx, `UPDATE batches SET amount_paid = amount_paid + $1 WHERE id = $2`, round2(o.owed), o.id); err != nil {
				return nil, fmt.Errorf("apply batch payment: %w", err)
			}
			paid += o.owed
		}
	}

	if paid == 0 {
		return nil, &domain.AppError{Code: "NOTHING_OWED", Message: "None of the selected items are unpaid", Status: 400}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return s.Get(ctx, branchID, supplierID)
}
