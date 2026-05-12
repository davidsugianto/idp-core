package mocks

import (
	"context"
	"reflect"

	permissionModel "github.com/davidsugianto/idp-core/internal/model/permission"
	"github.com/golang/mock/gomock"
)

// MockPermissionRepository is a mock implementation of the Permission repository
type MockPermissionRepository struct {
	ctrl     *gomock.Controller
	recorder *MockPermissionRepositoryMockRecorder
}

type MockPermissionRepositoryMockRecorder struct {
	mock *MockPermissionRepository
}

func NewMockPermissionRepository(ctrl *gomock.Controller) *MockPermissionRepository {
	mock := &MockPermissionRepository{ctrl: ctrl}
	mock.recorder = &MockPermissionRepositoryMockRecorder{mock}
	return mock
}

func (m *MockPermissionRepository) EXPECT() *MockPermissionRepositoryMockRecorder {
	return m.recorder
}

func (m *MockPermissionRepository) Create(ctx context.Context, permission *permissionModel.Permission) error {
	ret := m.ctrl.Call(m, "Create", ctx, permission)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockPermissionRepositoryMockRecorder) Create(ctx, permission interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockPermissionRepository)(nil).Create), ctx, permission)
}

func (m *MockPermissionRepository) GetByID(ctx context.Context, id string) (*permissionModel.Permission, error) {
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*permissionModel.Permission)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockPermissionRepositoryMockRecorder) GetByID(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockPermissionRepository)(nil).GetByID), ctx, id)
}

func (m *MockPermissionRepository) GetByName(ctx context.Context, name string) (*permissionModel.Permission, error) {
	ret := m.ctrl.Call(m, "GetByName", ctx, name)
	ret0, _ := ret[0].(*permissionModel.Permission)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockPermissionRepositoryMockRecorder) GetByName(ctx, name interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByName", reflect.TypeOf((*MockPermissionRepository)(nil).GetByName), ctx, name)
}

func (m *MockPermissionRepository) GetByResourceAction(ctx context.Context, resource, action string) (*permissionModel.Permission, error) {
	ret := m.ctrl.Call(m, "GetByResourceAction", ctx, resource, action)
	ret0, _ := ret[0].(*permissionModel.Permission)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockPermissionRepositoryMockRecorder) GetByResourceAction(ctx, resource, action interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByResourceAction", reflect.TypeOf((*MockPermissionRepository)(nil).GetByResourceAction), ctx, resource, action)
}

func (m *MockPermissionRepository) List(ctx context.Context, limit, offset int) ([]permissionModel.Permission, int64, error) {
	ret := m.ctrl.Call(m, "List", ctx, limit, offset)
	ret0, _ := ret[0].([]permissionModel.Permission)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockPermissionRepositoryMockRecorder) List(ctx, limit, offset interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockPermissionRepository)(nil).List), ctx, limit, offset)
}

func (m *MockPermissionRepository) Update(ctx context.Context, permission *permissionModel.Permission) error {
	ret := m.ctrl.Call(m, "Update", ctx, permission)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockPermissionRepositoryMockRecorder) Update(ctx, permission interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockPermissionRepository)(nil).Update), ctx, permission)
}

func (m *MockPermissionRepository) Delete(ctx context.Context, id string) error {
	ret := m.ctrl.Call(m, "Delete", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockPermissionRepositoryMockRecorder) Delete(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockPermissionRepository)(nil).Delete), ctx, id)
}
