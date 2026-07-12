package dto

type SettingsResponse struct {
	ExchangeRateUSDKHR float64 `json:"exchange_rate_usd_khr"`
	TaxRatePercent     float64 `json:"tax_rate_percent"`
	TaxEnabled         bool    `json:"tax_enabled"`
	InvoicePrefix      string  `json:"invoice_prefix"`
	LowStockThreshold  int     `json:"low_stock_threshold"`
	SalePackages       string  `json:"sale_packages,omitempty"`   // raw JSON, owner-editable POS packages
	LaborPresets       string  `json:"labor_presets,omitempty"`   // raw JSON, owner-editable labor buttons
	FeePresets         string  `json:"fee_presets,omitempty"`     // raw JSON, owner-editable fee buttons
	PaymentMethods     string  `json:"payment_methods,omitempty"` // raw JSON array of method names
	ShopName           string  `json:"shop_name"`                 // shown on printed invoices/receipts
	ShopAddress        string  `json:"shop_address,omitempty"`
	ShopPhone          string  `json:"shop_phone,omitempty"`
	ShopEmail          string  `json:"shop_email,omitempty"`
	FeatureBatchScan   bool    `json:"feature_batch_scan"` // opt-in batch-QR scan-to-install feature
}

type UpdateSettingRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

type ExchangeRateResponse struct {
	Rate      float64 `json:"rate"`
	UpdatedAt string  `json:"updated_at,omitempty"`
}
