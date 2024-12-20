// Code generated by MockGen. DO NOT EDIT.
// Source: ./envoy/auth_server/proto/gateway_endpoint_grpc.pb.go
//
// Generated by this command:
//
//	mockgen -source=./envoy/auth_server/proto/gateway_endpoint_grpc.pb.go -destination=./envoy/auth_server/endpoint_store/client_mock_test.go -package=endpointstore -mock_names=GatewayEndpointsClient=MockGatewayEndpointsClient
//

// Package endpointstore is a generated GoMock package.
package endpointstore

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
	grpc "google.golang.org/grpc"
	
	proto "github.com/buildwithgrove/path/envoy/auth_server/proto"
)

// MockGatewayEndpointsClient is a mock of GatewayEndpointsClient interface.
type MockGatewayEndpointsClient struct {
	ctrl     *gomock.Controller
	recorder *MockGatewayEndpointsClientMockRecorder
	isgomock struct{}
}

// MockGatewayEndpointsClientMockRecorder is the mock recorder for MockGatewayEndpointsClient.
type MockGatewayEndpointsClientMockRecorder struct {
	mock *MockGatewayEndpointsClient
}

// NewMockGatewayEndpointsClient creates a new mock instance.
func NewMockGatewayEndpointsClient(ctrl *gomock.Controller) *MockGatewayEndpointsClient {
	mock := &MockGatewayEndpointsClient{ctrl: ctrl}
	mock.recorder = &MockGatewayEndpointsClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGatewayEndpointsClient) EXPECT() *MockGatewayEndpointsClientMockRecorder {
	return m.recorder
}

// FetchAuthDataSync mocks base method.
func (m *MockGatewayEndpointsClient) FetchAuthDataSync(ctx context.Context, in *proto.AuthDataRequest, opts ...grpc.CallOption) (*proto.AuthDataResponse, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "FetchAuthDataSync", varargs...)
	ret0, _ := ret[0].(*proto.AuthDataResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchAuthDataSync indicates an expected call of FetchAuthDataSync.
func (mr *MockGatewayEndpointsClientMockRecorder) FetchAuthDataSync(ctx, in any, opts ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchAuthDataSync", reflect.TypeOf((*MockGatewayEndpointsClient)(nil).FetchAuthDataSync), varargs...)
}

// StreamAuthDataUpdates mocks base method.
func (m *MockGatewayEndpointsClient) StreamAuthDataUpdates(ctx context.Context, in *proto.AuthDataUpdatesRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[proto.AuthDataUpdate], error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "StreamAuthDataUpdates", varargs...)
	ret0, _ := ret[0].(grpc.ServerStreamingClient[proto.AuthDataUpdate])
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StreamAuthDataUpdates indicates an expected call of StreamAuthDataUpdates.
func (mr *MockGatewayEndpointsClientMockRecorder) StreamAuthDataUpdates(ctx, in any, opts ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StreamAuthDataUpdates", reflect.TypeOf((*MockGatewayEndpointsClient)(nil).StreamAuthDataUpdates), varargs...)
}

// MockGatewayEndpointsServer is a mock of GatewayEndpointsServer interface.
type MockGatewayEndpointsServer struct {
	ctrl     *gomock.Controller
	recorder *MockGatewayEndpointsServerMockRecorder
	isgomock struct{}
}

// MockGatewayEndpointsServerMockRecorder is the mock recorder for MockGatewayEndpointsServer.
type MockGatewayEndpointsServerMockRecorder struct {
	mock *MockGatewayEndpointsServer
}

// NewMockGatewayEndpointsServer creates a new mock instance.
func NewMockGatewayEndpointsServer(ctrl *gomock.Controller) *MockGatewayEndpointsServer {
	mock := &MockGatewayEndpointsServer{ctrl: ctrl}
	mock.recorder = &MockGatewayEndpointsServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGatewayEndpointsServer) EXPECT() *MockGatewayEndpointsServerMockRecorder {
	return m.recorder
}

// FetchAuthDataSync mocks base method.
func (m *MockGatewayEndpointsServer) FetchAuthDataSync(arg0 context.Context, arg1 *proto.AuthDataRequest) (*proto.AuthDataResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchAuthDataSync", arg0, arg1)
	ret0, _ := ret[0].(*proto.AuthDataResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchAuthDataSync indicates an expected call of FetchAuthDataSync.
func (mr *MockGatewayEndpointsServerMockRecorder) FetchAuthDataSync(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchAuthDataSync", reflect.TypeOf((*MockGatewayEndpointsServer)(nil).FetchAuthDataSync), arg0, arg1)
}

// StreamAuthDataUpdates mocks base method.
func (m *MockGatewayEndpointsServer) StreamAuthDataUpdates(arg0 *proto.AuthDataUpdatesRequest, arg1 grpc.ServerStreamingServer[proto.AuthDataUpdate]) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StreamAuthDataUpdates", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// StreamAuthDataUpdates indicates an expected call of StreamAuthDataUpdates.
func (mr *MockGatewayEndpointsServerMockRecorder) StreamAuthDataUpdates(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StreamAuthDataUpdates", reflect.TypeOf((*MockGatewayEndpointsServer)(nil).StreamAuthDataUpdates), arg0, arg1)
}

// mustEmbedUnimplementedGatewayEndpointsServer mocks base method.
func (m *MockGatewayEndpointsServer) mustEmbedUnimplementedGatewayEndpointsServer() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "mustEmbedUnimplementedGatewayEndpointsServer")
}

// mustEmbedUnimplementedGatewayEndpointsServer indicates an expected call of mustEmbedUnimplementedGatewayEndpointsServer.
func (mr *MockGatewayEndpointsServerMockRecorder) mustEmbedUnimplementedGatewayEndpointsServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "mustEmbedUnimplementedGatewayEndpointsServer", reflect.TypeOf((*MockGatewayEndpointsServer)(nil).mustEmbedUnimplementedGatewayEndpointsServer))
}

