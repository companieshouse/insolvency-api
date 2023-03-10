package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/companieshouse/insolvency-api/mocks"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/golang/mock/gomock"
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitIsValidStatementDate(t *testing.T) {
	transactionID := "123"
	apiURL := "https://api.companieshouse.gov.uk"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	Convey("request supplied is invalid - no attachment has been supplied", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService := mocks.NewMockService(mockCtrl)

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResourceDao, nil, nil)

		statement := generateStatement()
		statement.Attachments = []string{}

		validationErr, err := ValidateStatementDetails(mockService, &statement, transactionID, req)

		So(validationErr, ShouldContainSubstring, "please supply only one attachment")
		So(err, ShouldBeNil)
	})

	Convey("request supplied is invalid - more than one attachment has been supplied", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService := mocks.NewMockService(mockCtrl)

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResourceDao, nil, nil)

		statement := generateStatement()
		statement.Attachments = []string{
			"1234567890",
			"0987654321",
		}

		validationErr, err := ValidateStatementDetails(mockService, &statement, transactionID, req)

		So(validationErr, ShouldContainSubstring, "please supply only one attachment")
		So(err, ShouldBeNil)
	})

	Convey("error retrieving insolvency resource", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(&models.InsolvencyResourceDao{}, nil, fmt.Errorf("err"))

		statement := generateStatement()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateStatementDetails(mockService, &statement, transactionID, req)
		So(err.Error(), ShouldContainSubstring, "err")
		So(validationErr, ShouldBeEmpty)
	})

	Convey("error retrieving company details", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusTeapot, ""))

		mockService := mocks.NewMockService(mockCtrl)

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResourceDao, nil, nil)

		statement := generateStatement()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateStatementDetails(mockService, &statement, transactionID, req)
		So(validationErr, ShouldBeEmpty)
		So(err.Error(), ShouldContainSubstring, "error getting company details from DB")
	})

	Convey("error parsing appointedOn date", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService := mocks.NewMockService(mockCtrl)

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResourceDao, nil, nil)

		statement := generateStatement()
		statement.StatementDate = "2001/1/2"

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateStatementDetails(mockService, &statement, transactionID, req)
		So(validationErr, ShouldBeEmpty)
		So(err.Error(), ShouldContainSubstring, "error parsing date")
	})

	Convey("error parsing incorporatedOn date", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("error")))

		statement := generateStatement()
		mockService := mocks.NewMockService(mockCtrl)

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResourceDao, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateStatementDetails(mockService, &statement, transactionID, req)
		So(validationErr, ShouldBeEmpty)
		So(err.Error(), ShouldContainSubstring, "error parsing date")
	})

	Convey("invalid date - in the future", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService := mocks.NewMockService(mockCtrl)

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResourceDao, nil, nil)

		statement := generateStatement()
		statement.StatementDate = time.Now().AddDate(0, 0, 1).Format("2006-01-02")

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateStatementDetails(mockService, &statement, transactionID, req)
		So(validationErr, ShouldContainSubstring, "should not be in the future")
		So(err, ShouldBeNil)
	})

	Convey("invalid appointedOn date - before company was incorporated", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService := mocks.NewMockService(mockCtrl)

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResourceDao, nil, nil)

		statement := generateStatement()
		statement.StatementDate = "1999-01-01"

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateStatementDetails(mockService, &statement, transactionID, req)
		So(validationErr, ShouldContainSubstring, "before the company was incorporated")
		So(err, ShouldBeNil)
	})

	Convey("valid date", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService := mocks.NewMockService(mockCtrl)

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResourceDao, nil, nil)

		statement := generateStatement()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateStatementDetails(mockService, &statement, transactionID, req)
		So(validationErr, ShouldBeEmpty)
		So(err, ShouldBeNil)
	})

	Convey("nil dao", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService := mocks.NewMockService(mockCtrl)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateStatementDetails(mockService, nil, transactionID, req)
		So(validationErr, ShouldBeEmpty)
		So(err.Error(), ShouldContainSubstring, "nil DAO passed to service for validation")
	})

}

func generateStatement() models.StatementOfAffairsResourceDao {
	return models.StatementOfAffairsResourceDao{
		StatementDate: "2012-01-23",
		Attachments: []string{
			"123456789",
		},
	}
}
