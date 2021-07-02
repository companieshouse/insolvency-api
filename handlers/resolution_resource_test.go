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
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

func serveHandleCreateResolution(body []byte, service dao.Service, tranIDSet bool) *httptest.ResponseRecorder {
	path := "/transactions/123456789/insolvency/resolution"
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	if tranIDSet {
		req = mux.SetURLVars(req, map[string]string{"transaction_id": transactionID})
	}
	res := httptest.NewRecorder()

	handler := HandleCreateResolution(service)
	handler.ServeHTTP(res, req)

	return res
}

func TestUnitHandleCreateResolution(t *testing.T) {
	err := os.Chdir("..")
	if err != nil {
		log.ErrorR(nil, fmt.Errorf("error accessing root directory"))
	}

	Convey("Must need a transaction ID in the url", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		res := serveHandleCreateResolution(body, mock_dao.NewMockService(mockCtrl), false)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Error checking if transaction is closed against transaction api", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		// Expect the transaction api to be called and return an error
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		res := serveHandleCreateResolution(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Transaction is already closed and cannot be updated", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer httpmock.DeactivateAndReset()
		defer mockCtrl.Finish()

		// Expect the transaction api to be called and return an already closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		body, _ := json.Marshal(&models.InsolvencyRequest{})
		res := serveHandleCreateResolution(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusForbidden)
	})

	Convey("Failed to read request body", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		body := []byte(`{"first_name":error`)
		res := serveHandleCreateResolution(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
	})

	Convey("Incoming request has date of resolution missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		resolution := generateResolution()
		resolution.DateOfResolution = ""
		body, _ := json.Marshal(resolution)
		res := serveHandleCreateResolution(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "date_of_resolution is a required field")
	})

	Convey("Incoming request has invalid date format", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		resolution := generateResolution()
		resolution.DateOfResolution = "21-01-01"
		body, _ := json.Marshal(resolution)
		res := serveHandleCreateResolution(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "date_of_resolution does not match the 2006-01-02 format")
	})

	Convey("Incoming request has attachments missing", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		resolution := generateResolution()
		resolution.Attachments = nil
		body, _ := json.Marshal(resolution)
		res := serveHandleCreateResolution(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "attachments is a required field")
	})

	Convey("Incoming request has empty attachments array", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		resolution := generateResolution()
		resolution.Attachments = []string{}
		body, _ := json.Marshal(resolution)
		res := serveHandleCreateResolution(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "no attachment has been supplied")
	})

	Convey("Incoming request has more than one attachment", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		resolution := generateResolution()
		resolution.Attachments = []string{
			"1234567890",
			"0987654321",
		}
		body, _ := json.Marshal(resolution)
		res := serveHandleCreateResolution(body, mock_dao.NewMockService(mockCtrl), true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "only one attachment can be supplied: [1234567890 0987654321]")
	})

	Convey("Attachment is not associated with transaction", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		resolution := generateResolution()
		resolution.Attachments = []string{
			"111-222-333-444",
		}
		body, _ := json.Marshal(resolution)

		// Expect GetAttachmentFromInsolvencyResource to be called once and return an empty attachment model, nil
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, resolution.Attachments[0]).Return(models.AttachmentResourceDao{}, nil)

		res := serveHandleCreateResolution(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
		So(res.Body.String(), ShouldContainSubstring, "attachment not found on transaction")
	})

	Convey("Attachment is not of type resolution", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		resolution := generateResolution()
		resolution.Attachments = []string{
			"111-222-333-444",
		}
		body, _ := json.Marshal(resolution)

		attachment := generateAttachment()
		attachment.Type = "not-resolution"

		// Expect GetAttachmentFromInsolvencyResource to be called once and return attachment, nil
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, resolution.Attachments[0]).Return(attachment, nil)

		res := serveHandleCreateResolution(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusBadRequest)
		So(res.Body.String(), ShouldContainSubstring, "attachment is not a resolution")
	})

	Convey("Generic error when adding resolution resource to mongo", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		mockService := mock_dao.NewMockService(mockCtrl)

		resolution := generateResolution()
		body, _ := json.Marshal(resolution)

		attachment := generateAttachment()

		// Expect GetAttachmentFromInsolvencyResource to be called once and return attachment, nil
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, resolution.Attachments[0]).Return(attachment, nil)

		// Expect CreateResolutionResource to be called once and return an error
		mockService.EXPECT().CreateResolutionResource(gomock.Any(), transactionID).Return(http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)).Times(1)

		res := serveHandleCreateResolution(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusInternalServerError)
	})

	Convey("Error adding resolution resource to mongo - insolvency case not found", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))

		mockService := mock_dao.NewMockService(mockCtrl)

		resolution := generateResolution()
		body, _ := json.Marshal(resolution)

		attachment := generateAttachment()

		// Expect GetAttachmentFromInsolvencyResource to be called once and return attachment, nil
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, resolution.Attachments[0]).Return(attachment, nil)

		// Expect CreateResolutionResource to be called once and return an error
		mockService.EXPECT().CreateResolutionResource(gomock.Any(), transactionID).Return(http.StatusNotFound, fmt.Errorf("there was a problem handling your request for transaction %s not found", transactionID)).Times(1)

		res := serveHandleCreateResolution(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("Successfully add insolvency resource to mongo", t, func() {
		httpmock.Activate()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.DeactivateAndReset()

		// Expect the transaction api to be called and return an open transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponse))
		mockService := mock_dao.NewMockService(mockCtrl)

		resolution := generateResolution()
		body, _ := json.Marshal(resolution)

		attachment := models.AttachmentResourceDao{
			ID:     "1111",
			Type:   "resolution",
			Status: "status",
			Links:  models.AttachmentResourceLinksDao{},
		}

		// Expect GetAttachmentFromInsolvencyResource to be called once and return attachment, nil
		mockService.EXPECT().GetAttachmentFromInsolvencyResource(transactionID, resolution.Attachments[0]).Return(attachment, nil)

		// Expect CreateResolutionResource to be called once and return an error
		mockService.EXPECT().CreateResolutionResource(gomock.Any(), transactionID).Return(http.StatusCreated, nil).Times(1)

		res := serveHandleCreateResolution(body, mockService, true)

		So(res.Code, ShouldEqual, http.StatusOK)
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
