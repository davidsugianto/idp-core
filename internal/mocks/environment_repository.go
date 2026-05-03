package mocks

import (
	"context"
	"reflect"
	"time"

	"github.com/davidsugianto/idp-core/internal/model/environment"
	"github.com/golang/mock/gomock"
)

// MockEnvironmentRepository is a mock implementation of the Environment repository
type MockEnvironmentRepository struct {
	ctrl     *gomock.Controller
	recorder *MockEnvironmentRepositoryMockRecorder
}

type MockEnvironmentRepositoryMockRecorder struct {
	mock *MockEnvironmentRepository
}

func NewMockEnvironmentRepository(ctrl *gomock.Controller) *MockEnvironmentRepository {
	mock := &MockEnvironmentRepository{ctrl: ctrl}
	mock.recorder = &MockEnvironmentRepositoryMockRecorder{mock}
	return mock
}

func (m *MockEnvironmentRepository) EXPECT() *MockEnvironmentRepositoryMockRecorder {
	return m.recorder
}

func (m *MockEnvironmentRepository) Create(ctx context.Context, env *environment.Environment) error {
	ret := m.ctrl.Call(m, "Create", ctx, env)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockEnvironmentRepositoryMockRecorder) Create(ctx, env interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockEnvironmentRepository)(nil).Create), ctx, env)
}

func (m *MockEnvironmentRepository) GetByID(ctx context.Context, id string) (*environment.Environment, error) {
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*environment.Environment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockEnvironmentRepositoryMockRecorder) GetByID(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockEnvironmentRepository)(nil).GetByID), ctx, id)
}

func (m *MockEnvironmentRepository) GetByIDAndTeam(ctx context.Context, id, teamID string) (*environment.Environment, error) {
	ret := m.ctrl.Call(m, "GetByIDAndTeam", ctx, id, teamID)
	ret0, _ := ret[0].(*environment.Environment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockEnvironmentRepositoryMockRecorder) GetByIDAndTeam(ctx, id, teamID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByIDAndTeam", reflect.TypeOf((*MockEnvironmentRepository)(nil).GetByIDAndTeam), ctx, id, teamID)
}

func (m *MockEnvironmentRepository) GetByNamespace(ctx context.Context, namespace string) (*environment.Environment, error) {
	ret := m.ctrl.Call(m, "GetByNamespace", ctx, namespace)
	ret0, _ := ret[0].(*environment.Environment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockEnvironmentRepositoryMockRecorder) GetByNamespace(ctx, namespace interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByNamespace", reflect.TypeOf((*MockEnvironmentRepository)(nil).GetByNamespace), ctx, namespace)
}

func (m *MockEnvironmentRepository) ListByTeam(ctx context.Context, teamID string) ([]environment.Environment, error) {
	ret := m.ctrl.Call(m, "ListByTeam", ctx, teamID)
	ret0, _ := ret[0].([]environment.Environment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockEnvironmentRepositoryMockRecorder) ListByTeam(ctx, teamID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByTeam", reflect.TypeOf((*MockEnvironmentRepository)(nil).ListByTeam), ctx, teamID)
}

func (m *MockEnvironmentRepository) ListByStatus(ctx context.Context, status string) ([]environment.Environment, error) {
	ret := m.ctrl.Call(m, "ListByStatus", ctx, status)
	ret0, _ := ret[0].([]environment.Environment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockEnvironmentRepositoryMockRecorder) ListByStatus(ctx, status interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByStatus", reflect.TypeOf((*MockEnvironmentRepository)(nil).ListByStatus), ctx, status)
}

func (m *MockEnvironmentRepository) ListExpired(ctx context.Context) ([]environment.Environment, error) {
	ret := m.ctrl.Call(m, "ListExpired", ctx)
	ret0, _ := ret[0].([]environment.Environment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockEnvironmentRepositoryMockRecorder) ListExpired(ctx interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListExpired", reflect.TypeOf((*MockEnvironmentRepository)(nil).ListExpired), ctx)
}

func (m *MockEnvironmentRepository) Update(ctx context.Context, env *environment.Environment) error {
	ret := m.ctrl.Call(m, "Update", ctx, env)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockEnvironmentRepositoryMockRecorder) Update(ctx, env interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockEnvironmentRepository)(nil).Update), ctx, env)
}

func (m *MockEnvironmentRepository) UpdateStatus(ctx context.Context, id, teamID, status, lastError string) error {
	ret := m.ctrl.Call(m, "UpdateStatus", ctx, id, teamID, status, lastError)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockEnvironmentRepositoryMockRecorder) UpdateStatus(ctx, id, teamID, status, lastError interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateStatus", reflect.TypeOf((*MockEnvironmentRepository)(nil).UpdateStatus), ctx, id, teamID, status, lastError)
}

func (m *MockEnvironmentRepository) UpdateArgoAppName(ctx context.Context, id, teamID, argoAppName string) error {
	ret := m.ctrl.Call(m, "UpdateArgoAppName", ctx, id, teamID, argoAppName)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockEnvironmentRepositoryMockRecorder) UpdateArgoAppName(ctx, id, teamID, argoAppName interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateArgoAppName", reflect.TypeOf((*MockEnvironmentRepository)(nil).UpdateArgoAppName), ctx, id, teamID, argoAppName)
}

func (m *MockEnvironmentRepository) UpdateLastSync(ctx context.Context, id string, syncedAt time.Time) error {
	ret := m.ctrl.Call(m, "UpdateLastSync", ctx, id, syncedAt)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockEnvironmentRepositoryMockRecorder) UpdateLastSync(ctx, id, syncedAt interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateLastSync", reflect.TypeOf((*MockEnvironmentRepository)(nil).UpdateLastSync), ctx, id, syncedAt)
}

func (m *MockEnvironmentRepository) IncrementErrorCount(ctx context.Context, id string) error {
	ret := m.ctrl.Call(m, "IncrementErrorCount", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockEnvironmentRepositoryMockRecorder) IncrementErrorCount(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IncrementErrorCount", reflect.TypeOf((*MockEnvironmentRepository)(nil).IncrementErrorCount), ctx, id)
}

func (m *MockEnvironmentRepository) ResetErrorCount(ctx context.Context, id string) error {
	ret := m.ctrl.Call(m, "ResetErrorCount", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockEnvironmentRepositoryMockRecorder) ResetErrorCount(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResetErrorCount", reflect.TypeOf((*MockEnvironmentRepository)(nil).ResetErrorCount), ctx, id)
}

func (m *MockEnvironmentRepository) SoftDelete(ctx context.Context, id, teamID string) error {
	ret := m.ctrl.Call(m, "SoftDelete", ctx, id, teamID)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockEnvironmentRepositoryMockRecorder) SoftDelete(ctx, id, teamID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SoftDelete", reflect.TypeOf((*MockEnvironmentRepository)(nil).SoftDelete), ctx, id, teamID)
}

func (m *MockEnvironmentRepository) HardDelete(ctx context.Context, id string) error {
	ret := m.ctrl.Call(m, "HardDelete", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockEnvironmentRepositoryMockRecorder) HardDelete(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HardDelete", reflect.TypeOf((*MockEnvironmentRepository)(nil).HardDelete), ctx, id)
}
