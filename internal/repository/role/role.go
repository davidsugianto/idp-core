package role

import (
	"context"
	"errors"

	roleModel "github.com/davidsugianto/idp-core/internal/model/role"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Create creates a new role
func (r *repository) Create(ctx context.Context, role *roleModel.Role) error {
	if role.ID == "" {
		role.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).Create(role).Error
}

// GetByID retrieves a role by ID
func (r *repository) GetByID(ctx context.Context, id string) (*roleModel.Role, error) {
	var role roleModel.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		First(&role, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

// GetByName retrieves a role by name
func (r *repository) GetByName(ctx context.Context, name string) (*roleModel.Role, error) {
	var role roleModel.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		Where("name = ?", name).
		First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

// List retrieves a paginated list of roles
func (r *repository) List(ctx context.Context, limit, offset int) ([]roleModel.Role, int64, error) {
	var roles []roleModel.Role
	var total int64

	db := r.db.WithContext(ctx).Model(&roleModel.Role{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Preload("Permissions").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&roles).Error

	return roles, total, err
}

// Update updates a role
func (r *repository) Update(ctx context.Context, role *roleModel.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

// SoftDelete soft deletes a role
func (r *repository) SoftDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&roleModel.Role{}).Error
}
