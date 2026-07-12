package dto

import "time"

type CreateStocktakeRequest struct {
	Notes string `json:"notes,omitempty"`
	// Optional: pre-populate the sheet with every active product matching
	// these filters (either can be omitted; both empty = don't pre-populate).
	Type     string `json:"type,omitempty"`
	Category string `json:"category,omitempty"`
}

type AddItemRequest struct {
	ProductID int64 `json:"product_id" binding:"required"`
}

type SetCountRequest struct {
	// A pointer so an explicit 0 (product is completely out) can be told
	// apart from the field being omitted — Gin's "required" on a plain int
	// rejects the zero value, which would silently break counting out-of-stock items.
	CountedQty *float64 `json:"counted_qty" binding:"required,gte=0"`
}

type StocktakeListResponse struct {
	ID            int64      `json:"id"`
	Status        string     `json:"status"`
	Notes         string     `json:"notes,omitempty"`
	ItemCount     int        `json:"item_count"`
	CountedCount  int        `json:"counted_count"`
	VarianceCount int        `json:"variance_count"`
	CreatedByID   *int64     `json:"created_by_id,omitempty"`
	CreatedByName string     `json:"created_by_name,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

type StocktakeItemResponse struct {
	ID            int64      `json:"id"`
	ProductID     int64      `json:"product_id"`
	SKU           string     `json:"sku"`
	Barcode       string     `json:"barcode,omitempty"`
	ProductName   string     `json:"product_name"`
	ExpectedQty   float64    `json:"expected_qty"`
	CountedQty    *float64   `json:"counted_qty,omitempty"`
	Variance      *float64   `json:"variance,omitempty"`
	CountedByName string     `json:"counted_by_name,omitempty"`
	CountedAt     *time.Time `json:"counted_at,omitempty"`
}

type StocktakeDetailResponse struct {
	StocktakeListResponse
	Items []StocktakeItemResponse `json:"items"`
}

type FinalizeItemResult struct {
	ProductID int64  `json:"product_id"`
	SKU       string `json:"sku"`
	Message   string `json:"message"`
}

type FinalizeResult struct {
	Adjusted  int                  `json:"adjusted"`
	Unchanged int                  `json:"unchanged"`
	Skipped   int                  `json:"skipped"`
	Errors    []FinalizeItemResult `json:"errors,omitempty"`
}
