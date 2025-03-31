package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/go-session-handler/httpsession"
	"github.com/companieshouse/go-session-handler/session"
	"github.com/companieshouse/insolvency-api/dao"
	mock_dao "github.com/companieshouse/insolvency-api/mocks"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	pdfFilePath  = "handlers/attachment_test.pdf"
	txtFilePath  = "handlers/attachment_test.txt"
	attachmentID = "987654321"
)

func serveHandleSubmitAttachment(body []byte, service dao.Service, tranIDSet bool, helperService utils.HelperService, res *httptest.ResponseRecorder) *httptest.ResponseRecorder {
	ctx := context.WithValue(context.Background(), httpsession.ContextKeySession, &session.Session{})
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body)).WithContext(ctx)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=test_boundary")
	if tranIDSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}

	handler := HandleSubmitAttachment(service, helperService)
	handler.ServeHTTP(res, req)

	return res
}

func getBodyWithFile(attachmentType string, filePath string) (*bytes.Buffer, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.SetBoundary("test_boundary")
	part, err := writer.CreateFormFile("file", filePath)
	if attachmentType != "" {
		writer.WriteField("attachment_type", attachmentType)
	}
	if err != nil {
		writer.Close()
		return nil, err
	}
	io.Copy(part, file)
	writer.Close()
	return body, nil
}

func TestUnitHandleSubmitAttachment(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	helperService := utils.NewHelperService()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	Convey("Must have a transaction ID in the url", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)

		body, _ := json.Marshal(&models.InsolvencyRequest{})

		res := serveHandleSubmitAttachment(body, mockService, false, helperService, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "transaction ID is not in the URL path")
	})

	Convey("Error checking if transaction is closed against transaction api", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an error
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

		body, _ := json.Marshal(&models.InsolvencyRequest{})

		res := serveHandleSubmitAttachment(body, mockService, true, helperService, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
		So(res.Body.String(), ShouldContainSubstring, "error checking transaction status")
	})

	Convey("Transaction is already closed and cannot be updated", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an already closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		body, _ := json.Marshal(&models.InsolvencyRequest{})

		res := serveHandleSubmitAttachment(body, mockService, true, helperService, rec)

		So(res.Code, ShouldEqual, http.StatusForbidden)
		So(res.Body.String(), ShouldContainSubstring, "already closed and cannot be updated")
	})

	Convey("Failed to read form from request body", t, func() {
		mockService, _, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		body := []byte(`{"company_name":error`)

		res := serveHandleSubmitAttachment(body, mockService, true, helperService, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "error reading form from request")
	})

	Convey("Validation failed - invalid attachment type", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		body, err := getBodyWithFile("", pdfFilePath)
		if err != nil {
			t.Error(err)
		}

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(make([]models.AttachmentResourceDao, 0), nil)

		res := serveHandleSubmitAttachment((body).Bytes(), mockService, true, mockHelperService, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Validation failed - invalid attachment file format", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		body, err := getBodyWithFile("resolution", txtFilePath)
		if err != nil {
			t.Error(err)
		}

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(make([]models.AttachmentResourceDao, 0), nil)

		res := serveHandleSubmitAttachment((body).Bytes(), mockService, true, mockHelperService, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Validation failed - attachment with type has already been filed", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		body, err := getBodyWithFile("resolution", txtFilePath)
		if err != nil {
			t.Error(err)
		}

		attachments := []models.AttachmentResourceDao{
			{
				Type: "resolution",
			},
		}

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(attachments, nil)

		res := serveHandleSubmitAttachment((body).Bytes(), mockService, true, mockHelperService, rec)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error uploading attachment", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		body, err := getBodyWithFile("resolution", pdfFilePath)
		if err != nil {
			t.Error(err)
		}

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(make([]models.AttachmentResourceDao, 0), nil)

		res := serveHandleSubmitAttachment((body).Bytes(), mockService, true, mockHelperService, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Error updating DB", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))
		httpmock.RegisterResponder(http.MethodPost, `=~.*`, httpmock.NewStringResponder(http.StatusCreated, `{"id": "12345"}`))

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockService.EXPECT().AddAttachmentToInsolvencyResource(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("err"))
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(make([]models.AttachmentResourceDao, 0), nil)

		body, err := getBodyWithFile("resolution", pdfFilePath)
		if err != nil {
			t.Error(err)
		}

		res := serveHandleSubmitAttachment((body).Bytes(), mockService, true, mockHelperService, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Error when retrieving existing attachments from DB", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))
		httpmock.RegisterResponder(http.MethodPost, `=~.*`, httpmock.NewStringResponder(http.StatusCreated, `{"id": "12345"}`))

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(nil, fmt.Errorf("err"))

		body, err := getBodyWithFile("resolution", pdfFilePath)
		if err != nil {
			t.Error(err)
		}

		res := serveHandleSubmitAttachment((body).Bytes(), mockService, true, mockHelperService, rec)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Error when retrieving existing attachments from DB - insolvency case not found", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))
		httpmock.RegisterResponder(http.MethodPost, `=~.*`, httpmock.NewStringResponder(http.StatusCreated, `{"id": "12345"}`))

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(nil, nil)

		body, err := getBodyWithFile("resolution", pdfFilePath)
		if err != nil {
			t.Error(err)
		}

		res := serveHandleSubmitAttachment((body).Bytes(), mockService, true, mockHelperService, rec)

		So(res.Code, ShouldEqual, http.StatusNotFound)
		So(res.Body.String(), ShouldContainSubstring, "insolvency case not found")
	})

	Convey("Success", t, func() {
		mockService, mockHelperService, rec := mock_dao.CreateTestObjects(t)
		httpmock.Activate()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))
		httpmock.RegisterResponder(http.MethodPost, `=~.*`, httpmock.NewStringResponder(http.StatusCreated, `{"id": "12345"}`))

		daoResponse := models.AttachmentResourceDao{
			Type: "attachment",
		}

		mockHelperService.EXPECT().HandleTransactionIdExistsValidation(gomock.Any(), gomock.Any(), transactionID).Return(true, transactionID).AnyTimes()
		mockHelperService.EXPECT().HandleTransactionNotClosedValidation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
		mockService.EXPECT().AddAttachmentToInsolvencyResource(gomock.Any(), gomock.Any(), gomock.Any()).Return(&daoResponse, nil)
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(make([]models.AttachmentResourceDao, 0), nil)
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()

		body, err := getBodyWithFile("resolution", pdfFilePath)
		if err != nil {
			t.Error(err)
		}

		res := serveHandleSubmitAttachment((body).Bytes(), mockService, true, mockHelperService, rec)

		So(res.Code, ShouldEqual, http.StatusCreated)
	})
}

