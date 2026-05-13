package mocks

import (
	"context"
	"reflect"

	"github.com/davidsugianto/idp-core/internal/model/cost"
	"github.com/golang/mock/gomock"
)

// MockCostRepository is a mock implementation of the cost repository
type MockCostRepository struct {
	ctrl     *gomock.Controller
	recorder *MockCostRepositoryMockRecorder
}

type MockCostRepositoryMockRecorder struct {
	mock *MockCostRepository
}

func NewMockCostRepository(ctrl *gomock.Controller) *MockCostRepository {
	mock := &MockCostRepository{ctrl: ctrl}
	mock.recorder = &MockCostRepositoryMockRecorder{mock}
	return mock
}

func (m *MockCostRepository) EXPECT() *MockCostRepositoryMockRecorder {
	return m.recorder
}

func (m *MockCostRepository) Create(ctx context.Context, record *cost.CostRecord) error {
	ret := m.ctrl.Call(m, "Create", ctx, record)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockCostRepositoryMockRecorder) Create(ctx, record any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockCostRepository)(nil).Create), ctx, record)
}

func (m *MockCostRepository) BatchCreate(ctx context.Context, records []cost.CostRecord) error {
	ret := m.ctrl.Call(m, "BatchCreate", ctx, records)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockCostRepositoryMockRecorder) BatchCreate(ctx, records any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BatchCreate", reflect.TypeOf((*MockCostRepository)(nil).BatchCreate), ctx, records)
}

func (m *MockCostRepository) List(ctx context.Context, filter cost.CostFilter) ([]cost.CostRecord, int64, error) {
	ret := m.ctrl.Call(m, "List", ctx, filter)
	ret0, _ := ret[0].([]cost.CostRecord)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockCostRepositoryMockRecorder) List(ctx, filter any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockCostRepository)(nil).List), ctx, filter)
}

func (m *MockCostRepository) GetByTeamAndPeriod(ctx context.Context, teamID string, namespace string, start, end string) ([]cost.CostRecord, error) {
	ret := m.ctrl.Call(m, "GetByTeamAndPeriod", ctx, teamID, namespace, start, end)
	ret0, _ := ret[0].([]cost.CostRecord)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockCostRepositoryMockRecorder) GetByTeamAndPeriod(ctx, teamID, namespace, start, end any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByTeamAndPeriod", reflect.TypeOf((*MockCostRepository)(nil).GetByTeamAndPeriod), ctx, teamID, namespace, start, end)
}