package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/analytics/dto"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

// ---------------------------------------------------------------------------
// Sales trends + period-over-period comparison
// ---------------------------------------------------------------------------

func (s *Service) GetSales(ctx context.Context, branchID int64, from, to, granularity string) (*dto.SalesResponse, error) {
	if granularity != "week" && granularity != "month" {
		granularity = "day"
	}
	resp := &dto.SalesResponse{From: from, To: to, Granularity: granularity}

	// Totals for the range (non-voided invoices, by issue date).
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(total_usd), 0), COUNT(*), COALESCE(SUM(tax_amount), 0)
		FROM invoices
		WHERE branch_id = $1 AND status <> 'voided'
		  AND COALESCE(issued_at, created_at)::date BETWEEN $2::date AND $3::date`,
		branchID, from, to).Scan(&resp.RevenueUSD, &resp.InvoiceCount, &resp.TaxCollected)
	if err != nil {
		return nil, fmt.Errorf("sales totals: %w", err)
	}
	if resp.InvoiceCount > 0 {
		resp.AvgTicket = round2(resp.RevenueUSD / float64(resp.InvoiceCount))
	}
	resp.RevenueUSD = round2(resp.RevenueUSD)
	resp.TaxCollected = round2(resp.TaxCollected)

	// Previous period of equal length, immediately before [from, to].
	prevFrom, prevTo := previousPeriod(from, to)
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(total_usd), 0)
		FROM invoices
		WHERE branch_id = $1 AND status <> 'voided'
		  AND COALESCE(issued_at, created_at)::date BETWEEN $2::date AND $3::date`,
		branchID, prevFrom, prevTo).Scan(&resp.PrevRevenueUSD); err != nil {
		return nil, fmt.Errorf("prev sales: %w", err)
	}
	resp.PrevRevenueUSD = round2(resp.PrevRevenueUSD)
	if resp.PrevRevenueUSD > 0 {
		resp.RevenueChangePct = round2((resp.RevenueUSD - resp.PrevRevenueUSD) / resp.PrevRevenueUSD * 100)
	}

	// Time series bucketed by the requested granularity.
	rows, err := s.pool.Query(ctx, `
		SELECT to_char(date_trunc($4, COALESCE(issued_at, created_at)), 'YYYY-MM-DD') AS period,
		       COALESCE(SUM(total_usd), 0), COUNT(*)
		FROM invoices
		WHERE branch_id = $1 AND status <> 'voided'
		  AND COALESCE(issued_at, created_at)::date BETWEEN $2::date AND $3::date
		GROUP BY 1 ORDER BY 1`,
		branchID, from, to, granularity)
	if err != nil {
		return nil, fmt.Errorf("sales series: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var p dto.SalesPoint
		if err := rows.Scan(&p.Period, &p.RevenueUSD, &p.InvoiceCount); err != nil {
			return nil, fmt.Errorf("scan series: %w", err)
		}
		p.RevenueUSD = round2(p.RevenueUSD)
		resp.Series = append(resp.Series, p)
	}
	if resp.Series == nil {
		resp.Series = []dto.SalesPoint{}
	}

	// Payment method mix (by payment date in range).
	pmRows, err := s.pool.Query(ctx, `
		SELECT p.method, COUNT(*), COALESCE(SUM(p.amount), 0)
		FROM payments p
		JOIN invoices i ON i.id = p.invoice_id
		WHERE i.branch_id = $1 AND p.created_at::date BETWEEN $2::date AND $3::date
		GROUP BY p.method ORDER BY 3 DESC`,
		branchID, from, to)
	if err != nil {
		return nil, fmt.Errorf("payment mix: %w", err)
	}
	defer pmRows.Close()
	for pmRows.Next() {
		var pm dto.PaymentMethodTotal
		if err := pmRows.Scan(&pm.Method, &pm.Count, &pm.Total); err != nil {
			return nil, fmt.Errorf("scan payment mix: %w", err)
		}
		pm.Total = round2(pm.Total)
		resp.PaymentMethods = append(resp.PaymentMethods, pm)
	}
	if resp.PaymentMethods == nil {
		resp.PaymentMethods = []dto.PaymentMethodTotal{}
	}

	return resp, nil
}

