package service

import (
	"context"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/batch"
	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/inventory/dto"
	"github.com/cakeru/autostock/internal/inventory/models"
	telegrammodels "github.com/cakeru/autostock/internal/telegram/models"
	telegram "github.com/cakeru/autostock/internal/telegram/service"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) List(ctx context.Context, branchID int64, filter dto.ProductFilter) ([]models.Product, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 20
	}
	offset := (filter.Page - 1) * filter.PerPage

	rows, err := s.pool.Query(ctx, `
		SELECT id, branch_id, type, sku, COALESCE(barcode,''), name, COALESCE(description,''), COALESCE(category,''),
		       buy_price, sell_price, stock_quantity, reserved_quantity, min_stock_alert, unit, is_oil_product, is_bulk, rated_life_km,
		       COALESCE(tire_size,''), COALESCE(tire_brand,''), COALESCE(tire_model,''), COALESCE(tire_pattern,''), COALESCE(dot_code,''),
		       COALESCE(load_index,''), COALESCE(speed_rating,''), COALESCE(tire_type,''), COALESCE(location,''), COALESCE(image_url,''), is_active,
		       created_at, updated_at,
		       COUNT(*) OVER() as total_count
		FROM products
		WHERE branch_id = $1 AND is_active = true
		  AND ($2::text IS NULL OR type = $2)
		  AND ($3::text IS NULL OR name ILIKE '%' || $3 || '%')
		  AND ($4::text IS NULL OR tire_size = $4)
		  AND ($5::text IS NULL OR tire_brand = $5)
		  AND ($6::numeric IS NULL OR stock_quantity < $6)
		  AND ($7::text IS NULL OR LOWER(sku) = LOWER($7) OR barcode = $7)
		ORDER BY created_at DESC
		LIMIT $8 OFFSET $9`,
		branchID,
		nullIfEmpty(filter.Type),
		nullIfEmpty(filter.NameLike),
		nullIfEmpty(filter.TireSize),
		nullIfEmpty(filter.TireBrand),
		nullFloat(filter.StockQuantityLT),
		nullIfEmpty(filter.Code),
		filter.PerPage, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("query products: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	var total int
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(
			&p.ID, &p.BranchID, &p.Type, &p.SKU, &p.Barcode, &p.Name, &p.Description,
			&p.Category, &p.BuyPrice, &p.SellPrice, &p.StockQuantity, &p.ReservedQuantity,
			&p.MinStockAlert, &p.Unit, &p.IsOilProduct, &p.IsBulk, &p.RatedLifeKm, &p.TireSize, &p.TireBrand, &p.TireModel,
			&p.TirePattern, &p.DOTCode, &p.LoadIndex, &p.SpeedRating,
			&p.TireType, &p.Location, &p.ImageURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
			&total,
		); err != nil {
			return nil, 0, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, p)
	}

	if products == nil {
		products = []models.Product{}
	}

	return products, total, nil
}

func (s *Service) Get(ctx context.Context, branchID int64, id int64) (*models.Product, error) {
	var p models.Product
	err := s.pool.QueryRow(ctx, `
		SELECT id, branch_id, type, sku, COALESCE(barcode,''), name, COALESCE(description,''), COALESCE(category,''),
		       buy_price, sell_price, stock_quantity, reserved_quantity, min_stock_alert, unit, is_oil_product, is_bulk, rated_life_km,
		       COALESCE(tire_size,''), COALESCE(tire_brand,''), COALESCE(tire_model,''), COALESCE(tire_pattern,''), COALESCE(dot_code,''),
		       COALESCE(load_index,''), COALESCE(speed_rating,''), COALESCE(tire_type,''), COALESCE(location,''), COALESCE(image_url,''), is_active,
		       created_at, updated_at
		FROM products
		WHERE id = $1 AND branch_id = $2 AND is_active = true`, id, branchID).
		Scan(
			&p.ID, &p.BranchID, &p.Type, &p.SKU, &p.Barcode, &p.Name, &p.Description,
			&p.Category, &p.BuyPrice, &p.SellPrice, &p.StockQuantity, &p.ReservedQuantity,
			&p.MinStockAlert, &p.Unit, &p.IsOilProduct, &p.IsBulk, &p.RatedLifeKm, &p.TireSize, &p.TireBrand, &p.TireModel,
			&p.TirePattern, &p.DOTCode, &p.LoadIndex, &p.SpeedRating,
			&p.TireType, &p.Location, &p.ImageURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get product: %w", err)
	}
	return &p, nil
}

