package mocks

import (
	"context"
	"reflect"

	"github.com/davidsugianto/idp-core/internal/model/environment"
	"github.com/davidsugianto/idp-core/internal/pkg/argocd"
	"github.com/golang/mock/gomock"
)

// MockGitopsRepository is a mock implementation of the Gitops repository
type MockGitopsRepository struct {
	ctrl     *gomock.Controller
	recorder *MockGitopsRepositoryMockRecorder
}

type MockGitopsRepositoryMockRecorder struct {
	mock *MockGitopsRepository
}

func NewMockGitopsRepository(ctrl *gomock.Controller) *MockGitopsRepository {
	mock := &MockGitopsRepository{ctrl: ctrl}
	mock.recorder = &MockGitopsRepositoryMockRecorder{mock}
	return mock
}

func (m *MockGitopsRepository) EXPECT() *MockGitopsRepositoryMockRecorder {
	return m.recorder
}

func (m *MockGitopsRepository) CreateApplication(ctx context.Context, spec argocd.ApplicationSpec) error {
	ret := m.ctrl.Call(m, "CreateApplication", ctx, spec)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockGitopsRepositoryMockRecorder) CreateApplication(ctx, spec interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateApplication", reflect.TypeOf((*MockGitopsRepository)(nil).CreateApplication), ctx, spec)
}

func (m *MockGitopsRepository) GetApplicationStatus(ctx context.Context, name string) (*environment.ArgoStatus, error) {
	ret := m.ctrl.Call(m, "GetApplicationStatus", ctx, name)
	ret0, _ := ret[0].(*environment.ArgoStatus)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockGitopsRepositoryMockRecorder) GetApplicationStatus(ctx, name interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApplicationStatus", reflect.TypeOf((*MockGitopsRepository)(nil).GetApplicationStatus), ctx, name)
}

func (m *MockGitopsRepository) SyncApplication(ctx context.Context, name string) error {
	ret := m.ctrl.Call(m, "SyncApplication", ctx, name)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockGitopsRepositoryMockRecorder) SyncApplication(ctx, name interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SyncApplication", reflect.TypeOf((*MockGitopsRepository)(nil).SyncApplication), ctx, name)
}

func (m *MockGitopsRepository) DeleteApplication(ctx context.Context, name string) error {
	ret := m.ctrl.Call(m, "DeleteApplication", ctx, name)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockGitopsRepositoryMockRecorder) DeleteApplication(ctx, name interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteApplication", reflect.TypeOf((*MockGitopsRepository)(nil).DeleteApplication), ctx, name)
}
