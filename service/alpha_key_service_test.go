package service

import (
	"net/http"
	"testing"

	"github.com/companieshouse/go-sdk-manager/config"
	"github.com/companieshouse/insolvency-api/constants"
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
			httpmock.RegisterResponder(http.MethodGet, "http://localhost:18103/alphakey?name=companyName", httpmock.NewStringResponder(http.StatusOK, alphaKeyResponse("COMPANYNAME")))
			httpmock.RegisterResponder(http.MethodGet, "http://localhost:18103/alphakey?name=COMPANYNAME", httpmock.NewStringResponder(http.StatusOK, alphaKeyResponse("COMPANYNAME")))

			request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())
			err, statusCode := CheckCompanyNameAlphaKey("COMPANYNAME", request, &http.Request{})
			So(err, ShouldBeNil)
			So(statusCode, ShouldEqual, http.StatusOK)

		})

		Convey("Company name does not match with company profile by alphaley", func() {

			httpmock.RegisterResponder(http.MethodGet, "http://localhost:18103/alphakey?name=companyName", httpmock.NewStringResponder(http.StatusOK, alphaKeyResponse("COMPANYNAME")))
			httpmock.RegisterResponder(http.MethodGet, "http://localhost:18103/alphakey?name=ANOTHERNAME", httpmock.NewStringResponder(http.StatusOK, alphaKeyResponse("ANOTHERNAME")))

			request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())
			err, statusCode := CheckCompanyNameAlphaKey("ANOTHERNAME", request, &http.Request{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "company names do not match")
			So(statusCode, ShouldEqual, http.StatusBadRequest)

		})

		Convey("Error contacting the alpha key service api", func() {
			defer httpmock.Reset()
			httpmock.RegisterResponder(http.MethodGet, "http://localhost:18103/alphakey?name=companyName", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

			request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())
			err, statusCode := CheckCompanyNameAlphaKey("companyName", request, &http.Request{})
			So(err, ShouldNotBeNil)
			So(statusCode, ShouldEqual, http.StatusInternalServerError)
			So(err.Error(), ShouldContainSubstring, `error communicating with alphakey service`)
		})

		Convey("alphakeyapi url present in environment variable returns no error", func() {
			cfg, _ := config.Get()
			cfg.ALPHAKEYAPIURL = "http://localhost:8003"
			defer httpmock.Reset()
			httpmock.RegisterResponder(http.MethodGet, "http://localhost:8003/alphakey?name=companyName", httpmock.NewStringResponder(http.StatusOK, alphaKeyResponse("COMPANYNAME")))

			request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())
			err, statusCode := CheckCompanyNameAlphaKey("companyName", request, &http.Request{})
			So(err, ShouldBeNil)
			So(statusCode, ShouldEqual, http.StatusOK)
		})
	})

}