func previousPeriod(from, to string) (string, string) {
	layout := "2006-01-02"
	f, err1 := time.Parse(layout, from)
	t, err2 := time.Parse(layout, to)
	if err1 != nil || err2 != nil {
		return from, to
	}
	days := int(t.Sub(f).Hours()/24) + 1
	prevTo := f.AddDate(0, 0, -1)
	prevFrom := prevTo.AddDate(0, 0, -(days - 1))
	return prevFrom.Format(layout), prevTo.Format(layout)
}

// ---------------------------------------------------------------------------
// Receivables (AR aging)
// ---------------------------------------------------------------------------

func (s *Service) GetReceivables(ctx context.Context, branchID int64) (*dto.ReceivablesResponse, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT i.customer_id, COALESCE(c.name, 'Walk-in'),
		       (i.total_usd - i.paid_amount) AS outstanding,
		       GREATEST(0, CURRENT_DATE - COALESCE(i.issued_at, i.created_at)::date) AS age_days
		FROM invoices i
		LEFT JOIN customers c ON c.id = i.customer_id
		WHERE i.branch_id = $1 AND i.status <> 'voided'
		  AND i.payment_status IN ('unpaid', 'partial')
		  AND (i.total_usd - i.paid_amount) > 0`,
		branchID)
	if err != nil {
		return nil, fmt.Errorf("receivables: %w", err)
	}
	defer rows.Close()

	buckets := []dto.ARBucket{{Label: "0-30"}, {Label: "31-60"}, {Label: "61-90"}, {Label: "90+"}}
	type agg struct {
		id     *int64
		name   string
		count  int
		out    float64
		oldest int
	}
	custs := map[string]*agg{}
	var order []string
	var total float64

	for rows.Next() {
		var cid *int64
		var name string
		var out float64
		var age int
		if err := rows.Scan(&cid, &name, &out, &age); err != nil {
			return nil, fmt.Errorf("scan receivable: %w", err)
		}
		total += out
		bi := 3
		switch {
		case age <= 30:
			bi = 0
		case age <= 60:
			bi = 1
		case age <= 90:
			bi = 2
		}
		buckets[bi].Count++
		buckets[bi].OutstandingUSD += out

		key := name
		if cid != nil {
			key = fmt.Sprintf("c%d", *cid)
		}
		a := custs[key]
		if a == nil {
			a = &agg{id: cid, name: name}
			custs[key] = a
			order = append(order, key)
		}
		a.count++
		a.out += out
		if age > a.oldest {
			a.oldest = age
		}
	}

	resp := &dto.ReceivablesResponse{TotalOutstanding: round2(total), Buckets: buckets}
	for i := range resp.Buckets {
		resp.Buckets[i].OutstandingUSD = round2(resp.Buckets[i].OutstandingUSD)
	}
	for _, k := range order {
		a := custs[k]
		resp.Customers = append(resp.Customers, dto.ARCustomer{
			CustomerID: a.id, CustomerName: a.name, InvoiceCount: a.count,
			OutstandingUSD: round2(a.out), OldestDays: a.oldest,
		})
	}
	sort.Slice(resp.Customers, func(i, j int) bool {
		return resp.Customers[i].OutstandingUSD > resp.Customers[j].OutstandingUSD
	})
	if resp.Customers == nil {
		resp.Customers = []dto.ARCustomer{}
	}
	return resp, nil
}

// ---------------------------------------------------------------------------
// Inventory intelligence
// ---------------------------------------------------------------------------

func (s *Service) GetInventory(ctx context.Context, branchID int64, windowDays int) (*dto.InventoryResponse, error) {
	if windowDays <= 0 {
		windowDays = 90
	}
	resp := &dto.InventoryResponse{WindowDays: windowDays}

	// Valuation: FIFO on-hand cost from remaining batches, plus buy_price fallback
	// for products that have stock but no tracked batches.
	var batchCost, noBatchCost float64
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(b.quantity_remaining * b.unit_cost), 0)
		FROM batches b JOIN products p ON p.id = b.product_id
		WHERE p.branch_id = $1 AND p.is_active AND b.quantity_remaining > 0`, branchID).Scan(&batchCost); err != nil {
		return nil, fmt.Errorf("batch cost: %w", err)
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(p.buy_price * p.stock_quantity), 0)
		FROM products p
		WHERE p.branch_id = $1 AND p.is_active AND p.stock_quantity > 0
		  AND NOT EXISTS (SELECT 1 FROM batches b WHERE b.product_id = p.id AND b.quantity_remaining > 0)`,
		branchID).Scan(&noBatchCost); err != nil {
		return nil, fmt.Errorf("no-batch cost: %w", err)
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(sell_price * stock_quantity), 0), COALESCE(SUM(stock_quantity), 0), COUNT(*)
		FROM products WHERE branch_id = $1 AND is_active`, branchID).
		Scan(&resp.Valuation.RetailValue, &resp.Valuation.UnitsOnHand, &resp.Valuation.SKUCount); err != nil {
		return nil, fmt.Errorf("retail value: %w", err)
	}
	resp.Valuation.CostValue = round2(batchCost + noBatchCost)
	resp.Valuation.RetailValue = round2(resp.Valuation.RetailValue)
	resp.Valuation.PotentialProfit = round2(resp.Valuation.RetailValue - resp.Valuation.CostValue)

	// Units sold + revenue per product in the window.
	sold := map[int64]*dto.ProductStat{}
	rows, err := s.pool.Query(ctx, `
		SELECT ii.product_id, p.name, p.type, p.stock_quantity,
		       COALESCE(SUM(ii.quantity), 0), COALESCE(SUM(ii.total_usd), 0)
		FROM invoice_items ii
		JOIN invoices i ON i.id = ii.invoice_id
		JOIN products p ON p.id = ii.product_id
		WHERE i.branch_id = $1 AND i.status <> 'voided' AND ii.product_id IS NOT NULL
		  AND COALESCE(i.issued_at, i.created_at) >= NOW() - make_interval(days => $2)
		GROUP BY ii.product_id, p.name, p.type, p.stock_quantity`,
		branchID, windowDays)
	if err != nil {
		return nil, fmt.Errorf("sold per product: %w", err)
	}
	for rows.Next() {
		var st dto.ProductStat
		if err := rows.Scan(&st.ProductID, &st.Name, &st.Type, &st.StockQty, &st.QtySold, &st.RevenueUSD); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan sold: %w", err)
		}
		st.RevenueUSD = round2(st.RevenueUSD)
		sold[st.ProductID] = &st
	}
	rows.Close()

	// COGS per product (FIFO batch cost) in the window.
	var periodCOGS float64
	cogsRows, err := s.pool.Query(ctx, `
		SELECT m.product_id, COALESCE(SUM(-m.quantity_change * b.unit_cost), 0)
		FROM stock_movements m
		JOIN batches b ON b.id = m.batch_id
		JOIN invoices i ON m.reference_type = 'invoice' AND i.id = m.reference_id
		WHERE i.branch_id = $1 AND i.status <> 'voided' AND m.reason = 'invoice_issued'
		  AND COALESCE(i.issued_at, i.created_at) >= NOW() - make_interval(days => $2)
		GROUP BY m.product_id`, branchID, windowDays)
	if err != nil {
		return nil, fmt.Errorf("cogs per product: %w", err)
	}
	for cogsRows.Next() {
		var pid int64
		var cogs float64
		if err := cogsRows.Scan(&pid, &cogs); err != nil {
			cogsRows.Close()
			return nil, fmt.Errorf("scan cogs: %w", err)
		}
		periodCOGS += cogs
		if st, ok := sold[pid]; ok {
			st.ProfitUSD = round2(st.RevenueUSD - cogs)
		}
	}
	cogsRows.Close()

	// Top sellers by revenue.
	for _, st := range sold {
		resp.TopSellers = append(resp.TopSellers, *st)
	}
	sort.Slice(resp.TopSellers, func(i, j int) bool { return resp.TopSellers[i].RevenueUSD > resp.TopSellers[j].RevenueUSD })
	if len(resp.TopSellers) > 10 {
		resp.TopSellers = resp.TopSellers[:10]
	}
	if resp.TopSellers == nil {
		resp.TopSellers = []dto.ProductStat{}
	}

	if resp.Valuation.CostValue > 0 {
		resp.TurnoverRatio = round2(periodCOGS / resp.Valuation.CostValue)
	}

	// Dead stock: on hand but no sale within the window.
	dsRows, err := s.pool.Query(ctx, `
		SELECT p.id, p.name, p.stock_quantity, round(p.buy_price * p.stock_quantity, 2),
		       MAX(COALESCE(i.issued_at, i.created_at))::text
		FROM products p
		LEFT JOIN invoice_items ii ON ii.product_id = p.id
		LEFT JOIN invoices i ON i.id = ii.invoice_id AND i.status <> 'voided'
		WHERE p.branch_id = $1 AND p.is_active AND p.stock_quantity > 0
		GROUP BY p.id, p.name, p.stock_quantity, p.buy_price
		HAVING MAX(COALESCE(i.issued_at, i.created_at)) IS NULL
		    OR MAX(COALESCE(i.issued_at, i.created_at)) < NOW() - make_interval(days => $2)
		ORDER BY 4 DESC
		LIMIT 15`, branchID, windowDays)
	if err != nil {
		return nil, fmt.Errorf("dead stock: %w", err)
	}
	for dsRows.Next() {
		var d dto.DeadStockItem
		if err := dsRows.Scan(&d.ProductID, &d.Name, &d.StockQty, &d.CostValue, &d.LastSold); err != nil {
			dsRows.Close()
			return nil, fmt.Errorf("scan dead stock: %w", err)
		}
		resp.DeadStock = append(resp.DeadStock, d)
	}
	dsRows.Close()
	if resp.DeadStock == nil {
		resp.DeadStock = []dto.DeadStockItem{}
	}

	// Reorder suggestions: low stock or projected to run out within 21 days.
	prRows, err := s.pool.Query(ctx, `
		SELECT id, name, stock_quantity, COALESCE(min_stock_alert, 0)
		FROM products WHERE branch_id = $1 AND is_active`, branchID)
	if err != nil {
		return nil, fmt.Errorf("reorder products: %w", err)
	}
	for prRows.Next() {
		var pid int64
		var name string
		var stock, minStock int
		if err := prRows.Scan(&pid, &name, &stock, &minStock); err != nil {
			prRows.Close()
			return nil, fmt.Errorf("scan reorder: %w", err)
		}
		rate := 0.0
		if st, ok := sold[pid]; ok {
			rate = float64(st.QtySold) / float64(windowDays)
		}
		item := dto.ReorderItem{ProductID: pid, Name: name, StockQty: stock, MinStock: minStock, DailyRate: round2(rate)}
		var daysLeft *float64
		if rate > 0 {
			dl := round2(float64(stock) / rate)
			daysLeft = &dl
		}
		item.DaysLeft = daysLeft
		belowMin := minStock > 0 && stock < minStock
		runningOut := daysLeft != nil && *daysLeft <= 21
		if belowMin || runningOut {
			resp.Reorder = append(resp.Reorder, item)
		}
	}
	prRows.Close()
	sort.Slice(resp.Reorder, func(i, j int) bool {
		return reorderUrgency(resp.Reorder[i]) < reorderUrgency(resp.Reorder[j])
	})
	if len(resp.Reorder) > 15 {
		resp.Reorder = resp.Reorder[:15]
	}
	if resp.Reorder == nil {
		resp.Reorder = []dto.ReorderItem{}
	}

	return resp, nil
}

