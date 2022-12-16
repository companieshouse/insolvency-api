package service

import (
	"fmt"
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
	"github.com/gorilla/mux"
	"net/http"
)

func TransactionIdExists(w http.ResponseWriter, req *http.Request) (string, bool) {
	vars := mux.Vars(req)
	transactionID := utils.GetTransactionIDFromVars(vars)
	if transactionID == "" {
		log.ErrorR(req, fmt.Errorf("there is no transaction ID in the URL path"))
		m := models.NewMessageResponse("transaction ID is not in the URL path")
		utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
		return "", false
	}
	return transactionID, true
}

func ValidateTransactionNotClosed(w http.ResponseWriter, req *http.Request, transactionID string) (error, bool) {
	// Check if transaction is closed
	isTransactionClosed, err, httpStatus := CheckIfTransactionClosed(transactionID, req)
	if err != nil {
		log.ErrorR(req, fmt.Errorf("error checking transaction status for [%v]: [%s]", transactionID, err))
		m := models.NewMessageResponse(fmt.Sprintf("error checking transaction status for [%v]: [%s]", transactionID, err))
		utils.WriteJSONWithStatus(w, req, m, httpStatus)
		return nil, false
	}

	if isTransactionClosed {
		log.ErrorR(req, fmt.Errorf("transaction [%v] is already closed and cannot be updated", transactionID))
		m := models.NewMessageResponse(fmt.Sprintf("transaction [%v] is already closed and cannot be updated", transactionID))
		utils.WriteJSONWithStatus(w, req, m, httpStatus)
		return nil, false
	}
	return err, true
}

func ValidBodyDecode(w http.ResponseWriter, req *http.Request, err error, transactionID string) bool {
	if err != nil {
		log.ErrorR(req, fmt.Errorf("invalid request"))
		m := models.NewMessageResponse(fmt.Sprintf("failed to read request body for transaction %s", transactionID))
		utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
		return false
	}
	return true
}
