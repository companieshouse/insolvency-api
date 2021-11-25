package service

import (
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitIsUserOnEfsAllowList(t *testing.T) {

	Convey("Email auth intercept", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		Convey("Error communicating with  EFS API", func() {
			req, _ := http.NewRequest("GET", "", nil)

			defer httpmock.Reset()
			httpmock.RegisterResponder(
				http.MethodGet,
				"http://localhost:4001/efs-submission-api/company-authentication/allow-list/demo@ch.gov.uk",
				httpmock.NewStringResponder(http.StatusInternalServerError, ""),
			)

			userAllowed, err := IsUserOnEfsAllowList("demo@ch.gov.uk", req)
			So(userAllowed, ShouldBeFalse)
			So(err.Error(), ShouldEqual, "error communicating with the EFS submission api: [ch-api: got HTTP response code 500 with body: ]")
		})

		Convey("user allowed", func() {
			req, _ := http.NewRequest("GET", "", nil)

			defer httpmock.Reset()
			httpmock.RegisterResponder(
				http.MethodGet,
				"http://localhost:4001/efs-submission-api/company-authentication/allow-list/demo@ch.gov.uk",
				httpmock.NewStringResponder(http.StatusOK, "true"),
			)

			userAllowed, err := IsUserOnEfsAllowList("demo@ch.gov.uk", req)
			So(userAllowed, ShouldBeTrue)
			So(err, ShouldBeNil)
		})

		Convey("user not allowed", func() {
			req, _ := http.NewRequest("GET", "", nil)

			defer httpmock.Reset()
			httpmock.RegisterResponder(
				http.MethodGet,
				"http://localhost:4001/efs-submission-api/company-authentication/allow-list/demo@ch.gov.uk",
				httpmock.NewStringResponder(http.StatusOK, "false"),
			)

			userAllowed, err := IsUserOnEfsAllowList("demo@ch.gov.uk", req)
			So(userAllowed, ShouldBeFalse)
			So(err, ShouldBeNil)
		})
	})
}
