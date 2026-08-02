package dto

import "time"

type CustomerFilter struct {
	NameLike string `form:"name_like"`
	Phone    string `form:"phone"`
	Email    string `form:"email"`
	// Search matches broadly across name, phone, and the customer's vehicles
	// (plate / make / model) — so an admin who remembers the car but not the
	// owner can still find them.
	Search  string `form:"search"`
	Page    int    `form:"page"`
	PerPage int    `form:"per_page"`
}

type CreateCustomerRequest struct {
	Name         string `json:"name" binding:"required"`
	CustomerType string `json:"customer_type,omitempty" binding:"omitempty,oneof=garage retail company"`
	Phone        string `json:"phone,omitempty"`
	Email        string `json:"email,omitempty"`
	Address      string `json:"address,omitempty"`
	Notes        string `json:"notes,omitempty"`
}

type UpdateCustomerRequest struct {
	Name         *string `json:"name,omitempty"`
	CustomerType *string `json:"customer_type,omitempty" binding:"omitempty,oneof=garage retail company"`
	Phone        *string `json:"phone,omitempty"`
	Email        *string `json:"email,omitempty"`
	Address      *string `json:"address,omitempty"`
	Notes        *string `json:"notes,omitempty"`
}

type CustomerResponse struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	CustomerType  string    `json:"customer_type"`
	Phone         string    `json:"phone,omitempty"`
	Email         string    `json:"email,omitempty"`
	Address       string    `json:"address,omitempty"`
	Notes         string    `json:"notes,omitempty"`
	CustomerSince string    `json:"customer_since,omitempty"`
	IsActive      bool      `json:"is_active"`
	VehicleCount  int       `json:"vehicle_count"`
	VehiclePlates string    `json:"vehicle_plates,omitempty"` // comma-separated, for the list row
	TotalSpent    float64   `json:"total_spent"`
	LastVisit     string    `json:"last_visit,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CustomerStats are the value metrics shown at the top of a customer profile.
type CustomerStats struct {
	TotalSpent  float64 `json:"total_spent"`
	VisitCount  int     `json:"visit_count"`
	LastVisit   string  `json:"last_visit,omitempty"`
	Outstanding float64 `json:"outstanding"`
}

// ActivityItem is one entry in a customer's unified timeline (jobs + invoices).
type ActivityItem struct {
	Type        string     `json:"type"` // "job" | "invoice"
	ID          int64      `json:"id"`
	Ref         string     `json:"ref"`
	Date        *time.Time `json:"date,omitempty"`
	Title       string     `json:"title"`
	Status      string     `json:"status"`
	Amount      float64    `json:"amount"`
	Outstanding float64    `json:"outstanding"`
	Plate       string     `json:"plate,omitempty"`
}

type CreateVehicleRequest struct {
	PlateNumber string `json:"plate_number" binding:"required"`
	Make        string `json:"make,omitempty"`
	Model       string `json:"model,omitempty"`
	Year        *int   `json:"year,omitempty"`
	VIN         string `json:"vin,omitempty"`
	Color       string `json:"color,omitempty"`
	BodyType    string `json:"body_type,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

type UpdateVehicleRequest struct {
	PlateNumber *string `json:"plate_number,omitempty"`
	Make        *string `json:"make,omitempty"`
	Model       *string `json:"model,omitempty"`
	Year        *int    `json:"year,omitempty"`
	VIN         *string `json:"vin,omitempty"`
	Color       *string `json:"color,omitempty"`
	BodyType    *string `json:"body_type,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

type ServiceHistoryItem struct {
	ID            int64      `json:"id"`
	JobNumber     string     `json:"job_number"`
	Status        string     `json:"status"`
	Description   string     `json:"description,omitempty"`
	WorkPerformed string     `json:"work_performed,omitempty"`
	TotalAmount   float64    `json:"total_amount"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}
