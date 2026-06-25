package model

import "time"

// User represents a platform user (DESIGN_DOC 30.1).
type User struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"` // never serialize
	DisplayName  string     `json:"display_name"`
	Phone        string     `json:"phone,omitempty"`
	Email        string     `json:"email,omitempty"`
	Role         string     `json:"role"`   // super_admin / admin / mobile_user
	Status       string     `json:"status"` // active / disabled
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// UserSession represents a refresh token session (DESIGN_DOC 30.1).
type UserSession struct {
	ID               string     `json:"id"`
	UserID           string     `json:"user_id"`
	RefreshTokenHash string     `json:"-"`           // never serialize
	ClientType       string     `json:"client_type"` // admin / mobile
	ExpiresAt        time.Time  `json:"expires_at"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

// AuditLog records security-sensitive operations (DESIGN_DOC 27.3, 30.1).
type AuditLog struct {
	ID              string    `json:"id"`
	OperatorUserID  string    `json:"operator_user_id"`
	Action          string    `json:"action"`
	TargetType      string    `json:"target_type"`
	TargetID        string    `json:"target_id"`
	PayloadSnapshot string    `json:"payload_snapshot,omitempty"` // JSONB
	CreatedAt       time.Time `json:"created_at"`
}

// LoginRequest is the request body for POST /api/v1/auth/login.
type LoginRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	ClientType string `json:"client_type" binding:"required"` // admin / mobile
}

// RegisterRequest is the request body for POST /api/v1/auth/register (Admin).
type RegisterRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=64"`
	Password    string `json:"password" binding:"required,min=6,max=128"`
	DisplayName string `json:"display_name" binding:"required"`
	Phone       string `json:"phone,omitempty"`
	Email       string `json:"email,omitempty"`
	InviteCode  string `json:"invite_code,omitempty"`
}

// RefreshRequest is the request body for POST /api/v1/auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LogoutRequest is the request body for POST /api/v1/auth/logout.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthResponse is the response for login/register.
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	User         *User  `json:"user"`
}

// CreateUserRequest is the admin request for creating a user (DESIGN_DOC 29.2).
type CreateUserRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=64"`
	DisplayName string `json:"display_name" binding:"required"`
	Password    string `json:"password" binding:"required,min=6,max=128"`
	Role        string `json:"role" binding:"required,oneof=super_admin admin mobile_user"`
	Phone       string `json:"phone,omitempty"`
	Email       string `json:"email,omitempty"`
}

// UpdateUserRequest is the admin request for updating a user.
type UpdateUserRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Phone       *string `json:"phone,omitempty"`
	Email       *string `json:"email,omitempty"`
	Role        *string `json:"role,omitempty" binding:"omitempty,oneof=super_admin admin mobile_user"`
	Status      *string `json:"status,omitempty" binding:"omitempty,oneof=active disabled"`
}

// UpdateUserStatusRequest is the request to enable/disable a user.
type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active disabled"`
}

// ResetPasswordRequest is the admin request to reset a user's password.
type ResetPasswordRequest struct {
	Password string `json:"password" binding:"required,min=6,max=128"`
}

// UserListResponse wraps paginated user list.
type UserListResponse struct {
	Data       []User `json:"data"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	TotalCount int64  `json:"total_count"`
}
