package utils

import (
	"fmt"
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/models"
	"net/http"
)

// HelperService interface declares
type HelperService interface {
	HandleTransactionIdExistsValidation(w http.ResponseWriter, req *http.Request, transactionID string) (bool, string)
	HandleTransactionNotClosedValidation(w http.ResponseWriter, req *http.Request, transactionID string, isTransactionClosed bool, httpStatus int, err error) bool
	HandleBodyDecodedValidation(w http.ResponseWriter, req *http.Request, transactionID string, err error) bool
	HandleMandatoryFieldValidation(w http.ResponseWriter, req *http.Request, err string) bool
	HandleAttachmentValidation(w http.ResponseWriter, req *http.Request, transactionID string, attachment models.AttachmentResourceDao, err error) bool
	HandleAttachmentTypeValidation(w http.ResponseWriter, req *http.Request, responseMessage string, err error) int
	HandleEtagGenerationValidation(err error) bool
	HandleCreateResourceValidation(w http.ResponseWriter, req *http.Request, statusCode int, err error) bool
	// GenerateEtag generates a random etag which is generated on every write action
	GenerateEtag() (string, error)
}

type helperService struct {
}

// HandleTransactionIdExistsValidation implements HelperService
func (*helperService) HandleTransactionIdExistsValidation(w http.ResponseWriter, req *http.Request, transactionID string) (bool, string) {
	if transactionID == "" {
		log.ErrorR(req, fmt.Errorf("there is no transaction ID in the URL path"))
		m := models.NewMessageResponse("transaction ID is not in the URL path")
		WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
		return false, ""
	}
	return true, transactionID
}

// HandleTransactionNotClosedValidation implements HelperService
func (*helperService) HandleTransactionNotClosedValidation(w http.ResponseWriter, req *http.Request, transactionID string, isTransactionClosed bool, httpStatus int, err error) bool {
	if err != nil {
		log.ErrorR(req, fmt.Errorf("error checking transaction status for [%v]: [%s]", transactionID, err))
		m := models.NewMessageResponse(fmt.Sprintf("error checking transaction status for [%v]: [%s]", transactionID, err))
		WriteJSONWithStatus(w, req, m, httpStatus)
		return false
	}

	if isTransactionClosed {
		log.ErrorR(req, fmt.Errorf("transaction [%v] is already closed and cannot be updated", transactionID))
		m := models.NewMessageResponse(fmt.Sprintf("transaction [%v] is already closed and cannot be updated", transactionID))
		WriteJSONWithStatus(w, req, m, httpStatus)
		return false
	}
	return true
}

// HandleBodyDecodedValidation implements HelperService
func (*helperService) HandleBodyDecodedValidation(w http.ResponseWriter, req *http.Request, transactionID string, err error) bool {
	if err != nil {
		log.ErrorR(req, fmt.Errorf("invalid request"))
		m := models.NewMessageResponse(fmt.Sprintf("failed to read request body for transaction %s", transactionID))
		WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
		return false
	}
	return true
}

// HandleMandatoryFieldValidation implements HelperService
func (*helperService) HandleMandatoryFieldValidation(w http.ResponseWriter, req *http.Request, errs string) bool {
	if errs != "" {
		log.ErrorR(req, fmt.Errorf("invalid request - failed validation on the following: %s", errs))
		m := models.NewMessageResponse("invalid request body: " + errs)
		WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
		return false
	}
	return true
}

// HandleAttachmentValidation implements HelperService
func (*helperService) HandleAttachmentValidation(w http.ResponseWriter, req *http.Request, transactionID string, attachment models.AttachmentResourceDao, err error) bool {
	if attachment == (models.AttachmentResourceDao{}) {
		log.ErrorR(req, fmt.Errorf("failed to get attachment from insolvency resource in db for transaction [%s] with attachment id of [%s]: %v", transactionID, attachment, err))
		m := models.NewMessageResponse("attachment not found on transaction")
		WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
		return false
	}
	return true
}

// HandleAttachmentTypeValidation implements HelperService
func (*helperService) HandleAttachmentTypeValidation(w http.ResponseWriter, req *http.Request, responseMessage string, err error) int {
	log.ErrorR(req, err)
	m := models.NewMessageResponse(responseMessage)
	WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
	return http.StatusBadRequest
}

// HandleCreateResourceValidation implements HelperService
func (*helperService) HandleCreateResourceValidation(w http.ResponseWriter, req *http.Request, statusCode int, err error) bool {
	if err != nil {
		log.ErrorR(req, err)
		m := models.NewMessageResponse(err.Error())
		WriteJSONWithStatus(w, req, m, statusCode)
		return false
	}
	return true
}

// HandleEtagGenerationValidation implements HelperService
func (*helperService) HandleEtagGenerationValidation(err error) bool {
	if err != nil {
		log.Error(fmt.Errorf("error generating etag: [%s]", err))
		return false
	}
	return true
}

// NewHelperService will create a new instance of the HelperService interface.
func NewHelperService() HelperService {
	return &helperService{}
}
