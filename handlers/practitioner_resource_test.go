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

	practitionerResourceDao := models.PractitionerResourceDao{}
	practitionerResourceDao.Data.IPCode = "IPCode"
	practitionerResourceDao.Data.FirstName = "FirstName"
	practitionerResourceDao.Data.LastName = "LastName"
	practitionerResourceDao.Data.TelephoneNumber = "TelephoneNumber"
	practitionerResourceDao.Data.Email = "Email"
	practitionerResourceDao.Data.Address = models.AddressResourceDao{}
	practitionerResourceDao.Data.Role = "Role"
	practitionerResourceDao.Data.Links = models.PractitionerResourceLinksDao{}

	appointmentResourceDao := models.AppointmentResourceDao{}
	appointmentResourceDao.Data.AppointedOn = "2012-01-23"

	practitionerResourceDao.Data.Appointment = &appointmentResourceDao

	practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

	Convey("error if etag not generated", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)

		body, _ := json.Marshal(&models.InsolvencyRequest{})

		mockHelperService.EXPECT().GenerateEtag().Return("etag", fmt.Errorf("error generating etag: [%s]", "err")).AnyTimes()
		res := serveHandleCreateInsolvencyResource(body, mockService, false, mockHelperService, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
		So(res.Body.String(), ShouldContainSubstring, "error generating etag")
	})

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

		// Expect GetInsolvencyPractitionersResource to return a valid insolvency case
		mockService.EXPECT().GetInsolvencyPractitionersResource(gomock.Any()).Return(generateInsolvencyPractitionerAppointmentResources(), nil, nil)

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "the practitioner role supplied is not valid")
	})

	Convey("Incoming request has an invalid role - not final-liquidator for a CVL case", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		// Expect GetInsolvencyPractitionersResource to return a valid insolvency case
		insolvencyCase := generateInsolvencyPractitionerAppointmentResources()
		insolvencyCase.Data.CaseType = constants.CVL.String()
		mockService.EXPECT().GetInsolvencyPractitionersResource(gomock.Any()).Return(insolvencyCase, nil, nil)

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

		// Expect GetInsolvencyPractitionersResource to return an error
		mockService.EXPECT().GetInsolvencyPractitionersResource(gomock.Any()).Return(&models.InsolvencyResourceDao{}, nil, fmt.Errorf("failed to get insolvency"))

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
		So(res.Body.String(), ShouldContainSubstring, "failed to get insolvency")
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

		// Expect GetInsolvencyPractitionersResource to return a valid insolvency case
		mockService.EXPECT().GetInsolvencyPractitionersResource(gomock.Any()).Return(generateInsolvencyPractitionerAppointmentResources(), nil, nil)

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

		// Expect GetInsolvencyPractitionersResource to return a valid insolvency case
		mockService.EXPECT().GetInsolvencyPractitionersResource(gomock.Any()).Return(generateInsolvencyPractitionerAppointmentResources(), nil, nil)

		res := serveHandleCreatePractitionersResource(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "the first name contains a character which is not allowed")
	})

	Convey("Incoming request has telephone number and email missing and an invalid last name", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		// Expect GetInsolvencyPractitionersResource to return a valid insolvency case
		mockService.EXPECT().GetInsolvencyPractitionersResource(gomock.Any()).Return(generateInsolvencyPractitionerAppointmentResources(), nil, nil)

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

		mockService.EXPECT().GetInsolvencyPractitionersResource(gomock.Any()).Return(nil, nil, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)).Times(1)
		// Expect GetInsolvencyPractitionersResource to return a valid insolvency case
		mockService.EXPECT().GetInsolvencyPractitionersResource(gomock.Any()).Return(generateInsolvencyPractitionerAppointmentResources(), nil, nil).Times(2)
		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return(practitionerResourceDaos, nil)
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

		jsonPractitionersDaoMap := `{"VM04221441":  "/transactions/168570-809316-704268/insolvency/practitioners/VM04221441"}`

		dataDto := models.InsolvencyResourceDao{}
		dataDto.Data.CompanyNumber = "company_number"
		dataDto.Data.CaseType = "case_type"
		dataDto.Data.CompanyName = "company_name"
		dataDto.Data.Etag = "etag"
		dataDto.Data.Kind = "kind"
		dataDto.Data.Practitioners = jsonPractitionersDaoMap

		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetInsolvencyPractitionersResource(gomock.Any()).Return(&dataDto, nil, nil).Times(2)
		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return(practitionerResourceDaos, nil)
		mockService.EXPECT().CreatePractitionerResource(gomock.Any(), gomock.Any()).Return(200, nil)
		mockService.EXPECT().UpdateInsolvencyPractitioners(gomock.Any(), gomock.Any()).Return(404, fmt.Errorf("there was a problem handling your request for transaction id [%s] not found", transactionID))

		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusNotFound)
		So(res.Body.String(), ShouldContainSubstring, "there was a problem handling your request")
		So(res.Body.String(), ShouldContainSubstring, "not found")
	})

	Convey("Failed when practitioners are equal or more than 5", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		jsonPractitionersDaoMap := `{"VM04221441": "/transactions/168570-809316-704268/insolvency/practitioners/VM04221441",
			"VM04221442": "/transactions/168570-809316-704268/insolvency/practitioners/VM04221442",
			"VM04221444": "/transactions/168570-809316-704268/insolvency/practitioners/VM04221444",
			"VM04221445": "/transactions/168570-809316-704268/insolvency/practitioners/VM04221445",
			"VM04221446": "/transactions/168570-809316-704268/insolvency/practitioners/VM04221446"}`

		dataDto := models.InsolvencyResourceDao{}
		dataDto.Data.CompanyNumber = "company_number"
		dataDto.Data.CaseType = "case_type"
		dataDto.Data.CompanyName = "company_name"
		dataDto.Data.Etag = "etag"
		dataDto.Data.Kind = "kind"
		dataDto.Data.Practitioners = jsonPractitionersDaoMap

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		insolvencyCase := generateInsolvencyPractitionerAppointmentResources()
		insolvencyCase.Data.CaseType = constants.CVL.String()

		practitioner := generatePractitioner()
		body, _ := json.Marshal(practitioner)

		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetInsolvencyPractitionersResource(gomock.Any()).Return(&dataDto, nil, nil).Times(2)
		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return(practitionerResourceDaos, nil)
		mockService.EXPECT().CreatePractitionerResource(gomock.Any(), gomock.Any()).Return(200, nil)
		mockService.EXPECT().UpdateInsolvencyPractitioners(gomock.Any(), gomock.Any()).Return(200, nil)

		res := serveHandleCreatePractitionersResource(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "already has 5 practitioners")
	})

	Convey("Successfully add insolvency resource to mongo", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		jsonPractitionersDaoMap := `{"VM04221441": "/transactions/168570-809316-704268/insolvency/practitioners/VM04221441",
			"VM04221442": "/transactions/168570-809316-704268/insolvency/practitioners/VM04221442",
			"VM04221446": "/transactions/168570-809316-704268/insolvency/practitioners/VM04221446"}`

		dataDto := models.InsolvencyResourceDao{}
		dataDto.Data.CompanyNumber = "company_number"
		dataDto.Data.CaseType = "case_type"
		dataDto.Data.CompanyName = "company_name"
		dataDto.Data.Etag = "etag"
		dataDto.Data.Kind = "kind"
		dataDto.Data.Practitioners = jsonPractitionersDaoMap

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		insolvencyCase := generateInsolvencyPractitionerAppointmentResources()
		insolvencyCase.Data.CaseType = constants.CVL.String()

		practitioner := generatePractitioner()
		body, _ := json.Marshal(practitioner)

		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetInsolvencyPractitionersResource(gomock.Any()).Return(insolvencyCase, nil, nil).Times(2)
		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return(practitionerResourceDaos, nil)
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

func TestUnitHandleGetPractitionerResources(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	practitionerResourceDao := models.PractitionerResourceDao{}
	practitionerResourceDao.Data.IPCode = "IPCode"
	practitionerResourceDao.Data.FirstName = "FirstName"
	practitionerResourceDao.Data.LastName = "LastName"
	practitionerResourceDao.Data.TelephoneNumber = "TelephoneNumber"
	practitionerResourceDao.Data.Email = "Email"
	practitionerResourceDao.Data.Address = models.AddressResourceDao{}
	practitionerResourceDao.Data.Role = "Role"
	practitionerResourceDao.Data.Links = models.PractitionerResourceLinksDao{}

	appointmentResourceDao := models.AppointmentResourceDao{}
	appointmentResourceDao.Data.AppointedOn = "2012-01-23"

	practitionerResourceDao.Data.Appointment = &appointmentResourceDao

	practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

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
		insolvencyCase := generateInsolvencyPractitionerAppointmentResources()
		insolvencyCase.Data.CaseType = constants.CVL.String()

		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(nil, nil, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)).Times(1)
		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return(practitionerResourceDaos, nil).AnyTimes()

		res := serveGetPractitionersByIdssRequest(mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Error when retrieving practitioner resources from mongo - insolvency case not found", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		insolvencyCase := generateInsolvencyPractitionerAppointmentResources()
		insolvencyCase.Data.CaseType = constants.CVL.String()

		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(nil, nil, nil).Times(1)
		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return(practitionerResourceDaos, nil).AnyTimes()

		res := serveGetPractitionersByIdssRequest(mockService, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Error when retrieving practitioner resources from mongo - no practitioners assigned to insolvency case", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		insolvencyCase := generateInsolvencyPractitionerAppointmentResources()
		insolvencyCase.Data.CaseType = constants.CVL.String()

		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(nil, nil, nil).Times(1)
		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return(practitionerResourceDaos, nil).AnyTimes()

		res := serveGetPractitionersByIdssRequest(mockService, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Successfully retrieve practitioners for insolvency case", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		jsonPractitionersDaoMap := `
			"VM04221441":  "/transactions/168570-809316-704268/insolvency/practitioners/VM04221441",
			"VM042214412": "/transactions/168570-809316-704268/insolvency/practitioners/VM042214412",
			"VM04221443":  "/transactions/168570-809316-704268/insolvency/practitioners/VM042214413",
			"VM04221444":  "/transactions/168570-809316-704268/insolvency/practitioners/VM04221444",
			"VM04221445":  "/transactions/168570-809316-704268/insolvency/practitioners/VM04221445",
			"VM04221446":  "/transactions/168570-809316-704268/insolvency/practitioners/VM04221446"}`

		dataDto := models.InsolvencyResourceDao{}
		dataDto.Data.CompanyNumber = "company_number"
		dataDto.Data.CaseType = "case_type"
		dataDto.Data.CompanyName = "company_name"
		dataDto.Data.Etag = "etag"
		dataDto.Data.Kind = "kind"
		dataDto.Data.Practitioners = jsonPractitionersDaoMap

		mockService := mock_dao.NewMockService(mockCtrl)
		insolvencyCase := generateInsolvencyPractitionerAppointmentResources()
		insolvencyCase.Data.CaseType = constants.CVL.String()

		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(&dataDto, nil, fmt.Errorf("there was a problem handling your request for transaction %s already has 5 practitioners", transactionID)).Times(1)
		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return(practitionerResourceDaos, nil).AnyTimes()

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

	practitionerResourceDao := models.PractitionerResourceDao{}
	practitionerResourceDao.Data.IPCode = "IPCode"
	practitionerResourceDao.Data.FirstName = "FirstName"
	practitionerResourceDao.Data.LastName = "LastName"
	practitionerResourceDao.Data.TelephoneNumber = "TelephoneNumber"
	practitionerResourceDao.Data.Email = "Email"
	practitionerResourceDao.Data.Address = models.AddressResourceDao{}
	practitionerResourceDao.Data.Role = "Role"
	practitionerResourceDao.Data.Links = models.PractitionerResourceLinksDao{}

	appointmentResourceDao := models.AppointmentResourceDao{}
	appointmentResourceDao.Data.AppointedOn = "2012-01-23"

	practitionerResourceDao.Data.Appointment = &appointmentResourceDao

	practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

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
		// Expect GetPractitionersResource to return an error
		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return(practitionerResourceDaos, fmt.Errorf("error retrieving practitioner"))

		res := serveGetPractitionersByIdsRequest(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Practitioner resource not found", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect GetPractitionersResource to return an empty practitioner resource
		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return(nil, nil)

		res := serveGetPractitionersByIdsRequest(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Successfully retrieve practitioner resource", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect GetPractitionersResource to successfully return a practitioner resource
		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return(practitionerResourceDaos, nil)

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

func serveHandleAppointPractitioner(body []byte,
	service dao.Service,
	helperService utils.HelperService,
	tranIdSet bool,
	practitionerIDSet bool,
	res *httptest.ResponseRecorder) *httptest.ResponseRecorder {
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

	helperService := utils.NewHelperService()

	Convey("Must have a transaction ID in the url", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)

		body, _ := json.Marshal(&models.PractitionerAppointment{})

		res := serveHandleAppointPractitioner(body, mockService, helperService, false, false, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "there is no Transaction ID in the URL path")
	})

	Convey("Must have a practitioner ID in the url", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		body, _ := json.Marshal(&models.PractitionerAppointment{})

		res := serveHandleAppointPractitioner(body, mockService, helperService, true, false, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "there is no Practitioner ID in the URL path")
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

		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return(nil, fmt.Errorf("error"))
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

		mockService.EXPECT().GetInsolvencyPractitionersResource(gomock.Any()).Return(generateInsolvencyPractitionerAppointmentResources(), nil, fmt.Errorf("error"))
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

		appointmentResourceDao := models.AppointmentResourceDao{}
		appointmentResourceDao.Data.AppointedOn = "2012-01-23"

		insolvencyDao := models.InsolvencyResourceDao{}
		insolvencyDao.Data.CompanyNumber = "1234"
		insolvencyDao.Data.CaseType = "CVL"
		insolvencyDao.Data.CompanyName = "Company"

		practitionerResourceDao := models.PractitionerResourceDao{}
		practitionerResourceDao.Data.PractitionerId = practitionerID
		practitionerResourceDao.Data.Appointment = &appointmentResourceDao

		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2012-01-23",
			MadeBy:      "company",
		})

		mockHelperService.EXPECT().GenerateEtag().Return("etags", nil)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return(practitionerResourceDaos, nil)
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(&insolvencyDao, practitionerResourceDaos, nil)
		mockService.EXPECT().CreateAppointmentResource(gomock.Any()).Return(200, nil)
		mockService.EXPECT().UpdatePractitionerAppointment(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, nil)
		mockService.EXPECT().GetPractitionerAppointment(gomock.Any(), gomock.Any()).Return(&appointmentResourceDao, nil)

		res := serveHandleAppointPractitioner(body, mockService, mockHelperService, true, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "already appointed")
	})

	Convey("error checking practitioner details for appointment date", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		appointmentResourceDao := models.AppointmentResourceDao{}
		appointmentResourceDao.Data.AppointedOn = "data"

		practitionerResourceDao := models.PractitionerResourceDao{}
		practitionerResourceDao.Data.PractitionerId = practitionerID
		practitionerResourceDao.Data.Appointment = &appointmentResourceDao

		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		mockHelperService.EXPECT().GenerateEtag().Return("etags", nil)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return(practitionerResourceDaos, fmt.Errorf("error"))
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(&models.InsolvencyResourceDao{}, practitionerResourceDaos, fmt.Errorf("error"))
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

		appointmentResourceDao := models.AppointmentResourceDao{}
		appointmentResourceDao.Data.AppointedOn = "2033-01-01"

		practitionerResourceDao := models.PractitionerResourceDao{}
		practitionerResourceDao.Data.PractitionerId = practitionerID
		practitionerResourceDao.Data.Appointment = &appointmentResourceDao

		insolvencyDao := models.InsolvencyResourceDao{}
		insolvencyDao.Data.CompanyNumber = "1234"
		insolvencyDao.Data.CaseType = "CVL"
		insolvencyDao.Data.CompanyName = "Company"

		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2032-02-23",
			MadeBy:      "company",
		})

		mockHelperService.EXPECT().GenerateEtag().Return("etags", nil)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return(practitionerResourceDaos, nil).AnyTimes()
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(&insolvencyDao, practitionerResourceDaos, nil)
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

		insolvencyDao := models.InsolvencyResourceDao{}
		insolvencyDao.Data.CompanyNumber = "1234"
		insolvencyDao.Data.CaseType = "CVL"
		insolvencyDao.Data.CompanyName = "Company"

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2012-02-23",
			MadeBy:      "company",
		})

		mockHelperService.EXPECT().GenerateEtag().Return("etags", nil)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetInsolvencyPractitionersResource(gomock.Any()).Return(&insolvencyDao, nil, nil)
		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return(nil, nil)
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

		appointmentResourceDao := models.AppointmentResourceDao{}
		//appointmentResourceDao.Data.AppointedOn = "2012-02-23"
		appointmentResourceDao.Data.MadeBy = "company"
		appointmentResourceDao.Data.Links = models.AppointmentResourceLinksDao{
			Self: "/links/self",
		}

		practitionerResourceDao := models.PractitionerResourceDao{}
		practitionerResourceDao.Data.PractitionerId = practitionerID
		practitionerResourceDao.Data.Appointment = &appointmentResourceDao

		insolvencyDao := models.InsolvencyResourceDao{}
		insolvencyDao.Data.CompanyNumber = "1234"
		insolvencyDao.Data.CaseType = "CVL"
		insolvencyDao.Data.CompanyName = "Company"

		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2012-02-23",
			MadeBy:      "company",
		})

		mockHelperService.EXPECT().GenerateEtag().Return("etags", nil)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetPractitionerAppointment(gomock.Any(), gomock.Any()).Return(&models.AppointmentResourceDao{}, nil)
		mockService.EXPECT().GetInsolvencyPractitionersResource(gomock.Any()).Return(&insolvencyDao, practitionerResourceDaos, nil)
		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return([]models.PractitionerResourceDao{}, nil)
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
	practitionerResourceDao := models.PractitionerResourceDao{}
	practitionerResourceDao.Data.IPCode = "IPCode"
	practitionerResourceDao.Data.FirstName = "FirstName"
	practitionerResourceDao.Data.LastName = "LastName"
	practitionerResourceDao.Data.TelephoneNumber = "TelephoneNumber"
	practitionerResourceDao.Data.Email = "Email"
	practitionerResourceDao.Data.Address = models.AddressResourceDao{}
	practitionerResourceDao.Data.Role = "Role"
	practitionerResourceDao.Data.Links = models.PractitionerResourceLinksDao{}

	appointmentResourceDao := models.AppointmentResourceDao{}
	appointmentResourceDao.Data.AppointedOn = "2012-02-23"

	practitionerResourceDao.Data.Appointment = &appointmentResourceDao

	practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

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
		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return(practitionerResourceDaos, fmt.Errorf("error"))

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
		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return(nil, nil)

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2012-02-23",
			MadeBy:      "company",
		})
		res := serveHandleGetPractitionerAppointment(body, mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
		So(res.Body.String(), ShouldContainSubstring, "not found")
	})

	Convey("empty appointment returned", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)

		practitionerResourceDaos[0].Data.Appointment = nil
		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return(practitionerResourceDaos, nil)

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2012-02-23",
			MadeBy:      "company",
		})
		res := serveHandleGetPractitionerAppointment(body, mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
		So(res.Body.String(), ShouldContainSubstring, "no appointment found")
	})

	Convey("failed -no appointment returned when transactionID is not valid", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2012-02-23",
			MadeBy:      "company",
		})

		appointmentResourceDao := models.AppointmentResourceDao{}
		appointmentResourceDao.Data.AppointedOn = "2012-02-23"

		practitionerResourceDaos[0].Data.PractitionerId = practitionerID
		practitionerResourceDaos[0].Data.Appointment = &appointmentResourceDao
		practitionerResourceDaos[0].Data.Links.Appointment = "{\"00001234\":\"/transactions/X/insolvency/practitioners/00001234/appointment\"}"

		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return(practitionerResourceDaos, nil)

		res := serveHandleGetPractitionerAppointment(body, mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("failed -no appointment returned when practitionerID is not valid", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2012-02-23",
			MadeBy:      "company",
		})

		appointmentResourceDao := models.AppointmentResourceDao{}
		appointmentResourceDao.Data.AppointedOn = "2012-02-23"

		practitionerResourceDaos[0].Data.PractitionerId = "practitionerID"
		practitionerResourceDaos[0].Data.Appointment = &appointmentResourceDao
		practitionerResourceDaos[0].Data.Links.Appointment = "{\"00001234\":\"/transactions/123456789/insolvency/practitioners/00001234/appointment\"}"

		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return(practitionerResourceDaos, nil)

		res := serveHandleGetPractitionerAppointment(body, mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("success - appointment returned", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)

		body, _ := json.Marshal(models.PractitionerAppointment{
			AppointedOn: "2012-02-23",
			MadeBy:      "company",
		})

		appointmentResourceDao := models.AppointmentResourceDao{}
		appointmentResourceDao.Data.AppointedOn = "2012-02-23"

		practitionerResourceDaos[0].Data.PractitionerId = practitionerID
		practitionerResourceDaos[0].Data.Appointment = &appointmentResourceDao
		practitionerResourceDaos[0].Data.Links.Appointment = "{\"00001234\":\"/transactions/12345678/insolvency/practitioners/00001234/appointment\"}"

		mockService.EXPECT().GetPractitionersResource(gomock.Any()).Return(practitionerResourceDaos, nil)

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
