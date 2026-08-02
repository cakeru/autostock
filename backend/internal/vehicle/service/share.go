package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/vehicle/dto"
)

// EnsureShareLink returns the vehicle's share token, minting one if it doesn't
// have an active link yet. Repeated calls return the same token, so a staff
// member re-opening the share dialog doesn't silently invalidate a link the
// customer already has.
func (s *Service) EnsureShareLink(ctx context.Context, branchID, vehicleID int64) (string, error) {
	var existing *string
	err := s.pool.QueryRow(ctx,
		`SELECT share_token FROM vehicles WHERE id = $1 AND branch_id = $2`, vehicleID, branchID).Scan(&existing)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", domain.ErrNotFound
		}
		return "", fmt.Errorf("load share token: %w", err)
	}
	if existing != nil && *existing != "" {
		return *existing, nil
	}

	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate share token: %w", err)
	}
	token := hex.EncodeToString(buf)

	if _, err := s.pool.Exec(ctx,
		`UPDATE vehicles SET share_token = $1, updated_at = NOW() WHERE id = $2 AND branch_id = $3`,
		token, vehicleID, branchID); err != nil {
		return "", fmt.Errorf("save share token: %w", err)
	}
	return token, nil
}

// RevokeShareLink clears the token — the old URL immediately stops resolving.
func (s *Service) RevokeShareLink(ctx context.Context, branchID, vehicleID int64) error {
	result, err := s.pool.Exec(ctx,
		`UPDATE vehicles SET share_token = NULL, updated_at = NOW() WHERE id = $1 AND branch_id = $2`,
		vehicleID, branchID)
	if err != nil {
		return fmt.Errorf("revoke share token: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// GetPublicReport assembles the read-only condition report for a share token.
// The token is the sole authorization — no JWT, no branch context — so the
// lookup is by token alone and the payload stays customer-safe (no internal
// notes, no contact details, no row ids beyond what rendering needs).
func (s *Service) GetPublicReport(ctx context.Context, token string) (*dto.PublicReportResponse, error) {
	if len(token) < 32 {
		return nil, domain.ErrNotFound
	}

	var vehicleID, branchID, customerID int64
	var year *int
	r := &dto.PublicReportResponse{GeneratedAt: time.Now(), DistanceUnit: "km"}
	err := s.pool.QueryRow(ctx, `
		SELECT vh.id, vh.branch_id, vh.customer_id, vh.plate_number,
		       COALESCE(vh.make,''), COALESCE(vh.model,''), vh.year, COALESCE(vh.body_type,''),
		       COALESCE(c.name,'')
		FROM vehicles vh JOIN customers c ON c.id = vh.customer_id
		WHERE vh.share_token = $1`, token).
		Scan(&vehicleID, &branchID, &customerID, &r.PlateNumber,
			&r.Make, &r.Model, &year, &r.BodyType, &r.CustomerName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("resolve share token: %w", err)
	}
	if year != nil {
		r.Year = *year
	}

	shopRows, err := s.pool.Query(ctx,
		`SELECT key, value FROM settings WHERE branch_id = $1 AND key IN ('shop_name', 'shop_phone', 'shop_address', 'distance_unit')`,
		branchID)
	if err == nil {
		for shopRows.Next() {
			var k, v string
			if shopRows.Scan(&k, &v) == nil {
				switch k {
				case "shop_name":
					r.ShopName = v
				case "shop_phone":
					r.ShopPhone = v
				case "shop_address":
					r.ShopAddress = v
				case "distance_unit":
					if v == "mi" || v == "km" {
						r.DistanceUnit = v
					}
				}
			}
		}
		shopRows.Close()
	}

	if r.Due, err = s.GetDueForVehicle(ctx, branchID, vehicleID); err != nil {
		return nil, err
	}
	if r.WheelServices, err = s.ListWheelServices(ctx, branchID, vehicleID); err != nil {
		return nil, err
	}
	if r.PartStatuses, err = s.ListPartStatuses(ctx, branchID, vehicleID); err != nil {
		return nil, err
	}
	if r.Parts, err = s.ListParts(ctx, branchID, vehicleID); err != nil {
		return nil, err
	}
	if r.Visits, err = s.GetPublicTimeline(ctx, branchID, vehicleID); err != nil {
		return nil, err
	}
	return r, nil
}
