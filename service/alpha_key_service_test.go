package service

import (
	"github.com/companieshouse/insolvency-api/constants"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

func alphaKeyResponse(companyName string) string {
	return `
	{
		"sameAsAlphaKey": " ` + companyName + `",
		"orderedAlphaKey": "COMPANYNAME",
		"upperCaseName": "COMPANYNAME"
	}`
}
func TestUnitCheckCompanyNameValid(t *testing.T) {

	Convey("CheckCompanyInsolvencyValidFromAlphaKeyServiceAPI", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		Convey("Company name matches with company profile by alphaley", func() {
			httpmock.RegisterResponder(http.MethodGet, "http://localhost:4001/alphakey?name=companyName", httpmock.NewStringResponder(http.StatusOK, alphaKeyResponse("COMPANYNAME")))
			httpmock.RegisterResponder(http.MethodGet, "http://localhost:4001/alphakey?name=COMPANYNAME", httpmock.NewStringResponder(http.StatusOK, alphaKeyResponse("COMPANYNAME")))

			request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())
			err, statusCode := CheckCompanyNameAlphaKey("COMPANYNAME", request, &http.Request{})
			So(err, ShouldBeNil)
			So(statusCode, ShouldEqual, http.StatusOK)

		})

		Convey("Company name does not match with company profile by alphaley", func() {

			httpmock.RegisterResponder(http.MethodGet, "http://localhost:4001/alphakey?name=companyName", httpmock.NewStringResponder(http.StatusOK, alphaKeyResponse("COMPANYNAME")))
			httpmock.RegisterResponder(http.MethodGet, "http://localhost:4001/alphakey?name=ANOTHERNAME", httpmock.NewStringResponder(http.StatusOK, alphaKeyResponse("ANOTHERNAME")))

			request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())
			err, statusCode := CheckCompanyNameAlphaKey("ANOTHERNAME", request, &http.Request{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "company names do not match")
			So(statusCode, ShouldEqual, http.StatusBadRequest)

		})

		Convey("Error contacting the alpha key service api", func() {
			defer httpmock.Reset()
			req, _ := http.NewRequest("GET", "", nil)
			httpmock.RegisterResponder(http.MethodGet, "http://localhost:4001/alphakey/name=companyName", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

			request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())
			err, statusCode := CheckCompanyNameAlphaKey("companyName", request, req)
			So(err, ShouldNotBeNil)
			So(statusCode, ShouldEqual, http.StatusInternalServerError)
			So(err.Error(), ShouldEqual, `error communicating with the company profile api`)
		})
	})
}
