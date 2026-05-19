package mocks

import (
	"context"
	"reflect"

	endpointModel "github.com/davidsugianto/idp-core/internal/model/service_endpoint"
	serviceModel "github.com/davidsugianto/idp-core/internal/model/service"
	depModel "github.com/davidsugianto/idp-core/internal/model/service_dependency"
	svcEnvModel "github.com/davidsugianto/idp-core/internal/model/service_environment"
	versionModel "github.com/davidsugianto/idp-core/internal/model/service_version"
	"github.com/golang/mock/gomock"
)

type MockServiceRepository struct {
	ctrl     *gomock.Controller
	recorder *MockServiceRepositoryMockRecorder
}

type MockServiceRepositoryMockRecorder struct {
	mock *MockServiceRepository
}

func NewMockServiceRepository(ctrl *gomock.Controller) *MockServiceRepository {
	mock := &MockServiceRepository{ctrl: ctrl}
	mock.recorder = &MockServiceRepositoryMockRecorder{mock}
	return mock
}

func (m *MockServiceRepository) EXPECT() *MockServiceRepositoryMockRecorder {
	return m.recorder
}

func (m *MockServiceRepository) Create(ctx context.Context, svc *serviceModel.Service) error {
	ret := m.ctrl.Call(m, "Create", ctx, svc)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockServiceRepositoryMockRecorder) Create(ctx, svc interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockServiceRepository)(nil).Create), ctx, svc)
}

func (m *MockServiceRepository) GetByID(ctx context.Context, id string) (*serviceModel.Service, error) {
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*serviceModel.Service)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockServiceRepositoryMockRecorder) GetByID(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockServiceRepository)(nil).GetByID), ctx, id)
}

func (m *MockServiceRepository) GetByIDIncludingDeleted(ctx context.Context, id string) (*serviceModel.Service, error) {
	ret := m.ctrl.Call(m, "GetByIDIncludingDeleted", ctx, id)
	ret0, _ := ret[0].(*serviceModel.Service)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockServiceRepositoryMockRecorder) GetByIDIncludingDeleted(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByIDIncludingDeleted", reflect.TypeOf((*MockServiceRepository)(nil).GetByIDIncludingDeleted), ctx, id)
}

func (m *MockServiceRepository) List(ctx context.Context, req *serviceModel.ListServicesRequest) ([]serviceModel.Service, int64, error) {
	ret := m.ctrl.Call(m, "List", ctx, req)
	ret0, _ := ret[0].([]serviceModel.Service)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockServiceRepositoryMockRecorder) List(ctx, req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockServiceRepository)(nil).List), ctx, req)
}

func (m *MockServiceRepository) Update(ctx context.Context, svc *serviceModel.Service) error {
	ret := m.ctrl.Call(m, "Update", ctx, svc)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockServiceRepositoryMockRecorder) Update(ctx, svc interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockServiceRepository)(nil).Update), ctx, svc)
}

func (m *MockServiceRepository) Delete(ctx context.Context, id string) error {
	ret := m.ctrl.Call(m, "Delete", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockServiceRepositoryMockRecorder) Delete(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockServiceRepository)(nil).Delete), ctx, id)
}

func (m *MockServiceRepository) CreateVersion(ctx context.Context, v *versionModel.ServiceVersion) error {
	ret := m.ctrl.Call(m, "CreateVersion", ctx, v)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockServiceRepositoryMockRecorder) CreateVersion(ctx, v interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateVersion", reflect.TypeOf((*MockServiceRepository)(nil).CreateVersion), ctx, v)
}

func (m *MockServiceRepository) GetVersionByID(ctx context.Context, id string) (*versionModel.ServiceVersion, error) {
	ret := m.ctrl.Call(m, "GetVersionByID", ctx, id)
	ret0, _ := ret[0].(*versionModel.ServiceVersion)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockServiceRepositoryMockRecorder) GetVersionByID(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVersionByID", reflect.TypeOf((*MockServiceRepository)(nil).GetVersionByID), ctx, id)
}

func (m *MockServiceRepository) GetVersionByServiceAndVersion(ctx context.Context, serviceID, version string) (*versionModel.ServiceVersion, error) {
	ret := m.ctrl.Call(m, "GetVersionByServiceAndVersion", ctx, serviceID, version)
	ret0, _ := ret[0].(*versionModel.ServiceVersion)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockServiceRepositoryMockRecorder) GetVersionByServiceAndVersion(ctx, serviceID, version interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVersionByServiceAndVersion", reflect.TypeOf((*MockServiceRepository)(nil).GetVersionByServiceAndVersion), ctx, serviceID, version)
}

