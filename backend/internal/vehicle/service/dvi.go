package service

import (
	"context"
	"fmt"

	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/vehicle/dto"
)

// ListPartStatuses returns the current DVI condition per part for a vehicle.
// Parts with no row are simply "not checked" (grey) and are omitted — the UI
// treats an absent key as grey.
func (s *Service) ListPartStatuses(ctx context.Context, branchID, vehicleID int64) ([]dto.PartStatusResponse, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ps.part_key, ps.status, COALESCE(ps.note,''), COALESCE(u.full_name,''), ps.updated_at
		FROM vehicle_part_status ps
		LEFT JOIN users u ON u.id = ps.updated_by
		WHERE ps.vehicle_id = $1 AND ps.branch_id = $2
		ORDER BY ps.part_key`, vehicleID, branchID)
	if err != nil {
		return nil, fmt.Errorf("list part statuses: %w", err)
	}
	defer rows.Close()

	out := []dto.PartStatusResponse{}
	for rows.Next() {
		var p dto.PartStatusResponse
		if err := rows.Scan(&p.PartKey, &p.Status, &p.Note, &p.UpdatedByName, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan part status: %w", err)
		}
		out = append(out, p)
	}
	return out, nil
}

// SetPartStatus upserts one part's condition. Setting it to "grey" clears the
// row entirely (grey == not checked == no row), so the table only holds parts
// that have actually been assessed.
func (s *Service) SetPartStatus(ctx context.Context, branchID, vehicleID, userID int64, req *dto.SetPartStatusRequest) error {
	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM vehicles WHERE id = $1 AND branch_id = $2)`,
		vehicleID, branchID).Scan(&exists); err != nil || !exists {
		return domain.ErrNotFound
	}

	if req.Status == "grey" {
		_, err := s.pool.Exec(ctx,
			`DELETE FROM vehicle_part_status WHERE vehicle_id = $1 AND part_key = $2`, vehicleID, req.PartKey)
		if err != nil {
			return fmt.Errorf("clear part status: %w", err)
		}
		return nil
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO vehicle_part_status (branch_id, vehicle_id, part_key, status, note, updated_by, updated_at)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, NOW())
		ON CONFLICT (vehicle_id, part_key)
		DO UPDATE SET status = EXCLUDED.status, note = EXCLUDED.note, updated_by = EXCLUDED.updated_by, updated_at = NOW()`,
		branchID, vehicleID, req.PartKey, req.Status, req.Note, userID)
	if err != nil {
		return fmt.Errorf("set part status: %w", err)
	}
	return nil
}
