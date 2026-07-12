package service

import (
	"context"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/dashboard/dto"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) GetSummary(ctx context.Context, branchID int64) (*dto.SummaryResponse, error) {
	resp := &dto.SummaryResponse{}

	err := s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(total_usd), 0) FROM invoices
		 WHERE branch_id = $1 AND status = 'paid' AND created_at >= CURRENT_DATE`, branchID).Scan(&resp.TodayRevenue)
	if err != nil {
		return nil, fmt.Errorf("query today revenue: %w", err)
	}

	err = s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM service_jobs
		 WHERE branch_id = $1 AND created_at >= CURRENT_DATE`, branchID).Scan(&resp.TodayJobs)
	if err != nil {
		return nil, fmt.Errorf("query today jobs: %w", err)
	}

	err = s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM customers
		 WHERE branch_id = $1 AND is_active = true`, branchID).Scan(&resp.TotalCustomers)
	if err != nil {
		return nil, fmt.Errorf("query total customers: %w", err)
	}

	err = s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM products
		 WHERE branch_id = $1 AND is_active = true AND stock_quantity < min_stock_alert`, branchID).Scan(&resp.LowStockCount)
	if err != nil {
		return nil, fmt.Errorf("query low stock count: %w", err)
	}

	err = s.pool.QueryRow(ctx,
		`SELECT COUNT(*), COALESCE(SUM(total_usd - paid_amount), 0) FROM invoices
		 WHERE branch_id = $1 AND status = 'issued' AND payment_status != 'paid'`, branchID).
		Scan(&resp.UnpaidCount, &resp.OutstandingUSD)
	if err != nil {
		return nil, fmt.Errorf("query outstanding: %w", err)
	}

	invoiceRows, err := s.pool.Query(ctx, `
		SELECT i.id, i.invoice_number, COALESCE(c.name, 'Walk-in') as customer_name, i.total_usd, i.status
		FROM invoices i
		LEFT JOIN customers c ON c.id = i.customer_id
		WHERE i.branch_id = $1
		ORDER BY i.created_at DESC
		LIMIT 5`, branchID)
	if err != nil {
		return nil, fmt.Errorf("query recent invoices: %w", err)
	}
	defer invoiceRows.Close()

	for invoiceRows.Next() {
		var inv dto.RecentInvoice
		if err := invoiceRows.Scan(&inv.ID, &inv.InvoiceNumber, &inv.CustomerName, &inv.TotalUSD, &inv.Status); err != nil {
			return nil, fmt.Errorf("scan invoice: %w", err)
		}
		resp.RecentInvoices = append(resp.RecentInvoices, inv)
	}
	if resp.RecentInvoices == nil {
		resp.RecentInvoices = []dto.RecentInvoice{}
	}

	jobRows, err := s.pool.Query(ctx, `
		SELECT sj.id, sj.job_number, COALESCE(c.name, '') as customer_name, sj.status
		FROM service_jobs sj
		LEFT JOIN customers c ON c.id = sj.customer_id
		WHERE sj.branch_id = $1
		ORDER BY sj.created_at DESC
		LIMIT 5`, branchID)
	if err != nil {
		return nil, fmt.Errorf("query recent jobs: %w", err)
	}
	defer jobRows.Close()

	for jobRows.Next() {
		var job dto.RecentJob
		if err := jobRows.Scan(&job.ID, &job.JobNumber, &job.CustomerName, &job.Status); err != nil {
			return nil, fmt.Errorf("scan job: %w", err)
		}
		resp.RecentJobs = append(resp.RecentJobs, job)
	}
	if resp.RecentJobs == nil {
		resp.RecentJobs = []dto.RecentJob{}
	}

	return resp, nil
}

