package dto

import "time"

type CreatePORequest struct {
	SupplierID int64  `json:"supplier_id" binding:"required"`
	Notes      string `json:"notes,omitempty"`
}

type AddPOItemRequest struct {
	ProductID       int64   `json:"product_id" binding:"required"`
	QuantityOrdered int     `json:"quantity_ordered" binding:"required,gt=0"`
	UnitCost        float64 `json:"unit_cost" binding:"gte=0"`
}

// ReceiveRequest lists what actually showed up. Any line omitted receives its
// full remaining quantity_ordered - quantity_received; a line can be included
// with quantity=0 to explicitly receive nothing for it this trip. The invoice
// number/image (optional) apply to every batch created by this receive.
type ReceiveRequest struct {
	Items         []ReceiveLine `json:"items,omitempty"`
	Paid          bool          `json:"paid,omitempty"`
	InvoiceNumber string        `json:"invoice_number,omitempty"`
	InvoiceImage  string        `json:"invoice_image,omitempty"`
}

type ReceiveLine struct {
	ItemID   int64 `json:"item_id" binding:"required"`
	Quantity *int  `json:"quantity,omitempty"`
}

type POListResponse struct {
	ID            int64      `json:"id"`
	PONumber      string     `json:"po_number"`
	Status        string     `json:"status"`
	SupplierID    int64      `json:"supplier_id"`
	SupplierName  string     `json:"supplier_name"`
	Notes         string     `json:"notes,omitempty"`
	ItemCount     int        `json:"item_count"`
	TotalCost     float64    `json:"total_cost"`
	CreatedByID   *int64     `json:"created_by_id,omitempty"`
	CreatedByName string     `json:"created_by_name,omitempty"`
	OrderedAt     *time.Time `json:"ordered_at,omitempty"`
	ReceivedAt    *time.Time `json:"received_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type POItemResponse struct {
	ID               int64   `json:"id"`
	ProductID        int64   `json:"product_id"`
	SKU              string  `json:"sku"`
	ProductName      string  `json:"product_name"`
	QuantityOrdered  int     `json:"quantity_ordered"`
	QuantityReceived int     `json:"quantity_received"`
	UnitCost         float64 `json:"unit_cost"`
	TotalCost        float64 `json:"total_cost"`
}

type PODetailResponse struct {
	POListResponse
	Items []POItemResponse `json:"items"`
}

type ReceiveResult struct {
	Received int    `json:"received"` // total units received across all lines this trip
	Status   string `json:"status"`   // the PO's status after this receipt
}
