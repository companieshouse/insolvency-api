package service

import (
	"net/http"
	"testing"

	"github.com/companieshouse/insolvency-api/config"
	"github.com/jarcoal/httpmock"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitIsUserOnEfsAllowList(t *testing.T) {
	// Function response is now dependent on config (EFS API call is bypassed if DISABLE_EFS_ALLOW_LIST_AUTH is set true)
	cfg, _ := config.Get()

	Convey("Email auth intercept - DISABLE_EFS_ALLOW_LIST_AUTH unset or false", t, func() {
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

		Convey("user allowed because on EFS allow list", func() {
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

	Convey("Email auth intercept - DISABLE_EFS_ALLOW_LIST_AUTH set true", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// Simulate DISABLE_EFS_ALLOW_LIST_AUTH feature toggle being enabled in the environment
		cfg.IsEfsAllowListAuthDisabled = true

		Convey("user allowed because email address contains magic string and DISABLE_EFS_ALLOW_LIST_AUTH is toggled on in environment", func() {
			req, _ := http.NewRequest("GET", "", nil)

			defer httpmock.Reset()

			userAllowed, err := IsUserOnEfsAllowList("demo-ip-test@ch.gov.uk", req)
			So(userAllowed, ShouldBeTrue)
			So(err, ShouldBeNil)
		})

		Convey("user not allowed because email address does not contain magic string (even though EFS endpoint mocked as true if function tried to call it)", func() {
			req, _ := http.NewRequest("GET", "", nil)

			defer httpmock.Reset()
			httpmock.RegisterResponder(
				http.MethodGet,
				"http://localhost:4001/efs-submission-api/company-authentication/allow-list/demo@ch.gov.uk",
				httpmock.NewStringResponder(http.StatusOK, "true"),
			)

			userAllowed, err := IsUserOnEfsAllowList("demo-test@ch.gov.uk", req)
			So(userAllowed, ShouldBeFalse)
			So(err, ShouldBeNil)
		})

	})
}
