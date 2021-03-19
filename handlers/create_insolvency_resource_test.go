package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/constants"
	mock_dao "github.com/companieshouse/insolvency-api/mocks"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/gorilla/mux"

	"github.com/companieshouse/insolvency-api/dao"
	"github.com/golang/mock/gomock"
	"github.com/jarcoal/httpmock"

	. "github.com/smartystreets/goconvey/convey"
)

const transactionID = "12345678"

func serveHandleCreateInsolvencyResource(body []byte, service dao.Service, tranIdSet bool) *httptest.ResponseRecorder {
	path := "/transactions/123456789/insolvency"
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	if tranIdSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}
	res := httptest.NewRecorder()

	handler := HandleCreateInsolvencyResource(service)
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
			CompanyName: "companyName",
		})
		res := serveHandleCreateInsolvencyResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Incoming request has company name missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:      constants.MVL.String(),
			CompanyNumber: "12345678",
		})
		res := serveHandleCreateInsolvencyResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Incoming request has case type missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CompanyNumber: "12345678",
			CompanyName:   "companyName",
		})
		res := serveHandleCreateInsolvencyResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Incoming case type is not CVL", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:      constants.MVL.String(),
			CompanyNumber: "12345678",
			CompanyName:   "companyName",
		})
		res := serveHandleCreateInsolvencyResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error adding insolvency resource to mongo", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect CreateInsolvencyResource to be called once and return an error
		mockService.EXPECT().CreateInsolvencyResource(gomock.Any()).Return(errors.New("error when creating mongo resource")).Times(1)

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:      constants.CVL.String(),
			CompanyName:   "companyName",
			CompanyNumber: "12345678",
		})
		res := serveHandleCreateInsolvencyResource(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Successfully add insolvency resource to mongo", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		// Expect CreateInsolvencyResource to be called once and not return an error
		mockService.EXPECT().CreateInsolvencyResource(gomock.Any()).Return(nil).Times(1)

		body, _ := json.Marshal(&models.InsolvencyRequest{
			CaseType:      constants.CVL.String(),
			CompanyName:   "companyName",
			CompanyNumber: "12345678",
		})
		res := serveHandleCreateInsolvencyResource(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusCreated)
		// TODO: Check call to transaction API to update transaction resource
	})

}

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

		body, _ := json.Marshal([]models.PractitionerRequest{
			{
				FirstName: "First",
				LastName:  "Last",
				Address: models.Address{
					AddressLine1: "addressline1",
					Locality:     "locality",
				},
				Role: constants.FinalLiquidator.String(),
			},
		})
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Incoming request has first name missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body, _ := json.Marshal([]models.PractitionerRequest{
			{
				IPCode:   "1234",
				LastName: "Last",
				Address: models.Address{
					AddressLine1: "addressline1",
					Locality:     "locality",
				},
				Role: constants.FinalLiquidator.String(),
			},
		})
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Incoming request has last name missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body, _ := json.Marshal([]models.PractitionerRequest{
			{
				IPCode:    "1234",
				FirstName: "First",
				Address: models.Address{
					AddressLine1: "addressline1",
					Locality:     "locality",
				},
				Role: constants.FinalLiquidator.String(),
			},
		})
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Incoming request has address missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body, _ := json.Marshal([]models.PractitionerRequest{
			{
				IPCode:    "1234",
				FirstName: "First",
				LastName:  "Last",
				Role:      constants.FinalLiquidator.String(),
			},
		})
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Incoming request has address line 1 missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body, _ := json.Marshal([]models.PractitionerRequest{
			{
				IPCode:    "1234",
				FirstName: "First",
				LastName:  "Last",
				Address: models.Address{
					Locality: "locality",
				},
				Role: constants.FinalLiquidator.String(),
			},
		})
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Incoming request has locality missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body, _ := json.Marshal([]models.PractitionerRequest{
			{
				IPCode:    "1234",
				FirstName: "First",
				LastName:  "Last",
				Address: models.Address{
					AddressLine1: "addressline1",
				},
				Role: constants.FinalLiquidator.String(),
			},
		})
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Incoming request has role missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body, _ := json.Marshal([]models.PractitionerRequest{
			{
				IPCode:    "1234",
				FirstName: "First",
				LastName:  "Last",
				Address: models.Address{
					AddressLine1: "addressline1",
					Locality:     "locality",
				},
			},
		})
		res := serveHandleCreatePractitionersResource(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Incoming request has an invalid role", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		body, _ := json.Marshal([]models.PractitionerRequest{
			{
				IPCode:    "1234",
				FirstName: "First",
				LastName:  "Last",
				Address: models.Address{
					AddressLine1: "addressline1",
					Locality:     "locality",
				},
				Role: "error-role",
			},
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
		mockService.EXPECT().CreatePractitionersResource(gomock.Any(), transactionID).Return(errors.New("error when creating mongo resource")).Times(1)

		body, _ := json.Marshal([]models.PractitionerRequest{
			{
				IPCode:    "1234",
				FirstName: "First",
				LastName:  "Last",
				Address: models.Address{
					AddressLine1: "addressline1",
					Locality:     "locality",
				},
				Role: constants.FinalLiquidator.String(),
			},
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
		mockService.EXPECT().CreatePractitionersResource(gomock.Any(), transactionID).Return(dao.ErrorNotFound).Times(1)

		body, _ := json.Marshal([]models.PractitionerRequest{
			{
				IPCode:    "1234",
				FirstName: "First",
				LastName:  "Last",
				Address: models.Address{
					AddressLine1: "addressline1",
					Locality:     "locality",
				},
				Role: constants.FinalLiquidator.String(),
			},
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
		mockService.EXPECT().CreatePractitionersResource(gomock.Any(), transactionID).Return(dao.ErrorPractitionerLimitReached).Times(1)

		body, _ := json.Marshal([]models.PractitionerRequest{
			{
				IPCode:    "1234",
				FirstName: "First",
				LastName:  "Last",
				Address: models.Address{
					AddressLine1: "addressline1",
					Locality:     "locality",
				},
				Role: constants.FinalLiquidator.String(),
			},
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
		mockService.EXPECT().CreatePractitionersResource(gomock.Any(), transactionID).Return(dao.ErrorPractitionerLimitWillExceed).Times(1)

		body, _ := json.Marshal([]models.PractitionerRequest{
			{
				IPCode:    "1234",
				FirstName: "First",
				LastName:  "Last",
				Address: models.Address{
					AddressLine1: "addressline1",
					Locality:     "locality",
				},
				Role: constants.FinalLiquidator.String(),
			},
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
		mockService.EXPECT().CreatePractitionersResource(gomock.Any(), transactionID).Return(nil).Times(1)

		body, _ := json.Marshal([]models.PractitionerRequest{
			{
				IPCode:    "1234",
				FirstName: "First",
				LastName:  "Last",
				Address: models.Address{
					AddressLine1: "addressline1",
					Locality:     "locality",
				},
				Role: constants.FinalLiquidator.String(),
			},
		})
		res := serveHandleCreatePractitionersResource(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusCreated)
	})

}
