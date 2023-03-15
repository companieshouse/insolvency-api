package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/dao"
	"github.com/companieshouse/insolvency-api/mocks"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
	"github.com/golang/mock/gomock"

	"github.com/gorilla/mux"
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

func serveHandleCreateProgressReport(body []byte, service dao.Service, helperService utils.HelperService, tranIDSet bool, res *httptest.ResponseRecorder) *httptest.ResponseRecorder {
	path := "/transactions/123456789/insolvency/progress-report"
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	if tranIDSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}

	handler := HandleCreateProgressReport(service, helperService)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleCreateProgressReport(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	helperService := utils.NewHelperService()

	Convey("Must need a transaction ID in the url", t, func() {
		mockService, _, rec := mocks.CreateTestObjects(t)

		body, _ := json.Marshal(&models.InsolvencyRequest{})

		res := serveHandleCreateProgressReport(body, mockService, helperService, false, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "transaction ID is not in the URL path")
	})

	Convey("Error checking if transaction is closed against transaction api", t, func() {
		mockService, _, rec := mocks.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an error
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

		body, _ := json.Marshal(&models.InsolvencyRequest{})

		res := serveHandleCreateProgressReport(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
		So(res.Body.String(), ShouldContainSubstring, "error checking transaction status")
	})

	Convey("Transaction is already closed and cannot be updated", t, func() {
		mockService, _, rec := mocks.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an already closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		body, _ := json.Marshal(&models.InsolvencyRequest{})

		res := serveHandleCreateProgressReport(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusForbidden)
		So(res.Body.String(), ShouldContainSubstring, "already closed and cannot be updated")
	})

	Convey("Failed to read request body", t, func() {
		mockService, _, rec := mocks.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		body := []byte(`{"first_name":error`)

		res := serveHandleCreateProgressReport(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, fmt.Sprintf("failed to read request body for transaction %s", transactionID))
	})

	Convey("Incoming request has from date missing", t, func() {
		mockService, _, rec := mocks.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		progressReport := generateProgressReport()
		progressReport.FromDate = ""

		body, _ := json.Marshal(progressReport)
		res := serveHandleCreateProgressReport(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "from_date is a required field")
	})

	Convey("Incoming request has to date missing", t, func() {
		mockService, _, rec := mocks.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		progressReport := generateProgressReport()
		progressReport.ToDate = ""

		body, _ := json.Marshal(progressReport)

		res := serveHandleCreateProgressReport(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "to_date is a required field")
	})

	Convey("Incoming request has invalid from date format", t, func() {
		mockService, _, rec := mocks.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		progressReport := generateProgressReport()
		progressReport.FromDate = "21-01-01"

		body, _ := json.Marshal(progressReport)

		res := serveHandleCreateProgressReport(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "from_date does not match the 2006-01-02 format")
	})

	Convey("Incoming request has invalid to date format", t, func() {
		mockService, _, rec := mocks.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		progressReport := generateProgressReport()
		progressReport.ToDate = "21-01-01"

		body, _ := json.Marshal(progressReport)

		res := serveHandleCreateProgressReport(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "to_date does not match the 2006-01-02 format")
	})

	Convey("invalid from date - in the future", t, func() {
		mockService, mockHelperService, rec := mocks.CreateTestObjects(t)

		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		progressReport := generateProgressReport()
		progressReport.FromDate = time.Now().AddDate(0, 0, 1).Format("2006-01-02")

		insolvencyDao := generateInsolvencyPractitionerAppointmentResources()

		body, _ := json.Marshal(progressReport)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()
		mockHelperService.EXPECT().HandleEtagGenerationValidation(gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyDao, []models.PractitionerResourceDao{}, nil)

		res := serveHandleCreateProgressReport(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "from_date")
		So(res.Body.String(), ShouldContainSubstring, "should not be in the future or before the company was incorporated")
	})

	Convey("invalid to date - in the future", t, func() {
		mockService, mockHelperService, rec := mocks.CreateTestObjects(t)

		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		progressReport := generateProgressReport()
		progressReport.ToDate = time.Now().AddDate(0, 0, 1).Format("2006-01-02")

		insolvencyDao := generateInsolvencyPractitionerAppointmentResources()

		body, _ := json.Marshal(progressReport)

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()
		mockHelperService.EXPECT().HandleEtagGenerationValidation(gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyDao, []models.PractitionerResourceDao{}, nil)

		res := serveHandleCreateProgressReport(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "to_date")
		So(res.Body.String(), ShouldContainSubstring, "should not be in the future or before the company was incorporated")
	})

	Convey("invalid from date - before incorporation date", t, func() {
		mockService, mockHelperService, rec := mocks.CreateTestObjects(t)

		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		progressReport := generateProgressReport()
		progressReport.FromDate = "1999-01-02"

		insolvencyDao := generateInsolvencyPractitionerAppointmentResources()

		body, _ := json.Marshal(progressReport)

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()
		mockHelperService.EXPECT().HandleEtagGenerationValidation(gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyDao, []models.PractitionerResourceDao{}, nil)

		res := serveHandleCreateProgressReport(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "from_date")
		So(res.Body.String(), ShouldContainSubstring, "should not be in the future or before the company was incorporated")
	})

	Convey("invalid to date - before incorporation date", t, func() {
		mockService, mockHelperService, rec := mocks.CreateTestObjects(t)

		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		progressReport := generateProgressReport()
		progressReport.ToDate = "1999-06-26"

		insolvencyDao := generateInsolvencyPractitionerAppointmentResources()

		body, _ := json.Marshal(progressReport)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()
		mockHelperService.EXPECT().HandleEtagGenerationValidation(gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyDao, []models.PractitionerResourceDao{}, nil)

		res := serveHandleCreateProgressReport(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "to_date")
		So(res.Body.String(), ShouldContainSubstring, "should not be in the future or before the company was incorporated")
	})

	Convey("invalid to date - to date is before from date", t, func() {
		mockService, mockHelperService, rec := mocks.CreateTestObjects(t)

		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		progressReport := generateProgressReport()
		progressReport.FromDate = "2021-06-27"
		progressReport.ToDate = "2021-06-26"

		insolvencyDao := generateInsolvencyPractitionerAppointmentResources()

		body, _ := json.Marshal(progressReport)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()
		mockHelperService.EXPECT().HandleEtagGenerationValidation(gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyDao, []models.PractitionerResourceDao{}, nil)

		res := serveHandleCreateProgressReport(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "to_date")
		So(res.Body.String(), ShouldContainSubstring, "should not be before from_date")
	})

	Convey("Incoming request has attachments missing", t, func() {
		mockService, _, rec := mocks.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		statement := generateProgressReport()
		statement.Attachments = nil
		body, _ := json.Marshal(statement)

		res := serveHandleCreateProgressReport(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "attachments is a required field")
	})

	Convey("Attachment is not associated with transaction", t, func() {
		mockService, _, rec := mocks.CreateTestObjects(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		progressReport := generateProgressReport()
		body, _ := json.Marshal(progressReport)

		insolvencyDao := generateInsolvencyPractitionerAppointmentResources()

		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyDao, []models.PractitionerResourceDao{}, nil)
		// Expect GetAttachmentFromInsolvencyResource to be called once and return an empty attachment model, nil
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, progressReport.Attachments[0]).Return(models.AttachmentResourceDao{}, nil)

		res := serveHandleCreateProgressReport(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
		So(res.Body.String(), ShouldContainSubstring, "attachment not found on transaction")
	})

	Convey("Failed to validate progress report", t, func() {
		mockService, mockHelperService, rec := mocks.CreateTestObjects(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		progressReport := generateProgressReport()
		progressReport.Attachments = nil

		insolvencyDao := generateInsolvencyPractitionerAppointmentResources()
		insolvencyDao.Data.Attachments = []models.AttachmentResourceDao{{
			Type: "resolution",
		}}

		body, _ := json.Marshal(progressReport)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()
		mockHelperService.EXPECT().HandleEtagGenerationValidation(gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleAttachmentValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleAttachmentTypeValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(http.StatusInternalServerError).AnyTimes()

		mockService.EXPECT().GetAttachmentFromInsolvencyResource(gomock.Any(), gomock.Any()).Return(models.AttachmentResourceDao{}, nil)
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyDao, []models.PractitionerResourceDao{}, fmt.Errorf("error"))

		res := serveHandleCreateProgressReport(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
		So(res.Body.String(), ShouldContainSubstring, "there was a problem handling your request for transaction ID")
	})

	Convey("Validation errors are present - multiple attachments", t, func() {
		mockService, mockHelperService, rec := mocks.CreateTestObjects(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		progressReport := generateProgressReport()
		progressReport.Attachments = []string{
			"1234567890",
			"0987654321",
		}

		insolvencyDao := generateInsolvencyPractitionerAppointmentResources()
		insolvencyDao.Data.Attachments = nil

		body, _ := json.Marshal(progressReport)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()
		mockHelperService.EXPECT().HandleEtagGenerationValidation(gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleAttachmentValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleAttachmentTypeValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(http.StatusBadRequest).AnyTimes()

		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, progressReport.Attachments[0]).Return(models.AttachmentResourceDao{}, nil)
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyDao, []models.PractitionerResourceDao{}, nil)

		res := serveHandleCreateProgressReport(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "please supply only one attachment")
	})

	Convey("Validation errors are present - no attachment is present", t, func() {
		mockService, mockHelperService, rec := mocks.CreateTestObjects(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		progressReport := generateProgressReport()
		progressReport.Attachments = []string{}

		insolvencyDao := generateInsolvencyPractitionerAppointmentResources()

		body, _ := json.Marshal(progressReport)
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()
		mockHelperService.EXPECT().HandleEtagGenerationValidation(gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyDao, []models.PractitionerResourceDao{}, nil)

		res := serveHandleCreateProgressReport(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "please supply only one attachment")
	})

	Convey("Attachment is not of type progress-report", t, func() {
		mockService, _, rec := mocks.CreateTestObjects(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		progressReport := generateProgressReport()

		insolvencyDao := generateInsolvencyPractitionerAppointmentResources()

		attachment := generateAttachment()
		attachment.Type = "not-progress-report"

		body, _ := json.Marshal(progressReport)
		// Expect GetAttachmentFromInsolvencyResource to be called once and return attachment, nil
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, progressReport.Attachments[0]).Return(attachment, nil)
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyDao, []models.PractitionerResourceDao{}, nil)

		res := serveHandleCreateProgressReport(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "attachment is not a progress-report")
	})

	Convey("Generic error when adding progress report resource to mongo", t, func() {
		mockService, _, rec := mocks.CreateTestObjects(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		progressReport := generateProgressReport()

		insolvencyDao := generateInsolvencyPractitionerAppointmentResources()

		attachment := generateAttachment()
		attachment.Type = "progress-report"

		body, _ := json.Marshal(progressReport)
		// Expect GetAttachmentFromInsolvencyResource to be called once and return attachment, nil
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, progressReport.Attachments[0]).Return(attachment, nil)
		// Expect CreateProgressReportResource to be called and return an error
		mockService.EXPECT().CreateProgressReportResource(gomock.Any(), transactionID).Return(http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID))
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyDao, []models.PractitionerResourceDao{}, nil)

		res := serveHandleCreateProgressReport(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
		So(res.Body.String(), ShouldContainSubstring, "there was a problem handling your request")
	})

	Convey("Error adding progress report resource to mongo - insolvency case not found", t, func() {
		mockService, _, rec := mocks.CreateTestObjects(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		progressReport := generateProgressReport()

		attachment := generateAttachment()
		attachment.Type = "progress-report"

		insolvencyDao := generateInsolvencyPractitionerAppointmentResources()

		body, _ := json.Marshal(progressReport)
		// Expect GetAttachmentFromInsolvencyResource to be called once and return attachment, nil
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, progressReport.Attachments[0]).Return(attachment, nil)
		// Expect CreateProgressReportResource to be called and return an error
		mockService.EXPECT().CreateProgressReportResource(gomock.Any(), transactionID).Return(http.StatusNotFound, fmt.Errorf("there was a problem handling your request for transaction %s not found", transactionID))
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyDao, []models.PractitionerResourceDao{}, nil)

		res := serveHandleCreateProgressReport(body, mockService, helperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusNotFound)
		So(res.Body.String(), ShouldContainSubstring, "not found")
	})

	Convey("Successfully add insolvency resource to mongo", t, func() {
		mockService, mockHelperService, rec := mocks.CreateTestObjects(t)
		httpmock.Activate()

		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		progressReport := generateProgressReport()

		attachment := generateAttachment()
		attachment.Type = "progress-report"

		insolvencyDao := generateInsolvencyPractitionerAppointmentResources()

		body, _ := json.Marshal(progressReport)

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()
		mockHelperService.EXPECT().HandleEtagGenerationValidation(gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockHelperService.EXPECT().HandleCreateResourceValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()

		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, progressReport.Attachments[0]).Return(attachment, nil)
		mockHelperService.EXPECT().HandleAttachmentValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockService.EXPECT().CreateProgressReportResource(gomock.Any(), transactionID).Return(http.StatusOK, nil)
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyDao, []models.PractitionerResourceDao{}, nil)

		res := serveHandleCreateProgressReport(body, mockService, mockHelperService, true, rec)

		So(res.Code, ShouldEqual, http.StatusCreated)
		So(res.Body.String(), ShouldContainSubstring, "\"kind\":\"insolvency-resource#progress-report\"")
	})
}

