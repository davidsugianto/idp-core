package mocks

import (
	"context"
	"reflect"

	roleModel "github.com/davidsugianto/idp-core/internal/model/role"
	"github.com/golang/mock/gomock"
)

// MockRoleRepository is a mock implementation of the Role repository
type MockRoleRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRoleRepositoryMockRecorder
}

type MockRoleRepositoryMockRecorder struct {
	mock *MockRoleRepository
}

func NewMockRoleRepository(ctrl *gomock.Controller) *MockRoleRepository {
	mock := &MockRoleRepository{ctrl: ctrl}
	mock.recorder = &MockRoleRepositoryMockRecorder{mock}
	return mock
}

func (m *MockRoleRepository) EXPECT() *MockRoleRepositoryMockRecorder {
	return m.recorder
}

func (m *MockRoleRepository) Create(ctx context.Context, role *roleModel.Role) error {
	ret := m.ctrl.Call(m, "Create", ctx, role)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockRoleRepositoryMockRecorder) Create(ctx, role interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockRoleRepository)(nil).Create), ctx, role)
}

func (m *MockRoleRepository) GetByID(ctx context.Context, id string) (*roleModel.Role, error) {
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*roleModel.Role)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockRoleRepositoryMockRecorder) GetByID(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockRoleRepository)(nil).GetByID), ctx, id)
}

func (m *MockRoleRepository) GetByName(ctx context.Context, name string) (*roleModel.Role, error) {
	ret := m.ctrl.Call(m, "GetByName", ctx, name)
	ret0, _ := ret[0].(*roleModel.Role)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockRoleRepositoryMockRecorder) GetByName(ctx, name interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByName", reflect.TypeOf((*MockRoleRepository)(nil).GetByName), ctx, name)
}

func (m *MockRoleRepository) List(ctx context.Context, limit, offset int) ([]roleModel.Role, int64, error) {
	ret := m.ctrl.Call(m, "List", ctx, limit, offset)
	ret0, _ := ret[0].([]roleModel.Role)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockRoleRepositoryMockRecorder) List(ctx, limit, offset interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockRoleRepository)(nil).List), ctx, limit, offset)
}

func (m *MockRoleRepository) Update(ctx context.Context, role *roleModel.Role) error {
	ret := m.ctrl.Call(m, "Update", ctx, role)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockRoleRepositoryMockRecorder) Update(ctx, role interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockRoleRepository)(nil).Update), ctx, role)
}

func (m *MockRoleRepository) SoftDelete(ctx context.Context, id string) error {
	ret := m.ctrl.Call(m, "SoftDelete", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockRoleRepositoryMockRecorder) SoftDelete(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SoftDelete", reflect.TypeOf((*MockRoleRepository)(nil).SoftDelete), ctx, id)
}

func (m *MockRoleRepository) AddPermission(ctx context.Context, roleID, permissionID string) error {
	ret := m.ctrl.Call(m, "AddPermission", ctx, roleID, permissionID)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockRoleRepositoryMockRecorder) AddPermission(ctx, roleID, permissionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddPermission", reflect.TypeOf((*MockRoleRepository)(nil).AddPermission), ctx, roleID, permissionID)
}

func (m *MockRoleRepository) RemovePermission(ctx context.Context, roleID, permissionID string) error {
	ret := m.ctrl.Call(m, "RemovePermission", ctx, roleID, permissionID)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockRoleRepositoryMockRecorder) RemovePermission(ctx, roleID, permissionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemovePermission", reflect.TypeOf((*MockRoleRepository)(nil).RemovePermission), ctx, roleID, permissionID)
}

func (m *MockRoleRepository) GetRolePermissions(ctx context.Context, roleID string) ([]roleModel.Role, error) {
	ret := m.ctrl.Call(m, "GetRolePermissions", ctx, roleID)
	ret0, _ := ret[0].([]roleModel.Role)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockRoleRepositoryMockRecorder) GetRolePermissions(ctx, roleID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRolePermissions", reflect.TypeOf((*MockRoleRepository)(nil).GetRolePermissions), ctx, roleID)
}

func (m *MockRoleRepository) SetPermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	ret := m.ctrl.Call(m, "SetPermissions", ctx, roleID, permissionIDs)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockRoleRepositoryMockRecorder) SetPermissions(ctx, roleID, permissionIDs interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetPermissions", reflect.TypeOf((*MockRoleRepository)(nil).SetPermissions), ctx, roleID, permissionIDs)
}

