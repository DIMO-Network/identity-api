// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/DIMO-Network/identity-api/graph (interfaces: DeveloperLicenseRepository)
//
// Generated by this command:
//
//	mockgen -destination=./mock_developerlicense_test.go -package=graph github.com/DIMO-Network/identity-api/graph DeveloperLicenseRepository
//

// Package graph is a generated GoMock package.
package graph

import (
	context "context"
	reflect "reflect"

	model "github.com/DIMO-Network/identity-api/graph/model"
	gomock "go.uber.org/mock/gomock"
)

// MockDeveloperLicenseRepository is a mock of DeveloperLicenseRepository interface.
type MockDeveloperLicenseRepository struct {
	ctrl     *gomock.Controller
	recorder *MockDeveloperLicenseRepositoryMockRecorder
	isgomock struct{}
}

// MockDeveloperLicenseRepositoryMockRecorder is the mock recorder for MockDeveloperLicenseRepository.
type MockDeveloperLicenseRepositoryMockRecorder struct {
	mock *MockDeveloperLicenseRepository
}

// NewMockDeveloperLicenseRepository creates a new mock instance.
func NewMockDeveloperLicenseRepository(ctrl *gomock.Controller) *MockDeveloperLicenseRepository {
	mock := &MockDeveloperLicenseRepository{ctrl: ctrl}
	mock.recorder = &MockDeveloperLicenseRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDeveloperLicenseRepository) EXPECT() *MockDeveloperLicenseRepositoryMockRecorder {
	return m.recorder
}

// GetDeveloperLicenses mocks base method.
func (m *MockDeveloperLicenseRepository) GetDeveloperLicenses(ctx context.Context, first *int, after *string, last *int, before *string, filterBy *model.DeveloperLicenseFilterBy) (*model.DeveloperLicenseConnection, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDeveloperLicenses", ctx, first, after, last, before, filterBy)
	ret0, _ := ret[0].(*model.DeveloperLicenseConnection)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDeveloperLicenses indicates an expected call of GetDeveloperLicenses.
func (mr *MockDeveloperLicenseRepositoryMockRecorder) GetDeveloperLicenses(ctx, first, after, last, before, filterBy any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDeveloperLicenses", reflect.TypeOf((*MockDeveloperLicenseRepository)(nil).GetDeveloperLicenses), ctx, first, after, last, before, filterBy)
}

// GetLicense mocks base method.
func (m *MockDeveloperLicenseRepository) GetLicense(ctx context.Context, by model.DeveloperLicenseBy) (*model.DeveloperLicense, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLicense", ctx, by)
	ret0, _ := ret[0].(*model.DeveloperLicense)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLicense indicates an expected call of GetLicense.
func (mr *MockDeveloperLicenseRepositoryMockRecorder) GetLicense(ctx, by any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLicense", reflect.TypeOf((*MockDeveloperLicenseRepository)(nil).GetLicense), ctx, by)
}

// GetRedirectURIsForLicense mocks base method.
func (m *MockDeveloperLicenseRepository) GetRedirectURIsForLicense(ctx context.Context, obj *model.DeveloperLicense, first *int, after *string, last *int, before *string) (*model.RedirectURIConnection, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRedirectURIsForLicense", ctx, obj, first, after, last, before)
	ret0, _ := ret[0].(*model.RedirectURIConnection)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRedirectURIsForLicense indicates an expected call of GetRedirectURIsForLicense.
func (mr *MockDeveloperLicenseRepositoryMockRecorder) GetRedirectURIsForLicense(ctx, obj, first, after, last, before any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRedirectURIsForLicense", reflect.TypeOf((*MockDeveloperLicenseRepository)(nil).GetRedirectURIsForLicense), ctx, obj, first, after, last, before)
}

// GetSignersForLicense mocks base method.
func (m *MockDeveloperLicenseRepository) GetSignersForLicense(ctx context.Context, obj *model.DeveloperLicense, first *int, after *string, last *int, before *string) (*model.SignerConnection, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSignersForLicense", ctx, obj, first, after, last, before)
	ret0, _ := ret[0].(*model.SignerConnection)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSignersForLicense indicates an expected call of GetSignersForLicense.
func (mr *MockDeveloperLicenseRepositoryMockRecorder) GetSignersForLicense(ctx, obj, first, after, last, before any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSignersForLicense", reflect.TypeOf((*MockDeveloperLicenseRepository)(nil).GetSignersForLicense), ctx, obj, first, after, last, before)
}
