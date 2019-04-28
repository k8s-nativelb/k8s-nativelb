// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/proto/proto.pb.go

// Package proto is a generated GoMock package.
package proto

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	grpc "google.golang.org/grpc"
	reflect "reflect"
)

// MockNativeLoadBalancerAgentClient is a mock of NativeLoadBalancerAgentClient interface
type MockNativeLoadBalancerAgentClient struct {
	ctrl     *gomock.Controller
	recorder *MockNativeLoadBalancerAgentClientMockRecorder
}

// MockNativeLoadBalancerAgentClientMockRecorder is the mock recorder for MockNativeLoadBalancerAgentClient
type MockNativeLoadBalancerAgentClientMockRecorder struct {
	mock *MockNativeLoadBalancerAgentClient
}

// NewMockNativeLoadBalancerAgentClient creates a new mock instance
func NewMockNativeLoadBalancerAgentClient(ctrl *gomock.Controller) *MockNativeLoadBalancerAgentClient {
	mock := &MockNativeLoadBalancerAgentClient{ctrl: ctrl}
	mock.recorder = &MockNativeLoadBalancerAgentClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockNativeLoadBalancerAgentClient) EXPECT() *MockNativeLoadBalancerAgentClientMockRecorder {
	return m.recorder
}

// CreateServers mocks base method
func (m *MockNativeLoadBalancerAgentClient) CreateServers(ctx context.Context, in *FarmSpec, opts ...grpc.CallOption) (*Result, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateServers", varargs...)
	ret0, _ := ret[0].(*Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateServers indicates an expected call of CreateServers
func (mr *MockNativeLoadBalancerAgentClientMockRecorder) CreateServers(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateServers", reflect.TypeOf((*MockNativeLoadBalancerAgentClient)(nil).CreateServers), varargs...)
}

// UpdateServers mocks base method
func (m *MockNativeLoadBalancerAgentClient) UpdateServers(ctx context.Context, in *FarmSpec, opts ...grpc.CallOption) (*Result, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateServers", varargs...)
	ret0, _ := ret[0].(*Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateServers indicates an expected call of UpdateServers
func (mr *MockNativeLoadBalancerAgentClientMockRecorder) UpdateServers(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateServers", reflect.TypeOf((*MockNativeLoadBalancerAgentClient)(nil).UpdateServers), varargs...)
}

// DeleteServers mocks base method
func (m *MockNativeLoadBalancerAgentClient) DeleteServers(ctx context.Context, in *FarmSpec, opts ...grpc.CallOption) (*Result, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteServers", varargs...)
	ret0, _ := ret[0].(*Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteServers indicates an expected call of DeleteServers
func (mr *MockNativeLoadBalancerAgentClientMockRecorder) DeleteServers(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteServers", reflect.TypeOf((*MockNativeLoadBalancerAgentClient)(nil).DeleteServers), varargs...)
}

// InitAgent mocks base method
func (m *MockNativeLoadBalancerAgentClient) InitAgent(ctx context.Context, in *InitAgentData, opts ...grpc.CallOption) (*InitAgentResult, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "InitAgent", varargs...)
	ret0, _ := ret[0].(*InitAgentResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InitAgent indicates an expected call of InitAgent
func (mr *MockNativeLoadBalancerAgentClientMockRecorder) InitAgent(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitAgent", reflect.TypeOf((*MockNativeLoadBalancerAgentClient)(nil).InitAgent), varargs...)
}

// GetAgentStatus mocks base method
func (m *MockNativeLoadBalancerAgentClient) GetAgentStatus(ctx context.Context, in *Command, opts ...grpc.CallOption) (*AgentStatus, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetAgentStatus", varargs...)
	ret0, _ := ret[0].(*AgentStatus)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAgentStatus indicates an expected call of GetAgentStatus
func (mr *MockNativeLoadBalancerAgentClientMockRecorder) GetAgentStatus(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAgentStatus", reflect.TypeOf((*MockNativeLoadBalancerAgentClient)(nil).GetAgentStatus), varargs...)
}

// GetServersStats mocks base method
func (m *MockNativeLoadBalancerAgentClient) GetServersStats(ctx context.Context, in *Command, opts ...grpc.CallOption) (*ServersStats, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetServersStats", varargs...)
	ret0, _ := ret[0].(*ServersStats)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetServersStats indicates an expected call of GetServersStats
func (mr *MockNativeLoadBalancerAgentClientMockRecorder) GetServersStats(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetServersStats", reflect.TypeOf((*MockNativeLoadBalancerAgentClient)(nil).GetServersStats), varargs...)
}

// StopAgent mocks base method
func (m *MockNativeLoadBalancerAgentClient) StopAgent(ctx context.Context, in *Command, opts ...grpc.CallOption) (*Result, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "StopAgent", varargs...)
	ret0, _ := ret[0].(*Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StopAgent indicates an expected call of StopAgent
func (mr *MockNativeLoadBalancerAgentClientMockRecorder) StopAgent(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StopAgent", reflect.TypeOf((*MockNativeLoadBalancerAgentClient)(nil).StopAgent), varargs...)
}

// UpdateAgentSyncVersion mocks base method
func (m *MockNativeLoadBalancerAgentClient) UpdateAgentSyncVersion(ctx context.Context, in *InitAgentData, opts ...grpc.CallOption) (*Result, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateAgentSyncVersion", varargs...)
	ret0, _ := ret[0].(*Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateAgentSyncVersion indicates an expected call of UpdateAgentSyncVersion
func (mr *MockNativeLoadBalancerAgentClientMockRecorder) UpdateAgentSyncVersion(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAgentSyncVersion", reflect.TypeOf((*MockNativeLoadBalancerAgentClient)(nil).UpdateAgentSyncVersion), varargs...)
}

// MockNativeLoadBalancerAgentServer is a mock of NativeLoadBalancerAgentServer interface
type MockNativeLoadBalancerAgentServer struct {
	ctrl     *gomock.Controller
	recorder *MockNativeLoadBalancerAgentServerMockRecorder
}

// MockNativeLoadBalancerAgentServerMockRecorder is the mock recorder for MockNativeLoadBalancerAgentServer
type MockNativeLoadBalancerAgentServerMockRecorder struct {
	mock *MockNativeLoadBalancerAgentServer
}

// NewMockNativeLoadBalancerAgentServer creates a new mock instance
func NewMockNativeLoadBalancerAgentServer(ctrl *gomock.Controller) *MockNativeLoadBalancerAgentServer {
	mock := &MockNativeLoadBalancerAgentServer{ctrl: ctrl}
	mock.recorder = &MockNativeLoadBalancerAgentServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockNativeLoadBalancerAgentServer) EXPECT() *MockNativeLoadBalancerAgentServerMockRecorder {
	return m.recorder
}

// CreateServers mocks base method
func (m *MockNativeLoadBalancerAgentServer) CreateServers(arg0 context.Context, arg1 *FarmSpec) (*Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateServers", arg0, arg1)
	ret0, _ := ret[0].(*Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateServers indicates an expected call of CreateServers
func (mr *MockNativeLoadBalancerAgentServerMockRecorder) CreateServers(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateServers", reflect.TypeOf((*MockNativeLoadBalancerAgentServer)(nil).CreateServers), arg0, arg1)
}

// UpdateServers mocks base method
func (m *MockNativeLoadBalancerAgentServer) UpdateServers(arg0 context.Context, arg1 *FarmSpec) (*Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateServers", arg0, arg1)
	ret0, _ := ret[0].(*Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateServers indicates an expected call of UpdateServers
func (mr *MockNativeLoadBalancerAgentServerMockRecorder) UpdateServers(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateServers", reflect.TypeOf((*MockNativeLoadBalancerAgentServer)(nil).UpdateServers), arg0, arg1)
}

// DeleteServers mocks base method
func (m *MockNativeLoadBalancerAgentServer) DeleteServers(arg0 context.Context, arg1 *FarmSpec) (*Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteServers", arg0, arg1)
	ret0, _ := ret[0].(*Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteServers indicates an expected call of DeleteServers
func (mr *MockNativeLoadBalancerAgentServerMockRecorder) DeleteServers(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteServers", reflect.TypeOf((*MockNativeLoadBalancerAgentServer)(nil).DeleteServers), arg0, arg1)
}

// InitAgent mocks base method
func (m *MockNativeLoadBalancerAgentServer) InitAgent(arg0 context.Context, arg1 *InitAgentData) (*InitAgentResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InitAgent", arg0, arg1)
	ret0, _ := ret[0].(*InitAgentResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InitAgent indicates an expected call of InitAgent
func (mr *MockNativeLoadBalancerAgentServerMockRecorder) InitAgent(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitAgent", reflect.TypeOf((*MockNativeLoadBalancerAgentServer)(nil).InitAgent), arg0, arg1)
}

// GetAgentStatus mocks base method
func (m *MockNativeLoadBalancerAgentServer) GetAgentStatus(arg0 context.Context, arg1 *Command) (*AgentStatus, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAgentStatus", arg0, arg1)
	ret0, _ := ret[0].(*AgentStatus)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAgentStatus indicates an expected call of GetAgentStatus
func (mr *MockNativeLoadBalancerAgentServerMockRecorder) GetAgentStatus(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAgentStatus", reflect.TypeOf((*MockNativeLoadBalancerAgentServer)(nil).GetAgentStatus), arg0, arg1)
}

// GetServersStats mocks base method
func (m *MockNativeLoadBalancerAgentServer) GetServersStats(arg0 context.Context, arg1 *Command) (*ServersStats, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetServersStats", arg0, arg1)
	ret0, _ := ret[0].(*ServersStats)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetServersStats indicates an expected call of GetServersStats
func (mr *MockNativeLoadBalancerAgentServerMockRecorder) GetServersStats(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetServersStats", reflect.TypeOf((*MockNativeLoadBalancerAgentServer)(nil).GetServersStats), arg0, arg1)
}

// StopAgent mocks base method
func (m *MockNativeLoadBalancerAgentServer) StopAgent(arg0 context.Context, arg1 *Command) (*Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StopAgent", arg0, arg1)
	ret0, _ := ret[0].(*Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StopAgent indicates an expected call of StopAgent
func (mr *MockNativeLoadBalancerAgentServerMockRecorder) StopAgent(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StopAgent", reflect.TypeOf((*MockNativeLoadBalancerAgentServer)(nil).StopAgent), arg0, arg1)
}

// UpdateAgentSyncVersion mocks base method
func (m *MockNativeLoadBalancerAgentServer) UpdateAgentSyncVersion(arg0 context.Context, arg1 *InitAgentData) (*Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateAgentSyncVersion", arg0, arg1)
	ret0, _ := ret[0].(*Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateAgentSyncVersion indicates an expected call of UpdateAgentSyncVersion
func (mr *MockNativeLoadBalancerAgentServerMockRecorder) UpdateAgentSyncVersion(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAgentSyncVersion", reflect.TypeOf((*MockNativeLoadBalancerAgentServer)(nil).UpdateAgentSyncVersion), arg0, arg1)
}
