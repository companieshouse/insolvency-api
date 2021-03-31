package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/go-session-handler/httpsession"
	"github.com/companieshouse/go-session-handler/session"
	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/dao"
	mock_dao "github.com/companieshouse/insolvency-api/mocks"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

const companyName = "companyName"
const companyNumber = "01234567"
const transactionID = "12345678"

var companyProfileResponse = `
{
 "company_name": "` + companyName + `",
 "company_number": "` + companyNumber + `",
 "jurisdiction": "england-wales",
 "company_status": "active",
 "type": "private-shares-exemption-30",
 "registered_office_address" : {
   "postal_code" : "CF14 3UZ",
   "address_line_2" : "Cardiff",
   "address_line_1" : "1 Crown Way"
  }
}
`

func serveHandleCreateInsolvencyResource(body []byte, service dao.Service, tranIdSet bool) *httptest.ResponseRecorder {

	ctx := context.WithValue(context.Background(), httpsession.ContextKeySession, &session.Session{})
	handler := HandleCreateInsolvencyResource(service)

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body)).WithContext(ctx)

	if tranIdSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}

	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleCreateInsolvencyResource(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	Convey("Must need a transaction ID in the url", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		res := serveHandleCreateInsolvencyResource(body, mock_dao.NewMockService(mockCtrl), false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Failed to read request body", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body := []byte(`{"company_name":error`)
		res := serveHandleCreateInsolvencyResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	// TODO: Unit tests when checking company profile API for company

	Convey("Incoming request has company number missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:    constants.MVL.String(),
			CompanyName: companyName,
		})
		res := serveHandleCreateInsolvencyResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "company_number is a required field")
	})

	Convey("Incoming request has company name missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:      constants.MVL.String(),
			CompanyNumber: companyNumber,
		})
		res := serveHandleCreateInsolvencyResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "company_name is a required field")
	})

	Convey("Incoming request has case type missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CompanyNumber: companyNumber,
			CompanyName:   companyName,
		})
		res := serveHandleCreateInsolvencyResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "case_type is a required field")
	})

	Convey("Incoming case type is not CVL", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:      constants.MVL.String(),
			CompanyNumber: companyNumber,
			CompanyName:   companyName,
		})
		res := serveHandleCreateInsolvencyResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error calling company-profile-api for company details", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		// Expect the company profile api to be called and return a company not found
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/01234567", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

		mockService := mock_dao.NewMockService(mockCtrl)

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:      constants.CVL.String(),
			CompanyName:   companyName,
			CompanyNumber: companyNumber,
		})
		res := serveHandleCreateInsolvencyResource(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Company marked for insolvency isn't found", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		// Expect the company profile api to be called and return a company not found
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/01234567", httpmock.NewStringResponder(http.StatusNotFound, ""))

		mockService := mock_dao.NewMockService(mockCtrl)

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:      constants.CVL.String(),
			CompanyName:   companyName,
			CompanyNumber: companyNumber,
		})
		res := serveHandleCreateInsolvencyResource(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Error adding insolvency resource to mongo", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		// Expect the company profile api to be called and return a valid company
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/01234567", httpmock.NewStringResponder(http.StatusOK, companyProfileResponse))

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect CreateInsolvencyResource to be called once and return an error
		mockService.EXPECT().CreateInsolvencyResource(gomock.Any()).Return(errors.New("error when creating mongo resource")).Times(1)

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:      constants.CVL.String(),
			CompanyName:   companyName,
			CompanyNumber: companyNumber,
		})
		res := serveHandleCreateInsolvencyResource(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Successfully add insolvency resource to mongo", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		// Expect the company profile api to be called and return a valid company
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/01234567", httpmock.NewStringResponder(http.StatusOK, companyProfileResponse))

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect CreateInsolvencyResource to be called once and not return an error
		mockService.EXPECT().CreateInsolvencyResource(gomock.Any()).Return(nil).Times(1)

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:      constants.CVL.String(),
			CompanyName:   companyName,
			CompanyNumber: companyNumber,
		})
		res := serveHandleCreateInsolvencyResource(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusCreated)
		// TODO: Check call to transaction API to update transaction resource
	})
}
