package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/customer/dto"
	"github.com/cakeru/autostock/internal/customer/models"
	"github.com/cakeru/autostock/internal/domain"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) List(ctx context.Context, branchID int64, filter dto.CustomerFilter) ([]dto.CustomerResponse, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 20
	}
	offset := (filter.Page - 1) * filter.PerPage

	rows, err := s.pool.Query(ctx, `
		SELECT c.id, c.name, c.customer_type, COALESCE(c.phone,''), COALESCE(c.email,''), COALESCE(c.address,''), COALESCE(c.notes,''),
		       c.customer_since, c.is_active, c.created_at, c.updated_at,
		       (SELECT COUNT(*) FROM vehicles v WHERE v.customer_id = c.id) as vehicle_count,
		       COALESCE((SELECT string_agg(v.plate_number, ', ' ORDER BY v.created_at) FROM vehicles v WHERE v.customer_id = c.id), '') as plates,
		       COALESCE((SELECT SUM(i.total_usd) FROM invoices i WHERE i.customer_id = c.id AND i.status <> 'voided'), 0) as total_spent,
		       (SELECT MAX(COALESCE(i.issued_at, i.created_at)) FROM invoices i WHERE i.customer_id = c.id AND i.status <> 'voided') as last_visit
		FROM customers c
		WHERE c.branch_id = $1 AND c.is_active = true
		  AND ($2::text IS NULL OR c.name ILIKE '%' || $2 || '%')
		  AND ($3::text IS NULL OR c.phone = $3)
		  AND ($4::text IS NULL OR c.email = $4)
		  AND ($5::text IS NULL OR
		       c.name ILIKE '%' || $5 || '%' OR c.phone ILIKE '%' || $5 || '%'
		       OR EXISTS (SELECT 1 FROM vehicles v WHERE v.customer_id = c.id
		                  AND (v.plate_number ILIKE '%' || $5 || '%' OR v.make ILIKE '%' || $5 || '%' OR v.model ILIKE '%' || $5 || '%')))
		ORDER BY c.created_at DESC
		LIMIT $6 OFFSET $7`,
		branchID,
		nullIfEmpty(filter.NameLike),
		nullIfEmpty(filter.Phone),
		nullIfEmpty(filter.Email),
		nullIfEmpty(filter.Search),
		filter.PerPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query customers: %w", err)
	}
	defer rows.Close()

	var customers []dto.CustomerResponse
	for rows.Next() {
		var c dto.CustomerResponse
		var customerSince, lastVisit *time.Time
		if err := rows.Scan(&c.ID, &c.Name, &c.CustomerType, &c.Phone, &c.Email,
			&c.Address, &c.Notes, &customerSince, &c.IsActive,
			&c.CreatedAt, &c.UpdatedAt, &c.VehicleCount, &c.VehiclePlates, &c.TotalSpent, &lastVisit); err != nil {
			return nil, 0, fmt.Errorf("scan customer: %w", err)
		}
		if customerSince != nil {
			c.CustomerSince = customerSince.Format("2006-01-02")
		}
		if lastVisit != nil {
			c.LastVisit = lastVisit.Format("2006-01-02")
		}
		customers = append(customers, c)
	}

	var total int
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM customers c
		WHERE c.branch_id = $1 AND c.is_active = true
		  AND ($2::text IS NULL OR c.name ILIKE '%' || $2 || '%')
		  AND ($3::text IS NULL OR c.phone = $3)
		  AND ($4::text IS NULL OR c.email = $4)
		  AND ($5::text IS NULL OR
		       c.name ILIKE '%' || $5 || '%' OR c.phone ILIKE '%' || $5 || '%'
		       OR EXISTS (SELECT 1 FROM vehicles v WHERE v.customer_id = c.id
		                  AND (v.plate_number ILIKE '%' || $5 || '%' OR v.make ILIKE '%' || $5 || '%' OR v.model ILIKE '%' || $5 || '%')))`,
		branchID,
		nullIfEmpty(filter.NameLike),
		nullIfEmpty(filter.Phone),
		nullIfEmpty(filter.Email),
		nullIfEmpty(filter.Search)).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count customers: %w", err)
	}

	if customers == nil {
		customers = []dto.CustomerResponse{}
	}
	return customers, total, nil
}

func (s *Service) Get(ctx context.Context, branchID int64, id int64) (*dto.CustomerResponse, []models.Vehicle, error) {
	var c dto.CustomerResponse
	var customerSince *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, customer_type, COALESCE(phone,''), COALESCE(email,''), COALESCE(address,''), COALESCE(notes,''),
		       customer_since, is_active, created_at, updated_at
		FROM customers WHERE id = $1 AND branch_id = $2 AND is_active = true`, id, branchID).
		Scan(&c.ID, &c.Name, &c.CustomerType, &c.Phone, &c.Email, &c.Address,
			&c.Notes, &customerSince, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, domain.ErrNotFound
		}
		return nil, nil, fmt.Errorf("get customer: %w", err)
	}
	if customerSince != nil {
		c.CustomerSince = customerSince.Format("2006-01-02")
	}

	vehicles, err := s.ListVehicles(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	c.VehicleCount = len(vehicles)

	return &c, vehicles, nil
}

