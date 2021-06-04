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

func serveHandleSubmitAttachment(body []byte, service dao.Service, tranIDSet bool) *httptest.ResponseRecorder {
	ctx := context.WithValue(context.Background(), httpsession.ContextKeySession, &session.Session{})
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body)).WithContext(ctx)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=test_boundary")
	if tranIDSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}
	res := httptest.NewRecorder()

	handler := HandleSubmitAttachment(service)
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

	Convey("Must have a transaction ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		res := serveHandleSubmitAttachment(body, mock_dao.NewMockService(mockCtrl), false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Failed to read request body", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		body := []byte(`{"company_name":error`)
		res := serveHandleSubmitAttachment(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Validation failed - invalid attachment type", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		body, err := getBodyWithFile("", pdfFilePath)
		if err != nil {
			t.Error(err)
		}

		mockService := mock_dao.NewMockService(mockCtrl)
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(make([]models.AttachmentResourceDao, 0), nil)

		res := serveHandleSubmitAttachment((body).Bytes(), mockService, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Validation failed - invalid attachment file format", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		body, err := getBodyWithFile("resolution", txtFilePath)
		if err != nil {
			t.Error(err)
		}

		mockService := mock_dao.NewMockService(mockCtrl)
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(make([]models.AttachmentResourceDao, 0), nil)

		res := serveHandleSubmitAttachment((body).Bytes(), mockService, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Validation failed - attachment with type has already been filed", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		body, err := getBodyWithFile("resolution", txtFilePath)
		if err != nil {
			t.Error(err)
		}

		attachments := []models.AttachmentResourceDao{
			{
				Type: "resolution",
			},
		}

		mockService := mock_dao.NewMockService(mockCtrl)
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(attachments, nil)

		res := serveHandleSubmitAttachment((body).Bytes(), mockService, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error uploading attachment", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		body, err := getBodyWithFile("resolution", pdfFilePath)
		if err != nil {
			t.Error(err)
		}

		mockService := mock_dao.NewMockService(mockCtrl)
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(make([]models.AttachmentResourceDao, 0), nil)

		res := serveHandleSubmitAttachment((body).Bytes(), mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Error updating DB", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder(http.MethodPost, `=~.*`, httpmock.NewStringResponder(http.StatusCreated, `{"id": "12345"}`))

		mockService := mock_dao.NewMockService(mockCtrl)
		mockService.EXPECT().AddAttachmentToInsolvencyResource(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("err"))
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(make([]models.AttachmentResourceDao, 0), nil)

		body, err := getBodyWithFile("resolution", pdfFilePath)
		if err != nil {
			t.Error(err)
		}

		res := serveHandleSubmitAttachment((body).Bytes(), mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Error when retrieving existing attachments from DB", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder(http.MethodPost, `=~.*`, httpmock.NewStringResponder(http.StatusCreated, `{"id": "12345"}`))

		mockService := mock_dao.NewMockService(mockCtrl)
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(nil, fmt.Errorf("err"))

		body, err := getBodyWithFile("resolution", pdfFilePath)
		if err != nil {
			t.Error(err)
		}

		res := serveHandleSubmitAttachment((body).Bytes(), mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Success", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		httpmock.RegisterResponder(http.MethodPost, `=~.*`, httpmock.NewStringResponder(http.StatusCreated, `{"id": "12345"}`))

		mockService := mock_dao.NewMockService(mockCtrl)
		daoResponse := models.AttachmentResourceDao{
			Type: "attachment",
		}
		mockService.EXPECT().AddAttachmentToInsolvencyResource(gomock.Any(), gomock.Any(), gomock.Any()).Return(&daoResponse, nil)
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(make([]models.AttachmentResourceDao, 0), nil)

		body, err := getBodyWithFile("resolution", pdfFilePath)
		if err != nil {
			t.Error(err)
		}

		res := serveHandleSubmitAttachment((body).Bytes(), mockService, true)

		So(res.Code, ShouldEqual, http.StatusCreated)
	})
}

func serveHandleGetAttachmentDetails(service dao.Service, tranIDSet bool, attachmentIDSet bool) *httptest.ResponseRecorder {
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

	handler := HandleGetAttachmentDetails(service)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleGetAttachment(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	Convey("Must have a transaction ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveHandleGetAttachmentDetails(mock_dao.NewMockService(mockCtrl), false, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Must have an attachment ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		res := serveHandleGetAttachmentDetails(mock_dao.NewMockService(mockCtrl), true, false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Failed to get attachment from DB", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)

		// Expect GetAttachmentFromInsolvencyResource to be called once and return an error
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, attachmentID).Return(nil, fmt.Errorf("failed to get attachment from insolvency resource in db for transaction [%s] with attachment id of [%s]: %v", transactionID, attachmentID, err))

		res := serveHandleGetAttachmentDetails(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("No attachment associated with Insolvency case", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)

		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, attachmentID).Return(nil, nil).Times(1)

		res := serveHandleGetAttachmentDetails(mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "attachment id is not valid")
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

	Convey("Must have a transaction ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		res := serveHandleDownloadAttachment(body, mock_dao.NewMockService(mockCtrl), false, false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Must have an attachment ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		res := serveHandleDownloadAttachment(body, mock_dao.NewMockService(mockCtrl), true, false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error getting insolvency resource", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, attachmentID).Return(make([]models.AttachmentResourceDao, 0), fmt.Errorf("err"))

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		res := serveHandleDownloadAttachment(body, mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Invalid attachment ID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		response := []models.AttachmentResourceDao{
			{ID: "xyz"},
		}
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, attachmentID).Return(response, nil)

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		res := serveHandleDownloadAttachment(body, mockService, true, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error getting attachment details", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)
		response := []models.AttachmentResourceDao{
			{ID: attachmentID},
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
		httpmock.RegisterResponder(http.MethodGet, `=~` + attachmentID + `$`, httpmock.NewStringResponder(http.StatusOK, `{"av_status": "virus"}`))

		mockService := mock_dao.NewMockService(mockCtrl)
		response := []models.AttachmentResourceDao{
			{ID: attachmentID},
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
			httpmock.RegisterResponder(http.MethodGet, `=~` + attachmentID + `$`, httpmock.NewStringResponder(http.StatusOK, `{"av_status": "clean"}`))
			httpmock.RegisterResponder(http.MethodGet, `=~download$`, httpmock.NewStringResponder(http.StatusTeapot, ""))

			mockService := mock_dao.NewMockService(mockCtrl)
			response := []models.AttachmentResourceDao{
				{ID: attachmentID},
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

			httpmock.RegisterResponder(http.MethodGet, `=~` + attachmentID + `$`, httpmock.NewStringResponder(http.StatusOK, `{"av_status": "clean"}`))
			httpmock.RegisterResponder(http.MethodGet, `=~download$`, httpmock.NewStringResponder(http.StatusOK, `downloaded`))

			mockService := mock_dao.NewMockService(mockCtrl)
			response := []models.AttachmentResourceDao{
				{ID: attachmentID},
			}
			mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, attachmentID).Return(response, nil)

			body, _ := json.Marshal(&models.InsolvencyRequest{})
			res := serveHandleDownloadAttachment(body, mockService, true, true)

			So(res.Code, ShouldEqual, http.StatusOK)
		})
}
