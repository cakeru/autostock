package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/batch"
	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/purchaseorder/dto"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) Create(ctx context.Context, branchID, userID int64, req *dto.CreatePORequest) (*dto.POListResponse, error) {
	var supplierExists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM suppliers WHERE id = $1 AND branch_id = $2 AND is_active)`,
		req.SupplierID, branchID).Scan(&supplierExists); err != nil {
		return nil, fmt.Errorf("check supplier: %w", err)
	}
	if !supplierExists {
		return nil, domain.ErrNotFound
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext('po_number'))`)
	if err != nil {
		return nil, fmt.Errorf("advisory lock: %w", err)
	}

	year := fmt.Sprintf("%d", time.Now().Year())
	var seq int
	err = tx.QueryRow(ctx,
		`SELECT COALESCE(MAX(SUBSTRING(po_number FROM 9)::int), 0) + 1
		 FROM purchase_orders WHERE po_number LIKE $1`, "PO-"+year+"-%").Scan(&seq)
	if err != nil {
		return nil, fmt.Errorf("generate po number: %w", err)
	}
	poNumber := fmt.Sprintf("PO-%s-%04d", year, seq)

	var id int64
	err = tx.QueryRow(ctx, `
		INSERT INTO purchase_orders (branch_id, supplier_id, po_number, notes, created_by)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5)
		RETURNING id`,
		branchID, req.SupplierID, poNumber, req.Notes, userID).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create po: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return s.getListItem(ctx, branchID, id)
}

