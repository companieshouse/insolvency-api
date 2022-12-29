package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/dao"
	mock_dao "github.com/companieshouse/insolvency-api/mocks"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/jarcoal/httpmock"

	. "github.com/smartystreets/goconvey/convey"
)

func serveHandleCreateProgressReport(body []byte, service dao.Service, helperService utils.HelperService, tranIDSet bool) *httptest.ResponseRecorder {
	path := "/transactions/123456789/insolvency/progress-report"
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	if tranIDSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}
	res := httptest.NewRecorder()

	handler := HandleCreateProgressReport(service, helperService)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleCreateProgressReport(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockService := mock_dao.NewMockService(mockCtrl)
	mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

	Convey("Must need a transaction ID in the url", t, func() {
		body, _ := json.Marshal(&models.InsolvencyRequest{})

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), "").Return("", false, http.StatusBadRequest).AnyTimes()

		res := serveHandleCreateProgressReport(body, mockService, mockHelperService, false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error checking if transaction is closed against transaction api", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an error
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

		body, _ := json.Marshal(&models.InsolvencyRequest{})

		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, http.StatusInternalServerError, nil).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusInternalServerError).AnyTimes()

		res := serveHandleCreateProgressReport(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Transaction is already closed and cannot be updated", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mockService := mock_dao.NewMockService(mockCtrl)
		mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

		// Expect the transaction api to be called and return an already closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		body, _ := json.Marshal(&models.InsolvencyRequest{})

		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusForbidden, nil).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusInternalServerError).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, http.StatusForbidden).AnyTimes()

		res := serveHandleCreateProgressReport(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusForbidden)
	})

	Convey("Failed to read request body", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		body := []byte(`{"first_name":error`)

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusInternalServerError).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusInternalServerError, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, http.StatusInternalServerError).AnyTimes()

		res := serveHandleCreateProgressReport(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Generic error when adding progress report resource to mongo", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		mockService := mock_dao.NewMockService(mockCtrl)
		mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		statement := generateProgressReport()
		body, _ := json.Marshal(statement)

		attachment := generateAttachment()
		attachment.Type = "progress-report"

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusInternalServerError).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusForbidden, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusInternalServerError).AnyTimes()
		mockService.EXPECT().CreateProgressReportResource(gomock.Any(), transactionID).Return(http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID))
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()
		mockHelperService.EXPECT().HandleEtagGenerationValidation(gomock.Any()).Return(false).AnyTimes()

		res := serveHandleCreateProgressReport(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Error adding progress report resource to mongo - insolvency case not found", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		mockService := mock_dao.NewMockService(mockCtrl)
		mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		statement := generateStatement()
		body, _ := json.Marshal(statement)

		attachment := generateAttachment()
		attachment.Type = "progress-report"

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusInternalServerError).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusForbidden, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusInternalServerError).AnyTimes()
		mockService.EXPECT().CreateProgressReportResource(gomock.Any(), transactionID).Return(http.StatusNotFound, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID))
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()
		mockHelperService.EXPECT().HandleEtagGenerationValidation(gomock.Any()).Return(false).AnyTimes()

		res := serveHandleCreateProgressReport(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Successfully add insolvency resource to mongo", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mockService := mock_dao.NewMockService(mockCtrl)
		mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		statement := generateProgressReport()
		body, _ := json.Marshal(statement)

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusInternalServerError).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusForbidden, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusInternalServerError).AnyTimes()
		mockService.EXPECT().CreateProgressReportResource(gomock.Any(), transactionID).Return(http.StatusOK, nil)
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()
		mockHelperService.EXPECT().HandleEtagGenerationValidation(gomock.Any()).Return(false).AnyTimes()
		mockHelperService.EXPECT().HandleCreateResourceValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, http.StatusOK).AnyTimes()

		res := serveHandleCreateProgressReport(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusOK)
	})
}

func generateProgressReport() models.ProgressReport {
	return models.ProgressReport{
		FromDate: "2021-06-06",
		ToDate:   "2021-06-07",
		Attachments: []string{
			"123456789",
		},
	}
}
