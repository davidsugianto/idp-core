package permission

import (
	"context"
	"errors"

	permissionModel "github.com/davidsugianto/idp-core/internal/model/permission"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Create creates a new permission
func (r *repository) Create(ctx context.Context, permission *permissionModel.Permission) error {
	if permission.ID == "" {
		permission.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).Create(permission).Error
}

// GetByID retrieves a permission by ID
func (r *repository) GetByID(ctx context.Context, id string) (*permissionModel.Permission, error) {
	var permission permissionModel.Permission
	err := r.db.WithContext(ctx).First(&permission, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &permission, nil
}

// GetByName retrieves a permission by name
func (r *repository) GetByName(ctx context.Context, name string) (*permissionModel.Permission, error) {
	var permission permissionModel.Permission
	err := r.db.WithContext(ctx).
		Where("name = ?", name).
		First(&permission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &permission, nil
}

// GetByResourceAction retrieves a permission by resource and action
func (r *repository) GetByResourceAction(ctx context.Context, resource, action string) (*permissionModel.Permission, error) {
	var permission permissionModel.Permission
	err := r.db.WithContext(ctx).
		Where("resource = ? AND action = ?", resource, action).
		First(&permission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &permission, nil
}

// List retrieves a paginated list of permissions
func (r *repository) List(ctx context.Context, limit, offset int) ([]permissionModel.Permission, int64, error) {
	var permissions []permissionModel.Permission
	var total int64

	db := r.db.WithContext(ctx).Model(&permissionModel.Permission{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Order("resource, action").
		Limit(limit).
		Offset(offset).
		Find(&permissions).Error

	return permissions, total, err
}

// Update updates a permission
func (r *repository) Update(ctx context.Context, permission *permissionModel.Permission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

// Delete deletes a permission
func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&permissionModel.Permission{}).Error
}
