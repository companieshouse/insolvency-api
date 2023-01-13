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

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/go-session-handler/httpsession"
	"github.com/companieshouse/go-session-handler/session"
	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/dao"
	mock_dao "github.com/companieshouse/insolvency-api/mocks"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
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

var transactionProfileResponse = `
{
 "company_name": "` + companyName + `",
 "company_number": "` + companyNumber + `",
 "id": "` + transactionID + `",
 "status": "open"
}
`

var transactionProfileResponseClosed = `
{
 "status": "closed"
}
`
var alphakeyResponse = `
{
	"sameAsAlphaKey": "COMPANYNAME",
	"orderedAlphaKey": "COMPANYNAME",
	"upperCaseName": "COMPANYNAME"
}
`

func serveHandleCreateInsolvencyResource(body []byte, service dao.Service, tranIDSet bool, helperService utils.HelperService) *httptest.ResponseRecorder {
	ctx := context.WithValue(context.Background(), httpsession.ContextKeySession, &session.Session{})
	handler := HandleCreateInsolvencyResource(service, helperService)

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body)).WithContext(ctx)

	if tranIDSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}

	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleCreateInsolvencyResource(t *testing.T) {
	InTest = true

	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	Convey("Must need a transaction ID in the url", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), "").Return("", false, http.StatusBadRequest).AnyTimes()

		res := serveHandleCreateInsolvencyResource(body, mockService, false, mockHelperService)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Failed to read request body", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		body := []byte(`{"company_name":error`)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusInternalServerError).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, http.StatusBadRequest).AnyTimes()

		res := serveHandleCreateInsolvencyResource(body, mockService, true, mockHelperService)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Incoming request has company number missing", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:    constants.MVL.String(),
			CompanyName: companyName,
		})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, http.StatusBadRequest).AnyTimes()

		res := serveHandleCreateInsolvencyResource(body, mockService, true, mockHelperService)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "company_number is a required field")
	})

	Convey("Incoming request has company name missing", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:      constants.MVL.String(),
			CompanyNumber: companyNumber,
		})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, http.StatusBadRequest).AnyTimes()

		res := serveHandleCreateInsolvencyResource(body, mockService, true, mockHelperService)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "company_name is a required field")
	})

	Convey("Incoming request has case type missing", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CompanyNumber: companyNumber,
			CompanyName:   companyName,
		})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, http.StatusBadRequest).AnyTimes()

		res := serveHandleCreateInsolvencyResource(body, mockService, true, mockHelperService)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "case_type is a required field")
	})

	Convey("Incoming case type is not CVL", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:      constants.MVL.String(),
			CompanyNumber: companyNumber,
			CompanyName:   companyName,
		})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, http.StatusBadRequest).AnyTimes()

		res := serveHandleCreateInsolvencyResource(body, mockService, true, mockHelperService)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error calling transaction-api when checking transaction exists", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return a valid transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

		// Expect the alphakeyservice api to be called and return an alphakey
		httpmock.RegisterResponder(http.MethodGet, "http://localhost:18103/alphakey?name=companyName", httpmock.NewStringResponder(http.StatusOK, alphakeyResponse))

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:      constants.CVL.String(),
			CompanyName:   companyName,
			CompanyNumber: companyNumber,
		})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()

		res := serveHandleCreateInsolvencyResource(body, mockService, true, mockHelperService)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Transaction marked for insolvency isn't found", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return a valid transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusNotFound, transactionProfileResponse))

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:      constants.CVL.String(),
			CompanyName:   companyName,
			CompanyNumber: companyNumber,
		})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()

		res := serveHandleCreateInsolvencyResource(body, mockService, true, mockHelperService)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Error calling company-profile-api for company details", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return a valid transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		// Expect the company profile api to be called and return a company not found
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/01234567", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:      constants.CVL.String(),
			CompanyName:   companyName,
			CompanyNumber: companyNumber,
		})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()

		res := serveHandleCreateInsolvencyResource(body, mockService, true, mockHelperService)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Company marked for insolvency isn't found", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return a valid transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		// Expect the company profile api to be called and return a company not found
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/01234567", httpmock.NewStringResponder(http.StatusNotFound, ""))

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:      constants.CVL.String(),
			CompanyName:   companyName,
			CompanyNumber: companyNumber,
		})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()

		res := serveHandleCreateInsolvencyResource(body, mockService, true, mockHelperService)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Insolvency case already exists for transaction ID", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return a valid transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		// Expect the company profile api to be called and return a valid company
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/01234567", httpmock.NewStringResponder(http.StatusOK, companyProfileResponse))

		// Expect the alphakeyservice api to be called and return an alphakey
		httpmock.RegisterResponder(http.MethodGet, "http://localhost:18103/alphakey?name=companyName", httpmock.NewStringResponder(http.StatusOK, alphakeyResponse))

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:      constants.CVL.String(),
			CompanyName:   companyName,
			CompanyNumber: companyNumber,
		})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		// Expect CreateInsolvencyResource to be called once and return an error
		mockService.EXPECT().CreateInsolvencyResource(gomock.Any()).Return(errors.New("insolvency case already exists"), http.StatusConflict).Times(1)
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()
		res := serveHandleCreateInsolvencyResource(body, mockService, true, mockHelperService)

		So(res.Code, ShouldEqual, http.StatusConflict)
	})

	Convey("Error adding insolvency resource to mongo", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return a valid transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		// Expect the company profile api to be called and return a valid company
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/01234567", httpmock.NewStringResponder(http.StatusOK, companyProfileResponse))

		// Expect the alphakeyservice api to be called and return an alphakey
		httpmock.RegisterResponder(http.MethodGet, "http://localhost:18103/alphakey?name=companyName", httpmock.NewStringResponder(http.StatusOK, alphakeyResponse))

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:      constants.CVL.String(),
			CompanyName:   companyName,
			CompanyNumber: companyNumber,
		})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil)
		// Expect CreateInsolvencyResource to be called once and return an error
		mockService.EXPECT().CreateInsolvencyResource(gomock.Any()).Return(errors.New("error when creating mongo resource"), http.StatusInternalServerError).Times(1)

		res := serveHandleCreateInsolvencyResource(body, mockService, true, mockHelperService)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Error calling transaction api when patching transaction with new insolvency resource", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return a valid transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		// Expect the company profile api to be called and return a valid company
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/01234567", httpmock.NewStringResponder(http.StatusOK, companyProfileResponse))

		// Expect the alphakeyservice api to be called and return an alphakey
		httpmock.RegisterResponder(http.MethodGet, "http://localhost:18103/alphakey?name=companyName", httpmock.NewStringResponder(http.StatusOK, alphakeyResponse))

		// Expect the transaction api to be patched and return a success
		httpmock.RegisterResponder(http.MethodPatch, "http://localhost:4001/private/transactions/12345678", httpmock.NewStringResponder(http.StatusInternalServerError, transactionProfileResponse))

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:      constants.CVL.String(),
			CompanyName:   companyName,
			CompanyNumber: companyNumber,
		})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil)
		// Expect CreateInsolvencyResource to be called once and not return an error
		mockService.EXPECT().CreateInsolvencyResource(gomock.Any()).Return(nil, http.StatusCreated).Times(1)

		res := serveHandleCreateInsolvencyResource(body, mockService, true, mockHelperService)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Transaction not found when calling transaction api to patch transaction with new insolvency resource", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return a valid transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		// Expect the company profile api to be called and return a valid company
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/01234567", httpmock.NewStringResponder(http.StatusOK, companyProfileResponse))

		// Expect the alphakeyservice api to be called and return an alphakey
		httpmock.RegisterResponder(http.MethodGet, "http://localhost:18103/alphakey?name=companyName", httpmock.NewStringResponder(http.StatusOK, alphakeyResponse))

		// Expect the transaction api to be patched and return a success
		httpmock.RegisterResponder(http.MethodPatch, "http://localhost:4001/private/transactions/12345678", httpmock.NewStringResponder(http.StatusNotFound, ""))

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:      constants.CVL.String(),
			CompanyName:   companyName,
			CompanyNumber: companyNumber,
		})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil)
		// Expect CreateInsolvencyResource to be called once and not return an error
		mockService.EXPECT().CreateInsolvencyResource(gomock.Any()).Return(nil, http.StatusCreated).Times(1)

		res := serveHandleCreateInsolvencyResource(body, mockService, true, mockHelperService)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Successfully add insolvency resource to mongo", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return a valid transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		// Expect the company profile api to be called and return a valid company
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/01234567", httpmock.NewStringResponder(http.StatusOK, companyProfileResponse))

		// Expect the alphakeyservice api to be called and return an alphakey
		httpmock.RegisterResponder(http.MethodGet, "http://localhost:18103/alphakey?name=companyName", httpmock.NewStringResponder(http.StatusOK, alphakeyResponse))

		// Expect the transaction api to be patched and return a success
		httpmock.RegisterResponder(http.MethodPatch, "http://localhost:4001/private/transactions/12345678", httpmock.NewStringResponder(http.StatusNoContent, transactionProfileResponse))

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:      constants.CVL.String(),
			CompanyName:   companyName,
			CompanyNumber: companyNumber,
		})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil)
		// Expect CreateInsolvencyResource to be called once and not return an error
		mockService.EXPECT().CreateInsolvencyResource(gomock.Any()).Return(nil, http.StatusCreated).Times(1)

		res := serveHandleCreateInsolvencyResource(body, mockService, true, mockHelperService)

		So(res.Code, ShouldEqual, http.StatusCreated)
	})
}

