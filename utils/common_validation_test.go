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
		transactionId, valid, httpStatusCode := helperService.HandleTransactionIdExistsValidation(res, req, "")

		So(transactionId, ShouldEqual, "")
		So(valid, ShouldBeFalse)
		So(httpStatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Passes validation when transaction ID exists", t, func() {
		var req, res = prepareForTest()
		transactionId, valid, httpStatusCode := helperService.HandleTransactionIdExistsValidation(res, req, "1234567")

		So(transactionId, ShouldEqual, "1234567")
		So(valid, ShouldBeTrue)
		So(httpStatusCode, ShouldEqual, http.StatusOK)
	})

	Convey("Fails validation when transaction is closed against transaction api", t, func() {
		var req, res = prepareForTest()
		valid, httpStatusCode, err := helperService.HandleTransactionNotClosedValidation(res, req, "anything", false, http.StatusInternalServerError, errors.New("anything"))

		So(valid, ShouldBeFalse)
		So(httpStatusCode, ShouldEqual, http.StatusInternalServerError)
		So(err, ShouldBeNil)
	})

	Convey("Fails validation when transaction is already closed and cannot be updated", t, func() {
		var req, res = prepareForTest()
		valid, httpStatusCode, err := helperService.HandleTransactionNotClosedValidation(res, req, "anything", true, http.StatusInternalServerError, nil)

		So(valid, ShouldBeFalse)
		So(httpStatusCode, ShouldEqual, http.StatusInternalServerError)
		So(err, ShouldBeNil)
	})

	Convey("Passes validation when transaction is not closed", t, func() {
		var req, res = prepareForTest()
		valid, httpStatusCode, err := helperService.HandleTransactionNotClosedValidation(res, req, "anything", false, http.StatusOK, nil)

		So(valid, ShouldBeTrue)
		So(httpStatusCode, ShouldEqual, http.StatusOK)
		So(err, ShouldBeNil)
	})

	Convey("Fails validation when invalid request body", t, func() {
		var req, res = prepareForTest()
		valid, httpStatusCode := helperService.HandleBodyDecodedValidation(res, req, "anything", errors.New("anything"))

		So(valid, ShouldBeFalse)
		So(httpStatusCode, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Passes validation when valid request body", t, func() {
		var req, res = prepareForTest()
		valid, httpStatusCode := helperService.HandleBodyDecodedValidation(res, req, "anything", nil)

		So(valid, ShouldBeTrue)
		So(httpStatusCode, ShouldEqual, http.StatusOK)
	})

	Convey("Fails validation when missing mandatory field values check fails", t, func() {
		var req, res = prepareForTest()
		valid, httpStatusCode := helperService.HandleMandatoryFieldValidation(res, req, "anything")

		So(valid, ShouldBeFalse)
		So(httpStatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Fails validation when Attachment Resource check fails", t, func() {
		var req, res = prepareForTest()
		valid, httpStatusCode := helperService.HandleAttachmentValidation(res, req, "", models.AttachmentResourceDao{}, nil)

		So(valid, ShouldBeFalse)
		So(httpStatusCode, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Passes validation when Attachment Resource check succeeds", t, func() {
		var req, res = prepareForTest()
		attachmentDao := models.AttachmentResourceDao{
			ID: "anything",
		}

		valid, httpStatusCode := helperService.HandleAttachmentValidation(res, req, "", attachmentDao, nil)

		So(valid, ShouldBeTrue)
		So(httpStatusCode, ShouldEqual, http.StatusOK)
	})

	Convey("Records Attachment Type failure successfully", t, func() {
		var req, res = prepareForTest()
		httpStatusCode := helperService.HandleAttachmentTypeValidation(res, req, "", errors.New("anything"))

		So(httpStatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Fails validation when etag generation fails", t, func() {
		actual := helperService.HandleEtagGenerationValidation(errors.New("anything"))

		So(actual, ShouldBeFalse)
	})

	Convey("Passes validation when etag generation succeeds", t, func() {
		actual := helperService.HandleEtagGenerationValidation(nil)

		So(actual, ShouldBeTrue)
	})

	Convey("Passes generating etags", t, func() {
		sha1Hash, err := helperService.GenerateEtag()

		So(sha1Hash, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})

	Convey("Fails validation when create resource fails", t, func() {
		var req, res = prepareForTest()
		valid, _ := helperService.HandleCreateResourceValidation(res, req, http.StatusInternalServerError, errors.New("anything"))

		So(valid, ShouldBeFalse)
	})

	Convey("Passes validation when create resource succeeds", t, func() {
		var req, res = prepareForTest()
		valid, _ := helperService.HandleCreateResourceValidation(res, req, http.StatusInternalServerError, nil)

		So(valid, ShouldBeTrue)
	})

	Convey("Fails validation when missing mandatory field value check fails", t, func() {
		var req, res = prepareForTest()
		valid, httpStatusCode := helperService.HandleMandatoryFieldValidation(res, req, "anything")

		So(valid, ShouldBeFalse)
		So(httpStatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Passes validation when mandatory field value check succeeds", t, func() {
		var req, res = prepareForTest()
		valid, httpStatusCode := helperService.HandleMandatoryFieldValidation(res, req, "")

		So(valid, ShouldBeTrue)
		So(httpStatusCode, ShouldEqual, http.StatusOK)
	})
}
