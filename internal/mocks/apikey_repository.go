package mocks

import (
	"context"
	"reflect"

	"github.com/davidsugianto/idp-core/internal/model/apikey"
	"github.com/golang/mock/gomock"
)

// MockApiKeyRepository is a mock implementation of the API key repository
type MockApiKeyRepository struct {
	ctrl     *gomock.Controller
	recorder *MockApiKeyRepositoryMockRecorder
}

type MockApiKeyRepositoryMockRecorder struct {
	mock *MockApiKeyRepository
}

func NewMockApiKeyRepository(ctrl *gomock.Controller) *MockApiKeyRepository {
	mock := &MockApiKeyRepository{ctrl: ctrl}
	mock.recorder = &MockApiKeyRepositoryMockRecorder{mock}
	return mock
}

func (m *MockApiKeyRepository) EXPECT() *MockApiKeyRepositoryMockRecorder {
	return m.recorder
}

func (m *MockApiKeyRepository) Create(ctx context.Context, key *apikey.APIKey) error {
	ret := m.ctrl.Call(m, "Create", ctx, key)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockApiKeyRepositoryMockRecorder) Create(ctx, key interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockApiKeyRepository)(nil).Create), ctx, key)
}

func (m *MockApiKeyRepository) GetByID(ctx context.Context, id string) (*apikey.APIKey, error) {
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*apikey.APIKey)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockApiKeyRepositoryMockRecorder) GetByID(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockApiKeyRepository)(nil).GetByID), ctx, id)
}

func (m *MockApiKeyRepository) GetByKey(ctx context.Context, key string) (*apikey.APIKey, error) {
	ret := m.ctrl.Call(m, "GetByKey", ctx, key)
	ret0, _ := ret[0].(*apikey.APIKey)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockApiKeyRepositoryMockRecorder) GetByKey(ctx, key interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByKey", reflect.TypeOf((*MockApiKeyRepository)(nil).GetByKey), ctx, key)
}

func (m *MockApiKeyRepository) ListByTeam(ctx context.Context, teamID string) ([]apikey.APIKey, error) {
	ret := m.ctrl.Call(m, "ListByTeam", ctx, teamID)
	ret0, _ := ret[0].([]apikey.APIKey)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockApiKeyRepositoryMockRecorder) ListByTeam(ctx, teamID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByTeam", reflect.TypeOf((*MockApiKeyRepository)(nil).ListByTeam), ctx, teamID)
}

func (m *MockApiKeyRepository) ListActive(ctx context.Context) ([]apikey.APIKey, error) {
	ret := m.ctrl.Call(m, "ListActive", ctx)
	ret0, _ := ret[0].([]apikey.APIKey)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockApiKeyRepositoryMockRecorder) ListActive(ctx interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListActive", reflect.TypeOf((*MockApiKeyRepository)(nil).ListActive), ctx)
}

func (m *MockApiKeyRepository) Update(ctx context.Context, key *apikey.APIKey) error {
	ret := m.ctrl.Call(m, "Update", ctx, key)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockApiKeyRepositoryMockRecorder) Update(ctx, key interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockApiKeyRepository)(nil).Update), ctx, key)
}

func (m *MockApiKeyRepository) UpdateLastUsed(ctx context.Context, id string) error {
	ret := m.ctrl.Call(m, "UpdateLastUsed", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockApiKeyRepositoryMockRecorder) UpdateLastUsed(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateLastUsed", reflect.TypeOf((*MockApiKeyRepository)(nil).UpdateLastUsed), ctx, id)
}

func (m *MockApiKeyRepository) IncrementUsage(ctx context.Context, id string) error {
	ret := m.ctrl.Call(m, "IncrementUsage", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockApiKeyRepositoryMockRecorder) IncrementUsage(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IncrementUsage", reflect.TypeOf((*MockApiKeyRepository)(nil).IncrementUsage), ctx, id)
}

func (m *MockApiKeyRepository) Deactivate(ctx context.Context, id string) error {
	ret := m.ctrl.Call(m, "Deactivate", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockApiKeyRepositoryMockRecorder) Deactivate(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Deactivate", reflect.TypeOf((*MockApiKeyRepository)(nil).Deactivate), ctx, id)
}

func (m *MockApiKeyRepository) Delete(ctx context.Context, id string) error {
	ret := m.ctrl.Call(m, "Delete", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockApiKeyRepositoryMockRecorder) Delete(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockApiKeyRepository)(nil).Delete), ctx, id)
}

func (m *MockApiKeyRepository) DeleteExpired(ctx context.Context) error {
	ret := m.ctrl.Call(m, "DeleteExpired", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockApiKeyRepositoryMockRecorder) DeleteExpired(ctx interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteExpired", reflect.TypeOf((*MockApiKeyRepository)(nil).DeleteExpired), ctx)
}