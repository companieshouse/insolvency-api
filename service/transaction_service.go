package service

import (
	"fmt"
	"net/http"

	"github.com/companieshouse/go-sdk-manager/manager"
)

// CheckTransactionID will check with the transaction api that the provided transaction id exists
func CheckTransactionID(transactionID string, req *http.Request) (error, int) {

	// Create SDK session
	api, err := manager.GetPrivateSDK(req)
	if err != nil {
		return fmt.Errorf("error creating SDK to call transaction api: [%v]", err.Error()), http.StatusInternalServerError
	}

	// Call transaction api to retrieve details of the transaction
	transactionProfile, err := api.Transaction.Get(transactionID).Do()
	if err != nil {
		// If 404 then return the transaction not found
		if transactionProfile.HTTPStatusCode == http.StatusNotFound {
			return fmt.Errorf("transaction not found"), http.StatusNotFound
		}
		// Else return that there has been an error contacting the transaction api
		return fmt.Errorf("error communicating with the transaction api"), transactionProfile.HTTPStatusCode
	}

	return nil, http.StatusOK
}
