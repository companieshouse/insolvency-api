package service

import (
	"net/http"
	"testing"

	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestValidProgressReport(t *testing.T) {
	transactionID := "123"
	apiURL := "https://api.companieshouse.gov.uk"

	Convey("request supplied is invalid - no attachment has been supplied", t, func() {
		mockService, _, _ := utils.CreateTestObjects(t)
		httpmock.Activate()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)

		progressReport := generateProgressReport()
		progressReport.Attachments = []string{}

		validationErr, err := ValidateProgressReportDetails(mockService, &progressReport, transactionID, req)

		So(validationErr, ShouldContainSubstring, "please supply only one attachment")
		So(err, ShouldBeNil)
	})

	Convey("request supplied is invalid - more than one attachment has been supplied", t, func() {
		mockService, _, _ := utils.CreateTestObjects(t)
		httpmock.Activate()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

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
