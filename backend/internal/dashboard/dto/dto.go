package dto

type SummaryResponse struct {
	TodayRevenue   float64         `json:"today_revenue"`
	TodayJobs      int             `json:"today_jobs"`
	TotalCustomers int             `json:"total_customers"`
	LowStockCount  int             `json:"low_stock_count"`
	UnpaidCount    int             `json:"unpaid_count"`
	OutstandingUSD float64         `json:"outstanding_usd"`
	RecentInvoices []RecentInvoice `json:"recent_invoices"`
	RecentJobs     []RecentJob     `json:"recent_jobs"`
}

type RecentInvoice struct {
	ID            int64   `json:"id"`
	InvoiceNumber string  `json:"invoice_number"`
	CustomerName  string  `json:"customer_name"`
	TotalUSD      float64 `json:"total_usd"`
	Status        string  `json:"status"`
}

type RecentJob struct {
	ID           int64  `json:"id"`
	JobNumber    string `json:"job_number"`
	CustomerName string `json:"customer_name"`
	Status       string `json:"status"`
}

type DailyRevenueItem struct {
	Date          string  `json:"date"`
	RevenueUSD    float64 `json:"revenue_usd"`
	InvoiceCount  int     `json:"invoice_count"`
}

type PaymentMethodTotal struct {
	Method string  `json:"method"`
	Count  int     `json:"count"`
	Total  float64 `json:"total"`
}

type UserPaymentTotal struct {
	UserID   int64   `json:"user_id"`
	UserName string  `json:"user_name"`
	Count    int     `json:"count"`
	Total    float64 `json:"total"`
}

type ProfitCategory struct {
	Category  string  `json:"category"` // "Parts & tires" | "Labor" | "Fees"
	Revenue   float64 `json:"revenue"`
	Cost      float64 `json:"cost"`
	Profit    float64 `json:"profit"`
	MarginPct float64 `json:"margin_pct"`
}

type ProfitResponse struct {
	From         string           `json:"from"`
	To           string           `json:"to"`
	GrossRevenue float64          `json:"gross_revenue"`
	Discounts    float64          `json:"discounts"`
	Revenue      float64          `json:"revenue"` // net of discounts
	Cost         float64          `json:"cost"`
	GrossProfit  float64          `json:"gross_profit"`
	MarginPct    float64          `json:"margin_pct"`
	InvoiceCount int              `json:"invoice_count"`
	Categories   []ProfitCategory `json:"categories"`
}

type DayCloseResponse struct {
	Date            string              `json:"date"`
	PaymentMethods  []PaymentMethodTotal `json:"payment_methods"`
	VoidedCount     int                 `json:"voided_count"`
	VoidedTotal     float64             `json:"voided_total"`
	UserTotals      []UserPaymentTotal  `json:"user_totals"`
}
