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

func serveHandleCreateStatementOfAffairs(body []byte, service dao.Service, helperService utils.HelperService, tranIDSet bool) *httptest.ResponseRecorder {
	path := "/transactions/123456789/insolvency/statement-of-affairs"
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	if tranIDSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}
	res := httptest.NewRecorder()

	handler := HandleCreateStatementOfAffairs(service, helperService)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleCreateStatementOfAffairs(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	Convey("Must need a transaction ID in the url", t, func() {
		mockService := mock_dao.NewMockService(mockCtrl)
		mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

		body, _ := json.Marshal(&models.InsolvencyRequest{})

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), "").Return("", false, http.StatusBadRequest).AnyTimes()

		res := serveHandleCreateStatementOfAffairs(body, mockService, mockHelperService, false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error checking if transaction is closed against transaction api", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		mockService := mock_dao.NewMockService(mockCtrl)
		mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

		// Expect the transaction api to be called and return an error
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

		body, _ := json.Marshal(&models.InsolvencyRequest{})

		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, false, http.StatusInternalServerError).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusInternalServerError).AnyTimes()

		res := serveHandleCreateStatementOfAffairs(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Transaction is already closed and cannot be updated", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		mockService := mock_dao.NewMockService(mockCtrl)
		mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

		// Expect the transaction api to be called and return an already closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		body, _ := json.Marshal(&models.InsolvencyRequest{})

		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, true, http.StatusForbidden).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusInternalServerError).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, http.StatusForbidden).AnyTimes()

		res := serveHandleCreateStatementOfAffairs(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusForbidden)
	})

	Convey("Failed to read request body", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		mockService := mock_dao.NewMockService(mockCtrl)
		mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		body := []byte(`{"first_name":error`)

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusInternalServerError).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, true, http.StatusInternalServerError).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, http.StatusInternalServerError).AnyTimes()

		res := serveHandleCreateStatementOfAffairs(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Incoming request has statement date missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()
		mockService := mock_dao.NewMockService(mockCtrl)
		mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		statement := generateStatement()
		statement.StatementDate = ""
		body, _ := json.Marshal(statement)

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusInternalServerError).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, true, http.StatusInternalServerError).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusInternalServerError).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, http.StatusBadRequest).AnyTimes()

		res := serveHandleCreateStatementOfAffairs(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "statement_date is a required field")
	})

	Convey("Incoming request has invalid date format", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()
		mockService := mock_dao.NewMockService(mockCtrl)
		mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		statement := generateStatement()
		statement.StatementDate = "21-01-01"
		body, _ := json.Marshal(statement)

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusInternalServerError).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, true, http.StatusInternalServerError).AnyTimes()
		mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusInternalServerError).AnyTimes()
		mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, http.StatusBadRequest).AnyTimes()

		res := serveHandleCreateStatementOfAffairs(body, mockService, mockHelperService, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "statement_date does not match the 2006-01-02 format")
	})

	// Convey("Incoming request has attachments missing", t, func() {
	// 	httpmock.Activate()
	// 	mockCtrl := gomock.NewController(t)
	// 	defer mockCtrl.Finish()
	// 	defer httpmock.DeactivateAndReset()

	// 	// Expect the transaction api to be called and return an open transaction
	// 	httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

	// 	statement := generateStatement()
	// 	statement.Attachments = nil
	// 	body, _ := json.Marshal(statement)
	// 	res := serveHandleCreateStatementOfAffairs(body, mock_dao.NewMockService(mockCtrl), mockHelperService, true)

	// 	So(res.Code, ShouldEqual, http.StatusBadRequest)
	// 	So(res.Body.String(), ShouldContainSubstring, "attachments is a required field")
	// })

	//Convey("Attachment is not associated with transaction", t, func() {
	//	httpmock.Activate()
	//	defer httpmock.DeactivateAndReset()
	//	mockCtrl := gomock.NewController(t)
	//	defer mockCtrl.Finish()
	//
	//	mockService := mock_dao.NewMockService(mockCtrl)
	//
	//	httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))
	//	insolvencyDao := generateInsolvencyResource()
	//	mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyDao, nil)
	//
	//	// Expect the transaction api to be called and return an open transaction
	//	httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))
	//
	//	statement := generateStatement()
	//
	//	// Expect GetAttachmentFromInsolvencyResource to be called once and return an empty attachment model, nil
	//	mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, statement.Attachments[0]).Return(models.AttachmentResourceDao{}, nil)
	//
	//	body, _ := json.Marshal(statement)
	//	res := serveHandleCreateStatementOfAffairs(body, mockService, mockHelperService, true)
	//
	//	So(res.Code, ShouldEqual, http.StatusInternalServerError)
	//	So(res.Body.String(), ShouldContainSubstring, "attachment not found on transaction")
	//})

	// Convey("Failed to validate statement of affairs", t, func() {
	// 	httpmock.Activate()
	// 	defer httpmock.DeactivateAndReset()
	// 	mockCtrl := gomock.NewController(t)
	// 	defer mockCtrl.Finish()

	// 	mockService := mock_dao.NewMockService(mockCtrl)

	// 	httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))
	// 	mockService.EXPECT().GetInsolvencyResource(transactionID).Return(models.InsolvencyResourceDao{}, fmt.Errorf("error"))

	// 	// Expect the transaction api to be called and return an open transaction
	// 	httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

	// 	statement := generateStatement()

	// 	body, _ := json.Marshal(statement)
	// 	res := serveHandleCreateStatementOfAffairs(body, mockService, true)

	// 	So(res.Code, ShouldEqual, http.StatusInternalServerError)
	// 	So(res.Body.String(), ShouldContainSubstring, "there was a problem handling your request for transaction ID")
	// })

	// Convey("Validation errors are present - date is in the past", t, func() {
	// 	httpmock.Activate()
	// 	defer httpmock.DeactivateAndReset()
	// 	mockCtrl := gomock.NewController(t)
	// 	defer mockCtrl.Finish()

	// 	mockService := mock_dao.NewMockService(mockCtrl)

	// 	httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))
	// 	insolvencyDao := generateInsolvencyResource()
	// 	mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyDao, nil)

	// 	// Expect the transaction api to be called and return an open transaction
	// 	httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

	// 	statement := generateStatement()
	// 	statement.StatementDate = "1999-01-01"

	// 	body, _ := json.Marshal(statement)
	// 	res := serveHandleCreateStatementOfAffairs(body, mockService, true)

	// 	So(res.Code, ShouldEqual, http.StatusBadRequest)
	// 	So(res.Body.String(), ShouldContainSubstring, fmt.Sprintf("statement_date [%s] should not be in the future or before the company was incorporated", statement.StatementDate))
	// })

	// Convey("Validation errors are present - multiple attachments", t, func() {
	// 	httpmock.Activate()
	// 	defer httpmock.DeactivateAndReset()
	// 	mockCtrl := gomock.NewController(t)
	// 	defer mockCtrl.Finish()

	// 	mockService := mock_dao.NewMockService(mockCtrl)

	// 	httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))
	// 	insolvencyDao := generateInsolvencyResource()
	// 	mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyDao, nil)

	// 	// Expect the transaction api to be called and return an open transaction
	// 	httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

	// 	statement := generateStatement()
	// 	statement.Attachments = []string{
	// 		"1234567890",
	// 		"0987654321",
	// 	}

	// 	body, _ := json.Marshal(statement)
	// 	res := serveHandleCreateStatementOfAffairs(body, mockService, true)

	// 	So(res.Code, ShouldEqual, http.StatusBadRequest)
	// 	So(res.Body.String(), ShouldContainSubstring, "please supply only one attachment")
	// })

	// Convey("Validation errors are present - no attachment is present", t, func() {
	// 	httpmock.Activate()
	// 	defer httpmock.DeactivateAndReset()
	// 	mockCtrl := gomock.NewController(t)
	// 	defer mockCtrl.Finish()

	// 	mockService := mock_dao.NewMockService(mockCtrl)

	// 	httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))
	// 	insolvencyDao := generateInsolvencyResource()
	// 	mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyDao, nil)

	// 	// Expect the transaction api to be called and return an open transaction
	// 	httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

	// 	statement := generateStatement()
	// 	statement.Attachments = []string{}

	// 	body, _ := json.Marshal(statement)
	// 	res := serveHandleCreateStatementOfAffairs(body, mockService, true)

	// 	So(res.Code, ShouldEqual, http.StatusBadRequest)
	// 	So(res.Body.String(), ShouldContainSubstring, "please supply only one attachment")
	// })

	// Convey("Attachment is not of type statement-of-affairs-director or liquidator", t, func() {
	// 	httpmock.Activate()
	// 	defer httpmock.DeactivateAndReset()
	// 	mockCtrl := gomock.NewController(t)
	// 	defer mockCtrl.Finish()

	// 	mockService := mock_dao.NewMockService(mockCtrl)

	// 	httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))
	// 	insolvencyDao := generateInsolvencyResource()
	// 	mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyDao, nil)

	// 	// Expect the transaction api to be called and return an open transaction
	// 	httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

	// 	statement := generateStatement()

	// 	body, _ := json.Marshal(statement)

	// 	attachment := generateAttachment()
	// 	attachment.Type = "not-soa"

	// 	// Expect GetAttachmentFromInsolvencyResource to be called once and return attachment, nil
	// 	mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, statement.Attachments[0]).Return(attachment, nil)

	// 	res := serveHandleCreateStatementOfAffairs(body, mockService, true)

	// 	So(res.Code, ShouldEqual, http.StatusBadRequest)
	// 	So(res.Body.String(), ShouldContainSubstring, "attachment is not a statement-of-affairs-director")
	// })

	// Convey("Generic error when adding statement of affairs resource to mongo", t, func() {
	// 	httpmock.Activate()§
	// 	mockCtrl := gomock.NewController(t)
	// 	defer mockCtrl.Finish()
	// 	defer httpmock.DeactivateAndReset()
	// 	mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

	// 	httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

	// 	// Expect the transaction api to be called and return an open transaction
	// 	httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

	// 	statement := generateStatement()
	// 	body, _ := json.Marshal(statement)

	// 	attachment := generateAttachment()
	// 	attachment.Type = "statement-of-affairs-director"

	// 	mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusInternalServerError).AnyTimes()
	// 	mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, true, http.StatusForbidden).AnyTimes()
	// 	mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusInternalServerError).AnyTimes()
	// 	mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusBadRequest).AnyTimes()
	// 	mockService.EXPECT().CreateStatementOfAffairsResource(gomock.Any(), transactionID).Return(http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID))
	// 	mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()
	// 	mockHelperService.EXPECT().HandleEtagGenerationValidation(gomock.Any()).Return(false).AnyTimes()

	// 	res := serveHandleCreateStatementOfAffairs(body, mockService, mockHelperService, true)

	// 	So(res.Code, ShouldEqual, http.StatusInternalServerError)
	// })

	// Convey("Error adding statement of affairs resource to mongo - insolvency case not found", t, func() {
	// 	httpmock.Activate()
	// 	mockCtrl := gomock.NewController(t)
	// 	defer mockCtrl.Finish()
	// 	defer httpmock.DeactivateAndReset()
	// 	mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

	// 	httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

	// 	// Expect the transaction api to be called and return an open transaction
	// 	httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

	// 	statement := generateStatement()
	// 	body, _ := json.Marshal(statement)

	// 	attachment := generateAttachment()
	// 	attachment.Type = "statement-of-affairs-director"

	// 	mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusInternalServerError).AnyTimes()
	// 	mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, true, http.StatusForbidden).AnyTimes()
	// 	mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusInternalServerError).AnyTimes()
	// 	mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusBadRequest).AnyTimes()
	// 	mockService.EXPECT().CreateStatementOfAffairsResource(gomock.Any(), transactionID).Return(http.StatusNotFound, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID))
	// 	mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()
	// 	mockHelperService.EXPECT().HandleEtagGenerationValidation(gomock.Any()).Return(false).AnyTimes()

	// 	res := serveHandleCreateStatementOfAffairs(body, mockService, mockHelperService, true)

	// 	So(res.Code, ShouldEqual, http.StatusNotFound)
	// })

	// Convey("Successfully add insolvency resource to mongo", t, func() {
	// 	httpmock.Activate()
	// 	defer httpmock.DeactivateAndReset()
	// 	mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

	// 	httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

	// 	// Expect the transaction api to be called and return an open transaction
	// 	httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

	// 	statement := generateStatement()
	// 	body, _ := json.Marshal(statement)

	// 	mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(transactionID, true, http.StatusInternalServerError).AnyTimes()
	// 	mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, true, http.StatusForbidden).AnyTimes()
	// 	mockHelperService.EXPECT().HandleBodyDecodedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusInternalServerError).AnyTimes()
	// 	mockHelperService.EXPECT().HandleMandatoryFieldValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusBadRequest).AnyTimes()
	// 	mockService.EXPECT().CreateStatementOfAffairsResource(gomock.Any(), transactionID).Return(http.StatusCreated, nil).Times(1)
	// 	mockHelperService.EXPECT().HandleCreateResourceValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, http.StatusOK).AnyTimes()

	// 	res := serveHandleCreateStatementOfAffairs(body, mockService, mockHelperService, true)

	// 	So(res.Code, ShouldEqual, http.StatusOK)
	// })
}

