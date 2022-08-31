package service

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/companieshouse/api-sdk-go/companieshouseapi"
	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

func incomingInsolvencyRequest(companyNumber string, companyName string, caseType string) *models.InsolvencyRequest {
	return &models.InsolvencyRequest{
		CompanyNumber: companyNumber,
		CompanyName:   companyName,
		CaseType:      caseType,
	}
}

func companyProfileResponse(jurisdiction string, companyStatus string, companyType string) string {
	return `
{
 "company_name": "COMPANYNAME",
 "company_number": "01234567",
 "jurisdiction": "` + jurisdiction + `",
 "company_status": "` + companyStatus + `",
 "date_of_creation": "2000-06-26 00:00:00.000Z",
 "type": "` + companyType + `",
 "registered_office_address" : {
   "postal_code" : "CF14 3UZ",
   "address_line_2" : "Cardiff",
   "address_line_1" : "1 Crown Way"
  }
}
`
}

var apiURL = "https://api.companieshouse.gov.uk"

func TestUnitCheckCompanyInsolvencyValid(t *testing.T) {

	Convey("CheckCompanyInsolvencyValidFromCompanyProfileAPI", t, func() {
		var companyProfile *companieshouseapi.CompanyProfile
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		Convey("Company cannot be found on company profile api", func() {
			defer httpmock.Reset()
			httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/01234567", httpmock.NewStringResponder(http.StatusNotFound, "Message: Company not found"))

			request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())
			err, statusCode, companyProfile := CheckCompanyExists(request, &http.Request{})
			So(err, ShouldNotBeNil)
			So(statusCode, ShouldEqual, http.StatusNotFound)
			So(err.Error(), ShouldEqual, `company not found`)
			So(companyProfile, ShouldBeNil)
		})

		Convey("Error contacting the company profile api", func() {
			defer httpmock.Reset()
			httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/01234567", httpmock.NewStringResponder(http.StatusTeapot, ""))

			request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())
			err, statusCode, companyProfile := CheckCompanyExists(request, &http.Request{})
			So(err, ShouldNotBeNil)
			So(statusCode, ShouldEqual, http.StatusTeapot)
			So(err.Error(), ShouldEqual, `error communicating with the company profile api`)
			So(companyProfile, ShouldBeNil)
		})

		Convey("Jurisdiction of company is not allowed to create insolvency case", func() {

			var companyProfile *companieshouseapi.CompanyProfile
			json.Unmarshal([]byte(companyProfileResponse("scotland", "active", "private-shares-exemption-30")), &companyProfile)

			request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())
			err := CheckCompanyDetailsAreValid(companyProfile, request)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, `jurisdiction [scotland] not permitted`)
		})

		Convey("Company status is not allowed to create insolvency case", func() {
			json.Unmarshal([]byte(companyProfileResponse("england-wales", "dissolved", "private-shares-exemption-30")), &companyProfile)
			request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())
			// err, statusCode := CheckCompanyInsolvencyValid(request, &http.Request{})
			err := CheckCompanyDetailsAreValid(companyProfile, request)
			So(err, ShouldNotBeNil)
			// So(statusCode, ShouldEqual, http.StatusBadRequest)
			So(err.Error(), ShouldEqual, `company status [dissolved] not permitted`)
		})

		Convey("Company type is not allowed to create insolvency case", func() {
			json.Unmarshal([]byte(companyProfileResponse("england-wales", "active", "converted-or-closed")), &companyProfile)
			request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())
			err := CheckCompanyDetailsAreValid(companyProfile, request)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, `company type [converted-or-closed] not permitted`)
		})

		Convey("Company is allowed to start insolvency case", func() {
			json.Unmarshal([]byte(companyProfileResponse("england-wales", "active", "private-shares-exemption-30")), &companyProfile)
			request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())
			err := CheckCompanyDetailsAreValid(companyProfile, request)
			So(err, ShouldBeNil)
		})
	})
}

func TestUnitGetCompanyIncorporatedOn(t *testing.T) {
	Convey("CheckCompanyInsolvencyValidFromCompanyProfileAPI", t, func() {

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		companyNumber := "01234567"

		Convey("Company cannot be found on company profile api", func() {
			defer httpmock.Reset()
			httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/01234567", httpmock.NewStringResponder(http.StatusNotFound, "Message: Company not found"))

			date, err := GetCompanyIncorporatedOn(companyNumber, &http.Request{})
			So(date, ShouldBeEmpty)
			So(err.Error(), ShouldEqual, `company not found`)
		})

		Convey("Error contacting the company profile api", func() {
			defer httpmock.Reset()
			httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/01234567", httpmock.NewStringResponder(http.StatusTeapot, ""))

			date, err := GetCompanyIncorporatedOn(companyNumber, &http.Request{})
			So(date, ShouldBeEmpty)
			So(err.Error(), ShouldEqual, `error communicating with the company profile api`)
		})

		Convey("Successfully retrieve incorporated on date", func() {
			defer httpmock.Reset()
			httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/01234567", httpmock.NewStringResponder(http.StatusOK, companyProfileResponse("england-wales", "active", "private-shares-exemption-30")))

			date, err := GetCompanyIncorporatedOn(companyNumber, &http.Request{})
			So(date, ShouldEqual, "2000-06-26 00:00:00.000Z")
			So(err, ShouldBeNil)
		})
	})
}
