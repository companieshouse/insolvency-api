package service

import (
	"net/http"
	"testing"

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
 "company_name": "companyName",
 "company_number": "01234567",
 "jurisdiction": "` + jurisdiction + `",
 "company_status": "` + companyStatus + `",
 "type": "` + companyType + `",
 "registered_office_address" : {
   "postal_code" : "CF14 3UZ",
   "address_line_2" : "Cardiff",
   "address_line_1" : "1 Crown Way"
  }
}
`

}

func TestUnitCheckCompanyInsolvencyValid(t *testing.T) {

	Convey("CheckCompanyInsolvencyValidFromCompanyProfileAPI", t, func() {

		apiURL := "https://api.companieshouse.gov.uk"

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		Convey("Company cannot be found on company profile api", func() {
			defer httpmock.Reset()
			httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/01234567", httpmock.NewStringResponder(http.StatusNotFound, "Message: Company not found"))

			request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())
			err, statusCode := CheckCompanyInsolvencyValid(request, &http.Request{})
			So(err, ShouldNotBeNil)
			So(statusCode, ShouldEqual, http.StatusNotFound)
			So(err.Error(), ShouldEqual, `company not found`)
		})

		Convey("Error contacting the company profile api", func() {
			defer httpmock.Reset()
			httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/01234567", httpmock.NewStringResponder(http.StatusTeapot, ""))

			request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())
			err, statusCode := CheckCompanyInsolvencyValid(request, &http.Request{})
			So(err, ShouldNotBeNil)
			So(statusCode, ShouldEqual, http.StatusTeapot)
			So(err.Error(), ShouldEqual, `error communicating with the company profile api`)
		})

		Convey("Provided company name does not match company profile api", func() {
			defer httpmock.Reset()
			httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/01234567", httpmock.NewStringResponder(http.StatusOK,
				companyProfileResponse("england-wales", "active", "private-shares-exemption-30")))

			request := incomingInsolvencyRequest("01234567", "wrongName", constants.CVL.String())
			err, statusCode := CheckCompanyInsolvencyValid(request, &http.Request{})
			So(err, ShouldNotBeNil)
			So(statusCode, ShouldEqual, http.StatusBadRequest)
			So(err.Error(), ShouldEqual, `company names do not match - provided: [wrongName], expected: [companyName]`)
		})

		Convey("Jurisdiction of company is not allowed to create insolvency case", func() {
			defer httpmock.Reset()
			httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/01234567", httpmock.NewStringResponder(http.StatusOK,
				companyProfileResponse("scotland", "active", "private-shares-exemption-30")))

			request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())
			err, statusCode := CheckCompanyInsolvencyValid(request, &http.Request{})
			So(err, ShouldNotBeNil)
			So(statusCode, ShouldEqual, http.StatusBadRequest)
			So(err.Error(), ShouldEqual, `jurisdiction [scotland] not permitted`)
		})

		Convey("Company status is not allowed to create insolvency case", func() {
			defer httpmock.Reset()
			httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/01234567", httpmock.NewStringResponder(http.StatusOK,
				companyProfileResponse("england-wales", "dissolved", "private-shares-exemption-30")))

			request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())
			err, statusCode := CheckCompanyInsolvencyValid(request, &http.Request{})
			So(err, ShouldNotBeNil)
			So(statusCode, ShouldEqual, http.StatusBadRequest)
			So(err.Error(), ShouldEqual, `company status [dissolved] not permitted`)
		})

		Convey("Company type is not allowed to create insolvency case", func() {
			defer httpmock.Reset()
			httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/01234567", httpmock.NewStringResponder(http.StatusOK,
				companyProfileResponse("england-wales", "active", "converted-or-closed")))

			request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())
			err, statusCode := CheckCompanyInsolvencyValid(request, &http.Request{})
			So(err, ShouldNotBeNil)
			So(statusCode, ShouldEqual, http.StatusBadRequest)
			So(err.Error(), ShouldEqual, `company type [converted-or-closed] not permitted`)
		})

		Convey("Company is allowed to start insolvency case", func() {
			defer httpmock.Reset()
			httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/01234567", httpmock.NewStringResponder(http.StatusOK,
				companyProfileResponse("england-wales", "active", "private-shares-exemption-30")))

			request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())
			err, statusCode := CheckCompanyInsolvencyValid(request, &http.Request{})
			So(err, ShouldBeNil)
			So(statusCode, ShouldEqual, http.StatusOK)
		})
	})
}
