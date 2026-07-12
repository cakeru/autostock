package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/audit/dto"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) List(ctx context.Context, branchID int64, f dto.AuditFilter) ([]dto.AuditLogItem, int, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage < 1 || f.PerPage > 100 {
		f.PerPage = 25
	}
	offset := (f.Page - 1) * f.PerPage

	var userID *int64
	if f.UserID > 0 {
		userID = &f.UserID
	}

	rows, err := s.pool.Query(ctx, `
		SELECT a.id, a.action, a.entity_type, a.entity_id, a.user_id, COALESCE(u.full_name, u.username, ''),
		       a.created_at::text, COUNT(*) OVER() AS total
		FROM audit_logs a
		LEFT JOIN users u ON u.id = a.user_id
		WHERE a.branch_id = $1
		  AND ($2::text IS NULL OR a.action = $2)
		  AND ($3::text IS NULL OR a.entity_type = $3)
		  AND ($4::bigint IS NULL OR a.user_id = $4)
		  AND ($5::text IS NULL OR a.created_at >= $5::timestamptz)
		  AND ($6::text IS NULL OR a.created_at < ($6::date + 1))
		ORDER BY a.created_at DESC
		LIMIT $7 OFFSET $8`,
		branchID, nullIfEmpty(f.Action), nullIfEmpty(f.EntityType), userID,
		nullIfEmpty(f.From), nullIfEmpty(f.To), f.PerPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query audit logs: %w", err)
	}
	defer rows.Close()

	var items []dto.AuditLogItem
	var total int
	for rows.Next() {
		var it dto.AuditLogItem
		if err := rows.Scan(&it.ID, &it.Action, &it.EntityType, &it.EntityID, &it.UserID,
			&it.UserName, &it.CreatedAt, &total); err != nil {
			return nil, 0, fmt.Errorf("scan audit log: %w", err)
		}
		items = append(items, it)
	}
	if items == nil {
		items = []dto.AuditLogItem{}
	}
	return items, total, nil
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