// MockUnsafeGatewayEndpointsServer is a mock of UnsafeGatewayEndpointsServer interface.
type MockUnsafeGatewayEndpointsServer struct {
	ctrl     *gomock.Controller
	recorder *MockUnsafeGatewayEndpointsServerMockRecorder
	isgomock struct{}
}

// MockUnsafeGatewayEndpointsServerMockRecorder is the mock recorder for MockUnsafeGatewayEndpointsServer.
type MockUnsafeGatewayEndpointsServerMockRecorder struct {
	mock *MockUnsafeGatewayEndpointsServer
}

// NewMockUnsafeGatewayEndpointsServer creates a new mock instance.
func NewMockUnsafeGatewayEndpointsServer(ctrl *gomock.Controller) *MockUnsafeGatewayEndpointsServer {
	mock := &MockUnsafeGatewayEndpointsServer{ctrl: ctrl}
	mock.recorder = &MockUnsafeGatewayEndpointsServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUnsafeGatewayEndpointsServer) EXPECT() *MockUnsafeGatewayEndpointsServerMockRecorder {
	return m.recorder
}

// mustEmbedUnimplementedGatewayEndpointsServer mocks base method.
func (m *MockUnsafeGatewayEndpointsServer) mustEmbedUnimplementedGatewayEndpointsServer() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "mustEmbedUnimplementedGatewayEndpointsServer")
}

// mustEmbedUnimplementedGatewayEndpointsServer indicates an expected call of mustEmbedUnimplementedGatewayEndpointsServer.
func (mr *MockUnsafeGatewayEndpointsServerMockRecorder) mustEmbedUnimplementedGatewayEndpointsServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "mustEmbedUnimplementedGatewayEndpointsServer", reflect.TypeOf((*MockUnsafeGatewayEndpointsServer)(nil).mustEmbedUnimplementedGatewayEndpointsServer))
}