func serveHandleGetStatementOfAffairs(service dao.Service, tranIDSet bool) *httptest.ResponseRecorder {
	path := "/transactions/123456789/insolvency/statement-of-affairs"
	req := httptest.NewRequest(http.MethodPost, path, nil)
	if tranIDSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}
	res := httptest.NewRecorder()

	handler := HandleGetStatementOfAffairs(service)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleGetStatementOfAffairs(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	Convey("Must need a transaction ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveHandleGetStatementOfAffairs(mock_dao.NewMockService(mockCtrl), false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Failed to get statement of affairs from Insolvency resource", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()
		mockService := mock_dao.NewMockService(mockCtrl)

		// Expect GetStatementOfAffairsResource to be called once and return an error
		mockService.EXPECT().GetStatementOfAffairsResource(transactionID).Return(models.StatementOfAffairsResourceDao{}, fmt.Errorf("failed to get statement of affairs from insolvency resource in db for transaction [%s]: %v", transactionID, err))

		res := serveHandleGetStatementOfAffairs(mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Statment of affairs was not found on supplied transaction", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()
		mockService := mock_dao.NewMockService(mockCtrl)

		// Expect GetStatementOfAffairsResource to be called once and return nil
		mockService.EXPECT().GetStatementOfAffairsResource(transactionID).Return(models.StatementOfAffairsResourceDao{}, nil)

		res := serveHandleGetStatementOfAffairs(mockService, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Success - Statement of affairs was retrieved from insolvency resource", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()
		mockService := mock_dao.NewMockService(mockCtrl)

		statementOfAffairs := models.StatementOfAffairsResourceDao{
			StatementDate: "2021-06-06",
			Attachments: []string{
				"1223-3445-5667",
			},
		}
		// Expect GetStatementOfAffairsResource to be called once and return statement of affairs
		mockService.EXPECT().GetStatementOfAffairsResource(transactionID).Return(statementOfAffairs, nil)

		res := serveHandleGetStatementOfAffairs(mockService, true)

		So(res.Code, ShouldEqual, http.StatusOK)
	})
}

func serveHandleDeleteStatementOfAffairs(service dao.Service, tranIDSet bool) *httptest.ResponseRecorder {
	path := "/transactions/123456789/insolvency/statement-of-affairs"
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	if tranIDSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}
	res := httptest.NewRecorder()

	handler := HandleDeleteStatementOfAffairs(service)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleDeleteStatementOfAffairs(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	Convey("Must need a transaction ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveHandleDeleteStatementOfAffairs(mock_dao.NewMockService(mockCtrl), false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error checking if transaction is closed against transaction api", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		// Expect the transaction api to be called and return an error
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

		res := serveHandleDeleteStatementOfAffairs(mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Transaction is already closed and cannot be updated", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		// Expect the transaction api to be called and return an already closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		res := serveHandleDeleteStatementOfAffairs(mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusForbidden)
	})

	Convey("Error when deleting statement of affairs from DB", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		mockService := mock_dao.NewMockService(mockCtrl)
		mockService.EXPECT().DeleteStatementOfAffairsResource(transactionID).Return(http.StatusInternalServerError, fmt.Errorf("err"))

		res := serveHandleDeleteStatementOfAffairs(mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Statement of affairs not found when deleting from DB", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		mockService := mock_dao.NewMockService(mockCtrl)
		mockService.EXPECT().DeleteStatementOfAffairsResource(transactionID).Return(http.StatusNotFound, fmt.Errorf("err"))

		res := serveHandleDeleteStatementOfAffairs(mockService, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Successfully delete statement of affairs from DB", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		mockService := mock_dao.NewMockService(mockCtrl)
		mockService.EXPECT().DeleteStatementOfAffairsResource(transactionID).Return(http.StatusNoContent, nil)

		res := serveHandleDeleteStatementOfAffairs(mockService, true)

		So(res.Code, ShouldEqual, http.StatusNoContent)
	})
}

func generateStatement() models.StatementOfAffairs {
	return models.StatementOfAffairs{
		StatementDate: "2021-06-06",
		Attachments: []string{
			"123456789",
		},
	}
}
