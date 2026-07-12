package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/vehicle/dto"
)

var validPositions = map[string]bool{"FL": true, "FR": true, "RL": true, "RR": true, "SPARE": true}

// ListWheelServices returns every wheel snapshot for a vehicle, newest first,
// each with its corners and printout photos hydrated.
func (s *Service) ListWheelServices(ctx context.Context, branchID, vehicleID int64) ([]dto.WheelServiceResponse, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ws.id, ws.performed_at, ws.mileage, ws.invoice_id, COALESCE(i.invoice_number,''),
		       ws.service_job_id, COALESCE(sj.job_number,''), COALESCE(ws.notes,''),
		       COALESCE(u.full_name,''), ws.created_at
		FROM vehicle_wheel_services ws
		LEFT JOIN invoices i ON i.id = ws.invoice_id
		LEFT JOIN service_jobs sj ON sj.id = ws.service_job_id
		LEFT JOIN users u ON u.id = ws.created_by
		WHERE ws.vehicle_id = $1 AND ws.branch_id = $2
		ORDER BY ws.performed_at DESC, ws.id DESC`, vehicleID, branchID)
	if err != nil {
		return nil, fmt.Errorf("list wheel services: %w", err)
	}

	var services []dto.WheelServiceResponse
	var ids []int64
	for rows.Next() {
		var w dto.WheelServiceResponse
		if err := rows.Scan(&w.ID, &w.PerformedAt, &w.Mileage, &w.InvoiceID, &w.InvoiceNumber,
			&w.ServiceJobID, &w.JobNumber, &w.Notes, &w.CreatedByName, &w.CreatedAt); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan wheel service: %w", err)
		}
		w.Corners = []dto.CornerData{}
		w.Photos = []dto.PhotoResponse{}
		services = append(services, w)
		ids = append(ids, w.ID)
	}
	rows.Close()

	if len(ids) == 0 {
		return []dto.WheelServiceResponse{}, nil
	}

	byService := map[int64]int{}
	for i, w := range services {
		byService[w.ID] = i
	}

	cornerRows, err := s.pool.Query(ctx, `
		SELECT wheel_service_id, position, tire_product_id, COALESCE(tire_brand,''), COALESCE(tire_size,''),
		       COALESCE(tire_dot,''), tread_mm, tread_before_mm, pressure,
		       COALESCE(camber_before,''), COALESCE(camber_after,''),
		       COALESCE(caster_before,''), COALESCE(caster_after,''),
		       COALESCE(toe_before,''), COALESCE(toe_after,''), COALESCE(wear_note,'')
		FROM wheel_service_corners
		WHERE wheel_service_id = ANY($1)
		ORDER BY CASE position WHEN 'FL' THEN 1 WHEN 'FR' THEN 2 WHEN 'RL' THEN 3 WHEN 'RR' THEN 4 ELSE 5 END`, ids)
	if err != nil {
		return nil, fmt.Errorf("list corners: %w", err)
	}
	for cornerRows.Next() {
		var wsID int64
		var c dto.CornerData
		if err := cornerRows.Scan(&wsID, &c.Position, &c.TireProductID, &c.TireBrand, &c.TireSize,
			&c.TireDOT, &c.TreadMM, &c.TreadBeforeMM, &c.Pressure, &c.CamberBefore, &c.CamberAfter,
			&c.CasterBefore, &c.CasterAfter, &c.ToeBefore, &c.ToeAfter, &c.WearNote); err != nil {
			cornerRows.Close()
			return nil, fmt.Errorf("scan corner: %w", err)
		}
		if idx, ok := byService[wsID]; ok {
			services[idx].Corners = append(services[idx].Corners, c)
		}
	}
	cornerRows.Close()

	photoRows, err := s.pool.Query(ctx,
		`SELECT id, wheel_service_id, url FROM wheel_service_photos WHERE wheel_service_id = ANY($1) ORDER BY created_at`, ids)
	if err != nil {
		return nil, fmt.Errorf("list wheel photos: %w", err)
	}
	for photoRows.Next() {
		var wsID int64
		var p dto.PhotoResponse
		if err := photoRows.Scan(&p.ID, &wsID, &p.URL); err != nil {
			photoRows.Close()
			return nil, fmt.Errorf("scan wheel photo: %w", err)
		}
		if idx, ok := byService[wsID]; ok {
			services[idx].Photos = append(services[idx].Photos, p)
		}
	}
	photoRows.Close()

	return services, nil
}

func (s *Service) getWheelService(ctx context.Context, branchID, id int64) (*dto.WheelServiceResponse, error) {
	// Reuse the list path scoped to this vehicle's service — cheap enough and
	// keeps the hydration logic in one place.
	var vehicleID int64
	err := s.pool.QueryRow(ctx,
		`SELECT vehicle_id FROM vehicle_wheel_services WHERE id = $1 AND branch_id = $2`, id, branchID).Scan(&vehicleID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get wheel service: %w", err)
	}
	all, err := s.ListWheelServices(ctx, branchID, vehicleID)
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].ID == id {
			return &all[i], nil
		}
	}
	return nil, domain.ErrNotFound
}

func (s *Service) CreateWheelService(ctx context.Context, branchID, vehicleID, userID int64, req *dto.CreateWheelServiceRequest) (*dto.WheelServiceResponse, error) {
	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM vehicles WHERE id = $1 AND branch_id = $2)`,
		vehicleID, branchID).Scan(&exists); err != nil || !exists {
		return nil, domain.ErrNotFound
	}

	performedAt := time.Now()
	if req.PerformedAt != "" {
		parsed, err := time.Parse("2006-01-02", req.PerformedAt)
		if err != nil {
			return nil, &domain.AppError{Code: "INVALID_DATE", Message: "performed_at must be YYYY-MM-DD", Status: 400}
		}
		performedAt = parsed
	}

	seen := map[string]bool{}
	for _, c := range req.Corners {
		if !validPositions[c.Position] {
			return nil, &domain.AppError{Code: "INVALID_POSITION", Message: fmt.Sprintf("invalid wheel position %q", c.Position), Status: 400}
		}
		if seen[c.Position] {
			return nil, &domain.AppError{Code: "DUPLICATE_POSITION", Message: fmt.Sprintf("duplicate wheel position %q", c.Position), Status: 400}
		}
		seen[c.Position] = true
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var id int64
	err = tx.QueryRow(ctx, `
		INSERT INTO vehicle_wheel_services (branch_id, vehicle_id, performed_at, mileage, invoice_id, service_job_id, notes, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8)
		RETURNING id`,
		branchID, vehicleID, performedAt, req.Mileage, req.InvoiceID, req.ServiceJobID, req.Notes, userID).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create wheel service: %w", err)
	}

	for _, c := range req.Corners {
		_, err := tx.Exec(ctx, `
			INSERT INTO wheel_service_corners
			  (wheel_service_id, position, tire_product_id, tire_brand, tire_size, tire_dot,
			   tread_mm, tread_before_mm, pressure, camber_before, camber_after, caster_before, caster_after,
			   toe_before, toe_after, wear_note)
			VALUES ($1, $2, $3, NULLIF($4,''), NULLIF($5,''), NULLIF($6,''),
			   $7, $8, $9, NULLIF($10,''), NULLIF($11,''), NULLIF($12,''), NULLIF($13,''),
			   NULLIF($14,''), NULLIF($15,''), NULLIF($16,''))`,
			id, c.Position, c.TireProductID, c.TireBrand, c.TireSize, c.TireDOT,
			c.TreadMM, c.TreadBeforeMM, c.Pressure, c.CamberBefore, c.CamberAfter, c.CasterBefore, c.CasterAfter,
			c.ToeBefore, c.ToeAfter, c.WearNote)
		if err != nil {
			return nil, fmt.Errorf("insert corner %s: %w", c.Position, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit wheel service: %w", err)
	}
	return s.getWheelService(ctx, branchID, id)
}

// DeleteWheelService removes a snapshot and returns its photo URLs so the
// caller can clean storage (cascade drops the DB rows).
func (s *Service) DeleteWheelService(ctx context.Context, branchID, id int64) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.url FROM wheel_service_photos p
		JOIN vehicle_wheel_services ws ON ws.id = p.wheel_service_id
		WHERE p.wheel_service_id = $1 AND ws.branch_id = $2`, id, branchID)
	if err != nil {
		return nil, fmt.Errorf("list wheel photo urls: %w", err)
	}
	var urls []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan url: %w", err)
		}
		urls = append(urls, u)
	}
	rows.Close()

	result, err := s.pool.Exec(ctx, `DELETE FROM vehicle_wheel_services WHERE id = $1 AND branch_id = $2`, id, branchID)
	if err != nil {
		return nil, fmt.Errorf("delete wheel service: %w", err)
	}
	if result.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}
	return urls, nil
}

func (s *Service) AddWheelServicePhoto(ctx context.Context, branchID, wheelServiceID int64, url string) (*dto.PhotoResponse, error) {
	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM vehicle_wheel_services WHERE id = $1 AND branch_id = $2)`,
		wheelServiceID, branchID).Scan(&exists); err != nil || !exists {
		return nil, domain.ErrNotFound
	}
	var p dto.PhotoResponse
	err := s.pool.QueryRow(ctx,
		`INSERT INTO wheel_service_photos (wheel_service_id, url) VALUES ($1, $2) RETURNING id, url`,
		wheelServiceID, url).Scan(&p.ID, &p.URL)
	if err != nil {
		return nil, fmt.Errorf("add wheel photo: %w", err)
	}
	return &p, nil
}

