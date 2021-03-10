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

func serveHandleCreateInsolvencyResource(body []byte, service dao.Service, tranIdSet bool) *httptest.ResponseRecorder {
	path := "/transactions/123456789/insolvency"
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	if tranIdSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": "12345678"})
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
