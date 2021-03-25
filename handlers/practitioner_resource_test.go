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

		body, _ := json.Marshal(models.PractitionerRequest{
			FirstName: "First",
			LastName:  "Last",
			Address: models.Address{
				AddressLine1: "addressline1",
				Locality:     "locality",
			},
			Role: constants.FinalLiquidator.String(),
		})
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "ip_code is a required field")
	})

	Convey("Incoming request has first name missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body, _ := json.Marshal(models.PractitionerRequest{
			IPCode:   "1234",
			LastName: "Last",
			Address: models.Address{
				AddressLine1: "addressline1",
				Locality:     "locality",
			},
			Role: constants.FinalLiquidator.String(),
		})
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "first_name is a required field")
	})

	Convey("Incoming request has last name missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body, _ := json.Marshal(models.PractitionerRequest{
			IPCode:    "1234",
			FirstName: "First",
			Address: models.Address{
				AddressLine1: "addressline1",
				Locality:     "locality",
			},
			Role: constants.FinalLiquidator.String(),
		})
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "last_name is a required field")
	})

	Convey("Incoming request has address missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body, _ := json.Marshal(models.PractitionerRequest{
			IPCode:    "1234",
			FirstName: "First",
			LastName:  "Last",
			Role:      constants.FinalLiquidator.String(),
		})
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "address_line_1 is a required field, locality is a required field")
	})

	Convey("Incoming request has address line 1 missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body, _ := json.Marshal(models.PractitionerRequest{
			IPCode:    "1234",
			FirstName: "First",
			LastName:  "Last",
			Address: models.Address{
				Locality: "locality",
			},
			Role: constants.FinalLiquidator.String(),
		})
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "address_line_1 is a required field")
	})

	Convey("Incoming request has locality missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body, _ := json.Marshal(models.PractitionerRequest{
			IPCode:    "1234",
			FirstName: "First",
			LastName:  "Last",
			Address: models.Address{
				AddressLine1: "addressline1",
			},
			Role: constants.FinalLiquidator.String(),
		})
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "locality is a required field")
	})

	Convey("Incoming request has role missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body, _ := json.Marshal(models.PractitionerRequest{
			IPCode:    "1234",
			FirstName: "First",
			LastName:  "Last",
			Address: models.Address{
				AddressLine1: "addressline1",
				Locality:     "locality",
			},
		})
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "role is a required field")
	})

	Convey("Incoming request has an invalid role", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body, _ := json.Marshal(models.PractitionerRequest{
			IPCode:    "1234",
			FirstName: "First",
			LastName:  "Last",
			Address: models.Address{
				AddressLine1: "addressline1",
				Locality:     "locality",
			},
			Role: "error-role",
		})
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Generic error when adding practitioners resource to mongo", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect CreatePractitionersResource to be called once and return an error
		mockService.EXPECT().CreatePractitionersResource(gomock.Any(), transactionID).Return(fmt.Errorf("there was a problem handling your request for transaction %s", transactionID), http.StatusInternalServerError).Times(1)

		body, _ := json.Marshal(models.PractitionerRequest{
			IPCode:    "1234",
			FirstName: "First",
			LastName:  "Last",
			Address: models.Address{
				AddressLine1: "addressline1",
				Locality:     "locality",
			},
			Role: constants.FinalLiquidator.String(),
		})
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

		body, _ := json.Marshal(models.PractitionerRequest{
			IPCode:    "1234",
			FirstName: "First",
			LastName:  "Last",
			Address: models.Address{
				AddressLine1: "addressline1",
				Locality:     "locality",
			},
			Role: constants.FinalLiquidator.String(),
		})
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

		body, _ := json.Marshal(models.PractitionerRequest{
			IPCode:    "1234",
			FirstName: "First",
			LastName:  "Last",
			Address: models.Address{
				AddressLine1: "addressline1",
				Locality:     "locality",
			},
			Role: constants.FinalLiquidator.String(),
		})
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

		body, _ := json.Marshal(models.PractitionerRequest{
			IPCode:    "1234",
			FirstName: "First",
			LastName:  "Last",
			Address: models.Address{
				AddressLine1: "addressline1",
				Locality:     "locality",
			},
			Role: constants.FinalLiquidator.String(),
		})
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

		body, _ := json.Marshal(models.PractitionerRequest{
			IPCode:    "1234",
			FirstName: "First",
			LastName:  "Last",
			Address: models.Address{
				AddressLine1: "addressline1",
				Locality:     "locality",
			},
			Role: constants.FinalLiquidator.String(),
		})
		res := serveHandleCreatePractitionersResource(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusCreated)
	})

}
