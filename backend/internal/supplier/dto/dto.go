package dto

type CreateSupplierRequest struct {
	Name    string `json:"name" binding:"required"`
	Phone   string `json:"phone,omitempty"`
	Email   string `json:"email,omitempty"`
	Address string `json:"address,omitempty"`
	Notes   string `json:"notes,omitempty"`
}

type UpdateSupplierRequest struct {
	Name    *string `json:"name,omitempty"`
	Phone   *string `json:"phone,omitempty"`
	Email   *string `json:"email,omitempty"`
	Address *string `json:"address,omitempty"`
	Notes   *string `json:"notes,omitempty"`
}

type SupplierResponse struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	Phone          string  `json:"phone,omitempty"`
	Email          string  `json:"email,omitempty"`
	Address        string  `json:"address,omitempty"`
	Notes          string  `json:"notes,omitempty"`
	IsActive       bool    `json:"is_active"`
	TotalPurchased float64 `json:"total_purchased"`
	Outstanding    float64 `json:"outstanding"`
	PurchaseCount  int     `json:"purchase_count"`
	CreatedAt      string  `json:"created_at"`
}

type PurchaseItem struct {
	BatchID       int64   `json:"batch_id"`
	ProductID     *int64  `json:"product_id,omitempty"`
	ProductName   string  `json:"product_name"`
	Quantity      int     `json:"quantity"`
	UnitCost      float64 `json:"unit_cost"`
	TotalCost     float64 `json:"total_cost"`
	AmountPaid    float64 `json:"amount_paid"`
	Owed          float64 `json:"owed"`
	DOTCode       string  `json:"dot_code,omitempty"`
	InvoiceNumber string  `json:"invoice_number,omitempty"`
	InvoiceImage  string  `json:"invoice_image,omitempty"`
	ReceivedAt    string  `json:"received_at"`
}

// PayRequest selects the exact batches (receives) to pay off in full.
type PayRequest struct {
	BatchIDs []int64 `json:"batch_ids" binding:"required,min=1"`
}
