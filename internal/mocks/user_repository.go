package mocks

import (
	"context"
	"reflect"

	"github.com/davidsugianto/idp-core/internal/model/user"
	"github.com/golang/mock/gomock"
)

// MockUserRepository is a mock implementation of the User repository
type MockUserRepository struct {
	ctrl     *gomock.Controller
	recorder *MockUserRepositoryMockRecorder
}

type MockUserRepositoryMockRecorder struct {
	mock *MockUserRepository
}

func NewMockUserRepository(ctrl *gomock.Controller) *MockUserRepository {
	mock := &MockUserRepository{ctrl: ctrl}
	mock.recorder = &MockUserRepositoryMockRecorder{mock}
	return mock
}

func (m *MockUserRepository) EXPECT() *MockUserRepositoryMockRecorder {
	return m.recorder
}

func (m *MockUserRepository) Create(ctx context.Context, u *user.User) error {
	ret := m.ctrl.Call(m, "Create", ctx, u)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUserRepositoryMockRecorder) Create(ctx, u interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockUserRepository)(nil).Create), ctx, u)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*user.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserRepositoryMockRecorder) GetByID(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockUserRepository)(nil).GetByID), ctx, id)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	ret := m.ctrl.Call(m, "GetByEmail", ctx, email)
	ret0, _ := ret[0].(*user.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserRepositoryMockRecorder) GetByEmail(ctx, email interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByEmail", reflect.TypeOf((*MockUserRepository)(nil).GetByEmail), ctx, email)
}

func (m *MockUserRepository) GetByProviderID(ctx context.Context, provider, providerID string) (*user.User, error) {
	ret := m.ctrl.Call(m, "GetByProviderID", ctx, provider, providerID)
	ret0, _ := ret[0].(*user.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserRepositoryMockRecorder) GetByProviderID(ctx, provider, providerID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByProviderID", reflect.TypeOf((*MockUserRepository)(nil).GetByProviderID), ctx, provider, providerID)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]user.User, int64, error) {
	ret := m.ctrl.Call(m, "List", ctx, limit, offset)
	ret0, _ := ret[0].([]user.User)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockUserRepositoryMockRecorder) List(ctx, limit, offset interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockUserRepository)(nil).List), ctx, limit, offset)
}

func (m *MockUserRepository) ListByStatus(ctx context.Context, status string) ([]user.User, error) {
	ret := m.ctrl.Call(m, "ListByStatus", ctx, status)
	ret0, _ := ret[0].([]user.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserRepositoryMockRecorder) ListByStatus(ctx, status interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByStatus", reflect.TypeOf((*MockUserRepository)(nil).ListByStatus), ctx, status)
}

func (m *MockUserRepository) Update(ctx context.Context, u *user.User) error {
	ret := m.ctrl.Call(m, "Update", ctx, u)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUserRepositoryMockRecorder) Update(ctx, u interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockUserRepository)(nil).Update), ctx, u)
}

func (m *MockUserRepository) UpdateStatus(ctx context.Context, id, status string) error {
	ret := m.ctrl.Call(m, "UpdateStatus", ctx, id, status)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUserRepositoryMockRecorder) UpdateStatus(ctx, id, status interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateStatus", reflect.TypeOf((*MockUserRepository)(nil).UpdateStatus), ctx, id, status)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	ret := m.ctrl.Call(m, "UpdateLastLogin", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUserRepositoryMockRecorder) UpdateLastLogin(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateLastLogin", reflect.TypeOf((*MockUserRepository)(nil).UpdateLastLogin), ctx, id)
}

func (m *MockUserRepository) SoftDelete(ctx context.Context, id string) error {
	ret := m.ctrl.Call(m, "SoftDelete", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUserRepositoryMockRecorder) SoftDelete(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SoftDelete", reflect.TypeOf((*MockUserRepository)(nil).SoftDelete), ctx, id)
}

func (m *MockUserRepository) HardDelete(ctx context.Context, id string) error {
	ret := m.ctrl.Call(m, "HardDelete", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUserRepositoryMockRecorder) HardDelete(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HardDelete", reflect.TypeOf((*MockUserRepository)(nil).HardDelete), ctx, id)
}
