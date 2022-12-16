// Code generated by MockGen. DO NOT EDIT.
// Source: dao/service.go

package mocks

import (
	models "github.com/companieshouse/insolvency-api/models"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
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
	ret := m.ctrl.Call(m, "CreateInsolvencyResource", dao)
	ret0, _ := ret[0].(error)
	ret1, _ := ret[1].(int)
	return ret0, ret1
}

// CreateInsolvencyResource indicates an expected call of CreateInsolvencyResource
func (mr *MockServiceMockRecorder) CreateInsolvencyResource(dao interface{}) *gomock.Call {
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
	ret := m.ctrl.Call(m, "CreatePractitionersResource", dao, transactionID)
	ret0, _ := ret[0].(error)
	ret1, _ := ret[1].(int)
	return ret0, ret1
}

// CreatePractitionersResource indicates an expected call of CreatePractitionersResource
func (mr *MockServiceMockRecorder) CreatePractitionersResource(dao, transactionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePractitionersResource", reflect.TypeOf((*MockService)(nil).CreatePractitionersResource), dao, transactionID)
}

// GetPractitionerResources mocks base method
func (m *MockService) GetPractitionerResources(transactionID string) ([]models.PractitionerResourceDao, error) {
	ret := m.ctrl.Call(m, "GetPractitionerResources", transactionID)
	ret0, _ := ret[0].([]models.PractitionerResourceDao)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPractitionerResources indicates an expected call of GetPractitionerResources
func (mr *MockServiceMockRecorder) GetPractitionerResources(transactionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPractitionerResources", reflect.TypeOf((*MockService)(nil).GetPractitionerResources), transactionID)
}

// GetPractitionerResource mocks base method
func (m *MockService) GetPractitionerResource(practitionerID, transactionID string) (models.PractitionerResourceDao, error) {
	ret := m.ctrl.Call(m, "GetPractitionerResource", practitionerID, transactionID)
	ret0, _ := ret[0].(models.PractitionerResourceDao)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPractitionerResource indicates an expected call of GetPractitionerResource
func (mr *MockServiceMockRecorder) GetPractitionerResource(practitionerID, transactionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPractitionerResource", reflect.TypeOf((*MockService)(nil).GetPractitionerResource), practitionerID, transactionID)
}

// DeletePractitioner mocks base method
func (m *MockService) DeletePractitioner(practitionerID, transactionID string) (error, int) {
	ret := m.ctrl.Call(m, "DeletePractitioner", practitionerID, transactionID)
	ret0, _ := ret[0].(error)
	ret1, _ := ret[1].(int)
	return ret0, ret1
}

// DeletePractitioner indicates an expected call of DeletePractitioner
func (mr *MockServiceMockRecorder) DeletePractitioner(practitionerID, transactionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePractitioner", reflect.TypeOf((*MockService)(nil).DeletePractitioner), practitionerID, transactionID)
}

// AppointPractitioner mocks base method
func (m *MockService) AppointPractitioner(dao *models.AppointmentResourceDao, transactionID, practitionerID string) (error, int) {
	ret := m.ctrl.Call(m, "AppointPractitioner", dao, transactionID, practitionerID)
	ret0, _ := ret[0].(error)
	ret1, _ := ret[1].(int)
	return ret0, ret1
}

// AppointPractitioner indicates an expected call of AppointPractitioner
func (mr *MockServiceMockRecorder) AppointPractitioner(dao, transactionID, practitionerID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AppointPractitioner", reflect.TypeOf((*MockService)(nil).AppointPractitioner), dao, transactionID, practitionerID)
}

// DeletePractitionerAppointment mocks base method
func (m *MockService) DeletePractitionerAppointment(transactionID, practitionerID string) (error, int) {
	ret := m.ctrl.Call(m, "DeletePractitionerAppointment", transactionID, practitionerID)
	ret0, _ := ret[0].(error)
	ret1, _ := ret[1].(int)
	return ret0, ret1
}

// DeletePractitionerAppointment indicates an expected call of DeletePractitionerAppointment
func (mr *MockServiceMockRecorder) DeletePractitionerAppointment(transactionID, practitionerID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePractitionerAppointment", reflect.TypeOf((*MockService)(nil).DeletePractitionerAppointment), transactionID, practitionerID)
}

// AddAttachmentToInsolvencyResource mocks base method
func (m *MockService) AddAttachmentToInsolvencyResource(transactionID, fileID, attachmentType string) (*models.AttachmentResourceDao, error) {
	ret := m.ctrl.Call(m, "AddAttachmentToInsolvencyResource", transactionID, fileID, attachmentType)
	ret0, _ := ret[0].(*models.AttachmentResourceDao)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddAttachmentToInsolvencyResource indicates an expected call of AddAttachmentToInsolvencyResource
func (mr *MockServiceMockRecorder) AddAttachmentToInsolvencyResource(transactionID, fileID, attachmentType interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddAttachmentToInsolvencyResource", reflect.TypeOf((*MockService)(nil).AddAttachmentToInsolvencyResource), transactionID, fileID, attachmentType)
}

// GetAttachmentFromInsolvencyResource mocks base method
func (m *MockService) GetAttachmentFromInsolvencyResource(transactionID, attachmentID string) (models.AttachmentResourceDao, error) {
	ret := m.ctrl.Call(m, "GetAttachmentFromInsolvencyResource", transactionID, attachmentID)
	ret0, _ := ret[0].(models.AttachmentResourceDao)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAttachmentFromInsolvencyResource indicates an expected call of GetAttachmentFromInsolvencyResource
func (mr *MockServiceMockRecorder) GetAttachmentFromInsolvencyResource(transactionID, attachmentID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAttachmentFromInsolvencyResource", reflect.TypeOf((*MockService)(nil).GetAttachmentFromInsolvencyResource), transactionID, attachmentID)
}

// GetAttachmentResources mocks base method
func (m *MockService) GetAttachmentResources(transactionID string) ([]models.AttachmentResourceDao, error) {
	ret := m.ctrl.Call(m, "GetAttachmentResources", transactionID)
	ret0, _ := ret[0].([]models.AttachmentResourceDao)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAttachmentResources indicates an expected call of GetAttachmentResources
func (mr *MockServiceMockRecorder) GetAttachmentResources(transactionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAttachmentResources", reflect.TypeOf((*MockService)(nil).GetAttachmentResources), transactionID)
}

// DeleteAttachmentResource mocks base method
func (m *MockService) DeleteAttachmentResource(transactionID, attachmentID string) (int, error) {
	ret := m.ctrl.Call(m, "DeleteAttachmentResource", transactionID, attachmentID)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteAttachmentResource indicates an expected call of DeleteAttachmentResource
func (mr *MockServiceMockRecorder) DeleteAttachmentResource(transactionID, attachmentID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAttachmentResource", reflect.TypeOf((*MockService)(nil).DeleteAttachmentResource), transactionID, attachmentID)
}

// UpdateAttachmentStatus mocks base method
func (m *MockService) UpdateAttachmentStatus(transactionID, attachmentID, avStatus string) (int, error) {
	ret := m.ctrl.Call(m, "UpdateAttachmentStatus", transactionID, attachmentID, avStatus)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateAttachmentStatus indicates an expected call of UpdateAttachmentStatus
func (mr *MockServiceMockRecorder) UpdateAttachmentStatus(transactionID, attachmentID, avStatus interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAttachmentStatus", reflect.TypeOf((*MockService)(nil).UpdateAttachmentStatus), transactionID, attachmentID, avStatus)
}

// CreateStatementOfAffairsResource mocks base method
func (m *MockService) CreateStatementOfAffairsResource(dao *models.StatementOfAffairsResourceDao, transactionID string) (int, error) {
	ret := m.ctrl.Call(m, "CreateStatementOfAffairsResource", dao, transactionID)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateProgressReportResource mocks base method
func (m *MockService) CreateProgressReportResource(dao *models.ProgressReportResourceDao, transactionID string) (int, error) {
	ret := m.ctrl.Call(m, "CreateProgressReportResource", dao, transactionID)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateStatementOfAffairsResource indicates an expected call of CreateStatementOfAffairsResource
func (mr *MockServiceMockRecorder) CreateStatementOfAffairsResource(dao, transactionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateStatementOfAffairsResource", reflect.TypeOf((*MockService)(nil).CreateStatementOfAffairsResource), dao, transactionID)
}

// DeleteStatementOfAffairsResource mocks base method
func (m *MockService) DeleteStatementOfAffairsResource(transactionID string) (int, error) {
	ret := m.ctrl.Call(m, "DeleteStatementOfAffairsResource", transactionID)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteStatementOfAffairsResource indicates an expected call of DeleteStatementOfAffairsResource
func (mr *MockServiceMockRecorder) DeleteStatementOfAffairsResource(transactionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteStatementOfAffairsResource", reflect.TypeOf((*MockService)(nil).DeleteStatementOfAffairsResource), transactionID)
}

// CreateProgressReportResource indicates an expected call of CreateProgressReportResource
func (mr *MockServiceMockRecorder) CreateProgressReportResource(dao, transactionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateProgressReportResource", reflect.TypeOf((*MockService)(nil).CreateProgressReportResource), dao, transactionID)
}

// CreateResolutionResource mocks base method
func (m *MockService) CreateResolutionResource(dao *models.ResolutionResourceDao, transactionID string) (int, error) {
	ret := m.ctrl.Call(m, "CreateResolutionResource", dao, transactionID)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateResolutionResource indicates an expected call of CreateResolutionResource
func (mr *MockServiceMockRecorder) CreateResolutionResource(dao, transactionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateResolutionResource", reflect.TypeOf((*MockService)(nil).CreateResolutionResource), dao, transactionID)
}

// GetStatementOfAffairsResource mocks base method
func (m *MockService) GetStatementOfAffairsResource(transactionID string) (models.StatementOfAffairsResourceDao, error) {
	ret := m.ctrl.Call(m, "GetStatementOfAffairsResource", transactionID)
	ret0, _ := ret[0].(models.StatementOfAffairsResourceDao)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStatementOfAffairsResource indicates an expected call of GetStatementOfAffairsResource
func (mr *MockServiceMockRecorder) GetStatementOfAffairsResource(transactionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStatementOfAffairsResource", reflect.TypeOf((*MockService)(nil).GetStatementOfAffairsResource), transactionID)
}

// GetResolutionResource mocks base method
func (m *MockService) GetResolutionResource(transactionID string) (models.ResolutionResourceDao, error) {
	ret := m.ctrl.Call(m, "GetResolutionResource", transactionID)
	ret0, _ := ret[0].(models.ResolutionResourceDao)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetResolutionResource indicates an expected call of GetResolutionResource
func (mr *MockServiceMockRecorder) GetResolutionResource(transactionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetResolutionResource", reflect.TypeOf((*MockService)(nil).GetResolutionResource), transactionID)
}

// DeleteResolutionResource mocks base method
func (m *MockService) DeleteResolutionResource(transactionID string) (int, error) {
	ret := m.ctrl.Call(m, "DeleteResolutionResource", transactionID)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteResolutionResource indicates an expected call of DeleteResolutionResource
func (mr *MockServiceMockRecorder) DeleteResolutionResource(transactionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteResolutionResource", reflect.TypeOf((*MockService)(nil).DeleteResolutionResource), transactionID)
}
