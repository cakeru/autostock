package dto

import "time"

type InvoiceFilter struct {
	Status        string `form:"status"`
	PaymentStatus string `form:"payment_status"`
	CustomerID    int64  `form:"customer_id"`
	InvoiceNumber string `form:"invoice_number"`
	CreatedAtGTE  string `form:"created_at_gte"`
	CreatedAtLTE  string `form:"created_at_lte"`
	Page          int    `form:"page"`
	PerPage       int    `form:"per_page"`
}

type CreateInvoiceRequest struct {
	CustomerID    *int64              `json:"customer_id,omitempty"`
	VehicleID     *int64              `json:"vehicle_id,omitempty"`
	ServiceJobID  *int64              `json:"service_job_id,omitempty"`
	Mileage       *int                `json:"mileage,omitempty"`
	Items         []InvoiceItemReq    `json:"items" binding:"required,min=1,dive"`
	Discount      float64             `json:"discount"`
	ExchangeRate  float64   `json:"exchange_rate"`
	TaxRate       float64             `json:"tax_rate"`
	PaymentMethod string              `json:"payment_method"`
	Notes         string              `json:"notes"`
}

type InvoiceItemReq struct {
	ProductID    *int64  `json:"product_id,omitempty"`
	ItemType     string  `json:"item_type" binding:"required,oneof=product labor custom fee"`
	Description  string  `json:"description"`
	Quantity     float64 `json:"quantity" binding:"required,gt=0"`
	UnitPriceUSD float64 `json:"unit_price_usd" binding:"required,gte=0"`
}

type UpdateInvoiceItemRequest struct {
	Description  *string  `json:"description,omitempty"`
	Quantity     *float64 `json:"quantity,omitempty" binding:"omitempty,gt=0"`
	UnitPriceUSD *float64 `json:"unit_price_usd,omitempty" binding:"omitempty,gte=0"`
}

type UpdateInvoiceRequest struct {
	PaymentMethod *string `json:"payment_method,omitempty"`
	PaymentNotes  *string `json:"payment_notes,omitempty"`
	Notes         *string `json:"notes,omitempty"`
	VehicleID     *int64  `json:"vehicle_id,omitempty"`
	Mileage       *int    `json:"mileage,omitempty"`
	ClearVehicle  bool    `json:"clear_vehicle,omitempty"`
}

type RecordPaymentRequest struct {
	Amount         float64 `json:"amount" binding:"required,gt=0"` // base USD applied to the invoice
	Method         string  `json:"method" binding:"required"`
	Notes          string  `json:"notes,omitempty"`
	Currency       string  `json:"currency,omitempty"`        // USD | KHR (how it was tendered)
	TenderedAmount float64 `json:"tendered_amount,omitempty"` // amount in the tendered currency
	ExchangeRate   float64 `json:"exchange_rate,omitempty"`   // KHR per USD used
	Reference      string  `json:"reference,omitempty"`       // bank/wallet transfer Trx ID (e.g. ABA)
}

type PaymentResponse struct {
	ID             int64     `json:"id"`
	InvoiceID      int64     `json:"invoice_id"`
	Amount         float64   `json:"amount"`
	Method         string    `json:"method"`
	Currency       string    `json:"currency,omitempty"`
	TenderedAmount *float64  `json:"tendered_amount,omitempty"`
	ReceivedByName string    `json:"received_by_name,omitempty"`
	Notes          string    `json:"notes,omitempty"`
	Reference      string    `json:"reference,omitempty"`
	ProofURL       string    `json:"proof_url,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type VoidInvoiceRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type InvoiceListResponse struct {
	ID            int64      `json:"id"`
	InvoiceNumber string     `json:"invoice_number"`
	Status        string     `json:"status"`
	PaymentStatus string     `json:"payment_status"`
	CustomerID    *int64     `json:"customer_id,omitempty"`
	CustomerName  string     `json:"customer_name,omitempty"`
	VehicleID     *int64     `json:"vehicle_id,omitempty"`
	PlateNumber   string     `json:"plate_number,omitempty"`
	ServiceJobID  *int64     `json:"service_job_id,omitempty"`
	JobNumber     string     `json:"job_number,omitempty"`
	Subtotal      float64    `json:"subtotal"`
	TotalUSD      float64    `json:"total_usd"`
	ExchangeRate  float64    `json:"exchange_rate"`
	TotalKHR      float64    `json:"total_khr"`
	PaidAmount    float64    `json:"paid_amount"`
	IssuedAt      *time.Time `json:"issued_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	CreatedByID   *int64     `json:"created_by_id,omitempty"`
	CreatedByName string     `json:"created_by_name,omitempty"`
}

type InvoiceDetailResponse struct {
	ID             int64             `json:"id"`
	InvoiceNumber  string            `json:"invoice_number"`
	Status         string            `json:"status"`
	PaymentStatus  string            `json:"payment_status"`
	CustomerID     *int64            `json:"customer_id,omitempty"`
	CustomerName   string            `json:"customer_name,omitempty"`
	CustomerPhone  string            `json:"customer_phone,omitempty"`
	CustomerAddr   string            `json:"customer_address,omitempty"`
	VehicleID      *int64            `json:"vehicle_id,omitempty"`
	PlateNumber    string            `json:"plate_number,omitempty"`
	VehicleInfo    string            `json:"vehicle_info,omitempty"`
	Mileage        *int              `json:"mileage,omitempty"`
	ServiceJobID   *int64            `json:"service_job_id,omitempty"`
	JobNumber      string            `json:"job_number,omitempty"`
	Items          []InvoiceItemResp `json:"items"`
	Payments       []PaymentResponse `json:"payments"`
	Subtotal       float64           `json:"subtotal"`
	TaxRate        float64           `json:"tax_rate"`
	TaxAmount      float64           `json:"tax_amount"`
	Discount       float64           `json:"discount"`
	TotalUSD       float64           `json:"total_usd"`
	ExchangeRate   float64           `json:"exchange_rate"`
	TotalKHR       float64           `json:"total_khr"`
	PaidAmount     float64           `json:"paid_amount"`
	PaymentMethod  string            `json:"payment_method,omitempty"`
	PaymentNotes   string            `json:"payment_notes,omitempty"`
	Notes          string            `json:"notes,omitempty"`
	IssuedAt       *time.Time        `json:"issued_at,omitempty"`
	VoidedAt       *time.Time        `json:"voided_at,omitempty"`
	VoidReason     string            `json:"void_reason,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	CreatedByID    *int64            `json:"created_by_id,omitempty"`
	CreatedByName  string            `json:"created_by_name,omitempty"`
}

type InvoiceItemResp struct {
	ID           int64   `json:"id"`
	ProductID    *int64  `json:"product_id,omitempty"`
	ItemType     string  `json:"item_type"`
	Description  string  `json:"description"`
	Quantity     float64 `json:"quantity"`
	UnitPriceUSD float64 `json:"unit_price_usd"`
	TotalUSD     float64 `json:"total_usd"`
}

type CreateInvoiceFromJobRequest struct {
	Discount      float64 `json:"discount"`
	ExchangeRate  float64 `json:"exchange_rate" binding:"required"`
	PaymentMethod string  `json:"payment_method"`
	Notes         string  `json:"notes"`
}