func (m *MockServiceRepository) ListVersionsByService(ctx context.Context, serviceID string, req *versionModel.ListServiceVersionsRequest) ([]versionModel.ServiceVersion, int64, error) {
	ret := m.ctrl.Call(m, "ListVersionsByService", ctx, serviceID, req)
	ret0, _ := ret[0].([]versionModel.ServiceVersion)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockServiceRepositoryMockRecorder) ListVersionsByService(ctx, serviceID, req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListVersionsByService", reflect.TypeOf((*MockServiceRepository)(nil).ListVersionsByService), ctx, serviceID, req)
}

func (m *MockServiceRepository) UpdateVersion(ctx context.Context, v *versionModel.ServiceVersion) error {
	ret := m.ctrl.Call(m, "UpdateVersion", ctx, v)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockServiceRepositoryMockRecorder) UpdateVersion(ctx, v interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateVersion", reflect.TypeOf((*MockServiceRepository)(nil).UpdateVersion), ctx, v)
}

func (m *MockServiceRepository) GetActiveVersionsByService(ctx context.Context, serviceID string) ([]versionModel.ServiceVersion, error) {
	ret := m.ctrl.Call(m, "GetActiveVersionsByService", ctx, serviceID)
	ret0, _ := ret[0].([]versionModel.ServiceVersion)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockServiceRepositoryMockRecorder) GetActiveVersionsByService(ctx, serviceID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetActiveVersionsByService", reflect.TypeOf((*MockServiceRepository)(nil).GetActiveVersionsByService), ctx, serviceID)
}

func (m *MockServiceRepository) CreateEndpoint(ctx context.Context, ep *endpointModel.ServiceEndpoint) error {
	ret := m.ctrl.Call(m, "CreateEndpoint", ctx, ep)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockServiceRepositoryMockRecorder) CreateEndpoint(ctx, ep interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateEndpoint", reflect.TypeOf((*MockServiceRepository)(nil).CreateEndpoint), ctx, ep)
}

func (m *MockServiceRepository) GetEndpointByID(ctx context.Context, id string) (*endpointModel.ServiceEndpoint, error) {
	ret := m.ctrl.Call(m, "GetEndpointByID", ctx, id)
	ret0, _ := ret[0].(*endpointModel.ServiceEndpoint)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockServiceRepositoryMockRecorder) GetEndpointByID(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEndpointByID", reflect.TypeOf((*MockServiceRepository)(nil).GetEndpointByID), ctx, id)
}

func (m *MockServiceRepository) ListEndpointsByVersion(ctx context.Context, versionID string) ([]endpointModel.ServiceEndpoint, error) {
	ret := m.ctrl.Call(m, "ListEndpointsByVersion", ctx, versionID)
	ret0, _ := ret[0].([]endpointModel.ServiceEndpoint)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockServiceRepositoryMockRecorder) ListEndpointsByVersion(ctx, versionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEndpointsByVersion", reflect.TypeOf((*MockServiceRepository)(nil).ListEndpointsByVersion), ctx, versionID)
}

func (m *MockServiceRepository) UpdateEndpoint(ctx context.Context, ep *endpointModel.ServiceEndpoint) error {
	ret := m.ctrl.Call(m, "UpdateEndpoint", ctx, ep)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockServiceRepositoryMockRecorder) UpdateEndpoint(ctx, ep interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateEndpoint", reflect.TypeOf((*MockServiceRepository)(nil).UpdateEndpoint), ctx, ep)
}

