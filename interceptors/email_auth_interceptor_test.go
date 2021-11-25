package interceptors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jarcoal/httpmock"

	. "github.com/smartystreets/goconvey/convey"
)

func GetTestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func TestUnitEmailAuthIntercept(t *testing.T) {
	Convey("Email auth intercept", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		Convey("Error checking EFS allow list", func() {
			req, _ := http.NewRequest("GET", "", nil)
			req.Header.Set("ERIC-Authorised-User", "demo@companieshouse.gov.uk test")

			defer httpmock.Reset()
			httpmock.RegisterResponder(
				http.MethodGet,
				"http://localhost:4001/efs-submission-api/company-authentication/allow-list/demo@companieshouse.gov.uk",
				httpmock.NewStringResponder(http.StatusInternalServerError, ""),
			)

			w := httptest.NewRecorder()
			test := EmailAuthIntercept(GetTestHandler())
			test.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("User not allowed", func() {
			req, _ := http.NewRequest("GET", "", nil)
			req.Header.Set("ERIC-Authorised-User", "demo@companieshouse.gov.uk test")

			defer httpmock.Reset()
			httpmock.RegisterResponder(
				http.MethodGet,
				"http://localhost:4001/efs-submission-api/company-authentication/allow-list/demo@companieshouse.gov.uk?",
				httpmock.NewStringResponder(http.StatusOK, "false"),
			)

			w := httptest.NewRecorder()
			test := EmailAuthIntercept(GetTestHandler())
			test.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("User allowed", func() {
			req, _ := http.NewRequest("GET", "", nil)
			req.Header.Set("ERIC-Authorised-User", "demo@companieshouse.gov.uk test")

			defer httpmock.Reset()
			httpmock.RegisterResponder(
				http.MethodGet,
				"http://localhost:4001/efs-submission-api/company-authentication/allow-list/demo@companieshouse.gov.uk?",
				httpmock.NewStringResponder(http.StatusOK, "true"),
			)

			w := httptest.NewRecorder()
			test := EmailAuthIntercept(GetTestHandler())
			test.ServeHTTP(w, req)
			So(w.Code, ShouldEqual, http.StatusOK)
		})

	})
}
