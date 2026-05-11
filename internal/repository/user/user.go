package user

import (
	"context"
	"errors"
	"time"

	"github.com/davidsugianto/idp-core/internal/model/user"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Create creates a new user
func (r *repository) Create(ctx context.Context, u *user.User) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	if u.Status == "" {
		u.Status = "active"
	}
	if u.Provider == "" {
		u.Provider = "local"
	}
	return r.db.WithContext(ctx).Create(u).Error
}

// GetByID retrieves a user by ID
func (r *repository) GetByID(ctx context.Context, id string) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// GetByEmail retrieves a user by email
func (r *repository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// GetByProviderID retrieves a user by provider and provider ID
func (r *repository) GetByProviderID(ctx context.Context, provider, providerID string) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).
		Where("provider = ? AND provider_id = ?", provider, providerID).
		First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// List retrieves a paginated list of users
func (r *repository) List(ctx context.Context, limit, offset int) ([]user.User, int64, error) {
	var users []user.User
	var total int64

	query := r.db.WithContext(ctx).Model(&user.User{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// ListByStatus retrieves users by status
func (r *repository) ListByStatus(ctx context.Context, status string) ([]user.User, error) {
	var users []user.User
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Find(&users).Error
	return users, err
}

// Update updates a user
func (r *repository) Update(ctx context.Context, u *user.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

// UpdateStatus updates a user's status
func (r *repository) UpdateStatus(ctx context.Context, id, status string) error {
	return r.db.WithContext(ctx).
		Model(&user.User{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// UpdateLastLogin updates the user's last login timestamp
func (r *repository) UpdateLastLogin(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&user.User{}).
		Where("id = ?", id).
		Update("last_login_at", now).Error
}

// SoftDelete soft deletes a user
func (r *repository) SoftDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&user.User{}).Error
}

// HardDelete permanently deletes a user
func (r *repository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&user.User{}).Error
}