// Update edits a draft order's supplier and notes (ordered/received orders are
// locked — correct them by cancelling and reordering).
func (s *Service) Update(ctx context.Context, branchID, id int64, req *dto.CreatePORequest) (*dto.POListResponse, error) {
	st, err := s.status(ctx, branchID, id)
	if err != nil {
		return nil, err
	}
	if st != "draft" {
		return nil, &domain.AppError{Code: "PO_NOT_DRAFT", Message: "Only draft orders can be edited — cancel and reorder instead", Status: 400}
	}
	var supplierExists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM suppliers WHERE id = $1 AND branch_id = $2 AND is_active)`,
		req.SupplierID, branchID).Scan(&supplierExists); err != nil {
		return nil, fmt.Errorf("check supplier: %w", err)
	}
	if !supplierExists {
		return nil, domain.ErrNotFound
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE purchase_orders SET supplier_id = $1, notes = NULLIF($2, '') WHERE id = $3 AND branch_id = $4`,
		req.SupplierID, req.Notes, id, branchID)
	if err != nil {
		return nil, fmt.Errorf("update po: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}
	return s.getListItem(ctx, branchID, id)
}

func (s *Service) List(ctx context.Context, branchID int64) ([]dto.POListResponse, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT po.id, po.po_number, po.status, po.supplier_id, sup.name, COALESCE(po.notes, ''),
		       po.created_by, COALESCE(u.full_name, ''),
		       po.ordered_at, po.received_at, po.created_at,
		       COUNT(poi.id), COALESCE(SUM(poi.quantity_ordered * poi.unit_cost), 0)
		FROM purchase_orders po
		JOIN suppliers sup ON sup.id = po.supplier_id
		LEFT JOIN users u ON u.id = po.created_by
		LEFT JOIN purchase_order_items poi ON poi.purchase_order_id = po.id
		WHERE po.branch_id = $1
		GROUP BY po.id, sup.name, u.full_name
		ORDER BY po.created_at DESC`, branchID)
	if err != nil {
		return nil, fmt.Errorf("query purchase orders: %w", err)
	}
	defer rows.Close()

	var list []dto.POListResponse
	for rows.Next() {
		var po dto.POListResponse
		if err := rows.Scan(&po.ID, &po.PONumber, &po.Status, &po.SupplierID, &po.SupplierName, &po.Notes,
			&po.CreatedByID, &po.CreatedByName, &po.OrderedAt, &po.ReceivedAt, &po.CreatedAt,
			&po.ItemCount, &po.TotalCost); err != nil {
			return nil, fmt.Errorf("scan po: %w", err)
		}
		list = append(list, po)
	}
	if list == nil {
		list = []dto.POListResponse{}
	}
	return list, nil
}

func (s *Service) Get(ctx context.Context, branchID, id int64) (*dto.PODetailResponse, error) {
	li, err := s.getListItem(ctx, branchID, id)
	if err != nil {
		return nil, err
	}
	det := &dto.PODetailResponse{POListResponse: *li}

	rows, err := s.pool.Query(ctx, `
		SELECT poi.id, poi.product_id, p.sku, p.name, poi.quantity_ordered, poi.quantity_received, poi.unit_cost
		FROM purchase_order_items poi
		JOIN products p ON p.id = poi.product_id
		WHERE poi.purchase_order_id = $1
		ORDER BY poi.created_at`, id)
	if err != nil {
		return nil, fmt.Errorf("query po items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var it dto.POItemResponse
		if err := rows.Scan(&it.ID, &it.ProductID, &it.SKU, &it.ProductName, &it.QuantityOrdered, &it.QuantityReceived, &it.UnitCost); err != nil {
			return nil, fmt.Errorf("scan po item: %w", err)
		}
		it.TotalCost = float64(it.QuantityOrdered) * it.UnitCost
		det.Items = append(det.Items, it)
	}
	if det.Items == nil {
		det.Items = []dto.POItemResponse{}
	}
	return det, nil
}

func (s *Service) getListItem(ctx context.Context, branchID, id int64) (*dto.POListResponse, error) {
	var po dto.POListResponse
	err := s.pool.QueryRow(ctx, `
		SELECT po.id, po.po_number, po.status, po.supplier_id, sup.name, COALESCE(po.notes, ''),
		       po.created_by, COALESCE(u.full_name, ''),
		       po.ordered_at, po.received_at, po.created_at,
		       COUNT(poi.id), COALESCE(SUM(poi.quantity_ordered * poi.unit_cost), 0)
		FROM purchase_orders po
		JOIN suppliers sup ON sup.id = po.supplier_id
		LEFT JOIN users u ON u.id = po.created_by
		LEFT JOIN purchase_order_items poi ON poi.purchase_order_id = po.id
		WHERE po.id = $1 AND po.branch_id = $2
		GROUP BY po.id, sup.name, u.full_name`, id, branchID).
		Scan(&po.ID, &po.PONumber, &po.Status, &po.SupplierID, &po.SupplierName, &po.Notes,
			&po.CreatedByID, &po.CreatedByName, &po.OrderedAt, &po.ReceivedAt, &po.CreatedAt,
			&po.ItemCount, &po.TotalCost)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get po: %w", err)
	}
	return &po, nil
}

func (s *Service) status(ctx context.Context, branchID, id int64) (string, error) {
	var status string
	err := s.pool.QueryRow(ctx, `SELECT status FROM purchase_orders WHERE id = $1 AND branch_id = $2`, id, branchID).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", domain.ErrNotFound
		}
		return "", fmt.Errorf("check po: %w", err)
	}
	return status, nil
}

func (s *Service) AddItem(ctx context.Context, branchID, poID int64, req *dto.AddPOItemRequest) (*dto.POItemResponse, error) {
	status, err := s.status(ctx, branchID, poID)
	if err != nil {
		return nil, err
	}
	if status != "draft" {
		return nil, &domain.AppError{Code: "INVALID_STATUS", Message: "Can only add items to a draft purchase order", Status: 400}
	}

	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1 AND branch_id = $2 AND is_active)`,
		req.ProductID, branchID).Scan(&exists); err != nil || !exists {
		return nil, domain.ErrNotFound
	}

	var it dto.POItemResponse
	err = s.pool.QueryRow(ctx, `
		WITH ins AS (
			INSERT INTO purchase_order_items (purchase_order_id, product_id, quantity_ordered, unit_cost)
			VALUES ($1, $2, $3, $4)
			RETURNING id, product_id, quantity_ordered, quantity_received, unit_cost
		)
		SELECT ins.id, ins.product_id, p.sku, p.name, ins.quantity_ordered, ins.quantity_received, ins.unit_cost
		FROM ins JOIN products p ON p.id = ins.product_id`,
		poID, req.ProductID, req.QuantityOrdered, req.UnitCost).
		Scan(&it.ID, &it.ProductID, &it.SKU, &it.ProductName, &it.QuantityOrdered, &it.QuantityReceived, &it.UnitCost)
	if err != nil {
		return nil, fmt.Errorf("add po item: %w", err)
	}
	it.TotalCost = float64(it.QuantityOrdered) * it.UnitCost
	return &it, nil
}