// RecentTireOptions offers tires the customer recently bought for this vehicle,
// so filling a corner is a pick rather than retyping brand/size. Sourced from
// tire line items on the vehicle's invoices, most recent first.
func (s *Service) RecentTireOptions(ctx context.Context, branchID, vehicleID int64) ([]dto.TireOption, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT ON (p.id) p.id, p.name, COALESCE(p.tire_size,''), i.id, i.invoice_number, i.issued_at
		FROM invoices i
		JOIN invoice_items ii ON ii.invoice_id = i.id
		JOIN products p ON p.id = ii.product_id
		WHERE i.vehicle_id = $1 AND i.branch_id = $2 AND i.status <> 'voided' AND p.type = 'tire'
		ORDER BY p.id, i.issued_at DESC`, vehicleID, branchID)
	if err != nil {
		return nil, fmt.Errorf("recent tire options: %w", err)
	}
	defer rows.Close()

	out := []dto.TireOption{}
	for rows.Next() {
		var t dto.TireOption
		if err := rows.Scan(&t.ProductID, &t.Name, &t.Size, &t.InvoiceID, &t.InvoiceNumber, &t.PurchasedAt); err != nil {
			return nil, fmt.Errorf("scan tire option: %w", err)
		}
		out = append(out, t)
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// General parts-replaced log
// ---------------------------------------------------------------------------

func (s *Service) ListParts(ctx context.Context, branchID, vehicleID int64) ([]dto.PartResponse, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT vp.id, vp.part_name, COALESCE(vp.part_key,''), COALESCE(vp.position,''), vp.replaced_at, vp.mileage,
		       vp.product_id, vp.invoice_id, COALESCE(i.invoice_number,''),
		       COALESCE(u.full_name,''), vp.created_at
		FROM vehicle_parts vp
		LEFT JOIN invoices i ON i.id = vp.invoice_id
		LEFT JOIN users u ON u.id = vp.created_by
		WHERE vp.vehicle_id = $1 AND vp.branch_id = $2
		ORDER BY vp.replaced_at DESC, vp.id DESC`, vehicleID, branchID)
	if err != nil {
		return nil, fmt.Errorf("list parts: %w", err)
	}
	defer rows.Close()

	out := []dto.PartResponse{}
	for rows.Next() {
		var p dto.PartResponse
		if err := rows.Scan(&p.ID, &p.PartName, &p.PartKey, &p.Position, &p.ReplacedAt, &p.Mileage,
			&p.ProductID, &p.InvoiceID, &p.InvoiceNumber, &p.CreatedByName, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan part: %w", err)
		}
		out = append(out, p)
	}
	return out, nil
}

