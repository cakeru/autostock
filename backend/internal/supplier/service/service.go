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
	for rows.Next() {
		var it dto.PurchaseItem
		if err := rows.Scan(&it.BatchID, &it.ProductID, &it.ProductName, &it.Quantity, &it.UnitCost,
			&it.TotalCost, &it.AmountPaid, &it.Owed, &it.DOTCode, &it.ReceivedAt); err != nil {
			return nil, fmt.Errorf("scan purchase: %w", err)
		}
		out = append(out, it)
	}
	return out, nil
}

// Pay applies a payment to the supplier's oldest unpaid purchases (FIFO).
func (s *Service) Pay(ctx context.Context, branchID, supplierID int64, amount float64) (*dto.SupplierResponse, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM suppliers WHERE id=$1 AND branch_id=$2)`, supplierID, branchID).Scan(&exists); err != nil || !exists {
		return nil, domain.ErrNotFound
	}

	rows, err := tx.Query(ctx, `
		SELECT id, quantity_received * unit_cost - amount_paid AS owed
		FROM batches
		WHERE supplier_id = $1 AND branch_id = $2 AND (quantity_received * unit_cost - amount_paid) > 0
		ORDER BY received_at ASC`, supplierID, branchID)
	if err != nil {
		return nil, fmt.Errorf("owed batches: %w", err)
	}
	type owed struct {
		id  int64
		amt float64
	}
	var batches []owed
	for rows.Next() {
		var o owed
		if err := rows.Scan(&o.id, &o.amt); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan owed: %w", err)
		}
		batches = append(batches, o)
	}
	rows.Close()

	remaining := round2(amount)
	for _, b := range batches {
		if remaining <= 0 {
			break
		}
		pay := b.amt
		if pay > remaining {
			pay = remaining
		}
		if _, err := tx.Exec(ctx, `UPDATE batches SET amount_paid = amount_paid + $1 WHERE id = $2`, round2(pay), b.id); err != nil {
			return nil, fmt.Errorf("apply payment: %w", err)
		}
		remaining = round2(remaining - pay)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return s.Get(ctx, branchID, supplierID)
}