// SetImage updates a product's image_url (pass "" to clear) and returns the
// refreshed product. Scoped by branch.
func (s *Service) SetImage(ctx context.Context, branchID int64, id int64, url string) (*models.Product, error) {
	var p models.Product
	err := s.pool.QueryRow(ctx, `
		UPDATE products SET image_url = NULLIF($1, ''), updated_at = NOW()
		WHERE id = $2 AND branch_id = $3 AND is_active = true
		RETURNING id, branch_id, type, sku, COALESCE(barcode,''), name, COALESCE(description,''), COALESCE(category,''),
		          buy_price, sell_price, stock_quantity, reserved_quantity, min_stock_alert, unit, is_oil_product, is_bulk, rated_life_km,
		          COALESCE(tire_size,''), COALESCE(tire_brand,''), COALESCE(tire_model,''), COALESCE(tire_pattern,''), COALESCE(dot_code,''),
		          COALESCE(load_index,''), COALESCE(speed_rating,''), COALESCE(tire_type,''), COALESCE(location,''), COALESCE(image_url,''), is_active,
		          created_at, updated_at`, url, id, branchID).
		Scan(
			&p.ID, &p.BranchID, &p.Type, &p.SKU, &p.Barcode, &p.Name, &p.Description,
			&p.Category, &p.BuyPrice, &p.SellPrice, &p.StockQuantity, &p.ReservedQuantity,
			&p.MinStockAlert, &p.Unit, &p.IsOilProduct, &p.IsBulk, &p.RatedLifeKm, &p.TireSize, &p.TireBrand, &p.TireModel,
			&p.TirePattern, &p.DOTCode, &p.LoadIndex, &p.SpeedRating,
			&p.TireType, &p.Location, &p.ImageURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("set image: %w", err)
	}
	return &p, nil
}