func (s *Service) CreatePart(ctx context.Context, branchID, vehicleID, userID int64, req *dto.CreatePartRequest) (*dto.PartResponse, error) {
	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM vehicles WHERE id = $1 AND branch_id = $2)`,
		vehicleID, branchID).Scan(&exists); err != nil || !exists {
		return nil, domain.ErrNotFound
	}

	replacedAt := time.Now()
	if req.ReplacedAt != "" {
		parsed, err := time.Parse("2006-01-02", req.ReplacedAt)
		if err != nil {
			return nil, &domain.AppError{Code: "INVALID_DATE", Message: "replaced_at must be YYYY-MM-DD", Status: 400}
		}
		replacedAt = parsed
	}

	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO vehicle_parts (branch_id, vehicle_id, part_name, part_key, position, replaced_at, mileage, product_id, invoice_id, notes, created_by)
		VALUES ($1, $2, $3, NULLIF($4,''), NULLIF($5,''), $6, $7, $8, $9, NULLIF($10,''), $11)
		RETURNING id`,
		branchID, vehicleID, req.PartName, req.PartKey, req.Position, replacedAt, req.Mileage, req.ProductID, req.InvoiceID, req.Notes, userID).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create part: %w", err)
	}

	parts, err := s.ListParts(ctx, branchID, vehicleID)
	if err != nil {
		return nil, err
	}
	for i := range parts {
		if parts[i].ID == id {
			return &parts[i], nil
		}
	}
	return nil, domain.ErrNotFound
}

func (s *Service) DeletePart(ctx context.Context, branchID, id int64) error {
	result, err := s.pool.Exec(ctx, `DELETE FROM vehicle_parts WHERE id = $1 AND branch_id = $2`, id, branchID)
	if err != nil {
		return fmt.Errorf("delete part: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
