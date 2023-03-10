package service

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/companieshouse/insolvency-api/mocks"

	"github.com/companieshouse/insolvency-api/models"
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitValidProgressReport(t *testing.T) {
	transactionID := "123"

	insolvencyResourceDao, practitionerResourceDao, _ := generateInsolvencyPractitionerAppointmentResources()
	practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

	Convey("request supplied is invalid - no attachment has been supplied", t, func() {
		mockService, _, _ := mocks.CreateTestObjects(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResourceDao, practitionerResourceDaos, nil)

		progressReport := generateProgressReport()
		progressReport.Attachments = []string{}

		validationErr, err := ValidateProgressReportDetails(mockService, &progressReport, transactionID, req)

		So(validationErr, ShouldContainSubstring, "please supply only one attachment")
		So(err, ShouldBeNil)
	})

	Convey("request supplied is invalid - more than one attachment has been supplied", t, func() {
		mockService, _, _ := mocks.CreateTestObjects(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResourceDao, practitionerResourceDaos, nil)

		progressReport := generateProgressReport()
		progressReport.Attachments = []string{
			"1234567890",
			"0987654321",
		}

		validationErr, err := ValidateProgressReportDetails(mockService, &progressReport, transactionID, req)

		So(validationErr, ShouldContainSubstring, "please supply only one attachment")
		So(err, ShouldBeNil)
	})

	Convey("error retrieving insolvency resource", t, func() {
		mockService, _, _ := mocks.CreateTestObjects(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(&models.InsolvencyResourceDao{}, nil, fmt.Errorf("error"))

		progressReport := generateProgressReport()

		validationErr, err := ValidateProgressReportDetails(mockService, &progressReport, transactionID, req)

		So(validationErr, ShouldBeEmpty)
		So(err.Error(), ShouldContainSubstring, "error getting insolvency resource from DB")
	})

	Convey("error retrieving company details", t, func() {
		mockService, _, _ := mocks.CreateTestObjects(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusTeapot, ""))

		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResourceDao, practitionerResourceDaos, nil)

		progressReport := generateProgressReport()

		validationErr, err := ValidateProgressReportDetails(mockService, &progressReport, transactionID, req)

		So(validationErr, ShouldBeEmpty)
		So(err.Error(), ShouldContainSubstring, "error communicating with the company profile api")
	})

	Convey("error parsing from date", t, func() {
		mockService, _, _ := mocks.CreateTestObjects(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResourceDao, practitionerResourceDaos, nil)

		progressReport := generateProgressReport()
		progressReport.FromDate = "2001/1/2"

		validationErr, err := ValidateProgressReportDetails(mockService, &progressReport, transactionID, req)
		So(validationErr, ShouldBeEmpty)
		So(err.Error(), ShouldContainSubstring, "error parsing date")
	})

	Convey("error parsing to date", t, func() {
		mockService, _, _ := mocks.CreateTestObjects(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResourceDao, practitionerResourceDaos, nil)

		progressReport := generateProgressReport()
		progressReport.ToDate = "2001/1/2"

		validationErr, err := ValidateProgressReportDetails(mockService, &progressReport, transactionID, req)
		So(validationErr, ShouldBeEmpty)
		So(err.Error(), ShouldContainSubstring, "error parsing date")
	})

	Convey("invalid from date - in the future", t, func() {
		mockService, _, _ := mocks.CreateTestObjects(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResourceDao, practitionerResourceDaos, nil)

		progressReport := generateProgressReport()
		progressReport.FromDate = time.Now().AddDate(0, 0, 1).Format("2006-01-02")

		validationErr, err := ValidateProgressReportDetails(mockService, &progressReport, transactionID, req)
		So(validationErr, ShouldContainSubstring, "should not be in the future")
		So(err, ShouldBeNil)
	})

	Convey("invalid to date - in the future", t, func() {
		mockService, _, _ := mocks.CreateTestObjects(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResourceDao, practitionerResourceDaos, nil)

		progressReport := generateProgressReport()
		progressReport.ToDate = time.Now().AddDate(0, 0, 1).Format("2006-01-02")

		validationErr, err := ValidateProgressReportDetails(mockService, &progressReport, transactionID, req)
		So(validationErr, ShouldContainSubstring, "should not be in the future")
		So(err, ShouldBeNil)
	})

	Convey("invalid from date - before company was incorporated", t, func() {
		mockService, _, _ := mocks.CreateTestObjects(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResourceDao, practitionerResourceDaos, nil)

		progressReport := generateProgressReport()
		progressReport.FromDate = "1999-01-01"

		validationErr, err := ValidateProgressReportDetails(mockService, &progressReport, transactionID, req)
		So(validationErr, ShouldContainSubstring, "from_date")
		So(validationErr, ShouldContainSubstring, "before the company was incorporated")
		So(err, ShouldBeNil)
	})

	Convey("invalid to date - before company was incorporated", t, func() {
		mockService, _, _ := mocks.CreateTestObjects(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		//mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResourceDao, practitionerResourceDaos, nil)

		progressReport := generateProgressReport()
		progressReport.ToDate = "1999-01-01"

		validationErr, err := ValidateProgressReportDetails(mockService, &progressReport, transactionID, req)
		So(validationErr, ShouldContainSubstring, "to_date")
		So(validationErr, ShouldContainSubstring, "before the company was incorporated")
		So(err, ShouldBeNil)
	})

	Convey("valid dates", t, func() {
		mockService, _, _ := mocks.CreateTestObjects(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResourceDao, practitionerResourceDaos, nil)

		progressReport := generateProgressReport()

		validationErr, err := ValidateProgressReportDetails(mockService, &progressReport, transactionID, req)
		So(validationErr, ShouldBeEmpty)
		So(err, ShouldBeNil)
	})

	Convey("nil dao", t, func() {
		mockService, _, _ := mocks.CreateTestObjects(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		validationErr, err := ValidateProgressReportDetails(mockService, nil, transactionID, req)
		So(validationErr, ShouldBeEmpty)
		So(err.Error(), ShouldContainSubstring, "nil DAO passed to service for validation")
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
