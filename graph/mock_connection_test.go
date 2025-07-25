// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/DIMO-Network/identity-api/graph (interfaces: ConnectionRepository)
//
// Generated by this command:
//
//	mockgen -destination=./mock_connection_test.go -package=graph github.com/DIMO-Network/identity-api/graph ConnectionRepository
//

// Package graph is a generated GoMock package.
package graph

import (
	context "context"
	reflect "reflect"

	model "github.com/DIMO-Network/identity-api/graph/model"
	gomock "go.uber.org/mock/gomock"
)

// MockConnectionRepository is a mock of ConnectionRepository interface.
type MockConnectionRepository struct {
	ctrl     *gomock.Controller
	recorder *MockConnectionRepositoryMockRecorder
	isgomock struct{}
}

// MockConnectionRepositoryMockRecorder is the mock recorder for MockConnectionRepository.
type MockConnectionRepositoryMockRecorder struct {
	mock *MockConnectionRepository
}

// NewMockConnectionRepository creates a new mock instance.
func NewMockConnectionRepository(ctrl *gomock.Controller) *MockConnectionRepository {
	mock := &MockConnectionRepository{ctrl: ctrl}
	mock.recorder = &MockConnectionRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConnectionRepository) EXPECT() *MockConnectionRepositoryMockRecorder {
	return m.recorder
}

// GetConnection mocks base method.
func (m *MockConnectionRepository) GetConnection(ctx context.Context, by model.ConnectionBy) (*model.Connection, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetConnection", ctx, by)
	ret0, _ := ret[0].(*model.Connection)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetConnection indicates an expected call of GetConnection.
func (mr *MockConnectionRepositoryMockRecorder) GetConnection(ctx, by any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetConnection", reflect.TypeOf((*MockConnectionRepository)(nil).GetConnection), ctx, by)
}

// GetConnections mocks base method.
func (m *MockConnectionRepository) GetConnections(ctx context.Context, first *int, after *string, last *int, before *string) (*model.ConnectionConnection, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetConnections", ctx, first, after, last, before)
	ret0, _ := ret[0].(*model.ConnectionConnection)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetConnections indicates an expected call of GetConnections.
func (mr *MockConnectionRepositoryMockRecorder) GetConnections(ctx, first, after, last, before any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetConnections", reflect.TypeOf((*MockConnectionRepository)(nil).GetConnections), ctx, first, after, last, before)
}
