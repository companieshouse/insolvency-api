package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/companieshouse/insolvency-api/mocks"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/golang/mock/gomock"
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestValidProgressReport(t *testing.T) {
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
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)

		progressReport := generateProgressReport()
		progressReport.Attachments = []string{}

		validationErr, err := ValidateProgressReportDetails(mockService, &progressReport, transactionID, req)

		So(validationErr, ShouldContainSubstring, "please supply only one attachment")
		So(err, ShouldBeNil)
	})

	Convey("request supplied is invalid - more than one attachment has been supplied", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)

		progressReport := generateProgressReport()
		progressReport.Attachments = []string{
			"1234567890",
			"0987654321",
		}

		validationErr, err := ValidateProgressReportDetails(mockService, &progressReport, transactionID, req)

		So(validationErr, ShouldContainSubstring, "please supply only one attachment")
		So(err, ShouldBeNil)
	})

	Convey("valid date", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)

		progressReport := generateProgressReport()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateProgressReportDetails(mockService, &progressReport, transactionID, req)
		So(validationErr, ShouldBeEmpty)
		So(err, ShouldBeNil)
	})
}

func generateProgressReport() models.ProgressReportResourceDao {
	return models.ProgressReportResourceDao{
		FromDate: "2021-06-06",
		ToDate:   "2021-06-07",
		Attachments: []string{
			"123456789",
		},
	}
}
