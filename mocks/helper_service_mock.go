package mocks

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jarcoal/httpmock"

	reflect "reflect"

	"github.com/companieshouse/insolvency-api/models"

	gomock "github.com/golang/mock/gomock"
)

// MockHelperService is a mock of Service interface
type MockHelperService struct {
	ctrl     *gomock.Controller
	recorder *MockHelperServiceMockRecorder
}

// MockHelperServiceMockRecorder is the mock recorder for MockHelperService
type MockHelperServiceMockRecorder struct {
	mock *MockHelperService
}

// NewHelperMockHelperService creates a new mock instance
func NewHelperMockHelperService(ctrl *gomock.Controller) *MockHelperService {
	mock := &MockHelperService{ctrl: ctrl}
	mock.recorder = &MockHelperServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHelperService) EXPECT() *MockHelperServiceMockRecorder {
	return m.recorder
}

// GenerateEtag mocks base method
func (m *MockHelperService) GenerateEtag() (string, error) {
	ret := m.ctrl.Call(m, "GenerateEtag")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateEtag indicates an expected call of GenerateEtag
func (mr *MockHelperServiceMockRecorder) GenerateEtag() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateEtag", reflect.TypeOf((*MockHelperService)(nil).GenerateEtag))
}

// HandleTransactionIdExistsValidation mocks base method
func (m *MockHelperService) HandleTransactionIdExistsValidation(w http.ResponseWriter, req *http.Request, transactionID string) (bool, string) {
	ret := m.ctrl.Call(m, "HandleTransactionIdExistsValidation", w, req, transactionID)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(string)
	return ret0, ret1
}

// HandleTransactionIdExistsValidation indicates an expected call of HandleTransactionIdExistsValidation
func (mr *MockHelperServiceMockRecorder) HandleTransactionIdExistsValidation(w, req, transactionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleTransactionIdExistsValidation", reflect.TypeOf((*MockHelperService)(nil).HandleTransactionIdExistsValidation), w, req, transactionID)
}

// HandleTransactionNotClosedValidation mocks base method
func (m *MockHelperService) HandleTransactionNotClosedValidation(w http.ResponseWriter, req *http.Request, transactionID string, isTransactionClosed bool, httpStatus int, err error) bool {
	ret := m.ctrl.Call(m, "HandleTransactionNotClosedValidation", w, req, transactionID, isTransactionClosed, err, httpStatus)
	ret0, _ := ret[0].(bool)
	return ret0
}

// HandleTransactionNotClosedValidation indicates an expected call of HandleTransactionNotClosedValidation
func (mr *MockHelperServiceMockRecorder) HandleTransactionNotClosedValidation(w, req, transactionID, isTransactionClosed, httpStatus, err interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleTransactionNotClosedValidation", reflect.TypeOf((*MockHelperService)(nil).HandleTransactionNotClosedValidation), w, req, transactionID, isTransactionClosed, httpStatus, err)
}

// HandleBodyDecodedValidation indicates an expected call of HandleBodyDecodedValidation
func (mr *MockHelperServiceMockRecorder) HandleBodyDecodedValidation(http, req, transactionID, err interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleBodyDecodedValidation", reflect.TypeOf((*MockHelperService)(nil).HandleBodyDecodedValidation), http, req, transactionID, err)
}

// HandleBodyDecodedValidation mocks base method
func (m *MockHelperService) HandleBodyDecodedValidation(w http.ResponseWriter, req *http.Request, transactionID string, err error) bool {
	ret := m.ctrl.Call(m, "HandleBodyDecodedValidation", w, req, transactionID, err)
	ret0, _ := ret[0].(bool)
	return ret0
}

// HandleAttachmentTypeValidation indicates an expected call of HandleAttachmentTypeValidation
func (mr *MockHelperServiceMockRecorder) HandleAttachmentTypeValidation(http, req, responseMessage, err interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleAttachmentTypeValidation", reflect.TypeOf((*MockHelperService)(nil).HandleAttachmentTypeValidation), http, req, responseMessage, err)
}

