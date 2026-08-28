package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/batch"
	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/stocktake/dto"
	telegrammodels "github.com/cakeru/autostock/internal/telegram/models"
	telegram "github.com/cakeru/autostock/internal/telegram/service"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) Create(ctx context.Context, branchID, userID int64, req *dto.CreateStocktakeRequest) (*dto.StocktakeListResponse, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var st dto.StocktakeListResponse
	err = tx.QueryRow(ctx, `
		INSERT INTO stocktakes (branch_id, notes, created_by)
		VALUES ($1, NULLIF($2, ''), $3)
		RETURNING id, status, COALESCE(notes, ''), created_by, created_at`,
		branchID, req.Notes, userID,
	).Scan(&st.ID, &st.Status, &st.Notes, &st.CreatedByID, &st.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create stocktake: %w", err)
	}

	// Optional bulk pre-populate: snapshot every active product matching the
	// given filters at its current stock level.
	if req.Type != "" || req.Category != "" {
		_, err = tx.Exec(ctx, `
			INSERT INTO stocktake_items (stocktake_id, product_id, expected_qty)
			SELECT $1, id, stock_quantity FROM products
			WHERE branch_id = $2 AND is_active = true
			  AND ($3::text = '' OR type = $3)
			  AND ($4::text = '' OR category = $4)`,
			st.ID, branchID, req.Type, req.Category)
		if err != nil {
			return nil, fmt.Errorf("populate stocktake items: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	if st.CreatedByID != nil {
		_ = s.pool.QueryRow(ctx, `SELECT full_name FROM users WHERE id = $1`, *st.CreatedByID).Scan(&st.CreatedByName)
	}
	counts, _ := s.itemCounts(ctx, st.ID)
	st.ItemCount, st.CountedCount, st.VarianceCount = counts.total, counts.counted, counts.variances
	return &st, nil
}

func (s *Service) List(ctx context.Context, branchID int64) ([]dto.StocktakeListResponse, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT st.id, st.status, COALESCE(st.notes, ''), st.created_by, COALESCE(u.full_name, ''),
		       st.created_at, st.completed_at,
		       COUNT(sti.id) AS item_count,
		       COUNT(sti.counted_qty) AS counted_count,
		       COUNT(*) FILTER (WHERE sti.variance IS NOT NULL AND sti.variance != 0) AS variance_count
		FROM stocktakes st
		LEFT JOIN users u ON u.id = st.created_by
		LEFT JOIN stocktake_items sti ON sti.stocktake_id = st.id
		WHERE st.branch_id = $1
		GROUP BY st.id, u.full_name
		ORDER BY st.created_at DESC`, branchID)
	if err != nil {
		return nil, fmt.Errorf("query stocktakes: %w", err)
	}
	defer rows.Close()

	var list []dto.StocktakeListResponse
	for rows.Next() {
		var st dto.StocktakeListResponse
		if err := rows.Scan(&st.ID, &st.Status, &st.Notes, &st.CreatedByID, &st.CreatedByName,
			&st.CreatedAt, &st.CompletedAt, &st.ItemCount, &st.CountedCount, &st.VarianceCount); err != nil {
			return nil, fmt.Errorf("scan stocktake: %w", err)
		}
		list = append(list, st)
	}
	if list == nil {
		list = []dto.StocktakeListResponse{}
	}
	return list, nil
}

func (s *Service) Get(ctx context.Context, branchID, id int64) (*dto.StocktakeDetailResponse, error) {
	var st dto.StocktakeDetailResponse
	err := s.pool.QueryRow(ctx, `
		SELECT st.id, st.status, COALESCE(st.notes, ''), st.created_by, COALESCE(u.full_name, ''),
		       st.created_at, st.completed_at
		FROM stocktakes st
		LEFT JOIN users u ON u.id = st.created_by
		WHERE st.id = $1 AND st.branch_id = $2`, id, branchID).
		Scan(&st.ID, &st.Status, &st.Notes, &st.CreatedByID, &st.CreatedByName, &st.CreatedAt, &st.CompletedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get stocktake: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT sti.id, sti.product_id, p.sku, COALESCE(p.barcode, ''), p.name,
		       sti.expected_qty, sti.counted_qty, sti.variance,
		       COALESCE(u.full_name, ''), sti.counted_at
		FROM stocktake_items sti
		JOIN products p ON p.id = sti.product_id
		LEFT JOIN users u ON u.id = sti.counted_by
		WHERE sti.stocktake_id = $1
		ORDER BY p.name`, id)
	if err != nil {
		return nil, fmt.Errorf("query stocktake items: %w", err)
	}
	defer rows.Close()

	var counted, variances int
	for rows.Next() {
		var item dto.StocktakeItemResponse
		if err := rows.Scan(&item.ID, &item.ProductID, &item.SKU, &item.Barcode, &item.ProductName,
			&item.ExpectedQty, &item.CountedQty, &item.Variance, &item.CountedByName, &item.CountedAt); err != nil {
			return nil, fmt.Errorf("scan stocktake item: %w", err)
		}
		if item.CountedQty != nil {
			counted++
		}
		if item.Variance != nil && *item.Variance != 0 {
			variances++
		}
		st.Items = append(st.Items, item)
	}
	if st.Items == nil {
		st.Items = []dto.StocktakeItemResponse{}
	}
	st.ItemCount = len(st.Items)
	st.CountedCount = counted
	st.VarianceCount = variances

	return &st, nil
}

func (s *Service) AddItem(ctx context.Context, branchID, stocktakeID int64, productID int64) (*dto.StocktakeItemResponse, error) {
	status, err := s.status(ctx, branchID, stocktakeID)
	if err != nil {
		return nil, err
	}
	if status != "draft" {
		return nil, &domain.AppError{Code: "INVALID_STATUS", Message: "Can only add items to a draft stocktake", Status: 400}
	}

	var stockQty int
	err = s.pool.QueryRow(ctx,
		`SELECT stock_quantity FROM products WHERE id = $1 AND branch_id = $2 AND is_active = true`,
		productID, branchID).Scan(&stockQty)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("check product: %w", err)
	}

	var itemID int64
	err = s.pool.QueryRow(ctx, `
		INSERT INTO stocktake_items (stocktake_id, product_id, expected_qty)
		VALUES ($1, $2, $3)
		RETURNING id`,
		stocktakeID, productID, stockQty).Scan(&itemID)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, &domain.AppError{Code: "ALREADY_ADDED", Message: "This product is already on the count sheet", Status: 400}
		}
		return nil, fmt.Errorf("add item: %w", err)
	}

	return s.getItem(ctx, itemID)
}

