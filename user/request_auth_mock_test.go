// Code generated by MockGen. DO NOT EDIT.
// Source: request_authenticator.go
//
// Generated by this command:
//
//	mockgen -source=request_authenticator.go -destination=request_auth_mock_test.go -package=user
//

// Package user is a generated GoMock package.
package user

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// Mockcache is a mock of cache interface.
type Mockcache struct {
	ctrl     *gomock.Controller
	recorder *MockcacheMockRecorder
}

// MockcacheMockRecorder is the mock recorder for Mockcache.
type MockcacheMockRecorder struct {
	mock *Mockcache
}

// NewMockcache creates a new mock instance.
func NewMockcache(ctrl *gomock.Controller) *Mockcache {
	mock := &Mockcache{ctrl: ctrl}
	mock.recorder = &MockcacheMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Mockcache) EXPECT() *MockcacheMockRecorder {
	return m.recorder
}

// GetUserApp mocks base method.
func (m *Mockcache) GetUserApp(ctx context.Context, userAppID UserAppID) (UserApp, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserApp", ctx, userAppID)
	ret0, _ := ret[0].(UserApp)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// GetUserApp indicates an expected call of GetUserApp.
func (mr *MockcacheMockRecorder) GetUserApp(ctx, userAppID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserApp", reflect.TypeOf((*Mockcache)(nil).GetUserApp), ctx, userAppID)
}