func serveHandleGetValidationStatus(service dao.Service, tranIDSet bool) *httptest.ResponseRecorder {
	path := constants.TransactionsPath + transactionID + constants.ValidationStatusPath
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if tranIDSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}
	res := httptest.NewRecorder()

	handler := HandleGetValidationStatus(service)
	handler.ServeHTTP(res, req)

	return res
}

func serveHandleGetFilings(service dao.Service, tranIDSet bool) *httptest.ResponseRecorder {
	path := "/private/transactions/" + transactionID + "/insolvency/filings"
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if tranIDSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}
	res := httptest.NewRecorder()

	handler := HandleGetFilings(service)
	handler.ServeHTTP(res, req)

	return res
}

func createInsolvencyResource() models.InsolvencyResourceDao {
	return models.InsolvencyResourceDao{
		ID:            primitive.ObjectID{},
		TransactionID: transactionID,
		Etag:          "etag1234",
		Kind:          "insolvency",
		Data: models.InsolvencyResourceDaoData{
			CompanyNumber: companyNumber,
			CompanyName:   companyName,
			CaseType:      "insolvency",
			Practitioners: []models.PractitionerResourceDao{
				{
					Appointment: &models.AppointmentResourceDao{
						AppointedOn: "2020-01-01",
						MadeBy:      "creditors",
					},
				},
			},
		},
		Links: models.InsolvencyResourceLinksDao{
			Self:             "/transactions/123456789/insolvency",
			ValidationStatus: "/transactions/123456789/insolvency/validation-status",
		},
	}
}