func (s *Service) SetCount(ctx context.Context, branchID, itemID, userID int64, countedQty float64) (*dto.StocktakeItemResponse, error) {
	var status string
	err := s.pool.QueryRow(ctx, `
		SELECT st.status FROM stocktake_items sti
		JOIN stocktakes st ON st.id = sti.stocktake_id
		WHERE sti.id = $1 AND st.branch_id = $2`, itemID, branchID).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("check item: %w", err)
	}
	if status != "draft" {
		return nil, &domain.AppError{Code: "INVALID_STATUS", Message: "Can only enter counts on a draft stocktake", Status: 400}
	}

	// The variance shown here (against expected_qty as snapshotted when the
	// line was added) is informational for the counting staff. The
	// authoritative variance is recalculated against live stock at Finalize,
	// so counts entered mid-shift don't get double-adjusted for sales that
	// happen after this count is taken but before the sheet is finalized.
	_, err = s.pool.Exec(ctx, `
		UPDATE stocktake_items
		SET counted_qty = $1, variance = $1 - expected_qty, counted_by = $2, counted_at = NOW()
		WHERE id = $3`, countedQty, userID, itemID)
	if err != nil {
		return nil, fmt.Errorf("set count: %w", err)
	}

	return s.getItem(ctx, itemID)
}