func (s *Service) GetDailyRevenue(ctx context.Context, branchID int64) ([]dto.DailyRevenueItem, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			DATE(created_at)::text as date,
			COALESCE(SUM(total_usd), 0) as revenue_usd,
			COUNT(*) as invoice_count
		FROM invoices
		WHERE branch_id = $1 AND status = 'paid'
		  AND created_at >= CURRENT_DATE - INTERVAL '6 days'
		GROUP BY DATE(created_at)
		ORDER BY date ASC`, branchID)
	if err != nil {
		return nil, fmt.Errorf("query daily revenue: %w", err)
	}
	defer rows.Close()

	var items []dto.DailyRevenueItem
	for rows.Next() {
		var item dto.DailyRevenueItem
		if err := rows.Scan(&item.Date, &item.RevenueUSD, &item.InvoiceCount); err != nil {
			return nil, fmt.Errorf("scan daily revenue: %w", err)
		}
		items = append(items, item)
	}

	if items == nil {
		items = []dto.DailyRevenueItem{}
	}
	return items, nil
}

// GetProfit computes revenue, cost of goods sold, and gross margin over a date
// range (by invoice issue date, excluding voided invoices). COGS is the true
// FIFO cost: each product line's stock came from specific batches, and each of
// those consumption movements is joined to its batch's unit_cost. Labor and fee
// lines have no COGS.
func (s *Service) GetProfit(ctx context.Context, branchID int64, from, to string) (*dto.ProfitResponse, error) {
	resp := &dto.ProfitResponse{From: from, To: to}

	// Revenue by category from invoice line items on non-voided invoices in range.
	rows, err := s.pool.Query(ctx, `
		SELECT CASE WHEN ii.item_type = 'product' THEN 'Parts & tires'
		            WHEN ii.item_type = 'labor' THEN 'Labor'
		            ELSE 'Fees' END AS category,
		       COALESCE(SUM(ii.total_usd), 0)
		FROM invoice_items ii
		JOIN invoices i ON i.id = ii.invoice_id
		WHERE i.branch_id = $1 AND i.status <> 'voided'
		  AND i.issued_at::date BETWEEN $2::date AND $3::date
		GROUP BY category`, branchID, from, to)
	if err != nil {
		return nil, fmt.Errorf("query revenue: %w", err)
	}
	cats := map[string]*dto.ProfitCategory{
		"Parts & tires": {Category: "Parts & tires"},
		"Labor":         {Category: "Labor"},
		"Fees":          {Category: "Fees"},
	}
	for rows.Next() {
		var name string
		var rev float64
		if err := rows.Scan(&name, &rev); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan revenue: %w", err)
		}
		if c, ok := cats[name]; ok {
			c.Revenue = rev
		}
	}
	rows.Close()

	// COGS: FIFO batch cost of the units sold on those invoices.
	var cogs float64
	err = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(-m.quantity_change * b.unit_cost), 0)
		FROM stock_movements m
		JOIN batches b ON b.id = m.batch_id
		JOIN invoices i ON m.reference_type = 'invoice' AND i.id = m.reference_id
		WHERE i.branch_id = $1 AND i.status <> 'voided'
		  AND m.reason = 'invoice_issued'
		  AND i.issued_at::date BETWEEN $2::date AND $3::date`, branchID, from, to).Scan(&cogs)
	if err != nil {
		return nil, fmt.Errorf("query cogs: %w", err)
	}
	cats["Parts & tires"].Cost = math.Round(cogs*100) / 100

	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM invoices
		WHERE branch_id = $1 AND status <> 'voided'
		  AND issued_at::date BETWEEN $2::date AND $3::date`, branchID, from, to).Scan(&resp.InvoiceCount)
	if err != nil {
		return nil, fmt.Errorf("count invoices: %w", err)
	}

	// Negotiated discounts reduce what the shop actually collected.
	var discounts float64
	err = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(discount), 0) FROM invoices
		WHERE branch_id = $1 AND status <> 'voided'
		  AND issued_at::date BETWEEN $2::date AND $3::date`, branchID, from, to).Scan(&discounts)
	if err != nil {
		return nil, fmt.Errorf("query discounts: %w", err)
	}

	for _, name := range []string{"Parts & tires", "Labor", "Fees"} {
		c := cats[name]
		c.Profit = math.Round((c.Revenue-c.Cost)*100) / 100
		if c.Revenue > 0 {
			c.MarginPct = math.Round(c.Profit/c.Revenue*1000) / 10
		}
		resp.GrossRevenue += c.Revenue
		resp.Cost += c.Cost
		resp.Categories = append(resp.Categories, *c)
	}
	if discounts > 0 {
		resp.Categories = append(resp.Categories, dto.ProfitCategory{Category: "Discounts", Revenue: -discounts, Profit: -discounts})
	}
	resp.Discounts = math.Round(discounts*100) / 100
	resp.GrossRevenue = math.Round(resp.GrossRevenue*100) / 100
	resp.Revenue = math.Round((resp.GrossRevenue-discounts)*100) / 100
	resp.Cost = math.Round(resp.Cost*100) / 100
	resp.GrossProfit = math.Round((resp.Revenue-resp.Cost)*100) / 100
	if resp.Revenue > 0 {
		resp.MarginPct = math.Round(resp.GrossProfit/resp.Revenue*1000) / 10
	}
	return resp, nil
}

