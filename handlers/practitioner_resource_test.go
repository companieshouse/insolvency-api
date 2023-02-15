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

	helperService := utils.NewHelperService()

	practitionerResourceDto := models.PractitionerResourceDto{
		Data: models.PractitionerResourceDao{
			IPCode:          "ip_code",
			FirstName:       "first_name",
			LastName:        "last_name",
			TelephoneNumber: "telephone_number,omitempty",
			Email:           "email,omitempty",
			Appointment: &models.AppointmentResourceDao{
				AppointedOn: "2012-01-23",
			},
		},
	}
	practitionerResourceDtos := append([]models.PractitionerResourceDto{}, practitionerResourceDto)

	Convey("Must need a transaction ID in the url", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)

		body, _ := json.Marshal(&models.InsolvencyRequest{})

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, false, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "transaction ID is not in the URL path")
	})

	Convey("Error checking if transaction is closed against transaction api", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an error
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

		body, _ := json.Marshal(&models.InsolvencyRequest{})

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
		So(res.Body.String(), ShouldContainSubstring, "error checking transaction status")
	})

	Convey("Transaction is already closed and cannot be updated", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an already closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		body, _ := json.Marshal(&models.InsolvencyRequest{})

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusForbidden)
		So(res.Body.String(), ShouldContainSubstring, "already closed and cannot be updated")
	})

	Convey("Failed to read request body", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		body := []byte(`{"first_name":error`)

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "failed to read request body for transaction")
	})

	Convey("Incoming request has IP code missing", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.IPCode = ""
		body, _ := json.Marshal(practitioner)

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "ip_code is a required field")
	})

	Convey("Incoming request has invalid IP code - not a number", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.IPCode = "+1234"
		body, _ := json.Marshal(practitioner)
		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "ip_code must be a valid number")
	})

	Convey("Incoming request has invalid IP code - more than 8 characters in length", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.IPCode = "123456789"
		body, _ := json.Marshal(practitioner)

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "ip_code must be a maximum of 8 characters in length")
	})

	Convey("Incoming request has first name missing", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.FirstName = ""
		body, _ := json.Marshal(practitioner)

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "first_name is a required field")
	})

	Convey("Incoming request has last name missing", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.LastName = ""
		body, _ := json.Marshal(practitioner)

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "last_name is a required field")
	})

	Convey("Incoming request has address missing", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.Address = models.Address{}
		body, _ := json.Marshal(practitioner)

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "address_line_1 is a required field, locality is a required field")
	})

	Convey("Incoming request has address premises missing", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.Address = models.Address{
			Locality: "locality",
		}
		body, _ := json.Marshal(practitioner)

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "premises is a required field")
	})

	Convey("Incoming request has address line 1 missing", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.Address = models.Address{
			Locality: "locality",
		}
		body, _ := json.Marshal(practitioner)

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "address_line_1 is a required field")
	})

	Convey("Incoming request has locality missing", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.Address = models.Address{
			AddressLine1: "addressline1",
		}
		body, _ := json.Marshal(practitioner)

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "locality is a required field")
	})

	Convey("Incoming request has address postcode missing", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.Address = models.Address{
			Locality: "locality",
		}
		body, _ := json.Marshal(practitioner)

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "postal_code is a required field")
	})

	Convey("Incoming request has role missing", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.Role = ""
		body, _ := json.Marshal(practitioner)

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "role is a required field")
	})

	Convey("Incoming request has an invalid role", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.Role = "error-role"
		body, _ := json.Marshal(practitioner)

		// Expect GetInsolvencyResource to return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(generateInsolvencyResource(), nil)

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "the practitioner role supplied is not valid")
	})

	Convey("Incoming request has an invalid role - not final-liquidator for a CVL case", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
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

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "invalid request body: the practitioner role must be final-liquidator")
	})

	Convey("Error retrieving insolvency case when validating practitioner", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.Role = constants.Receiver.String()
		body, _ := json.Marshal(practitioner)

		// Expect GetInsolvencyResource to return an error
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(models.InsolvencyResourceDao{}, fmt.Errorf("error retrieving insolvency case"))

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
		So(res.Body.String(), ShouldContainSubstring, "failed to validate the practitioner request supplied")
	})

	Convey("Incoming request has telephone number and email missing", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = ""
		practitioner.Email = ""
		body, _ := json.Marshal(practitioner)
		// Expect GetInsolvencyResource to return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(generateInsolvencyResource(), nil)

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "invalid request body: either telephone_number or email are required")
	})

	Convey("Incoming request has invalid first name", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		practitioner.FirstName = "J4ck"
		body, _ := json.Marshal(practitioner)

		// Expect GetInsolvencyResource to return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(generateInsolvencyResource(), nil)

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "the first name contains a character which is not allowed")
	})

	Convey("Incoming request has telephone number and email missing and an invalid last name", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))
		// Expect GetInsolvencyResource to return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(generateInsolvencyResource(), nil)

		practitioner := generatePractitioner()
		practitioner.LastName = "wr0ng"
		practitioner.TelephoneNumber = ""
		practitioner.Email = ""
		body, _ := json.Marshal(practitioner)

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

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

		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetInsolvencyPractitionerByTransactionID(gomock.Any()).Return(nil, "", fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)).Times(1)
		// Expect GetInsolvencyResource to return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(generateInsolvencyResource(), nil)
		mockService.EXPECT().GetPractitionersByIdsFromPractitioner(gomock.Any(), gomock.Any()).Return(practitionerResourceDtos, nil)
		mockService.EXPECT().CreatePractitionerResource(gomock.Any(), gomock.Any()).Return(200, nil)
		mockService.EXPECT().UpdateInsolvencyPractitioners(gomock.Any(), gomock.Any()).Return(200, nil)

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
		So(res.Body.String(), ShouldContainSubstring, "there was a problem handling your request")
	})

	Convey("Error adding practitioners resource to mongo - insolvency case not found", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		body, _ := json.Marshal(practitioner)

		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetInsolvencyPractitionerByTransactionID(gomock.Any()).Return(map[string]string{"id": "practitionerID"}, "", nil).Times(1)
		// Expect GetInsolvencyResource to return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(generateInsolvencyResource(), nil)
		mockService.EXPECT().GetPractitionersByIdsFromPractitioner(gomock.Any(), gomock.Any()).Return(practitionerResourceDtos, nil)
		mockService.EXPECT().CreatePractitionerResource(gomock.Any(), gomock.Any()).Return(200, nil)
		mockService.EXPECT().UpdateInsolvencyPractitioners(gomock.Any(), gomock.Any()).Return(404, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID))

		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusNotFound)
		So(res.Body.String(), ShouldContainSubstring, "there was a problem handling your request")
	})

	Convey("Error adding practitioners resource to mongo - 5 practitioners already exist", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		practitioner := generatePractitioner()
		body, _ := json.Marshal(practitioner)

		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetInsolvencyPractitionerByTransactionID(gomock.Any()).Return(nil, "", fmt.Errorf("there was a problem handling your request for transaction %s already has 5 practitioners", transactionID)).Times(1)
		// Expect GetInsolvencyResource to return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(generateInsolvencyResource(), nil)
		mockService.EXPECT().GetPractitionersByIdsFromPractitioner(gomock.Any(), gomock.Any()).Return(practitionerResourceDtos, nil)
		mockService.EXPECT().CreatePractitionerResource(gomock.Any(), gomock.Any()).Return(200, nil)
		mockService.EXPECT().UpdateInsolvencyPractitioners(gomock.Any(), gomock.Any()).Return(200, nil)

		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
		So(res.Body.String(), ShouldContainSubstring, "there was a problem handling your request")
		So(res.Body.String(), ShouldContainSubstring, "already has 5 practitioners")
	})

	Convey("Successfully add insolvency resource to mongo", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		jsonPractitionersDaoMap := map[string]string{
			"VM04221441":  "/transactions/168570-809316-704268/insolvency/practitioners/VM04221441",
			"VM042214412": "/transactions/168570-809316-704268/insolvency/practitioners/VM042214412",
			"VM04221443":  "/transactions/168570-809316-704268/insolvency/practitioners/VM042214413",
			"VM04221444":  "/transactions/168570-809316-704268/insolvency/practitioners/VM04221444",
			"VM04221445":  "/transactions/168570-809316-704268/insolvency/practitioners/VM04221445",
			"VM04221446":  "/transactions/168570-809316-704268/insolvency/practitioners/VM04221446"}
		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		insolvencyCase := generateInsolvencyResource()
		insolvencyCase.Data.CaseType = constants.CVL.String()

		practitioner := generatePractitioner()
		body, _ := json.Marshal(practitioner)

		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetInsolvencyPractitionerByTransactionID(gomock.Any()).Return(jsonPractitionersDaoMap, "", nil).Times(1)
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(insolvencyCase, nil)
		mockService.EXPECT().GetPractitionersByIdsFromPractitioner(gomock.Any(), gomock.Any()).Return(practitionerResourceDtos, nil)
		mockService.EXPECT().CreatePractitionerResource(gomock.Any(), gomock.Any()).Return(200, nil)
		mockService.EXPECT().UpdateInsolvencyPractitioners(gomock.Any(), gomock.Any()).Return(200, nil)

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

func serveGetPractitionersByIdssRequest(service dao.Service, tranIDSet bool) *httptest.ResponseRecorder {
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

func TestUnitHandleGetPractitionersByIdss(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	practitionerResourceDto := models.PractitionerResourceDto{
		Data: models.PractitionerResourceDao{
			IPCode:          "ip_code",
			FirstName:       "first_name",
			LastName:        "last_name",
			TelephoneNumber: "telephone_number,omitempty",
			Email:           "email,omitempty",
			Appointment: &models.AppointmentResourceDao{
				AppointedOn: "2012-01-23",
			},
		},
	}
	practitionerResourceDtos := append([]models.PractitionerResourceDto{}, practitionerResourceDto)

	Convey("Must need a transactionID in the URL", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveGetPractitionersByIdssRequest(mock_dao.NewMockService(mockCtrl), false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Generic error when retrieving practitioner resources from mongo", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		insolvencyCase := generateInsolvencyResource()
		insolvencyCase.Data.CaseType = constants.CVL.String()

		mockService.EXPECT().GetInsolvencyPractitionerByTransactionID(transactionID).Return(nil, "", fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)).Times(1)
		mockService.EXPECT().GetPractitionersByIdsFromPractitioner(gomock.Any(), gomock.Any()).Return(practitionerResourceDtos, nil).AnyTimes()

		res := serveGetPractitionersByIdssRequest(mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Error when retrieving practitioner resources from mongo - insolvency case not found", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		insolvencyCase := generateInsolvencyResource()
		insolvencyCase.Data.CaseType = constants.CVL.String()

		mockService.EXPECT().GetInsolvencyPractitionerByTransactionID(transactionID).Return(nil, "", nil).Times(1)
		mockService.EXPECT().GetPractitionersByIdsFromPractitioner(gomock.Any(), gomock.Any()).Return(practitionerResourceDtos, nil).AnyTimes()

		res := serveGetPractitionersByIdssRequest(mockService, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Error when retrieving practitioner resources from mongo - no practitioners assigned to insolvency case", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		insolvencyCase := generateInsolvencyResource()
		insolvencyCase.Data.CaseType = constants.CVL.String()

		mockService.EXPECT().GetInsolvencyPractitionerByTransactionID(transactionID).Return(nil, "", nil).Times(1)
		mockService.EXPECT().GetPractitionersByIdsFromPractitioner(gomock.Any(), gomock.Any()).Return(practitionerResourceDtos, nil).AnyTimes()

		res := serveGetPractitionersByIdssRequest(mockService, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Successfully retrieve practitioners for insolvency case", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		jsonPractitionersDaoMap := map[string]string{
			"VM04221441":  "/transactions/168570-809316-704268/insolvency/practitioners/VM04221441",
			"VM042214412": "/transactions/168570-809316-704268/insolvency/practitioners/VM042214412",
			"VM04221443":  "/transactions/168570-809316-704268/insolvency/practitioners/VM042214413",
			"VM04221444":  "/transactions/168570-809316-704268/insolvency/practitioners/VM04221444",
			"VM04221445":  "/transactions/168570-809316-704268/insolvency/practitioners/VM04221445",
			"VM04221446":  "/transactions/168570-809316-704268/insolvency/practitioners/VM04221446"}

		mockService := mock_dao.NewMockService(mockCtrl)
		insolvencyCase := generateInsolvencyResource()
		insolvencyCase.Data.CaseType = constants.CVL.String()

		mockService.EXPECT().GetInsolvencyPractitionerByTransactionID(transactionID).Return(jsonPractitionersDaoMap, "", fmt.Errorf("there was a problem handling your request for transaction %s already has 5 practitioners", transactionID)).Times(1)
		mockService.EXPECT().GetPractitionersByIdsFromPractitioner(gomock.Any(), gomock.Any()).Return(practitionerResourceDtos, nil).AnyTimes()

		res := serveGetPractitionersByIdssRequest(mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})
}

func serveGetPractitionersByIdsRequest(service dao.Service, tranIDSet bool, practIDSet bool) *httptest.ResponseRecorder {
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

	practitionerResourceDto := models.PractitionerResourceDto{
		Data: models.PractitionerResourceDao{
			IPCode:          "ip_code",
			FirstName:       "first_name",
			LastName:        "last_name",
			TelephoneNumber: "telephone_number,omitempty",
			Email:           "email,omitempty",
			Appointment: &models.AppointmentResourceDao{
				AppointedOn: "2012-01-23",
			},
		},
	}
	practitionerResourceDtos := append([]models.PractitionerResourceDto{}, practitionerResourceDto)

	Convey("Must need a transactionID in the URL", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveGetPractitionersByIdsRequest(mock_dao.NewMockService(mockCtrl), false, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Must need a practitionerID in the URL", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveGetPractitionersByIdsRequest(mock_dao.NewMockService(mockCtrl), true, false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Must need a transactionID and practitionerID in the URL", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveGetPractitionersByIdsRequest(mock_dao.NewMockService(mockCtrl), false, false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error when retrieving a practitioner resource from the DB", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect GetPractitionersByIdsFromPractitioner to return an error
		mockService.EXPECT().GetPractitionersByIdsFromPractitioner(gomock.Any(), gomock.Any()).Return(practitionerResourceDtos, fmt.Errorf("error retrieving practitioner"))

		res := serveGetPractitionersByIdsRequest(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Practitioner resource not found", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect GetPractitionersByIdsFromPractitioner to return an empty practitioner resource
		mockService.EXPECT().GetPractitionersByIdsFromPractitioner(gomock.Any(), gomock.Any()).Return(nil, nil)

		res := serveGetPractitionersByIdsRequest(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Successfully retrieve practitioner resource", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect GetPractitionersByIdsFromPractitioner to successfully return a practitioner resource
		mockService.EXPECT().GetPractitionersByIdsFromPractitioner(gomock.Any(), gomock.Any()).Return(practitionerResourceDtos, nil)

		res := serveGetPractitionersByIdsRequest(mockService, true, true)

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
		mockService.EXPECT().DeletePractitioner(practitionerID, transactionID).Return(http.StatusBadRequest, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)).Times(1)

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
		mockService.EXPECT().DeletePractitioner(practitionerID, transactionID).Return(http.StatusNotFound, nil).Times(1)

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
		mockService.EXPECT().DeletePractitioner(practitionerID, transactionID).Return(http.StatusNoContent, nil).Times(1)

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

	practitionerResourceDto := models.PractitionerResourceDto{
		ID: practitionerID,
		Data: models.PractitionerResourceDao{
			IPCode:          "ip_code",
			FirstName:       "first_name",
			LastName:        "last_name",
			TelephoneNumber: "telephone_number,omitempty",
			Email:           "email,omitempty",
			Appointment: &models.AppointmentResourceDao{
				AppointedOn: "2012-01-23",
			},
		},
	}
	practitionerResourceDtos := append([]models.PractitionerResourceDto{}, practitionerResourceDto)

	helperService := utils.NewHelperService()

	Convey("Must have a transaction ID in the url", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)

		body, _ := json.Marshal(&models.PractitionerAppointment{})

		res := serveHandleAppointPractitioner(body, mockService, helperService, false, false, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "transaction ID is not in the URL path")
	})

	Convey("Must have a practitioner ID in the url", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		body, _ := json.Marshal(&models.PractitionerAppointment{})

		res := serveHandleAppointPractitioner(body, mockService, helperService, true, false, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error checking if transaction is closed against transaction api", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an error
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

		body, _ := json.Marshal(&models.PractitionerAppointment{})

		res := serveHandleAppointPractitioner(body, mockService, helperService, true, true, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
		So(res.Body.String(), ShouldContainSubstring, "error checking transaction status")
	})

	Convey("Transaction is already closed and cannot be updated", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an already closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		body, _ := json.Marshal(&models.PractitionerAppointment{})

		res := serveHandleAppointPractitioner(body, mockService, helperService, true, true, rec)

		So(res.Code, ShouldEqual, http.StatusForbidden)
		So(res.Body.String(), ShouldContainSubstring, "is already closed and cannot be updated")
	})

	Convey("Failed to read request body", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		body := []byte(`{"appointed_on":error`)

		res := serveHandleAppointPractitioner(body, mockService, helperService, true, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "failed to read request body for transaction")
	})

	Convey("mandatory fields not supplied", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		body, _ := json.Marshal(models.PractitionerAppointment{})

		res := serveHandleAppointPractitioner(body, mockService, helperService, true, true, rec)

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

		mockHelperService.EXPECT().GenerateEtag().Return("etags", nil)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		//mockService.EXPECT().UpdateInsolvencyPractitionerAppointment(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, nil).Times(1)
		mockService.EXPECT().GetPractitionersByIdsFromPractitioner(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))
		mockService.EXPECT().CreateAppointmentResource(gomock.Any()).Return(200, nil)
		mockService.EXPECT().UpdatePractitionerAppointment(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, nil)

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

		mockHelperService.EXPECT().GenerateEtag().Return("etags", nil)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetPractitionersByIdsFromPractitioner(gomock.Any(), gomock.Any()).Return(practitionerResourceDtos, fmt.Errorf("error"))
		mockService.EXPECT().CreateAppointmentResource(gomock.Any()).Return(200, nil)
		mockService.EXPECT().UpdatePractitionerAppointment(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, nil)

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

		mockHelperService.EXPECT().GenerateEtag().Return("etags", nil)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetPractitionersByIdsFromPractitioner(gomock.Any(), gomock.Any()).Return(practitionerResourceDtos, nil)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyDao, nil)
		mockService.EXPECT().CreateAppointmentResource(gomock.Any()).Return(200, nil)
		mockService.EXPECT().UpdatePractitionerAppointment(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, nil)

		res := serveHandleAppointPractitioner(body, mockService, mockHelperService, true, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "already appointed")
	})

	Convey("error checking practitioner details for appointment date", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		mockHelperService.EXPECT().GenerateEtag().Return("etags", nil)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetPractitionersByIdsFromPractitioner(gomock.Any(), gomock.Any()).Return(practitionerResourceDtos, fmt.Errorf("error"))
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(models.InsolvencyResourceDao{}, fmt.Errorf("error"))
		mockService.EXPECT().CreateAppointmentResource(gomock.Any()).Return(200, nil)
		mockService.EXPECT().UpdatePractitionerAppointment(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, nil)

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
				Appointment: &models.AppointmentResourceDao{
					AppointedOn: "2033-01-01",
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
			AppointedOn: "2032-02-23",
			MadeBy:      "company",
		})

		mockHelperService.EXPECT().GenerateEtag().Return("etags", nil)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetPractitionersByIdsFromPractitioner(gomock.Any(), gomock.Any()).Return(practitionerResourceDtos, nil).AnyTimes()
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyDao, nil)
		mockService.EXPECT().CreateAppointmentResource(gomock.Any()).Return(200, nil)
		mockService.EXPECT().UpdatePractitionerAppointment(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, nil)

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

		practitionersDao := []models.PractitionerResourceDao{}
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

		mockHelperService.EXPECT().GenerateEtag().Return("etags", nil)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetPractitionersByIdsFromPractitioner(gomock.Any(), gomock.Any()).Return(nil, nil)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyDao, nil)
		mockService.EXPECT().GetPractitionerAppointment(gomock.Any(), gomock.Any()).Return(&models.AppointmentResourceDao{}, fmt.Errorf("error occured"))
		mockService.EXPECT().CreateAppointmentResource(gomock.Any()).Return(200, nil)
		mockService.EXPECT().UpdatePractitionerAppointment(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, nil)

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

		mockHelperService.EXPECT().GenerateEtag().Return("etags", nil)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyDao, nil)
		mockService.EXPECT().GetPractitionerAppointment(gomock.Any(), gomock.Any()).Return(&models.AppointmentResourceDao{}, nil)
		mockService.EXPECT().GetPractitionerAppointment(gomock.Any(), gomock.Any()).Return(&models.AppointmentResourceDao{}, nil)
		mockService.EXPECT().GetPractitionersByIdsFromPractitioner(gomock.Any(), gomock.Any()).Return([]models.PractitionerResourceDto{}, nil)
		mockService.EXPECT().CreateAppointmentResource(gomock.Any()).Return(200, nil)
		mockService.EXPECT().UpdatePractitionerAppointment(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, nil)

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
	practitionerResourceDto := models.PractitionerResourceDto{
		Data: models.PractitionerResourceDao{
			IPCode:          "ip_code",
			FirstName:       "first_name",
			LastName:        "last_name",
			TelephoneNumber: "telephone_number,omitempty",
			Email:           "email,omitempty",
			Appointment: &models.AppointmentResourceDao{
				AppointedOn: "2012-01-23",
			},
		},
	}
	practitionerResourceDtos := append([]models.PractitionerResourceDto{}, practitionerResourceDto)

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
		mockService.EXPECT().GetPractitionersByIdsFromPractitioner(gomock.Any(), gomock.Any()).Return(practitionerResourceDtos, fmt.Errorf("error"))

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
		mockService.EXPECT().GetPractitionersByIdsFromPractitioner(gomock.Any(), gomock.Any()).Return(nil, nil)

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

		practitionerResourceDtos[0].Data.Appointment = nil
		mockService.EXPECT().GetPractitionersByIdsFromPractitioner(gomock.Any(), gomock.Any()).Return(practitionerResourceDtos, nil)

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

		mockService := mock_dao.NewMockService(mockCtrl)

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2012-02-23",
			MadeBy:      "company",
		})

		practitionerResourceDtos[0].Data.Appointment = &models.AppointmentResourceDao{
			AppointedOn: "2012-01-23",
		}

		mockService.EXPECT().GetPractitionersByIdsFromPractitioner(gomock.Any(), gomock.Any()).Return(practitionerResourceDtos, nil)

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
		mockService.EXPECT().DeletePractitionerAppointment(transactionID, practitionerID).Return(http.StatusBadRequest, fmt.Errorf("err"))

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
		mockService.EXPECT().DeletePractitionerAppointment(transactionID, practitionerID).Return(http.StatusNoContent, nil)

		res := serveHandleDeletePractitionerAppointment(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusNoContent)
	})
}
