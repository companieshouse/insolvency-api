package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/dao"
	mock_dao "github.com/companieshouse/insolvency-api/mocks"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"

	. "github.com/smartystreets/goconvey/convey"
)

func serveHandleCreateProgressReport(body []byte, service dao.Service, tranIDSet bool) *httptest.ResponseRecorder {
	path := "/transactions/123456789/insolvency/progress-report"
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	if tranIDSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}
	res := httptest.NewRecorder()

	handler := HandleCreateProgressReport(service)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleCreateProgressReport(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	Convey("Must need a transaction ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		res := serveHandleCreateProgressReport(body, mock_dao.NewMockService(mockCtrl), false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error checking if transaction is closed against transaction api", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		// Expect the transaction api to be called and return an error
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		res := serveHandleCreateProgressReport(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Transaction is already closed and cannot be updated", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		// Expect the transaction api to be called and return an already closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		res := serveHandleCreateProgressReport(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusForbidden)
	})

	Convey("Failed to read request body", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		body := []byte(`{"first_name":error`)
		res := serveHandleCreateProgressReport(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Successfully add insolvency resource to mongo", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		mockService := mock_dao.NewMockService(mockCtrl)

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		statement := generateProgressReport()
		body, _ := json.Marshal(statement)

		// Expect CreateStatementOfAffairsResource to be called once and return an error
		mockService.EXPECT().CreateProgressReportResource(gomock.Any(), transactionID).Return(http.StatusCreated, nil).Times(1)

		res := serveHandleCreateProgressReport(body, mockService, true)

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
