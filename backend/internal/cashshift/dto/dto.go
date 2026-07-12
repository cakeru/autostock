package dto

type OpenRequest struct {
	OpeningFloat float64 `json:"opening_float" binding:"gte=0"`
}

type CloseRequest struct {
	ClosingAmount float64 `json:"closing_amount" binding:"gte=0"`
	Note          string  `json:"note,omitempty"`
}

type ShiftResponse struct {
	ID             int64    `json:"id"`
	Status         string   `json:"status"`
	OpenedByID     *int64   `json:"opened_by_id,omitempty"`
	OpenedByName   string   `json:"opened_by_name,omitempty"`
	OpeningFloat   float64  `json:"opening_float"`
	OpenedAt       string   `json:"opened_at"`
	CashSales      float64  `json:"cash_sales"`      // cash payments taken during the shift
	ExpectedAmount float64  `json:"expected_amount"` // opening_float + cash_sales
	ClosingAmount  *float64 `json:"closing_amount,omitempty"`
	OverShort      *float64 `json:"over_short,omitempty"`
	Note           string   `json:"note,omitempty"`
	ClosedByName   string   `json:"closed_by_name,omitempty"`
	ClosedAt       *string  `json:"closed_at,omitempty"`
}