// Lower is more urgent: soonest to run out, then lowest stock.
func reorderUrgency(r dto.ReorderItem) float64 {
	if r.DaysLeft != nil {
		return *r.DaysLeft
	}
	return 1000 + float64(r.StockQty)
}

// ---------------------------------------------------------------------------
// Customer analytics
// ---------------------------------------------------------------------------

func (s *Service) GetCustomers(ctx context.Context, branchID int64, windowDays int) (*dto.CustomersResponse, error) {
	if windowDays <= 0 {
		windowDays = 90
	}
	resp := &dto.CustomersResponse{WindowDays: windowDays}

	rows, err := s.pool.Query(ctx, `
		SELECT i.customer_id, c.name, COUNT(*), COALESCE(SUM(i.total_usd), 0),
		       MAX(COALESCE(i.issued_at, i.created_at))::text,
		       (CURRENT_DATE - MAX(COALESCE(i.issued_at, i.created_at))::date)
		FROM invoices i
		JOIN customers c ON c.id = i.customer_id
		WHERE i.branch_id = $1 AND i.status <> 'voided' AND i.customer_id IS NOT NULL
		GROUP BY i.customer_id, c.name
		ORDER BY 4 DESC`, branchID)
	if err != nil {
		return nil, fmt.Errorf("customers: %w", err)
	}
	defer rows.Close()

	var all []dto.CustomerStat
	var totalSpent float64
	var totalVisits int
	for rows.Next() {
		var st dto.CustomerStat
		var last string
		var since int
		if err := rows.Scan(&st.CustomerID, &st.Name, &st.InvoiceCount, &st.TotalSpent, &last, &since); err != nil {
			return nil, fmt.Errorf("scan customer: %w", err)
		}
		st.TotalSpent = round2(st.TotalSpent)
		st.LastVisit = &last
		st.DaysSinceVisit = &since
		totalSpent += st.TotalSpent
		totalVisits += st.InvoiceCount
		if st.InvoiceCount > 1 {
			resp.RepeatCustomers++
		}
		if since > windowDays {
			resp.ChurnRiskCount++
			resp.ChurnRisk = append(resp.ChurnRisk, st)
		}
		all = append(all, st)
	}

	resp.TotalCustomers = len(all)
	if totalVisits > 0 {
		resp.AvgSpendPerVisit = round2(totalSpent / float64(totalVisits))
	}

	// New customers: first purchase within the window.
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM (
			SELECT customer_id, MIN(COALESCE(issued_at, created_at)) AS first_at
			FROM invoices WHERE branch_id = $1 AND status <> 'voided' AND customer_id IS NOT NULL
			GROUP BY customer_id
		) t WHERE t.first_at >= NOW() - make_interval(days => $2)`,
		branchID, windowDays).Scan(&resp.NewCustomers); err != nil {
		return nil, fmt.Errorf("new customers: %w", err)
	}

	if len(all) > 10 {
		resp.TopCustomers = all[:10]
	} else {
		resp.TopCustomers = all
	}
	if resp.TopCustomers == nil {
		resp.TopCustomers = []dto.CustomerStat{}
	}
	sort.Slice(resp.ChurnRisk, func(i, j int) bool {
		return *resp.ChurnRisk[i].DaysSinceVisit > *resp.ChurnRisk[j].DaysSinceVisit
	})
	if len(resp.ChurnRisk) > 10 {
		resp.ChurnRisk = resp.ChurnRisk[:10]
	}
	if resp.ChurnRisk == nil {
		resp.ChurnRisk = []dto.CustomerStat{}
	}
	return resp, nil
}

// GetTechnicians reports each technician's completed jobs, the revenue from
// those jobs' invoices, logged hours, and what they cost to pay (salary
// prorated for the period and/or commission on the labor they billed) over a
// date range.
func (s *Service) GetTechnicians(ctx context.Context, branchID int64, from, to string) (*dto.TechniciansResponse, error) {
	resp := &dto.TechniciansResponse{From: from, To: to, Technicians: []dto.TechStat{}}
	days := daysInRange(from, to)

	rows, err := s.pool.Query(ctx, `
		SELECT e.id, e.name, e.pay_type, e.base_salary, e.hourly_rate, e.commission_rate,
		       COUNT(sj.id),
		       COALESCE(SUM(CASE WHEN i.id IS NOT NULL THEN i.total_usd ELSE 0 END), 0),
		       COALESCE(SUM(sj.actual_hours), 0),
		       COALESCE(SUM(li.labor_amt), 0)
		FROM employees e
		JOIN service_jobs sj ON sj.assigned_to = e.id AND sj.status = 'completed'
		  AND sj.completed_at::date BETWEEN $2::date AND $3::date
		LEFT JOIN invoices i ON i.id = sj.invoice_id AND i.status <> 'voided'
		LEFT JOIN LATERAL (
		    SELECT COALESCE(SUM(ii.total_usd), 0) AS labor_amt
		    FROM invoice_items ii WHERE ii.invoice_id = i.id AND ii.item_type = 'labor'
		) li ON true
		WHERE e.branch_id = $1
		GROUP BY e.id, e.name, e.pay_type, e.base_salary, e.hourly_rate, e.commission_rate
		ORDER BY 8 DESC`, branchID, from, to)
	if err != nil {
		return nil, fmt.Errorf("technicians: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var t dto.TechStat
		var payType string
		var baseSalary, hourlyRate, commissionRate, laborRevenue float64
		if err := rows.Scan(&t.EmployeeID, &t.Name, &payType, &baseSalary, &hourlyRate, &commissionRate,
			&t.JobsCompleted, &t.Revenue, &t.Hours, &laborRevenue); err != nil {
			return nil, fmt.Errorf("scan tech: %w", err)
		}
		t.Revenue = round2(t.Revenue)
		t.SalaryCost, t.CommissionCost = computePayCost(payType, baseSalary, hourlyRate, commissionRate, laborRevenue, t.Hours, days)
		t.PayrollCost = round2(t.SalaryCost + t.CommissionCost)
		resp.Technicians = append(resp.Technicians, t)
	}
	return resp, nil
}

// computePayCost applies one employee's pay formula for the period:
// salary is the monthly base_salary annualized then prorated by days in
// range; hourly is hourly_rate times hours actually logged on completed jobs
// in range; commission is commission_rate% of the labor revenue they billed
// in range; hybrid combines prorated salary with commission.
func computePayCost(payType string, baseSalary, hourlyRate, commissionRate, laborRevenue, hours float64, days int) (salaryCost, commissionCost float64) {
	switch payType {
	case "salary":
		salaryCost = baseSalary * 12.0 / 365.0 * float64(days)
	case "hourly":
		salaryCost = hourlyRate * hours
	case "commission":
		commissionCost = commissionRate / 100 * laborRevenue
	case "hybrid":
		salaryCost = baseSalary * 12.0 / 365.0 * float64(days)
		commissionCost = commissionRate / 100 * laborRevenue
	}
	return round2(salaryCost), round2(commissionCost)
}

func daysInRange(from, to string) int {
	f, err1 := time.Parse("2006-01-02", from)
	t, err2 := time.Parse("2006-01-02", to)
	if err1 != nil || err2 != nil {
		return 30
	}
	days := int(t.Sub(f).Hours()/24) + 1
	if days < 1 {
		return 1
	}
	return days
}

// ---------------------------------------------------------------------------
// Profit & Loss (with operating expenses)
// ---------------------------------------------------------------------------

func (s *Service) GetPnL(ctx context.Context, branchID int64, from, to string) (*dto.PnLResponse, error) {
	resp := &dto.PnLResponse{From: from, To: to}

	var subtotal, discount float64
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(subtotal), 0), COALESCE(SUM(discount), 0)
		FROM invoices
		WHERE branch_id = $1 AND status <> 'voided'
		  AND issued_at::date BETWEEN $2::date AND $3::date`,
		branchID, from, to).Scan(&subtotal, &discount); err != nil {
		return nil, fmt.Errorf("pnl revenue: %w", err)
	}
	resp.Revenue = round2(subtotal - discount)

	var cogs float64
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(-m.quantity_change * b.unit_cost), 0)
		FROM stock_movements m
		JOIN batches b ON b.id = m.batch_id
		JOIN invoices i ON m.reference_type = 'invoice' AND i.id = m.reference_id
		WHERE i.branch_id = $1 AND i.status <> 'voided' AND m.reason = 'invoice_issued'
		  AND i.issued_at::date BETWEEN $2::date AND $3::date`,
		branchID, from, to).Scan(&cogs); err != nil {
		return nil, fmt.Errorf("pnl cogs: %w", err)
	}
	resp.COGS = round2(cogs)

	// Refunds reduce what the shop actually kept.
	var returns float64
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(refund_amount), 0) FROM returns
		WHERE branch_id = $1 AND created_at::date BETWEEN $2::date AND $3::date`,
		branchID, from, to).Scan(&returns); err != nil {
		return nil, fmt.Errorf("pnl returns: %w", err)
	}
	resp.Returns = round2(returns)
	resp.GrossProfit = round2(resp.Revenue - resp.Returns - resp.COGS)

	payroll, payrollTotal, err := s.payrollForRange(ctx, branchID, from, to)
	if err != nil {
		return nil, err
	}
	resp.PayrollBreakdown = payroll
	resp.Payroll = payrollTotal

	rows, err := s.pool.Query(ctx, `
		SELECT category, COALESCE(SUM(amount_usd), 0)
		FROM expenses
		WHERE branch_id = $1 AND spent_at BETWEEN $2::date AND $3::date
		GROUP BY category ORDER BY 2 DESC`, branchID, from, to)
	if err != nil {
		return nil, fmt.Errorf("pnl expenses: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var ec dto.PnLExpenseCategory
		if err := rows.Scan(&ec.Category, &ec.Amount); err != nil {
			return nil, fmt.Errorf("scan expense cat: %w", err)
		}
		ec.Amount = round2(ec.Amount)
		resp.Expenses += ec.Amount
		resp.ExpenseCategories = append(resp.ExpenseCategories, ec)
	}
	if resp.ExpenseCategories == nil {
		resp.ExpenseCategories = []dto.PnLExpenseCategory{}
	}
	resp.Expenses = round2(resp.Expenses)
	resp.NetProfit = round2(resp.GrossProfit - resp.Expenses - resp.Payroll)
	if resp.Revenue > 0 {
		resp.GrossMarginPct = round2(resp.GrossProfit / resp.Revenue * 100)
		resp.NetMarginPct = round2(resp.NetProfit / resp.Revenue * 100)
	}
	return resp, nil
}

