package interceptors

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/companieshouse/chs.go/authentication"
	"github.com/jarcoal/httpmock"

	. "github.com/smartystreets/goconvey/convey"
)

func getTestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func testContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(
		ctx,
		authentication.ContextKeyUserDetails,
		authentication.AuthUserDetails{Email: "demo@companieshouse.gov.uk"},
	)
	return ctx
}

func invalidTestContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(
		ctx,
		authentication.ContextKeyUserDetails,
		"invalid",
	)
	return ctx
}

func TestUnitEmailAuthIntercept(t *testing.T) {
	Convey("Email auth intercept", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		Convey("Invalid user details in context", func() {
			req, _ := http.NewRequestWithContext(invalidTestContext(), "GET", "", nil)

			w := httptest.NewRecorder()
			test := EmailAuthIntercept(getTestHandler())
			test.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Error checking EFS allow list", func() {
			req, _ := http.NewRequestWithContext(testContext(), "GET", "", nil)

			defer httpmock.Reset()
			httpmock.RegisterResponder(
				http.MethodGet,
				"http://localhost:4001/efs-submission-api/company-authentication/allow-list/demo@companieshouse.gov.uk",
				httpmock.NewStringResponder(http.StatusInternalServerError, ""),
			)

			w := httptest.NewRecorder()
			test := EmailAuthIntercept(getTestHandler())
			test.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("User not allowed", func() {
			req, _ := http.NewRequestWithContext(testContext(), "GET", "", nil)

			defer httpmock.Reset()
			httpmock.RegisterResponder(
				http.MethodGet,
				"http://localhost:4001/efs-submission-api/company-authentication/allow-list/demo@companieshouse.gov.uk?",
				httpmock.NewStringResponder(http.StatusOK, "false"),
			)

			w := httptest.NewRecorder()
			test := EmailAuthIntercept(getTestHandler())
			test.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("User allowed", func() {
			req, _ := http.NewRequestWithContext(testContext(), "GET", "", nil)

			defer httpmock.Reset()
			httpmock.RegisterResponder(
				http.MethodGet,
				"http://localhost:4001/efs-submission-api/company-authentication/allow-list/demo@companieshouse.gov.uk?",
				httpmock.NewStringResponder(http.StatusOK, "true"),
			)

			w := httptest.NewRecorder()
			test := EmailAuthIntercept(getTestHandler())
			test.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusOK)
		})
	})
}