func serveHandleGetAttachmentDetails(service dao.Service, tranIDSet bool, attachmentIDSet bool, helperService utils.HelperService) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	vars := make(map[string]string)
	if tranIDSet {
		vars["transaction_id"] = transactionID
	}
	if attachmentIDSet {
		vars["attachment_id"] = attachmentID
	}
	req = mux.SetURLVars(req, vars)
	res := httptest.NewRecorder()

	handler := HandleGetAttachmentDetails(service, helperService)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleGetAttachment(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockService := mock_dao.NewMockService(mockCtrl)
	mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

	Convey("Must have a transaction ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil)

		res := serveHandleGetAttachmentDetails(mockService, false, true, mockHelperService)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Must have an attachment ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveHandleGetAttachmentDetails(mockService, true, false, mockHelperService)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Failed to get attachment from DB", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		httpmock.RegisterResponder(http.MethodGet, `=~.*`, httpmock.NewStringResponder(http.StatusOK, `{"name": "file"}`))

		// Expect GetAttachmentFromInsolvencyResource to be called once and return an error
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, attachmentID).Return(models.AttachmentResourceDao{}, fmt.Errorf("failed to get attachment from insolvency resource in db for transaction [%s] with attachment id of [%s]: %v", transactionID, attachmentID, err))

		res := serveHandleGetAttachmentDetails(mockService, true, true, mockHelperService)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("No attachment associated with Insolvency case", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		httpmock.RegisterResponder(http.MethodGet, `=~.*`, httpmock.NewStringResponder(http.StatusOK, `{"name": "file"}`))

		// Expect GetAttachmentFromInsolvencyResource to be called once and return nothing
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, attachmentID).Return(models.AttachmentResourceDao{}, nil).Times(1)

		res := serveHandleGetAttachmentDetails(mockService, true, true, mockHelperService)

		So(res.Code, ShouldEqual, http.StatusNotFound)
		So(res.Body.String(), ShouldContainSubstring, "attachment id is not valid")
	})

	Convey("Success", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		httpmock.RegisterResponder(http.MethodGet, `=~.*`, httpmock.NewStringResponder(http.StatusOK, `{"name": "file"}`))

		attachment := models.AttachmentResourceDao{
			ID:     "1111",
			Type:   "resolution",
			Status: "status",
			Links:  models.AttachmentResourceLinksDao{},
		}

		// Expect GetAttachmentFromInsolvencyResource to be called once and return the attachment
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, attachmentID).Return(attachment, nil)

		res := serveHandleGetAttachmentDetails(mockService, true, true, mockHelperService)

		So(res.Code, ShouldEqual, http.StatusOK)
	})
}

