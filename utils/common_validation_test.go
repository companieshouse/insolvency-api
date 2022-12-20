package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/companieshouse/insolvency-api/models"

	. "github.com/smartystreets/goconvey/convey"
)

func prepareForTest() (*http.Request, *httptest.ResponseRecorder) {
	path := "/anything"
	body, _ := json.Marshal(&models.InsolvencyRequest{})
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	res := httptest.NewRecorder()

	return req, res
}

func TestUnitHandleRequestValidation(t *testing.T) {
	helperService := NewHelperService()
	Convey("Fails validation when transaction ID does exists", t, func() {
		var req, res = prepareForTest()
		_, actual, _ := helperService.HandleTransactionIdExistsValidation(res, req, "")

		So(actual, ShouldBeFalse)
	})

	Convey("Passes validation when transaction ID exists", t, func() {
		var req, res = prepareForTest()
		_, actual, _ := helperService.HandleTransactionIdExistsValidation(res, req, "1234567")

		So(actual, ShouldBeTrue)
	})

	Convey("Fails validation when transaction is closed against transaction api", t, func() {
		var req, res = prepareForTest()
		_, actual, _ := helperService.HandleTransactionNotClosedValidation(res, req, "anything", false, errors.New("anything"), http.StatusInternalServerError)

		So(actual, ShouldBeFalse)
	})

	Convey("Fails validation when transaction is already closed and cannot be updated", t, func() {
		var req, res = prepareForTest()
		_, actual, _ := helperService.HandleTransactionNotClosedValidation(res, req, "anything", true, nil, http.StatusInternalServerError)

		So(actual, ShouldBeFalse)
	})

	Convey("Passes validation when transaction is not closed", t, func() {
		var req, res = prepareForTest()
		_, actual, _ := helperService.HandleTransactionNotClosedValidation(res, req, "anything", false, nil, http.StatusInternalServerError)

		So(actual, ShouldBeTrue)
	})

	Convey("Fails validation when invalid request body", t, func() {
		var req, res = prepareForTest()
		actual, _ := helperService.HandleBodyDecodedValidation(res, req, "anything", errors.New("anything"))

		So(actual, ShouldBeFalse)
	})

	Convey("Passes validation when valid request body", t, func() {
		var req, res = prepareForTest()
		actual, _ := helperService.HandleBodyDecodedValidation(res, req, "anything", nil)

		So(actual, ShouldBeTrue)
	})

	Convey("Fails validation when etag generation fails", t, func() {
		actual := helperService.HandleEtagGenerationValidation(errors.New("anything"))

		So(actual, ShouldBeFalse)
	})

	Convey("Passes validation when etag generation succeeds", t, func() {
		actual := helperService.HandleEtagGenerationValidation(nil)

		So(actual, ShouldBeTrue)
	})
}
