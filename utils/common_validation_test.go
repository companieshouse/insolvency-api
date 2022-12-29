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
		_, actual, _ := helperService.HandleTransactionNotClosedValidation(res, req, "anything", false, http.StatusInternalServerError, errors.New("anything"))

		So(actual, ShouldBeFalse)
	})

	Convey("Fails validation when transaction is already closed and cannot be updated", t, func() {
		var req, res = prepareForTest()
		_, actual, _ := helperService.HandleTransactionNotClosedValidation(res, req, "anything", true, http.StatusInternalServerError, nil)

		So(actual, ShouldBeFalse)
	})

	Convey("Passes validation when transaction is not closed", t, func() {
		var req, res = prepareForTest()
		_, actual, _ := helperService.HandleTransactionNotClosedValidation(res, req, "anything", false, http.StatusInternalServerError, nil)

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

	Convey("Fails validation when missing mandatory field values check fails", t, func() {
		var req, res = prepareForTest()
		actual, httpStatusCode := helperService.HandleMandatoryFieldValidation(res, req, "anything")

		So(actual, ShouldBeFalse)
		So(httpStatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Passes validation when mandatory field values check succeeds", t, func() {
		var req, res = prepareForTest()
		actual, httpStatusCode := helperService.HandleMandatoryFieldValidation(res, req, "")

		So(actual, ShouldBeTrue)
		So(httpStatusCode, ShouldEqual, http.StatusOK)
	})

	Convey("Fails validation when missing mandatory fields check fails", t, func() {
		var req, res = prepareForTest()
		actual, httpStatusCode := helperService.HandleMandatoryFieldValidation(res, req, "anything")

		So(actual, ShouldBeFalse)
		So(httpStatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Fails validation when Statement Details check fails", t, func() {
		var req, res = prepareForTest()
		actual, httpStatusCode := helperService.HandleStatementDetailsValidation(res, req, "", "", errors.New(""))

		So(actual, ShouldBeFalse)
		So(httpStatusCode, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Fails validation when Statement Details check fails", t, func() {
		var req, res = prepareForTest()
		actual, httpStatusCode := helperService.HandleStatementDetailsValidation(res, req, "", "anything", nil)

		So(actual, ShouldBeFalse)
		So(httpStatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Passes validation when Statement Details check succeeds", t, func() {
		var req, res = prepareForTest()
		actual, httpStatusCode := helperService.HandleStatementDetailsValidation(res, req, "", "", nil)

		So(actual, ShouldBeTrue)
		So(httpStatusCode, ShouldEqual, http.StatusOK)
	})

	Convey("Fails validation when Attachment Resource check fails", t, func() {
		var req, res = prepareForTest()
		actual, httpStatusCode := helperService.HandleAttachmentResourceValidation(res, req, "", models.AttachmentResourceDao{}, nil)

		So(actual, ShouldBeFalse)
		So(httpStatusCode, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Passes validation when Attachment Resource check succeeds", t, func() {
		var req, res = prepareForTest()
		attachmentDao := models.AttachmentResourceDao{
			ID: "anything",
		}

		actual, httpStatusCode := helperService.HandleAttachmentResourceValidation(res, req, "", attachmentDao, nil)

		So(actual, ShouldBeTrue)
		So(httpStatusCode, ShouldEqual, http.StatusOK)
	})

	Convey("Records Attachment Type failure successfully", t, func() {
		var req, res = prepareForTest()
		httpStatusCode := helperService.HandleAttachmentTypeValidation(res, req, "", errors.New("anything"))

		So(httpStatusCode, ShouldEqual, http.StatusBadRequest)
	})

	//1*	HandleAttachmentTypeValidation(w http.ResponseWriter, req *http.Request, responseMessage string, err error) int

	Convey("Fails validation when etag generation fails", t, func() {
		actual := helperService.HandleEtagGenerationValidation(errors.New("anything"))

		So(actual, ShouldBeFalse)
	})

	Convey("Passes validation when etag generation succeeds", t, func() {
		actual := helperService.HandleEtagGenerationValidation(nil)

		So(actual, ShouldBeTrue)
	})

	Convey("Passes generating etags", t, func() {
		actual, _ := helperService.GenerateEtag()

		So(actual, ShouldNotBeNil)
	})

	Convey("Fails validation when create resource fails", t, func() {
		var req, res = prepareForTest()
		actual, _ := helperService.HandleCreateResourceValidation(res, req, http.StatusInternalServerError, errors.New("anything"))

		So(actual, ShouldBeFalse)
	})

	Convey("Passes validation when create resource succeeds", t, func() {
		var req, res = prepareForTest()
		actual, _ := helperService.HandleCreateResourceValidation(res, req, http.StatusInternalServerError, nil)

		So(actual, ShouldBeTrue)
	})

	Convey("Fails validation when missing mandatory field value check fails", t, func() {
		var req, res = prepareForTest()
		actual, httpStatusCode := helperService.HandleMandatoryFieldValidation(res, req, "anything")

		So(actual, ShouldBeFalse)
		So(httpStatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Passes validation when mandatory field value check succeeds", t, func() {
		var req, res = prepareForTest()
		actual, httpStatusCode := helperService.HandleMandatoryFieldValidation(res, req, "")

		So(actual, ShouldBeTrue)
		So(httpStatusCode, ShouldEqual, http.StatusOK)
	})
}
