package apikey

import (
	"context"
	"time"

	"github.com/davidsugianto/idp-core/internal/model/apikey"
	"gorm.io/gorm"
)

// Repository defines the interface for API key persistence operations
type Repository interface {
	Create(ctx context.Context, key *apikey.APIKey) error
	GetByID(ctx context.Context, id string) (*apikey.APIKey, error)
	GetByKey(ctx context.Context, key string) (*apikey.APIKey, error)
	ListByTeam(ctx context.Context, teamID string) ([]apikey.APIKey, error)
	ListActive(ctx context.Context) ([]apikey.APIKey, error)
	Update(ctx context.Context, key *apikey.APIKey) error
	UpdateLastUsed(ctx context.Context, id string) error
	IncrementUsage(ctx context.Context, id string) error
	Deactivate(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) error
}

type repository struct {
	db *gorm.DB
}

type Dependencies struct {
	Database *gorm.DB
}

func New(deps Dependencies) Repository {
	return &repository{db: deps.Database}
}

// Create persists a new API key
func (r *repository) Create(ctx context.Context, key *apikey.APIKey) error {
	return r.db.WithContext(ctx).Create(key).Error
}

// GetByID retrieves an API key by ID
func (r *repository) GetByID(ctx context.Context, id string) (*apikey.APIKey, error) {
	var key apikey.APIKey
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// GetByKey retrieves an API key by the key value (for authentication)
func (r *repository) GetByKey(ctx context.Context, key string) (*apikey.APIKey, error) {
	var apiKey apikey.APIKey
	err := r.db.WithContext(ctx).
		Where("key = ? AND is_active = ?", key, true).
		First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// ListByTeam lists all API keys for a team
func (r *repository) ListByTeam(ctx context.Context, teamID string) ([]apikey.APIKey, error) {
	var keys []apikey.APIKey
	err := r.db.WithContext(ctx).
		Where("team_id = ?", teamID).
		Order("created_at DESC").
		Find(&keys).Error
	return keys, err
}

// ListActive lists all active API keys
func (r *repository) ListActive(ctx context.Context) ([]apikey.APIKey, error) {
	var keys []apikey.APIKey
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Order("created_at DESC").
		Find(&keys).Error
	return keys, err
}

// Update updates an API key
func (r *repository) Update(ctx context.Context, key *apikey.APIKey) error {
	return r.db.WithContext(ctx).Save(key).Error
}

// UpdateLastUsed updates the last used timestamp
func (r *repository) UpdateLastUsed(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&apikey.APIKey{}).
		Where("id = ?", id).
		Update("last_used_at", now).Error
}

// IncrementUsage increments the usage count
func (r *repository) IncrementUsage(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&apikey.APIKey{}).
		Where("id = ?", id).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error
}

// Deactivate deactivates an API key
func (r *repository) Deactivate(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&apikey.APIKey{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

// Delete soft deletes an API key
func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&apikey.APIKey{}).Error
}

// DeleteExpired removes all expired API keys
func (r *repository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Delete(&apikey.APIKey{}).Error
}
