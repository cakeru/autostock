package models

import "time"

type Setting struct {
	ID          int64     `json:"id"`
	BranchID    *int64    `json:"branch_id,omitempty"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
