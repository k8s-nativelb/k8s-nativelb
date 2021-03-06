// Code generated by MockGen. DO NOT EDIT.
// Source: handler.go

// Package handler is a generated GoMock package.
package handler

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockHandlerInterface is a mock of HandlerInterface interface
type MockHandlerInterface struct {
	ctrl     *gomock.Controller
	recorder *MockHandlerInterfaceMockRecorder
}

// MockHandlerInterfaceMockRecorder is the mock recorder for MockHandlerInterface
type MockHandlerInterfaceMockRecorder struct {
	mock *MockHandlerInterface
}

// NewMockHandlerInterface creates a new mock instance
func NewMockHandlerInterface(ctrl *gomock.Controller) *MockHandlerInterface {
	mock := &MockHandlerInterface{ctrl: ctrl}
	mock.recorder = &MockHandlerInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHandlerInterface) EXPECT() *MockHandlerInterfaceMockRecorder {
	return m.recorder
}

// GetPid mocks base method
func (m *MockHandlerInterface) GetPid(arg0 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPid", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPid indicates an expected call of GetPid
func (mr *MockHandlerInterfaceMockRecorder) GetPid(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPid", reflect.TypeOf((*MockHandlerInterface)(nil).GetPid), arg0)
}

// CheckHaproxyConfig mocks base method
func (m *MockHandlerInterface) CheckHaproxyConfig() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckHaproxyConfig")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckHaproxyConfig indicates an expected call of CheckHaproxyConfig
func (mr *MockHandlerInterfaceMockRecorder) CheckHaproxyConfig() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckHaproxyConfig", reflect.TypeOf((*MockHandlerInterface)(nil).CheckHaproxyConfig))
}

// CheckNginxConfig mocks base method
func (m *MockHandlerInterface) CheckNginxConfig() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckNginxConfig")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckNginxConfig indicates an expected call of CheckNginxConfig
func (mr *MockHandlerInterfaceMockRecorder) CheckNginxConfig() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckNginxConfig", reflect.TypeOf((*MockHandlerInterface)(nil).CheckNginxConfig))
}

// CheckKeepalivedConfig mocks base method
func (m *MockHandlerInterface) CheckKeepalivedConfig() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckKeepalivedConfig")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckKeepalivedConfig indicates an expected call of CheckKeepalivedConfig
func (mr *MockHandlerInterfaceMockRecorder) CheckKeepalivedConfig() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckKeepalivedConfig", reflect.TypeOf((*MockHandlerInterface)(nil).CheckKeepalivedConfig))
}

// StartHaproxy mocks base method
func (m *MockHandlerInterface) StartHaproxy() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartHaproxy")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StartHaproxy indicates an expected call of StartHaproxy
func (mr *MockHandlerInterfaceMockRecorder) StartHaproxy() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartHaproxy", reflect.TypeOf((*MockHandlerInterface)(nil).StartHaproxy))
}

// StartNginx mocks base method
func (m *MockHandlerInterface) StartNginx() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartNginx")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StartNginx indicates an expected call of StartNginx
func (mr *MockHandlerInterfaceMockRecorder) StartNginx() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartNginx", reflect.TypeOf((*MockHandlerInterface)(nil).StartNginx))
}

// StartKeepalived mocks base method
func (m *MockHandlerInterface) StartKeepalived() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartKeepalived")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StartKeepalived indicates an expected call of StartKeepalived
func (mr *MockHandlerInterfaceMockRecorder) StartKeepalived() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartKeepalived", reflect.TypeOf((*MockHandlerInterface)(nil).StartKeepalived))
}

// ReloadHaproxy mocks base method
func (m *MockHandlerInterface) ReloadHaproxy(arg0 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReloadHaproxy", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReloadHaproxy indicates an expected call of ReloadHaproxy
func (mr *MockHandlerInterfaceMockRecorder) ReloadHaproxy(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReloadHaproxy", reflect.TypeOf((*MockHandlerInterface)(nil).ReloadHaproxy), arg0)
}

// ReloadNginx mocks base method
func (m *MockHandlerInterface) ReloadNginx(arg0 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReloadNginx", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReloadNginx indicates an expected call of ReloadNginx
func (mr *MockHandlerInterfaceMockRecorder) ReloadNginx(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReloadNginx", reflect.TypeOf((*MockHandlerInterface)(nil).ReloadNginx), arg0)
}

// ReloadKeepalived mocks base method
func (m *MockHandlerInterface) ReloadKeepalived(arg0 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReloadKeepalived", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReloadKeepalived indicates an expected call of ReloadKeepalived
func (mr *MockHandlerInterfaceMockRecorder) ReloadKeepalived(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReloadKeepalived", reflect.TypeOf((*MockHandlerInterface)(nil).ReloadKeepalived), arg0)
}

// StopHaproxy mocks base method
func (m *MockHandlerInterface) StopHaproxy(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StopHaproxy", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// StopHaproxy indicates an expected call of StopHaproxy
func (mr *MockHandlerInterfaceMockRecorder) StopHaproxy(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StopHaproxy", reflect.TypeOf((*MockHandlerInterface)(nil).StopHaproxy), arg0)
}

// StopNginx mocks base method
func (m *MockHandlerInterface) StopNginx(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StopNginx", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// StopNginx indicates an expected call of StopNginx
func (mr *MockHandlerInterfaceMockRecorder) StopNginx(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StopNginx", reflect.TypeOf((*MockHandlerInterface)(nil).StopNginx), arg0)
}

// StopKeepalived mocks base method
func (m *MockHandlerInterface) StopKeepalived(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StopKeepalived", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// StopKeepalived indicates an expected call of StopKeepalived
func (mr *MockHandlerInterfaceMockRecorder) StopKeepalived(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StopKeepalived", reflect.TypeOf((*MockHandlerInterface)(nil).StopKeepalived), arg0)
}
