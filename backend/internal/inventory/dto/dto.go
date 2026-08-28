package dto

type ProductFilter struct {
	Type            string `form:"type"`
	NameLike        string `form:"name_like"`
	TireSize        string `form:"tire_size"`
	TireBrand       string `form:"tire_brand"`
	StockQuantityLT float64 `form:"stock_quantity_lt"`
	Category        string `form:"category"`
	Code            string `form:"code"` // exact barcode or SKU (POS scan)
	Page            int    `form:"page"`
	PerPage         int    `form:"per_page"`
}

type CreateProductRequest struct {
	Type          string  `json:"type" binding:"required,oneof=tire part labor consumable"`
	SKU           string  `json:"sku" binding:"required"`
	Barcode       string  `json:"barcode,omitempty"`
	Name          string  `json:"name" binding:"required"`
	Description   string  `json:"description,omitempty"`
	Category      string  `json:"category,omitempty"`
	BuyPrice      float64 `json:"buy_price"`
	SellPrice     float64 `json:"sell_price"`
	StockQuantity float64 `json:"stock_quantity"`
	MinStockAlert float64 `json:"min_stock_alert"`
	Unit          string  `json:"unit"`
	IsOilProduct  bool    `json:"is_oil_product,omitempty"`
	IsBulk        bool    `json:"is_bulk,omitempty"`
	LifeKm        *int    `json:"life_km,omitempty"`
	LifeDays      *int    `json:"life_days,omitempty"`
	LifeMonths    *int    `json:"life_months,omitempty"`
	TireSize      string  `json:"tire_size,omitempty"`
	TireBrand     string  `json:"tire_brand,omitempty"`
	TireModel     string  `json:"tire_model,omitempty"`
	TirePattern   string  `json:"tire_pattern,omitempty"`
	DOTCode       string  `json:"dot_code,omitempty"`
	LoadIndex     string  `json:"load_index,omitempty"`
	SpeedRating   string  `json:"speed_rating,omitempty"`
	TireType      string  `json:"tire_type,omitempty"`
	Location      string  `json:"location,omitempty"`
}

type UpdateProductRequest struct {
	Type          *string  `json:"type,omitempty" binding:"omitempty,oneof=tire part labor consumable"`
	Barcode       *string  `json:"barcode,omitempty"`
	Name          *string  `json:"name,omitempty"`
	Description   *string  `json:"description,omitempty"`
	Category      *string  `json:"category,omitempty"`
	BuyPrice      *float64 `json:"buy_price,omitempty"`
	SellPrice     *float64 `json:"sell_price,omitempty"`
	MinStockAlert *float64 `json:"min_stock_alert,omitempty"`
	Unit          *string  `json:"unit,omitempty"`
	IsOilProduct  *bool    `json:"is_oil_product,omitempty"`
	IsBulk        *bool    `json:"is_bulk,omitempty"`
	LifeKm        *int     `json:"life_km,omitempty"`
	LifeDays      *int     `json:"life_days,omitempty"`
	LifeMonths    *int     `json:"life_months,omitempty"`
	TireSize      *string  `json:"tire_size,omitempty"`
	TireBrand     *string  `json:"tire_brand,omitempty"`
	TireModel     *string  `json:"tire_model,omitempty"`
	TirePattern   *string  `json:"tire_pattern,omitempty"`
	DOTCode       *string  `json:"dot_code,omitempty"`
	LoadIndex     *string  `json:"load_index,omitempty"`
	SpeedRating   *string  `json:"speed_rating,omitempty"`
	TireType      *string  `json:"tire_type,omitempty"`
	Location      *string  `json:"location,omitempty"`
}

// ReceiveInvoice is one supplier invoice attached to a receive. A receive can
// carry several invoices (e.g. a $100 purchase split into four $25 invoices)
// that are paid off one at a time.
type ReceiveInvoice struct {
	InvoiceNumber string  `json:"invoice_number,omitempty"`
	InvoiceImage  string  `json:"invoice_image,omitempty"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
}

type ReceiveStockRequest struct {
	Quantity   float64 `json:"quantity" binding:"required,gt=0"`
	UnitCost   float64 `json:"unit_cost,omitempty"`
	SupplierID *int64  `json:"supplier_id,omitempty"`
	Paid       bool    `json:"paid,omitempty"` // paid to supplier on delivery
	Supplier   string  `json:"supplier,omitempty"`
	DOTCode    string  `json:"dot_code,omitempty"`
	Notes      string  `json:"notes,omitempty"`
	Invoices   []ReceiveInvoice `json:"invoices,omitempty" binding:"dive"`
}

type AdjustStockRequest struct {
	QuantityChange float64 `json:"quantity_change" binding:"required"`
	Reason         string `json:"reason" binding:"required"`
}