func serveHandleGetProgressReport(service dao.Service, tranIDSet bool) *httptest.ResponseRecorder {
	path := "/transactions/123456789/insolvency/progress-report"
	req := httptest.NewRequest(http.MethodPost, path, nil)
	if tranIDSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}
	rec := httptest.NewRecorder()

	handler := HandleGetProgressReport(service)
	handler.ServeHTTP(rec, req)

	return rec
}

func TestUnitHandleGetProgressReport(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	Convey("Must need a transaction ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveHandleGetProgressReport(mocks.NewMockService(mockCtrl), false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Failed to get progress report from Insolvency resource", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect GetProgressReportResource to be called once and return an error
		mockService.EXPECT().GetProgressReportResource(transactionID).Return(&models.ProgressReportResourceDao{}, fmt.Errorf("failed to get progress report from insolvency resource in db for transaction [%s]: %v", transactionID, err))

		res := serveHandleGetProgressReport(mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Progress report was not found on supplied transaction", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect GetProgressReportResource to be called once and return nil
		mockService.EXPECT().GetProgressReportResource(transactionID).Return(&models.ProgressReportResourceDao{}, nil)

		res := serveHandleGetProgressReport(mockService, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Success - Progress report was retrieved from insolvency resource", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		progressReport := models.ProgressReportResourceDao{
			Etag:     "6f143c1f8109d834263eb764c5f020a0ae3ff78ee1789477179cb80f",
			Kind:     "insolvency-resource#progress-report",
			FromDate: "2021-06-06",
			ToDate:   "2022-06-05",
			Attachments: []string{
				"1223-3445-5667",
			},
			Links: models.ProgressReportResourceLinksDao{
				Self: "/transactions/12345678/insolvency/progress-report",
			},
		}

		// Expect GetProgressReportResource to be called once and return statement of affairs
		mockService.EXPECT().GetProgressReportResource(transactionID).Return(&progressReport, nil)

		res := serveHandleGetProgressReport(mockService, true)

		So(res.Code, ShouldEqual, http.StatusOK)
		So(res.Body.String(), ShouldContainSubstring, "etag")
		So(res.Body.String(), ShouldContainSubstring, "kind")
		So(res.Body.String(), ShouldContainSubstring, "links")
		So(res.Body.String(), ShouldContainSubstring, "from_date")
		So(res.Body.String(), ShouldContainSubstring, "to_date")
		So(res.Body.String(), ShouldContainSubstring, "attachments")
	})
}

