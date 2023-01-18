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

const practitionerID = "00001234"

func serveHandleCreatePractitionersResource(body []byte, service dao.Service, helperService utils.HelperService, tranIDSet bool, res *httptest.ResponseRecorder) *httptest.ResponseRecorder {
	path := "/transactions/123456789/insolvency/practitioners"
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	if tranIDSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}

	handler := HandleCreatePractitionersResource(service, helperService)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleCreatePractitionersResource(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	Convey("Must need a transaction ID in the url", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), "").Return(false, "").AnyTimes()

		http.Error(rec, "", http.StatusBadRequest)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, false, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error checking if transaction is closed against transaction api", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an error
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false).AnyTimes()

		http.Error(rec, "", http.StatusInternalServerError)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Transaction is already closed and cannot be updated", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an already closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false).AnyTimes()

		http.Error(rec, "", http.StatusForbidden)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusForbidden)
	})

	Convey("Failed to read request body", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		body := []byte(`{"first_name":error`)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false).AnyTimes()

		http.Error(rec, "", http.StatusBadRequest)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Incoming request has IP code missing", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.IPCode = ""
		body, _ := json.Marshal(practitioner)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false).AnyTimes()

		http.Error(rec, "ip_code is a required field", http.StatusBadRequest)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "ip_code is a required field")
	})

	Convey("Incoming request has invalid IP code - not a number", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.IPCode = "+1234"
		body, _ := json.Marshal(practitioner)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false).AnyTimes()

		http.Error(rec, "ip_code must be a valid number", http.StatusBadRequest)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "ip_code must be a valid number")
	})

	Convey("Incoming request has invalid IP code - more than 8 characters in length", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.IPCode = "123456789"
		body, _ := json.Marshal(practitioner)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false).AnyTimes()

		http.Error(rec, "ip_code must be a maximum of 8 characters in length", http.StatusBadRequest)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "ip_code must be a maximum of 8 characters in length")
	})

	Convey("Incoming request has first name missing", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.FirstName = ""
		body, _ := json.Marshal(practitioner)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false).AnyTimes()

		http.Error(rec, "first_name is a required field", http.StatusBadRequest)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "first_name is a required field")
	})

	Convey("Incoming request has last name missing", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.LastName = ""
		body, _ := json.Marshal(practitioner)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false).AnyTimes()

		http.Error(rec, "last_name is a required field", http.StatusBadRequest)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "last_name is a required field")
	})

	Convey("Incoming request has address missing", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.Address = models.Address{}
		body, _ := json.Marshal(practitioner)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false).AnyTimes()

		http.Error(rec, "address_line_1 is a required field, locality is a required field", http.StatusBadRequest)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "address_line_1 is a required field, locality is a required field")
	})

	Convey("Incoming request has address premises missing", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.Address = models.Address{
			Locality: "locality",
		}
		body, _ := json.Marshal(practitioner)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false).AnyTimes()

		http.Error(rec, "premises is a required field", http.StatusBadRequest)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "premises is a required field")
	})

	Convey("Incoming request has address line 1 missing", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.Address = models.Address{
			Locality: "locality",
		}
		body, _ := json.Marshal(practitioner)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false).AnyTimes()

		http.Error(rec, "address_line_1 is a required field", http.StatusBadRequest)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "address_line_1 is a required field")
	})

	Convey("Incoming request has locality missing", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.Address = models.Address{
			AddressLine1: "addressline1",
		}
		body, _ := json.Marshal(practitioner)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false).AnyTimes()

		http.Error(rec, "locality is a required field", http.StatusBadRequest)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "locality is a required field")
	})

	Convey("Incoming request has address postcode missing", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.Address = models.Address{
			Locality: "locality",
		}
		body, _ := json.Marshal(practitioner)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false).AnyTimes()

		http.Error(rec, "postal_code is a required field", http.StatusBadRequest)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "postal_code is a required field")
	})

	Convey("Incoming request has role missing", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.Role = ""
		body, _ := json.Marshal(practitioner)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false).AnyTimes()

		http.Error(rec, "role is a required field", http.StatusBadRequest)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "role is a required field")
	})

	Convey("Incoming request has an invalid role", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.Role = "error-role"
		body, _ := json.Marshal(practitioner)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false).AnyTimes()
		// Expect GetInsolvencyResource to return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(generateInsolvencyResource(), nil)

		http.Error(rec, "", http.StatusBadRequest)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Incoming request has an invalid role - not final-liquidator for a CVL case", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		// Expect GetInsolvencyResource to return a valid insolvency case
		insolvencyCase := generateInsolvencyResource()
		insolvencyCase.Data.CaseType = constants.CVL.String()
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(insolvencyCase, nil)

		practitioner := generatePractitioner()
		practitioner.Role = constants.Receiver.String()
		body, _ := json.Marshal(practitioner)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		http.Error(rec, "", http.StatusBadRequest)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error retrieving insolvency case when validating practitioner", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.Role = constants.Receiver.String()
		body, _ := json.Marshal(practitioner)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		// Expect GetInsolvencyResource to return an error
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(models.InsolvencyResourceDao{}, fmt.Errorf("error retrieving insolvency case"))

		http.Error(rec, "", http.StatusInternalServerError)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Incoming request has telephone number and email missing", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = ""
		practitioner.Email = ""
		body, _ := json.Marshal(practitioner)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		// Expect GetInsolvencyResource to return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(generateInsolvencyResource(), nil)

		http.Error(rec, "", http.StatusBadRequest)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Incoming request has invalid first name", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.FirstName = "J4ck"
		body, _ := json.Marshal(practitioner)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		// Expect GetInsolvencyResource to return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(generateInsolvencyResource(), nil)

		http.Error(rec, "the first name contains a character which is not allowed", http.StatusBadRequest)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "the first name contains a character which is not allowed")
	})

	Convey("Incoming request has telephone number and email missing and an invalid last name", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(generateInsolvencyResource(), nil)

		practitioner := generatePractitioner()
		practitioner.LastName = "wr0ng"
		practitioner.TelephoneNumber = ""
		practitioner.Email = ""
		body, _ := json.Marshal(practitioner)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		// Expect GetInsolvencyResource to return a valid insolvency case

		http.Error(rec, "invalid request body: either telephone_number or email are required, the last name contains a character which is not allowed", http.StatusBadRequest)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "invalid request body: either telephone_number or email are required")
		So(res.Body.String(), ShouldContainSubstring, "the last name contains a character which is not allowed")
	})

	Convey("Generic error when adding practitioners resource to mongo", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		body, _ := json.Marshal(practitioner)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		// Expect CreatePractitionersResource to be called once and return an error
		mockService.EXPECT().CreatePractitionersResource(gomock.Any(), transactionID).Return(fmt.Errorf("there was a problem handling your request for transaction %s", transactionID), http.StatusInternalServerError).Times(1)
		// Expect GetInsolvencyResource to return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(generateInsolvencyResource(), nil)

		http.Error(rec, "Server error", http.StatusInternalServerError)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Error adding practitioners resource to mongo - insolvency case not found", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		body, _ := json.Marshal(practitioner)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		// Expect CreatePractitionersResource to be called once and return an error
		mockService.EXPECT().CreatePractitionersResource(gomock.Any(), transactionID).Return(fmt.Errorf("there was a problem handling your request for transaction %s not found", transactionID), http.StatusNotFound).Times(1)
		// Expect GetInsolvencyResource to return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(generateInsolvencyResource(), nil)

		http.Error(rec, "Server error", http.StatusNotFound)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Error adding practitioners resource to mongo - 5 practitioners already exist", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		body, _ := json.Marshal(practitioner)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		// Expect CreatePractitionersResource to be called once and return an error
		mockService.EXPECT().CreatePractitionersResource(gomock.Any(), transactionID).Return(fmt.Errorf("there was a problem handling your request for transaction %s already has 5 practitioners", transactionID), http.StatusBadRequest).Times(1)
		// Expect GetInsolvencyResource to return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(generateInsolvencyResource(), nil)

		http.Error(rec, "", http.StatusBadRequest)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error adding practitioners resource to mongo - the limit of practitioners will exceed", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		body, _ := json.Marshal(practitioner)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		// Expect CreatePractitioners to be called once and return an error
		mockService.EXPECT().CreatePractitionersResource(gomock.Any(), transactionID).Return(fmt.Errorf("there was a problem handling your request for transaction %s will have more than 5 practitioners", transactionID), http.StatusBadRequest).Times(1)
		// Expect GetInsolvencyResource to return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(generateInsolvencyResource(), nil)

		http.Error(rec, "", http.StatusBadRequest)
		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Successfully add insolvency resource to mongo", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		insolvencyCase := generateInsolvencyResource()
		insolvencyCase.Data.CaseType = constants.CVL.String()

		practitioner := generatePractitioner()
		body, _ := json.Marshal(practitioner)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		// Expect CreatePractitionersResource to be called once and not return an error
		mockService.EXPECT().CreatePractitionersResource(gomock.Any(), transactionID).Return(nil, http.StatusCreated).Times(1)
		// Expect GetInsolvencyResource to return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(insolvencyCase, nil)

		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusCreated)
	})
}

