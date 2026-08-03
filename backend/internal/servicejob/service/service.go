package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/servicejob/dto"
	telegrammodels "github.com/cakeru/autostock/internal/telegram/models"
	telegram "github.com/cakeru/autostock/internal/telegram/service"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) List(ctx context.Context, branchID int64, filter dto.ServiceJobFilter) ([]dto.ServiceJobListResponse, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 20
	}
	offset := (filter.Page - 1) * filter.PerPage

	var customerID *int64
	if filter.CustomerID > 0 {
		customerID = &filter.CustomerID
	}
	var assignedTo *int64
	if filter.AssignedTo > 0 {
		assignedTo = &filter.AssignedTo
	}

	rows, err := s.pool.Query(ctx, `
		SELECT sj.id, sj.branch_id, sj.job_number, sj.status, sj.priority,
		       sj.customer_id, COALESCE(c.name, ''), COALESCE(c.phone, ''),
		       sj.vehicle_id,
		       CASE WHEN v.id IS NOT NULL THEN COALESCE(v.plate_number, '') ELSE '' END,
		       CASE WHEN v.id IS NOT NULL THEN CONCAT_WS(' ', COALESCE(v.make,''), COALESCE(v.model,''), COALESCE(v.year::text,'')) ELSE '' END,
		       COALESCE(sj.description, ''), sj.invoice_id, sj.scheduled_at,
		       sj.assigned_to, COALESCE(asg.name, ''),
		       sj.quote_approved_at, sj.quote_approved_by,
		       sj.created_by, COALESCE(cred.full_name, ''),
		       sj.created_at, sj.updated_at,
		       COUNT(*) OVER() as total_count
		FROM service_jobs sj
		LEFT JOIN customers c ON c.id = sj.customer_id
		LEFT JOIN vehicles v ON v.id = sj.vehicle_id
		LEFT JOIN users cred ON cred.id = sj.created_by
		LEFT JOIN employees asg ON asg.id = sj.assigned_to
		WHERE sj.branch_id = $1
		  AND ($2::text IS NULL OR sj.status = $2)
		  AND ($3::bigint IS NULL OR sj.customer_id = $3)
		  AND ($4::text IS NULL OR (sj.scheduled_at IS NOT NULL AND sj.scheduled_at >= $4::timestamptz))
		  AND ($5::bigint IS NULL OR sj.assigned_to = $5)
		ORDER BY (CASE WHEN $4::text IS NOT NULL THEN sj.scheduled_at END) ASC NULLS LAST, sj.created_at DESC
		LIMIT $6 OFFSET $7`,
		branchID, nullIfEmpty(filter.Status), customerID, nullIfEmpty(filter.ScheduledFrom), assignedTo, filter.PerPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query service jobs: %w", err)
	}
	defer rows.Close()

	var jobs []dto.ServiceJobListResponse
	var total int
	for rows.Next() {
		var j dto.ServiceJobListResponse
		if err := rows.Scan(&j.ID, &j.BranchID, &j.JobNumber, &j.Status, &j.Priority,
			&j.CustomerID, &j.CustomerName, &j.CustomerPhone,
			&j.VehicleID, &j.PlateNumber, &j.VehicleInfo,
			&j.Description, &j.InvoiceID, &j.ScheduledAt,
			&j.AssignedToID, &j.AssignedToName,
			&j.QuoteApprovedAt, &j.QuoteApprovedBy,
			&j.CreatedByID, &j.CreatedByName,
			&j.CreatedAt, &j.UpdatedAt,
			&total); err != nil {
			return nil, 0, fmt.Errorf("scan job: %w", err)
		}
		jobs = append(jobs, j)
	}

	if jobs == nil {
		jobs = []dto.ServiceJobListResponse{}
	}
	return jobs, total, nil
}

