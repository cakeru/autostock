package domain

import "time"

type Branch struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	Email     string    `json:"email,omitempty"`
	LogoURL   string    `json:"logo_url,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID           int64     `json:"id"`
	BranchID     int64     `json:"branch_id"`
	Username     string    `json:"username"`
	Email        string    `json:"email,omitempty"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name,omitempty"`
	Role         string    `json:"role"`
	Permissions  []string  `json:"permissions"`
	IsActive     bool      `json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
