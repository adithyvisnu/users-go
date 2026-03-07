package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system.
type User struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"        json:"id"`
	TenantID  uuid.UUID      `gorm:"type:uuid;not null;index"    json:"tenant_id"`
	Email     string         `gorm:"uniqueIndex;not null"        json:"email"`
	Name      string         `gorm:"not null"                    json:"name"`
	Role      string         `gorm:"default:'member'"            json:"role"` // admin | member
	Active    bool           `gorm:"default:true"                json:"active"`
	CreatedAt time.Time      `                                   json:"created_at"`
	UpdatedAt time.Time      `                                   json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                       json:"-"`
}

// BeforeCreate sets a UUID primary key automatically.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// ─── Request / Response DTOs ──────────────────────────────────────────────────

// CreateUserRequest is the payload for POST /users.
// @Description Payload to create a new user
type CreateUserRequest struct {
	Email string `json:"email" binding:"required,email"    example:"alice@example.com"`
	Name  string `json:"name"  binding:"required,min=2"    example:"Alice Smith"`
	Role  string `json:"role"  binding:"omitempty,oneof=admin member" example:"member"`
}

// UpdateUserRequest is the payload for PATCH /users/:id.
// @Description Payload to update an existing user
type UpdateUserRequest struct {
	Name   *string `json:"name"   binding:"omitempty,min=2" example:"Alice Johnson"`
	Role   *string `json:"role"   binding:"omitempty,oneof=admin member" example:"admin"`
	Active *bool   `json:"active" binding:"omitempty"       example:"true"`
}

// UserResponse is the public-facing representation of a user.
// @Description User object returned by the API
type UserResponse struct {
	ID        uuid.UUID `json:"id"         example:"550e8400-e29b-41d4-a716-446655440000"`
	TenantID  uuid.UUID `json:"tenant_id"  example:"660e8400-e29b-41d4-a716-446655440001"`
	Email     string    `json:"email"      example:"alice@example.com"`
	Name      string    `json:"name"       example:"Alice Smith"`
	Role      string    `json:"role"       example:"member"`
	Active    bool      `json:"active"     example:"true"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-15T10:00:00Z"`
}

// ListUsersResponse wraps a paginated list of users.
// @Description Paginated list of users
type ListUsersResponse struct {
	Data  []UserResponse `json:"data"`
	Total int64          `json:"total"   example:"42"`
	Page  int            `json:"page"    example:"1"`
	Limit int            `json:"limit"   example:"20"`
}

// ErrorResponse is the standard error envelope.
// @Description Standard error response
type ErrorResponse struct {
	Error   string `json:"error"             example:"user not found"`
	Code    string `json:"code,omitempty"    example:"NOT_FOUND"`
	Details string `json:"details,omitempty" example:"no user with id ..."`
}

// ToResponse converts a User model to a UserResponse DTO.
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		TenantID:  u.TenantID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		Active:    u.Active,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
