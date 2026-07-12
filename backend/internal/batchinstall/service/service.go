// Package service backs the optional batch-scan traceability feature: resolving
// a scanned batch QR, and recording which batch a mechanic actually fitted to a
// car. Everything here is an additive log — it never mutates stock or the sale.
package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/batchinstall/dto"
	"github.com/cakeru/autostock/internal/domain"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// batchNoExpr renders the human batch label (B-2026-0066) in SQL, matching the
// format used everywhere else (StockHistory, batch list).
const batchNoExpr = `'B-' || to_char(b.received_at, 'YYYY') || '-' || lpad(b.id::text, 4, '0')`

// parseCode turns a scanned QR payload into a batch id. Accepts "KSB:<id>", a
// bare numeric id, or the human "B-YYYY-NNNN" label (the trailing number is the
// zero-padded id).
func parseCode(code string) (int64, bool) {
	s := strings.TrimSpace(code)
	s = strings.TrimPrefix(strings.ToUpper(s), "KSB:")
	if id, err := strconv.ParseInt(s, 10, 64); err == nil && id > 0 {
		return id, true
	}
	// B-YYYY-NNNN → take the last dash-separated group as the id.
	if parts := strings.Split(s, "-"); len(parts) >= 2 {
		if id, err := strconv.ParseInt(strings.TrimLeft(parts[len(parts)-1], "0"), 10, 64); err == nil && id > 0 {
			return id, true
		}
	}
	return 0, false
}

func (s *Service) ResolveCode(ctx context.Context, branchID int64, code string) (*dto.BatchInfo, error) {
	id, ok := parseCode(code)
	if !ok {
		return nil, &domain.AppError{Code: "INVALID_CODE", Message: "That doesn't look like a batch label", Status: 400}
	}
	var bi dto.BatchInfo
	err := s.pool.QueryRow(ctx, `
		SELECT b.id, `+batchNoExpr+`, b.product_id, p.name,
		       COALESCE(p.tire_size,''), COALESCE(b.dot_code,''), COALESCE(b.supplier,''),
		       b.quantity_remaining, b.received_at
		FROM batches b JOIN products p ON p.id = b.product_id
		WHERE b.id = $1 AND b.branch_id = $2`, id, branchID).
		Scan(&bi.BatchID, &bi.BatchNo, &bi.ProductID, &bi.ProductName,
			&bi.TireSize, &bi.DOTCode, &bi.Supplier, &bi.QuantityRemaining, &bi.ReceivedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, &domain.AppError{Code: "NOT_FOUND", Message: "No batch matches that label in this branch", Status: 404}
		}
		return nil, fmt.Errorf("resolve batch: %w", err)
	}
	return &bi, nil
}

