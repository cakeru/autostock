package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/expense/dto"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) Create(ctx context.Context, branchID, userID int64, req *dto.CreateExpenseRequest) (*dto.ExpenseResponse, error) {
	spentAt := req.SpentAt
	if spentAt == "" {
		spentAt = time.Now().Format("2006-01-02")
	}

	var e dto.ExpenseResponse
	err := s.pool.QueryRow(ctx, `
		INSERT INTO expenses (branch_id, category, description, amount_usd, spent_at, created_by)
		VALUES ($1, $2, $3, $4, $5::date, $6)
		RETURNING id, category, COALESCE(description, ''), amount_usd, spent_at::text, created_by, created_at::text`,
		branchID, req.Category, req.Description, req.AmountUSD, spentAt, userID).
		Scan(&e.ID, &e.Category, &e.Description, &e.AmountUSD, &e.SpentAt, &e.CreatedBy, &e.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create expense: %w", err)
	}
	return &e, nil
}

func (s *Service) List(ctx context.Context, branchID int64, from, to string) ([]dto.ExpenseResponse, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT e.id, e.category, COALESCE(e.description, ''), e.amount_usd, e.spent_at::text,
		       e.created_by, COALESCE(u.full_name, ''), e.created_at::text
		FROM expenses e
		LEFT JOIN users u ON u.id = e.created_by
		WHERE e.branch_id = $1
		  AND ($2::date IS NULL OR e.spent_at >= $2::date)
		  AND ($3::date IS NULL OR e.spent_at <= $3::date)
		ORDER BY e.spent_at DESC, e.id DESC`,
		branchID, nullIfEmpty(from), nullIfEmpty(to))
	if err != nil {
		return nil, fmt.Errorf("list expenses: %w", err)
	}
	defer rows.Close()

	var out []dto.ExpenseResponse
	for rows.Next() {
		var e dto.ExpenseResponse
		if err := rows.Scan(&e.ID, &e.Category, &e.Description, &e.AmountUSD, &e.SpentAt,
			&e.CreatedBy, &e.CreatedName, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan expense: %w", err)
		}
		out = append(out, e)
	}
	if out == nil {
		out = []dto.ExpenseResponse{}
	}
	return out, nil
}

func (s *Service) Delete(ctx context.Context, branchID, id int64) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM expenses WHERE id = $1 AND branch_id = $2`, id, branchID)
	if err != nil {
		return fmt.Errorf("delete expense: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