func TestUnitHandleGetValidationStatus(t *testing.T) {
	InTest = true

	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockService := mock_dao.NewMockService(mockCtrl)

	Convey("Must need a transaction ID in the url", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		res := serveHandleGetValidationStatus(mock_dao.NewMockService(mockCtrl), false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Insolvency case not found in DB", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// Expect GetInsolvencyResource to be called once and return an error for the insolvency case
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(models.InsolvencyResourceDao{}, fmt.Errorf("there was a problem handling your request for transaction [%s] - insolvency case not found", transactionID)).Times(1)

		res := serveHandleGetValidationStatus(mockService, true)

		So(res.Code, ShouldEqual, http.StatusOK)
		So(res.Body.String(), ShouldContainSubstring, fmt.Sprintf("insolvency case with transactionID [%s] not found", transactionID))
	})

	Convey("Error returning insolvency case from DB", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// Expect GetInsolvencyResource to be called once and return an error for the insolvency case
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(models.InsolvencyResourceDao{}, errors.New("error getting insolvency case from DB")).Times(1)

		res := serveHandleGetValidationStatus(mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
		So(res.Body.String(), ShouldContainSubstring, `there was a problem handling your request`)
	})

	Convey("Case is found valid for submission", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(createInsolvencyResource(), nil).Times(1)

		res := serveHandleGetValidationStatus(mockService, true)

		So(res.Code, ShouldEqual, http.StatusOK)
		So(res.Body.String(), ShouldContainSubstring, `"is_valid":true`)
		So(res.Body.String(), ShouldContainSubstring, `"errors":[]`)
	})
}