func (s *Service) RemoveItem(ctx context.Context, branchID, itemID int64) error {
	result, err := s.pool.Exec(ctx, `
		DELETE FROM stocktake_items sti
		USING stocktakes st
		WHERE sti.id = $1 AND st.id = sti.stocktake_id AND st.branch_id = $2 AND st.status = 'draft'`,
		itemID, branchID)
	if err != nil {
		return fmt.Errorf("remove item: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Service) Cancel(ctx context.Context, branchID, id int64) error {
	result, err := s.pool.Exec(ctx,
		`UPDATE stocktakes SET status = 'cancelled', updated_at = NOW() WHERE id = $1 AND branch_id = $2 AND status = 'draft'`,
		id, branchID)
	if err != nil {
		return fmt.Errorf("cancel stocktake: %w", err)
	}
	if result.RowsAffected() == 0 {
		return &domain.AppError{Code: "INVALID_STATUS", Message: "Only a draft stocktake can be cancelled", Status: 400}
	}
	return nil
}

// Finalize applies every counted line as a real stock adjustment, recomputed
// against each product's *live* stock (not the stale snapshot from when the
// line was added), so legitimate sales/receipts mid-count aren't misread as
// shrinkage. Lines left uncounted are simply skipped. Each line is independent
// so one bad line (e.g. would drop stock below what's reserved for a job)
// doesn't block the rest of the sheet.
func (s *Service) Finalize(ctx context.Context, branchID, id, userID int64) (*dto.FinalizeResult, error) {
	status, err := s.status(ctx, branchID, id)
	if err != nil {
		return nil, err
	}
	if status != "draft" {
		return nil, &domain.AppError{Code: "INVALID_STATUS", Message: "Only a draft stocktake can be finalized", Status: 400}
	}

	rows, err := s.pool.Query(ctx, `
		SELECT sti.id, sti.product_id, p.sku, sti.counted_qty
		FROM stocktake_items sti
		JOIN products p ON p.id = sti.product_id
		WHERE sti.stocktake_id = $1 AND sti.counted_qty IS NOT NULL`, id)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	type line struct {
		itemID, productID int64
		sku               string
		countedQty        float64
	}
	var lines []line
	for rows.Next() {
		var l line
		if err := rows.Scan(&l.itemID, &l.productID, &l.sku, &l.countedQty); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan item: %w", err)
		}
		lines = append(lines, l)
	}
	rows.Close()

	result := &dto.FinalizeResult{}
	for _, l := range lines {
		variance, err := s.applyStocktakeLine(ctx, branchID, id, userID, l.itemID, l.productID, l.countedQty)
		if err != nil {
			result.Skipped++
			result.Errors = append(result.Errors, dto.FinalizeItemResult{ProductID: l.productID, SKU: l.sku, Message: err.Error()})
			continue
		}
		if variance == 0 {
			result.Unchanged++
		} else {
			result.Adjusted++
		}
	}

	_, err = s.pool.Exec(ctx,
		`UPDATE stocktakes SET status = 'completed', completed_by = $1, completed_at = NOW(), updated_at = NOW() WHERE id = $2 AND branch_id = $3`,
		userID, id, branchID)
	if err != nil {
		return nil, fmt.Errorf("complete stocktake: %w", err)
	}

	return result, nil
}

// applyStocktakeLine re-checks live stock for one product and, if the count
// differs, applies the adjustment within its own transaction (guarded the
// same way a manual adjustment is: never below zero, never below what's
// reserved for scheduled jobs).
func (s *Service) applyStocktakeLine(ctx context.Context, branchID, stocktakeID, userID, itemID, productID int64, countedQty float64) (float64, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var liveStock, reserved float64
	var sku, name string
	err = tx.QueryRow(ctx,
		`SELECT stock_quantity, reserved_quantity, sku, name FROM products WHERE id = $1 AND branch_id = $2 AND is_active = true FOR UPDATE`,
		productID, branchID).Scan(&liveStock, &reserved, &sku, &name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("product no longer exists")
		}
		return 0, fmt.Errorf("check stock: %w", err)
	}

	variance := countedQty - liveStock
	if variance == 0 {
		if _, err := tx.Exec(ctx, `UPDATE stocktake_items SET variance = 0 WHERE id = $1`, itemID); err != nil {
			return 0, fmt.Errorf("record variance: %w", err)
		}
		return 0, tx.Commit(ctx)
	}

	if countedQty < reserved {
		return 0, fmt.Errorf("counted %g is below %g units reserved for scheduled jobs", countedQty, reserved)
	}

	if _, err := tx.Exec(ctx, `UPDATE products SET stock_quantity = $1, updated_at = NOW() WHERE id = $2`, countedQty, productID); err != nil {
		return 0, fmt.Errorf("update stock: %w", err)
	}

	refID := stocktakeID
	if variance > 0 {
		batchID, err := batch.Create(ctx, tx, branchID, productID, variance, 0, nil, 0, "Stocktake", "", "Found on physical count", &userID, "", "")
		if err != nil {
			return 0, err
		}
		if err := batch.RecordMovement(ctx, tx, branchID, productID, variance, "stocktake_found", "stocktake", &refID, &batchID, &userID); err != nil {
			return 0, err
		}
	} else {
		if err := batch.ConsumeFIFO(ctx, tx, branchID, productID, -variance, "stocktake_shrinkage", "stocktake", &refID, &userID); err != nil {
			return 0, err
		}
		if err := telegram.LogEvent(ctx, tx, branchID, telegrammodels.TopicAlerts, "stocktake_shrinkage", "product", productID, map[string]any{
			"sku": sku, "name": name, "variance": variance,
		}); err != nil {
			return 0, err
		}
	}

	if _, err := tx.Exec(ctx, `UPDATE stocktake_items SET variance = $1 WHERE id = $2`, variance, itemID); err != nil {
		return 0, fmt.Errorf("record variance: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return variance, nil
}

func (s *Service) status(ctx context.Context, branchID, id int64) (string, error) {
	var status string
	err := s.pool.QueryRow(ctx, `SELECT status FROM stocktakes WHERE id = $1 AND branch_id = $2`, id, branchID).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", domain.ErrNotFound
		}
		return "", fmt.Errorf("check stocktake: %w", err)
	}
	return status, nil
}