func (s *Service) Create(ctx context.Context, branchID int64, userID int64, req *dto.CreateProductRequest) (*models.Product, error) {
	var count int
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM products WHERE branch_id = $1 AND sku = $2 AND is_active = true`,
		branchID, req.SKU).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("check sku: %w", err)
	}
	if count > 0 {
		return nil, domain.ErrDuplicateSKU
	}

	unit := req.Unit
	if unit == "" {
		unit = "piece"
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var p models.Product
	err = tx.QueryRow(ctx, `
		INSERT INTO products (
		    branch_id, type, sku, name, description, category,
		    buy_price, sell_price, stock_quantity, min_stock_alert, unit,
		    tire_size, tire_brand, tire_model, tire_pattern, dot_code,
		    load_index, speed_rating, tire_type, location, barcode, is_oil_product, rated_life_km, is_bulk)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,
		        $12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)
		RETURNING id, branch_id, type, sku, COALESCE(barcode,''), name, COALESCE(description,''), COALESCE(category,''),
		          buy_price, sell_price, stock_quantity, reserved_quantity, min_stock_alert, unit, is_oil_product, is_bulk, rated_life_km,
		          COALESCE(tire_size,''), COALESCE(tire_brand,''), COALESCE(tire_model,''), COALESCE(tire_pattern,''), COALESCE(dot_code,''),
		          COALESCE(load_index,''), COALESCE(speed_rating,''), COALESCE(tire_type,''), COALESCE(location,''), is_active,
		          created_at, updated_at`,
		branchID, req.Type, req.SKU, req.Name, req.Description,
		req.Category, req.BuyPrice, req.SellPrice, req.StockQuantity,
		req.MinStockAlert, unit,
		req.TireSize, req.TireBrand, req.TireModel, req.TirePattern,
		req.DOTCode, req.LoadIndex, req.SpeedRating, req.TireType, req.Location, req.Barcode, req.IsOilProduct, req.RatedLifeKm, req.IsBulk,
	).Scan(
		&p.ID, &p.BranchID, &p.Type, &p.SKU, &p.Barcode, &p.Name, &p.Description,
		&p.Category, &p.BuyPrice, &p.SellPrice, &p.StockQuantity, &p.ReservedQuantity,
		&p.MinStockAlert, &p.Unit, &p.IsOilProduct, &p.IsBulk, &p.RatedLifeKm, &p.TireSize, &p.TireBrand, &p.TireModel,
		&p.TirePattern, &p.DOTCode, &p.LoadIndex, &p.SpeedRating,
		&p.TireType, &p.Location, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}

	// Opening stock becomes the product's first batch so the batch ledger
	// starts consistent with stock_quantity.
	if p.StockQuantity > 0 {
		batchID, err := batch.Create(ctx, tx, branchID, p.ID, p.StockQuantity, p.BuyPrice, nil, 0, "Opening stock", p.DOTCode, "", &userID)
		if err != nil {
			return nil, err
		}
		if err := batch.RecordMovement(ctx, tx, branchID, p.ID, p.StockQuantity, "opening", "batch", &batchID, &batchID, &userID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return &p, nil
}

func (s *Service) Update(ctx context.Context, branchID int64, id int64, req *dto.UpdateProductRequest) (*models.Product, error) {
	_, err := s.Get(ctx, branchID, id)
	if err != nil {
		return nil, err
	}

	var p models.Product
	err = s.pool.QueryRow(ctx, `
		UPDATE products
		SET name       = COALESCE(NULLIF($1, ''), name),
		    description  = COALESCE($2, description),
		    category     = COALESCE($3, category),
		    buy_price    = COALESCE($4, buy_price),
		    sell_price   = COALESCE($5, sell_price),
		    min_stock_alert = COALESCE($6, min_stock_alert),
		    unit         = COALESCE($7, unit),
		    tire_size    = COALESCE($8, tire_size),
		    tire_brand   = COALESCE($9, tire_brand),
		    tire_model   = COALESCE($10, tire_model),
		    tire_pattern = COALESCE($11, tire_pattern),
		    dot_code     = COALESCE($12, dot_code),
		    load_index   = COALESCE($13, load_index),
		    speed_rating = COALESCE($14, speed_rating),
		    tire_type    = COALESCE($15, tire_type),
		    location     = COALESCE($16, location),
		    barcode      = COALESCE($19, barcode),
		    is_oil_product = COALESCE($20, is_oil_product),
		    rated_life_km = COALESCE($21, rated_life_km),
		    is_bulk      = COALESCE($22, is_bulk),
		    type         = COALESCE($23, type),
		    updated_at   = NOW()
		WHERE id = $17 AND branch_id = $18 AND is_active = true
		RETURNING id, branch_id, type, sku, COALESCE(barcode,''), name, COALESCE(description,''), COALESCE(category,''),
		          buy_price, sell_price, stock_quantity, reserved_quantity, min_stock_alert, unit, is_oil_product, is_bulk, rated_life_km,
		          COALESCE(tire_size,''), COALESCE(tire_brand,''), COALESCE(tire_model,''), COALESCE(tire_pattern,''), COALESCE(dot_code,''),
		          COALESCE(load_index,''), COALESCE(speed_rating,''), COALESCE(tire_type,''), COALESCE(location,''), is_active,
		          created_at, updated_at`,
		req.Name, req.Description, req.Category, req.BuyPrice, req.SellPrice,
		req.MinStockAlert, req.Unit,
		req.TireSize, req.TireBrand, req.TireModel, req.TirePattern,
		req.DOTCode, req.LoadIndex, req.SpeedRating, req.TireType, req.Location,
		id, branchID, req.Barcode, req.IsOilProduct, req.RatedLifeKm, req.IsBulk, req.Type,
	).Scan(
		&p.ID, &p.BranchID, &p.Type, &p.SKU, &p.Barcode, &p.Name, &p.Description,
		&p.Category, &p.BuyPrice, &p.SellPrice, &p.StockQuantity, &p.ReservedQuantity,
		&p.MinStockAlert, &p.Unit, &p.IsOilProduct, &p.IsBulk, &p.RatedLifeKm, &p.TireSize, &p.TireBrand, &p.TireModel,
		&p.TirePattern, &p.DOTCode, &p.LoadIndex, &p.SpeedRating,
		&p.TireType, &p.Location, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("update product: %w", err)
	}

	return &p, nil
}

func (s *Service) ReceiveStock(ctx context.Context, branchID int64, id int64, userID int64, req *dto.ReceiveStockRequest) (*models.Product, error) {
	_, err := s.Get(ctx, branchID, id)
	if err != nil {
		return nil, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var p models.Product
	err = tx.QueryRow(ctx, `
		UPDATE products
		SET stock_quantity = stock_quantity + $1,
		    buy_price = CASE WHEN $2 > 0 THEN $2 ELSE buy_price END,
		    updated_at = NOW()
		WHERE id = $3 AND branch_id = $4 AND is_active = true
		RETURNING id, branch_id, type, sku, COALESCE(barcode,''), name, COALESCE(description,''), COALESCE(category,''),
		          buy_price, sell_price, stock_quantity, reserved_quantity, min_stock_alert, unit, is_oil_product, is_bulk, rated_life_km,
		          COALESCE(tire_size,''), COALESCE(tire_brand,''), COALESCE(tire_model,''), COALESCE(tire_pattern,''), COALESCE(dot_code,''),
		          COALESCE(load_index,''), COALESCE(speed_rating,''), COALESCE(tire_type,''), COALESCE(location,''), is_active,
		          created_at, updated_at`,
		req.Quantity, req.UnitCost, id, branchID,
	).Scan(
		&p.ID, &p.BranchID, &p.Type, &p.SKU, &p.Barcode, &p.Name, &p.Description,
		&p.Category, &p.BuyPrice, &p.SellPrice, &p.StockQuantity, &p.ReservedQuantity,
		&p.MinStockAlert, &p.Unit, &p.IsOilProduct, &p.IsBulk, &p.RatedLifeKm, &p.TireSize, &p.TireBrand, &p.TireModel,
		&p.TirePattern, &p.DOTCode, &p.LoadIndex, &p.SpeedRating,
		&p.TireType, &p.Location, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("receive stock: %w", err)
	}

	// Each receipt is a traceable intake batch. Cost defaults to the product's
	// current buy price when the receipt doesn't state one; the DOT/lot code
	// falls back to the product's if not given per-intake.
	cost := req.UnitCost
	if cost == 0 {
		cost = p.BuyPrice
	}
	dot := req.DOTCode
	if dot == "" {
		dot = p.DOTCode
	}
	// If paid on delivery, mark the whole purchase paid; otherwise it's payable.
	amountPaid := 0.0
	if req.Paid {
		amountPaid = req.Quantity * cost
	}
	batchID, err := batch.Create(ctx, tx, branchID, id, req.Quantity, cost, req.SupplierID, amountPaid, req.Supplier, dot, req.Notes, &userID)
	if err != nil {
		return nil, err
	}
	if err := batch.RecordMovement(ctx, tx, branchID, id, req.Quantity, "received", "batch", &batchID, &batchID, &userID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return &p, nil
}

func (s *Service) AdjustStock(ctx context.Context, branchID int64, id int64, userID int64, req *dto.AdjustStockRequest) (*models.Product, error) {
	_, err := s.Get(ctx, branchID, id)
	if err != nil {
		return nil, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var currentStock, reservedQty float64
	err = tx.QueryRow(ctx, `SELECT stock_quantity, reserved_quantity FROM products WHERE id = $1 AND branch_id = $2 AND is_active = true FOR UPDATE`,
		id, branchID).Scan(&currentStock, &reservedQty)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	newStock := currentStock + req.QuantityChange
	if newStock < 0 {
		return nil, &domain.AppError{Code: "INSUFFICIENT_STOCK", Message: fmt.Sprintf("Cannot adjust stock below 0: have %g, change %g", currentStock, req.QuantityChange), Status: 400}
	}
	if newStock < reservedQty {
		return nil, &domain.AppError{Code: "INSUFFICIENT_STOCK", Message: fmt.Sprintf("Cannot adjust stock below %g — %g units are reserved for scheduled jobs", reservedQty, reservedQty), Status: 400}
	}

	var p models.Product
	err = tx.QueryRow(ctx, `
		UPDATE products
		SET stock_quantity = stock_quantity + $1,
		    updated_at = NOW()
		WHERE id = $2 AND branch_id = $3 AND is_active = true
		RETURNING id, branch_id, type, sku, COALESCE(barcode,''), name, COALESCE(description,''), COALESCE(category,''),
		          buy_price, sell_price, stock_quantity, reserved_quantity, min_stock_alert, unit, is_oil_product, is_bulk, rated_life_km,
		          COALESCE(tire_size,''), COALESCE(tire_brand,''), COALESCE(tire_model,''), COALESCE(tire_pattern,''), COALESCE(dot_code,''),
		          COALESCE(load_index,''), COALESCE(speed_rating,''), COALESCE(tire_type,''), COALESCE(location,''), is_active,
		          created_at, updated_at`,
		req.QuantityChange, id, branchID,
	).Scan(
		&p.ID, &p.BranchID, &p.Type, &p.SKU, &p.Barcode, &p.Name, &p.Description,
		&p.Category, &p.BuyPrice, &p.SellPrice, &p.StockQuantity, &p.ReservedQuantity,
		&p.MinStockAlert, &p.Unit, &p.IsOilProduct, &p.IsBulk, &p.RatedLifeKm, &p.TireSize, &p.TireBrand, &p.TireModel,
		&p.TirePattern, &p.DOTCode, &p.LoadIndex, &p.SpeedRating,
		&p.TireType, &p.Location, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("adjust stock: %w", err)
	}

	if currentStock >= p.MinStockAlert && p.StockQuantity < p.MinStockAlert {
		if err := telegram.LogEvent(ctx, tx, branchID, telegrammodels.TopicAlerts, "low_stock", "product", id, map[string]any{
			"sku": p.SKU, "name": p.Name, "stock_qty": p.StockQuantity, "min_alert": p.MinStockAlert,
		}); err != nil {
			return nil, err
		}
	}

	// Keep the batch ledger in step: a decrease draws down existing batches
	// FIFO; an increase becomes its own small adjustment batch so it stays
	// traceable and SUM(batches.remaining) still equals stock_quantity.
	if req.QuantityChange < 0 {
		if err := batch.ConsumeFIFO(ctx, tx, branchID, id, -req.QuantityChange, req.Reason, "product", &id, &userID); err != nil {
			return nil, err
		}
	} else {
		batchID, err := batch.Create(ctx, tx, branchID, id, req.QuantityChange, p.BuyPrice, nil, 0, "Adjustment", p.DOTCode, req.Reason, &userID)
		if err != nil {
			return nil, err
		}
		if err := batch.RecordMovement(ctx, tx, branchID, id, req.QuantityChange, req.Reason, "batch", &batchID, &batchID, &userID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return &p, nil
}

func (s *Service) Delete(ctx context.Context, branchID int64, id int64) error {
	var reserved float64
	err := s.pool.QueryRow(ctx,
		`SELECT reserved_quantity FROM products WHERE id = $1 AND branch_id = $2 AND is_active = true`,
		id, branchID).Scan(&reserved)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.ErrNotFound
		}
		return fmt.Errorf("check reservation: %w", err)
	}
	if reserved > 0 {
		return &domain.AppError{Code: "PRODUCT_RESERVED", Message: fmt.Sprintf("Cannot delete: %g units are reserved for scheduled jobs", reserved), Status: 400}
	}

	result, err := s.pool.Exec(ctx,
		`UPDATE products SET is_active = false, updated_at = NOW() WHERE id = $1 AND branch_id = $2 AND is_active = true`,
		id, branchID)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ListMovements returns the stock ledger for one product, newest first, with a
// running balance computed backwards from the product's current stock (so it is
// correct even though the opening balance has no ledger row). Movements tied to
// an invoice carry that invoice's number for display.
func (s *Service) ListMovements(ctx context.Context, branchID, productID int64, page, perPage int) ([]models.StockMovement, int, error) {
	if _, err := s.Get(ctx, branchID, productID); err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	rows, err := s.pool.Query(ctx, `
		SELECT m.id, m.quantity_change, m.reason, COALESCE(m.reference_type, ''),
		       COALESCE(m.reference_id, 0), COALESCE(i.invoice_number, ''),
		       CASE WHEN m.batch_id IS NULL THEN ''
		            ELSE 'B-' || to_char(b.received_at, 'YYYY') || '-' || lpad(b.id::text, 4, '0') END AS batch_no,
		       COALESCE(u.full_name, ''),
		       (SELECT stock_quantity FROM products WHERE id = $1)
		         - COALESCE(SUM(m.quantity_change) OVER (
		             ORDER BY m.created_at DESC, m.id DESC
		             ROWS BETWEEN UNBOUNDED PRECEDING AND 1 PRECEDING), 0) AS balance_after,
		       m.created_at,
		       COUNT(*) OVER() AS total
		FROM stock_movements m
		LEFT JOIN invoices i ON m.reference_type = 'invoice' AND i.id = m.reference_id
		LEFT JOIN batches b ON b.id = m.batch_id
		LEFT JOIN users u ON u.id = m.recorded_by
		WHERE m.product_id = $1 AND m.branch_id = $2
		ORDER BY m.created_at DESC, m.id DESC
		LIMIT $3 OFFSET $4`, productID, branchID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query movements: %w", err)
	}
	defer rows.Close()

	var movements []models.StockMovement
	var total int
	for rows.Next() {
		var m models.StockMovement
		if err := rows.Scan(&m.ID, &m.QuantityChange, &m.Reason, &m.ReferenceType,
			&m.ReferenceID, &m.InvoiceNumber, &m.BatchNo, &m.RecordedByName,
			&m.BalanceAfter, &m.CreatedAt, &total); err != nil {
			return nil, 0, fmt.Errorf("scan movement: %w", err)
		}
		movements = append(movements, m)
	}
	if movements == nil {
		movements = []models.StockMovement{}
	}
	return movements, total, nil
}

// ListBatches returns the intake batches for a product, newest first.
func (s *Service) ListBatches(ctx context.Context, branchID, productID int64) ([]models.Batch, error) {
	if _, err := s.Get(ctx, branchID, productID); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT b.id,
		       'B-' || to_char(b.received_at, 'YYYY') || '-' || lpad(b.id::text, 4, '0') AS batch_no,
		       COALESCE(b.supplier, ''), COALESCE(b.dot_code, ''), b.unit_cost,
		       b.quantity_received, b.quantity_remaining, COALESCE(b.notes, ''),
		       COALESCE(u.full_name, ''), b.received_at
		FROM batches b
		LEFT JOIN users u ON u.id = b.received_by
		WHERE b.product_id = $1 AND b.branch_id = $2
		ORDER BY b.received_at DESC, b.id DESC`, productID, branchID)
	if err != nil {
		return nil, fmt.Errorf("query batches: %w", err)
	}
	defer rows.Close()

	var batches []models.Batch
	for rows.Next() {
		var b models.Batch
		if err := rows.Scan(&b.ID, &b.BatchNo, &b.Supplier, &b.DOTCode, &b.UnitCost,
			&b.QuantityReceived, &b.QuantityRemaining, &b.Notes, &b.ReceivedByName, &b.ReceivedAt); err != nil {
			return nil, fmt.Errorf("scan batch: %w", err)
		}
		batches = append(batches, b)
	}
	if batches == nil {
		batches = []models.Batch{}
	}
	return batches, nil
}