func generatePractitioner() models.PractitionerRequest {
	return models.PractitionerRequest{
		IPCode:          "1234",
		FirstName:       "Joe",
		LastName:        "Bloggs",
		TelephoneNumber: "07777777777",
		Email:           "a@b.com",
		Address: models.Address{
			Premises:     "premises",
			AddressLine1: "addressline1",
			Locality:     "locality",
			PostalCode:   "postcode",
		},
		Role: constants.FinalLiquidator.String(),
	}
}

func serveGetPractitionerResourcesRequest(service dao.Service, tranIDSet bool) *httptest.ResponseRecorder {
	path := constants.TransactionsPath + transactionID + "/insolvency/practitioners"
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if tranIDSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}
	res := httptest.NewRecorder()

	handler := HandleGetPractitionerResources(service)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleGetPractitionerResources(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	Convey("Must need a transactionID in the URL", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveGetPractitionerResourcesRequest(mock_dao.NewMockService(mockCtrl), false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Generic error when retrieving practitioner resources from mongo", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect GetPractitionersResource to be called once and return an error
		mockService.EXPECT().GetPractitionerResources(transactionID).Return(nil, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)).Times(1)

		res := serveGetPractitionerResourcesRequest(mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Error when retrieving practitioner resources from mongo - insolvency case not found", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect GetPractitionersResource to be called once and return nil, nil
		mockService.EXPECT().GetPractitionerResources(transactionID).Return(nil, nil).Times(1)

		res := serveGetPractitionerResourcesRequest(mockService, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Error when retrieving practitioner resources from mongo - no practitioners assigned to insolvency case", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		var practitionerResources []models.PractitionerResourceDao
		// Expect GetPractitionersResource to be called once and return empty list, nil
		mockService.EXPECT().GetPractitionerResources(transactionID).Return(practitionerResources, nil).Times(1)

		res := serveGetPractitionerResourcesRequest(mockService, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Successfully retrieve practitioners for insolvency case", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		practitionerResources := []models.PractitionerResourceDao{
			{
				IPCode:    "IPCode",
				FirstName: "FirstName",
				LastName:  "LastName",
				Address: models.AddressResourceDao{
					AddressLine1: "AddressLine1",
					Locality:     "Locality",
				},
				Role: "Role",
			},
		}
		// Expect GetPractitionersResource to be called once and return list of practitioners, nil
		mockService.EXPECT().GetPractitionerResources(transactionID).Return(practitionerResources, nil).Times(1)

		res := serveGetPractitionerResourcesRequest(mockService, true)

		So(res.Code, ShouldEqual, http.StatusOK)
	})
}

func serveGetPractitionerResourceRequest(service dao.Service, tranIDSet bool, practIDSet bool) *httptest.ResponseRecorder {
	path := constants.TransactionsPath + transactionID + constants.PractitionersPath + practitionerID
	req := httptest.NewRequest(http.MethodGet, path, nil)
	vars := make(map[string]string)
	if tranIDSet {
		vars["transaction_id"] = transactionID
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}
	if practIDSet {
		vars["practitioner_id"] = practitionerID
		req = mux.SetURLVars(req, vars)
	}
	res := httptest.NewRecorder()

	handler := HandleGetPractitionerResource(service)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleGetPractitionerResource(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	Convey("Must need a transactionID in the URL", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveGetPractitionerResourceRequest(mock_dao.NewMockService(mockCtrl), false, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Must need a practitionerID in the URL", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveGetPractitionerResourceRequest(mock_dao.NewMockService(mockCtrl), true, false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Must need a transactionID and practitionerID in the URL", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveGetPractitionerResourceRequest(mock_dao.NewMockService(mockCtrl), false, false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error when retrieving a practitioner resource from the DB", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect GetPractitionerResource to return an error
		mockService.EXPECT().GetPractitionerResource(gomock.Any(), gomock.Any()).Return(models.PractitionerResourceDao{}, fmt.Errorf("error retrieving practitioner"))

		res := serveGetPractitionerResourceRequest(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Practitioner resource not found", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect GetPractitionerResource to return an empty practitioner resource
		mockService.EXPECT().GetPractitionerResource(gomock.Any(), gomock.Any()).Return(models.PractitionerResourceDao{}, nil)

		res := serveGetPractitionerResourceRequest(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Successfully retrieve practitioner resource", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		practitionerResource := models.PractitionerResourceDao{
			ID:              "1234",
			IPCode:          "1111",
			FirstName:       "First",
			LastName:        "Last",
			TelephoneNumber: "1234",
			Email:           "firstlast@email.com",
			Address:         models.AddressResourceDao{},
			Role:            "final-liquidator",
			Links:           models.PractitionerResourceLinksDao{},
			Appointment:     nil,
		}

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect GetPractitionerResource to successfully return a practitioner resource
		mockService.EXPECT().GetPractitionerResource(gomock.Any(), gomock.Any()).Return(practitionerResource, nil)

		res := serveGetPractitionerResourceRequest(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusOK)
	})
}

func serveDeletePractitionerRequest(service dao.Service, tranIdSet bool, practIdSet bool) *httptest.ResponseRecorder {
	path := constants.TransactionsPath + transactionID + constants.PractitionersPath + practitionerID
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	vars := make(map[string]string)
	if tranIdSet {
		vars["transaction_id"] = transactionID
		req = mux.SetURLVars(req, vars)
	}

	if practIdSet {
		vars["practitioner_id"] = practitionerID
		req = mux.SetURLVars(req, vars)
	}
	res := httptest.NewRecorder()

	handler := HandleDeletePractitioner(service)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleDeletePractitioner(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	Convey("Must need a transactionID in the URL", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveDeletePractitionerRequest(mock_dao.NewMockService(mockCtrl), false, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Must need a practitionerID in the URL", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveDeletePractitionerRequest(mock_dao.NewMockService(mockCtrl), true, false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error checking if transaction is closed against transaction api", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		// Expect the transaction api to be called and return an error
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

		res := serveDeletePractitionerRequest(mock_dao.NewMockService(mockCtrl), true, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Transaction is already closed and cannot be updated", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		// Expect the transaction api to be called and return an already closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		res := serveDeletePractitionerRequest(mock_dao.NewMockService(mockCtrl), true, true)

		So(res.Code, ShouldEqual, http.StatusForbidden)
	})

	Convey("Generic error when deleting practitioner resource from mongo", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect DeletePractitioner to be called once and return an error
		mockService.EXPECT().DeletePractitioner(practitionerID, transactionID).Return(fmt.Errorf("there was a problem handling your request for transaction %s", transactionID), http.StatusBadRequest).Times(1)

		res := serveDeletePractitionerRequest(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error when retrieving practitioner resources from mongo - insolvency case or practitioner not found", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect DeletePractitioner to be called once and return nil, 404
		mockService.EXPECT().DeletePractitioner(practitionerID, transactionID).Return(nil, http.StatusNotFound).Times(1)

		res := serveDeletePractitionerRequest(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Successfully retrieve practitioners for insolvency case", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect DeletePractitioner to be called once and return http status NoContent, nil
		mockService.EXPECT().DeletePractitioner(practitionerID, transactionID).Return(nil, http.StatusNoContent).Times(1)

		res := serveDeletePractitionerRequest(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusNoContent)
	})
}

func serveHandleAppointPractitioner(body []byte, service dao.Service, helperService utils.HelperService, tranIdSet bool, practitionerIDSet bool, res *httptest.ResponseRecorder) *httptest.ResponseRecorder {
	path := "/transactions/123456789/insolvency/practitioners/abcd/appointment"
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	vars := make(map[string]string)
	if tranIdSet {
		vars["transaction_id"] = transactionID
	}

	if practitionerIDSet {
		vars["practitioner_id"] = practitionerID
	}
	req = mux.SetURLVars(req, vars)

	handler := HandleAppointPractitioner(service, helperService)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleAppointPractitioner(t *testing.T) {
	apiURL := "https://api.companieshouse.gov.uk"

	Convey("Must have a transaction ID in the url", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		body, _ := json.Marshal(&models.PractitionerAppointment{})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), "").Return(false, "").AnyTimes()

		http.Error(rec, "", http.StatusBadRequest)
		res := serveHandleAppointPractitioner(body, mockService, mockHelperService, false, false, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Must have a practitioner ID in the url", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		body, _ := json.Marshal(&models.PractitionerAppointment{})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, "").AnyTimes()

		http.Error(rec, "", http.StatusBadRequest)
		res := serveHandleAppointPractitioner(body, mockService, mockHelperService, true, false, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error checking if transaction is closed against transaction api", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an error
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

		body, _ := json.Marshal(&models.PractitionerAppointment{})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false).AnyTimes()

		http.Error(rec, "", http.StatusInternalServerError)
		res := serveHandleAppointPractitioner(body, mockService, mockHelperService, true, true, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Transaction is already closed and cannot be updated", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an already closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		body, _ := json.Marshal(&models.PractitionerAppointment{})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false).AnyTimes()

		http.Error(rec, "", http.StatusForbidden)
		res := serveHandleAppointPractitioner(body, mockService, mockHelperService, true, true, rec)

		So(res.Code, ShouldEqual, http.StatusForbidden)
	})

	Convey("Failed to read request body", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		body := []byte(`{"appointed_on":error`)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false).AnyTimes()

		http.Error(rec, fmt.Sprintf("failed to read request body for transaction %s", transactionID), http.StatusBadRequest)
		res := serveHandleAppointPractitioner(body, mockService, mockHelperService, true, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("mandatory fields not supplied", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		body, _ := json.Marshal(models.PractitionerAppointment{})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false).AnyTimes()

		http.Error(rec, "made_by is a required field", http.StatusBadRequest)
		res := serveHandleAppointPractitioner(body, mockService, mockHelperService, true, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "made_by is a required field")
	})

	Convey("invalid made_by field supplied", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2012-02-23",
			MadeBy:      "invalid",
		})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		http.Error(rec, "made_by supplied is not valid", http.StatusBadRequest)
		res := serveHandleAppointPractitioner(body, mockService, mockHelperService, true, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "made_by supplied is not valid")
	})

	Convey("error checking practitioner details", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2012-02-23",
			MadeBy:      "company",
		})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockService.EXPECT().GetPractitionerResources(transactionID).Return(nil, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)).Times(1)

		res := serveHandleAppointPractitioner(body, mockService, mockHelperService, true, true, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
		So(res.Body.String(), ShouldContainSubstring, fmt.Sprintf("there was a problem handling your request for transaction ID [%s]", transactionID))
	})

	Convey("practitioner already appointed", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		practitionersDao := []models.PractitionerResourceDao{
			{
				ID: practitionerID,
				Appointment: &models.AppointmentResourceDao{
					AppointedOn: "2012-01-23",
				},
			},
		}
		insolvencyDao := models.InsolvencyResourceDao{
			Data: models.InsolvencyResourceDaoData{
				CompanyNumber: "1234",
				CaseType:      "CVL",
				CompanyName:   "Company",
				Practitioners: practitionersDao,
			},
		}

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2012-02-23",
			MadeBy:      "company",
		})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockService.EXPECT().GetPractitionerResources(transactionID).Return(practitionersDao, nil).AnyTimes()
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyDao, nil)

		res := serveHandleAppointPractitioner(body, mockService, mockHelperService, true, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "already appointed")
	})

	Convey("error checking practitioner details for appointment date", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitionersDao := []models.PractitionerResourceDao{{ID: practitionerID}}
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockService.EXPECT().GetPractitionerResources(transactionID).Return(practitionersDao, nil).Times(1)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(models.InsolvencyResourceDao{}, fmt.Errorf("error"))

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2012-02-23",
			MadeBy:      "company",
		})

		res := serveHandleAppointPractitioner(body, mockService, mockHelperService, true, true, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
		So(res.Body.String(), ShouldContainSubstring, fmt.Sprintf("there was a problem handling your request for transaction ID [%s]", transactionID))
	})

	Convey("appointment date invalid - date differs", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitionersDao := []models.PractitionerResourceDao{
			{
				ID: practitionerID,
			},
			{
				ID: "123",
				Appointment: &models.AppointmentResourceDao{
					AppointedOn: "2013-01-01",
				},
			},
		}

		insolvencyDao := models.InsolvencyResourceDao{
			Data: models.InsolvencyResourceDaoData{
				CompanyNumber: "1234",
				CaseType:      "CVL",
				CompanyName:   "Company",
				Practitioners: practitionersDao,
			},
		}

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2012-02-23",
			MadeBy:      "company",
		})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockService.EXPECT().GetPractitionerResources(transactionID).Return(practitionersDao, nil).Times(1)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyDao, nil)

		res := serveHandleAppointPractitioner(body, mockService, mockHelperService, true, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("error storing appointment in DB", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitionersDao := []models.PractitionerResourceDao{{ID: practitionerID}}
		insolvencyDao := models.InsolvencyResourceDao{
			Data: models.InsolvencyResourceDaoData{
				CompanyNumber: "1234",
				CaseType:      "CVL",
				CompanyName:   "Company",
				Practitioners: practitionersDao,
			},
		}

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2012-02-23",
			MadeBy:      "company",
		})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockService.EXPECT().GetPractitionerResources(transactionID).Return(practitionersDao, nil).Times(1)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyDao, nil)
		mockService.EXPECT().AppointPractitioner(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("err"), http.StatusInternalServerError)

		res := serveHandleAppointPractitioner(body, mockService, mockHelperService, true, true, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("error getting practitioner for response", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitionersDao := []models.PractitionerResourceDao{{ID: practitionerID}}
		insolvencyDao := models.InsolvencyResourceDao{
			Data: models.InsolvencyResourceDaoData{
				CompanyNumber: "1234",
				CaseType:      "CVL",
				CompanyName:   "Company",
				Practitioners: practitionersDao,
			},
		}

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2012-02-23",
			MadeBy:      "company",
		})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockService.EXPECT().GetPractitionerResources(transactionID).Return(practitionersDao, nil).Times(1)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyDao, nil)
		mockService.EXPECT().AppointPractitioner(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 0)
		mockService.EXPECT().GetPractitionerResource(gomock.Any(), gomock.Any()).Return(models.PractitionerResourceDao{}, fmt.Errorf("error"))

		res := serveHandleAppointPractitioner(body, mockService, mockHelperService, true, true, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("empty practitioner returned", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitionersDao := []models.PractitionerResourceDao{{ID: practitionerID}}
		insolvencyDao := models.InsolvencyResourceDao{
			Data: models.InsolvencyResourceDaoData{
				CompanyNumber: "1234",
				CaseType:      "CVL",
				CompanyName:   "Company",
				Practitioners: practitionersDao,
			},
		}

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2012-02-23",
			MadeBy:      "company",
		})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockService.EXPECT().GetPractitionerResources(transactionID).Return(practitionersDao, nil).Times(1)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyDao, nil)
		mockService.EXPECT().AppointPractitioner(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 0)
		mockService.EXPECT().GetPractitionerResource(gomock.Any(), gomock.Any()).Return(models.PractitionerResourceDao{}, nil)

		res := serveHandleAppointPractitioner(body, mockService, mockHelperService, true, true, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("successful appointment", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitionersDao := []models.PractitionerResourceDao{
			{
				ID: "321",
				Appointment: &models.AppointmentResourceDao{
					AppointedOn: "2012-02-23",
					MadeBy:      "company",
					Links: models.AppointmentResourceLinksDao{
						Self: "/links/self",
					},
				},
			},
		}
		insolvencyDao := models.InsolvencyResourceDao{
			Data: models.InsolvencyResourceDaoData{
				CompanyNumber: "1234",
				CaseType:      "CVL",
				CompanyName:   "Company",
				Practitioners: practitionersDao,
			},
		}

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2012-02-23",
			MadeBy:      "company",
		})
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockService.EXPECT().GetPractitionerResources(transactionID).Return(practitionersDao, nil).Times(1)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyDao, nil)
		mockService.EXPECT().AppointPractitioner(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 0)
		mockService.EXPECT().GetPractitionerResource(gomock.Any(), gomock.Any()).Return(practitionersDao[0], nil)

		res := serveHandleAppointPractitioner(body, mockService, mockHelperService, true, true, rec)

		So(res.Code, ShouldEqual, http.StatusCreated)
	})
}

