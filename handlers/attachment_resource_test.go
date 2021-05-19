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

func getBodyWithFile(attachmentType string) (*bytes.Buffer, error) {
	file, err := os.Open("handlers/attachment_test.txt")
	if err != nil {
		return nil, err
	}

	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.SetBoundary("test_boundary")
	part, err := writer.CreateFormFile("file", "handlers/attachment_test.txt")
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

	Convey("Validation failed", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		body, err := getBodyWithFile("")
		if err != nil {
			t.Error(err)
		}

		res := serveHandleSubmitAttachment((body).Bytes(), mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error uploading attachment", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		body, err := getBodyWithFile("resolution")
		if err != nil {
			t.Error(err)
		}

		res := serveHandleSubmitAttachment((body).Bytes(), mock_dao.NewMockService(mockCtrl), true)

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

		body, err := getBodyWithFile("resolution")
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

		body, err := getBodyWithFile("resolution")
		if err != nil {
			t.Error(err)
		}

		res := serveHandleSubmitAttachment((body).Bytes(), mockService, true)

		So(res.Code, ShouldEqual, http.StatusCreated)
	})
}
