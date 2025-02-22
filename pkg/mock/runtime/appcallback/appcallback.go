// Code generated by MockGen. DO NOT EDIT.
// Source: spec/proto/runtime/v1/appcallback.pb.go

// Package mock_appcallback is a generated GoMock package.
package mock_appcallback

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	runtime "mosn.io/layotto/spec/proto/runtime/v1"
)

// MockAppCallbackClient is a mock of AppCallbackClient interface.
type MockAppCallbackClient struct {
	ctrl     *gomock.Controller
	recorder *MockAppCallbackClientMockRecorder
}

// MockAppCallbackClientMockRecorder is the mock recorder for MockAppCallbackClient.
type MockAppCallbackClientMockRecorder struct {
	mock *MockAppCallbackClient
}

// NewMockAppCallbackClient creates a new mock instance.
func NewMockAppCallbackClient(ctrl *gomock.Controller) *MockAppCallbackClient {
	mock := &MockAppCallbackClient{ctrl: ctrl}
	mock.recorder = &MockAppCallbackClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAppCallbackClient) EXPECT() *MockAppCallbackClientMockRecorder {
	return m.recorder
}

// ListTopicSubscriptions mocks base method.
func (m *MockAppCallbackClient) ListTopicSubscriptions(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*runtime.ListTopicSubscriptionsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListTopicSubscriptions", varargs...)
	ret0, _ := ret[0].(*runtime.ListTopicSubscriptionsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListTopicSubscriptions indicates an expected call of ListTopicSubscriptions.
func (mr *MockAppCallbackClientMockRecorder) ListTopicSubscriptions(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListTopicSubscriptions", reflect.TypeOf((*MockAppCallbackClient)(nil).ListTopicSubscriptions), varargs...)
}

// OnTopicEvent mocks base method.
func (m *MockAppCallbackClient) OnTopicEvent(ctx context.Context, in *runtime.TopicEventRequest, opts ...grpc.CallOption) (*runtime.TopicEventResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "OnTopicEvent", varargs...)
	ret0, _ := ret[0].(*runtime.TopicEventResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OnTopicEvent indicates an expected call of OnTopicEvent.
func (mr *MockAppCallbackClientMockRecorder) OnTopicEvent(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnTopicEvent", reflect.TypeOf((*MockAppCallbackClient)(nil).OnTopicEvent), varargs...)
}

// MockAppCallbackServer is a mock of AppCallbackServer interface.
type MockAppCallbackServer struct {
	ctrl     *gomock.Controller
	recorder *MockAppCallbackServerMockRecorder
}

// MockAppCallbackServerMockRecorder is the mock recorder for MockAppCallbackServer.
type MockAppCallbackServerMockRecorder struct {
	mock *MockAppCallbackServer
}

// NewMockAppCallbackServer creates a new mock instance.
func NewMockAppCallbackServer(ctrl *gomock.Controller) *MockAppCallbackServer {
	mock := &MockAppCallbackServer{ctrl: ctrl}
	mock.recorder = &MockAppCallbackServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAppCallbackServer) EXPECT() *MockAppCallbackServerMockRecorder {
	return m.recorder
}

// ListTopicSubscriptions mocks base method.
func (m *MockAppCallbackServer) ListTopicSubscriptions(arg0 context.Context, arg1 *empty.Empty) (*runtime.ListTopicSubscriptionsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListTopicSubscriptions", arg0, arg1)
	ret0, _ := ret[0].(*runtime.ListTopicSubscriptionsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListTopicSubscriptions indicates an expected call of ListTopicSubscriptions.
func (mr *MockAppCallbackServerMockRecorder) ListTopicSubscriptions(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListTopicSubscriptions", reflect.TypeOf((*MockAppCallbackServer)(nil).ListTopicSubscriptions), arg0, arg1)
}

// OnTopicEvent mocks base method.
func (m *MockAppCallbackServer) OnTopicEvent(arg0 context.Context, arg1 *runtime.TopicEventRequest) (*runtime.TopicEventResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OnTopicEvent", arg0, arg1)
	ret0, _ := ret[0].(*runtime.TopicEventResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OnTopicEvent indicates an expected call of OnTopicEvent.
func (mr *MockAppCallbackServerMockRecorder) OnTopicEvent(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnTopicEvent", reflect.TypeOf((*MockAppCallbackServer)(nil).OnTopicEvent), arg0, arg1)
}