func serveHandleDownloadAttachment(body []byte, service dao.Service, tranIDSet bool, attachmentIDSet bool) *httptest.ResponseRecorder {
	ctx := context.WithValue(context.Background(), httpsession.ContextKeySession, &session.Session{})
	req := httptest.NewRequest(http.MethodGet, "/test", bytes.NewReader(body)).WithContext(ctx)
	vars := make(map[string]string)
	if tranIDSet {
		vars["transaction_id"] = transactionID
	}
	if attachmentIDSet {
		vars["attachment_id"] = attachmentID
	}

	req = mux.SetURLVars(req, vars)

	res := httptest.NewRecorder()

	handler := HandleDownloadAttachment(service)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleDownloadAttachment(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockService := mock_dao.NewMockService(mockCtrl)

	Convey("Must have a transaction ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		res := serveHandleDownloadAttachment(body, mockService, false, false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Must have an attachment ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		res := serveHandleDownloadAttachment(body, mockService, true, false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error getting insolvency resource", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, attachmentID).Return(models.AttachmentResourceDao{}, fmt.Errorf("err"))

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		res := serveHandleDownloadAttachment(body, mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Invalid attachment ID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, attachmentID).Return(models.AttachmentResourceDao{}, nil)

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		res := serveHandleDownloadAttachment(body, mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Error getting attachment details", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		response := models.AttachmentResourceDao{
			ID: attachmentID,
		}

		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, attachmentID).Return(response, nil)

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		res := serveHandleDownloadAttachment(body, mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("status not clean", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, `=~`+attachmentID+`$`, httpmock.NewStringResponder(http.StatusOK, `{"av_status": "virus"}`))

		response := models.AttachmentResourceDao{
			ID: attachmentID,
		}
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, attachmentID).Return(response, nil)

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		res := serveHandleDownloadAttachment(body, mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusForbidden)
	})

	Convey("Error downloading attachment", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, `=~`+attachmentID+`$`, httpmock.NewStringResponder(http.StatusOK, `{"av_status": "clean"}`))
		httpmock.RegisterResponder(http.MethodGet, `=~download$`, httpmock.NewStringResponder(http.StatusTeapot, ""))

		response := models.AttachmentResourceDao{
			ID: attachmentID,
		}
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, attachmentID).Return(response, nil)

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		res := serveHandleDownloadAttachment(body, mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Successful download", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodGet, `=~`+attachmentID+`$`, httpmock.NewStringResponder(http.StatusOK, `{"av_status": "clean"}`))
		httpmock.RegisterResponder(http.MethodGet, `=~download$`, httpmock.NewStringResponder(http.StatusOK, `downloaded`))

		response := models.AttachmentResourceDao{
			ID: attachmentID,
		}
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, attachmentID).Return(response, nil)

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		res := serveHandleDownloadAttachment(body, mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusOK)
	})
}

func serveHandleDeleteAttachment(service dao.Service, tranIDSet bool, attachmentIDSet bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, "/test", nil)
	vars := make(map[string]string)
	if tranIDSet {
		vars["transaction_id"] = transactionID
	}
	if attachmentIDSet {
		vars["attachment_id"] = attachmentID
	}
	req = mux.SetURLVars(req, vars)
	res := httptest.NewRecorder()

	handler := HandleDeleteAttachment(service)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleDeleteAttachment(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockService := mock_dao.NewMockService(mockCtrl)

	Convey("Must have a transaction ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveHandleDeleteAttachment(mockService, false, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Must have an attachment ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveHandleDeleteAttachment(mockService, true, false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error checking if transaction is closed against transaction api", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		// Expect the transaction api to be called and return an error
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

		res := serveHandleDeleteAttachment(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Transaction is already closed and cannot be updated", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		// Expect the transaction api to be called and return an already closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		res := serveHandleDeleteAttachment(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusForbidden)
	})

	Convey("Failed to delete attachment from DB", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))
		httpmock.RegisterResponder(http.MethodDelete, `=~.*`, httpmock.NewStringResponder(http.StatusNoContent, ``))

		// Expect DeleteAttachmentResource to be called once and return an error
		mockService.EXPECT().DeleteAttachmentResource(transactionID, attachmentID).Return(http.StatusInternalServerError, fmt.Errorf("err"))

		res := serveHandleDeleteAttachment(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Attachment resource not found in DB", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))
		httpmock.RegisterResponder(http.MethodDelete, `=~.*`, httpmock.NewStringResponder(http.StatusNoContent, ``))

		// Expect DeleteAttachmentResource to be called once and return a not found
		mockService.EXPECT().DeleteAttachmentResource(transactionID, attachmentID).Return(http.StatusNotFound, fmt.Errorf("attachment not found"))

		res := serveHandleDeleteAttachment(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Success", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))
		httpmock.RegisterResponder(http.MethodDelete, `=~.*`, httpmock.NewStringResponder(http.StatusNoContent, ``))

		// Expect GetAttachmentFromInsolvencyResource to be called once and return an error
		mockService.EXPECT().DeleteAttachmentResource(transactionID, attachmentID).Return(http.StatusNoContent, nil)

		res := serveHandleDeleteAttachment(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusNoContent)
	})

}