func serveHandleGetPractitionerAppointment(body []byte, service dao.Service, tranIdSet bool, practitionerIDSet bool) *httptest.ResponseRecorder {
	path := "/transactions/123456789/insolvency/practitioners/abcd/appointment"
	req := httptest.NewRequest(http.MethodGet, path, bytes.NewReader(body))
	vars := make(map[string]string)
	if tranIdSet {
		vars["transaction_id"] = transactionID
	}

	if practitionerIDSet {
		vars["practitioner_id"] = practitionerID
	}
	req = mux.SetURLVars(req, vars)
	res := httptest.NewRecorder()

	handler := HandleGetPractitionerAppointment(service)
	handler.ServeHTTP(res, req)

	return res
}

func companyProfileDateResponse(dateOfCreation string) string {
	return `
{
 "company_name": "companyName",
 "company_number": "01234567",
 "jurisdiction": "england-wales",
 "company_status": "active",
 "type": "private-shares-exemption-30",
 "date_of_creation": "` + dateOfCreation + `",
 "registered_office_address" : {
   "postal_code" : "CF14 3UZ",
   "address_line_2" : "Cardiff",
   "address_line_1" : "1 Crown Way"
  }
}
`

}

func TestUnitHandleGetPractitionerAppointment(t *testing.T) {
	Convey("Must have a transaction ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		body, _ := json.Marshal(&models.PractitionerAppointment{})
		res := serveHandleGetPractitionerAppointment(body, mock_dao.NewMockService(mockCtrl), false, false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Must have a practitioner ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		body, _ := json.Marshal(&models.PractitionerAppointment{})
		res := serveHandleGetPractitionerAppointment(body, mock_dao.NewMockService(mockCtrl), true, false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("error getting practitioner for response", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		mockService.EXPECT().GetPractitionerResource(gomock.Any(), gomock.Any()).Return(models.PractitionerResourceDao{}, fmt.Errorf("error"))

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2012-02-23",
			MadeBy:      "company",
		})
		res := serveHandleGetPractitionerAppointment(body, mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("empty practitioner returned", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		mockService.EXPECT().GetPractitionerResource(gomock.Any(), gomock.Any()).Return(models.PractitionerResourceDao{}, nil)

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2012-02-23",
			MadeBy:      "company",
		})
		res := serveHandleGetPractitionerAppointment(body, mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("empty appointment returned", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		mockService.EXPECT().GetPractitionerResource(gomock.Any(), gomock.Any()).Return(models.PractitionerResourceDao{ID: "123"}, nil)

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2012-02-23",
			MadeBy:      "company",
		})
		res := serveHandleGetPractitionerAppointment(body, mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("success - appointment returned", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		practitionersDao := models.PractitionerResourceDao{
			ID: "321",
			Appointment: &models.AppointmentResourceDao{
				AppointedOn: "2012-02-23",
				MadeBy:      "company",
				Links: models.AppointmentResourceLinksDao{
					Self: "/links/self",
				},
			},
		}

		mockService := mock_dao.NewMockService(mockCtrl)
		mockService.EXPECT().GetPractitionerResource(gomock.Any(), gomock.Any()).Return(practitionersDao, nil)

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2012-02-23",
			MadeBy:      "company",
		})
		res := serveHandleGetPractitionerAppointment(body, mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusOK)
	})
}