func (s *Service) RecordInstall(ctx context.Context, branchID, userID int64, req *dto.RecordInstallRequest) (*dto.InstallResponse, error) {
	// Verify the batch is real and in this branch, and pull the vehicle from the
	// chosen job (the scan is anchored on an open job, not entered by hand).
	var productID int64
	if err := s.pool.QueryRow(ctx,
		`SELECT product_id FROM batches WHERE id = $1 AND branch_id = $2`,
		req.BatchID, branchID).Scan(&productID); err != nil {
		if err == pgx.ErrNoRows {
			return nil, &domain.AppError{Code: "NOT_FOUND", Message: "Batch not found", Status: 404}
		}
		return nil, fmt.Errorf("check batch: %w", err)
	}

	var vehicleID *int64
	if err := s.pool.QueryRow(ctx,
		`SELECT vehicle_id FROM service_jobs WHERE id = $1 AND branch_id = $2`,
		req.ServiceJobID, branchID).Scan(&vehicleID); err != nil {
		if err == pgx.ErrNoRows {
			return nil, &domain.AppError{Code: "NOT_FOUND", Message: "Job not found", Status: 404}
		}
		return nil, fmt.Errorf("check job: %w", err)
	}

	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO batch_installs
		    (branch_id, batch_id, product_id, vehicle_id, service_job_id, position, note, installed_by, mechanic_employee_id)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6,''), NULLIF($7,''), $8, $9)
		RETURNING id`,
		branchID, req.BatchID, productID, vehicleID, req.ServiceJobID,
		req.Position, req.Note, userID, req.MechanicEmployeeID).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("record install: %w", err)
	}
	return s.getInstall(ctx, branchID, id)
}

func (s *Service) getInstall(ctx context.Context, branchID, id int64) (*dto.InstallResponse, error) {
	var r dto.InstallResponse
	err := s.pool.QueryRow(ctx, `
		SELECT bi.id, bi.batch_id, `+batchNoExpr+`, p.name, COALESCE(p.tire_size,''), COALESCE(b.dot_code,''),
		       bi.vehicle_id, COALESCE(v.plate_number,''), bi.service_job_id, COALESCE(sj.job_number,''),
		       COALESCE(bi.position,''), COALESCE(bi.note,''),
		       COALESCE(e.name,''), COALESCE(u.full_name,''), bi.installed_at
		FROM batch_installs bi
		JOIN batches b ON b.id = bi.batch_id
		JOIN products p ON p.id = bi.product_id
		LEFT JOIN vehicles v ON v.id = bi.vehicle_id
		LEFT JOIN service_jobs sj ON sj.id = bi.service_job_id
		LEFT JOIN employees e ON e.id = bi.mechanic_employee_id
		LEFT JOIN users u ON u.id = bi.installed_by
		WHERE bi.id = $1 AND bi.branch_id = $2`, id, branchID).
		Scan(&r.ID, &r.BatchID, &r.BatchNo, &r.ProductName, &r.TireSize, &r.DOTCode,
			&r.VehicleID, &r.PlateNumber, &r.ServiceJobID, &r.JobNumber,
			&r.Position, &r.Note, &r.MechanicName, &r.InstalledByName, &r.InstalledAt)
	if err != nil {
		return nil, fmt.Errorf("load install: %w", err)
	}
	return &r, nil
}

// ListForVehicle returns the batches actually fitted to a car, newest first —
// feeds the vehicle service timeline and sharpens the recall list.
func (s *Service) ListForVehicle(ctx context.Context, branchID, vehicleID int64) ([]dto.InstallResponse, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT bi.id, bi.batch_id, `+batchNoExpr+`, p.name, COALESCE(p.tire_size,''), COALESCE(b.dot_code,''),
		       bi.vehicle_id, COALESCE(v.plate_number,''), bi.service_job_id, COALESCE(sj.job_number,''),
		       COALESCE(bi.position,''), COALESCE(bi.note,''),
		       COALESCE(e.name,''), COALESCE(u.full_name,''), bi.installed_at
		FROM batch_installs bi
		JOIN batches b ON b.id = bi.batch_id
		JOIN products p ON p.id = bi.product_id
		LEFT JOIN vehicles v ON v.id = bi.vehicle_id
		LEFT JOIN service_jobs sj ON sj.id = bi.service_job_id
		LEFT JOIN employees e ON e.id = bi.mechanic_employee_id
		LEFT JOIN users u ON u.id = bi.installed_by
		WHERE bi.vehicle_id = $1 AND bi.branch_id = $2
		ORDER BY bi.installed_at DESC, bi.id DESC`, vehicleID, branchID)
	if err != nil {
		return nil, fmt.Errorf("list installs: %w", err)
	}
	defer rows.Close()

	out := []dto.InstallResponse{}
	for rows.Next() {
		var r dto.InstallResponse
		if err := rows.Scan(&r.ID, &r.BatchID, &r.BatchNo, &r.ProductName, &r.TireSize, &r.DOTCode,
			&r.VehicleID, &r.PlateNumber, &r.ServiceJobID, &r.JobNumber,
			&r.Position, &r.Note, &r.MechanicName, &r.InstalledByName, &r.InstalledAt); err != nil {
			return nil, fmt.Errorf("scan install: %w", err)
		}
		out = append(out, r)
	}
	return out, nil
}

// OpenJobs lists jobs still being worked (pending/in_progress) so the mechanic
// can attach a scan to the car currently on the ramp.
func (s *Service) OpenJobs(ctx context.Context, branchID int64) ([]dto.OpenJob, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT sj.id, sj.job_number, sj.status, sj.vehicle_id,
		       COALESCE(v.plate_number,''), COALESCE(v.make,''), COALESCE(v.model,''), COALESCE(c.name,'')
		FROM service_jobs sj
		LEFT JOIN vehicles v ON v.id = sj.vehicle_id
		LEFT JOIN customers c ON c.id = sj.customer_id
		WHERE sj.branch_id = $1 AND sj.status IN ('pending', 'in_progress')
		ORDER BY sj.created_at DESC`, branchID)
	if err != nil {
		return nil, fmt.Errorf("list open jobs: %w", err)
	}
	defer rows.Close()

	out := []dto.OpenJob{}
	for rows.Next() {
		var j dto.OpenJob
		if err := rows.Scan(&j.ID, &j.JobNumber, &j.Status, &j.VehicleID,
			&j.PlateNumber, &j.Make, &j.Model, &j.CustomerName); err != nil {
			return nil, fmt.Errorf("scan open job: %w", err)
		}
		out = append(out, j)
	}
	return out, nil
}

// Mechanics is a minimal active-employee list for the "who fitted it" picker —
// exposed under the scan permission so a non-sales mechanic device can read it.
func (s *Service) Mechanics(ctx context.Context, branchID int64) ([]dto.Mechanic, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name FROM employees
		WHERE branch_id = $1 AND is_active = true
		ORDER BY name`, branchID)
	if err != nil {
		return nil, fmt.Errorf("list mechanics: %w", err)
	}
	defer rows.Close()

	out := []dto.Mechanic{}
	for rows.Next() {
		var m dto.Mechanic
		if err := rows.Scan(&m.ID, &m.Name); err != nil {
			return nil, fmt.Errorf("scan mechanic: %w", err)
		}
		out = append(out, m)
	}
	return out, nil
}
