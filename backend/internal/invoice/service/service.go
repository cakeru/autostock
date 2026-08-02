package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/batch"
	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/invoice/dto"
	telegrammodels "github.com/cakeru/autostock/internal/telegram/models"
	telegram "github.com/cakeru/autostock/internal/telegram/service"
	"github.com/cakeru/autostock/internal/vehicleevent"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) List(ctx context.Context, branchID int64, filter dto.InvoiceFilter) ([]dto.InvoiceListResponse, int, error) {
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

	rows, err := s.pool.Query(ctx, `
		SELECT i.id, i.invoice_number, i.status, i.payment_status,
		       i.customer_id, COALESCE(c.name, ''),
		       i.vehicle_id,
		       CASE WHEN v.id IS NOT NULL THEN COALESCE(v.plate_number, '') ELSE '' END,
		       i.service_job_id,
		       CASE WHEN sj.id IS NOT NULL THEN sj.job_number ELSE '' END,
		       i.subtotal, i.total_usd, i.exchange_rate, i.total_khr,
		       i.paid_amount, i.issued_at, i.created_at,
		       i.created_by, COALESCE(cred.full_name, ''),
		       COUNT(*) OVER() as total_count
		FROM invoices i
		LEFT JOIN customers c ON c.id = i.customer_id
		LEFT JOIN vehicles v ON v.id = i.vehicle_id
		LEFT JOIN service_jobs sj ON sj.id = i.service_job_id
		LEFT JOIN users cred ON cred.id = i.created_by
		WHERE i.branch_id = $1
		  AND ($2::text IS NULL OR i.status = $2)
		  AND ($3::text IS NULL OR i.payment_status = $3)
		  AND ($4::bigint IS NULL OR i.customer_id = $4)
		  AND ($5::text IS NULL OR i.invoice_number ILIKE '%' || $5 || '%')
		  AND ($6::text IS NULL OR i.created_at >= $6::timestamptz)
		  AND ($7::text IS NULL OR i.created_at <= $7::timestamptz)
		ORDER BY i.created_at DESC
		LIMIT $8 OFFSET $9`,
		branchID, nullIfEmpty(filter.Status), nullIfEmpty(filter.PaymentStatus),
		customerID, nullIfEmpty(filter.InvoiceNumber),
		nullIfEmpty(filter.CreatedAtGTE), nullIfEmpty(filter.CreatedAtLTE),
		filter.PerPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query invoices: %w", err)
	}
	defer rows.Close()

	var invoices []dto.InvoiceListResponse
	var total int
	for rows.Next() {
		var inv dto.InvoiceListResponse
		if err := rows.Scan(&inv.ID, &inv.InvoiceNumber, &inv.Status, &inv.PaymentStatus,
			&inv.CustomerID, &inv.CustomerName, &inv.VehicleID, &inv.PlateNumber,
			&inv.ServiceJobID, &inv.JobNumber, &inv.Subtotal, &inv.TotalUSD,
			&inv.ExchangeRate, &inv.TotalKHR, &inv.PaidAmount, &inv.IssuedAt, &inv.CreatedAt,
			&inv.CreatedByID, &inv.CreatedByName,
			&total); err != nil {
			return nil, 0, fmt.Errorf("scan invoice: %w", err)
		}
		invoices = append(invoices, inv)
	}

	if invoices == nil {
		invoices = []dto.InvoiceListResponse{}
	}
	return invoices, total, nil
}

func (s *Service) Get(ctx context.Context, branchID int64, id int64) (*dto.InvoiceDetailResponse, error) {
	var inv dto.InvoiceDetailResponse
	err := s.pool.QueryRow(ctx, `
		SELECT i.id, i.invoice_number, i.status, i.payment_status,
		       i.customer_id, COALESCE(c.name, ''), COALESCE(c.phone, ''), COALESCE(c.address, ''),
		       i.vehicle_id,
		       CASE WHEN v.id IS NOT NULL THEN COALESCE(v.plate_number, '') ELSE '' END,
		       CASE WHEN v.id IS NOT NULL THEN CONCAT_WS(' ', COALESCE(v.make,''), COALESCE(v.model,''), COALESCE(v.year::text,'')) ELSE '' END,
		       i.mileage,
		       i.service_job_id,
		       CASE WHEN sj.id IS NOT NULL THEN sj.job_number ELSE '' END,
		       i.subtotal, i.tax_rate, i.tax_amount, i.discount,
		       i.total_usd, i.exchange_rate, i.total_khr,
		       i.paid_amount, COALESCE(i.payment_method, ''), COALESCE(i.payment_notes, ''),
		       COALESCE(i.notes, ''), i.issued_at,
		       i.voided_at, COALESCE(i.void_reason, ''),
		       i.created_by, COALESCE(cred.full_name, ''),
		       i.created_at, i.updated_at
		FROM invoices i
		LEFT JOIN customers c ON c.id = i.customer_id
		LEFT JOIN vehicles v ON v.id = i.vehicle_id
		LEFT JOIN service_jobs sj ON sj.id = i.service_job_id
		LEFT JOIN users cred ON cred.id = i.created_by
		WHERE i.id = $1 AND i.branch_id = $2`, id, branchID).
		Scan(&inv.ID, &inv.InvoiceNumber, &inv.Status, &inv.PaymentStatus,
			&inv.CustomerID, &inv.CustomerName, &inv.CustomerPhone, &inv.CustomerAddr,
			&inv.VehicleID, &inv.PlateNumber, &inv.VehicleInfo, &inv.Mileage,
			&inv.ServiceJobID, &inv.JobNumber,
			&inv.Subtotal, &inv.TaxRate, &inv.TaxAmount, &inv.Discount,
			&inv.TotalUSD, &inv.ExchangeRate, &inv.TotalKHR,
			&inv.PaidAmount, &inv.PaymentMethod, &inv.PaymentNotes,
			&inv.Notes, &inv.IssuedAt,
			&inv.VoidedAt, &inv.VoidReason,
			&inv.CreatedByID, &inv.CreatedByName,
			&inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get invoice: %w", err)
	}

	items, err := s.ListItems(ctx, id)
	if err != nil {
		return nil, err
	}
	inv.Items = items
	if inv.Items == nil {
		inv.Items = []dto.InvoiceItemResp{}
	}

	payments, err := s.GetPayments(ctx, id)
	if err != nil {
		return nil, err
	}
	inv.Payments = payments

	return &inv, nil
}

