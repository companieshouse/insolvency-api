package service

import (
	"fmt"
	"net/http"

	"github.com/companieshouse/go-sdk-manager/manager"
)

// IsUserOnEfsAllowList uses the sdk to call the EFS api and return a boolean depending on whether or not the email address is on the allow list
func IsUserOnEfsAllowList(emailAddress string, req *http.Request) (bool, error) {
	api, err := manager.GetInternalSDK(req)
	if err != nil {
		return false, fmt.Errorf("error creating SDK to call transaction api: [%v]", err.Error())
	}

	isUserAllowed, err := api.Efs.IsUserOnAllowList(emailAddress).Do()
	if err != nil {
		return false, fmt.Errorf("error communicating with the EFS submission api: [%s]", err)
	}

	if isUserAllowed == nil {
		return false, fmt.Errorf("error communicating with the EFS submission API: no response received")
	}

	return isUserAllowed.UserAllowed, nil
}
