package dto

type CreateDepositRequest struct {
	CustomerID int64   `json:"customer_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Note       string  `json:"note,omitempty"`
}

type ApplyDepositRequest struct {
	InvoiceID int64 `json:"invoice_id" binding:"required"`
}

type DepositResponse struct {
	ID            int64   `json:"id"`
	CustomerID    int64   `json:"customer_id"`
	CustomerName  string  `json:"customer_name,omitempty"`
	Amount        float64 `json:"amount"`
	Note          string  `json:"note,omitempty"`
	Status        string  `json:"status"`
	InvoiceID     *int64  `json:"invoice_id,omitempty"`
	InvoiceNumber string  `json:"invoice_number,omitempty"`
	CreatedAt     string  `json:"created_at"`
	SettledAt     *string `json:"settled_at,omitempty"`
}
