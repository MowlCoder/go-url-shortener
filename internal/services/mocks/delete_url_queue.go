// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/services/delete_url_queue.go
//
// Generated by this command:
//
//	mockgen.exe -source=./internal/services/delete_url_queue.go -package=servicesmocks -destination=./internal/services/mocks/delete_url_queue.go
//
// Package servicesmocks is a generated GoMock package.
package servicesmocks

import (
	context "context"
	reflect "reflect"

	domain "github.com/MowlCoder/go-url-shortener/internal/domain"
	gomock "go.uber.org/mock/gomock"
)

// MockurlStorage is a mock of urlStorage interface.
type MockurlStorage struct {
	ctrl     *gomock.Controller
	recorder *MockurlStorageMockRecorder
}

// MockurlStorageMockRecorder is the mock recorder for MockurlStorage.
type MockurlStorageMockRecorder struct {
	mock *MockurlStorage
}

// NewMockurlStorage creates a new mock instance.
func NewMockurlStorage(ctrl *gomock.Controller) *MockurlStorage {
	mock := &MockurlStorage{ctrl: ctrl}
	mock.recorder = &MockurlStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockurlStorage) EXPECT() *MockurlStorageMockRecorder {
	return m.recorder
}

// DoDeleteURLTasks mocks base method.
func (m *MockurlStorage) DoDeleteURLTasks(ctx context.Context, tasks []domain.DeleteURLsTask) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DoDeleteURLTasks", ctx, tasks)
	ret0, _ := ret[0].(error)
	return ret0
}

// DoDeleteURLTasks indicates an expected call of DoDeleteURLTasks.
func (mr *MockurlStorageMockRecorder) DoDeleteURLTasks(ctx, tasks any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DoDeleteURLTasks", reflect.TypeOf((*MockurlStorage)(nil).DoDeleteURLTasks), ctx, tasks)
}

// Mocklogger is a mock of logger interface.
type Mocklogger struct {
	ctrl     *gomock.Controller
	recorder *MockloggerMockRecorder
}

// MockloggerMockRecorder is the mock recorder for Mocklogger.
type MockloggerMockRecorder struct {
	mock *Mocklogger
}

// NewMocklogger creates a new mock instance.
func NewMocklogger(ctrl *gomock.Controller) *Mocklogger {
	mock := &Mocklogger{ctrl: ctrl}
	mock.recorder = &MockloggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Mocklogger) EXPECT() *MockloggerMockRecorder {
	return m.recorder
}

// Info mocks base method.
func (m *Mocklogger) Info(msg string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Info", msg)
}

// Info indicates an expected call of Info.
func (mr *MockloggerMockRecorder) Info(msg any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*Mocklogger)(nil).Info), msg)
}
