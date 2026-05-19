package mocks

import (
	"context"
	"reflect"

	"github.com/davidsugianto/idp-core/internal/model/rightsizing"
	"github.com/golang/mock/gomock"
)

type MockRightsizingRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRightsizingRepositoryMockRecorder
}

type MockRightsizingRepositoryMockRecorder struct {
	mock *MockRightsizingRepository
}

func NewMockRightsizingRepository(ctrl *gomock.Controller) *MockRightsizingRepository {
	mock := &MockRightsizingRepository{ctrl: ctrl}
	mock.recorder = &MockRightsizingRepositoryMockRecorder{mock}
	return mock
}

func (m *MockRightsizingRepository) EXPECT() *MockRightsizingRepositoryMockRecorder {
	return m.recorder
}

func (m *MockRightsizingRepository) Create(ctx context.Context, rec *rightsizing.RightsizingRecommendation) error {
	ret := m.ctrl.Call(m, "Create", ctx, rec)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockRightsizingRepositoryMockRecorder) Create(ctx, rec interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockRightsizingRepository)(nil).Create), ctx, rec)
}

func (m *MockRightsizingRepository) GetByID(ctx context.Context, id string) (*rightsizing.RightsizingRecommendation, error) {
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*rightsizing.RightsizingRecommendation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockRightsizingRepositoryMockRecorder) GetByID(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockRightsizingRepository)(nil).GetByID), ctx, id)
}

func (m *MockRightsizingRepository) List(ctx context.Context, req *rightsizing.ListRecommendationsRequest) ([]rightsizing.RightsizingRecommendation, int64, error) {
	ret := m.ctrl.Call(m, "List", ctx, req)
	ret0, _ := ret[0].([]rightsizing.RightsizingRecommendation)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockRightsizingRepositoryMockRecorder) List(ctx, req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockRightsizingRepository)(nil).List), ctx, req)
}

func (m *MockRightsizingRepository) Update(ctx context.Context, rec *rightsizing.RightsizingRecommendation) error {
	ret := m.ctrl.Call(m, "Update", ctx, rec)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockRightsizingRepositoryMockRecorder) Update(ctx, rec interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockRightsizingRepository)(nil).Update), ctx, rec)
}

func (m *MockRightsizingRepository) Delete(ctx context.Context, id string) error {
	ret := m.ctrl.Call(m, "Delete", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockRightsizingRepositoryMockRecorder) Delete(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockRightsizingRepository)(nil).Delete), ctx, id)
}

func (m *MockRightsizingRepository) DeletePendingByWorkload(ctx context.Context, namespace, workloadName, workloadType string) error {
	ret := m.ctrl.Call(m, "DeletePendingByWorkload", ctx, namespace, workloadName, workloadType)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockRightsizingRepositoryMockRecorder) DeletePendingByWorkload(ctx, namespace, workloadName, workloadType interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePendingByWorkload", reflect.TypeOf((*MockRightsizingRepository)(nil).DeletePendingByWorkload), ctx, namespace, workloadName, workloadType)
}

func (m *MockRightsizingRepository) ListPendingByWorkload(ctx context.Context, namespace, workloadName, workloadType string) ([]rightsizing.RightsizingRecommendation, error) {
	ret := m.ctrl.Call(m, "ListPendingByWorkload", ctx, namespace, workloadName, workloadType)
	ret0, _ := ret[0].([]rightsizing.RightsizingRecommendation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockRightsizingRepositoryMockRecorder) ListPendingByWorkload(ctx, namespace, workloadName, workloadType interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListPendingByWorkload", reflect.TypeOf((*MockRightsizingRepository)(nil).ListPendingByWorkload), ctx, namespace, workloadName, workloadType)
}

func (m *MockRightsizingRepository) ExistsPendingForContainer(ctx context.Context, namespace, workloadName, workloadType, containerName string) (bool, error) {
	ret := m.ctrl.Call(m, "ExistsPendingForContainer", ctx, namespace, workloadName, workloadType, containerName)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockRightsizingRepositoryMockRecorder) ExistsPendingForContainer(ctx, namespace, workloadName, workloadType, containerName interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExistsPendingForContainer", reflect.TypeOf((*MockRightsizingRepository)(nil).ExistsPendingForContainer), ctx, namespace, workloadName, workloadType, containerName)
}
