package dto

type ReturnItemInput struct {
	InvoiceItemID int64   `json:"invoice_item_id" binding:"required"`
	Quantity      float64 `json:"quantity" binding:"required,gt=0"`
}

type CreateReturnRequest struct {
	InvoiceID    int64             `json:"invoice_id" binding:"required"`
	RefundMethod string            `json:"refund_method" binding:"required,oneof=cash store_credit"`
	Reason       string            `json:"reason,omitempty"`
	Items        []ReturnItemInput `json:"items" binding:"required,min=1"`
}

type ReturnItemResp struct {
	InvoiceItemID int64   `json:"invoice_item_id"`
	ProductID     *int64  `json:"product_id,omitempty"`
	Description   string  `json:"description"`
	Quantity      float64 `json:"quantity"`
	UnitPrice     float64 `json:"unit_price"`
	Total         float64 `json:"total"`
}

type ReturnResponse struct {
	ID            int64            `json:"id"`
	InvoiceID     int64            `json:"invoice_id"`
	RefundAmount  float64          `json:"refund_amount"`
	RefundMethod  string           `json:"refund_method"`
	Reason        string           `json:"reason,omitempty"`
	CreatedByName string           `json:"created_by_name,omitempty"`
	CreatedAt     string           `json:"created_at"`
	Items         []ReturnItemResp `json:"items"`
}

// InvoiceReturns bundles a invoice's credit notes plus, per invoice_item, how
// much has already been returned (so the UI can cap remaining quantities).
type InvoiceReturns struct {
	Returns        []ReturnResponse    `json:"returns"`
	ReturnedByItem map[string]float64  `json:"returned_by_item"`
}
