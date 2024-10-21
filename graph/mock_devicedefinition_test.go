// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/DIMO-Network/identity-api/graph (interfaces: DeviceDefinitionRepository)
//
// Generated by this command:
//
//	mockgen -destination=./mock_devicedefinition_test.go -package=graph github.com/DIMO-Network/identity-api/graph DeviceDefinitionRepository
//

// Package graph is a generated GoMock package.
package graph

import (
	context "context"
	reflect "reflect"

	model "github.com/DIMO-Network/identity-api/graph/model"
	gomock "go.uber.org/mock/gomock"
)

// MockDeviceDefinitionRepository is a mock of DeviceDefinitionRepository interface.
type MockDeviceDefinitionRepository struct {
	ctrl     *gomock.Controller
	recorder *MockDeviceDefinitionRepositoryMockRecorder
	isgomock struct{}
}

// MockDeviceDefinitionRepositoryMockRecorder is the mock recorder for MockDeviceDefinitionRepository.
type MockDeviceDefinitionRepositoryMockRecorder struct {
	mock *MockDeviceDefinitionRepository
}

// NewMockDeviceDefinitionRepository creates a new mock instance.
func NewMockDeviceDefinitionRepository(ctrl *gomock.Controller) *MockDeviceDefinitionRepository {
	mock := &MockDeviceDefinitionRepository{ctrl: ctrl}
	mock.recorder = &MockDeviceDefinitionRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDeviceDefinitionRepository) EXPECT() *MockDeviceDefinitionRepositoryMockRecorder {
	return m.recorder
}

// GetDeviceDefinition mocks base method.
func (m *MockDeviceDefinitionRepository) GetDeviceDefinition(ctx context.Context, by model.DeviceDefinitionBy) (*model.DeviceDefinition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDeviceDefinition", ctx, by)
	ret0, _ := ret[0].(*model.DeviceDefinition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDeviceDefinition indicates an expected call of GetDeviceDefinition.
func (mr *MockDeviceDefinitionRepositoryMockRecorder) GetDeviceDefinition(ctx, by any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDeviceDefinition", reflect.TypeOf((*MockDeviceDefinitionRepository)(nil).GetDeviceDefinition), ctx, by)
}

// GetDeviceDefinitions mocks base method.
func (m *MockDeviceDefinitionRepository) GetDeviceDefinitions(ctx context.Context, tableID, first *int, after *string, last *int, before *string, filterBy *model.DeviceDefinitionFilter) (*model.DeviceDefinitionConnection, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDeviceDefinitions", ctx, tableID, first, after, last, before, filterBy)
	ret0, _ := ret[0].(*model.DeviceDefinitionConnection)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDeviceDefinitions indicates an expected call of GetDeviceDefinitions.
func (mr *MockDeviceDefinitionRepositoryMockRecorder) GetDeviceDefinitions(ctx, tableID, first, after, last, before, filterBy any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDeviceDefinitions", reflect.TypeOf((*MockDeviceDefinitionRepository)(nil).GetDeviceDefinitions), ctx, tableID, first, after, last, before, filterBy)
}
