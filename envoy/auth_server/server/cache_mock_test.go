// Code generated by MockGen. DO NOT EDIT.
// Source: server.go
//
// Generated by this command:
//
//	mockgen -source=server.go -destination=cache_mock_test.go -package=server
//

// Package server is a generated GoMock package.
package server

import (
	reflect "reflect"

	user "github.com/buildwithgrove/auth-server/user"
	gomock "go.uber.org/mock/gomock"
)

// MockendpointDataCache is a mock of endpointDataCache interface.
type MockendpointDataCache struct {
	ctrl     *gomock.Controller
	recorder *MockendpointDataCacheMockRecorder
}

// MockendpointDataCacheMockRecorder is the mock recorder for MockendpointDataCache.
type MockendpointDataCacheMockRecorder struct {
	mock *MockendpointDataCache
}

// NewMockendpointDataCache creates a new mock instance.
func NewMockendpointDataCache(ctrl *gomock.Controller) *MockendpointDataCache {
	mock := &MockendpointDataCache{ctrl: ctrl}
	mock.recorder = &MockendpointDataCacheMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockendpointDataCache) EXPECT() *MockendpointDataCacheMockRecorder {
	return m.recorder
}

// GetGatewayEndpoint mocks base method.
func (m *MockendpointDataCache) GetGatewayEndpoint(arg0 user.EndpointID) (user.GatewayEndpoint, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGatewayEndpoint", arg0)
	ret0, _ := ret[0].(user.GatewayEndpoint)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// GetGatewayEndpoint indicates an expected call of GetGatewayEndpoint.
func (mr *MockendpointDataCacheMockRecorder) GetGatewayEndpoint(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGatewayEndpoint", reflect.TypeOf((*MockendpointDataCache)(nil).GetGatewayEndpoint), arg0)
}

// MockAuthorizer is a mock of Authorizer interface.
type MockAuthorizer struct {
	ctrl     *gomock.Controller
	recorder *MockAuthorizerMockRecorder
}

// MockAuthorizerMockRecorder is the mock recorder for MockAuthorizer.
type MockAuthorizerMockRecorder struct {
	mock *MockAuthorizer
}

// NewMockAuthorizer creates a new mock instance.
func NewMockAuthorizer(ctrl *gomock.Controller) *MockAuthorizer {
	mock := &MockAuthorizer{ctrl: ctrl}
	mock.recorder = &MockAuthorizerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAuthorizer) EXPECT() *MockAuthorizerMockRecorder {
	return m.recorder
}

// authorizeRequest mocks base method.
func (m *MockAuthorizer) authorizeRequest(arg0 user.ProviderUserID, arg1 user.GatewayEndpoint) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "authorizeRequest", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// authorizeRequest indicates an expected call of authorizeRequest.
func (mr *MockAuthorizerMockRecorder) authorizeRequest(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "authorizeRequest", reflect.TypeOf((*MockAuthorizer)(nil).authorizeRequest), arg0, arg1)
}
