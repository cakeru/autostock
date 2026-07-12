package dto

type ExpenseResponse struct {
	ID          int64   `json:"id"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	AmountUSD   float64 `json:"amount_usd"`
	SpentAt     string  `json:"spent_at"`
	CreatedBy   *int64  `json:"created_by,omitempty"`
	CreatedName string  `json:"created_by_name,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

type CreateExpenseRequest struct {
	Category    string  `json:"category" binding:"required"`
	Description string  `json:"description"`
	AmountUSD   float64 `json:"amount_usd" binding:"required,gt=0"`
	SpentAt     string  `json:"spent_at"` // YYYY-MM-DD; defaults to today when empty
}
