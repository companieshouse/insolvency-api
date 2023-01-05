package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/dao"
	mock_dao "github.com/companieshouse/insolvency-api/mocks"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

func serveHandleCreateResolution(body []byte, service dao.Service, helperService utils.HelperService, tranIDSet bool) *httptest.ResponseRecorder {
	path := "/transactions/123456789/insolvency/resolution"
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	if tranIDSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}
	res := httptest.NewRecorder()

	handler := HandleCreateResolution(service, helperService)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleCreateResolution(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	Convey("Must need a transaction ID in the url", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		body, _ := json.Marshal(&models.InsolvencyRequest{})

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), "").Return("", false, http.StatusBadRequest).AnyTimes()

		res := serveHandleCreateResolution(body, mockService, mockHelperService, false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error checking if transaction is closed against transaction api", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an error
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

		body, _ := json.Marshal(&models.InsolvencyRequest{})

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusInternalServerError).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, http.StatusInternalServerError, nil).AnyTimes()

		res := serveHandleCreateResolution(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Transaction is already closed and cannot be updated", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an already closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		body, _ := json.Marshal(&models.InsolvencyRequest{})

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusInternalServerError).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, http.StatusForbidden, nil).AnyTimes()

		res := serveHandleCreateResolution(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusForbidden)
	})

	Convey("Failed to read request body", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		body := []byte(`{"first_name":error`)

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusInternalServerError).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, http.StatusInternalServerError).AnyTimes()

		res := serveHandleCreateResolution(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Incoming request has date of resolution missing", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		resolution := generateResolution()
		resolution.DateOfResolution = ""
		body, _ := json.Marshal(resolution)

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, http.StatusBadRequest).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil)

		res := serveHandleCreateResolution(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "date_of_resolution is a required field")
	})

	Convey("Incoming request has invalid date format", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		resolution := generateResolution()
		resolution.DateOfResolution = "21-01-01"
		body, _ := json.Marshal(resolution)

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, http.StatusBadRequest).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil)

		res := serveHandleCreateResolution(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "date_of_resolution does not match the 2006-01-02 format")
	})

	Convey("Incoming request has attachments missing", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		resolution := generateResolution()
		resolution.Attachments = nil
		body, _ := json.Marshal(resolution)

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, http.StatusBadRequest).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil)

		res := serveHandleCreateResolution(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "attachments is a required field")
	})

	Convey("Attachment is not associated with transaction", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		resolution := generateResolution()

		body, _ := json.Marshal(resolution)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)
		// Expect GetAttachmentFromInsolvencyResource to be called once and return an empty attachment model, nil
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, resolution.Attachments[0]).Return(models.AttachmentResourceDao{}, nil)
		mockHelperService.EXPECT().HandleAttachmentValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, http.StatusInternalServerError).AnyTimes()

		res := serveHandleCreateResolution(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
		So(res.Body.String(), ShouldContainSubstring, "attachment not found on transaction")
	})

	Convey("Failed to validate resolution", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		resolution := generateResolution()

		body, _ := json.Marshal(resolution)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(models.InsolvencyResourceDao{}, fmt.Errorf("error"))

		res := serveHandleCreateResolution(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
		So(res.Body.String(), ShouldContainSubstring, "there was a problem handling your request for transaction ID")
	})

	Convey("Validation errors are present", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		resolution := generateResolution()
		resolution.DateOfResolution = "1999-01-01"

		body, _ := json.Marshal(resolution)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)

		res := serveHandleCreateResolution(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, fmt.Sprintf("date_of_resolution [%s] should not be in the future or before the company was incorporated", resolution.DateOfResolution))
	})

	Convey("Validation errors are present", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		resolution := generateResolution()
		resolution.Attachments = []string{
			"1234567890",
			"0987654321",
		}

		body, _ := json.Marshal(resolution)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()

		res := serveHandleCreateResolution(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "please supply only one attachment")
	})

	Convey("Validation errors are present", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		resolution := generateResolution()
		resolution.Attachments = []string{}

		body, _ := json.Marshal(resolution)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()

		res := serveHandleCreateResolution(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "please supply only one attachment")
	})

	Convey("Attachment is not of type resolution", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		resolution := generateResolution()

		attachment := generateAttachment()
		attachment.Type = "not-resolution"

		body, _ := json.Marshal(resolution)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)
		// Expect GetAttachmentFromInsolvencyResource to be called once and return attachment, nil
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, resolution.Attachments[0]).Return(attachment, nil)
		mockHelperService.EXPECT().HandleAttachmentValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleAttachmentTypeValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(http.StatusBadRequest).AnyTimes()

		res := serveHandleCreateResolution(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "attachment is not a resolution")
	})

	Convey("Generic error when adding resolution resource to mongo", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		resolution := generateResolution()
		body, _ := json.Marshal(resolution)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)
		// Expect GetAttachmentFromInsolvencyResource to be called once and return attachment, nil
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, resolution.Attachments[0]).Return(generateAttachment(), nil)
		mockHelperService.EXPECT().HandleAttachmentValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleAttachmentTypeValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(http.StatusInternalServerError).AnyTimes()
		// Expect CreateResolutionResource to be called once and return an error
		mockService.EXPECT().CreateResolutionResource(gomock.Any(), transactionID).Return(http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)).Times(1)
		mockHelperService.EXPECT().HandleCreateResourceValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, http.StatusInternalServerError).AnyTimes()

		res := serveHandleCreateResolution(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Error adding resolution resource to mongo - insolvency case not found", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		resolution := generateResolution()
		body, _ := json.Marshal(resolution)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)
		// Expect GetAttachmentFromInsolvencyResource to be called once and return attachment, nil
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, resolution.Attachments[0]).Return(generateAttachment(), nil)
		mockHelperService.EXPECT().HandleAttachmentValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleAttachmentTypeValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(http.StatusOK).AnyTimes()
		// Expect CreateResolutionResource to be called once and return an error
		mockService.EXPECT().CreateResolutionResource(gomock.Any(), transactionID).Return(http.StatusNotFound, fmt.Errorf("there was a problem handling your request for transaction %s not found", transactionID)).Times(1)
		mockHelperService.EXPECT().HandleCreateResourceValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, http.StatusNotFound).AnyTimes()

		res := serveHandleCreateResolution(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Successfully add insolvency resource to mongo", t, func() {
		mockService, mockHelperService := utils.CreateTestServices(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		resolution := generateResolution()

		attachment := models.AttachmentResourceDao{
			ID:     "1111",
			Type:   "resolution",
			Status: "status",
			Links:  models.AttachmentResourceLinksDao{},
		}

		body, _ := json.Marshal(resolution)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK, nil).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)
		// Expect GetAttachmentFromInsolvencyResource to be called once and return attachment, nil
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, resolution.Attachments[0]).Return(attachment, nil)
		mockHelperService.EXPECT().HandleAttachmentValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()
		mockHelperService.EXPECT().HandleAttachmentTypeValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(http.StatusOK).AnyTimes()
		// Expect CreateResolutionResource to be called once and return an error
		mockService.EXPECT().CreateResolutionResource(gomock.Any(), transactionID).Return(http.StatusCreated, nil).Times(1)
		mockHelperService.EXPECT().HandleCreateResourceValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()

		res := serveHandleCreateResolution(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusOK)
		So(res.Body.String(), ShouldContainSubstring, "\"date_of_resolution\":\"2021-06-06\"")
	})
}