func (s *Service) Get(ctx context.Context, branchID int64, id int64) (*dto.ServiceJobDetailResponse, error) {
	var j dto.ServiceJobDetailResponse
	err := s.pool.QueryRow(ctx, `
		SELECT sj.id, sj.branch_id, sj.job_number, sj.status, sj.priority,
		       sj.customer_id, COALESCE(c.name, ''), COALESCE(c.phone, ''),
		       sj.vehicle_id,
		       CASE WHEN v.id IS NOT NULL THEN COALESCE(v.plate_number, '') ELSE '' END,
		       CASE WHEN v.id IS NOT NULL THEN CONCAT_WS(' ', COALESCE(v.make,''), COALESCE(v.model,''), COALESCE(v.year::text,'')) ELSE '' END,
		       sj.mileage, sj.mileage_unit,
		       sj.description, COALESCE(sj.diagnosis, ''), COALESCE(sj.work_performed, ''),
		       sj.estimated_hours, sj.actual_hours, sj.started_at, sj.completed_at,
		       sj.invoice_id, COALESCE(sj.notes, ''), sj.scheduled_at,
		       sj.assigned_to, COALESCE(asg.name, ''),
		       sj.quote_approved_at, sj.quote_approved_by,
		       COALESCE(sj.discount, 0),
		       sj.created_by, COALESCE(cred.full_name, ''),
		       sj.created_at, sj.updated_at
		FROM service_jobs sj
		LEFT JOIN customers c ON c.id = sj.customer_id
		LEFT JOIN vehicles v ON v.id = sj.vehicle_id
		LEFT JOIN users cred ON cred.id = sj.created_by
		LEFT JOIN employees asg ON asg.id = sj.assigned_to
		WHERE sj.id = $1 AND sj.branch_id = $2`, id, branchID).
		Scan(&j.ID, &j.BranchID, &j.JobNumber, &j.Status, &j.Priority,
			&j.CustomerID, &j.CustomerName, &j.CustomerPhone,
			&j.VehicleID, &j.PlateNumber, &j.VehicleInfo, &j.Mileage, &j.MileageUnit,
			&j.Description, &j.Diagnosis, &j.WorkPerformed,
			&j.EstimatedHours, &j.ActualHours, &j.StartedAt, &j.CompletedAt,
			&j.InvoiceID, &j.Notes, &j.ScheduledAt,
			&j.AssignedToID, &j.AssignedToName,
			&j.QuoteApprovedAt, &j.QuoteApprovedBy, &j.Discount,
			&j.CreatedByID, &j.CreatedByName,
			&j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get job: %w", err)
	}

	items, err := s.ListItems(ctx, id)
	if err != nil {
		return nil, err
	}
	j.Items = items

	var total float64
	err = s.pool.QueryRow(ctx, `SELECT COALESCE(SUM(total_price), 0)::decimal(10,2) FROM service_job_items WHERE service_job_id = $1`, id).Scan(&total)
	if err == nil {
		j.TotalAmount = total
	}

	if j.Items == nil {
		j.Items = []dto.ServiceJobItemResponse{}
	}
	return &j, nil
}

