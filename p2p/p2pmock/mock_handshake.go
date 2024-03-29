// Code generated by MockGen. DO NOT EDIT.
// Source: handshake.go

// Package p2pmock is a generated GoMock package.
package p2pmock

import (
	context "context"
	p2pcommon "github.com/aergoio/aergo/v2/p2p/p2pcommon"
	types "github.com/aergoio/aergo/v2/types"
	gomock "github.com/golang/mock/gomock"
	io "io"
	reflect "reflect"
	time "time"
)

// MockHSHandlerFactory is a mock of HSHandlerFactory interface
type MockHSHandlerFactory struct {
	ctrl     *gomock.Controller
	recorder *MockHSHandlerFactoryMockRecorder
}

// MockHSHandlerFactoryMockRecorder is the mock recorder for MockHSHandlerFactory
type MockHSHandlerFactoryMockRecorder struct {
	mock *MockHSHandlerFactory
}

// NewMockHSHandlerFactory creates a new mock instance
func NewMockHSHandlerFactory(ctrl *gomock.Controller) *MockHSHandlerFactory {
	mock := &MockHSHandlerFactory{ctrl: ctrl}
	mock.recorder = &MockHSHandlerFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHSHandlerFactory) EXPECT() *MockHSHandlerFactoryMockRecorder {
	return m.recorder
}

// CreateHSHandler mocks base method
func (m *MockHSHandlerFactory) CreateHSHandler(outbound bool, pid types.PeerID) p2pcommon.HSHandler {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateHSHandler", outbound, pid)
	ret0, _ := ret[0].(p2pcommon.HSHandler)
	return ret0
}

// CreateHSHandler indicates an expected call of CreateHSHandler
func (mr *MockHSHandlerFactoryMockRecorder) CreateHSHandler(outbound, pid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateHSHandler", reflect.TypeOf((*MockHSHandlerFactory)(nil).CreateHSHandler), outbound, pid)
}

// MockHSHandler is a mock of HSHandler interface
type MockHSHandler struct {
	ctrl     *gomock.Controller
	recorder *MockHSHandlerMockRecorder
}

// MockHSHandlerMockRecorder is the mock recorder for MockHSHandler
type MockHSHandlerMockRecorder struct {
	mock *MockHSHandler
}

// NewMockHSHandler creates a new mock instance
func NewMockHSHandler(ctrl *gomock.Controller) *MockHSHandler {
	mock := &MockHSHandler{ctrl: ctrl}
	mock.recorder = &MockHSHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHSHandler) EXPECT() *MockHSHandlerMockRecorder {
	return m.recorder
}

