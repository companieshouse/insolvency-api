package service

import (
	"fmt"
	"net/http"

	"github.com/companieshouse/go-sdk-manager/manager"
)

// IsUserOnEfsAllowList uses the sdk to call the EFS api and return a boolean depending on whether or not the email address is on the allow list
func IsUserOnEfsAllowList(emailAddress string, req *http.Request) (bool, error) {
	// Create Private SDK session
	api, err := manager.GetPrivateSDK(req)
	if err != nil {
		return false, fmt.Errorf("error creating SDK to call transaction api: [%v]", err.Error())
	}

	// Patch transaction api with insolvency resource
	isUserAllowed, err := api.Efs.IsUserOnAllowList(emailAddress).Do()
	if err != nil {
		// Return that there has been an error contacting the transaction api
		return isUserAllowed.IsUserOnEfsAllowList, fmt.Errorf("error communicating with the EFS submission api")
	}

	return isUserAllowed.IsUserOnEfsAllowList, nil

}