func serveHandleGetResolution(service dao.Service, tranIDSet bool) *httptest.ResponseRecorder {
	path := "/transactions/123456789/insolvency/resolution"
	req := httptest.NewRequest(http.MethodPost, path, nil)
	if tranIDSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}
	res := httptest.NewRecorder()

	handler := HandleGetResolution(service)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleGetResolution(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockService := mock_dao.NewMockService(mockCtrl)

	Convey("Must need a transaction ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveHandleGetResolution(mockService, false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Failed to get resolution from Insolvency resource", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// Expect GetResolutionResource to be called once and return an error
		mockService.EXPECT().GetResolutionResource(transactionID).Return(models.ResolutionResourceDao{}, fmt.Errorf("failed to get resolution from insolvency resource in db for transaction [%s]: %v", transactionID, err))

		res := serveHandleGetResolution(mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Resolution was not found on supplied transaction", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// Expect GetResolutionResource to be called once and return nil
		mockService.EXPECT().GetResolutionResource(transactionID).Return(models.ResolutionResourceDao{}, nil)

		res := serveHandleGetResolution(mockService, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Success - Resolution was retrieved from insolvency resource", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		resolution := models.ResolutionResourceDao{
			DateOfResolution: "2021-06-06",
			Attachments: []string{
				"1223-3445-5667",
			},
		}
		// Expect GetResolutionResource to be called once and return a resolution
		mockService.EXPECT().GetResolutionResource(transactionID).Return(resolution, nil)

		res := serveHandleGetResolution(mockService, true)

		So(res.Code, ShouldEqual, http.StatusOK)
	})
}