func (s *Service) Create(ctx context.Context, branchID int64, userID int64, req *dto.CreateServiceJobRequest) (*dto.ServiceJobListResponse, error) {
	priority := req.Priority
	if priority == "" {
		priority = "normal"
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext('job_number'))`)
	if err != nil {
		return nil, fmt.Errorf("advisory lock: %w", err)
	}

	year := fmt.Sprintf("%d", time.Now().Year())
	var seq int
	err = tx.QueryRow(ctx,
		`SELECT COALESCE(MAX(SUBSTRING(job_number FROM 10)::int), 0) + 1
		 FROM service_jobs WHERE job_number LIKE $1`, "JOB-"+year+"-%").Scan(&seq)
	if err != nil {
		return nil, fmt.Errorf("generate job number: %w", err)
	}
	jobNumber := fmt.Sprintf("JOB-%s-%04d", year, seq)

	mileageUnit := "km"
	if req.VehicleID != nil {
		_ = s.pool.QueryRow(ctx, `SELECT distance_unit FROM vehicles WHERE id = $1`, *req.VehicleID).Scan(&mileageUnit)
		if mileageUnit != "mi" {
			mileageUnit = "km"
		}
	}

	var j dto.ServiceJobListResponse
	err = tx.QueryRow(ctx, `
		INSERT INTO service_jobs (branch_id, job_number, customer_id, vehicle_id, mileage, mileage_unit, description, priority, estimated_hours, discount, notes, created_by, scheduled_at, assigned_to)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NULLIF($13, '')::timestamptz, $14)
		RETURNING id, branch_id, job_number, status, priority, COALESCE(description,''), created_at, updated_at, created_by`,
		branchID, jobNumber, req.CustomerID, req.VehicleID, req.Mileage, mileageUnit, req.Description, priority, req.EstimatedHours, req.Discount, req.Notes, userID, req.ScheduledAt, req.AssignedTo).
		Scan(&j.ID, &j.BranchID, &j.JobNumber, &j.Status, &j.Priority, &j.Description, &j.CreatedAt, &j.UpdatedAt, &j.CreatedByID)
	if err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}

	// Optional line items (e.g. when a POS cart is saved as a job). Product
	// lines reserve stock so the units this job is promising can't be sold
	// out from under it by the POS or another job before it's invoiced.
	for _, it := range req.Items {
		itemType := normalizeItemType(it.ItemType, it.ProductID)
		if _, err := tx.Exec(ctx, `
			INSERT INTO service_job_items (service_job_id, product_id, item_type, description, quantity, unit_price, total_price, vehicle_event_type)
			VALUES ($1, $2, $3, $4, $5, $6, $5::numeric * $6::numeric, $7)`,
			j.ID, it.ProductID, itemType, it.Description, it.Quantity, it.UnitPrice, it.VehicleEventType); err != nil {
			return nil, fmt.Errorf("add job item: %w", err)
		}
		if itemType == "product" && it.ProductID != nil {
			if err := reserveStock(ctx, tx, branchID, *it.ProductID, it.Quantity); err != nil {
				return nil, err
			}
		}
	}

	var customerName, vehicleInfo, assignedToName string
	_ = tx.QueryRow(ctx, `
		SELECT COALESCE(c.name, ''),
		       CASE WHEN v.id IS NOT NULL THEN CONCAT_WS(' ', v.make, v.model, v.year::text) ELSE '' END,
		       COALESCE(e.name, '')
		FROM (SELECT 1) x
		LEFT JOIN customers c ON c.id = $1::bigint
		LEFT JOIN vehicles v ON v.id = $2::bigint
		LEFT JOIN employees e ON e.id = $3::bigint`,
		req.CustomerID, req.VehicleID, req.AssignedTo).Scan(&customerName, &vehicleInfo, &assignedToName)

	if err := telegram.LogEvent(ctx, tx, branchID, telegrammodels.TopicJobs, "job_created", "service_job", j.ID, map[string]any{
		"job_number": j.JobNumber, "customer_name": customerName, "vehicle_info": vehicleInfo,
		"description": j.Description, "assigned_to_name": assignedToName, "status": j.Status,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	if j.CreatedByID != nil {
		_ = s.pool.QueryRow(ctx, `SELECT full_name FROM users WHERE id = $1`, *j.CreatedByID).Scan(&j.CreatedByName)
	}
	return &j, nil
}

var allowedTransitions = map[string][]string{
	"pending":     {"in_progress", "cancelled"},
	"in_progress": {"completed", "cancelled"},
	"completed":   {},
	"cancelled":   {},
}

func (s *Service) Update(ctx context.Context, branchID int64, id int64, req *dto.UpdateServiceJobRequest) (*dto.ServiceJobDetailResponse, error) {
	current, err := s.Get(ctx, branchID, id)
	if err != nil {
		return nil, err
	}

	if req.Status != nil && *req.Status != current.Status {
		valid, ok := allowedTransitions[current.Status]
		if !ok {
			return nil, &domain.AppError{Code: "INVALID_TRANSITION", Message: fmt.Sprintf("No transitions allowed from '%s'", current.Status), Status: 400}
		}
		found := false
		for _, s := range valid {
			if s == *req.Status {
				found = true
				break
			}
		}
		if !found {
			return nil, &domain.AppError{Code: "INVALID_TRANSITION", Message: fmt.Sprintf("Cannot transition from '%s' to '%s'", current.Status, *req.Status), Status: 400}
		}
	}

	var startedAt, completedAt *time.Time
	if req.StartedAt != nil && *req.StartedAt != "" {
		t, err := time.Parse(time.RFC3339, *req.StartedAt)
		if err != nil {
			return nil, fmt.Errorf("parse started_at: %w", err)
		}
		startedAt = &t
	}
	if req.CompletedAt != nil && *req.CompletedAt != "" {
		t, err := time.Parse(time.RFC3339, *req.CompletedAt)
		if err != nil {
			return nil, fmt.Errorf("parse completed_at: %w", err)
		}
		completedAt = &t
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Cancelling a job that never got invoiced releases every unit it was
	// holding back for the customer.
	if req.Status != nil && *req.Status == "cancelled" && current.InvoiceID == nil {
		if err := releaseJobReservations(ctx, tx, id); err != nil {
			return nil, err
		}
	}

	_, err = tx.Exec(ctx, `
		UPDATE service_jobs
		SET status = COALESCE($1, status),
		    priority = COALESCE($2, priority),
		    diagnosis = COALESCE($3, diagnosis),
		    work_performed = COALESCE($4, work_performed),
		    estimated_hours = COALESCE($5, estimated_hours),
		    actual_hours = COALESCE($6, actual_hours),
		    started_at = CASE WHEN $7::timestamptz IS NOT NULL THEN $7 ELSE started_at END,
		    completed_at = CASE WHEN $8::timestamptz IS NOT NULL THEN $8 ELSE completed_at END,
		    notes = COALESCE($9, notes),
		    scheduled_at = CASE WHEN NULLIF($12::text, '') IS NOT NULL THEN NULLIF($12::text, '')::timestamptz ELSE scheduled_at END,
		    assigned_to = COALESCE($13, assigned_to),
		    mileage = COALESCE($14, mileage),
		    updated_at = NOW()
		WHERE id = $10 AND branch_id = $11`,
		req.Status, req.Priority, req.Diagnosis, req.WorkPerformed,
		req.EstimatedHours, req.ActualHours, startedAt, completedAt, req.Notes, id, branchID, req.ScheduledAt, req.AssignedTo, req.Mileage)
	if err != nil {
		return nil, fmt.Errorf("update job: %w", err)
	}

	if req.Status != nil && *req.Status != current.Status {
		if err := telegram.LogEvent(ctx, tx, branchID, telegrammodels.TopicJobs, "job_status_changed", "service_job", id, map[string]any{
			"job_number": current.JobNumber, "customer_name": current.CustomerName, "vehicle_info": current.VehicleInfo,
			"description": current.Description, "assigned_to_name": current.AssignedToName,
			"old_status": current.Status, "new_status": *req.Status,
		}); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return s.Get(ctx, branchID, id)
}

func (s *Service) Delete(ctx context.Context, branchID int64, id int64) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// A job in a deletable state (not completed) can only have live
	// reservations, never a real deduction yet — release them all.
	if err := releaseJobReservations(ctx, tx, id); err != nil {
		return err
	}

	result, err := tx.Exec(ctx, `DELETE FROM service_jobs WHERE id = $1 AND branch_id = $2 AND (status IS NULL OR status NOT IN ('completed', 'invoiced'))`, id, branchID)
	if err != nil {
		return fmt.Errorf("delete job: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (s *Service) Complete(ctx context.Context, branchID int64, id int64) (*dto.ServiceJobDetailResponse, error) {
	current, err := s.Get(ctx, branchID, id)
	if err != nil {
		return nil, err
	}

	if current.Status != "in_progress" && current.Status != "pending" {
		return nil, &domain.AppError{Code: "INVALID_STATUS", Message: fmt.Sprintf("Cannot complete a '%s' job", current.Status), Status: 400}
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now()
	if _, err = tx.Exec(ctx, `UPDATE service_jobs SET status = 'completed', completed_at = $1, updated_at = NOW() WHERE id = $2 AND branch_id = $3`, now, id, branchID); err != nil {
		return nil, fmt.Errorf("complete job: %w", err)
	}

	if err := telegram.LogEvent(ctx, tx, branchID, telegrammodels.TopicJobs, "job_status_changed", "service_job", id, map[string]any{
		"job_number": current.JobNumber, "customer_name": current.CustomerName, "vehicle_info": current.VehicleInfo,
		"description": current.Description, "assigned_to_name": current.AssignedToName,
		"old_status": current.Status, "new_status": "completed",
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return s.Get(ctx, branchID, id)
}

func (s *Service) ApproveQuote(ctx context.Context, branchID int64, id int64, userID int64) (*dto.ServiceJobDetailResponse, error) {
	current, err := s.Get(ctx, branchID, id)
	if err != nil {
		return nil, err
	}
	if current.Status == "completed" || current.Status == "cancelled" {
		return nil, &domain.AppError{Code: "INVALID_STATUS", Message: fmt.Sprintf("Cannot approve quote on a '%s' job", current.Status), Status: 400}
	}
	if current.QuoteApprovedAt != nil {
		return nil, &domain.AppError{Code: "ALREADY_APPROVED", Message: "Quote has already been approved", Status: 400}
	}

	_, err = s.pool.Exec(ctx,
		`UPDATE service_jobs SET quote_approved_at = NOW(), quote_approved_by = $1, updated_at = NOW() WHERE id = $2 AND branch_id = $3`,
		userID, id, branchID)
	if err != nil {
		return nil, fmt.Errorf("approve quote: %w", err)
	}
	return s.Get(ctx, branchID, id)
}

func (s *Service) AddItem(ctx context.Context, branchID int64, jobID int64, req *dto.AddItemRequest) (*dto.ServiceJobItemResponse, error) {
	var status string
	err := s.pool.QueryRow(ctx, `SELECT status FROM service_jobs WHERE id = $1 AND branch_id = $2`, jobID, branchID).Scan(&status)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	itemType := normalizeItemType(req.ItemType, req.ProductID)

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Only an active job's product lines hold stock — a completed job's items
	// were already released into a real deduction when it was invoiced, and a
	// cancelled job never should have held any.
	if itemType == "product" && req.ProductID != nil && status != "completed" && status != "cancelled" {
		if err := reserveStock(ctx, tx, branchID, *req.ProductID, req.Quantity); err != nil {
			return nil, err
		}
	}

	var item dto.ServiceJobItemResponse
	err = tx.QueryRow(ctx, `
		WITH ins AS (
			INSERT INTO service_job_items (service_job_id, product_id, item_type, description, quantity, unit_price, total_price, vehicle_event_type)
			VALUES ($1, $2, $3, $4, $5, $6, $5::numeric * $6::numeric, $7)
			RETURNING id, product_id, item_type, description, quantity, unit_price, total_price
		)
		SELECT ins.id, ins.product_id, ins.item_type, COALESCE(p.name, ''),
		       COALESCE(ins.description, ''), ins.quantity, ins.unit_price, ins.total_price
		FROM ins LEFT JOIN products p ON p.id = ins.product_id`,
		jobID, req.ProductID, itemType, req.Description, req.Quantity, req.UnitPrice, req.VehicleEventType).
		Scan(&item.ID, &item.ProductID, &item.ItemType, &item.ProductName, &item.Description,
			&item.Quantity, &item.UnitPrice, &item.TotalPrice)
	if err != nil {
		return nil, fmt.Errorf("add item: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return &item, nil
}

// normalizeItemType defaults a line's type: an explicit type wins, otherwise a
// line with a product is a "product" and one without is "custom".
func normalizeItemType(t string, productID *int64) string {
	switch t {
	case "product", "labor", "fee", "custom":
		return t
	}
	if productID != nil {
		return "product"
	}
	return "custom"
}

func (s *Service) RemoveItem(ctx context.Context, branchID int64, itemID int64) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Release any reservation this line was holding before deleting it — but
	// only if the job hasn't been invoiced yet, since an invoiced job already
	// converted its reservation into a real stock deduction.
	var productID *int64
	var qty float64
	var invoiceID *int64
	err = tx.QueryRow(ctx, `
		SELECT sji.product_id, sji.quantity, sj.invoice_id
		FROM service_job_items sji
		JOIN service_jobs sj ON sj.id = sji.service_job_id
		WHERE sji.id = $1 AND sj.branch_id = $2`, itemID, branchID).
		Scan(&productID, &qty, &invoiceID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.ErrNotFound
		}
		return fmt.Errorf("find item: %w", err)
	}

	result, err := tx.Exec(ctx, `DELETE FROM service_job_items WHERE id = $1`, itemID)
	if err != nil {
		return fmt.Errorf("remove item: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	if productID != nil && invoiceID == nil {
		if err := releaseStock(ctx, tx, *productID, qty); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (s *Service) ListItems(ctx context.Context, jobID int64) ([]dto.ServiceJobItemResponse, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT sji.id, sji.product_id, sji.item_type, COALESCE(p.name, '') as product_name,
		       COALESCE(sji.description, '') as description, sji.quantity, sji.unit_price, sji.total_price
		FROM service_job_items sji
		LEFT JOIN products p ON p.id = sji.product_id
		WHERE sji.service_job_id = $1
		ORDER BY sji.created_at`, jobID)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	var items []dto.ServiceJobItemResponse
	for rows.Next() {
		var item dto.ServiceJobItemResponse
		if err := rows.Scan(&item.ID, &item.ProductID, &item.ItemType, &item.ProductName,
			&item.Description, &item.Quantity, &item.UnitPrice, &item.TotalPrice); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, item)
	}
	if items == nil {
		items = []dto.ServiceJobItemResponse{}
	}
	return items, nil
}

