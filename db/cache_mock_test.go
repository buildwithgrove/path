// Code generated by MockGen. DO NOT EDIT.
// Source: cache.go
//
// Generated by this command:
//
//	mockgen -source=cache.go -destination=cache_mock_test.go -package=db
//

// Package db is a generated GoMock package.
package db

import (
	context "context"
	reflect "reflect"

	user "github.com/buildwithgrove/path/user"
	gomock "go.uber.org/mock/gomock"
)

// MockDriver is a mock of Driver interface.
type MockDriver struct {
	ctrl     *gomock.Controller
	recorder *MockDriverMockRecorder
}

// MockDriverMockRecorder is the mock recorder for MockDriver.
type MockDriverMockRecorder struct {
	mock *MockDriver
}

// NewMockDriver creates a new mock instance.
func NewMockDriver(ctrl *gomock.Controller) *MockDriver {
	mock := &MockDriver{ctrl: ctrl}
	mock.recorder = &MockDriverMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDriver) EXPECT() *MockDriverMockRecorder {
	return m.recorder
}

// GetUserApps mocks base method.
func (m *MockDriver) GetUserApps(ctx context.Context) (map[user.UserAppID]user.UserApp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserApps", ctx)
	ret0, _ := ret[0].(map[user.UserAppID]user.UserApp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserApps indicates an expected call of GetUserApps.
func (mr *MockDriverMockRecorder) GetUserApps(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserApps", reflect.TypeOf((*MockDriver)(nil).GetUserApps), ctx)
}