func TestUnitHandleGetFilings(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	Convey("Must need a transaction ID in the url", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		res := serveHandleGetFilings(mock_dao.NewMockService(mockCtrl), false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error checking transaction status against transaction api", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()
		mockService := mock_dao.NewMockService(mockCtrl)

		// Expect the transaction api to be called and return a closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

		res := serveHandleGetFilings(mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Transaction is still open", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()
		mockService := mock_dao.NewMockService(mockCtrl)

		// Expect the transaction api to be called and return a closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		res := serveHandleGetFilings(mockService, true)

		So(res.Code, ShouldEqual, http.StatusForbidden)
	})

	Convey("Error generating filings", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()
		mockService := mock_dao.NewMockService(mockCtrl)

		// Expect the transaction api to be called and return a closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(models.InsolvencyResourceDao{}, fmt.Errorf("error")).Times(1)

		res := serveHandleGetFilings(mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Generate filing for 600 case with one practitioner", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()
		mockService := mock_dao.NewMockService(mockCtrl)

		// Expect the transaction api to be called and return a closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(createInsolvencyResource(), nil).Times(1)

		res := serveHandleGetFilings(mockService, true)

		So(res.Code, ShouldEqual, http.StatusOK)
		So(res.Body.String(), ShouldContainSubstring, `"case_type":"insolvency"`)
		So(res.Body.String(), ShouldContainSubstring, `"company_name":"companyName"`)
		So(res.Body.String(), ShouldContainSubstring, `"company_number":"01234567"`)
		So(res.Body.String(), ShouldContainSubstring, `"kind":"insolvency#600"`)
	})

	Convey("Generate filing for LRESEX case with no practitioners", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()
		mockService := mock_dao.NewMockService(mockCtrl)

		// Expect the transaction api to be called and return a closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		insolvencyCase := createInsolvencyResource()
		insolvencyCase.Data.Practitioners = []models.PractitionerResourceDao{}
		insolvencyCase.Data.Attachments = []models.AttachmentResourceDao{{
			Type: "resolution",
		}}

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyCase, nil).Times(1)

		res := serveHandleGetFilings(mockService, true)

		So(res.Code, ShouldEqual, http.StatusOK)
		So(res.Body.String(), ShouldContainSubstring, `"case_type":"insolvency"`)
		So(res.Body.String(), ShouldContainSubstring, `"company_name":"companyName"`)
		So(res.Body.String(), ShouldContainSubstring, `"company_number":"01234567"`)
		So(res.Body.String(), ShouldContainSubstring, `"kind":"insolvency#LRESEX"`)
		So(res.Body.String(), ShouldContainSubstring, `"Type":"resolution"`)
		So(res.Body.String(), ShouldContainSubstring, `"practitioners":[]`)
	})

	Convey(`Generate filing for LIQ02 case with one practitioner and "statement-of-affairs-director"`, t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()
		mockService := mock_dao.NewMockService(mockCtrl)

		// Expect the transaction api to be called and return a closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		insolvencyCase := createInsolvencyResource()
		insolvencyCase.Data.Practitioners[0].Appointment = nil
		insolvencyCase.Data.Attachments = []models.AttachmentResourceDao{{
			Type: "statement-of-affairs-director",
		}}

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyCase, nil).Times(1)

		res := serveHandleGetFilings(mockService, true)

		So(res.Code, ShouldEqual, http.StatusOK)
		So(res.Body.String(), ShouldContainSubstring, `"case_type":"insolvency"`)
		So(res.Body.String(), ShouldContainSubstring, `"company_name":"companyName"`)
		So(res.Body.String(), ShouldContainSubstring, `"company_number":"01234567"`)
		So(res.Body.String(), ShouldContainSubstring, `"kind":"insolvency#LIQ02"`)
		So(res.Body.String(), ShouldContainSubstring, `"Type":"statement-of-affairs-director"`)
	})
}