// payrollForRange costs out every active employee for the period — unlike
// GetTechnicians (which only lists employees who completed a job in range),
// this covers everyone on payroll, since a salaried employee is owed pay
// whether or not they were assigned any jobs.
func (s *Service) payrollForRange(ctx context.Context, branchID int64, from, to string) ([]dto.PnLPayrollLine, float64, error) {
	days := daysInRange(from, to)

	rows, err := s.pool.Query(ctx, `
		SELECT e.id, e.name, e.pay_type, e.base_salary, e.hourly_rate, e.commission_rate,
		       COALESCE(SUM(sj.actual_hours), 0),
		       COALESCE(SUM(li.labor_amt), 0)
		FROM employees e
		LEFT JOIN service_jobs sj ON sj.assigned_to = e.id AND sj.status = 'completed'
		  AND sj.completed_at::date BETWEEN $2::date AND $3::date
		LEFT JOIN invoices i ON i.id = sj.invoice_id AND i.status <> 'voided'
		LEFT JOIN LATERAL (
		    SELECT COALESCE(SUM(ii.total_usd), 0) AS labor_amt
		    FROM invoice_items ii WHERE ii.invoice_id = i.id AND ii.item_type = 'labor'
		) li ON true
		WHERE e.branch_id = $1 AND e.is_active = true
		GROUP BY e.id, e.name, e.pay_type, e.base_salary, e.hourly_rate, e.commission_rate
		ORDER BY e.name`, branchID, from, to)
	if err != nil {
		return nil, 0, fmt.Errorf("payroll: %w", err)
	}
	defer rows.Close()

	var breakdown []dto.PnLPayrollLine
	var total float64
	for rows.Next() {
		var line dto.PnLPayrollLine
		var baseSalary, hourlyRate, commissionRate, hours, laborRevenue float64
		if err := rows.Scan(&line.EmployeeID, &line.Name, &line.PayType, &baseSalary, &hourlyRate, &commissionRate,
			&hours, &laborRevenue); err != nil {
			return nil, 0, fmt.Errorf("scan payroll: %w", err)
		}
		line.SalaryCost, line.CommissionCost = computePayCost(line.PayType, baseSalary, hourlyRate, commissionRate, laborRevenue, hours, days)
		line.Total = round2(line.SalaryCost + line.CommissionCost)
		if line.Total == 0 {
			continue
		}
		total += line.Total
		breakdown = append(breakdown, line)
	}
	if breakdown == nil {
		breakdown = []dto.PnLPayrollLine{}
	}
	return breakdown, round2(total), nil
}
