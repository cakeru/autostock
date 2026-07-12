package dto

// ---------- Sales ----------

type SalesPoint struct {
	Period       string  `json:"period"`
	RevenueUSD   float64 `json:"revenue_usd"`
	InvoiceCount int     `json:"invoice_count"`
}

type PaymentMethodTotal struct {
	Method string  `json:"method"`
	Count  int     `json:"count"`
	Total  float64 `json:"total"`
}

type SalesResponse struct {
	From             string               `json:"from"`
	To               string               `json:"to"`
	Granularity      string               `json:"granularity"`
	RevenueUSD       float64              `json:"revenue_usd"`
	InvoiceCount     int                  `json:"invoice_count"`
	AvgTicket        float64              `json:"avg_ticket"`
	TaxCollected     float64              `json:"tax_collected"`
	PrevRevenueUSD   float64              `json:"prev_revenue_usd"`
	RevenueChangePct float64              `json:"revenue_change_pct"`
	Series           []SalesPoint         `json:"series"`
	PaymentMethods   []PaymentMethodTotal `json:"payment_methods"`
}

// ---------- Receivables ----------

type ARBucket struct {
	Label          string  `json:"label"`
	Count          int     `json:"count"`
	OutstandingUSD float64 `json:"outstanding_usd"`
}

type ARCustomer struct {
	CustomerID     *int64  `json:"customer_id"`
	CustomerName   string  `json:"customer_name"`
	InvoiceCount   int     `json:"invoice_count"`
	OutstandingUSD float64 `json:"outstanding_usd"`
	OldestDays     int     `json:"oldest_days"`
}

type ReceivablesResponse struct {
	TotalOutstanding float64      `json:"total_outstanding"`
	Buckets          []ARBucket   `json:"buckets"`
	Customers        []ARCustomer `json:"customers"`
}

// ---------- Inventory ----------

type InventoryValuation struct {
	CostValue       float64 `json:"cost_value"`
	RetailValue     float64 `json:"retail_value"`
	PotentialProfit float64 `json:"potential_profit"`
	SKUCount        int     `json:"sku_count"`
	UnitsOnHand     int     `json:"units_on_hand"`
}

type ProductStat struct {
	ProductID  int64   `json:"product_id"`
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	QtySold    int     `json:"qty_sold"`
	RevenueUSD float64 `json:"revenue_usd"`
	ProfitUSD  float64 `json:"profit_usd"`
	StockQty   int     `json:"stock_qty"`
}

type DeadStockItem struct {
	ProductID int64   `json:"product_id"`
	Name      string  `json:"name"`
	StockQty  int     `json:"stock_qty"`
	CostValue float64 `json:"cost_value"`
	LastSold  *string `json:"last_sold,omitempty"`
}

type ReorderItem struct {
	ProductID int64    `json:"product_id"`
	Name      string   `json:"name"`
	StockQty  int      `json:"stock_qty"`
	MinStock  int      `json:"min_stock"`
	DailyRate float64  `json:"daily_rate"`
	DaysLeft  *float64 `json:"days_left,omitempty"`
}

type InventoryResponse struct {
	WindowDays    int                `json:"window_days"`
	Valuation     InventoryValuation `json:"valuation"`
	TurnoverRatio float64            `json:"turnover_ratio"`
	TopSellers    []ProductStat      `json:"top_sellers"`
	DeadStock     []DeadStockItem    `json:"dead_stock"`
	Reorder       []ReorderItem      `json:"reorder"`
}

// ---------- Customers ----------

type CustomerStat struct {
	CustomerID     int64   `json:"customer_id"`
	Name           string  `json:"name"`
	InvoiceCount   int     `json:"invoice_count"`
	TotalSpent     float64 `json:"total_spent"`
	LastVisit      *string `json:"last_visit,omitempty"`
	DaysSinceVisit *int    `json:"days_since_visit,omitempty"`
}

type CustomersResponse struct {
	WindowDays       int            `json:"window_days"`
	TotalCustomers   int            `json:"total_customers"`
	NewCustomers     int            `json:"new_customers"`
	RepeatCustomers  int            `json:"repeat_customers"`
	ChurnRiskCount   int            `json:"churn_risk_count"`
	AvgSpendPerVisit float64        `json:"avg_spend_per_visit"`
	TopCustomers     []CustomerStat `json:"top_customers"`
	ChurnRisk        []CustomerStat `json:"churn_risk"`
}

// ---------- P&L ----------

// ---------- Technicians ----------

type TechStat struct {
	EmployeeID     int64   `json:"employee_id"`
	Name           string  `json:"name"`
	JobsCompleted  int     `json:"jobs_completed"`
	Revenue        float64 `json:"revenue"`
	Hours          float64 `json:"hours"`
	SalaryCost     float64 `json:"salary_cost"`
	CommissionCost float64 `json:"commission_cost"`
	PayrollCost    float64 `json:"payroll_cost"`
}

type TechniciansResponse struct {
	From        string     `json:"from"`
	To          string     `json:"to"`
	Technicians []TechStat `json:"technicians"`
}

type PnLExpenseCategory struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
}

type PnLPayrollLine struct {
	EmployeeID     int64   `json:"employee_id"`
	Name           string  `json:"name"`
	PayType        string  `json:"pay_type"`
	SalaryCost     float64 `json:"salary_cost"`
	CommissionCost float64 `json:"commission_cost"`
	Total          float64 `json:"total"`
}

type PnLResponse struct {
	From              string               `json:"from"`
	To                string               `json:"to"`
	Revenue           float64              `json:"revenue"`
	Returns           float64              `json:"returns"`
	COGS              float64              `json:"cogs"`
	GrossProfit       float64              `json:"gross_profit"`
	Payroll           float64              `json:"payroll"`
	PayrollBreakdown  []PnLPayrollLine     `json:"payroll_breakdown"`
	Expenses          float64              `json:"expenses"`
	NetProfit         float64              `json:"net_profit"`
	GrossMarginPct    float64              `json:"gross_margin_pct"`
	NetMarginPct      float64              `json:"net_margin_pct"`
	ExpenseCategories []PnLExpenseCategory `json:"expense_categories"`
}
