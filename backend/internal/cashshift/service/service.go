package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/cashshift/dto"
	"github.com/cakeru/autostock/internal/domain"
	telegrammodels "github.com/cakeru/autostock/internal/telegram/models"
	telegram "github.com/cakeru/autostock/internal/telegram/service"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

// cashSales sums cash payments taken in the branch within [from, to].
func (s *Service) cashSales(ctx context.Context, branchID int64, from time.Time, to *time.Time) (float64, error) {
	var total float64
	var err error
	if to == nil {
		err = s.pool.QueryRow(ctx, `
			SELECT COALESCE(SUM(p.amount), 0) FROM payments p
			JOIN invoices i ON i.id = p.invoice_id
			WHERE i.branch_id = $1 AND p.method = 'cash' AND p.created_at >= $2`,
			branchID, from).Scan(&total)
	} else {
		err = s.pool.QueryRow(ctx, `
			SELECT COALESCE(SUM(p.amount), 0) FROM payments p
			JOIN invoices i ON i.id = p.invoice_id
			WHERE i.branch_id = $1 AND p.method = 'cash' AND p.created_at >= $2 AND p.created_at <= $3`,
			branchID, from, *to).Scan(&total)
	}
	return round2(total), err
}

// GetCurrent returns the branch's open drawer (with live cash sales) or nil.
func (s *Service) GetCurrent(ctx context.Context, branchID int64) (*dto.ShiftResponse, error) {
	var r dto.ShiftResponse
	var openedAt time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT c.id, c.status, c.opened_by, COALESCE(u.full_name, u.username, ''), c.opening_float, c.opened_at
		FROM cash_shifts c LEFT JOIN users u ON u.id = c.opened_by
		WHERE c.branch_id = $1 AND c.status = 'open'`, branchID).
		Scan(&r.ID, &r.Status, &r.OpenedByID, &r.OpenedByName, &r.OpeningFloat, &openedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get current shift: %w", err)
	}
	r.OpenedAt = openedAt.Format(time.RFC3339)
	sales, err := s.cashSales(ctx, branchID, openedAt, nil)
	if err != nil {
		return nil, fmt.Errorf("live cash sales: %w", err)
	}
	r.CashSales = sales
	r.ExpectedAmount = round2(r.OpeningFloat + sales)
	return &r, nil
}

func (s *Service) Open(ctx context.Context, branchID, userID int64, req *dto.OpenRequest) (*dto.ShiftResponse, error) {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO cash_shifts (branch_id, opened_by, opening_float) VALUES ($1, $2, $3)`,
		branchID, userID, req.OpeningFloat)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, &domain.AppError{Code: "DRAWER_OPEN", Message: "A cash drawer is already open. Close it first.", Status: 400}
		}
		return nil, fmt.Errorf("open shift: %w", err)
	}
	return s.GetCurrent(ctx, branchID)
}

// Close reconciles and closes the branch's currently open drawer.
func (s *Service) Close(ctx context.Context, branchID, userID int64, req *dto.CloseRequest) (*dto.ShiftResponse, error) {
	var id int64
	var openedAt time.Time
	var openingFloat float64
	err := s.pool.QueryRow(ctx,
		`SELECT id, opened_at, opening_float FROM cash_shifts WHERE branch_id = $1 AND status = 'open'`,
		branchID).Scan(&id, &openedAt, &openingFloat)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get shift: %w", err)
	}

	now := time.Now()
	sales, err := s.cashSales(ctx, branchID, openedAt, &now)
	if err != nil {
		return nil, fmt.Errorf("shift cash sales: %w", err)
	}
	expected := round2(openingFloat + sales)
	overShort := round2(req.ClosingAmount - expected)

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE cash_shifts
		SET status = 'closed', closed_by = $1, closing_amount = $2, cash_sales = $3,
		    expected_amount = $4, over_short = $5, note = NULLIF($6, ''), closed_at = $7
		WHERE id = $8`,
		userID, req.ClosingAmount, sales, expected, overShort, req.Note, now, id)
	if err != nil {
		return nil, fmt.Errorf("close shift: %w", err)
	}

	if overShort != 0 {
		if err := telegram.LogEvent(ctx, tx, branchID, telegrammodels.TopicAlerts, "cash_discrepancy", "cash_shift", id, map[string]any{
			"expected": expected, "actual": req.ClosingAmount, "over_short": overShort,
		}); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return s.getByID(ctx, branchID, id)
}

func (s *Service) getByID(ctx context.Context, branchID, id int64) (*dto.ShiftResponse, error) {
	rows, err := s.query(ctx, `WHERE c.branch_id = $1 AND c.id = $2 LIMIT 1`, branchID, id)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, domain.ErrNotFound
	}
	return &rows[0], nil
}

func (s *Service) List(ctx context.Context, branchID int64, limit int) ([]dto.ShiftResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.query(ctx, `WHERE c.branch_id = $1 ORDER BY c.opened_at DESC LIMIT $2`, branchID, limit)
}

// query builds ShiftResponses for a WHERE/ORDER clause.
func (s *Service) query(ctx context.Context, clause string, args ...interface{}) ([]dto.ShiftResponse, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT c.id, c.status, c.opened_by, COALESCE(uo.full_name, uo.username, ''),
		       c.opening_float, c.opened_at,
		       COALESCE(c.cash_sales, 0), COALESCE(c.expected_amount, c.opening_float),
		       c.closing_amount, c.over_short, COALESCE(c.note, ''),
		       COALESCE(uc.full_name, uc.username, ''), c.closed_at
		FROM cash_shifts c
		LEFT JOIN users uo ON uo.id = c.opened_by
		LEFT JOIN users uc ON uc.id = c.closed_by `+clause, args...)
	if err != nil {
		return nil, fmt.Errorf("query shifts: %w", err)
	}
	defer rows.Close()

	out := []dto.ShiftResponse{}
	for rows.Next() {
		var r dto.ShiftResponse
		var openedAt time.Time
		var closedAt *time.Time
		if err := rows.Scan(&r.ID, &r.Status, &r.OpenedByID, &r.OpenedByName,
			&r.OpeningFloat, &openedAt, &r.CashSales, &r.ExpectedAmount,
			&r.ClosingAmount, &r.OverShort, &r.Note, &r.ClosedByName, &closedAt); err != nil {
			return nil, fmt.Errorf("scan shift: %w", err)
		}
		r.OpenedAt = openedAt.Format(time.RFC3339)
		if closedAt != nil {
			s := closedAt.Format(time.RFC3339)
			r.ClosedAt = &s
		}
		out = append(out, r)
	}
	return out, nil
}