func (m *MockServiceRepository) DeleteEndpoint(ctx context.Context, id string) error {
	ret := m.ctrl.Call(m, "DeleteEndpoint", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockServiceRepositoryMockRecorder) DeleteEndpoint(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteEndpoint", reflect.TypeOf((*MockServiceRepository)(nil).DeleteEndpoint), ctx, id)
}

func (m *MockServiceRepository) SearchServices(ctx context.Context, query string) ([]serviceModel.Service, error) {
	ret := m.ctrl.Call(m, "SearchServices", ctx, query)
	ret0, _ := ret[0].([]serviceModel.Service)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockServiceRepositoryMockRecorder) SearchServices(ctx, query interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchServices", reflect.TypeOf((*MockServiceRepository)(nil).SearchServices), ctx, query)
}

func (m *MockServiceRepository) ListEndpointsByType(ctx context.Context, endpointType string) ([]endpointModel.ServiceEndpoint, error) {
	ret := m.ctrl.Call(m, "ListEndpointsByType", ctx, endpointType)
	ret0, _ := ret[0].([]endpointModel.ServiceEndpoint)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockServiceRepositoryMockRecorder) ListEndpointsByType(ctx, endpointType interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEndpointsByType", reflect.TypeOf((*MockServiceRepository)(nil).ListEndpointsByType), ctx, endpointType)
}

func (m *MockServiceRepository) ExistsByName(ctx context.Context, name string, teamID string) (bool, error) {
	ret := m.ctrl.Call(m, "ExistsByName", ctx, name, teamID)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockServiceRepositoryMockRecorder) ExistsByName(ctx, name, teamID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExistsByName", reflect.TypeOf((*MockServiceRepository)(nil).ExistsByName), ctx, name, teamID)
}

// Dependency methods
func (m *MockServiceRepository) CreateDependency(ctx context.Context, dep *depModel.ServiceDependency) error {
	ret := m.ctrl.Call(m, "CreateDependency", ctx, dep)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockServiceRepositoryMockRecorder) CreateDependency(ctx, dep interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateDependency", reflect.TypeOf((*MockServiceRepository)(nil).CreateDependency), ctx, dep)
}

func (m *MockServiceRepository) GetDependencyByID(ctx context.Context, id string) (*depModel.ServiceDependency, error) {
	ret := m.ctrl.Call(m, "GetDependencyByID", ctx, id)
	ret0, _ := ret[0].(*depModel.ServiceDependency)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockServiceRepositoryMockRecorder) GetDependencyByID(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDependencyByID", reflect.TypeOf((*MockServiceRepository)(nil).GetDependencyByID), ctx, id)
}

func (m *MockServiceRepository) ListDependenciesByService(ctx context.Context, serviceID string, req *depModel.ListDependenciesRequest) ([]depModel.ServiceDependency, int64, error) {
	ret := m.ctrl.Call(m, "ListDependenciesByService", ctx, serviceID, req)
	ret0, _ := ret[0].([]depModel.ServiceDependency)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockServiceRepositoryMockRecorder) ListDependenciesByService(ctx, serviceID, req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListDependenciesByService", reflect.TypeOf((*MockServiceRepository)(nil).ListDependenciesByService), ctx, serviceID, req)
}

func (m *MockServiceRepository) ListDependentsByService(ctx context.Context, serviceID string, req *depModel.ListDependenciesRequest) ([]depModel.ServiceDependency, int64, error) {
	ret := m.ctrl.Call(m, "ListDependentsByService", ctx, serviceID, req)
	ret0, _ := ret[0].([]depModel.ServiceDependency)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockServiceRepositoryMockRecorder) ListDependentsByService(ctx, serviceID, req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListDependentsByService", reflect.TypeOf((*MockServiceRepository)(nil).ListDependentsByService), ctx, serviceID, req)
}

func (m *MockServiceRepository) UpdateDependency(ctx context.Context, dep *depModel.ServiceDependency) error {
	ret := m.ctrl.Call(m, "UpdateDependency", ctx, dep)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockServiceRepositoryMockRecorder) UpdateDependency(ctx, dep interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateDependency", reflect.TypeOf((*MockServiceRepository)(nil).UpdateDependency), ctx, dep)
}

func (m *MockServiceRepository) DeleteDependency(ctx context.Context, id string) error {
	ret := m.ctrl.Call(m, "DeleteDependency", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockServiceRepositoryMockRecorder) DeleteDependency(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteDependency", reflect.TypeOf((*MockServiceRepository)(nil).DeleteDependency), ctx, id)
}

func (m *MockServiceRepository) ExistsDependency(ctx context.Context, serviceID, dependsOnServiceID string) (bool, error) {
	ret := m.ctrl.Call(m, "ExistsDependency", ctx, serviceID, dependsOnServiceID)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockServiceRepositoryMockRecorder) ExistsDependency(ctx, serviceID, dependsOnServiceID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExistsDependency", reflect.TypeOf((*MockServiceRepository)(nil).ExistsDependency), ctx, serviceID, dependsOnServiceID)
}

func (m *MockServiceRepository) ListAllDependencies(ctx context.Context) ([]depModel.ServiceDependency, error) {
	ret := m.ctrl.Call(m, "ListAllDependencies", ctx)
	ret0, _ := ret[0].([]depModel.ServiceDependency)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockServiceRepositoryMockRecorder) ListAllDependencies(ctx interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAllDependencies", reflect.TypeOf((*MockServiceRepository)(nil).ListAllDependencies), ctx)
}

// Service environment methods
func (m *MockServiceRepository) CreateServiceEnvironment(ctx context.Context, se *svcEnvModel.ServiceEnvironment) error {
	ret := m.ctrl.Call(m, "CreateServiceEnvironment", ctx, se)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockServiceRepositoryMockRecorder) CreateServiceEnvironment(ctx, se interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateServiceEnvironment", reflect.TypeOf((*MockServiceRepository)(nil).CreateServiceEnvironment), ctx, se)
}

func (m *MockServiceRepository) GetServiceEnvironmentByID(ctx context.Context, id string) (*svcEnvModel.ServiceEnvironment, error) {
	ret := m.ctrl.Call(m, "GetServiceEnvironmentByID", ctx, id)
	ret0, _ := ret[0].(*svcEnvModel.ServiceEnvironment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockServiceRepositoryMockRecorder) GetServiceEnvironmentByID(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetServiceEnvironmentByID", reflect.TypeOf((*MockServiceRepository)(nil).GetServiceEnvironmentByID), ctx, id)
}

func (m *MockServiceRepository) ListServiceEnvironmentsByVersion(ctx context.Context, versionID string, req *svcEnvModel.ListDeploymentsRequest) ([]svcEnvModel.ServiceEnvironment, int64, error) {
	ret := m.ctrl.Call(m, "ListServiceEnvironmentsByVersion", ctx, versionID, req)
	ret0, _ := ret[0].([]svcEnvModel.ServiceEnvironment)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockServiceRepositoryMockRecorder) ListServiceEnvironmentsByVersion(ctx, versionID, req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListServiceEnvironmentsByVersion", reflect.TypeOf((*MockServiceRepository)(nil).ListServiceEnvironmentsByVersion), ctx, versionID, req)
}

func (m *MockServiceRepository) ListServiceEnvironmentsByService(ctx context.Context, serviceID string, req *svcEnvModel.ListDeploymentsRequest) ([]svcEnvModel.ServiceEnvironment, int64, error) {
	ret := m.ctrl.Call(m, "ListServiceEnvironmentsByService", ctx, serviceID, req)
	ret0, _ := ret[0].([]svcEnvModel.ServiceEnvironment)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockServiceRepositoryMockRecorder) ListServiceEnvironmentsByService(ctx, serviceID, req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListServiceEnvironmentsByService", reflect.TypeOf((*MockServiceRepository)(nil).ListServiceEnvironmentsByService), ctx, serviceID, req)
}

func (m *MockServiceRepository) ListServiceEnvironmentsByEnvironment(ctx context.Context, environmentID string, req *svcEnvModel.ListDeploymentsRequest) ([]svcEnvModel.ServiceEnvironment, int64, error) {
	ret := m.ctrl.Call(m, "ListServiceEnvironmentsByEnvironment", ctx, environmentID, req)
	ret0, _ := ret[0].([]svcEnvModel.ServiceEnvironment)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockServiceRepositoryMockRecorder) ListServiceEnvironmentsByEnvironment(ctx, environmentID, req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListServiceEnvironmentsByEnvironment", reflect.TypeOf((*MockServiceRepository)(nil).ListServiceEnvironmentsByEnvironment), ctx, environmentID, req)
}

func (m *MockServiceRepository) UpdateServiceEnvironment(ctx context.Context, se *svcEnvModel.ServiceEnvironment) error {
	ret := m.ctrl.Call(m, "UpdateServiceEnvironment", ctx, se)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockServiceRepositoryMockRecorder) UpdateServiceEnvironment(ctx, se interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateServiceEnvironment", reflect.TypeOf((*MockServiceRepository)(nil).UpdateServiceEnvironment), ctx, se)
}

func (m *MockServiceRepository) DeleteServiceEnvironment(ctx context.Context, id string) error {
	ret := m.ctrl.Call(m, "DeleteServiceEnvironment", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockServiceRepositoryMockRecorder) DeleteServiceEnvironment(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteServiceEnvironment", reflect.TypeOf((*MockServiceRepository)(nil).DeleteServiceEnvironment), ctx, id)
}

func (m *MockServiceRepository) GetActiveDeployment(ctx context.Context, versionID, environmentID string) (*svcEnvModel.ServiceEnvironment, error) {
	ret := m.ctrl.Call(m, "GetActiveDeployment", ctx, versionID, environmentID)
	ret0, _ := ret[0].(*svcEnvModel.ServiceEnvironment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockServiceRepositoryMockRecorder) GetActiveDeployment(ctx, versionID, environmentID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetActiveDeployment", reflect.TypeOf((*MockServiceRepository)(nil).GetActiveDeployment), ctx, versionID, environmentID)
}