func (s *Service) Create(ctx context.Context, branchID int64, req *dto.CreateCustomerRequest) (*dto.CustomerResponse, error) {
	var c dto.CustomerResponse
	var customerSince *time.Time
	err := s.pool.QueryRow(ctx, `
		INSERT INTO customers (branch_id, name, customer_type, phone, email, address, notes, customer_since)
		VALUES ($1, $2, COALESCE(NULLIF($3, ''), 'retail'), $4, $5, $6, $7, CURRENT_DATE)
		RETURNING id, name, customer_type, COALESCE(phone,''), COALESCE(email,''), COALESCE(address,''), COALESCE(notes,''), customer_since, is_active, created_at, updated_at`,
		branchID, req.Name, req.CustomerType, req.Phone, req.Email, req.Address, req.Notes).
		Scan(&c.ID, &c.Name, &c.CustomerType, &c.Phone, &c.Email, &c.Address,
			&c.Notes, &customerSince, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create customer: %w", err)
	}
	if customerSince != nil {
		c.CustomerSince = customerSince.Format("2006-01-02")
	}
	return &c, nil
}

func (s *Service) Update(ctx context.Context, branchID int64, id int64, req *dto.UpdateCustomerRequest) (*dto.CustomerResponse, error) {
	var c dto.CustomerResponse
	var customerSince *time.Time
	err := s.pool.QueryRow(ctx, `
		UPDATE customers
		SET name = COALESCE(NULLIF($1, ''), name),
		    customer_type = COALESCE(NULLIF($2, ''), customer_type),
		    phone = COALESCE($3, phone),
		    email = COALESCE($4, email),
		    address = COALESCE($5, address),
		    notes = COALESCE($6, notes),
		    updated_at = NOW()
		WHERE id = $7 AND branch_id = $8 AND is_active = true
		RETURNING id, name, customer_type, COALESCE(phone,''), COALESCE(email,''), COALESCE(address,''), COALESCE(notes,''), customer_since, is_active, created_at, updated_at`,
		req.Name, req.CustomerType, req.Phone, req.Email, req.Address, req.Notes, id, branchID).
		Scan(&c.ID, &c.Name, &c.CustomerType, &c.Phone, &c.Email, &c.Address,
			&c.Notes, &customerSince, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("update customer: %w", err)
	}
	if customerSince != nil {
		c.CustomerSince = customerSince.Format("2006-01-02")
	}
	return &c, nil
}

