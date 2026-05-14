package mocks

import (
	"context"
	"reflect"

	"github.com/golang/mock/gomock"
)

type MockSlackNotifier struct {
	ctrl     *gomock.Controller
	recorder *MockSlackNotifierMockRecorder
}

type MockSlackNotifierMockRecorder struct {
	mock *MockSlackNotifier
}

func NewMockSlackNotifier(ctrl *gomock.Controller) *MockSlackNotifier {
	mock := &MockSlackNotifier{ctrl: ctrl}
	mock.recorder = &MockSlackNotifierMockRecorder{mock}
	return mock
}

func (m *MockSlackNotifier) EXPECT() *MockSlackNotifierMockRecorder {
	return m.recorder
}

func (m *MockSlackNotifier) SendAlert(ctx context.Context, channel string, title string, fields map[string]string) error {
	ret := m.ctrl.Call(m, "SendAlert", ctx, channel, title, fields)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockSlackNotifierMockRecorder) SendAlert(ctx, channel, title, fields any) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendAlert", reflect.TypeOf((*MockSlackNotifier)(nil).SendAlert), ctx, channel, title, fields)
}

func (m *MockSlackNotifier) Channel() string {
	ret := m.ctrl.Call(m, "Channel")
	ret0, _ := ret[0].(string)
	return ret0
}

func (mr *MockSlackNotifierMockRecorder) Channel() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Channel", reflect.TypeOf((*MockSlackNotifier)(nil).Channel))
}