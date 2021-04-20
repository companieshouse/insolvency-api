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
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/jarcoal/httpmock"

	. "github.com/smartystreets/goconvey/convey"
)

const practitionerID = "00001234"

func serveHandleCreatePractitionersResource(body []byte, service dao.Service, tranIdSet bool) *httptest.ResponseRecorder {
	path := "/transactions/123456789/insolvency/practitioners"
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	if tranIdSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}
	res := httptest.NewRecorder()

	handler := HandleCreatePractitionersResource(service)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleCreatePractitionersResource(t *testing.T) {
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
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Failed to read request body", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body := []byte(`{"first_name":error`)
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Incoming request has IP code missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.IPCode = ""
		body, _ := json.Marshal(practitioner)
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "ip_code is a required field")
	})

	Convey("Incoming request has first name missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.FirstName = ""
		body, _ := json.Marshal(practitioner)
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "first_name is a required field")
	})

	Convey("Incoming request has last name missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.LastName = ""
		body, _ := json.Marshal(practitioner)
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "last_name is a required field")
	})

	Convey("Incoming request has address missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.Address = models.Address{}
		body, _ := json.Marshal(practitioner)
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "address_line_1 is a required field, locality is a required field")
	})

	Convey("Incoming request has address line 1 missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.Address = models.Address{
			Locality: "locality",
		}
		body, _ := json.Marshal(practitioner)
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "address_line_1 is a required field")
	})

	Convey("Incoming request has locality missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.Address = models.Address{
			AddressLine1: "addressline1",
		}
		body, _ := json.Marshal(practitioner)
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "locality is a required field")
	})

	Convey("Incoming request has role missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.Role = ""
		body, _ := json.Marshal(practitioner)
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "role is a required field")
	})

	Convey("Incoming request has an invalid role", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.Role = "error-role"
		body, _ := json.Marshal(practitioner)
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Incoming request has telephone number and email missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = ""
		practitioner.Email = ""
		body, _ := json.Marshal(practitioner)
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "invalid request body: either telephone_number or email are required")
	})

	Convey("Generic error when adding practitioners resource to mongo", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect CreatePractitionersResource to be called once and return an error
		mockService.EXPECT().CreatePractitionersResource(gomock.Any(), transactionID).Return(fmt.Errorf("there was a problem handling your request for transaction %s", transactionID), http.StatusInternalServerError).Times(1)

		practitioner := generatePractitioner()
		body, _ := json.Marshal(practitioner)
		res := serveHandleCreatePractitionersResource(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Error adding practitioners resource to mongo - insolvency case not found", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect CreatePractitionersResource to be called once and return an error
		mockService.EXPECT().CreatePractitionersResource(gomock.Any(), transactionID).Return(fmt.Errorf("there was a problem handling your request for transaction %s not found", transactionID), http.StatusNotFound).Times(1)

		practitioner := generatePractitioner()
		body, _ := json.Marshal(practitioner)
		res := serveHandleCreatePractitionersResource(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Error adding practitioners resource to mongo - 5 practitioners already exist", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect CreatePractitionersResource to be called once and return an error
		mockService.EXPECT().CreatePractitionersResource(gomock.Any(), transactionID).Return(fmt.Errorf("there was a problem handling your request for transaction %s already has 5 practitioners", transactionID), http.StatusBadRequest).Times(1)

		practitioner := generatePractitioner()
		body, _ := json.Marshal(practitioner)
		res := serveHandleCreatePractitionersResource(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error adding practitioners resource to mongo - the limit of practitioners will exceed", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect CreatePractitioners to be called once and return an error
		mockService.EXPECT().CreatePractitionersResource(gomock.Any(), transactionID).Return(fmt.Errorf("there was a problem handling your request for transaction %s will have more than 5 practitioners", transactionID), http.StatusBadRequest).Times(1)

		practitioner := generatePractitioner()
		body, _ := json.Marshal(practitioner)
		res := serveHandleCreatePractitionersResource(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Successfully add insolvency resource to mongo", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect CreatePractitionersResource to be called once and not return an error
		mockService.EXPECT().CreatePractitionersResource(gomock.Any(), transactionID).Return(nil, http.StatusCreated).Times(1)

		practitioner := generatePractitioner()
		body, _ := json.Marshal(practitioner)
		res := serveHandleCreatePractitionersResource(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusCreated)
	})

}

func generatePractitioner() models.PractitionerRequest {
	return models.PractitionerRequest{
		IPCode:          "1234",
		FirstName:       "Joe",
		LastName:        "Bloggs",
		TelephoneNumber: "123456",
		Email:           "email",
		Address: models.Address{
			AddressLine1: "addressline1",
			Locality:     "locality",
		},
		Role: constants.FinalLiquidator.String(),
	}
}

func serveGetPractitionerResourcesRequest(service dao.Service, tranIdSet bool) *httptest.ResponseRecorder {
	path := "/transactions/" + transactionID + "/insolvency/practitioners"
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if tranIdSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}
	res := httptest.NewRecorder()

	handler := HandleGetPractitionerResources(service)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleGetractitionerResources(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	Convey("Must need a transactionID in the URL", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		res := serveGetPractitionerResourcesRequest(mock_dao.NewMockService(mockCtrl), false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Generic error when retrieving practitioner resources from mongo", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect GetPractitionersResource to be called once and return an error
		mockService.EXPECT().GetPractitionerResources(transactionID).Return(nil, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)).Times(1)

		res := serveGetPractitionerResourcesRequest(mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Generic error when retrieving practitioner resources from mongo", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect GetPractitionersResource to be called once and return an error
		mockService.EXPECT().GetPractitionerResources(transactionID).Return(nil, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)).Times(1)

		res := serveGetPractitionerResourcesRequest(mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Generic error when retrieving practitioner resources from mongo", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect GetPractitionersResource to be called once and return an error
		mockService.EXPECT().GetPractitionerResources(transactionID).Return(nil, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)).Times(1)

		res := serveGetPractitionerResourcesRequest(mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Error when retrieving practitioner resources from mongo - insolvency case not found", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect GetPractitionersResource to be called once and return nil, nil
		mockService.EXPECT().GetPractitionerResources(transactionID).Return(nil, nil).Times(1)

		res := serveGetPractitionerResourcesRequest(mockService, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Error when retrieving practitioner resources from mongo - no practitioners assigned to insolvency case", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		var practitionerResources []models.PractitionerResourceDao
		// Expect GetPractitionersResource to be called once and return empty list, nil
		mockService.EXPECT().GetPractitionerResources(transactionID).Return(practitionerResources, nil).Times(1)

		res := serveGetPractitionerResourcesRequest(mockService, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Successfully retrieve practitioners for insolvency case", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
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

func serveDeletePractitionerRequest(service dao.Service, tranIdSet bool, practIdSet bool) *httptest.ResponseRecorder {
	path := "/transactions/" + transactionID + "/insolvency/practitioners/" + practitionerID
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
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		res := serveDeletePractitionerRequest(mock_dao.NewMockService(mockCtrl), false, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Must need a practitionerID in the URL", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		res := serveDeletePractitionerRequest(mock_dao.NewMockService(mockCtrl), true, false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Generic error when deleting practitioner resource from mongo", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect DeletePractitioner to be called once and return an error
		mockService.EXPECT().DeletePractitioner(practitionerID, transactionID).Return(fmt.Errorf("there was a problem handling your request for transaction %s", transactionID), http.StatusBadRequest).Times(1)

		res := serveDeletePractitionerRequest(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error when retrieving practitioner resources from mongo - insolvency case or practitioner not found", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect DeletePractitioner to be called once and return nil, 404
		mockService.EXPECT().DeletePractitioner(practitionerID, transactionID).Return(nil, http.StatusNotFound).Times(1)

		res := serveDeletePractitionerRequest(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Successfully retrieve practitioners for insolvency case", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect DeletePractitioner to be called once and return http status NoContent, nil
		mockService.EXPECT().DeletePractitioner(practitionerID, transactionID).Return(nil, http.StatusNoContent).Times(1)

		res := serveDeletePractitionerRequest(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusNoContent)
	})

}
