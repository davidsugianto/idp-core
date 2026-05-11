package team

import (
	"context"
	"errors"

	"github.com/davidsugianto/idp-core/internal/model/team"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Create creates a new team
func (r *repository) Create(ctx context.Context, t *team.Team) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	if t.Status == "" {
		t.Status = "active"
	}
	return r.db.WithContext(ctx).Create(t).Error
}

// GetByID retrieves a team by ID
func (r *repository) GetByID(ctx context.Context, id string) (*team.Team, error) {
	var t team.Team
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

// GetBySlug retrieves a team by slug
func (r *repository) GetBySlug(ctx context.Context, slug string) (*team.Team, error) {
	var t team.Team
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

// List retrieves a paginated list of teams
func (r *repository) List(ctx context.Context, limit, offset int) ([]team.Team, int64, error) {
	var teams []team.Team
	var total int64

	query := r.db.WithContext(ctx).Model(&team.Team{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&teams).Error; err != nil {
		return nil, 0, err
	}

	return teams, total, nil
}

// ListByStatus retrieves teams by status
func (r *repository) ListByStatus(ctx context.Context, status string) ([]team.Team, error) {
	var teams []team.Team
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Find(&teams).Error
	return teams, err
}

// Update updates a team
func (r *repository) Update(ctx context.Context, t *team.Team) error {
	return r.db.WithContext(ctx).Save(t).Error
}

// SoftDelete soft deletes a team
func (r *repository) SoftDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&team.Team{}).Error
}
