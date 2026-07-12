package models

import "time"

type ServiceJob struct {
	ID             int64      `json:"id"`
	BranchID       int64      `json:"branch_id"`
	CustomerID     *int64     `json:"customer_id,omitempty"`
	VehicleID      *int64     `json:"vehicle_id,omitempty"`
	InvoiceID      *int64     `json:"invoice_id,omitempty"`
	JobNumber      string     `json:"job_number"`
	Status         string     `json:"status"`
	Priority       string     `json:"priority"`
	Description    string     `json:"description"`
	Diagnosis      string     `json:"diagnosis,omitempty"`
	WorkPerformed  string     `json:"work_performed,omitempty"`
	EstimatedHours float64    `json:"estimated_hours,omitempty"`
	ActualHours    float64    `json:"actual_hours,omitempty"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	Notes          string     `json:"notes,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	CreatedByID    *int64     `json:"created_by_id,omitempty"`
}

type ServiceJobItem struct {
	ID           int64     `json:"id"`
	ServiceJobID int64     `json:"service_job_id"`
	ProductID    int64     `json:"product_id"`
	Description  string    `json:"description,omitempty"`
	Quantity     float64   `json:"quantity"`
	UnitPrice    float64   `json:"unit_price"`
	TotalPrice   float64   `json:"total_price"`
	CreatedAt    time.Time `json:"created_at"`
}
