package dto

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string   `json:"access_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int      `json:"expires_in"`
	User         UserInfo `json:"user"`
}

type UserInfo struct {
	ID          int64    `json:"id"`
	Username    string   `json:"username"`
	Email       string   `json:"email,omitempty"`
	FullName    string   `json:"full_name,omitempty"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	BranchID    int64    `json:"branch_id"`
	IsActive    bool     `json:"is_active"`
}

type CreateUserRequest struct {
	Username    string   `json:"username" binding:"required"`
	Password    string   `json:"password" binding:"required,min=6"`
	Email       string   `json:"email,omitempty"`
	FullName    string   `json:"full_name,omitempty"`
	Role        string   `json:"role" binding:"required,oneof=admin staff"`
	Permissions []string `json:"permissions"`
}

type UpdateUserRequest struct {
	Email       *string   `json:"email,omitempty"`
	FullName    *string   `json:"full_name,omitempty"`
	Permissions *[]string `json:"permissions,omitempty"`
	IsActive    *bool     `json:"is_active,omitempty"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}
