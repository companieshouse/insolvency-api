package service

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/go-sdk-manager/manager"
	"github.com/companieshouse/insolvency-api/constants"
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
func ValidateAttachmentDetails(attachmentType string, header *multipart.FileHeader, file multipart.File) (string, error) {
	var errs []string

	// Check attachment is of valid type
	if !constants.IsAttachmentTypeValid(attachmentType) {
		errs = append(errs, "attachment_type is invalid")
	}

	// Check if attachment file format is PDF
	fileHeader := make([]byte, 512)

	// Copy the headers into the FileHeader buffer
	if _, err := file.Read(fileHeader); err != nil {
		log.Error(fmt.Errorf("error reading file: [%s]", err))
		return "", fmt.Errorf("error reading file")
	}

	fileType := http.DetectContentType(fileHeader)
	if fileType != "application/pdf" {
		errs = append(errs, "attachment file format is not pdf")
	}

	// Check if attachment size is less than 4.5MB
	fmt.Println("File size: ", header.Size)
	if header.Size > maxFileSize {
		errs = append(errs, "attachment file size is too large to be processed")
	}

	return strings.Join(errs, ", "), nil
}