func serveHandleDeleteResolution(service dao.Service, tranIDSet bool) *httptest.ResponseRecorder {
	path := "/transactions/123456789/insolvency/resolution"
	req := httptest.NewRequest(http.MethodPost, path, nil)
	if tranIDSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}
	res := httptest.NewRecorder()

	handler := HandleDeleteResolution(service)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleDeleteResolution(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockService := mock_dao.NewMockService(mockCtrl)

	Convey("Must need a transaction ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveHandleDeleteResolution(mockService, false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error checking if transaction is closed against transaction api", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an error
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

		res := serveHandleDeleteResolution(mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Transaction is already closed and cannot be updated", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an already closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		res := serveHandleDeleteResolution(mockService, true)

		So(res.Code, ShouldEqual, http.StatusForbidden)
	})

	Convey("Failed to delete resolution from Insolvency resource", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		// Expect DeleteResolutionResource to be called once and return an error
		mockService.EXPECT().DeleteResolutionResource(transactionID).Return(http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction id [%s] - could not delete resolution", transactionID))

		res := serveHandleDeleteResolution(mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Resolution was not found on supplied transaction", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		// Expect DeleteResolutionResource to be called once and return an error
		mockService.EXPECT().DeleteResolutionResource(transactionID).Return(http.StatusNotFound, fmt.Errorf("there was a problem handling your request for transaction id [%s] - resolution not found", transactionID))

		res := serveHandleDeleteResolution(mockService, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Success - Resolution was deleted from insolvency resource", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		// Expect DeleteResolutionResource to be called once and delete resolution
		mockService.EXPECT().DeleteResolutionResource(transactionID).Return(http.StatusNoContent, nil)

		res := serveHandleDeleteResolution(mockService, true)

		So(res.Code, ShouldEqual, http.StatusNoContent)
	})
}

func generateResolution() models.Resolution {
	return models.Resolution{
		DateOfResolution: "2021-06-06",
		Attachments: []string{
			"123456789",
		},
	}
}

func generateAttachment() models.AttachmentResourceDao {
	return models.AttachmentResourceDao{
		ID:     "1111",
		Type:   "resolution",
		Status: "status",
		Links:  models.AttachmentResourceLinksDao{},
	}
}

func generateInsolvencyResource() models.InsolvencyResourceDao {
	return models.InsolvencyResourceDao{
		Data: models.InsolvencyResourceDaoData{
			CompanyNumber: "1234",
			CaseType:      "CVL",
			CompanyName:   "Company",
		},
	}
}
