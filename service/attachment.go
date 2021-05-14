package service

import (
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/go-sdk-manager/manager"
	"github.com/companieshouse/insolvency-api/dao"
)

func UploadAttachment(req *http.Request, dao dao.Service) (string, ResponseType, error) {
	// Decode the incoming request
	err := req.ParseMultipartForm(32 << 20) // TODO what should this value be set to?
	if err != nil {
		err = fmt.Errorf("error parsing form: [%v]", err)
		log.ErrorR(req, err)
		return "", InvalidData, err
	}

	file, header, err := req.FormFile("file")

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
