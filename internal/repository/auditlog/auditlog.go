package auditlog

import (
	"context"
	"errors"

	"github.com/davidsugianto/idp-core/internal/model/auditlog"
	"gorm.io/gorm"
)

// Create persists a new audit log entry
func (r *repository) Create(ctx context.Context, log *auditlog.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetByID retrieves an audit log entry by ID
func (r *repository) GetByID(ctx context.Context, id string) (*auditlog.AuditLog, error) {
	var log auditlog.AuditLog
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&log).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

// List retrieves audit logs with filtering and pagination
func (r *repository) List(ctx context.Context, filter auditlog.AuditLogFilter) ([]auditlog.AuditLog, int64, error) {
	var logs []auditlog.AuditLog
	var total int64

	db := r.db.WithContext(ctx).Model(&auditlog.AuditLog{})

	if filter.UserID != "" {
		db = db.Where("user_id = ?", filter.UserID)
	}
	if filter.TeamID != "" {
		db = db.Where("team_id = ?", filter.TeamID)
	}
	if filter.Action != "" {
		db = db.Where("action = ?", filter.Action)
	}
	if filter.ResourceType != "" {
		db = db.Where("resource_type = ?", filter.ResourceType)
	}
	if filter.ResourceID != "" {
		db = db.Where("resource_id = ?", filter.ResourceID)
	}
	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}
	if filter.StartDate != nil {
		db = db.Where("created_at >= ?", filter.StartDate)
	}
	if filter.EndDate != nil {
		db = db.Where("created_at <= ?", filter.EndDate)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	err := db.Order("created_at DESC").
		Limit(limit).
		Offset(filter.Offset).
		Find(&logs).Error

	return logs, total, err
}