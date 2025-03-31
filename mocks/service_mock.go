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

// GetPractitionerAppointment mocks base method
func (m *MockService) GetPractitionerAppointment(transactionID string, practitionerID string) (*models.AppointmentResourceDao, error) {
	ret := m.ctrl.Call(m, "GetPractitionerAppointment", transactionID, practitionerID)
	ret0, _ := ret[0].(*models.AppointmentResourceDao)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPractitionerAppointment indicates an expected call of GetPractitionerAppointment
func (mr *MockServiceMockRecorder) GetPractitionerAppointment(transactionID, practitionerID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPractitionerAppointment", reflect.TypeOf((*MockService)(nil).GetPractitionerAppointment), transactionID, practitionerID)
}

// CreateInsolvencyResource mocks base methodÂ
func (m *MockService) CreateInsolvencyResource(dao *models.InsolvencyResourceDao) (int, error) {
	ret := m.ctrl.Call(m, "CreateInsolvencyResource", dao)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateInsolvencyResource indicates an expected call of CreateInsolvencyResource
func (mr *MockServiceMockRecorder) CreateInsolvencyResource(dao interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateInsolvencyResource", reflect.TypeOf((*MockService)(nil).CreateInsolvencyResource), dao)
}

// GetInsolvencyResource mocks base method
func (m *MockService) GetInsolvencyResource(transactionID string) (*models.InsolvencyResourceDao, error) {
	ret := m.ctrl.Call(m, "GetInsolvencyResource", transactionID)
	ret0, _ := ret[0].(*models.InsolvencyResourceDao)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetInsolvencyResource indicates an expected call of GetInsolvencyResource
func (mr *MockServiceMockRecorder) GetInsolvencyResource(transactionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInsolvencyResource", reflect.TypeOf((*MockService)(nil).GetInsolvencyResource), transactionID)
}

// GetInsolvencyAndExpandedPractitionerResources mocks base method
func (m *MockService) GetInsolvencyAndExpandedPractitionerResources(transactionID string) (*models.InsolvencyResourceDao,[]models.PractitionerResourceDao, error) {
	ret := m.ctrl.Call(m, "GetInsolvencyAndExpandedPractitionerResources", transactionID)
	ret0, _ := ret[0].(*models.InsolvencyResourceDao)
	ret1, _ := ret[1].([]models.PractitionerResourceDao)
	ret2, _ := ret[2].(error)
	return ret0, ret1,ret2
}

// GetInsolvencyAndExpandedPractitionerResources indicates an expected call of GetInsolvencyAndExpandedPractitionerResources
func (mr *MockServiceMockRecorder) GetInsolvencyAndExpandedPractitionerResources(transactionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInsolvencyAndExpandedPractitionerResources", reflect.TypeOf((*MockService)(nil).GetInsolvencyAndExpandedPractitionerResources), transactionID)
}

// CreatePractitionerResource mocks base method
func (m *MockService) CreatePractitionerResource(practitionerResourceDto *models.PractitionerResourceDao, transactionID string) (int, error) {
	ret := m.ctrl.Call(m, "CreatePractitionerResource", practitionerResourceDto, transactionID)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreatePractitionerResource indicates an expected call of CreatePractitionerResource
func (mr *MockServiceMockRecorder) CreatePractitionerResource(practitionerResourceDto, transactionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePractitionerResource", reflect.TypeOf((*MockService)(nil).CreatePractitionerResource), practitionerResourceDto, transactionID)
}

// CreateAppointmentResource mocks base method.
func (m *MockService) CreateAppointmentResource(dao *models.AppointmentResourceDao) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateAppointmentResource",dao)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0,ret1
}

// CreateAppointmentResource indicates an expected call of CreateAppointmentResource.
func (mr *MockServiceMockRecorder) CreateAppointmentResource(transactionInterface interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateAppointmentResource", reflect.TypeOf((*MockService)(nil).CreateAppointmentResource), transactionInterface)
}
 
// AddPractitionerToInsolvencyResource mocks base method
func (m *MockService) AddPractitionerToInsolvencyResource(transactionID, practitionerID, practitionerLink string, ) (int, error)  {
	ret := m.ctrl.Call(m, "AddPractitionerToInsolvencyResource", transactionID, practitionerID, practitionerLink)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddPractitionerToInsolvencyResource indicates an expected call of AddPractitionerToInsolvencyResource
func (mr *MockServiceMockRecorder) AddPractitionerToInsolvencyResource(transactionID, practitionerID, practitionerLink interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddPractitionerToInsolvencyResource", reflect.TypeOf((*MockService)(nil).AddPractitionerToInsolvencyResource), transactionID, practitionerID, practitionerLink)
}

// GetSinglePractitionerResource mocks base method
func (m *MockService) GetSinglePractitionerResource(transactionID, practitionerID string) (*models.PractitionerResourceDao, error)  {
	ret := m.ctrl.Call(m, "GetSinglePractitionerResource", transactionID, practitionerID)
	ret0, _ := ret[0].(*models.PractitionerResourceDao)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSinglePractitionerResource indicates an expected call of GetSinglePractitionerResource
func (mr *MockServiceMockRecorder) GetSinglePractitionerResource(transactionID, practitionerID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSinglePractitionerResource", reflect.TypeOf((*MockService)(nil).GetSinglePractitionerResource), transactionID, practitionerID)
}

// GetAllPractitionerResourcesForTransactionID mocks base method
func (m *MockService) GetAllPractitionerResourcesForTransactionID(transactionID string) ([]models.PractitionerResourceDao, error)  {
	ret := m.ctrl.Call(m, "GetAllPractitionerResourcesForTransactionID", transactionID)
	ret0, _ := ret[0].([]models.PractitionerResourceDao)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllPractitionerResourcesForTransactionID indicates an expected call of GetAllPractitionerResourcesForTransactionID
func (mr *MockServiceMockRecorder) GetAllPractitionerResourcesForTransactionID(transactionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllPractitionerResourcesForTransactionID", reflect.TypeOf((*MockService)(nil).GetAllPractitionerResourcesForTransactionID), transactionID)
}

// DeletePractitioner mocks base method
func (m *MockService) DeletePractitioner(transactionID, practitionerID string) (int, error) {
	ret := m.ctrl.Call(m, "DeletePractitioner", transactionID, practitionerID)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeletePractitioner indicates an expected call of DeletePractitioner
func (mr *MockServiceMockRecorder) DeletePractitioner(transactionID, practitionerID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePractitioner", reflect.TypeOf((*MockService)(nil).DeletePractitioner), transactionID, practitionerID)
}

// UpdatePractitionerAppointment mocks base method
func (m *MockService) UpdatePractitionerAppointment(appointmentResourceDto *models.AppointmentResourceDao,transactionID, practitionerID string) (int, error) {
	ret := m.ctrl.Call(m, "UpdatePractitionerAppointment",appointmentResourceDto,transactionID, practitionerID)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdatePractitionerAppointment indicates an expected call of UpdatePractitionerAppointment
func (mr *MockServiceMockRecorder) UpdatePractitionerAppointment(appointmentResourceDto,transactionID, practitionerID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePractitionerAppointment", reflect.TypeOf((*MockService)(nil).UpdatePractitionerAppointment),appointmentResourceDto,transactionID, practitionerID)
}

 
// DeletePractitionerAppointment mocks base method
func (m *MockService) DeletePractitionerAppointment(transactionID, practitionerID, etag string) (int, error) {
	ret := m.ctrl.Call(m, "DeletePractitionerAppointment", transactionID, practitionerID, etag)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeletePractitionerAppointment indicates an expected call of DeletePractitionerAppointment
func (mr *MockServiceMockRecorder) DeletePractitionerAppointment(transactionID, practitionerID, etag interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePractitionerAppointment", reflect.TypeOf((*MockService)(nil).DeletePractitionerAppointment), transactionID, practitionerID, etag)
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

// GetProgressReportResource mocks base method
func (m *MockService) GetProgressReportResource(transactionID string) (*models.ProgressReportResourceDao, error) {
	ret := m.ctrl.Call(m, "GetProgressReportResource", transactionID)
	ret0, _ := ret[0].(*models.ProgressReportResourceDao)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProgressReportResource indicates an expected call of GetProgressReportResource
func (mr *MockServiceMockRecorder) GetProgressReportResource(transactionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProgressReportResource", reflect.TypeOf((*MockService)(nil).GetProgressReportResource), transactionID)
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

// DeleteProgressReportResource
func (m *MockService) DeleteProgressReportResource(transactionID string) (int, error) {
	ret := m.ctrl.Call(m, "DeleteProgressReportResource", transactionID)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteProgressReportResource indicates an expected call of DeleteProgressReportResource
func (mr *MockServiceMockRecorder) DeleteProgressReportResource(transactionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteProgressReportResource", reflect.TypeOf((*MockService)(nil).DeleteProgressReportResource), transactionID)
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
