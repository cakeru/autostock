package dto

import "time"

type ScheduleRequest struct {
	Name          string `json:"name" binding:"required,max=60"`
	Cron          string `json:"cron" binding:"required,max=64"`
	Enabled       *bool  `json:"enabled"`
	RetentionDays int    `json:"retention_days" binding:"required,min=1,max=365"`
}

type ScheduleResponse struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	Cron          string     `json:"cron"`
	Enabled       bool       `json:"enabled"`
	RetentionDays int        `json:"retention_days"`
	LastRunAt     *time.Time `json:"last_run_at"`
	LastStatus    string     `json:"last_status"`
	LastError     string     `json:"last_error,omitempty"`
	NextRunAt     *time.Time `json:"next_run_at"`
	LatestFile    string     `json:"latest_file,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}
