package mocks

import (
	"context"
	"reflect"

	"github.com/davidsugianto/idp-core/internal/model/resourcequota"
	"github.com/golang/mock/gomock"
)

type MockQuotaRepository struct {
	ctrl     *gomock.Controller
	recorder *MockQuotaRepositoryMockRecorder
}

type MockQuotaRepositoryMockRecorder struct {
	mock *MockQuotaRepository
}

func NewMockQuotaRepository(ctrl *gomock.Controller) *MockQuotaRepository {
	mock := &MockQuotaRepository{ctrl: ctrl}
	mock.recorder = &MockQuotaRepositoryMockRecorder{mock}
	return mock
}

func (m *MockQuotaRepository) EXPECT() *MockQuotaRepositoryMockRecorder {
	return m.recorder
}

func (m *MockQuotaRepository) Create(ctx context.Context, quota *resourcequota.ResourceQuota) error {
	ret := m.ctrl.Call(m, "Create", ctx, quota)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockQuotaRepositoryMockRecorder) Create(ctx, quota interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockQuotaRepository)(nil).Create), ctx, quota)
}

func (m *MockQuotaRepository) GetByID(ctx context.Context, id string) (*resourcequota.ResourceQuota, error) {
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*resourcequota.ResourceQuota)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockQuotaRepositoryMockRecorder) GetByID(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockQuotaRepository)(nil).GetByID), ctx, id)
}

func (m *MockQuotaRepository) GetByNamespace(ctx context.Context, namespace string) (*resourcequota.ResourceQuota, error) {
	ret := m.ctrl.Call(m, "GetByNamespace", ctx, namespace)
	ret0, _ := ret[0].(*resourcequota.ResourceQuota)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockQuotaRepositoryMockRecorder) GetByNamespace(ctx, namespace interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByNamespace", reflect.TypeOf((*MockQuotaRepository)(nil).GetByNamespace), ctx, namespace)
}

func (m *MockQuotaRepository) List(ctx context.Context, req *resourcequota.ListResourceQuotasRequest) ([]resourcequota.ResourceQuota, int64, error) {
	ret := m.ctrl.Call(m, "List", ctx, req)
	ret0, _ := ret[0].([]resourcequota.ResourceQuota)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockQuotaRepositoryMockRecorder) List(ctx, req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockQuotaRepository)(nil).List), ctx, req)
}

func (m *MockQuotaRepository) Update(ctx context.Context, quota *resourcequota.ResourceQuota) error {
	ret := m.ctrl.Call(m, "Update", ctx, quota)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockQuotaRepositoryMockRecorder) Update(ctx, quota interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockQuotaRepository)(nil).Update), ctx, quota)
}

func (m *MockQuotaRepository) Delete(ctx context.Context, id string) error {
	ret := m.ctrl.Call(m, "Delete", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockQuotaRepositoryMockRecorder) Delete(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockQuotaRepository)(nil).Delete), ctx, id)
}

func (m *MockQuotaRepository) UpdateUsage(ctx context.Context, namespace string, usage *resourcequota.UsageResponse) error {
	ret := m.ctrl.Call(m, "UpdateUsage", ctx, namespace, usage)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockQuotaRepositoryMockRecorder) UpdateUsage(ctx, namespace, usage interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUsage", reflect.TypeOf((*MockQuotaRepository)(nil).UpdateUsage), ctx, namespace, usage)
}

func (m *MockQuotaRepository) GetActiveByNamespace(ctx context.Context, namespace string) (*resourcequota.ResourceQuota, error) {
	ret := m.ctrl.Call(m, "GetActiveByNamespace", ctx, namespace)
	ret0, _ := ret[0].(*resourcequota.ResourceQuota)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockQuotaRepositoryMockRecorder) GetActiveByNamespace(ctx, namespace interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetActiveByNamespace", reflect.TypeOf((*MockQuotaRepository)(nil).GetActiveByNamespace), ctx, namespace)
}

func (m *MockQuotaRepository) ListActiveByTeam(ctx context.Context, teamID string) ([]resourcequota.ResourceQuota, error) {
	ret := m.ctrl.Call(m, "ListActiveByTeam", ctx, teamID)
	ret0, _ := ret[0].([]resourcequota.ResourceQuota)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockQuotaRepositoryMockRecorder) ListActiveByTeam(ctx, teamID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListActiveByTeam", reflect.TypeOf((*MockQuotaRepository)(nil).ListActiveByTeam), ctx, teamID)
}

func (m *MockQuotaRepository) ExistsForNamespace(ctx context.Context, namespace string) (bool, error) {
	ret := m.ctrl.Call(m, "ExistsForNamespace", ctx, namespace)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockQuotaRepositoryMockRecorder) ExistsForNamespace(ctx, namespace interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExistsForNamespace", reflect.TypeOf((*MockQuotaRepository)(nil).ExistsForNamespace), ctx, namespace)
}
