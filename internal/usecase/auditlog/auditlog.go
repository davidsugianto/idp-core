package auditlog

import (
	"context"

	"github.com/davidsugianto/idp-core/internal/model/auditlog"
	"github.com/google/uuid"
)

// Create creates a new audit log entry
func (u *usecase) Create(ctx context.Context, req auditlog.CreateAuditLogRequest) (*auditlog.AuditLog, error) {
	log := &auditlog.AuditLog{
		ID:            uuid.New().String(),
		UserID:        req.UserID,
		UserEmail:     req.UserEmail,
		ActorType:     req.ActorType,
		Action:        req.Action,
		ResourceType:  req.ResourceType,
		ResourceID:    req.ResourceID,
		TeamID:        req.TeamID,
		EnvironmentID: req.EnvironmentID,
		IPAddress:     req.IPAddress,
		UserAgent:     req.UserAgent,
		RequestMethod: req.RequestMethod,
		RequestPath:   req.RequestPath,
		RequestID:     req.RequestID,
		OldValues:     req.OldValues,
		NewValues:     req.NewValues,
		Status:        req.Status,
		ErrorMessage:  req.ErrorMessage,
	}

	if log.Status == "" {
		log.Status = auditlog.StatusSuccess
	}

	if err := u.auditLogRepo.Create(ctx, log); err != nil {
		return nil, err
	}

	return log, nil
}

// Get retrieves an audit log entry by ID
func (u *usecase) Get(ctx context.Context, id string) (*auditlog.AuditLogResponse, error) {
	log, err := u.auditLogRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return auditlog.ToAuditLogResponse(log), nil
}

// List retrieves audit logs with filtering
func (u *usecase) List(ctx context.Context, filter auditlog.AuditLogFilter) (*auditlog.AuditLogListResponse, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 200 {
		filter.Limit = 200
	}

	logs, total, err := u.auditLogRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	return auditlog.ToAuditLogListResponse(logs, total), nil
}
