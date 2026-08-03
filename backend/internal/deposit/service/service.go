package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/deposit/dto"
	"github.com/cakeru/autostock/internal/domain"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

const selectDeposit = `
	SELECT d.id, d.customer_id, COALESCE(c.name, ''), d.amount, COALESCE(d.note, ''), d.status,
	       d.invoice_id, COALESCE(i.invoice_number, ''), d.created_at::text, d.settled_at::text
	FROM deposits d
	LEFT JOIN customers c ON c.id = d.customer_id
	LEFT JOIN invoices i ON i.id = d.invoice_id
`

func scanDeposit(row pgx.Row) (*dto.DepositResponse, error) {
	var d dto.DepositResponse
	var settled *string
	if err := row.Scan(&d.ID, &d.CustomerID, &d.CustomerName, &d.Amount, &d.Note, &d.Status,
		&d.InvoiceID, &d.InvoiceNumber, &d.CreatedAt, &settled); err != nil {
		return nil, err
	}
	d.SettledAt = settled
	return &d, nil
}

func (s *Service) getByID(ctx context.Context, branchID, id int64) (*dto.DepositResponse, error) {
	d, err := scanDeposit(s.pool.QueryRow(ctx, selectDeposit+` WHERE d.branch_id = $1 AND d.id = $2`, branchID, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get deposit: %w", err)
	}
	return d, nil
}

func (s *Service) Create(ctx context.Context, branchID, userID int64, req *dto.CreateDepositRequest) (*dto.DepositResponse, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO deposits (branch_id, customer_id, amount, note, created_by)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5) RETURNING id`,
		branchID, req.CustomerID, req.Amount, req.Note, userID).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create deposit: %w", err)
	}
	return s.getByID(ctx, branchID, id)
}

// Update edits a held deposit's amount/note (applied/refunded deposits are
// settled — correct those by adjusting the invoice instead).
func (s *Service) Update(ctx context.Context, branchID, id int64, req *dto.CreateDepositRequest) (*dto.DepositResponse, error) {
	d, err := s.getByID(ctx, branchID, id)
	if err != nil {
		return nil, err
	}
	if d.Status != "held" {
		return nil, &domain.AppError{Code: "DEPOSIT_SETTLED", Message: "Only held deposits can be edited", Status: 400}
	}
	if _, err := s.pool.Exec(ctx, `
		UPDATE deposits SET customer_id = $1, amount = $2, note = NULLIF($3, '') WHERE id = $4 AND branch_id = $5`,
		req.CustomerID, req.Amount, req.Note, id, branchID); err != nil {
		return nil, fmt.Errorf("update deposit: %w", err)
	}
	return s.getByID(ctx, branchID, id)
}

func (s *Service) List(ctx context.Context, branchID, customerID int64, status string) ([]dto.DepositResponse, error) {
	var cust *int64
	if customerID > 0 {
		cust = &customerID
	}
	var st *string
	if status != "" {
		st = &status
	}
	rows, err := s.pool.Query(ctx, selectDeposit+`
		WHERE d.branch_id = $1
		  AND ($2::bigint IS NULL OR d.customer_id = $2)
		  AND ($3::text IS NULL OR d.status = $3)
		ORDER BY d.created_at DESC`, branchID, cust, st)
	if err != nil {
		return nil, fmt.Errorf("list deposits: %w", err)
	}
	defer rows.Close()

	out := []dto.DepositResponse{}
	for rows.Next() {
		d, err := scanDeposit(rows)
		if err != nil {
			return nil, fmt.Errorf("scan deposit: %w", err)
		}
		out = append(out, *d)
	}
	return out, nil
}

// Apply records a held deposit as a payment against one of the customer's
// invoices, then marks the deposit applied.
func (s *Service) Apply(ctx context.Context, branchID, userID, id, invoiceID int64) (*dto.DepositResponse, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var depCustomer int64
	var amount float64
	var status string
	if err := tx.QueryRow(ctx,
		`SELECT customer_id, amount, status FROM deposits WHERE id = $1 AND branch_id = $2 FOR UPDATE`,
		id, branchID).Scan(&depCustomer, &amount, &status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get deposit: %w", err)
	}
	if status != "held" {
		return nil, &domain.AppError{Code: "SETTLED", Message: "Deposit has already been settled", Status: 400}
	}

	var invCustomer *int64
	var total, paid float64
	var invStatus string
	if err := tx.QueryRow(ctx,
		`SELECT customer_id, total_usd, paid_amount, status FROM invoices WHERE id = $1 AND branch_id = $2 FOR UPDATE`,
		invoiceID, branchID).Scan(&invCustomer, &total, &paid, &invStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &domain.AppError{Code: "INVOICE_NOT_FOUND", Message: "Invoice not found", Status: 404}
		}
		return nil, fmt.Errorf("get invoice: %w", err)
	}
	if invStatus == "voided" {
		return nil, &domain.AppError{Code: "VOIDED", Message: "Cannot apply a deposit to a voided invoice", Status: 400}
	}
	if invCustomer == nil || *invCustomer != depCustomer {
		return nil, &domain.AppError{Code: "CUSTOMER_MISMATCH", Message: "Invoice belongs to a different customer", Status: 400}
	}
	owed := total - paid
	if amount > owed+0.001 {
		return nil, &domain.AppError{Code: "EXCEEDS_OWED", Message: "Deposit exceeds the amount owed on this invoice", Status: 400}
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO payments (invoice_id, amount, method, received_by, notes)
		VALUES ($1, $2, 'deposit', $3, 'Deposit applied')`, invoiceID, amount, userID); err != nil {
		return nil, fmt.Errorf("record deposit payment: %w", err)
	}

	var sumPaid float64
	if err := tx.QueryRow(ctx, `SELECT COALESCE(SUM(amount), 0) FROM payments WHERE invoice_id = $1`, invoiceID).Scan(&sumPaid); err != nil {
		return nil, fmt.Errorf("sum payments: %w", err)
	}
	paymentStatus := "partial"
	newStatus := invStatus
	if sumPaid >= total {
		paymentStatus = "paid"
		newStatus = "paid"
	}
	if _, err := tx.Exec(ctx,
		`UPDATE invoices SET paid_amount = $1, payment_status = $2, status = $3, updated_at = NOW() WHERE id = $4`,
		sumPaid, paymentStatus, newStatus, invoiceID); err != nil {
		return nil, fmt.Errorf("update invoice: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`UPDATE deposits SET status = 'applied', invoice_id = $1, settled_at = NOW() WHERE id = $2`,
		invoiceID, id); err != nil {
		return nil, fmt.Errorf("mark deposit applied: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return s.getByID(ctx, branchID, id)
}

func (s *Service) Refund(ctx context.Context, branchID, id int64) (*dto.DepositResponse, error) {
	tag, err := s.pool.Exec(ctx,
		`UPDATE deposits SET status = 'refunded', settled_at = NOW() WHERE id = $1 AND branch_id = $2 AND status = 'held'`,
		id, branchID)
	if err != nil {
		return nil, fmt.Errorf("refund deposit: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, &domain.AppError{Code: "SETTLED", Message: "Only a held deposit can be refunded", Status: 400}
	}
	return s.getByID(ctx, branchID, id)
}
