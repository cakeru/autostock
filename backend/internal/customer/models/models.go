package models

import "time"

type Customer struct {
	ID           int64     `json:"id"`
	BranchID     int64     `json:"branch_id"`
	Name         string    `json:"name"`
	Phone        string    `json:"phone,omitempty"`
	Email        string    `json:"email,omitempty"`
	Address      string    `json:"address,omitempty"`
	Notes        string    `json:"notes,omitempty"`
	CustomerSince string   `json:"customer_since,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Vehicle struct {
	ID          int64     `json:"id"`
	CustomerID  int64     `json:"customer_id"`
	PlateNumber string    `json:"plate_number"`
	Make        string    `json:"make,omitempty"`
	Model       string    `json:"model,omitempty"`
	Year        int       `json:"year,omitempty"`
	VIN         string    `json:"vin,omitempty"`
	Color       string    `json:"color,omitempty"`
	BodyType    string    `json:"body_type,omitempty"`
	Notes       string    `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
