package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/adithyvisnu/user-go/internal/cache"
	"github.com/adithyvisnu/user-go/internal/models"
)

// UserRepository handles all persistence for users.
// Every read checks Redis first; writes invalidate affected keys.
type UserRepository struct {
	db    *gorm.DB
	cache *cache.Cache
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *gorm.DB, c *cache.Cache) *UserRepository {
	return &UserRepository{db: db, cache: c}
}

// ─── Read ─────────────────────────────────────────────────────────────────────

// GetByID fetches a user by ID, checking Redis before hitting PostgreSQL.
func (r *UserRepository) GetByID(ctx context.Context, tenantID, userID uuid.UUID) (*models.User, error) {
	key := cache.UserKey(tenantID.String(), userID.String())

	// 1. Cache check
	var user models.User
	hit, err := r.cache.Get(ctx, key, &user)
	if err != nil {
		// Cache errors are non-fatal — log and fall through to DB
		fmt.Printf("[cache] GET error for key %s: %v\n", key, err)
	}
	if hit {
		return &user, nil
	}

	// 2. Database fallback
	result := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", userID, tenantID).
		First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if result.Error != nil {
		return nil, fmt.Errorf("db get user: %w", result.Error)
	}

	// 3. Populate cache for next read
	if err := r.cache.Set(ctx, key, &user, cache.UserTTL); err != nil {
		fmt.Printf("[cache] SET error for key %s: %v\n", key, err)
	}

	return &user, nil
}

// List returns a paginated list of users for a tenant, cache-aside.
func (r *UserRepository) List(ctx context.Context, tenantID uuid.UUID, page, limit int) ([]models.User, int64, error) {
	key := cache.UserListKey(tenantID.String(), page, limit)

	// 1. Cache check
	type listPayload struct {
		Users []models.User `json:"users"`
		Total int64         `json:"total"`
	}
	var payload listPayload
	hit, err := r.cache.Get(ctx, key, &payload)
	if err != nil {
		fmt.Printf("[cache] GET error for key %s: %v\n", key, err)
	}
	if hit {
		return payload.Users, payload.Total, nil
	}

	// 2. Database fallback
	var users []models.User
	var total int64
	offset := (page - 1) * limit

	if err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("tenant_id = ?", tenantID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("db count users: %w", err)
	}

	if err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("db list users: %w", err)
	}

	// 3. Populate cache
	if err := r.cache.Set(ctx, key, listPayload{Users: users, Total: total}, cache.UserListTTL); err != nil {
		fmt.Printf("[cache] SET error for key %s: %v\n", key, err)
	}

	return users, total, nil
}

// ─── Write ────────────────────────────────────────────────────────────────────

// Create inserts a new user and invalidates list cache.
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("db create user: %w", err)
	}
	r.bustListCache(ctx, user.TenantID)
	return nil
}

// Update applies changes to an existing user and invalidates caches.
func (r *UserRepository) Update(ctx context.Context, user *models.User, changes map[string]any) error {
	if err := r.db.WithContext(ctx).Model(user).Updates(changes).Error; err != nil {
		return fmt.Errorf("db update user: %w", err)
	}
	r.bustUserCache(ctx, user.TenantID, user.ID)
	r.bustListCache(ctx, user.TenantID)
	return nil
}

// Delete soft-deletes a user (GORM DeletedAt) and invalidates caches.
func (r *UserRepository) Delete(ctx context.Context, tenantID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", userID, tenantID).
		Delete(&models.User{})

	if result.Error != nil {
		return fmt.Errorf("db delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	r.bustUserCache(ctx, tenantID, userID)
	r.bustListCache(ctx, tenantID)
	return nil
}

// ─── Cache invalidation helpers ───────────────────────────────────────────────

func (r *UserRepository) bustUserCache(ctx context.Context, tenantID, userID uuid.UUID) {
	key := cache.UserKey(tenantID.String(), userID.String())
	if err := r.cache.Delete(ctx, key); err != nil {
		fmt.Printf("[cache] DELETE error for key %s: %v\n", key, err)
	}
}

func (r *UserRepository) bustListCache(ctx context.Context, tenantID uuid.UUID) {
	pattern := cache.UserListPattern(tenantID.String())
	if err := r.cache.DeletePattern(ctx, pattern); err != nil {
		fmt.Printf("[cache] DELETE PATTERN error for pattern %s: %v\n", pattern, err)
	}
}

// ─── Sentinel errors ──────────────────────────────────────────────────────────

var ErrNotFound = errors.New("user not found")
