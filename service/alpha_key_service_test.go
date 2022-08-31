package service

import (
	"net/http"
	"testing"

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

		httpmock.RegisterResponder(http.MethodGet, "http://localhost:4001/alphakey?name=companyName", httpmock.NewStringResponder(http.StatusOK, alphaKeyResponse("COMPANYNAME")))
		httpmock.RegisterResponder(http.MethodGet, "http://localhost:4001/alphakey?name=COMPANYNAME", httpmock.NewStringResponder(http.StatusOK, alphaKeyResponse("COMPANYNAME")))

		request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())
		err, statusCode := CheckCompanyNameAlphaKey("COMPANYNAME", request, &http.Request{})
		So(err, ShouldBeNil)
		So(statusCode, ShouldEqual, http.StatusOK)

	})

	Convey("CheckCompanyInsolvencyInValidFromAlphaKeyServiceAPI", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodGet, "http://localhost:4001/alphakey?name=companyName", httpmock.NewStringResponder(http.StatusOK, alphaKeyResponse("COMPANYNAME")))
		httpmock.RegisterResponder(http.MethodGet, "http://localhost:4001/alphakey?name=ANOTHERNAME", httpmock.NewStringResponder(http.StatusOK, alphaKeyResponse("ANOTHERNAME")))

		request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())
		err, statusCode := CheckCompanyNameAlphaKey("ANOTHERNAME", request, &http.Request{})
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "company names do not match")
		So(statusCode, ShouldEqual, http.StatusBadRequest)

	})
}
