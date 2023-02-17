package service

import (
	"fmt"
	"net/http"

	"github.com/companieshouse/go-sdk-manager/manager"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/transformers"
)

// CheckTransactionID will check with the transaction api that the provided transaction id exists
func CheckTransactionID(transactionID string, req *http.Request) (error, int) {

	// Create SDK session
	api, err := manager.GetSDK(req)

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

	return nil, transactionProfile.HTTPStatusCode
}

// PatchTransactionWithInsolvencyResource will patch the provided transaction with the created insolvency resource
func PatchTransactionWithInsolvencyResource(transactionID string, insolvencyResource *models.InsolvencyResourceDto, req *http.Request) (error, int) {

	// Create Private SDK session
	api, err := manager.GetPrivateSDK(req)

	if err != nil {
		return fmt.Errorf("error creating SDK to call transaction api: [%v]", err.Error()), http.StatusInternalServerError
	}

	// Patch transaction api with insolvency resource
	transactionProfile, err := api.Transaction.Patch(transactionID, transformers.InsolvencyResourceDaoToTransactionResource(insolvencyResource)).Do()

	if err != nil {
		// If 404 then return the transaction not found
		if transactionProfile.HTTPStatusCode == http.StatusNotFound {
			return fmt.Errorf("transaction not found"), http.StatusNotFound
		}
		// Else return that there has been an error contacting the transaction api
		return fmt.Errorf("error communication with the transaction api"), transactionProfile.HTTPStatusCode
	}

	return nil, transactionProfile.HTTPStatusCode
}

// CheckIfTransactionClosed checks against the transaction api if the transaction is closed or not
func CheckIfTransactionClosed(transactionID string, req *http.Request) (bool, error, int) {

	// Create SDK session
	api, err := manager.GetSDK(req)

	if err != nil {
		return false, fmt.Errorf("error creating SDK to call transaction api: [%v]", err.Error()), http.StatusInternalServerError
	}

	// Call transaction api to retrieve details of the transaction
	transactionProfile, err := api.Transaction.Get(transactionID).Do()

	if err != nil {
		// If 404 then return the transaction not found
		if transactionProfile.HTTPStatusCode == http.StatusNotFound {
			return false, fmt.Errorf("transaction not found"), http.StatusNotFound
		}
		// Else return that there has been an error contacting the transaction api
		return false, fmt.Errorf("error getting transaction from transaction api: [%v]", err.Error()), transactionProfile.HTTPStatusCode
	}

	if transactionProfile.Status == "closed" {
		return true, nil, http.StatusForbidden
	}

	return false, nil, http.StatusOK
}
