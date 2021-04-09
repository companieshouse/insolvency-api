package service

import (
	"net/http"
	"testing"

	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/transformers"

	"github.com/jarcoal/httpmock"

	. "github.com/smartystreets/goconvey/convey"
)

func incomingInsolvencyResourceDao() *models.InsolvencyResourceDao {
	request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())
	return transformers.InsolvencyResourceRequestToDB(request, "87654321")
}

func transactionProfileResponse() string {
	return `
{
 "id": "87654321",
 "company_name": "companyName",
 "company_number": "01234567"
}
`
}

func TestUnitCheckTransactionID(t *testing.T) {

	Convey("CheckTransactionIDFromTransactionAPI", t, func() {

		apiURL := "https://api.companieshouse.gov.uk"

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		Convey("Transaction cannot be found on transaction api", func() {
			defer httpmock.Reset()

			httpmock.RegisterResponder(http.MethodGet, apiURL+"/transactions/87654321", httpmock.NewStringResponder(http.StatusNotFound, "Message: Transaction not found"))

			err, statusCode := CheckTransactionID("87654321", &http.Request{})
			So(err, ShouldNotBeNil)
			So(statusCode, ShouldEqual, http.StatusNotFound)
			So(err.Error(), ShouldEqual, `transaction not found`)
		})

		Convey("Error contacting the transaction api", func() {
			defer httpmock.Reset()

			httpmock.RegisterResponder(http.MethodGet, apiURL+"/transactions/87654321", httpmock.NewStringResponder(http.StatusTeapot, ""))

			err, statusCode := CheckTransactionID("87654321", &http.Request{})
			So(err, ShouldNotBeNil)
			So(statusCode, ShouldEqual, http.StatusTeapot)
			So(err.Error(), ShouldEqual, `error communicating with the transaction api`)
		})

		Convey("Provided transaction successfully returned", func() {
			defer httpmock.Reset()

			httpmock.RegisterResponder(http.MethodGet, apiURL+"/transactions/87654321", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse()))

			err, statusCode := CheckTransactionID("87654321", &http.Request{})

			So(err, ShouldBeNil)
			So(statusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func TestUnitPatchTransactionWithInsolvencyResource(t *testing.T) {

	Convey("PatchTransactionWithInsolvencyResourceOnTransactionAPI", t, func() {

		privateApiURL := "http://localhost:4001"

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		Convey("Transaction cannot be found on transaction api", func() {
			defer httpmock.Reset()

			httpmock.RegisterResponder(http.MethodPatch, privateApiURL+"/private/transactions/87654321", httpmock.NewStringResponder(http.StatusNotFound, "Message: Transaction not found"))

			err, statusCode := PatchTransactionWithInsolvencyResource("87654321", incomingInsolvencyResourceDao(), &http.Request{})
			So(err, ShouldNotBeNil)
			So(statusCode, ShouldEqual, http.StatusNotFound)
			So(err.Error(), ShouldEqual, `transaction not found`)
		})

		Convey("Error contacting the transaction api", func() {
			defer httpmock.Reset()

			httpmock.RegisterResponder(http.MethodPatch, privateApiURL+"/private/transactions/87654321", httpmock.NewStringResponder(http.StatusTeapot, ""))

			err, statusCode := PatchTransactionWithInsolvencyResource("87654321", incomingInsolvencyResourceDao(), &http.Request{})
			So(err, ShouldNotBeNil)
			So(statusCode, ShouldEqual, http.StatusTeapot)
			So(err.Error(), ShouldEqual, `error communication with the transaction api`)
		})

		Convey("Provided transaction successfully patched", func() {
			defer httpmock.Reset()

			httpmock.RegisterResponder(http.MethodPatch, privateApiURL+"/private/transactions/87654321", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse()))

			err, statusCode := PatchTransactionWithInsolvencyResource("87654321", incomingInsolvencyResourceDao(), &http.Request{})

			So(err, ShouldBeNil)
			So(statusCode, ShouldEqual, http.StatusOK)
		})
	})
}
