package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/adithyvisnu/user-go/internal/models",
	"github.com/adithyvisnu/user-go/internal/repository"
)

// UserService holds the business logic for user operations.
type UserService struct {
	repo *repository.UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// GetUser retrieves a single user by ID, scoped to the tenant.
func (s *UserService) GetUser(ctx context.Context, tenantID, userID uuid.UUID) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, tenantID, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrUserNotFound
	}
	return user, err
}

// ListUsers returns a paginated list of users for a tenant.
func (s *UserService) ListUsers(ctx context.Context, tenantID uuid.UUID, page, limit int) ([]models.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.List(ctx, tenantID, page, limit)
}

// CreateUser validates and creates a new user.
func (s *UserService) CreateUser(ctx context.Context, tenantID uuid.UUID, req *models.CreateUserRequest) (*models.User, error) {
	role := req.Role
	if role == "" {
		role = "member"
	}

	user := &models.User{
		TenantID: tenantID,
		Email:    req.Email,
		Name:     req.Name,
		Role:     role,
		Active:   true,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		// Postgres unique violation on email
		if isUniqueViolation(err) {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

// UpdateUser applies partial updates to an existing user.
func (s *UserService) UpdateUser(ctx context.Context, tenantID, userID uuid.UUID, req *models.UpdateUserRequest) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, tenantID, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	changes := make(map[string]any)
	if req.Name != nil {
		changes["name"] = *req.Name
	}
	if req.Role != nil {
		changes["role"] = *req.Role
	}
	if req.Active != nil {
		changes["active"] = *req.Active
	}

	if len(changes) == 0 {
		return user, nil // nothing to do
	}

	if err := s.repo.Update(ctx, user, changes); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	// Re-fetch to return the updated record
	return s.repo.GetByID(ctx, tenantID, userID)
}

// DeleteUser soft-deletes a user.
func (s *UserService) DeleteUser(ctx context.Context, tenantID, userID uuid.UUID) error {
	err := s.repo.Delete(ctx, tenantID, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrUserNotFound
	}
	return err
}

// ─── Sentinel errors ──────────────────────────────────────────────────────────

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailTaken   = errors.New("email already in use")
)

// isUniqueViolation checks for Postgres error code 23505.
func isUniqueViolation(err error) bool {
	return err != nil && containsCode(err.Error(), "23505")
}

func containsCode(msg, code string) bool {
	for i := 0; i+len(code) <= len(msg); i++ {
		if msg[i:i+len(code)] == code {
			return true
		}
	}
	return false
}