func (s *Service) generateJobNumber(ctx context.Context) (string, error) {
	year := fmt.Sprintf("%d", time.Now().Year())
	var seq int
	err := s.pool.QueryRow(ctx,
		`SELECT COALESCE(MAX(SUBSTRING(job_number FROM 10)::int), 0) + 1
		 FROM service_jobs WHERE job_number LIKE $1`, "JOB-"+year+"-%").Scan(&seq)
	if err != nil {
		return "", fmt.Errorf("generate job number: %w", err)
	}
	return fmt.Sprintf("JOB-%s-%04d", year, seq), nil
}

// reserveStock holds units against a job's future completion so POS sales and
// other jobs can't sell them out from under it. Fails if not enough is
// actually available (on hand minus whatever's already reserved elsewhere).
func reserveStock(ctx context.Context, tx pgx.Tx, branchID, productID int64, qty float64) error {
	if qty <= 0 {
		return nil
	}
	var stockQty, reserved float64
	err := tx.QueryRow(ctx,
		`SELECT stock_quantity, reserved_quantity FROM products WHERE id = $1 AND branch_id = $2 AND is_active = true FOR UPDATE`,
		productID, branchID).Scan(&stockQty, &reserved)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.ErrNotFound
		}
		return fmt.Errorf("check stock for reservation: %w", err)
	}
	available := stockQty - reserved
	if available < qty {
		return &domain.AppError{
			Code:    "INSUFFICIENT_STOCK",
			Message: fmt.Sprintf("Not enough available stock: %g available (%g on hand, %g already reserved for other jobs)", available, stockQty, reserved),
			Status:  400,
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE products SET reserved_quantity = reserved_quantity + $1 WHERE id = $2`, qty, productID); err != nil {
		return fmt.Errorf("reserve stock: %w", err)
	}
	return nil
}

// releaseStock gives back units a job is no longer holding (item removed, job
// cancelled/deleted). Clamped at 0 so it's safe even if called more than once.
func releaseStock(ctx context.Context, tx pgx.Tx, productID int64, qty float64) error {
	if qty <= 0 {
		return nil
	}
	if _, err := tx.Exec(ctx, `UPDATE products SET reserved_quantity = GREATEST(reserved_quantity - $1, 0) WHERE id = $2`, qty, productID); err != nil {
		return fmt.Errorf("release stock: %w", err)
	}
	return nil
}

// releaseJobReservations releases every product line a job is currently
// holding stock for — used when a job is cancelled or deleted.
func releaseJobReservations(ctx context.Context, tx pgx.Tx, jobID int64) error {
	rows, err := tx.Query(ctx,
		`SELECT product_id, quantity FROM service_job_items WHERE service_job_id = $1 AND product_id IS NOT NULL`,
		jobID)
	if err != nil {
		return fmt.Errorf("list job reservations: %w", err)
	}
	type line struct {
		productID int64
		qty       float64
	}
	var lines []line
	for rows.Next() {
		var l line
		if err := rows.Scan(&l.productID, &l.qty); err != nil {
			rows.Close()
			return fmt.Errorf("scan job reservation: %w", err)
		}
		lines = append(lines, l)
	}
	rows.Close()

	for _, l := range lines {
		if err := releaseStock(ctx, tx, l.productID, l.qty); err != nil {
			return err
		}
	}
	return nil
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