type itemCountRow struct {
	total, counted, variances int
}

func (s *Service) itemCounts(ctx context.Context, stocktakeID int64) (itemCountRow, error) {
	var c itemCountRow
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(counted_qty), COUNT(*) FILTER (WHERE variance IS NOT NULL AND variance != 0)
		FROM stocktake_items WHERE stocktake_id = $1`, stocktakeID).Scan(&c.total, &c.counted, &c.variances)
	return c, err
}

func (s *Service) getItem(ctx context.Context, itemID int64) (*dto.StocktakeItemResponse, error) {
	var item dto.StocktakeItemResponse
	err := s.pool.QueryRow(ctx, `
		SELECT sti.id, sti.product_id, p.sku, COALESCE(p.barcode, ''), p.name,
		       sti.expected_qty, sti.counted_qty, sti.variance,
		       COALESCE(u.full_name, ''), sti.counted_at
		FROM stocktake_items sti
		JOIN products p ON p.id = sti.product_id
		LEFT JOIN users u ON u.id = sti.counted_by
		WHERE sti.id = $1`, itemID).
		Scan(&item.ID, &item.ProductID, &item.SKU, &item.Barcode, &item.ProductName,
			&item.ExpectedQty, &item.CountedQty, &item.Variance, &item.CountedByName, &item.CountedAt)
	if err != nil {
		return nil, fmt.Errorf("get item: %w", err)
	}
	return &item, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