// Handle mocks base method
func (m *MockHSHandler) Handle(s io.ReadWriteCloser, ttl time.Duration) (*p2pcommon.HandshakeResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Handle", s, ttl)
	ret0, _ := ret[0].(*p2pcommon.HandshakeResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Handle indicates an expected call of Handle
func (mr *MockHSHandlerMockRecorder) Handle(s, ttl interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Handle", reflect.TypeOf((*MockHSHandler)(nil).Handle), s, ttl)
}

// MockVersionedManager is a mock of VersionedManager interface
type MockVersionedManager struct {
	ctrl     *gomock.Controller
	recorder *MockVersionedManagerMockRecorder
}

// MockVersionedManagerMockRecorder is the mock recorder for MockVersionedManager
type MockVersionedManagerMockRecorder struct {
	mock *MockVersionedManager
}

// NewMockVersionedManager creates a new mock instance
func NewMockVersionedManager(ctrl *gomock.Controller) *MockVersionedManager {
	mock := &MockVersionedManager{ctrl: ctrl}
	mock.recorder = &MockVersionedManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockVersionedManager) EXPECT() *MockVersionedManagerMockRecorder {
	return m.recorder
}

// FindBestP2PVersion mocks base method
func (m *MockVersionedManager) FindBestP2PVersion(versions []p2pcommon.P2PVersion) p2pcommon.P2PVersion {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindBestP2PVersion", versions)
	ret0, _ := ret[0].(p2pcommon.P2PVersion)
	return ret0
}

// FindBestP2PVersion indicates an expected call of FindBestP2PVersion
func (mr *MockVersionedManagerMockRecorder) FindBestP2PVersion(versions interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindBestP2PVersion", reflect.TypeOf((*MockVersionedManager)(nil).FindBestP2PVersion), versions)
}

// GetVersionedHandshaker mocks base method
func (m *MockVersionedManager) GetVersionedHandshaker(version p2pcommon.P2PVersion, peerID types.PeerID, rwc io.ReadWriteCloser) (p2pcommon.VersionedHandshaker, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVersionedHandshaker", version, peerID, rwc)
	ret0, _ := ret[0].(p2pcommon.VersionedHandshaker)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVersionedHandshaker indicates an expected call of GetVersionedHandshaker
func (mr *MockVersionedManagerMockRecorder) GetVersionedHandshaker(version, peerID, rwc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVersionedHandshaker", reflect.TypeOf((*MockVersionedManager)(nil).GetVersionedHandshaker), version, peerID, rwc)
}

// GetBestChainID mocks base method
func (m *MockVersionedManager) GetBestChainID() *types.ChainID {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBestChainID")
	ret0, _ := ret[0].(*types.ChainID)
	return ret0
}

// GetBestChainID indicates an expected call of GetBestChainID
func (mr *MockVersionedManagerMockRecorder) GetBestChainID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBestChainID", reflect.TypeOf((*MockVersionedManager)(nil).GetBestChainID))
}

// GetChainID mocks base method
func (m *MockVersionedManager) GetChainID(no types.BlockNo) *types.ChainID {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetChainID", no)
	ret0, _ := ret[0].(*types.ChainID)
	return ret0
}

// GetChainID indicates an expected call of GetChainID
func (mr *MockVersionedManagerMockRecorder) GetChainID(no interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetChainID", reflect.TypeOf((*MockVersionedManager)(nil).GetChainID), no)
}

// MockVersionedHandshaker is a mock of VersionedHandshaker interface
type MockVersionedHandshaker struct {
	ctrl     *gomock.Controller
	recorder *MockVersionedHandshakerMockRecorder
}

// MockVersionedHandshakerMockRecorder is the mock recorder for MockVersionedHandshaker
type MockVersionedHandshakerMockRecorder struct {
	mock *MockVersionedHandshaker
}

// NewMockVersionedHandshaker creates a new mock instance
func NewMockVersionedHandshaker(ctrl *gomock.Controller) *MockVersionedHandshaker {
	mock := &MockVersionedHandshaker{ctrl: ctrl}
	mock.recorder = &MockVersionedHandshakerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockVersionedHandshaker) EXPECT() *MockVersionedHandshakerMockRecorder {
	return m.recorder
}

// DoForOutbound mocks base method
func (m *MockVersionedHandshaker) DoForOutbound(ctx context.Context) (*p2pcommon.HandshakeResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DoForOutbound", ctx)
	ret0, _ := ret[0].(*p2pcommon.HandshakeResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DoForOutbound indicates an expected call of DoForOutbound
func (mr *MockVersionedHandshakerMockRecorder) DoForOutbound(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DoForOutbound", reflect.TypeOf((*MockVersionedHandshaker)(nil).DoForOutbound), ctx)
}

// DoForInbound mocks base method
func (m *MockVersionedHandshaker) DoForInbound(ctx context.Context) (*p2pcommon.HandshakeResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DoForInbound", ctx)
	ret0, _ := ret[0].(*p2pcommon.HandshakeResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DoForInbound indicates an expected call of DoForInbound
func (mr *MockVersionedHandshakerMockRecorder) DoForInbound(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DoForInbound", reflect.TypeOf((*MockVersionedHandshaker)(nil).DoForInbound), ctx)
}

// GetMsgRW mocks base method
func (m *MockVersionedHandshaker) GetMsgRW() p2pcommon.MsgReadWriter {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMsgRW")
	ret0, _ := ret[0].(p2pcommon.MsgReadWriter)
	return ret0
}

// GetMsgRW indicates an expected call of GetMsgRW
func (mr *MockVersionedHandshakerMockRecorder) GetMsgRW() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMsgRW", reflect.TypeOf((*MockVersionedHandshaker)(nil).GetMsgRW))
}
