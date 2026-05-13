package cost

import (
	"context"

	"github.com/davidsugianto/idp-core/internal/model/cost"
	"gorm.io/gorm"
)

// Repository defines the interface for cost record persistence operations
type Repository interface {
	Create(ctx context.Context, record *cost.CostRecord) error
	BatchCreate(ctx context.Context, records []cost.CostRecord) error
	List(ctx context.Context, filter cost.CostFilter) ([]cost.CostRecord, int64, error)
	GetByTeamAndPeriod(ctx context.Context, teamID string, namespace string, start, end string) ([]cost.CostRecord, error)
}

type repository struct {
	db *gorm.DB
}

// Dependencies holds the dependencies for the cost repository
type Dependencies struct {
	Database *gorm.DB
}

// New creates a new cost repository
func New(deps Dependencies) Repository {
	return &repository{db: deps.Database}
}

// Create persists a new cost record
func (r *repository) Create(ctx context.Context, record *cost.CostRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// BatchCreate creates multiple cost records in a single transaction
func (r *repository) BatchCreate(ctx context.Context, records []cost.CostRecord) error {
	if len(records) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(records, 100).Error
}

// List retrieves cost records with filtering and pagination
func (r *repository) List(ctx context.Context, filter cost.CostFilter) ([]cost.CostRecord, int64, error) {
	var records []cost.CostRecord
	var total int64

	db := r.db.WithContext(ctx).Model(&cost.CostRecord{})

	if filter.TeamID != "" {
		db = db.Where("team_id = ?", filter.TeamID)
	}
	if filter.EnvironmentID != "" {
		db = db.Where("environment_id = ?", filter.EnvironmentID)
	}
	if filter.Namespace != "" {
		db = db.Where("namespace = ?", filter.Namespace)
	}
	if filter.StartDate != nil {
		db = db.Where("period_start >= ?", filter.StartDate)
	}
	if filter.EndDate != nil {
		db = db.Where("period_end <= ?", filter.EndDate)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	err := db.Order("period_start DESC").
		Limit(limit).
		Offset(filter.Offset).
		Find(&records).Error

	return records, total, err
}

// GetByTeamAndPeriod retrieves cost records for a specific team within a time range
func (r *repository) GetByTeamAndPeriod(ctx context.Context, teamID string, namespace string, start, end string) ([]cost.CostRecord, error) {
	var records []cost.CostRecord
	db := r.db.WithContext(ctx).Model(&cost.CostRecord{})

	if teamID != "" {
		db = db.Where("team_id = ?", teamID)
	}
	if namespace != "" {
		db = db.Where("namespace = ?", namespace)
	}
	if start != "" {
		db = db.Where("period_start >= ?", start)
	}
	if end != "" {
		db = db.Where("period_end <= ?", end)
	}

	err := db.Order("period_start DESC").Find(&records).Error
	return records, err
}