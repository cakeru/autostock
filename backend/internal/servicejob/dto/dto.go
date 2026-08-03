package dto

import "time"

type ServiceJobFilter struct {
	Status        string `form:"status"`
	CustomerID    int64  `form:"customer_id"`
	VehicleID     int64  `form:"vehicle_id"`
	AssignedTo    int64  `form:"assigned_to"`
	CreatedAtGTE  string `form:"created_at_gte"`
	CreatedAtLTE  string `form:"created_at_lte"`
	ScheduledFrom string `form:"scheduled_from"` // agenda: only jobs booked at/after this, ordered by time
	Page          int    `form:"page"`
	PerPage       int    `form:"per_page"`
}

type CreateServiceJobRequest struct {
	CustomerID     *int64         `json:"customer_id,omitempty"`
	VehicleID      *int64         `json:"vehicle_id,omitempty"`
	Mileage        *int           `json:"mileage,omitempty"`
	Description    string         `json:"description" binding:"required"`
	Priority       string         `json:"priority" binding:"omitempty,oneof=low normal high urgent"`
	EstimatedHours float64        `json:"estimated_hours,omitempty"`
	ScheduledAt    string         `json:"scheduled_at,omitempty"` // RFC3339 / ISO datetime; empty = walk-in
	AssignedTo     *int64         `json:"assigned_to,omitempty"`
	Discount       float64        `json:"discount,omitempty"` // agreed whole-sale discount, carried to the invoice
	Notes          string         `json:"notes,omitempty"`
	Items          []JobItemInput `json:"items,omitempty"`
}

// JobItemInput is a line added when creating a job (e.g. from a saved POS cart).
type JobItemInput struct {
	ProductID        *int64  `json:"product_id,omitempty"`
	ItemType         string  `json:"item_type,omitempty"` // product | labor | fee | custom
	Description      string  `json:"description,omitempty"`
	Quantity         float64 `json:"quantity" binding:"required,gt=0"`
	UnitPrice        float64 `json:"unit_price" binding:"gte=0"`
	VehicleEventType *string `json:"vehicle_event_type,omitempty" binding:"omitempty,oneof=service"`
}

type UpdateServiceJobRequest struct {
	Status         *string  `json:"status,omitempty" binding:"omitempty,oneof=pending in_progress completed cancelled"`
	Priority       *string  `json:"priority,omitempty" binding:"omitempty,oneof=low normal high urgent"`
	Diagnosis      *string  `json:"diagnosis,omitempty"`
	WorkPerformed  *string  `json:"work_performed,omitempty"`
	EstimatedHours *float64 `json:"estimated_hours,omitempty"`
	ActualHours    *float64 `json:"actual_hours,omitempty"`
	StartedAt      *string  `json:"started_at,omitempty"`
	CompletedAt    *string  `json:"completed_at,omitempty"`
	ScheduledAt    *string  `json:"scheduled_at,omitempty"`
	AssignedTo     *int64   `json:"assigned_to,omitempty"`
	Mileage        *int     `json:"mileage,omitempty"`
	Notes          *string  `json:"notes,omitempty"`
}

type AddItemRequest struct {
	ProductID        *int64  `json:"product_id,omitempty"`
	ItemType         string  `json:"item_type,omitempty"` // product | labor | fee | custom
	Description      string  `json:"description,omitempty"`
	Quantity         float64 `json:"quantity" binding:"required,gt=0"`
	UnitPrice        float64 `json:"unit_price" binding:"gte=0"`
	VehicleEventType *string `json:"vehicle_event_type,omitempty" binding:"omitempty,oneof=service"`
}

type ServiceJobListResponse struct {
	ID              int64      `json:"id"`
	BranchID        int64      `json:"branch_id"`
	JobNumber       string     `json:"job_number"`
	Status          string     `json:"status"`
	Priority        string     `json:"priority"`
	CustomerID      *int64     `json:"customer_id,omitempty"`
	CustomerName    string     `json:"customer_name,omitempty"`
	CustomerPhone   string     `json:"customer_phone,omitempty"`
	VehicleID       *int64     `json:"vehicle_id,omitempty"`
	PlateNumber     string     `json:"plate_number,omitempty"`
	VehicleInfo     string     `json:"vehicle_info,omitempty"`
	MileageUnit     string     `json:"mileage_unit"`
	Description     string     `json:"description"`
	InvoiceID       *int64     `json:"invoice_id,omitempty"`
	ScheduledAt     *time.Time `json:"scheduled_at,omitempty"`
	AssignedToID    *int64     `json:"assigned_to_id,omitempty"`
	AssignedToName  string     `json:"assigned_to_name,omitempty"`
	QuoteApprovedAt *time.Time `json:"quote_approved_at,omitempty"`
	QuoteApprovedBy *int64     `json:"quote_approved_by,omitempty"`
	CreatedByID     *int64     `json:"created_by_id,omitempty"`
	CreatedByName   string     `json:"created_by_name,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type ServiceJobDetailResponse struct {
	ID              int64                     `json:"id"`
	BranchID        int64                     `json:"branch_id"`
	JobNumber       string                    `json:"job_number"`
	Status          string                    `json:"status"`
	Priority        string                    `json:"priority"`
	CustomerID      *int64                    `json:"customer_id,omitempty"`
	CustomerName    string                    `json:"customer_name,omitempty"`
	CustomerPhone   string                    `json:"customer_phone,omitempty"`
	VehicleID       *int64                    `json:"vehicle_id,omitempty"`
	PlateNumber     string                    `json:"plate_number,omitempty"`
	VehicleInfo     string                    `json:"vehicle_info,omitempty"`
	Mileage         *int                      `json:"mileage,omitempty"`
	MileageUnit     string                    `json:"mileage_unit"`
	Description     string                    `json:"description"`
	Diagnosis       string                    `json:"diagnosis,omitempty"`
	WorkPerformed   string                    `json:"work_performed,omitempty"`
	EstimatedHours  *float64                  `json:"estimated_hours,omitempty"`
	ActualHours     *float64                  `json:"actual_hours,omitempty"`
	StartedAt       *time.Time                `json:"started_at,omitempty"`
	CompletedAt     *time.Time                `json:"completed_at,omitempty"`
	ScheduledAt     *time.Time                `json:"scheduled_at,omitempty"`
	AssignedToID    *int64                    `json:"assigned_to_id,omitempty"`
	AssignedToName  string                    `json:"assigned_to_name,omitempty"`
	InvoiceID       *int64                    `json:"invoice_id,omitempty"`
	QuoteApprovedAt *time.Time                `json:"quote_approved_at,omitempty"`
	QuoteApprovedBy *int64                    `json:"quote_approved_by,omitempty"`
	Discount        float64                   `json:"discount,omitempty"`
	Notes           string                    `json:"notes,omitempty"`
	Items           []ServiceJobItemResponse  `json:"items"`
	TotalAmount     float64                   `json:"total_amount"`
	CreatedByID     *int64                    `json:"created_by_id,omitempty"`
	CreatedByName   string                    `json:"created_by_name,omitempty"`
	CreatedAt       time.Time                 `json:"created_at"`
	UpdatedAt       time.Time                 `json:"updated_at"`
}

type ServiceJobItemResponse struct {
	ID          int64   `json:"id"`
	ProductID   *int64  `json:"product_id,omitempty"`
	ItemType    string  `json:"item_type"`
	ProductName string  `json:"product_name,omitempty"`
	Description string  `json:"description,omitempty"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
}
