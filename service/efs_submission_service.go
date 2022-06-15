package service

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/go-sdk-manager/manager"
	"github.com/companieshouse/insolvency-api/config"
)

// IsUserOnEfsAllowList uses the sdk to call the EFS api and return a boolean depending on whether or not the email
// address is on the allow list
func IsUserOnEfsAllowList(emailAddress string, req *http.Request) (bool, error) {
	api, err := manager.GetPrivateSDK(req)
	if err != nil {
		return false, fmt.Errorf("error creating private SDK to call transaction api: [%v]", err.Error())
	}

	// Get environment config - only required whilst feature flag to disable EFS lookup exists
	cfg, err := config.Get()
	if err != nil {
		return false, fmt.Errorf("error configuring service: %w. Exiting", err)
	}

	// Check from Env Var or Command Line Flag if EFS Allow List Auth has been disabled AND email address contains
	// 'magic string' in which case the API call is bypassed and a 'true' value is returned to parent
	if cfg.IsEfsAllowListAuthDisabled {
		// Our 'magic string' to bypass EFS Allow List if it is in email address is 'ip-test'
		isMatch, err := regexp.MatchString("ip-test", emailAddress)
		if err != nil {
			return false, fmt.Errorf("EFS Allow List API call disabled by environment variable, but unable to check email address for regex match")
		}
		log.InfoR(req, fmt.Sprintf("EFS Allow List API call disabled by environment variable for email address: %s. Mocked API response: %t", emailAddress, isMatch))
		return isMatch, nil
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
