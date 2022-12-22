package utils

import (
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/models"
)

// HelperService interface declares
type HelperService interface {
	// HandleTransactionIdExistsValidation
	HandleTransactionIdExistsValidation(w http.ResponseWriter, req *http.Request, transactionID string) (string, bool, int)
	// HandleTransactionNotClosedValidation
	HandleTransactionNotClosedValidation(w http.ResponseWriter, req *http.Request, transactionID string, isTransactionClosed bool, err error, httpStatus int) (error, bool, int)
	// HandleBodyDecodedValidation
	HandleBodyDecodedValidation(w http.ResponseWriter, req *http.Request, transactionID string, err error) (bool, int)
	// HandleEtagGenerationValidation
	HandleEtagGenerationValidation(err error) bool
	// HandleCreateResourceValidation
	HandleCreateResourceValidation(w http.ResponseWriter, req *http.Request, err error, statusCode int) (bool, int)
	// GenerateEtag generates a random etag which is generated on every write action
	GenerateEtag() (string, error)
}

type helperService struct {
}

// HandleBodyDecodedValidation implements HelperService
func (*helperService) HandleBodyDecodedValidation(w http.ResponseWriter, req *http.Request, transactionID string, err error) (bool, int) {
	if err != nil {
		log.ErrorR(req, fmt.Errorf("invalid request"))
		m := models.NewMessageResponse(fmt.Sprintf("failed to read request body for transaction %s", transactionID))
		WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
		return false, http.StatusInternalServerError
	}
	return true, http.StatusOK
}

// HandleCreateResourceValidation implements HelperService
func (*helperService) HandleCreateResourceValidation(w http.ResponseWriter, req *http.Request, err error, statusCode int) (bool, int) {
	if err != nil {
		log.ErrorR(req, err)
		m := models.NewMessageResponse(err.Error())
		WriteJSONWithStatus(w, req, m, statusCode)
		return false, statusCode
	}
	return true, statusCode
}

// HandleEtagGenerationValidation implements HelperService
func (*helperService) HandleEtagGenerationValidation(err error) bool {
	if err != nil {
		log.Error(fmt.Errorf("error generating etag: [%s]", err))
		return false
	}
	return true
}

// HandleTransactionIdExistsValidation implements HelperService
func (*helperService) HandleTransactionIdExistsValidation(w http.ResponseWriter, req *http.Request, transactionID string) (string, bool, int) {
	if transactionID == "" {
		log.ErrorR(req, fmt.Errorf("there is no transaction ID in the URL path"))
		m := models.NewMessageResponse("transaction ID is not in the URL path")
		WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
		return "", false, http.StatusBadRequest
	}
	return transactionID, true, http.StatusOK
}

// HandleTransactionNotClosedValidation implements HelperService
func (*helperService) HandleTransactionNotClosedValidation(w http.ResponseWriter, req *http.Request, transactionID string, isTransactionClosed bool, err error, httpStatus int) (error, bool, int) {
	if err != nil {
		log.ErrorR(req, fmt.Errorf("error checking transaction status for [%v]: [%s]", transactionID, err))
		m := models.NewMessageResponse(fmt.Sprintf("error checking transaction status for [%v]: [%s]", transactionID, err))
		WriteJSONWithStatus(w, req, m, httpStatus)
		return nil, false, httpStatus
	}

	if isTransactionClosed {
		log.ErrorR(req, fmt.Errorf("transaction [%v] is already closed and cannot be updated", transactionID))
		m := models.NewMessageResponse(fmt.Sprintf("transaction [%v] is already closed and cannot be updated", transactionID))
		WriteJSONWithStatus(w, req, m, httpStatus)
		return nil, false, httpStatus
	}
	return err, true, httpStatus
}

// NewHelperService will create a new instance of the HelperService interface.
func NewHelperService() HelperService {
	return &helperService{}
}