func (s *Service) GetDayClose(ctx context.Context, branchID int64, date string) (*dto.DayCloseResponse, error) {
	resp := &dto.DayCloseResponse{Date: date}

	// Payments by method
	rows, err := s.pool.Query(ctx, `
		SELECT p.method, COUNT(*), COALESCE(SUM(p.amount), 0)
		FROM payments p
		JOIN invoices i ON i.id = p.invoice_id
		WHERE i.branch_id = $1 AND DATE(p.created_at) = $2::date
		GROUP BY p.method
		ORDER BY p.method`, branchID, date)
	if err != nil {
		return nil, fmt.Errorf("query payments by method: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var pm dto.PaymentMethodTotal
		if err := rows.Scan(&pm.Method, &pm.Count, &pm.Total); err != nil {
			return nil, fmt.Errorf("scan payment method: %w", err)
		}
		resp.PaymentMethods = append(resp.PaymentMethods, pm)
	}
	if resp.PaymentMethods == nil {
		resp.PaymentMethods = []dto.PaymentMethodTotal{}
	}

	// Voided invoices
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(total_usd), 0)
		FROM invoices
		WHERE branch_id = $1 AND status = 'voided' AND DATE(voided_at) = $2::date`, branchID, date).
		Scan(&resp.VoidedCount, &resp.VoidedTotal)
	if err != nil {
		return nil, fmt.Errorf("query voided: %w", err)
	}

	// Per-user totals
	userRows, err := s.pool.Query(ctx, `
		SELECT p.received_by, COALESCE(u.full_name, '') as user_name, COUNT(*), COALESCE(SUM(p.amount), 0)
		FROM payments p
		JOIN invoices i ON i.id = p.invoice_id
		LEFT JOIN users u ON u.id = p.received_by
		WHERE i.branch_id = $1 AND DATE(p.created_at) = $2::date
		GROUP BY p.received_by, u.full_name
		ORDER BY user_name`, branchID, date)
	if err != nil {
		return nil, fmt.Errorf("query user totals: %w", err)
	}
	defer userRows.Close()
	for userRows.Next() {
		var ut dto.UserPaymentTotal
		if err := userRows.Scan(&ut.UserID, &ut.UserName, &ut.Count, &ut.Total); err != nil {
			return nil, fmt.Errorf("scan user total: %w", err)
		}
		resp.UserTotals = append(resp.UserTotals, ut)
	}
	if resp.UserTotals == nil {
		resp.UserTotals = []dto.UserPaymentTotal{}
	}

	return resp, nil
}
