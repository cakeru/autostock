package dto

import "time"

type CreateEmployeeRequest struct {
	Name           string  `json:"name" binding:"required"`
	Position       string  `json:"position,omitempty"`
	Phone          string  `json:"phone,omitempty"`
	Email          string  `json:"email,omitempty"`
	PayType        string  `json:"pay_type,omitempty" binding:"omitempty,oneof=salary hourly commission hybrid"`
	BaseSalary     float64 `json:"base_salary,omitempty"`
	HourlyRate     float64 `json:"hourly_rate,omitempty"`
	CommissionRate float64 `json:"commission_rate,omitempty"`
	HireDate       string  `json:"hire_date,omitempty"` // YYYY-MM-DD
	Notes          string  `json:"notes,omitempty"`
}

type UpdateEmployeeRequest struct {
	Name           *string  `json:"name,omitempty"`
	Position       *string  `json:"position,omitempty"`
	Phone          *string  `json:"phone,omitempty"`
	Email          *string  `json:"email,omitempty"`
	PayType        *string  `json:"pay_type,omitempty" binding:"omitempty,oneof=salary hourly commission hybrid"`
	BaseSalary     *float64 `json:"base_salary,omitempty"`
	HourlyRate     *float64 `json:"hourly_rate,omitempty"`
	CommissionRate *float64 `json:"commission_rate,omitempty"`
	HireDate       *string  `json:"hire_date,omitempty"`
	Notes          *string  `json:"notes,omitempty"`
}

// CreateAccountRequest gives an existing employee profile a login — the
// profile (pay info, job assignment history) stays exactly as it was.
type CreateAccountRequest struct {
	Username    string   `json:"username" binding:"required"`
	Password    string   `json:"password" binding:"required,min=6"`
	Role        string   `json:"role" binding:"required,oneof=admin staff"`
	Permissions []string `json:"permissions,omitempty"`
}

type EmployeeResponse struct {
	ID             int64     `json:"id"`
	UserID         *int64    `json:"user_id,omitempty"`
	Username       string    `json:"username,omitempty"`
	Name           string    `json:"name"`
	Position       string    `json:"position,omitempty"`
	Phone          string    `json:"phone,omitempty"`
	Email          string    `json:"email,omitempty"`
	PayType        string    `json:"pay_type"`
	BaseSalary     float64   `json:"base_salary"`
	HourlyRate     float64   `json:"hourly_rate"`
	CommissionRate float64   `json:"commission_rate"`
	HireDate       string    `json:"hire_date,omitempty"`
	Notes          string    `json:"notes,omitempty"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}
