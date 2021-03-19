package service

import (
	"fmt"
	"net/http"

	"github.com/companieshouse/api-sdk-go/companieshouseapi"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/go-sdk-manager/manager"
)

// CheckCompanyInsolvencyValid will check that the company is valid to be made insolvent against the company profile api
func CheckCompanyInsolvencyValid(companyNumber string, req *http.Request) (bool, error, int) {

	// Create SDK session
	api, err := manager.GetSDK(req)
	if err != nil {
		log.ErrorR(req, err, log.Data{"company_number": companyNumber})
		return false, fmt.Errorf("error creating SDK to call company profile: [%v]", err), http.StatusInternalServerError
	}

	// Call company profile api to retrieve company details
	companyProfile, err := api.Profile.Get(companyNumber).Do()
	if err != nil {
		log.ErrorR(req, err, log.Data{"company_number": companyNumber})
		return false, fmt.Errorf("error calling company profile: [%v]", err), http.StatusInternalServerError
	}

	// If an error is returned when checking the company details the company is not valid
	if err := checkCompanyDetailsAreValid(companyProfile); err != nil {
		return false, err, http.StatusBadRequest
	}

	// If no errors then the company is valid for insolvency
	return true, nil, http.StatusCreated

}

// checkCompanyDetailsAreValid checks the incoming company profile to see if it's valid for insolvency
func checkCompanyDetailsAreValid(profile *companieshouseapi.CompanyProfile) error {
	//TODO: Check is company is valid for insolvency
	return nil
}