func (s *Service) Delete(ctx context.Context, branchID int64, id int64) error {
	result, err := s.pool.Exec(ctx, `UPDATE customers SET is_active = false, updated_at = NOW() WHERE id = $1 AND branch_id = $2 AND is_active = true`, id, branchID)
	if err != nil {
		return fmt.Errorf("delete customer: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Service) ListVehicles(ctx context.Context, customerID int64) ([]models.Vehicle, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, customer_id, plate_number, COALESCE(make,''), COALESCE(model,''), COALESCE(year, 0), COALESCE(vin,''), COALESCE(color,''), COALESCE(body_type,''), distance_unit, COALESCE(notes,''), created_at, updated_at
		FROM vehicles WHERE customer_id = $1 ORDER BY created_at DESC`, customerID)
	if err != nil {
		return nil, fmt.Errorf("query vehicles: %w", err)
	}
	defer rows.Close()

	var vehicles []models.Vehicle
	for rows.Next() {
		var v models.Vehicle
		if err := rows.Scan(&v.ID, &v.CustomerID, &v.PlateNumber, &v.Make, &v.Model,
			&v.Year, &v.VIN, &v.Color, &v.BodyType, &v.DistanceUnit, &v.Notes, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan vehicle: %w", err)
		}
		vehicles = append(vehicles, v)
	}
	if vehicles == nil {
		vehicles = []models.Vehicle{}
	}
	return vehicles, nil
}

func (s *Service) GetVehicle(ctx context.Context, id int64) (*models.Vehicle, error) {
	var v models.Vehicle
	err := s.pool.QueryRow(ctx, `
		SELECT id, customer_id, plate_number, COALESCE(make,''), COALESCE(model,''), COALESCE(year, 0), COALESCE(vin,''), COALESCE(color,''), COALESCE(body_type,''), distance_unit, COALESCE(notes,''), created_at, updated_at
		FROM vehicles WHERE id = $1`, id).
		Scan(&v.ID, &v.CustomerID, &v.PlateNumber, &v.Make, &v.Model,
			&v.Year, &v.VIN, &v.Color, &v.BodyType, &v.DistanceUnit, &v.Notes, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get vehicle: %w", err)
	}
	return &v, nil
}

func (s *Service) CreateVehicle(ctx context.Context, customerID int64, req *dto.CreateVehicleRequest) (*models.Vehicle, error) {
	var branchID int64
	if err := s.pool.QueryRow(ctx, `SELECT branch_id FROM customers WHERE id = $1`, customerID).Scan(&branchID); err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("look up customer branch: %w", err)
	}

	var v models.Vehicle
	err := s.pool.QueryRow(ctx, `
		INSERT INTO vehicles (branch_id, customer_id, plate_number, make, model, year, vin, color, body_type, distance_unit, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULLIF($9, ''), COALESCE(NULLIF($10, ''), 'km'), $11)
		RETURNING id, customer_id, plate_number, COALESCE(make,''), COALESCE(model,''), COALESCE(year, 0), COALESCE(vin,''), COALESCE(color,''), COALESCE(body_type,''), distance_unit, COALESCE(notes,''), created_at, updated_at`,
		branchID, customerID, req.PlateNumber, req.Make, req.Model, req.Year, req.VIN, req.Color, req.BodyType, req.DistanceUnit, req.Notes).
		Scan(&v.ID, &v.CustomerID, &v.PlateNumber, &v.Make, &v.Model,
			&v.Year, &v.VIN, &v.Color, &v.BodyType, &v.DistanceUnit, &v.Notes, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create vehicle: %w", err)
	}
	return &v, nil
}

func (s *Service) UpdateVehicle(ctx context.Context, id int64, req *dto.UpdateVehicleRequest) (*models.Vehicle, error) {
	var v models.Vehicle
	err := s.pool.QueryRow(ctx, `
		UPDATE vehicles
		SET plate_number = COALESCE(NULLIF($1, ''), plate_number),
		    make = COALESCE($2, make),
		    model = COALESCE($3, model),
		    year = COALESCE($4, year),
		    vin = COALESCE($5, vin),
		    color = COALESCE($6, color),
		    body_type = COALESCE($7, body_type),
		    distance_unit = COALESCE(NULLIF($8, ''), distance_unit),
		    notes = COALESCE($9, notes),
		    updated_at = NOW()
		WHERE id = $10
		RETURNING id, customer_id, plate_number, COALESCE(make,''), COALESCE(model,''), COALESCE(year, 0), COALESCE(vin,''), COALESCE(color,''), COALESCE(body_type,''), distance_unit, COALESCE(notes,''), created_at, updated_at`,
		req.PlateNumber, req.Make, req.Model, req.Year, req.VIN, req.Color, req.BodyType, req.DistanceUnit, req.Notes, id).
		Scan(&v.ID, &v.CustomerID, &v.PlateNumber, &v.Make, &v.Model,
			&v.Year, &v.VIN, &v.Color, &v.BodyType, &v.DistanceUnit, &v.Notes, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("update vehicle: %w", err)
	}
	return &v, nil
}

func (s *Service) DeleteVehicle(ctx context.Context, id int64) error {
	result, err := s.pool.Exec(ctx, `DELETE FROM vehicles WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete vehicle: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// GetStats computes the value metrics for a customer's profile header from
// their non-voided invoices (the source of truth for money).
func (s *Service) GetStats(ctx context.Context, customerID int64) (*dto.CustomerStats, error) {
	var stats dto.CustomerStats
	var lastVisit *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(total_usd), 0),
		       COUNT(*),
		       MAX(COALESCE(issued_at, created_at)),
		       COALESCE(SUM(CASE WHEN payment_status IN ('unpaid','partial') THEN total_usd - paid_amount ELSE 0 END), 0)
		FROM invoices
		WHERE customer_id = $1 AND status <> 'voided'`, customerID).
		Scan(&stats.TotalSpent, &stats.VisitCount, &lastVisit, &stats.Outstanding)
	if err != nil {
		return nil, fmt.Errorf("customer stats: %w", err)
	}
	if lastVisit != nil {
		stats.LastVisit = lastVisit.Format("2006-01-02")
	}
	return &stats, nil
}

// GetActivity returns a unified timeline of the customer's service jobs and
// invoices, newest first.
func (s *Service) GetActivity(ctx context.Context, customerID int64) ([]dto.ActivityItem, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT * FROM (
			SELECT 'job' AS type, sj.id, sj.job_number AS ref,
			       COALESCE(sj.completed_at, sj.created_at) AS date,
			       COALESCE(NULLIF(sj.description, ''), 'Service job') AS title,
			       sj.status,
			       COALESCE(SUM(sji.quantity * sji.unit_price), 0)::numeric(10,2) AS amount,
			       0::numeric(10,2) AS outstanding,
			       COALESCE(v.plate_number, '') AS plate
			FROM service_jobs sj
			LEFT JOIN service_job_items sji ON sji.service_job_id = sj.id
			LEFT JOIN vehicles v ON v.id = sj.vehicle_id
			WHERE sj.customer_id = $1
			GROUP BY sj.id, v.plate_number
			UNION ALL
			SELECT 'invoice', i.id, i.invoice_number,
			       COALESCE(i.issued_at, i.created_at),
			       'Sale', i.status,
			       i.total_usd,
			       CASE WHEN i.payment_status IN ('unpaid','partial') THEN (i.total_usd - i.paid_amount) ELSE 0 END,
			       COALESCE(v.plate_number, '')
			FROM invoices i
			LEFT JOIN vehicles v ON v.id = i.vehicle_id
			WHERE i.customer_id = $1 AND i.status <> 'voided'
		) t
		ORDER BY date DESC NULLS LAST
		LIMIT 50`, customerID)
	if err != nil {
		return nil, fmt.Errorf("query activity: %w", err)
	}
	defer rows.Close()

	items := []dto.ActivityItem{}
	for rows.Next() {
		var a dto.ActivityItem
		if err := rows.Scan(&a.Type, &a.ID, &a.Ref, &a.Date, &a.Title, &a.Status, &a.Amount, &a.Outstanding, &a.Plate); err != nil {
			return nil, fmt.Errorf("scan activity: %w", err)
		}
		items = append(items, a)
	}
	return items, nil
}

func (s *Service) GetServiceHistory(ctx context.Context, customerID int64) ([]dto.ServiceHistoryItem, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT sj.id, sj.job_number, sj.status, COALESCE(sj.description,''), COALESCE(sj.work_performed,''), sj.completed_at,
		       COALESCE(SUM(sji.quantity * sji.unit_price), 0)::decimal(10,2) as total_amount
		FROM service_jobs sj
		LEFT JOIN service_job_items sji ON sji.service_job_id = sj.id
		WHERE sj.customer_id = $1
		GROUP BY sj.id
		ORDER BY sj.created_at DESC`, customerID)
	if err != nil {
		return nil, fmt.Errorf("query service history: %w", err)
	}
	defer rows.Close()

	var items []dto.ServiceHistoryItem
	for rows.Next() {
		var item dto.ServiceHistoryItem
		if err := rows.Scan(&item.ID, &item.JobNumber, &item.Status, &item.Description,
			&item.WorkPerformed, &item.CompletedAt, &item.TotalAmount); err != nil {
			return nil, fmt.Errorf("scan service history: %w", err)
		}
		items = append(items, item)
	}
	if items == nil {
		items = []dto.ServiceHistoryItem{}
	}
	return items, nil
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
