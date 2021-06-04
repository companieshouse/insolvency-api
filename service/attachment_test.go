package service

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/mocks"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/golang/mock/gomock"
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitValidateAttachmentDetails(t *testing.T) {

	Convey("Invalid attachment details - invalid attachment type", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(make([]models.AttachmentResourceDao, 0), nil)

		validationErrs, err := ValidateAttachmentDetails(mockService, transactionID, "invalid", createHeader())
		So(validationErrs, ShouldEqual, "attachment_type is invalid")
		So(err, ShouldBeNil)
	})

	Convey("Error when validating attachment details - error retrieving attachments", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(nil, fmt.Errorf("error getting attachments for transaction ID [%s]", transactionID))

		validationErrs, err := ValidateAttachmentDetails(mockService, transactionID, "resolution", createHeader())
		So(validationErrs, ShouldBeEmpty)
		So(err.Error(), ShouldContainSubstring, "error getting attachments for transaction ID")
	})

	Convey("Invalid attachment details - attempt to file attachment with type that has already been filed", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(generateAttachment(), nil)

		validationErrs, err := ValidateAttachmentDetails(mockService, transactionID, "resolution", createHeader())
		So(validationErrs, ShouldEqual, fmt.Sprintf("attachment of type [%s] has already been filed for insolvency case with transaction ID [%s]", "resolution", transactionID))
		So(err, ShouldBeNil)
	})

	Convey("Invalid attachment details - statement of affairs liquidator filed for case", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		attachmentResources := generateAttachment()
		attachmentResources[0].Type = constants.StatementOfAffairsLiquidator.String()

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(attachmentResources, nil)

		validationErrs, err := ValidateAttachmentDetails(mockService, transactionID, "resolution", createHeader())
		So(validationErrs, ShouldEqual, fmt.Sprintf("attachment of type [%s] has been filed for insolvency case with transaction ID [%s] - no other attachments can be filed for this case", constants.StatementOfAffairsLiquidator.String(), transactionID))
		So(err, ShouldBeNil)
	})

	Convey("Invalid attachment details - attempt to file statement of affairs liquidator for case that already has attachments", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		attachmentResources := generateAttachment()
		attachmentResources[0].Type = constants.Resolution.String()

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(attachmentResources, nil)

		validationErrs, err := ValidateAttachmentDetails(mockService, transactionID, constants.StatementOfAffairsLiquidator.String(), createHeader())
		So(validationErrs, ShouldEqual, fmt.Sprintf("attachment of type [%s] cannot be filed for insolvency case with transaction ID [%s] - other attachments have already been filed for this case", constants.StatementOfAffairsLiquidator.String(), transactionID))
		So(err, ShouldBeNil)
	})

	Convey("Invalid attachment details - invalid file format in header and name", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(make([]models.AttachmentResourceDao, 0), nil)

		header := createHeader()
		header.Header.Set("Content-Type", "invalid")
		header.Filename = "test.txt"

		validationErrs, err := ValidateAttachmentDetails(mockService, transactionID, "resolution", header)
		So(validationErrs, ShouldEqual, "attachment file format should be pdf")
		So(err, ShouldBeNil)
	})

	Convey("Invalid attachment details - invalid file format in header and name", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(make([]models.AttachmentResourceDao, 0), nil)

		header := createHeader()
		header.Size = 10000000000

		validationErrs, err := ValidateAttachmentDetails(mockService, transactionID, "resolution", header)
		So(validationErrs, ShouldEqual, "attachment file size is too large to be processed")
		So(err, ShouldBeNil)
	})

	Convey("Valid attachment details - with existing attachments filed", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		attachmentResources := generateAttachment()
		attachmentResources[0].Type = constants.Resolution.String()

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(attachmentResources, nil)

		validationErrs, err := ValidateAttachmentDetails(mockService, transactionID, constants.StatementOfAffairsDirector.String(), createHeader())
		So(validationErrs, ShouldBeEmpty)
		So(err, ShouldBeNil)
	})

	Convey("Valid attachment details - without existing attachments filed", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetAttachmentResources(transactionID).Return(make([]models.AttachmentResourceDao, 0), nil)

		validationErrs, err := ValidateAttachmentDetails(mockService, transactionID, "resolution", createHeader())
		So(validationErrs, ShouldBeEmpty)
		So(err, ShouldBeNil)
	})
}

func createHeader() *multipart.FileHeader {
	return &multipart.FileHeader{
		Filename: "test.pdf",
		Header: textproto.MIMEHeader{
			"Content-Type": []string{
				"application/pdf",
			},
		},
		Size: 1,
	}
}

func generateAttachment() []models.AttachmentResourceDao {
	return []models.AttachmentResourceDao{
		{
			ID:     "1111",
			Type:   "resolution",
			Status: "submitted",
			Links: models.AttachmentResourceLinksDao{
				Self:     "/transaction/" + transactionID + "/insolvency/attachments/1111",
				Download: "/transaction/" + transactionID + "/insolvency/attachments/1111/download",
			},
		},
	}
}

func TestUnitDownloadAttachment(t *testing.T) {
	attachmentID := "123"

	Convey("Download attachment - no response", t, func() {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		responseType, err := DownloadAttachment(attachmentID, req, w)
		So(responseType, ShouldEqual, Error)
		So(err.Error(), ShouldEqual, "error communicating with the File Transfer API: [error downloading file, no response]")
	})

	Convey("Download attachment - success", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, `=~.*`, httpmock.NewStringResponder(http.StatusOK, `{"id": "12345"}`))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		responseType, err := DownloadAttachment(attachmentID, req, w)
		So(responseType, ShouldEqual, Success)
		So(err, ShouldBeNil)
	})
}
