// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/DIMO-Network/identity-api/graph (interfaces: DCNRepository)
//
// Generated by this command:
//
//	mockgen -destination=./mock_dcn_test.go -package=graph github.com/DIMO-Network/identity-api/graph DCNRepository
//

// Package graph is a generated GoMock package.
package graph

import (
	context "context"
	reflect "reflect"

	model "github.com/DIMO-Network/identity-api/graph/model"
	gomock "go.uber.org/mock/gomock"
)

// MockDCNRepository is a mock of DCNRepository interface.
type MockDCNRepository struct {
	ctrl     *gomock.Controller
	recorder *MockDCNRepositoryMockRecorder
	isgomock struct{}
}

// MockDCNRepositoryMockRecorder is the mock recorder for MockDCNRepository.
type MockDCNRepositoryMockRecorder struct {
	mock *MockDCNRepository
}

// NewMockDCNRepository creates a new mock instance.
func NewMockDCNRepository(ctrl *gomock.Controller) *MockDCNRepository {
	mock := &MockDCNRepository{ctrl: ctrl}
	mock.recorder = &MockDCNRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDCNRepository) EXPECT() *MockDCNRepositoryMockRecorder {
	return m.recorder
}

// GetDCN mocks base method.
func (m *MockDCNRepository) GetDCN(ctx context.Context, by model.DCNBy) (*model.Dcn, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDCN", ctx, by)
	ret0, _ := ret[0].(*model.Dcn)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDCN indicates an expected call of GetDCN.
func (mr *MockDCNRepositoryMockRecorder) GetDCN(ctx, by any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDCN", reflect.TypeOf((*MockDCNRepository)(nil).GetDCN), ctx, by)
}

// GetDCNByName mocks base method.
func (m *MockDCNRepository) GetDCNByName(ctx context.Context, name string) (*model.Dcn, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDCNByName", ctx, name)
	ret0, _ := ret[0].(*model.Dcn)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDCNByName indicates an expected call of GetDCNByName.
func (mr *MockDCNRepositoryMockRecorder) GetDCNByName(ctx, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDCNByName", reflect.TypeOf((*MockDCNRepository)(nil).GetDCNByName), ctx, name)
}

// GetDCNByNode mocks base method.
func (m *MockDCNRepository) GetDCNByNode(ctx context.Context, node []byte) (*model.Dcn, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDCNByNode", ctx, node)
	ret0, _ := ret[0].(*model.Dcn)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDCNByNode indicates an expected call of GetDCNByNode.
func (mr *MockDCNRepositoryMockRecorder) GetDCNByNode(ctx, node any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDCNByNode", reflect.TypeOf((*MockDCNRepository)(nil).GetDCNByNode), ctx, node)
}

// GetDCNs mocks base method.
func (m *MockDCNRepository) GetDCNs(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *model.DCNFilter) (*model.DCNConnection, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDCNs", ctx, first, after, last, before, filterBy)
	ret0, _ := ret[0].(*model.DCNConnection)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDCNs indicates an expected call of GetDCNs.
func (mr *MockDCNRepositoryMockRecorder) GetDCNs(ctx, first, after, last, before, filterBy any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDCNs", reflect.TypeOf((*MockDCNRepository)(nil).GetDCNs), ctx, first, after, last, before, filterBy)
}
