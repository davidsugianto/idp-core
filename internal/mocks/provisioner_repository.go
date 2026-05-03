package mocks

import (
	"context"
	"reflect"

	"github.com/davidsugianto/idp-core/internal/model/environment"
	"github.com/golang/mock/gomock"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// MockProvisionerRepository is a mock implementation of the Provisioner repository
type MockProvisionerRepository struct {
	ctrl     *gomock.Controller
	recorder *MockProvisionerRepositoryMockRecorder
}

type MockProvisionerRepositoryMockRecorder struct {
	mock *MockProvisionerRepository
}

func NewMockProvisionerRepository(ctrl *gomock.Controller) *MockProvisionerRepository {
	mock := &MockProvisionerRepository{ctrl: ctrl}
	mock.recorder = &MockProvisionerRepositoryMockRecorder{mock}
	return mock
}

func (m *MockProvisionerRepository) EXPECT() *MockProvisionerRepositoryMockRecorder {
	return m.recorder
}

func (m *MockProvisionerRepository) CreateNamespace(ctx context.Context, name string, labels map[string]string) error {
	ret := m.ctrl.Call(m, "CreateNamespace", ctx, name, labels)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockProvisionerRepositoryMockRecorder) CreateNamespace(ctx, name, labels interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateNamespace", reflect.TypeOf((*MockProvisionerRepository)(nil).CreateNamespace), ctx, name, labels)
}

func (m *MockProvisionerRepository) DeleteNamespace(ctx context.Context, name string) error {
	ret := m.ctrl.Call(m, "DeleteNamespace", ctx, name)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockProvisionerRepositoryMockRecorder) DeleteNamespace(ctx, name interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteNamespace", reflect.TypeOf((*MockProvisionerRepository)(nil).DeleteNamespace), ctx, name)
}

func (m *MockProvisionerRepository) GetNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	ret := m.ctrl.Call(m, "GetNamespace", ctx, name)
	ret0, _ := ret[0].(*corev1.Namespace)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockProvisionerRepositoryMockRecorder) GetNamespace(ctx, name interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNamespace", reflect.TypeOf((*MockProvisionerRepository)(nil).GetNamespace), ctx, name)
}

func (m *MockProvisionerRepository) NamespaceExists(ctx context.Context, name string) (bool, error) {
	ret := m.ctrl.Call(m, "NamespaceExists", ctx, name)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockProvisionerRepositoryMockRecorder) NamespaceExists(ctx, name interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NamespaceExists", reflect.TypeOf((*MockProvisionerRepository)(nil).NamespaceExists), ctx, name)
}

func (m *MockProvisionerRepository) CreateResourceQuota(ctx context.Context, namespace, name, cpu, memory string) error {
	ret := m.ctrl.Call(m, "CreateResourceQuota", ctx, namespace, name, cpu, memory)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockProvisionerRepositoryMockRecorder) CreateResourceQuota(ctx, namespace, name, cpu, memory interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateResourceQuota", reflect.TypeOf((*MockProvisionerRepository)(nil).CreateResourceQuota), ctx, namespace, name, cpu, memory)
}

func (m *MockProvisionerRepository) DeleteResourceQuota(ctx context.Context, namespace, name string) error {
	ret := m.ctrl.Call(m, "DeleteResourceQuota", ctx, namespace, name)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockProvisionerRepositoryMockRecorder) DeleteResourceQuota(ctx, namespace, name interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteResourceQuota", reflect.TypeOf((*MockProvisionerRepository)(nil).DeleteResourceQuota), ctx, namespace, name)
}

func (m *MockProvisionerRepository) CreateNetworkPolicy(ctx context.Context, namespace, name string, allowNamespaceLabels map[string]string) error {
	ret := m.ctrl.Call(m, "CreateNetworkPolicy", ctx, namespace, name, allowNamespaceLabels)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockProvisionerRepositoryMockRecorder) CreateNetworkPolicy(ctx, namespace, name, allowNamespaceLabels interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateNetworkPolicy", reflect.TypeOf((*MockProvisionerRepository)(nil).CreateNetworkPolicy), ctx, namespace, name, allowNamespaceLabels)
}

func (m *MockProvisionerRepository) DeleteNetworkPolicy(ctx context.Context, namespace, name string) error {
	ret := m.ctrl.Call(m, "DeleteNetworkPolicy", ctx, namespace, name)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockProvisionerRepositoryMockRecorder) DeleteNetworkPolicy(ctx, namespace, name interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteNetworkPolicy", reflect.TypeOf((*MockProvisionerRepository)(nil).DeleteNetworkPolicy), ctx, namespace, name)
}

func (m *MockProvisionerRepository) GetPodSummary(namespace string) (environment.PodSummary, bool) {
	ret := m.ctrl.Call(m, "GetPodSummary", namespace)
	ret0, _ := ret[0].(environment.PodSummary)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

func (mr *MockProvisionerRepositoryMockRecorder) GetPodSummary(namespace interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPodSummary", reflect.TypeOf((*MockProvisionerRepository)(nil).GetPodSummary), namespace)
}

func (m *MockProvisionerRepository) GetDeploymentSummary(namespace string) (environment.DeploymentSummary, bool) {
	ret := m.ctrl.Call(m, "GetDeploymentSummary", namespace)
	ret0, _ := ret[0].(environment.DeploymentSummary)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

func (mr *MockProvisionerRepositoryMockRecorder) GetDeploymentSummary(namespace interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDeploymentSummary", reflect.TypeOf((*MockProvisionerRepository)(nil).GetDeploymentSummary), namespace)
}

func (m *MockProvisionerRepository) GetWorkloads(namespace string) ([]*appsv1.Deployment, error) {
	ret := m.ctrl.Call(m, "GetWorkloads", namespace)
	ret0, _ := ret[0].([]*appsv1.Deployment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockProvisionerRepositoryMockRecorder) GetWorkloads(namespace interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWorkloads", reflect.TypeOf((*MockProvisionerRepository)(nil).GetWorkloads), namespace)
}

func (m *MockProvisionerRepository) GetPods(namespace string) ([]*corev1.Pod, error) {
	ret := m.ctrl.Call(m, "GetPods", namespace)
	ret0, _ := ret[0].([]*corev1.Pod)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockProvisionerRepositoryMockRecorder) GetPods(namespace interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPods", reflect.TypeOf((*MockProvisionerRepository)(nil).GetPods), namespace)
}

func (m *MockProvisionerRepository) StartInformers(ctx context.Context) error {
	ret := m.ctrl.Call(m, "StartInformers", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockProvisionerRepositoryMockRecorder) StartInformers(ctx interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartInformers", reflect.TypeOf((*MockProvisionerRepository)(nil).StartInformers), ctx)
}

func (m *MockProvisionerRepository) StopInformers() {
	m.ctrl.Call(m, "StopInformers")
}

func (mr *MockProvisionerRepositoryMockRecorder) StopInformers() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StopInformers", reflect.TypeOf((*MockProvisionerRepository)(nil).StopInformers))
}
