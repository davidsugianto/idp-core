package mocks

import (
	"context"
	"reflect"
	"time"

	prometheusPkg "github.com/davidsugianto/idp-core/internal/pkg/prometheus"
	"github.com/golang/mock/gomock"
)

type MockMonitoringRepository struct {
	ctrl     *gomock.Controller
	recorder *MockMonitoringRepositoryMockRecorder
}

type MockMonitoringRepositoryMockRecorder struct {
	mock *MockMonitoringRepository
}

func NewMockMonitoringRepository(ctrl *gomock.Controller) *MockMonitoringRepository {
	mock := &MockMonitoringRepository{ctrl: ctrl}
	mock.recorder = &MockMonitoringRepositoryMockRecorder{mock}
	return mock
}

func (m *MockMonitoringRepository) EXPECT() *MockMonitoringRepositoryMockRecorder {
	return m.recorder
}

func (m *MockMonitoringRepository) Query(ctx context.Context, query string) ([]prometheusPkg.QueryResult, error) {
	ret := m.ctrl.Call(m, "Query", ctx, query)
	ret0, _ := ret[0].([]prometheusPkg.QueryResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockMonitoringRepositoryMockRecorder) Query(ctx, query interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockMonitoringRepository)(nil).Query), ctx, query)
}

func (m *MockMonitoringRepository) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) ([]prometheusPkg.RangeQueryResult, error) {
	ret := m.ctrl.Call(m, "QueryRange", ctx, query, start, end, step)
	ret0, _ := ret[0].([]prometheusPkg.RangeQueryResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockMonitoringRepositoryMockRecorder) QueryRange(ctx, query, start, end, step interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryRange", reflect.TypeOf((*MockMonitoringRepository)(nil).QueryRange), ctx, query, start, end, step)
}
