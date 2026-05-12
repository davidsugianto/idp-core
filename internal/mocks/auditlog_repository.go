package mocks

import (
	"context"
	"reflect"

	"github.com/davidsugianto/idp-core/internal/model/auditlog"
	"github.com/golang/mock/gomock"
)

// MockAuditLogRepository is a mock implementation of the audit log repository
type MockAuditLogRepository struct {
	ctrl     *gomock.Controller
	recorder *MockAuditLogRepositoryMockRecorder
}

type MockAuditLogRepositoryMockRecorder struct {
	mock *MockAuditLogRepository
}

func NewMockAuditLogRepository(ctrl *gomock.Controller) *MockAuditLogRepository {
	mock := &MockAuditLogRepository{ctrl: ctrl}
	mock.recorder = &MockAuditLogRepositoryMockRecorder{mock}
	return mock
}

func (m *MockAuditLogRepository) EXPECT() *MockAuditLogRepositoryMockRecorder {
	return m.recorder
}

func (m *MockAuditLogRepository) Create(ctx context.Context, log *auditlog.AuditLog) error {
	ret := m.ctrl.Call(m, "Create", ctx, log)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockAuditLogRepositoryMockRecorder) Create(ctx, log interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockAuditLogRepository)(nil).Create), ctx, log)
}

func (m *MockAuditLogRepository) GetByID(ctx context.Context, id string) (*auditlog.AuditLog, error) {
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*auditlog.AuditLog)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockAuditLogRepositoryMockRecorder) GetByID(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockAuditLogRepository)(nil).GetByID), ctx, id)
}

func (m *MockAuditLogRepository) List(ctx context.Context, filter auditlog.AuditLogFilter) ([]auditlog.AuditLog, int64, error) {
	ret := m.ctrl.Call(m, "List", ctx, filter)
	ret0, _ := ret[0].([]auditlog.AuditLog)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockAuditLogRepositoryMockRecorder) List(ctx, filter interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockAuditLogRepository)(nil).List), ctx, filter)
}