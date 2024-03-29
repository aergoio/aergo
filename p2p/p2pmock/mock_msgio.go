// Code generated by MockGen. DO NOT EDIT.
// Source: msgio.go

// Package p2pmock is a generated GoMock package.
package p2pmock

import (
	p2pcommon "github.com/aergoio/aergo/v2/p2p/p2pcommon"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockMsgReadWriter is a mock of MsgReadWriter interface
type MockMsgReadWriter struct {
	ctrl     *gomock.Controller
	recorder *MockMsgReadWriterMockRecorder
}

// MockMsgReadWriterMockRecorder is the mock recorder for MockMsgReadWriter
type MockMsgReadWriterMockRecorder struct {
	mock *MockMsgReadWriter
}

// NewMockMsgReadWriter creates a new mock instance
func NewMockMsgReadWriter(ctrl *gomock.Controller) *MockMsgReadWriter {
	mock := &MockMsgReadWriter{ctrl: ctrl}
	mock.recorder = &MockMsgReadWriterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMsgReadWriter) EXPECT() *MockMsgReadWriterMockRecorder {
	return m.recorder
}

// ReadMsg mocks base method
func (m *MockMsgReadWriter) ReadMsg() (p2pcommon.Message, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadMsg")
	ret0, _ := ret[0].(p2pcommon.Message)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadMsg indicates an expected call of ReadMsg
func (mr *MockMsgReadWriterMockRecorder) ReadMsg() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadMsg", reflect.TypeOf((*MockMsgReadWriter)(nil).ReadMsg))
}

// WriteMsg mocks base method
func (m *MockMsgReadWriter) WriteMsg(msg p2pcommon.Message) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteMsg", msg)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteMsg indicates an expected call of WriteMsg
func (mr *MockMsgReadWriterMockRecorder) WriteMsg(msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteMsg", reflect.TypeOf((*MockMsgReadWriter)(nil).WriteMsg), msg)
}

// Close mocks base method
func (m *MockMsgReadWriter) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close
func (mr *MockMsgReadWriterMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockMsgReadWriter)(nil).Close))
}

// AddIOListener mocks base method
func (m *MockMsgReadWriter) AddIOListener(l p2pcommon.MsgIOListener) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddIOListener", l)
}

// AddIOListener indicates an expected call of AddIOListener
func (mr *MockMsgReadWriterMockRecorder) AddIOListener(l interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddIOListener", reflect.TypeOf((*MockMsgReadWriter)(nil).AddIOListener), l)
}

// MockMsgIOListener is a mock of MsgIOListener interface
type MockMsgIOListener struct {
	ctrl     *gomock.Controller
	recorder *MockMsgIOListenerMockRecorder
}

// MockMsgIOListenerMockRecorder is the mock recorder for MockMsgIOListener
type MockMsgIOListenerMockRecorder struct {
	mock *MockMsgIOListener
}

// NewMockMsgIOListener creates a new mock instance
func NewMockMsgIOListener(ctrl *gomock.Controller) *MockMsgIOListener {
	mock := &MockMsgIOListener{ctrl: ctrl}
	mock.recorder = &MockMsgIOListenerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMsgIOListener) EXPECT() *MockMsgIOListenerMockRecorder {
	return m.recorder
}

// OnRead mocks base method
func (m *MockMsgIOListener) OnRead(protocol p2pcommon.SubProtocol, read int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "OnRead", protocol, read)
}

// OnRead indicates an expected call of OnRead
func (mr *MockMsgIOListenerMockRecorder) OnRead(protocol, read interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnRead", reflect.TypeOf((*MockMsgIOListener)(nil).OnRead), protocol, read)
}

// OnWrite mocks base method
func (m *MockMsgIOListener) OnWrite(protocol p2pcommon.SubProtocol, write int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "OnWrite", protocol, write)
}

// OnWrite indicates an expected call of OnWrite
func (mr *MockMsgIOListenerMockRecorder) OnWrite(protocol, write interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnWrite", reflect.TypeOf((*MockMsgIOListener)(nil).OnWrite), protocol, write)
}
