package models

import "time"

type Product struct {
	ID            int64     `json:"id"`
	BranchID      int64     `json:"branch_id"`
	Type          string    `json:"type"`
	SKU           string    `json:"sku"`
	Barcode       string    `json:"barcode,omitempty"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	Category      string    `json:"category,omitempty"`
	BuyPrice      float64   `json:"buy_price"`
	SellPrice     float64   `json:"sell_price"`
	StockQuantity float64   `json:"stock_quantity"`
	ReservedQuantity float64 `json:"reserved_quantity"`
	MinStockAlert float64   `json:"min_stock_alert"`
	Unit          string    `json:"unit"`
	IsOilProduct  bool      `json:"is_oil_product,omitempty"`
	IsBulk        bool      `json:"is_bulk,omitempty"`
	LifeKm        *int      `json:"life_km,omitempty"`
	LifeDays      *int      `json:"life_days,omitempty"`
	LifeMonths    *int      `json:"life_months,omitempty"`
	TireSize      string    `json:"tire_size,omitempty"`
	TireBrand     string    `json:"tire_brand,omitempty"`
	TireModel     string    `json:"tire_model,omitempty"`
	TirePattern   string    `json:"tire_pattern,omitempty"`
	DOTCode       string    `json:"dot_code,omitempty"`
	LoadIndex     string    `json:"load_index,omitempty"`
	SpeedRating   string    `json:"speed_rating,omitempty"`
	TireType      string    `json:"tire_type,omitempty"`
	Location      string    `json:"location,omitempty"`
	ImageURL      string    `json:"image_url,omitempty"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type StockMovement struct {
	ID             int64     `json:"id"`
	QuantityChange float64   `json:"quantity_change"`
	Reason         string    `json:"reason"`
	ReferenceType  string    `json:"reference_type,omitempty"`
	ReferenceID    int64     `json:"reference_id,omitempty"`
	InvoiceNumber  string    `json:"invoice_number,omitempty"`
	BatchNo        string    `json:"batch_no,omitempty"`
	RecordedByName string    `json:"recorded_by_name,omitempty"`
	BalanceAfter   float64   `json:"balance_after"`
	CreatedAt      time.Time `json:"created_at"`
}

type Batch struct {
	ID                int64     `json:"id"`
	BatchNo           string    `json:"batch_no"`
	Supplier          string    `json:"supplier,omitempty"`
	DOTCode           string    `json:"dot_code,omitempty"`
	UnitCost          float64   `json:"unit_cost"`
	QuantityReceived  float64   `json:"quantity_received"`
	QuantityRemaining float64   `json:"quantity_remaining"`
	Notes             string    `json:"notes,omitempty"`
	ReceivedByName    string    `json:"received_by_name,omitempty"`
	ReceivedAt        time.Time `json:"received_at"`
}

type BatchConsumer struct {
	InvoiceID     int64     `json:"invoice_id"`
	InvoiceNumber string    `json:"invoice_number"`
	CustomerName  string    `json:"customer_name"`
	Quantity      float64   `json:"quantity"`
	CreatedAt     time.Time `json:"created_at"`
}

type LowStockProduct struct {
	ID               int64   `json:"id"`
	SKU              string  `json:"sku"`
	Name             string  `json:"name"`
	StockQuantity    float64 `json:"stock_quantity"`
	ReservedQuantity float64 `json:"reserved_quantity"`
	MinStockAlert    float64 `json:"min_stock_alert"`
	SellPrice        float64 `json:"sell_price"`
}
