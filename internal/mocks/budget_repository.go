package mocks

import (
	"context"
	"reflect"
	"time"

	"github.com/davidsugianto/idp-core/internal/model/budget"
	"github.com/golang/mock/gomock"
)

type MockBudgetRepository struct {
	ctrl     *gomock.Controller
	recorder *MockBudgetRepositoryMockRecorder
}

type MockBudgetRepositoryMockRecorder struct {
	mock *MockBudgetRepository
}

func NewMockBudgetRepository(ctrl *gomock.Controller) *MockBudgetRepository {
	mock := &MockBudgetRepository{ctrl: ctrl}
	mock.recorder = &MockBudgetRepositoryMockRecorder{mock}
	return mock
}

func (m *MockBudgetRepository) EXPECT() *MockBudgetRepositoryMockRecorder {
	return m.recorder
}

func (m *MockBudgetRepository) Create(ctx context.Context, b *budget.Budget) error {
	ret := m.ctrl.Call(m, "Create", ctx, b)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockBudgetRepositoryMockRecorder) Create(ctx, b any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockBudgetRepository)(nil).Create), ctx, b)
}

func (m *MockBudgetRepository) GetByID(ctx context.Context, id string) (*budget.Budget, error) {
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*budget.Budget)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockBudgetRepositoryMockRecorder) GetByID(ctx, id any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockBudgetRepository)(nil).GetByID), ctx, id)
}

func (m *MockBudgetRepository) ListByTeam(ctx context.Context, teamID string) ([]budget.Budget, error) {
	ret := m.ctrl.Call(m, "ListByTeam", ctx, teamID)
	ret0, _ := ret[0].([]budget.Budget)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockBudgetRepositoryMockRecorder) ListByTeam(ctx, teamID any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByTeam", reflect.TypeOf((*MockBudgetRepository)(nil).ListByTeam), ctx, teamID)
}

func (m *MockBudgetRepository) ListActive(ctx context.Context) ([]budget.Budget, error) {
	ret := m.ctrl.Call(m, "ListActive", ctx)
	ret0, _ := ret[0].([]budget.Budget)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockBudgetRepositoryMockRecorder) ListActive(ctx any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListActive", reflect.TypeOf((*MockBudgetRepository)(nil).ListActive), ctx)
}

func (m *MockBudgetRepository) Update(ctx context.Context, b *budget.Budget) error {
	ret := m.ctrl.Call(m, "Update", ctx, b)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockBudgetRepositoryMockRecorder) Update(ctx, b any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockBudgetRepository)(nil).Update), ctx, b)
}

func (m *MockBudgetRepository) Delete(ctx context.Context, id string) error {
	ret := m.ctrl.Call(m, "Delete", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockBudgetRepositoryMockRecorder) Delete(ctx, id any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockBudgetRepository)(nil).Delete), ctx, id)
}

func (m *MockBudgetRepository) CreateAlert(ctx context.Context, alert *budget.BudgetAlert) error {
	ret := m.ctrl.Call(m, "CreateAlert", ctx, alert)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockBudgetRepositoryMockRecorder) CreateAlert(ctx, alert any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateAlert", reflect.TypeOf((*MockBudgetRepository)(nil).CreateAlert), ctx, alert)
}

func (m *MockBudgetRepository) GetAlertsByBudget(ctx context.Context, budgetID string) ([]budget.BudgetAlert, error) {
	ret := m.ctrl.Call(m, "GetAlertsByBudget", ctx, budgetID)
	ret0, _ := ret[0].([]budget.BudgetAlert)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockBudgetRepositoryMockRecorder) GetAlertsByBudget(ctx, budgetID any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAlertsByBudget", reflect.TypeOf((*MockBudgetRepository)(nil).GetAlertsByBudget), ctx, budgetID)
}

func (m *MockBudgetRepository) GetLatestAlertForThreshold(ctx context.Context, budgetID string, threshold int, periodStart time.Time) (*budget.BudgetAlert, error) {
	ret := m.ctrl.Call(m, "GetLatestAlertForThreshold", ctx, budgetID, threshold, periodStart)
	ret0, _ := ret[0].(*budget.BudgetAlert)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockBudgetRepositoryMockRecorder) GetLatestAlertForThreshold(ctx, budgetID, threshold, periodStart any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLatestAlertForThreshold", reflect.TypeOf((*MockBudgetRepository)(nil).GetLatestAlertForThreshold), ctx, budgetID, threshold, periodStart)
}