func serveHandleDeletePractitionerAppointment(service dao.Service, tranIdSet bool, practitionerIDSet bool) *httptest.ResponseRecorder {
	path := "/transactions/123456789/insolvency/practitioners/abcd/appointment"
	req := httptest.NewRequest(http.MethodGet, path, nil)
	vars := make(map[string]string)
	if tranIdSet {
		vars["transaction_id"] = transactionID
	}

	if practitionerIDSet {
		vars["practitioner_id"] = practitionerID
	}
	req = mux.SetURLVars(req, vars)
	res := httptest.NewRecorder()

	handler := HandleDeletePractitionerAppointment(service)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleDeletePractitionerAppointment(t *testing.T) {
	Convey("Must have a transaction ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveHandleDeletePractitionerAppointment(mock_dao.NewMockService(mockCtrl), false, false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Must have a practitioner ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveHandleDeletePractitionerAppointment(mock_dao.NewMockService(mockCtrl), true, false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error checking if transaction is closed against transaction api", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		// Expect the transaction api to be called and return an error
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

		res := serveHandleDeletePractitionerAppointment(mock_dao.NewMockService(mockCtrl), true, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Transaction is already closed and cannot be updated", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		// Expect the transaction api to be called and return an already closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		res := serveHandleDeletePractitionerAppointment(mock_dao.NewMockService(mockCtrl), true, true)

		So(res.Code, ShouldEqual, http.StatusForbidden)
	})

	Convey("Generic error when deleting practitioner appointment from mongo", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		mockService := mock_dao.NewMockService(mockCtrl)
		mockService.EXPECT().DeletePractitionerAppointment(transactionID, practitionerID).Return(fmt.Errorf("err"), http.StatusBadRequest)

		res := serveHandleDeletePractitionerAppointment(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Successful deletion of appointment", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		mockService := mock_dao.NewMockService(mockCtrl)
		mockService.EXPECT().DeletePractitionerAppointment(transactionID, practitionerID).Return(nil, http.StatusNoContent)

		res := serveHandleDeletePractitionerAppointment(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusNoContent)
	})
}
