package auditlog

import (
	"context"

	"github.com/davidsugianto/idp-core/internal/model/auditlog"
	"gorm.io/gorm"
)

// Repository defines the interface for audit log persistence operations
type Repository interface {
	Create(ctx context.Context, log *auditlog.AuditLog) error
	GetByID(ctx context.Context, id string) (*auditlog.AuditLog, error)
	List(ctx context.Context, filter auditlog.AuditLogFilter) ([]auditlog.AuditLog, int64, error)
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