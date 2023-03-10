package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// panickingHandler is an http handler for test use, that just panics
func panickingHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		panic("Message from runtime panic")
	})
}

// notPanickingHandler is an http handler for test use, that does not panic
func notPanickingHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w, "Everything is OK")
	})
}

func TestUnitRecoveryHandler(t *testing.T) {
	Convey("Recovery handler should catch runtime panics", t, func() {
		handler := RecoveryHandler(panickingHandler())
		req := httptest.NewRequest(http.MethodGet, "/test", bytes.NewReader(nil))
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		So(res.Code, ShouldEqual, http.StatusInternalServerError)
		So(res.Body.String(), ShouldContainSubstring, "there was a problem handling your request")
	})

	Convey("Recovery handler should silently wrap successful handlers", t, func() {
		handler := RecoveryHandler(notPanickingHandler())
		req := httptest.NewRequest(http.MethodGet, "/test", bytes.NewReader(nil))
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		So(res.Code, ShouldEqual, http.StatusOK)
		So(res.Body.String(), ShouldEqual, "Everything is OK\n")
	})
}
