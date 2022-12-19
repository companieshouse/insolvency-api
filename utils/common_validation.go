package utils

import (
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/models"
)

func HandleTransactionIdExistsValidation(w http.ResponseWriter, req *http.Request, transactionID string) (string, bool) {
	if transactionID == "" {
		log.ErrorR(req, fmt.Errorf("there is no transaction ID in the URL path"))
		m := models.NewMessageResponse("transaction ID is not in the URL path")
		WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
		return "", false
	}
	return transactionID, true
}

func HandleTransactionNotClosedValidation(w http.ResponseWriter, req *http.Request, transactionID string, isTransactionClosed bool, err error, httpStatus int) (error, bool) {
	if err != nil {
		log.ErrorR(req, fmt.Errorf("error checking transaction status for [%v]: [%s]", transactionID, err))
		m := models.NewMessageResponse(fmt.Sprintf("error checking transaction status for [%v]: [%s]", transactionID, err))
		WriteJSONWithStatus(w, req, m, httpStatus)
		return nil, false
	}

	if isTransactionClosed {
		log.ErrorR(req, fmt.Errorf("transaction [%v] is already closed and cannot be updated", transactionID))
		m := models.NewMessageResponse(fmt.Sprintf("transaction [%v] is already closed and cannot be updated", transactionID))
		WriteJSONWithStatus(w, req, m, httpStatus)
		return nil, false
	}
	return err, true
}

func HandleBodyDecodedValidation(w http.ResponseWriter, req *http.Request, transactionID string, err error) bool {
	if err != nil {
		log.ErrorR(req, fmt.Errorf("invalid request"))
		m := models.NewMessageResponse(fmt.Sprintf("failed to read request body for transaction %s", transactionID))
		WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
		return false
	}
	return true
}

func HandleEtagGenerationValidation(err error) bool {
	if err != nil {
		log.Error(fmt.Errorf("error generating etag: [%s]", err))
		return false
	}
	return true
}

func HandleCreateProgressReportResourceValidation(w http.ResponseWriter, req *http.Request, err error, statusCode int) bool {
	if err != nil {
		log.ErrorR(req, err)
		m := models.NewMessageResponse(err.Error())
		WriteJSONWithStatus(w, req, m, statusCode)
		return false
	}
	return true
}
