package mocks

import (
	"net/http"

	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
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

// HandleTransactionIdExistsValidation mocks base method
func (m *MockHelperService) HandleTransactionIdExistsValidation(w http.ResponseWriter, req *http.Request, transactionID string) (string, bool, int) {
	ret := m.ctrl.Call(m, "HandleTransactionIdExistsValidation", w, req, transactionID)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(int)
	return ret0, ret1, ret2
}

// HandleTransactionIdExistsValidation indicates an expected call of HandleTransactionIdExistsValidation
func (mr *MockHelperServiceMockRecorder) HandleTransactionIdExistsValidation(w, req, transactionID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleTransactionIdExistsValidation", reflect.TypeOf((*MockHelperService)(nil).HandleTransactionIdExistsValidation), w, req, transactionID)
}

// HandleTransactionNotClosedValidation mocks base method
func (m *MockHelperService) HandleTransactionNotClosedValidation(w http.ResponseWriter, req *http.Request, transactionID string, isTransactionClosed bool, err error, httpStatus int) (error, bool, int) {
	ret := m.ctrl.Call(m, "HandleTransactionNotClosedValidation", w, req, transactionID, isTransactionClosed, err, httpStatus)
	ret0, _ := ret[0].(error)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[1].(int)
	return ret0, ret1, ret2
}

// HandleTransactionNotClosedValidation indicates an expected call of HandleTransactionNotClosedValidation
func (mr *MockHelperServiceMockRecorder) HandleTransactionNotClosedValidation(w, req, transactionID, isTransactionClosed, err, httpStatus interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleTransactionNotClosedValidation", reflect.TypeOf((*MockHelperService)(nil).HandleTransactionNotClosedValidation), w, req, transactionID, isTransactionClosed, err, httpStatus)
}

// HandleBodyDecodedValidation mocks base method
func (m *MockHelperService) HandleBodyDecodedValidation(w http.ResponseWriter, req *http.Request, transactionID string, err error) (bool, int) {
	ret := m.ctrl.Call(m, "HandleBodyDecodedValidation", w, req, transactionID, err)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(int)
	return ret0, ret1
}

// HandleBodyDecodedValidation indicates an expected call of HandleBodyDecodedValidation
func (mr *MockHelperServiceMockRecorder) HandleBodyDecodedValidation(http, req, transactionID, err interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleBodyDecodedValidation", reflect.TypeOf((*MockHelperService)(nil).HandleBodyDecodedValidation), http, req, transactionID, err)
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

// HandleCreateProgressReportResourceValidation mocks base method
func (m *MockHelperService) HandleCreateProgressReportResourceValidation(w http.ResponseWriter, req *http.Request, err error, statusCode int) (bool, int) {
	ret := m.ctrl.Call(m, "HandleCreateProgressReportResourceValidation", w, req, err, statusCode)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[0].(int)
	return ret0, ret1
}

// HandleCreateProgressReportResourceValidation indicates an expected call of HandleCreateProgressReportResourceValidation
func (mr *MockHelperServiceMockRecorder) HandleCreateProgressReportResourceValidation(w, req, err, statusCode interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleCreateProgressReportResourceValidation", reflect.TypeOf((*MockHelperService)(nil).HandleCreateProgressReportResourceValidation), w, req, err, statusCode)
}