// BatchConsumers lists the invoices (and customers) that drew stock from a
// batch — the recall view for "who got a unit from this bad batch".
func (s *Service) BatchConsumers(ctx context.Context, branchID, batchID int64) ([]models.BatchConsumer, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT i.id, i.invoice_number, COALESCE(c.name, 'Walk-in'),
		       -SUM(m.quantity_change) AS qty, MIN(m.created_at)
		FROM stock_movements m
		JOIN batches b ON b.id = m.batch_id AND b.branch_id = $2
		JOIN invoices i ON m.reference_type = 'invoice' AND i.id = m.reference_id
		LEFT JOIN customers c ON c.id = i.customer_id
		WHERE m.batch_id = $1 AND m.reason = 'invoice_issued'
		GROUP BY i.id, i.invoice_number, c.name
		ORDER BY MIN(m.created_at) DESC`, batchID, branchID)
	if err != nil {
		return nil, fmt.Errorf("query consumers: %w", err)
	}
	defer rows.Close()

	var consumers []models.BatchConsumer
	for rows.Next() {
		var c models.BatchConsumer
		if err := rows.Scan(&c.InvoiceID, &c.InvoiceNumber, &c.CustomerName, &c.Quantity, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan consumer: %w", err)
		}
		consumers = append(consumers, c)
	}
	if consumers == nil {
		consumers = []models.BatchConsumer{}
	}
	return consumers, nil
}

// GetLowStock flags products whose *available* stock (on hand minus what's
// already reserved for scheduled jobs) has fallen under the alert threshold —
// a heavily-reserved product can look fine on the shelf count while having
// nothing left to actually sell.
func (s *Service) GetLowStock(ctx context.Context, branchID int64) ([]models.LowStockProduct, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, sku, name, stock_quantity, reserved_quantity, min_stock_alert, sell_price
		FROM products
		WHERE branch_id = $1 AND is_active = true
		  AND (stock_quantity - reserved_quantity) < min_stock_alert
		ORDER BY (stock_quantity - reserved_quantity) ASC`, branchID)
	if err != nil {
		return nil, fmt.Errorf("query low stock: %w", err)
	}
	defer rows.Close()

	var products []models.LowStockProduct
	for rows.Next() {
		var p models.LowStockProduct
		if err := rows.Scan(&p.ID, &p.SKU, &p.Name, &p.StockQuantity, &p.ReservedQuantity, &p.MinStockAlert, &p.SellPrice); err != nil {
			return nil, fmt.Errorf("scan low stock: %w", err)
		}
		products = append(products, p)
	}

	if products == nil {
		products = []models.LowStockProduct{}
	}
	return products, nil
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nullInt(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

func nullFloat(f float64) *float64 {
	if f == 0 {
		return nil
	}
	return &f
}

func ceilDiv(a, b int) int {
	if a == 0 {
		return 0
	}
	return int(math.Ceil(float64(a) / float64(b)))
}
