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

func TestUnitIsValidResolutionRequest(t *testing.T) {
	Convey("Resolution request supplied is invalid - no attachment has been supplied", t, func() {
		resolution := generateResolution()
		resolution.Attachments = []string{}

		err := ValidateResolutionRequest(models.Resolution(resolution))

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "please supply only one attachment")
	})

	Convey("Practitioner request supplied is invalid - more than one attachment has been supplied", t, func() {
		resolution := generateResolution()
		resolution.Attachments = []string{
			"1234567890",
			"0987654321",
		}

		err := ValidateResolutionRequest(models.Resolution(resolution))

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "please supply only one attachment")
	})
}

func TestUnitIsValidResolutionDate(t *testing.T) {
	transactionID := "123"
	apiURL := "https://api.companieshouse.gov.uk"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	Convey("error retrieving insolvency resource", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(models.InsolvencyResourceDao{}, fmt.Errorf("err"))

		resolution := generateResolution()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateResolutionDate(mockService, &resolution, transactionID, req)
		So(err.Error(), ShouldContainSubstring, "err")
		So(validationErr, ShouldBeEmpty)
	})

	Convey("error retrieving company details", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusTeapot, ""))

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)

		resolution := generateResolution()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateResolutionDate(mockService, &resolution, transactionID, req)
		So(validationErr, ShouldBeEmpty)
		So(err.Error(), ShouldContainSubstring, "error getting company details from DB")
	})

	Convey("error parsing appointedOn date", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)

		resolution := generateResolution()
		resolution.DateOfResolution = "2001/1/2"

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateResolutionDate(mockService, &resolution, transactionID, req)
		So(validationErr, ShouldBeEmpty)
		So(err.Error(), ShouldContainSubstring, "error parsing date")
	})

	Convey("error parsing incorporatedOn date", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("error")))

		resolution := generateResolution()
		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateResolutionDate(mockService, &resolution, transactionID, req)
		So(validationErr, ShouldBeEmpty)
		So(err.Error(), ShouldContainSubstring, "error parsing date")
	})

	Convey("invalid date - in the future", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)

		resolution := generateResolution()
		resolution.DateOfResolution = time.Now().AddDate(0, 0, 1).Format("2006-01-02")

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateResolutionDate(mockService, &resolution, transactionID, req)
		So(validationErr, ShouldContainSubstring, "should not be in the future")
		So(err, ShouldBeNil)
	})

	Convey("invalid appointedOn date - before company was incorporated", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)

		resolution := generateResolution()
		resolution.DateOfResolution = "1999-01-01"

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateResolutionDate(mockService, &resolution, transactionID, req)
		So(validationErr, ShouldContainSubstring, "before the company was incorporated")
		So(err, ShouldBeNil)
	})

	Convey("valid appointment", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)

		resolution := generateResolution()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateResolutionDate(mockService, &resolution, transactionID, req)
		So(validationErr, ShouldBeEmpty)
		So(err, ShouldBeNil)
	})

}

func generateResolution() models.ResolutionResourceDao {
	return models.ResolutionResourceDao{
		DateOfResolution: "2012-01-23",
		Attachments: []string{
			"123456789",
		},
	}
}
