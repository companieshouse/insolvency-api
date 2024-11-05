package service

import (
	"net/http"
	"testing"

	"github.com/companieshouse/insolvency-api/constants"
	mock_dao "github.com/companieshouse/insolvency-api/mocks"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/transformers"
	"github.com/companieshouse/insolvency-api/utils"
	"github.com/golang/mock/gomock"

	"github.com/jarcoal/httpmock"

	. "github.com/smartystreets/goconvey/convey"
)

func incomingInsolvencyResourceDao(helperService utils.HelperService) *models.InsolvencyResourceDao {
	request := incomingInsolvencyRequest("01234567", "companyName", constants.CVL.String())

	res := transformers.InsolvencyResourceRequestToDB(request, "87654321")
	return res
}

func transactionProfileResponse(status string) string {
	return `
{
 "id": "87654321",
 "company_name": "companyName",
 "company_number": "01234567",
 "status": "` + status + `"
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

			httpmock.RegisterResponder(http.MethodGet, apiURL+"/transactions/87654321", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse("open")))

			err, statusCode := CheckTransactionID("87654321", &http.Request{})

			So(err, ShouldBeNil)
			So(statusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func TestUnitPatchTransactionWithInsolvencyResource(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

	Convey("PatchTransactionWithInsolvencyResourceOnTransactionAPI", t, func() {

		privateApiURL := "http://localhost:4001"

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		Convey("Transaction cannot be found on transaction api", func() {
			defer httpmock.Reset()

			httpmock.RegisterResponder(http.MethodPatch, privateApiURL+"/private/transactions/87654321", httpmock.NewStringResponder(http.StatusNotFound, "Message: Transaction not found"))

			err, statusCode := PatchTransactionWithInsolvencyResource("87654321", incomingInsolvencyResourceDao(mockHelperService), &http.Request{})
			So(err, ShouldNotBeNil)
			So(statusCode, ShouldEqual, http.StatusNotFound)
			So(err.Error(), ShouldEqual, `transaction not found`)
		})

		Convey("Error contacting the transaction api", func() {
			defer httpmock.Reset()

			httpmock.RegisterResponder(http.MethodPatch, privateApiURL+"/private/transactions/87654321", httpmock.NewStringResponder(http.StatusTeapot, ""))

			err, statusCode := PatchTransactionWithInsolvencyResource("87654321", incomingInsolvencyResourceDao(mockHelperService), &http.Request{})
			So(err, ShouldNotBeNil)
			So(statusCode, ShouldEqual, http.StatusTeapot)
			So(err.Error(), ShouldEqual, `error communication with the transaction api`)
		})

		Convey("Provided transaction successfully patched", func() {
			defer httpmock.Reset()

			httpmock.RegisterResponder(http.MethodPatch, privateApiURL+"/private/transactions/87654321", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse("open")))

			err, statusCode := PatchTransactionWithInsolvencyResource("87654321", incomingInsolvencyResourceDao(mockHelperService), &http.Request{})

			So(err, ShouldBeNil)
			So(statusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func TestUnitCheckIfTransactionClosed(t *testing.T) {

	Convey("CheckIfTransactionClosed", t, func() {

		apiURL := "https://api.companieshouse.gov.uk"

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		Convey("Transaction cannot be found on transaction api", func() {
			defer httpmock.Reset()

			httpmock.RegisterResponder(http.MethodGet, apiURL+"/transactions/87654321", httpmock.NewStringResponder(http.StatusNotFound, "Message: Transaction not found"))

			_, err, statusCode := CheckIfTransactionClosed("87654321", &http.Request{})
			So(err, ShouldNotBeNil)
			So(statusCode, ShouldEqual, http.StatusNotFound)
			So(err.Error(), ShouldEqual, `transaction not found`)
		})

		Convey("Error contacting the transaction api", func() {
			defer httpmock.Reset()

			httpmock.RegisterResponder(http.MethodGet, apiURL+"/transactions/87654321", httpmock.NewStringResponder(http.StatusTeapot, ""))

			_, err, statusCode := CheckIfTransactionClosed("87654321", &http.Request{})
			So(err, ShouldNotBeNil)
			So(statusCode, ShouldEqual, http.StatusTeapot)
			So(err.Error(), ShouldContainSubstring, `error getting transaction from transaction api`)
		})

		Convey("Provided transaction successfully returned as a closed transaction", func() {
			defer httpmock.Reset()

			httpmock.RegisterResponder(http.MethodGet, apiURL+"/transactions/87654321", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse("closed")))

			isTransactionClosed, err, statusCode := CheckIfTransactionClosed("87654321", &http.Request{})

			So(isTransactionClosed, ShouldBeTrue)
			So(err, ShouldBeNil)
			So(statusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("Provided transaction successfully returned as an open transaction", func() {
			defer httpmock.Reset()

			httpmock.RegisterResponder(http.MethodGet, apiURL+"/transactions/87654321", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse("open")))

			isTransactionClosed, err, statusCode := CheckIfTransactionClosed("87654321", &http.Request{})

			So(isTransactionClosed, ShouldBeFalse)
			So(err, ShouldBeNil)
			So(statusCode, ShouldEqual, http.StatusOK)
		})
	})
}