func (m *MockRoleRepository) AssignRole(ctx context.Context, userRole *roleModel.UserRole) error {
	ret := m.ctrl.Call(m, "AssignRole", ctx, userRole)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockRoleRepositoryMockRecorder) AssignRole(ctx, userRole interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AssignRole", reflect.TypeOf((*MockRoleRepository)(nil).AssignRole), ctx, userRole)
}

func (m *MockRoleRepository) RevokeRole(ctx context.Context, userID, roleID, teamID string) error {
	ret := m.ctrl.Call(m, "RevokeRole", ctx, userID, roleID, teamID)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockRoleRepositoryMockRecorder) RevokeRole(ctx, userID, roleID, teamID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RevokeRole", reflect.TypeOf((*MockRoleRepository)(nil).RevokeRole), ctx, userID, roleID, teamID)
}

func (m *MockRoleRepository) GetUserRoles(ctx context.Context, userID string) ([]roleModel.UserRole, error) {
	ret := m.ctrl.Call(m, "GetUserRoles", ctx, userID)
	ret0, _ := ret[0].([]roleModel.UserRole)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockRoleRepositoryMockRecorder) GetUserRoles(ctx, userID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserRoles", reflect.TypeOf((*MockRoleRepository)(nil).GetUserRoles), ctx, userID)
}

func (m *MockRoleRepository) GetUserRolesByTeam(ctx context.Context, userID, teamID string) ([]roleModel.UserRole, error) {
	ret := m.ctrl.Call(m, "GetUserRolesByTeam", ctx, userID, teamID)
	ret0, _ := ret[0].([]roleModel.UserRole)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockRoleRepositoryMockRecorder) GetUserRolesByTeam(ctx, userID, teamID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserRolesByTeam", reflect.TypeOf((*MockRoleRepository)(nil).GetUserRolesByTeam), ctx, userID, teamID)
}

func (m *MockRoleRepository) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	ret := m.ctrl.Call(m, "GetUserPermissions", ctx, userID)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockRoleRepositoryMockRecorder) GetUserPermissions(ctx, userID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserPermissions", reflect.TypeOf((*MockRoleRepository)(nil).GetUserPermissions), ctx, userID)
}

func (m *MockRoleRepository) GetUserPermissionsByTeam(ctx context.Context, userID, teamID string) ([]string, error) {
	ret := m.ctrl.Call(m, "GetUserPermissionsByTeam", ctx, userID, teamID)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockRoleRepositoryMockRecorder) GetUserPermissionsByTeam(ctx, userID, teamID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserPermissionsByTeam", reflect.TypeOf((*MockRoleRepository)(nil).GetUserPermissionsByTeam), ctx, userID, teamID)
}

func (m *MockRoleRepository) HasPermission(ctx context.Context, userID, resource, action string) (bool, error) {
	ret := m.ctrl.Call(m, "HasPermission", ctx, userID, resource, action)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockRoleRepositoryMockRecorder) HasPermission(ctx, userID, resource, action interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasPermission", reflect.TypeOf((*MockRoleRepository)(nil).HasPermission), ctx, userID, resource, action)
}

func (m *MockRoleRepository) HasTeamPermission(ctx context.Context, userID, teamID, resource, action string) (bool, error) {
	ret := m.ctrl.Call(m, "HasTeamPermission", ctx, userID, teamID, resource, action)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockRoleRepositoryMockRecorder) HasTeamPermission(ctx, userID, teamID, resource, action interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasTeamPermission", reflect.TypeOf((*MockRoleRepository)(nil).HasTeamPermission), ctx, userID, teamID, resource, action)
}
