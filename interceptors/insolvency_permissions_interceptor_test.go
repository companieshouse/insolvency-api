package interceptors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/companieshouse/chs.go/authentication"

	. "github.com/smartystreets/goconvey/convey"
)

func setTokenHeader(req *http.Request, permissions string) {
	req.Header.Set("ERIC-Authorised-Token-Permissions", permissions)
}

func TestUnitInsolvencyPermissionsIntercept(t *testing.T) {
	Convey("Insolvency permissions intercept", t, func() {

		Convey("Invalid token header", func() {
			req, _ := http.NewRequest("GET", "", nil)
			setTokenHeader(req, "invalid=invalid=invalid")

			w := httptest.NewRecorder()

			test := InsolvencyPermissionsIntercept(getTestHandler())
			test.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Read request", func() {
			req, _ := http.NewRequest("GET", "", nil)
			setTokenHeader(req, authentication.PermissionKeyInsolvencyCases+"=read")

			w := httptest.NewRecorder()

			test := InsolvencyPermissionsIntercept(getTestHandler())
			test.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("Update request", func() {
			req, _ := http.NewRequest("POST", "", nil)
			setTokenHeader(req, authentication.PermissionKeyInsolvencyCases+"=update")

			w := httptest.NewRecorder()

			test := InsolvencyPermissionsIntercept(getTestHandler())
			test.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("Incorrect token permission", func() {
			req, _ := http.NewRequest("POST", "", nil)
			setTokenHeader(req, authentication.PermissionKeyInsolvencyCases+"=read")

			w := httptest.NewRecorder()

			test := InsolvencyPermissionsIntercept(getTestHandler())
			test.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

	})
}
