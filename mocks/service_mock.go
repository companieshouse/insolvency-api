// Code generated by MockGen. DO NOT EDIT.
// Source: dao/service.go

// Package mock_dao is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	models "github.com/companieshouse/insolvency-api/models"
	gomock "github.com/golang/mock/gomock"
)

// MockService is a mock of Service interface
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
}

// MockServiceMockRecorder is the mock recorder for MockService
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// CreateInsolvencyResource mocks base method
func (m *MockService) CreateInsolvencyResource(dao *models.InsolvencyResourceDao) (error, int) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateInsolvencyResource", dao)
	ret0, _ := ret[0].(error)
	ret1, _ := ret[1].(int)
	return ret0, ret1
}

// CreateInsolvencyResource indicates an expected call of CreateInsolvencyResource
func (mr *MockServiceMockRecorder) CreateInsolvencyResource(dao interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateInsolvencyResource", reflect.TypeOf((*MockService)(nil).CreateInsolvencyResource), dao)
}

// GetInsolvencyResource mocks base method
func (m *MockService) GetInsolvencyResource(transactionID string) (models.InsolvencyResourceDao, error) {
	ret := m.ctrl.Call(m, "GetInsolvencyResource", transactionID)
	ret0, _ := ret[0].(models.InsolvencyResourceDao)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetInsolvencyResource indicates an expected call of GetInsolvencyResource
func (mr *MockServiceMockRecorder) GetInsolvencyResource(transactionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInsolvencyResource", reflect.TypeOf((*MockService)(nil).GetInsolvencyResource), transactionID)
}

// CreatePractitionersResource mocks base method
func (m *MockService) CreatePractitionersResource(dao *models.PractitionerResourceDao, transactionID string) (error, int) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatePractitionersResource", dao, transactionID)
	ret0, _ := ret[0].(error)
	ret1, _ := ret[1].(int)
	return ret0, ret1
}

// CreatePractitionersResource indicates an expected call of CreatePractitionersResource
func (mr *MockServiceMockRecorder) CreatePractitionersResource(dao, transactionID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePractitionersResource", reflect.TypeOf((*MockService)(nil).CreatePractitionersResource), dao, transactionID)
}

// GetPractitionerResources mocks base method
func (m *MockService) GetPractitionerResources(transactionID string) ([]models.PractitionerResourceDao, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPractitionerResources", transactionID)
	ret0, _ := ret[0].([]models.PractitionerResourceDao)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPractitionerResources indicates an expected call of GetPractitionerResources
func (mr *MockServiceMockRecorder) GetPractitionerResources(transactionID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPractitionerResources", reflect.TypeOf((*MockService)(nil).GetPractitionerResources), transactionID)
}

// GetPractitionerResource mocks base method
func (m *MockService) GetPractitionerResource(practitionerID, transactionID string) (models.PractitionerResourceDao, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPractitionerResource", practitionerID, transactionID)
	ret0, _ := ret[0].(models.PractitionerResourceDao)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPractitionerResource indicates an expected call of GetPractitionerResource
func (mr *MockServiceMockRecorder) GetPractitionerResource(practitionerID, transactionID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPractitionerResource", reflect.TypeOf((*MockService)(nil).GetPractitionerResource), practitionerID, transactionID)
}

// DeletePractitioner mocks base method
func (m *MockService) DeletePractitioner(practitionerID, transactionID string) (error, int) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeletePractitioner", practitionerID, transactionID)
	ret0, _ := ret[0].(error)
	ret1, _ := ret[1].(int)
	return ret0, ret1
}

// DeletePractitioner indicates an expected call of DeletePractitioner
func (mr *MockServiceMockRecorder) DeletePractitioner(practitionerID, transactionID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePractitioner", reflect.TypeOf((*MockService)(nil).DeletePractitioner), practitionerID, transactionID)
}

// AppointPractitioner mocks base method
func (m *MockService) AppointPractitioner(dao *models.AppointmentResourceDao, transactionID, practitionerID string) (error, int) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AppointPractitioner", dao, transactionID, practitionerID)
	ret0, _ := ret[0].(error)
	ret1, _ := ret[1].(int)
	return ret0, ret1
}

// AppointPractitioner indicates an expected call of AppointPractitioner
func (mr *MockServiceMockRecorder) AppointPractitioner(dao, transactionID, practitionerID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AppointPractitioner", reflect.TypeOf((*MockService)(nil).AppointPractitioner), dao, transactionID, practitionerID)
}

// DeletePractitionerAppointment mocks base method
func (m *MockService) DeletePractitionerAppointment(transactionID, practitionerID string) (error, int) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeletePractitionerAppointment", transactionID, practitionerID)
	ret0, _ := ret[0].(error)
	ret1, _ := ret[1].(int)
	return ret0, ret1
}

// DeletePractitionerAppointment indicates an expected call of DeletePractitionerAppointment
func (mr *MockServiceMockRecorder) DeletePractitionerAppointment(transactionID, practitionerID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePractitionerAppointment", reflect.TypeOf((*MockService)(nil).DeletePractitionerAppointment), transactionID, practitionerID)
}
