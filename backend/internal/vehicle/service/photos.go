package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/vehicle/dto"
)

func (s *Service) ListGalleryPhotos(ctx context.Context, branchID, vehicleID int64) ([]dto.GalleryPhotoResponse, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.id, p.url, COALESCE(p.caption,''), COALESCE(p.phase,''), COALESCE(u.full_name,''), p.created_at,
		       COALESCE(p.taken_at, p.created_at), p.customer_visible
		FROM vehicle_photos p
		LEFT JOIN users u ON u.id = p.created_by
		WHERE p.vehicle_id = $1 AND p.branch_id = $2
		ORDER BY p.created_at ASC, p.id ASC`, vehicleID, branchID)
	if err != nil {
		return nil, fmt.Errorf("list gallery photos: %w", err)
	}
	defer rows.Close()

	out := []dto.GalleryPhotoResponse{}
	for rows.Next() {
		var p dto.GalleryPhotoResponse
		if err := rows.Scan(&p.ID, &p.URL, &p.Caption, &p.Phase, &p.CreatedByName, &p.CreatedAt, &p.TakenAt, &p.CustomerVisible); err != nil {
			return nil, fmt.Errorf("scan gallery photo: %w", err)
		}
		out = append(out, p)
	}
	return out, nil
}

func (s *Service) AddGalleryPhoto(ctx context.Context, branchID, vehicleID, userID int64, url string, takenAt *time.Time) (*dto.GalleryPhotoResponse, error) {
	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM vehicles WHERE id = $1 AND branch_id = $2)`,
		vehicleID, branchID).Scan(&exists); err != nil || !exists {
		return nil, domain.ErrNotFound
	}
	var p dto.GalleryPhotoResponse
	err := s.pool.QueryRow(ctx, `
		INSERT INTO vehicle_photos (branch_id, vehicle_id, url, created_by, taken_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, url, COALESCE(caption,''), COALESCE(phase,''), created_at, COALESCE(taken_at, created_at), customer_visible`,
		branchID, vehicleID, url, userID, takenAt).Scan(&p.ID, &p.URL, &p.Caption, &p.Phase, &p.CreatedAt, &p.TakenAt, &p.CustomerVisible)
	if err != nil {
		return nil, fmt.Errorf("add gallery photo: %w", err)
	}
	return &p, nil
}

func (s *Service) UpdateGalleryPhoto(ctx context.Context, branchID, photoID int64, req *dto.UpdateGalleryPhotoRequest) (*dto.GalleryPhotoResponse, error) {
	if req.Phase != nil && *req.Phase != "" && *req.Phase != "before" && *req.Phase != "after" {
		return nil, &domain.AppError{Code: "INVALID_PHASE", Message: "phase must be before, after or empty", Status: 400}
	}
	var p dto.GalleryPhotoResponse
	// COALESCE($n, existing) leaves an unspecified field untouched; NULLIF('') on
	// phase turns an explicit "" into a real NULL (clears the tag).
	err := s.pool.QueryRow(ctx, `
		UPDATE vehicle_photos SET
		    caption = COALESCE($1, caption),
		    phase = CASE WHEN $2::varchar IS NULL THEN phase ELSE NULLIF($2, '') END,
		    customer_visible = COALESCE($5, customer_visible)
		WHERE id = $3 AND branch_id = $4
		RETURNING id, url, COALESCE(caption,''), COALESCE(phase,''), created_at, COALESCE(taken_at, created_at), customer_visible`,
		req.Caption, req.Phase, photoID, branchID, req.CustomerVisible).Scan(&p.ID, &p.URL, &p.Caption, &p.Phase, &p.CreatedAt, &p.TakenAt, &p.CustomerVisible)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("update gallery photo: %w", err)
	}
	return &p, nil
}

// DeleteGalleryPhoto removes the row and returns its URL so the caller can
// clean the file out of storage.
func (s *Service) DeleteGalleryPhoto(ctx context.Context, branchID, photoID int64) (string, error) {
	var url string
	err := s.pool.QueryRow(ctx,
		`DELETE FROM vehicle_photos WHERE id = $1 AND branch_id = $2 RETURNING url`, photoID, branchID).Scan(&url)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", domain.ErrNotFound
		}
		return "", fmt.Errorf("delete gallery photo: %w", err)
	}
	return url, nil
}