// HandleAttachmentTypeValidation mocks base method
func (m *MockHelperService) HandleAttachmentTypeValidation(w http.ResponseWriter, req *http.Request, responseMessage string, err error) int {
	ret := m.ctrl.Call(m, "HandleAttachmentTypeValidation", w, req, responseMessage, err)
	ret0, _ := ret[0].(int)
	return ret0
}

// HandleMandatoryFieldValidation indicates an expected call of HandleMandatoryFieldValidation
func (mr *MockHelperServiceMockRecorder) HandleMandatoryFieldValidation(http, req, errs, err interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleMandatoryFieldValidation", reflect.TypeOf((*MockHelperService)(nil).HandleMandatoryFieldValidation), http, req, errs)
}

// HandleMandatoryFieldValidation mocks base method
func (m *MockHelperService) HandleMandatoryFieldValidation(w http.ResponseWriter, req *http.Request, errs string) bool {
	ret := m.ctrl.Call(m, "HandleMandatoryFieldValidation", w, req, errs)
	ret0, _ := ret[0].(bool)
	return ret0
}

// HandleAttachmentValidation indicates an expected call of HandleAttachmentValidation
func (mr *MockHelperServiceMockRecorder) HandleAttachmentValidation(http, req, transactionID, dao, err interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleAttachmentValidation", reflect.TypeOf((*MockHelperService)(nil).HandleAttachmentValidation), http, req, transactionID, dao, err)
}

// HandleAttachmentValidation mocks base method
func (m *MockHelperService) HandleAttachmentValidation(w http.ResponseWriter, req *http.Request, transactionID string, attachment models.AttachmentResourceDao, err error) bool {
	ret := m.ctrl.Call(m, "HandleAttachmentValidation", w, req, transactionID, attachment, err)
	ret0, _ := ret[0].(bool)
	return ret0
}

// HandleEtagGenerationValidation mocks base method
func (m *MockHelperService) HandleEtagGenerationValidation(err error) bool {
	ret := m.ctrl.Call(m, "HandleEtagGenerationValidation", err)
	ret0, _ := ret[0].(bool)
	return ret0
}

// HandleEtagGenerationValidation indicates an expected call of HandleEtagGenerationValidation
func (mr *MockHelperServiceMockRecorder) HandleEtagGenerationValidation(err interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleEtagGenerationValidation", reflect.TypeOf((*MockHelperService)(nil).HandleEtagGenerationValidation), err)
}

// HandleCreateResourceValidation indicates an expected call of HandleCreateResourceValidation
func (mr *MockHelperServiceMockRecorder) HandleCreateResourceValidation(w, req, httpStatus, err interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleCreateResourceValidation", reflect.TypeOf((*MockHelperService)(nil).HandleCreateResourceValidation), w, req, httpStatus, err)
}

// HandleCreateResourceValidation mocks base method
func (m *MockHelperService) HandleCreateResourceValidation(w http.ResponseWriter, req *http.Request, httpStatus int, err error) bool {
	ret := m.ctrl.Call(m, "HandleCreateResourceValidation", w, req, err, httpStatus)
	ret0, _ := ret[0].(bool)
	return ret0
}

// HandleCreateResourceValidation indicates an expected call of HandleCreateResourceValidation
func (mr *MockHelperServiceMockRecorder) HandleDeleteResourceValidation(w, req, resourceType interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleDeleteResourceValidation", reflect.TypeOf((*MockHelperService)(nil).HandleDeleteResourceValidation), w, req, resourceType)
}

// HandleCreateResourceValidation mocks base method
func (m *MockHelperService) HandleDeleteResourceValidation(w http.ResponseWriter, req *http.Request, resourceType string) (bool, string) {
	ret := m.ctrl.Call(m, "HandleCreateResourceValidation", w, req, resourceType)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(string)
	return ret0, ret1
}

func CreateTestObjects(t *testing.T) (*MockService, *MockHelperService, *httptest.ResponseRecorder) {
	defer httpmock.DeactivateAndReset()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	return NewMockService(mockCtrl), NewHelperMockHelperService(mockCtrl), httptest.NewRecorder()
}
