package dto

type AuditFilter struct {
	Action     string `form:"action"`
	EntityType string `form:"entity_type"`
	UserID     int64  `form:"user_id"`
	From       string `form:"from"`
	To         string `form:"to"`
	Page       int    `form:"page"`
	PerPage    int    `form:"per_page"`
}

type AuditLogItem struct {
	ID         int64  `json:"id"`
	Action     string `json:"action"`
	EntityType string `json:"entity_type"`
	EntityID   *int64 `json:"entity_id,omitempty"`
	UserID     *int64 `json:"user_id,omitempty"`
	UserName   string `json:"user_name"`
	CreatedAt  string `json:"created_at"`
}