func (s *Service) RemoveItem(ctx context.Context, branchID, itemID int64) error {
	result, err := s.pool.Exec(ctx, `
		DELETE FROM purchase_order_items poi
		USING purchase_orders po
		WHERE poi.id = $1 AND po.id = poi.purchase_order_id AND po.branch_id = $2 AND po.status = 'draft'`,
		itemID, branchID)
	if err != nil {
		return fmt.Errorf("remove po item: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Service) Place(ctx context.Context, branchID, id, userID int64) (*dto.POListResponse, error) {
	status, err := s.status(ctx, branchID, id)
	if err != nil {
		return nil, err
	}
	if status != "draft" {
		return nil, &domain.AppError{Code: "INVALID_STATUS", Message: "Only a draft purchase order can be placed", Status: 400}
	}

	var itemCount int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM purchase_order_items WHERE purchase_order_id = $1`, id).Scan(&itemCount); err != nil {
		return nil, fmt.Errorf("count items: %w", err)
	}
	if itemCount == 0 {
		return nil, &domain.AppError{Code: "NO_ITEMS", Message: "Add at least one line before placing the order", Status: 400}
	}

	_, err = s.pool.Exec(ctx,
		`UPDATE purchase_orders SET status = 'ordered', ordered_at = NOW(), updated_at = NOW() WHERE id = $1 AND branch_id = $2`,
		id, branchID)
	if err != nil {
		return nil, fmt.Errorf("place po: %w", err)
	}
	_ = userID
	return s.getListItem(ctx, branchID, id)
}

func (s *Service) Cancel(ctx context.Context, branchID, id int64) error {
	result, err := s.pool.Exec(ctx,
		`UPDATE purchase_orders SET status = 'cancelled', updated_at = NOW() WHERE id = $1 AND branch_id = $2 AND status IN ('draft', 'ordered')`,
		id, branchID)
	if err != nil {
		return fmt.Errorf("cancel po: %w", err)
	}
	if result.RowsAffected() == 0 {
		return &domain.AppError{Code: "INVALID_STATUS", Message: "Only a draft or ordered purchase order (nothing received yet) can be cancelled", Status: 400}
	}
	return nil
}

// Receive posts what actually arrived: for each requested line (or every open
// line, if none are specified) it creates a real intake batch — same ledger a
// manual "Receive Stock" action uses — linked back to this PO and its
// supplier, then advances the PO to partial/received depending on how much of
// the full order is now in.
func (s *Service) Receive(ctx context.Context, branchID, poID, userID int64, req *dto.ReceiveRequest) (*dto.ReceiveResult, error) {
	status, err := s.status(ctx, branchID, poID)
	if err != nil {
		return nil, err
	}
	if status != "ordered" && status != "partial" {
		return nil, &domain.AppError{Code: "INVALID_STATUS", Message: "Place the order before receiving against it", Status: 400}
	}

	var supplierID int64
	var poNumber string
	if err := s.pool.QueryRow(ctx, `SELECT supplier_id, po_number FROM purchase_orders WHERE id = $1`, poID).Scan(&supplierID, &poNumber); err != nil {
		return nil, fmt.Errorf("load po: %w", err)
	}
	var supplierName string
	_ = s.pool.QueryRow(ctx, `SELECT name FROM suppliers WHERE id = $1`, supplierID).Scan(&supplierName)

	type line struct {
		itemID, productID       int64
		qtyOrdered, qtyReceived int
		unitCost                float64
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, product_id, quantity_ordered, quantity_received, unit_cost
		FROM purchase_order_items WHERE purchase_order_id = $1`, poID)
	if err != nil {
		return nil, fmt.Errorf("query po items: %w", err)
	}
	var lines []line
	for rows.Next() {
		var l line
		if err := rows.Scan(&l.itemID, &l.productID, &l.qtyOrdered, &l.qtyReceived, &l.unitCost); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan po item: %w", err)
		}
		lines = append(lines, l)
	}
	rows.Close()

	// Map of item_id -> requested quantity, when the caller specified a subset.
	requested := map[int64]int{}
	hasRequest := len(req.Items) > 0
	for _, l := range req.Items {
		if l.Quantity != nil {
			requested[l.ItemID] = *l.Quantity
		} else {
			requested[l.ItemID] = -1 // sentinel: receive full remaining
		}
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	totalReceived := 0
	for _, l := range lines {
		remaining := l.qtyOrdered - l.qtyReceived
		if remaining <= 0 {
			continue
		}
		qty := remaining
		if hasRequest {
			reqQty, ok := requested[l.itemID]
			if !ok {
				continue // this line wasn't part of the request
			}
			if reqQty >= 0 {
				qty = reqQty
			}
		}
		if qty <= 0 {
			continue
		}
		if qty > remaining {
			qty = remaining
		}

		amountPaid := 0.0
		if req.Paid {
			amountPaid = float64(qty) * l.unitCost
		}
		batchID, err := batch.Create(ctx, tx, branchID, l.productID, float64(qty), l.unitCost, &supplierID, amountPaid, supplierName, "", "PO "+poNumber, &userID, req.InvoiceNumber, req.InvoiceImage)
		if err != nil {
			return nil, err
		}
		if err := batch.RecordMovement(ctx, tx, branchID, l.productID, float64(qty), "po_received", "purchase_order", &poID, &batchID, &userID); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `UPDATE products SET stock_quantity = stock_quantity + $1, updated_at = NOW() WHERE id = $2`, qty, l.productID); err != nil {
			return nil, fmt.Errorf("update stock: %w", err)
		}
		if _, err := tx.Exec(ctx, `UPDATE purchase_order_items SET quantity_received = quantity_received + $1 WHERE id = $2`, qty, l.itemID); err != nil {
			return nil, fmt.Errorf("update po item: %w", err)
		}
		totalReceived += qty
	}

	if totalReceived == 0 {
		return nil, &domain.AppError{Code: "NOTHING_TO_RECEIVE", Message: "Nothing was received — every line is already fully received or none were specified", Status: 400}
	}

	var fullyReceived bool
	if err := tx.QueryRow(ctx,
		`SELECT COUNT(*) = 0 FROM purchase_order_items WHERE purchase_order_id = $1 AND quantity_received < quantity_ordered`,
		poID).Scan(&fullyReceived); err != nil {
		return nil, fmt.Errorf("check completion: %w", err)
	}

	newStatus := "partial"
	if fullyReceived {
		newStatus = "received"
		if _, err := tx.Exec(ctx,
			`UPDATE purchase_orders SET status = 'received', received_at = NOW(), updated_at = NOW() WHERE id = $1`, poID); err != nil {
			return nil, fmt.Errorf("complete po: %w", err)
		}
	} else {
		if _, err := tx.Exec(ctx,
			`UPDATE purchase_orders SET status = 'partial', updated_at = NOW() WHERE id = $1`, poID); err != nil {
			return nil, fmt.Errorf("update po status: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &dto.ReceiveResult{Received: totalReceived, Status: newStatus}, nil
}