func serveHandleDeleteProgressReport(service dao.Service, helperService utils.HelperService, tranIDSet bool) *httptest.ResponseRecorder {
	path := "/transactions/123456789/insolvency/progress-report"
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	if tranIDSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}
	res := httptest.NewRecorder()

	handler := HandleDeleteProgressReport(service, helperService)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleDeleteProgressReport(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	helperService := utils.NewHelperService()

	Convey("Must need a transaction ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveHandleDeleteProgressReport(mocks.NewMockService(mockCtrl), helperService, false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error checking if transaction is closed against transaction api", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		// Expect the transaction api to be called and return an error
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

		res := serveHandleDeleteProgressReport(mocks.NewMockService(mockCtrl), helperService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Transaction is already closed and cannot be updated", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		// Expect the transaction api to be called and return an already closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		res := serveHandleDeleteProgressReport(mocks.NewMockService(mockCtrl), helperService, true)

		So(res.Code, ShouldEqual, http.StatusForbidden)
	})

	Convey("Error when deleting progress report from DB", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		// Expect the deletion of progress report to return an error
		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().DeleteProgressReportResource(transactionID).Return(http.StatusInternalServerError, fmt.Errorf("err"))

		res := serveHandleDeleteProgressReport(mockService, helperService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Progress report not found when deleting from DB", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().DeleteProgressReportResource(transactionID).Return(http.StatusNotFound, fmt.Errorf("err"))

		res := serveHandleDeleteProgressReport(mockService, helperService, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Successfully delete progress report from DB", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().DeleteProgressReportResource(transactionID).Return(http.StatusNoContent, nil)

		res := serveHandleDeleteProgressReport(mockService, helperService, true)

		So(res.Code, ShouldEqual, http.StatusNoContent)
	})

}

func generateProgressReport() models.ProgressReport {
	return models.ProgressReport{
		FromDate: "2021-06-06",
		ToDate:   "2021-06-07",
		Attachments: []string{
			"123456789",
		},
	}
}
