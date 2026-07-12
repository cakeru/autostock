package models

import "time"

type Invoice struct {
	ID             int64      `json:"id"`
	BranchID       int64      `json:"branch_id"`
	InvoiceNumber  string     `json:"invoice_number"`
	CustomerID     *int64     `json:"customer_id,omitempty"`
	VehicleID      *int64     `json:"vehicle_id,omitempty"`
	ServiceJobID   *int64     `json:"service_job_id,omitempty"`
	Subtotal       float64    `json:"subtotal"`
	TaxRate        float64    `json:"tax_rate"`
	TaxAmount      float64    `json:"tax_amount"`
	Discount       float64    `json:"discount"`
	TotalUSD       float64    `json:"total_usd"`
	ExchangeRate   float64    `json:"exchange_rate"`
	TotalKHR       float64    `json:"total_khr"`
	PaymentStatus  string     `json:"payment_status"`
	Status         string     `json:"status"`
	PaidAmount     float64    `json:"paid_amount"`
	PaymentMethod  string     `json:"payment_method,omitempty"`
	Notes          string     `json:"notes,omitempty"`
	VoidedAt       *time.Time `json:"voided_at,omitempty"`
	VoidReason     string     `json:"void_reason,omitempty"`
	IssuedAt       *time.Time `json:"issued_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	CreatedByID    *int64     `json:"created_by_id,omitempty"`
}

type InvoiceItem struct {
	ID           int64     `json:"id"`
	InvoiceID    int64     `json:"invoice_id"`
	ProductID    *int64    `json:"product_id,omitempty"`
	ItemType     string    `json:"item_type"`
	Description  string    `json:"description"`
	Quantity     float64   `json:"quantity"`
	UnitPriceUSD float64   `json:"unit_price_usd"`
	TotalUSD     float64   `json:"total_usd"`
	CreatedAt    time.Time `json:"created_at"`
}