func (s *Service) Create(ctx context.Context, branchID int64, userID int64, req *dto.CreateInvoiceRequest) (*dto.InvoiceListResponse, error) {
	exchangeRate := req.ExchangeRate
	if exchangeRate == 0 {
		err := s.pool.QueryRow(ctx,
			`SELECT value::numeric FROM settings WHERE branch_id = $1 AND key = 'exchange_rate_usd_khr'`,
			branchID).Scan(&exchangeRate)
		if err != nil {
			exchangeRate = 4050
		}
	}

	var subtotal float64
	for _, item := range req.Items {
		subtotal += math.Round(item.Quantity*item.UnitPriceUSD*100) / 100
	}

	// Tax is applied from branch settings so every path (POS sale or a job
	// converted to an invoice) is taxed identically — the settings are the single
	// source of truth rather than whatever the caller happened to send.
	taxRate := s.taxRateFor(ctx, branchID)
	taxAmount := math.Round(subtotal*taxRate/100*100) / 100
	discount := req.Discount
	totalUSD := math.Round((subtotal+taxAmount-discount)*100) / 100
	totalKHR := math.Round(totalUSD*exchangeRate*100) / 100
	paymentStatus := "unpaid"

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext('invoice_number'))`)
	if err != nil {
		return nil, fmt.Errorf("advisory lock: %w", err)
	}

	year := fmt.Sprintf("%d", time.Now().Year())
	var seq int
	err = tx.QueryRow(ctx,
		`SELECT COALESCE(MAX(SUBSTRING(invoice_number FROM 10)::int), 0) + 1
		 FROM invoices WHERE invoice_number LIKE $1`, "INV-"+year+"-%").Scan(&seq)
	if err != nil {
		return nil, fmt.Errorf("generate invoice number: %w", err)
	}
	invoiceNumber := fmt.Sprintf("INV-%s-%04d", year, seq)

	// Oversell check: verify *available* stock (on hand minus whatever's
	// reserved for other scheduled jobs) for each product item. When invoicing
	// a job's own items, its reservation is released just before this call, so
	// its own held units count as available here. Also captures each
	// product's low-stock threshold so the deduction loop below can flag a
	// crossing without a second round trip.
	type stockInfo struct {
		sku, name, productType  string
		isOil                   bool
		ratedLifeKm             *int
		oilIntervalKm           *int
		oilIntervalMonths       *int
		oldQty, minAlert        float64
	}
	stockByProduct := map[int64]stockInfo{}
	for _, item := range req.Items {
		if item.ProductID != nil {
			var stockQty, reservedQty, minAlert float64
			var sku, name, productType string
			var isOil bool
			var ratedLifeKm *int
			var oilIntervalKm, oilIntervalMonths *int
			err := tx.QueryRow(ctx,
				`SELECT stock_quantity, reserved_quantity, min_stock_alert, sku, name, type, is_oil_product, rated_life_km, oil_interval_km, oil_interval_months FROM products WHERE id = $1 AND branch_id = $2 AND is_active = true FOR UPDATE`,
				*item.ProductID, branchID).Scan(&stockQty, &reservedQty, &minAlert, &sku, &name, &productType, &isOil, &ratedLifeKm, &oilIntervalKm, &oilIntervalMonths)
			if err != nil {
				return nil, fmt.Errorf("check stock for product %d: %w", *item.ProductID, err)
			}
			available := stockQty - reservedQty
			if available < item.Quantity {
				return nil, &domain.AppError{
					Code:    "INSUFFICIENT_STOCK",
					Message: fmt.Sprintf("insufficient available stock for product %d: %g available (%g on hand, %g reserved for other jobs), need %g", *item.ProductID, available, stockQty, reservedQty, item.Quantity),
					Status:  400,
				}
			}
			stockByProduct[*item.ProductID] = stockInfo{sku: sku, name: name, productType: productType, isOil: isOil, ratedLifeKm: ratedLifeKm, oilIntervalKm: oilIntervalKm, oilIntervalMonths: oilIntervalMonths, oldQty: stockQty, minAlert: minAlert}
		}
	}

	var inv dto.InvoiceListResponse
	err = tx.QueryRow(ctx, `
		INSERT INTO invoices (branch_id, invoice_number, customer_id, vehicle_id, service_job_id, mileage,
		                      subtotal, tax_rate, tax_amount, discount,
		                      total_usd, exchange_rate, total_khr,
		                      payment_status, payment_method, notes, status, issued_at, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,'issued',NOW(),$17)
		RETURNING id, invoice_number, status, payment_status, customer_id, vehicle_id, service_job_id,
		          subtotal, total_usd, exchange_rate, total_khr, paid_amount, issued_at, created_at, created_by`,
		branchID, invoiceNumber, req.CustomerID, req.VehicleID, req.ServiceJobID, req.Mileage,
		subtotal, taxRate, taxAmount, discount,
		totalUSD, exchangeRate, totalKHR,
		paymentStatus, req.PaymentMethod, req.Notes, userID).
		Scan(&inv.ID, &inv.InvoiceNumber, &inv.Status, &inv.PaymentStatus,
			&inv.CustomerID, &inv.VehicleID, &inv.ServiceJobID,
			&inv.Subtotal, &inv.TotalUSD, &inv.ExchangeRate,
			&inv.TotalKHR, &inv.PaidAmount, &inv.IssuedAt, &inv.CreatedAt,
			&inv.CreatedByID)
	if err != nil {
		return nil, fmt.Errorf("create invoice: %w", err)
	}

	if inv.CreatedByID != nil {
		_ = s.pool.QueryRow(ctx, `SELECT full_name FROM users WHERE id = $1`, *inv.CreatedByID).Scan(&inv.CreatedByName)
	}

	for _, item := range req.Items {
		desc := item.Description
		if desc == "" {
			desc = item.ItemType
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO invoice_items (invoice_id, product_id, item_type, description, quantity, unit_price_usd, total_usd)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			inv.ID, item.ProductID, item.ItemType, desc, item.Quantity, item.UnitPriceUSD, item.Quantity*item.UnitPriceUSD)
		if err != nil {
			return nil, fmt.Errorf("add invoice item: %w", err)
		}
	}

	// Deduct stock for each product item within the transaction. The product
	// counter stays authoritative for oversell; FIFO consumption links the sale
	// to the specific intake batches it drew from (traceability + recall).
	for _, item := range req.Items {
		if item.ProductID != nil {
			_, err := tx.Exec(ctx,
				`UPDATE products SET stock_quantity = stock_quantity - $1 WHERE id = $2`,
				item.Quantity, *item.ProductID)
			if err != nil {
				return nil, fmt.Errorf("deduct stock for product %d: %w", *item.ProductID, err)
			}
			if err := batch.ConsumeFIFO(ctx, tx, branchID, *item.ProductID, item.Quantity, "invoice_issued", "invoice", &inv.ID, &userID); err != nil {
				return nil, fmt.Errorf("consume batches for product %d: %w", *item.ProductID, err)
			}

			if info, ok := stockByProduct[*item.ProductID]; ok {
				newQty := info.oldQty - item.Quantity
				if info.oldQty >= info.minAlert && newQty < info.minAlert {
					if err := telegram.LogEvent(ctx, tx, branchID, telegrammodels.TopicAlerts, "low_stock", "product", *item.ProductID, map[string]any{
						"sku": info.sku, "name": info.name, "stock_qty": newQty, "min_alert": info.minAlert,
					}); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	if req.ServiceJobID != nil {
		result, err := tx.Exec(ctx, `UPDATE service_jobs SET invoice_id = $1 WHERE id = $2`, inv.ID, *req.ServiceJobID)
		if err != nil {
			return nil, fmt.Errorf("link service job: %w", err)
		}
		if result.RowsAffected() == 0 {
			return nil, domain.ErrNotFound
		}
	}

	// Auto-log a tire/oil service event when this sale includes one, so the
	// vehicle's "due for service" estimate has real history to project from.
	// At most one event per type per invoice — buying 4 tires is one install
	// event, not four.
	if req.VehicleID != nil {
		var tireName, oilName string
		var tireLifeKm *int
		var oilIntervalKm, oilIntervalMonths *int
		for _, item := range req.Items {
			if item.ProductID == nil {
				continue
			}
			info, ok := stockByProduct[*item.ProductID]
			if !ok {
				continue
			}
			if info.productType == "tire" && tireName == "" {
				tireName = info.name
				tireLifeKm = info.ratedLifeKm // the installed tire's rated life anchors its reminder
			}
			if info.isOil && oilName == "" {
				oilName = info.name
				oilIntervalKm = info.oilIntervalKm
				oilIntervalMonths = info.oilIntervalMonths
			}
		}
		occurredAt := time.Now()
		if inv.IssuedAt != nil {
			occurredAt = *inv.IssuedAt
		}
		if tireName != "" {
			if err := vehicleevent.LogEvent(ctx, tx, branchID, *req.VehicleID, "tire", req.Mileage, occurredAt, &inv.ID, req.ServiceJobID, tireName, tireLifeKm, nil, &userID); err != nil {
				return nil, err
			}
		}
		if oilName != "" {
			if err := vehicleevent.LogEvent(ctx, tx, branchID, *req.VehicleID, "oil", req.Mileage, occurredAt, &inv.ID, req.ServiceJobID, oilName, oilIntervalKm, oilIntervalMonths, &userID); err != nil {
				return nil, err
			}
		}
	}

	var customerName string
	_ = tx.QueryRow(ctx, `SELECT COALESCE(name, '') FROM customers WHERE id = $1`, req.CustomerID).Scan(&customerName)
	if err := telegram.LogEvent(ctx, tx, branchID, telegrammodels.TopicSales, "sale_issued", "invoice", inv.ID, map[string]any{
		"invoice_number": inv.InvoiceNumber, "customer_name": customerName,
		"total_usd": inv.TotalUSD, "item_count": len(req.Items),
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	s.logAudit(ctx, branchID, userID, "invoice_created", "invoice", inv.ID)

	inv.Subtotal = subtotal
	return &inv, nil
}

// taxRateFor returns the branch's active tax rate (0 when tax is disabled),
// making settings the single source of truth for tax across all invoice paths.
func (s *Service) taxRateFor(ctx context.Context, branchID int64) float64 {
	var enabled bool
	if err := s.pool.QueryRow(ctx,
		`SELECT value = 'true' FROM settings WHERE branch_id = $1 AND key = 'tax_enabled'`,
		branchID).Scan(&enabled); err != nil || !enabled {
		return 0
	}
	var rate float64
	_ = s.pool.QueryRow(ctx,
		`SELECT value::numeric FROM settings WHERE branch_id = $1 AND key = 'tax_rate_percent'`,
		branchID).Scan(&rate)
	return rate
}

func (s *Service) CreateFromServiceJob(ctx context.Context, branchID int64, userID int64, jobID int64, req *dto.CreateInvoiceFromJobRequest) (*dto.InvoiceListResponse, error) {
	var customerID, vehicleID, existingInvoiceID *int64
	var mileage *int
	var status string
	var jobDiscount float64
	err := s.pool.QueryRow(ctx, `SELECT customer_id, vehicle_id, mileage, status, invoice_id, COALESCE(discount, 0) FROM service_jobs WHERE id = $1`, jobID).
		Scan(&customerID, &vehicleID, &mileage, &status, &existingInvoiceID, &jobDiscount)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	if existingInvoiceID != nil {
		return nil, &domain.AppError{Code: "ALREADY_INVOICED", Message: "Service job already has an invoice", Status: 400}
	}

	if status != "completed" {
		return nil, &domain.AppError{Code: "INVALID_STATUS", Message: "Service job must be completed before invoicing", Status: 400}
	}

	rows, err := s.pool.Query(ctx, `
		SELECT sji.product_id, sji.item_type, COALESCE(sji.description, ''), sji.quantity, sji.unit_price
		FROM service_job_items sji
		WHERE sji.service_job_id = $1`, jobID)
	if err != nil {
		return nil, fmt.Errorf("get job items: %w", err)
	}
	defer rows.Close()

	var items []dto.InvoiceItemReq
	for rows.Next() {
		var item dto.InvoiceItemReq
		if err := rows.Scan(&item.ProductID, &item.ItemType, &item.Description, &item.Quantity, &item.UnitPriceUSD); err != nil {
			return nil, fmt.Errorf("scan job item: %w", err)
		}
		items = append(items, item)
	}

	if len(items) == 0 {
		return nil, &domain.AppError{Code: "NO_ITEMS", Message: "Service job has no items to invoice", Status: 400}
	}

	// Release this job's own reservations before the real deduction below —
	// otherwise its own held units would count against themselves in the
	// oversell check and invoicing would fail on the job's own stock.
	if err := s.releaseJobReservations(ctx, jobID); err != nil {
		return nil, err
	}

	// The discount agreed when the cart was saved as a job survives the
	// conversion; an explicit one on the request wins.
	discount := jobDiscount
	if req.Discount > 0 {
		discount = req.Discount
	}

	return s.Create(ctx, branchID, userID, &dto.CreateInvoiceRequest{
		CustomerID:    customerID,
		VehicleID:     vehicleID,
		ServiceJobID:  &jobID,
		Mileage:       mileage,
		Items:         items,
		Discount:      discount,
		ExchangeRate:  req.ExchangeRate,
		PaymentMethod: req.PaymentMethod,
		Notes:         req.Notes,
	})
}

func (s *Service) Update(ctx context.Context, branchID int64, id int64, req *dto.UpdateInvoiceRequest) (*dto.InvoiceDetailResponse, error) {
	inv, err := s.Get(ctx, branchID, id)
	if err != nil {
		return nil, err
	}

	if inv.Status == "voided" {
		return nil, &domain.AppError{Code: "VOIDED", Message: "Cannot update a voided invoice", Status: 400}
	}

	oldVehicleID := inv.VehicleID
	newVehicleID := req.VehicleID
	if req.ClearVehicle {
		newVehicleID = nil
	}
	vehicleChanged := oldVehicleID == nil || newVehicleID == nil || *oldVehicleID != *newVehicleID

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE invoices
		SET payment_method = COALESCE($1, payment_method),
		    payment_notes = COALESCE($2, payment_notes),
		    notes = COALESCE($3, notes),
		    vehicle_id = CASE WHEN $6 THEN NULL ELSE COALESCE($4, vehicle_id) END,
		    mileage = COALESCE($5, mileage),
		    updated_at = NOW()
		WHERE id = $7 AND branch_id = $8`,
		req.PaymentMethod, req.PaymentNotes, req.Notes, req.VehicleID, req.Mileage,
		req.ClearVehicle, id, branchID)
	if err != nil {
		return nil, fmt.Errorf("update invoice: %w", err)
	}

	// The vehicle's tire/oil service events are logged when an invoice is
	// issued, so attaching a vehicle after the fact (or switching to another)
	// must re-sync them: drop this invoice's events and re-log for the vehicle
	// it now belongs to, so the due-for-service estimate stays accurate.
	if vehicleChanged {
		if _, err := tx.Exec(ctx, `DELETE FROM vehicle_service_events WHERE invoice_id = $1`, id); err != nil {
			return nil, fmt.Errorf("clear vehicle service events: %w", err)
		}
		if newVehicleID != nil {
			occurredAt := time.Now()
			if inv.IssuedAt != nil {
				occurredAt = *inv.IssuedAt
			}
			if err := s.logVehicleEvents(ctx, tx, branchID, *newVehicleID, id, req.Mileage, occurredAt); err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return s.Get(ctx, branchID, id)
}

// logVehicleEvents mirrors the tire/oil service-event logging that Create
// performs at issue time, so a vehicle attached to an existing invoice still
// gets its due-for-service history. Idempotent per (vehicle, type, invoice).
func (s *Service) logVehicleEvents(ctx context.Context, tx pgx.Tx, branchID, vehicleID, invoiceID int64, mileage *int, occurredAt time.Time) error {
	rows, err := tx.Query(ctx, `
		SELECT p.type, p.is_oil_product, p.name, p.rated_life_km, p.oil_interval_km, p.oil_interval_months
		FROM invoice_items ii
		JOIN products p ON p.id = ii.product_id
		WHERE ii.invoice_id = $1 AND ii.product_id IS NOT NULL`, invoiceID)
	if err != nil {
		return fmt.Errorf("query invoice products: %w", err)
	}
	defer rows.Close()

	var tireName, oilName string
	var tireLifeKm, oilIntervalKm, oilIntervalMonths *int
	for rows.Next() {
		var ptype string
		var isOil bool
		var name string
		var lifeKm, oilKm, oilMonths *int
		if err := rows.Scan(&ptype, &isOil, &name, &lifeKm, &oilKm, &oilMonths); err != nil {
			return fmt.Errorf("scan invoice product: %w", err)
		}
		if ptype == "tire" && tireName == "" {
			tireName = name
			tireLifeKm = lifeKm
		}
		if isOil && oilName == "" {
			oilName = name
			oilIntervalKm = oilKm
			oilIntervalMonths = oilMonths
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate invoice products: %w", err)
	}

	var invID = invoiceID
	if tireName != "" {
		if err := vehicleevent.LogEvent(ctx, tx, branchID, vehicleID, "tire", mileage, occurredAt, &invID, nil, tireName, tireLifeKm, nil, nil); err != nil {
			return err
		}
	}
	if oilName != "" {
		if err := vehicleevent.LogEvent(ctx, tx, branchID, vehicleID, "oil", mileage, occurredAt, &invID, nil, oilName, oilIntervalKm, oilIntervalMonths, nil); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) RecordPayment(ctx context.Context, branchID int64, id int64, userID int64, req *dto.RecordPaymentRequest) (*dto.PaymentResponse, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var status string
	var totalUSD float64
	err = tx.QueryRow(ctx, `SELECT status, total_usd FROM invoices WHERE id = $1 AND branch_id = $2 FOR UPDATE`, id, branchID).
		Scan(&status, &totalUSD)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	if status == "voided" {
		return nil, &domain.AppError{Code: "VOIDED", Message: "Cannot record payment on voided invoice", Status: 400}
	}

	var sumPaid float64
	err = tx.QueryRow(ctx, `SELECT COALESCE(SUM(amount), 0) FROM payments WHERE invoice_id = $1`, id).Scan(&sumPaid)
	if err != nil {
		return nil, fmt.Errorf("sum payments: %w", err)
	}

	totalPaid := sumPaid + req.Amount
	if totalPaid > totalUSD {
		return nil, &domain.AppError{Code: "PAYMENT_EXCEEDS_TOTAL", Message: "Payment would exceed invoice total", Status: 400}
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}
	var p dto.PaymentResponse
	err = tx.QueryRow(ctx, `
		INSERT INTO payments (invoice_id, amount, method, received_by, notes, currency, tendered_amount, exchange_rate, reference)
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, 0), NULLIF($8, 0), NULLIF($9, ''))
		RETURNING id, invoice_id, amount, method, COALESCE(notes, ''), currency, tendered_amount, COALESCE(reference, ''), created_at`,
		id, req.Amount, req.Method, userID, req.Notes, currency, req.TenderedAmount, req.ExchangeRate, req.Reference).
		Scan(&p.ID, &p.InvoiceID, &p.Amount, &p.Method, &p.Notes, &p.Currency, &p.TenderedAmount, &p.Reference, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("record payment: %w", err)
	}

	if userID > 0 {
		_ = s.pool.QueryRow(ctx, `SELECT full_name FROM users WHERE id = $1`, userID).Scan(&p.ReceivedByName)
	}

	paymentStatus := "unpaid"
	if totalPaid >= totalUSD {
		paymentStatus = "paid"
	} else if totalPaid > 0 {
		paymentStatus = "partial"
	}

	invoiceStatus := status
	if paymentStatus == "paid" {
		invoiceStatus = "paid"
	}

	_, err = tx.Exec(ctx,
		`UPDATE invoices SET paid_amount = $1, payment_status = $2, status = $3, updated_at = NOW() WHERE id = $4`,
		totalPaid, paymentStatus, invoiceStatus, id)
	if err != nil {
		return nil, fmt.Errorf("update invoice payment: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	s.logAudit(ctx, branchID, userID, "payment_recorded", "invoice", id)
	return &p, nil
}

func (s *Service) GetPayments(ctx context.Context, invoiceID int64) ([]dto.PaymentResponse, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.id, p.invoice_id, p.amount, p.method, COALESCE(p.currency, 'USD'), p.tendered_amount,
		       COALESCE(u.full_name, ''), COALESCE(p.notes, ''), COALESCE(p.reference, ''), COALESCE(p.proof_url, ''), p.created_at
		FROM payments p
		LEFT JOIN users u ON u.id = p.received_by
		WHERE p.invoice_id = $1
		ORDER BY p.created_at`, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("query payments: %w", err)
	}
	defer rows.Close()

	var payments []dto.PaymentResponse
	for rows.Next() {
		var pay dto.PaymentResponse
		if err := rows.Scan(&pay.ID, &pay.InvoiceID, &pay.Amount, &pay.Method, &pay.Currency, &pay.TenderedAmount, &pay.ReceivedByName, &pay.Notes, &pay.Reference, &pay.ProofURL, &pay.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan payment: %w", err)
		}
		payments = append(payments, pay)
	}
	if payments == nil {
		payments = []dto.PaymentResponse{}
	}
	return payments, nil
}

// SetPaymentProof attaches a stored proof photo to a payment.
func (s *Service) SetPaymentProof(ctx context.Context, branchID, paymentID int64, proofURL string) error {
	result, err := s.pool.Exec(ctx, `
		UPDATE payments p
		SET proof_url = $1
		FROM invoices i
		WHERE p.id = $2 AND i.id = p.invoice_id AND i.branch_id = $3 AND i.status <> 'voided'`,
		proofURL, paymentID, branchID)
	if err != nil {
		return fmt.Errorf("set payment proof: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Service) Void(ctx context.Context, branchID int64, id int64, userID int64, reason string) (*dto.InvoiceDetailResponse, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var status, invoiceNumber string
	err = tx.QueryRow(ctx, `SELECT status, invoice_number FROM invoices WHERE id = $1 AND branch_id = $2`, id, branchID).Scan(&status, &invoiceNumber)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	if status == "voided" {
		return nil, &domain.AppError{Code: "ALREADY_VOIDED", Message: "Invoice is already voided", Status: 400}
	}

	// Restore stock to the exact batches this invoice consumed, using the
	// issue movements as the record of what came from where.
	type consumed struct {
		productID int64
		batchID   *int64
		qty       float64
	}
	var restores []consumed
	rows, err := tx.Query(ctx, `
		SELECT product_id, batch_id, -SUM(quantity_change)
		FROM stock_movements
		WHERE reference_type = 'invoice' AND reference_id = $1 AND reason = 'invoice_issued'
		GROUP BY product_id, batch_id`, id)
	if err != nil {
		return nil, fmt.Errorf("get consumed batches: %w", err)
	}
	for rows.Next() {
		var c consumed
		if err := rows.Scan(&c.productID, &c.batchID, &c.qty); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan consumed: %w", err)
		}
		restores = append(restores, c)
	}
	rows.Close()

	for _, c := range restores {
		if _, err := tx.Exec(ctx,
			`UPDATE products SET stock_quantity = stock_quantity + $1 WHERE id = $2`,
			c.qty, c.productID); err != nil {
			return nil, fmt.Errorf("restore stock for product %d: %w", c.productID, err)
		}
		if c.batchID != nil {
			if _, err := tx.Exec(ctx,
				`UPDATE batches SET quantity_remaining = quantity_remaining + $1 WHERE id = $2`,
				c.qty, *c.batchID); err != nil {
				return nil, fmt.Errorf("restore batch %d: %w", *c.batchID, err)
			}
		}
		if err := batch.RecordMovement(ctx, tx, branchID, c.productID, c.qty, "invoice_voided", "invoice", &id, c.batchID, &userID); err != nil {
			return nil, err
		}
	}

	_, err = tx.Exec(ctx,
		`UPDATE invoices SET status = 'voided', payment_status = 'voided', voided_at = NOW(), void_reason = $1, voided_by = $2, updated_at = NOW() WHERE id = $3`,
		reason, userID, id)
	if err != nil {
		return nil, fmt.Errorf("void invoice: %w", err)
	}

	if err := telegram.LogEvent(ctx, tx, branchID, telegrammodels.TopicSales, "sale_voided", "invoice", id, map[string]any{
		"invoice_number": invoiceNumber, "reason": reason,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	s.logAudit(ctx, branchID, userID, "invoice_voided", "invoice", id)

	return s.Get(ctx, branchID, id)
}

func (s *Service) ListItems(ctx context.Context, invoiceID int64) ([]dto.InvoiceItemResp, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, product_id, item_type, description, quantity, unit_price_usd, total_usd
		FROM invoice_items WHERE invoice_id = $1 ORDER BY created_at`, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	defer rows.Close()

	var items []dto.InvoiceItemResp
	for rows.Next() {
		var item dto.InvoiceItemResp
		if err := rows.Scan(&item.ID, &item.ProductID, &item.ItemType, &item.Description,
			&item.Quantity, &item.UnitPriceUSD, &item.TotalUSD); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, item)
	}
	if items == nil {
		items = []dto.InvoiceItemResp{}
	}
	return items, nil
}

func (s *Service) AddItem(ctx context.Context, branchID int64, userID int64, invoiceID int64, req *dto.InvoiceItemReq) (*dto.InvoiceItemResp, error) {
	desc := req.Description
	if desc == "" {
		desc = req.ItemType
	}

	var status string
	err := s.pool.QueryRow(ctx, `SELECT status FROM invoices WHERE id = $1 AND branch_id = $2`, invoiceID, branchID).Scan(&status)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	if status == "voided" {
		return nil, &domain.AppError{Code: "VOIDED", Message: "Cannot modify items on a voided invoice", Status: 400}
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var item dto.InvoiceItemResp
	err = tx.QueryRow(ctx, `
		INSERT INTO invoice_items (invoice_id, product_id, item_type, description, quantity, unit_price_usd, total_usd)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, product_id, item_type, description, quantity, unit_price_usd, total_usd`,
		invoiceID, req.ProductID, req.ItemType, desc, req.Quantity, req.UnitPriceUSD, req.Quantity*req.UnitPriceUSD).
		Scan(&item.ID, &item.ProductID, &item.ItemType, &item.Description,
			&item.Quantity, &item.UnitPriceUSD, &item.TotalUSD)
	if err != nil {
		return nil, fmt.Errorf("add item: %w", err)
	}

	// Draft invoices never touched stock; everything else has already been
	// deducted once, so a line added now must deduct as an issue would.
	if status != "draft" && req.ProductID != nil {
		if err := s.consumeStock(ctx, tx, branchID, invoiceID, *req.ProductID, req.Quantity, &userID); err != nil {
			return nil, err
		}
	}

	if err := s.recalculateInvoice(ctx, tx, invoiceID); err != nil {
		return nil, err
	}

	if err := s.syncVehicleEvents(ctx, tx, branchID, invoiceID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return &item, nil
}

func (s *Service) RemoveItem(ctx context.Context, branchID int64, userID int64, itemID int64) error {
	var invoiceID int64
	var status string
	var productID *int64
	var qty float64
	err := s.pool.QueryRow(ctx, `
		SELECT invoice_id, i.status, ii.product_id, ii.quantity FROM invoice_items ii
		JOIN invoices i ON i.id = ii.invoice_id
		WHERE ii.id = $1 AND i.branch_id = $2`, itemID, branchID).Scan(&invoiceID, &status, &productID, &qty)
	if err != nil {
		return domain.ErrNotFound
	}
	if status == "voided" {
		return &domain.AppError{Code: "VOIDED", Message: "Cannot modify items on a voided invoice", Status: 400}
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM invoice_items WHERE id = $1`, itemID)
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	if status != "draft" && productID != nil {
		if err := s.restoreStock(ctx, tx, branchID, invoiceID, *productID, qty, &userID); err != nil {
			return err
		}
	}

	if err := s.recalculateInvoice(ctx, tx, invoiceID); err != nil {
		return err
	}

	if err := s.syncVehicleEvents(ctx, tx, branchID, invoiceID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Service) UpdateItem(ctx context.Context, branchID, userID, invoiceID, itemID int64, req *dto.UpdateInvoiceItemRequest) (*dto.InvoiceItemResp, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var status string
	var productID *int64
	var oldQty, oldPrice float64
	err = tx.QueryRow(ctx, `
		SELECT i.status, ii.product_id, ii.quantity, ii.unit_price_usd
		FROM invoice_items ii JOIN invoices i ON i.id = ii.invoice_id
		WHERE ii.id = $1 AND i.id = $2 AND i.branch_id = $3`, itemID, invoiceID, branchID).
		Scan(&status, &productID, &oldQty, &oldPrice)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	if status == "voided" {
		return nil, &domain.AppError{Code: "VOIDED", Message: "Cannot modify items on a voided invoice", Status: 400}
	}

	newQty := oldQty
	if req.Quantity != nil {
		newQty = *req.Quantity
	}
	newPrice := oldPrice
	if req.UnitPriceUSD != nil {
		newPrice = *req.UnitPriceUSD
	}
	description := ""
	if req.Description != nil {
		description = *req.Description
	}

	if status != "draft" && productID != nil && newQty != oldQty {
		delta := newQty - oldQty
		if delta > 0 {
			if err := s.consumeStock(ctx, tx, branchID, invoiceID, *productID, delta, &userID); err != nil {
				return nil, err
			}
		} else {
			if err := s.restoreStock(ctx, tx, branchID, invoiceID, *productID, -delta, &userID); err != nil {
				return nil, err
			}
		}
	}

	var item dto.InvoiceItemResp
	err = tx.QueryRow(ctx, `
		UPDATE invoice_items
		SET description = COALESCE(NULLIF($1, ''), description),
		    quantity = COALESCE($2, quantity),
		    unit_price_usd = COALESCE($3, unit_price_usd),
		    total_usd = $4
		WHERE id = $5
		RETURNING id, product_id, item_type, description, quantity, unit_price_usd, total_usd`,
		description, req.Quantity, req.UnitPriceUSD, math.Round(newQty*newPrice*100)/100, itemID).
		Scan(&item.ID, &item.ProductID, &item.ItemType, &item.Description,
			&item.Quantity, &item.UnitPriceUSD, &item.TotalUSD)
	if err != nil {
		return nil, fmt.Errorf("update item: %w", err)
	}

	if err := s.recalculateInvoice(ctx, tx, invoiceID); err != nil {
		return nil, err
	}

	if err := s.syncVehicleEvents(ctx, tx, branchID, invoiceID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return &item, nil
}

// consumeStock deducts a product line sold on an issued invoice: oversell
// check, product counter, FIFO batch draw-down, and low-stock alert.
func (s *Service) consumeStock(ctx context.Context, tx pgx.Tx, branchID, invoiceID, productID int64, qty float64, userID *int64) error {
	var stockQty, reservedQty, minAlert float64
	var sku, name string
	err := tx.QueryRow(ctx,
		`SELECT stock_quantity, reserved_quantity, min_stock_alert, sku, name FROM products WHERE id = $1 AND branch_id = $2 AND is_active = true FOR UPDATE`,
		productID, branchID).Scan(&stockQty, &reservedQty, &minAlert, &sku, &name)
	if err != nil {
		return fmt.Errorf("check stock for product %d: %w", productID, err)
	}
	if available := stockQty - reservedQty; available < qty {
		return &domain.AppError{
			Code:    "INSUFFICIENT_STOCK",
			Message: fmt.Sprintf("insufficient available stock for product %d: %g available (%g on hand, %g reserved), need %g", productID, available, stockQty, reservedQty, qty),
			Status:  400,
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE products SET stock_quantity = stock_quantity - $1 WHERE id = $2`, qty, productID); err != nil {
		return fmt.Errorf("deduct stock for product %d: %w", productID, err)
	}
	if err := batch.ConsumeFIFO(ctx, tx, branchID, productID, qty, "invoice_edited", "invoice", &invoiceID, userID); err != nil {
		return err
	}
	if stockQty >= minAlert && stockQty-qty < minAlert {
		if err := telegram.LogEvent(ctx, tx, branchID, telegrammodels.TopicAlerts, "low_stock", "product", productID, map[string]any{
			"sku": sku, "name": name, "stock_qty": stockQty - qty, "min_alert": minAlert,
		}); err != nil {
			return err
		}
	}
	return nil
}

// restoreStock returns product qty to stock and to the exact intake batches
// this invoice drew from (newest first), so batch FIFO stays truthful.
func (s *Service) restoreStock(ctx context.Context, tx pgx.Tx, branchID, invoiceID, productID int64, qty float64, userID *int64) error {
	if _, err := tx.Exec(ctx, `UPDATE products SET stock_quantity = stock_quantity + $1 WHERE id = $2`, qty, productID); err != nil {
		return fmt.Errorf("restore stock for product %d: %w", productID, err)
	}
	remaining := qty
	rows, err := tx.Query(ctx, `
		SELECT batch_id, -SUM(quantity_change)
		FROM stock_movements
		WHERE reference_type = 'invoice' AND reference_id = $1 AND reason IN ('invoice_issued', 'invoice_edited')
		  AND product_id = $2 AND quantity_change < 0
		GROUP BY batch_id
		ORDER BY MAX(created_at) DESC`, invoiceID, productID)
	if err != nil {
		return fmt.Errorf("get consumed batches: %w", err)
	}
	type restore struct {
		batchID *int64
		qty     float64
	}
	var restores []restore
	for rows.Next() {
		var r restore
		if err := rows.Scan(&r.batchID, &r.qty); err != nil {
			rows.Close()
			return fmt.Errorf("scan consumed batch: %w", err)
		}
		restores = append(restores, r)
	}
	rows.Close()
	for _, r := range restores {
		if remaining <= 0 {
			break
		}
		back := r.qty
		if back > remaining {
			back = remaining
		}
		remaining -= back
		if r.batchID != nil {
			if _, err := tx.Exec(ctx, `UPDATE batches SET quantity_remaining = quantity_remaining + $1 WHERE id = $2`, back, *r.batchID); err != nil {
				return fmt.Errorf("restore batch %d: %w", *r.batchID, err)
			}
		}
		if err := batch.RecordMovement(ctx, tx, branchID, productID, back, "invoice_edited", "invoice", &invoiceID, r.batchID, userID); err != nil {
			return err
		}
	}
	return nil
}

// syncVehicleEvents re-derives the invoice's tire/oil service events from its
// current items. Called after any item mutation; the log is idempotent, so
// delete-then-replay keeps the vehicle's due-for-service history exact.
func (s *Service) syncVehicleEvents(ctx context.Context, tx pgx.Tx, branchID, invoiceID int64) error {
	var vehicleID *int64
	var mileage *int
	var issuedAt *time.Time
	err := tx.QueryRow(ctx, `SELECT vehicle_id, mileage, issued_at FROM invoices WHERE id = $1`, invoiceID).
		Scan(&vehicleID, &mileage, &issuedAt)
	if err != nil {
		return fmt.Errorf("get invoice vehicle: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM vehicle_service_events WHERE invoice_id = $1`, invoiceID); err != nil {
		return fmt.Errorf("clear vehicle service events: %w", err)
	}
	if vehicleID == nil {
		return nil
	}
	occurredAt := time.Now()
	if issuedAt != nil {
		occurredAt = *issuedAt
	}
	return s.logVehicleEvents(ctx, tx, branchID, *vehicleID, invoiceID, mileage, occurredAt)
}

func (s *Service) recalculateInvoice(ctx context.Context, tx pgx.Tx, invoiceID int64) error {
	var subtotal, taxRate, discount, exchangeRate float64
	err := tx.QueryRow(ctx,
		`SELECT COALESCE(SUM(total_usd), 0) FROM invoice_items WHERE invoice_id = $1`, invoiceID).Scan(&subtotal)
	if err != nil {
		return fmt.Errorf("sum items: %w", err)
	}

	err = tx.QueryRow(ctx,
		`SELECT tax_rate, discount, exchange_rate FROM invoices WHERE id = $1`, invoiceID).Scan(&taxRate, &discount, &exchangeRate)
	if err != nil {
		return fmt.Errorf("get invoice: %w", err)
	}

	taxAmount := math.Round(subtotal*taxRate/100*100) / 100
	totalUSD := math.Round((subtotal+taxAmount-discount)*100) / 100
	totalKHR := math.Round(totalUSD*exchangeRate*100) / 100

	// Editing a paid invoice may change the total; the recorded paid amount is
	// what was actually received, capped at the total so the invoice never
	// shows overpaid (staff refund any excess separately). Deriving it from the
	// payment ledger means it recovers correctly when the total rises again.
	var received float64
	err = tx.QueryRow(ctx, `SELECT COALESCE(SUM(amount), 0) FROM payments WHERE invoice_id = $1`, invoiceID).Scan(&received)
	if err != nil {
		return fmt.Errorf("sum payments: %w", err)
	}
	paidAmount := received
	if paidAmount > totalUSD {
		paidAmount = totalUSD
	}
	paymentStatus := "unpaid"
	if paidAmount >= totalUSD {
		paymentStatus = "paid"
	} else if paidAmount > 0 {
		paymentStatus = "partial"
	}

	_, err = tx.Exec(ctx,
		`UPDATE invoices SET subtotal = $1, tax_amount = $2, total_usd = $3, total_khr = $4,
		                      paid_amount = $5, payment_status = $6, updated_at = NOW() WHERE id = $7`,
		subtotal, taxAmount, totalUSD, totalKHR, paidAmount, paymentStatus, invoiceID)
	if err != nil {
		return fmt.Errorf("update invoice: %w", err)
	}
	return nil
}

// releaseJobReservations releases every product line a service job is
// currently holding stock for. Called right before a job's items are
// converted into a real invoice deduction, so the hold doesn't double-count
// against itself.
func (s *Service) releaseJobReservations(ctx context.Context, jobID int64) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

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
		if l.qty <= 0 {
			continue
		}
		if _, err := tx.Exec(ctx,
			`UPDATE products SET reserved_quantity = GREATEST(reserved_quantity - $1, 0) WHERE id = $2`,
			l.qty, l.productID); err != nil {
			return fmt.Errorf("release reservation for product %d: %w", l.productID, err)
		}
	}

	return tx.Commit(ctx)
}

func (s *Service) generateInvoiceNumber(ctx context.Context) (string, error) {
	year := fmt.Sprintf("%d", time.Now().Year())
	var seq int
	err := s.pool.QueryRow(ctx,
		`SELECT COALESCE(MAX(SUBSTRING(invoice_number FROM 10)::int), 0) + 1
		 FROM invoices WHERE invoice_number LIKE $1`, "INV-"+year+"-%").Scan(&seq)
	if err != nil {
		return "", fmt.Errorf("generate invoice number: %w", err)
	}
	return fmt.Sprintf("INV-%s-%04d", year, seq), nil
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (s *Service) logAudit(ctx context.Context, branchID, userID int64, action, entityType string, entityID int64) {
	_, _ = s.pool.Exec(ctx,
		`INSERT INTO audit_logs (branch_id, user_id, action, entity_type, entity_id) VALUES ($1, $2, $3, $4, $5)`,
		branchID, userID, action, entityType, entityID)
}
