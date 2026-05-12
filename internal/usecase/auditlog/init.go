package auditlog

import (
	"context"

	"github.com/davidsugianto/idp-core/internal/model/auditlog"
	auditlogRepo "github.com/davidsugianto/idp-core/internal/repository/auditlog"
)

// Usecase defines the interface for audit log business logic
type Usecase interface {
	Create(ctx context.Context, req auditlog.CreateAuditLogRequest) (*auditlog.AuditLog, error)
	Get(ctx context.Context, id string) (*auditlog.AuditLogResponse, error)
	List(ctx context.Context, filter auditlog.AuditLogFilter) (*auditlog.AuditLogListResponse, error)
}

type usecase struct {
	auditLogRepo auditlogRepo.Repository
}

type Dependencies struct {
	AuditLogRepo auditlogRepo.Repository
}

func New(deps Dependencies) Usecase {
	return &usecase{auditLogRepo: deps.AuditLogRepo}
}
