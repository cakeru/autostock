package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// BOM + utf8 so Excel opens Khmer names/addresses correctly.
const csvBOM = "\xEF\xBB\xBF"

func writeHeader(w *csv.Writer, cols ...string) {
	_ = w.Write(cols)
}

func fmtDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04")
}

func fmtNum(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}

// ExportInvoices writes every non-voided invoice as one CSV row, for the
// accountant's books.
func (s *Service) ExportInvoices(ctx context.Context, branchID int64, w io.Writer) error {
	bw := newCSVWriter(w)
	defer bw.Flush()
	writeHeader(bw,
		"invoice_number", "issued_at", "customer_name", "customer_type", "plate_number",
		"status", "payment_status", "subtotal_usd", "discount_usd", "tax_usd",
		"total_usd", "total_khr", "paid_usd", "payment_method", "created_by", "notes")

	rows, err := s.pool.Query(ctx, `
		SELECT i.invoice_number, COALESCE(i.issued_at, i.created_at), COALESCE(c.name, ''),
		       COALESCE(c.customer_type, 'retail'), COALESCE(v.plate_number, ''),
		       i.status, COALESCE(i.payment_status, ''), i.subtotal, i.discount, i.tax_amount,
		       i.total_usd, i.total_khr, i.paid_amount, COALESCE(i.payment_method, ''),
		       COALESCE(cred.full_name, ''), COALESCE(i.notes, '')
		FROM invoices i
		LEFT JOIN customers c ON c.id = i.customer_id
		LEFT JOIN vehicles v ON v.id = i.vehicle_id
		LEFT JOIN users cred ON cred.id = i.created_by
		WHERE i.branch_id = $1 AND i.status <> 'voided'
		ORDER BY i.created_at`, branchID)
	if err != nil {
		return fmt.Errorf("export invoices: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var number, custName, custType, plate, status, payStatus, payMethod, createdBy, notes string
		var issuedAt time.Time
		var subtotal, discount, tax, totalUSD, totalKHR, paid float64
		if err := rows.Scan(&number, &issuedAt, &custName, &custType, &plate,
			&status, &payStatus, &subtotal, &discount, &tax,
			&totalUSD, &totalKHR, &paid, &payMethod, &createdBy, &notes); err != nil {
			return fmt.Errorf("scan invoice export row: %w", err)
		}
		_ = bw.Write([]string{
			number, issuedAt.Format("2006-01-02 15:04"), custName, custType, plate,
			status, payStatus, fmtNum(subtotal), fmtNum(discount), fmtNum(tax),
			fmtNum(totalUSD), strconv.FormatFloat(totalKHR, 'f', 0, 64), fmtNum(paid),
			payMethod, createdBy, notes,
		})
	}
	return rows.Err()
}

// ExportCustomers writes every active customer plus their vehicles and totals.
func (s *Service) ExportCustomers(ctx context.Context, branchID int64, w io.Writer) error {
	bw := newCSVWriter(w)
	defer bw.Flush()
	writeHeader(bw,
		"name", "customer_type", "phone", "email", "address",
		"vehicles", "vehicle_count", "total_spent_usd", "last_visit", "customer_since")

	rows, err := s.pool.Query(ctx, `
		SELECT c.name, COALESCE(c.customer_type, 'retail'), COALESCE(c.phone,''), COALESCE(c.email,''), COALESCE(c.address,''),
		       COALESCE((SELECT string_agg(v.plate_number, ', ' ORDER BY v.created_at) FROM vehicles v WHERE v.customer_id = c.id), ''),
		       (SELECT COUNT(*) FROM vehicles v WHERE v.customer_id = c.id),
		       COALESCE((SELECT SUM(i.total_usd) FROM invoices i WHERE i.customer_id = c.id AND i.status <> 'voided'), 0),
		       (SELECT MAX(COALESCE(i.issued_at, i.created_at)) FROM invoices i WHERE i.customer_id = c.id AND i.status <> 'voided'),
		       c.customer_since
		FROM customers c
		WHERE c.branch_id = $1 AND c.is_active = true
		ORDER BY c.created_at`, branchID)
	if err != nil {
		return fmt.Errorf("export customers: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var name, custType, phone, email, addr, plates string
		var vehCount int
		var spent float64
		var lastVisit, since *time.Time
		if err := rows.Scan(&name, &custType, &phone, &email, &addr, &plates,
			&vehCount, &spent, &lastVisit, &since); err != nil {
			return fmt.Errorf("scan customer export row: %w", err)
		}
		_ = bw.Write([]string{
			name, custType, phone, email, addr, plates,
			strconv.Itoa(vehCount), fmtNum(spent), fmtDate(lastVisit), fmtDate(since),
		})
	}
	return rows.Err()
}

// ExportProducts writes the current catalog with stock and pricing.
func (s *Service) ExportProducts(ctx context.Context, branchID int64, w io.Writer) error {
	bw := newCSVWriter(w)
	defer bw.Flush()
	writeHeader(bw,
		"sku", "barcode", "type", "name", "category", "unit",
		"buy_price_usd", "sell_price_usd", "stock_quantity", "location",
		"life_km", "life_days", "life_months", "is_oil_product", "is_bulk", "is_active")

	rows, err := s.pool.Query(ctx, `
		SELECT sku, COALESCE(barcode,''), type, name, COALESCE(category,''), COALESCE(unit,'piece'),
		       buy_price, sell_price, stock_quantity, COALESCE(location,''),
		       COALESCE(life_km, 0), COALESCE(life_days, 0), COALESCE(life_months, 0),
		       is_oil_product, is_bulk, is_active
		FROM products WHERE branch_id = $1
		ORDER BY created_at`, branchID)
	if err != nil {
		return fmt.Errorf("export products: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var sku, barcode, ptype, name, category, unit, location string
		var buy, sell, stock, lifeKm, lifeDays, lifeMonths float64
		var isOil, isBulk, isActive bool
		if err := rows.Scan(&sku, &barcode, &ptype, &name, &category, &unit,
			&buy, &sell, &stock, &location, &lifeKm, &lifeDays, &lifeMonths,
			&isOil, &isBulk, &isActive); err != nil {
			return fmt.Errorf("scan product export row: %w", err)
		}
		_ = bw.Write([]string{
			sku, barcode, ptype, name, category, unit,
			fmtNum(buy), fmtNum(sell), strconv.FormatFloat(stock, 'f', -1, 64), location,
			strconv.FormatFloat(lifeKm, 'f', 0, 64), strconv.FormatFloat(lifeDays, 'f', 0, 64),
			strconv.FormatFloat(lifeMonths, 'f', 0, 64),
			strconv.FormatBool(isOil), strconv.FormatBool(isBulk), strconv.FormatBool(isActive),
		})
	}
	return rows.Err()
}

func newCSVWriter(w io.Writer) *csv.Writer {
	bw := csv.NewWriter(w)
	bw.UseCRLF = true // Excel-friendly line endings
	_, _ = io.WriteString(w, csvBOM)
	return bw
}
