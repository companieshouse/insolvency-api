package service

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/go-sdk-manager/manager"
	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/dao"
)

const maxFileSize = 1048576 * 4 // 4MB

// UploadAttachment sends a file to be uploaded to the File Transfer API and returns the ID
func UploadAttachment(file multipart.File, header *multipart.FileHeader, req *http.Request) (string, ResponseType, error) {
	// Create SDK session
	api, err := manager.GetSDK(req)
	if err != nil {
		err = fmt.Errorf("error creating SDK to upload attachment: [%v]", err)
		log.ErrorR(req, err)
		return "", Error, err
	}

	uploadedFileResponse, err := api.FileTransfer.UploadFile(file, header).Do()
	if err != nil {
		err = fmt.Errorf("error communicating with the File Transfer API: [%v]", err)
		log.ErrorR(req, err)
		return "", Error, err
	}

	if uploadedFileResponse == nil {
		err = fmt.Errorf("error uploading file: [%v]", err)
		log.ErrorR(req, err)
		return "", Error, err
	}
	return uploadedFileResponse.Id, Success, nil
}

// ValidateAttachmentDetails checks that the incoming attachment details are valid
func ValidateAttachmentDetails(svc dao.Service, transactionID string, attachmentType string, header *multipart.FileHeader) (string, error) {
	var errs []string

	// Check attachment is of valid type
	if !constants.IsAttachmentTypeValid(attachmentType) {
		errs = append(errs, "attachment_type is invalid")
	}

	// Check if attachment has already been filed
	attachments, err := svc.GetAttachmentResources(transactionID)
	if err != nil {
		return "", err
	}
	if len(attachments) > 0 {
		for _, a := range attachments {
			if a.Type == attachmentType {
				errs = append(errs, fmt.Sprintf("attachment of type [%s] has already been filed for insolvency case with transaction ID [%s]", attachmentType, transactionID))
				break
			}
		}
	}

	// Check file type is PDF
	fileType := header.Header.Get("Content-Type")
	if fileType != "application/pdf" && header.Filename[len(header.Filename)-3:] != "pdf" {
		errs = append(errs, "attachment file format should be pdf")
	}

	// Check if attachment size is less than maxFileSize
	if header.Size > maxFileSize {
		errs = append(errs, "attachment file size is too large to be processed")
	}

	return strings.Join(errs, ", "), nil
}