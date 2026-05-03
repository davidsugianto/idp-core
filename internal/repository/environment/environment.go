package environment

import (
	"context"
	"errors"
	"time"

	"github.com/davidsugianto/idp-core/internal/model/environment"
	"gorm.io/gorm"
)

// Create persists a new environment
func (r *repository) Create(ctx context.Context, env *environment.Environment) error {
	return r.db.WithContext(ctx).Create(env).Error
}

// GetByID retrieves an environment by ID
func (r *repository) GetByID(ctx context.Context, id string) (*environment.Environment, error) {
	var env environment.Environment
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&env).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &env, nil
}

// GetByIDAndTeam retrieves an environment by ID and team ID (for tenant isolation)
func (r *repository) GetByIDAndTeam(ctx context.Context, id, teamID string) (*environment.Environment, error) {
	var env environment.Environment
	err := r.db.WithContext(ctx).
		Where("id = ? AND team_id = ?", id, teamID).
		First(&env).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &env, nil
}

// GetByNamespace retrieves an environment by namespace
func (r *repository) GetByNamespace(ctx context.Context, namespace string) (*environment.Environment, error) {
	var env environment.Environment
	err := r.db.WithContext(ctx).Where("namespace = ?", namespace).First(&env).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &env, nil
}

// ListByTeam lists all environments for a team
func (r *repository) ListByTeam(ctx context.Context, teamID string) ([]environment.Environment, error) {
	var envs []environment.Environment
	err := r.db.WithContext(ctx).
		Where("team_id = ?", teamID).
		Order("created_at DESC").
		Find(&envs).Error
	return envs, err
}

// ListByStatus lists all environments with a specific status
func (r *repository) ListByStatus(ctx context.Context, status string) ([]environment.Environment, error) {
	var envs []environment.Environment
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Find(&envs).Error
	return envs, err
}

// ListExpired lists all environments that have expired
func (r *repository) ListExpired(ctx context.Context) ([]environment.Environment, error) {
	var envs []environment.Environment
	err := r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Where("status != ?", "deleting").
		Order("expires_at ASC").
		Find(&envs).Error
	return envs, err
}

// Update updates an environment
func (r *repository) Update(ctx context.Context, env *environment.Environment) error {
	return r.db.WithContext(ctx).Save(env).Error
}

// UpdateStatus updates the status and last error of an environment
func (r *repository) UpdateStatus(ctx context.Context, id, teamID, status, lastError string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if lastError != "" {
		updates["last_error"] = lastError
	}

	return r.db.WithContext(ctx).
		Model(&environment.Environment{}).
		Where("id = ? AND team_id = ?", id, teamID).
		Updates(updates).Error
}

// UpdateArgoAppName updates the ArgoCD application name
func (r *repository) UpdateArgoAppName(ctx context.Context, id, teamID, argoAppName string) error {
	return r.db.WithContext(ctx).
		Model(&environment.Environment{}).
		Where("id = ? AND team_id = ?", id, teamID).
		Update("argo_app_name", argoAppName).Error
}

// UpdateLastSync updates the last sync timestamp
func (r *repository) UpdateLastSync(ctx context.Context, id string, syncedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&environment.Environment{}).
		Where("id = ?", id).
		Update("last_sync_at", syncedAt).Error
}

// IncrementErrorCount increments the error count for an environment
func (r *repository) IncrementErrorCount(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&environment.Environment{}).
		Where("id = ?", id).
		UpdateColumn("error_count", gorm.Expr("error_count + 1")).Error
}

// ResetErrorCount resets the error count to zero
func (r *repository) ResetErrorCount(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&environment.Environment{}).
		Where("id = ?", id).
		Update("error_count", 0).Error
}

// SoftDelete soft deletes an environment
func (r *repository) SoftDelete(ctx context.Context, id, teamID string) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND team_id = ?", id, teamID).
		Delete(&environment.Environment{}).Error
}

// HardDelete permanently deletes an environment
func (r *repository) HardDelete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&environment.Environment{}).Error
}
