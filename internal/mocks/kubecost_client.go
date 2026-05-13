package mocks

import (
	"context"
	"reflect"

	"github.com/davidsugianto/idp-core/internal/pkg/kubecost"
	"github.com/golang/mock/gomock"
)

// MockKubecostClient is a mock implementation of the Kubecost client
type MockKubecostClient struct {
	ctrl     *gomock.Controller
	recorder *MockKubecostClientMockRecorder
}

type MockKubecostClientMockRecorder struct {
	mock *MockKubecostClient
}

func NewMockKubecostClient(ctrl *gomock.Controller) *MockKubecostClient {
	mock := &MockKubecostClient{ctrl: ctrl}
	mock.recorder = &MockKubecostClientMockRecorder{mock}
	return mock
}

func (m *MockKubecostClient) EXPECT() *MockKubecostClientMockRecorder {
	return m.recorder
}

func (m *MockKubecostClient) GetAllocation(ctx context.Context, req kubecost.AllocationRequest) (*kubecost.AllocationResponse, error) {
	ret := m.ctrl.Call(m, "GetAllocation", ctx, req)
	ret0, _ := ret[0].(*kubecost.AllocationResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockKubecostClientMockRecorder) GetAllocation(ctx, req any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllocation", reflect.TypeOf((*MockKubecostClient)(nil).GetAllocation), ctx, req)
}
