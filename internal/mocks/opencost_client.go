package mocks

import (
	"context"
	"reflect"

	"github.com/davidsugianto/idp-core/internal/pkg/opencost"
	"github.com/golang/mock/gomock"
)

// MockOpenCostClient is a mock implementation of the OpenCost client
type MockOpenCostClient struct {
	ctrl     *gomock.Controller
	recorder *MockOpenCostClientMockRecorder
}

type MockOpenCostClientMockRecorder struct {
	mock *MockOpenCostClient
}

func NewMockOpenCostClient(ctrl *gomock.Controller) *MockOpenCostClient {
	mock := &MockOpenCostClient{ctrl: ctrl}
	mock.recorder = &MockOpenCostClientMockRecorder{mock}
	return mock
}

func (m *MockOpenCostClient) EXPECT() *MockOpenCostClientMockRecorder {
	return m.recorder
}

func (m *MockOpenCostClient) GetAllocation(ctx context.Context, req opencost.AllocationRequest) (*opencost.AllocationResponse, error) {
	ret := m.ctrl.Call(m, "GetAllocation", ctx, req)
	ret0, _ := ret[0].(*opencost.AllocationResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockOpenCostClientMockRecorder) GetAllocation(ctx, req any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllocation", reflect.TypeOf((*MockOpenCostClient)(nil).GetAllocation), ctx, req)
}