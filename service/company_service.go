package service

import (
	"fmt"
	"net/http"

	"github.com/companieshouse/insolvency-api/constants"

	"github.com/companieshouse/insolvency-api/models"

	"github.com/companieshouse/api-sdk-go/companieshouseapi"

	"github.com/companieshouse/go-sdk-manager/manager"
)

// CheckCompanyInsolvencyValid will check that the company is valid to be made insolvent against the company profile api
func CheckCompanyInsolvencyValid(insolvencyRequest *models.InsolvencyRequest, req *http.Request) (error, int) {

	// Create SDK session
	api, err := manager.GetSDK(req)
	if err != nil {
		return fmt.Errorf("error creating SDK to call company profile: [%v]", err.Error()), http.StatusInternalServerError
	}

	// Call company profile api to retrieve company details
	companyProfile, err := api.Profile.Get(insolvencyRequest.CompanyNumber).Do()
	if err != nil {
		// If 404 then return that company not found
		if companyProfile.HTTPStatusCode == http.StatusNotFound {
			return fmt.Errorf("company not found"), http.StatusNotFound
		}
		// Else there has been an error contacting the company profile api
		return fmt.Errorf("error communicating with the company profile api"), companyProfile.HTTPStatusCode
	}

	// check company is valid for insolvency
	if err := checkCompanyDetailsAreValid(companyProfile, insolvencyRequest); err != nil {
		return err, http.StatusBadRequest
	}

	// If no errors then the company is valid for insolvency
	return nil, http.StatusOK

}

// checkCompanyDetailsAreValid checks the incoming company profile to see if it's valid for insolvency
func checkCompanyDetailsAreValid(companyProfile *companieshouseapi.CompanyProfile, insolvencyRequest *models.InsolvencyRequest) error {

	// Check company name in request and company profile match
	if companyProfile.CompanyName != insolvencyRequest.CompanyName {
		return fmt.Errorf("company names do not match - provided: [%s], expected: [%s]", insolvencyRequest.CompanyName, companyProfile.CompanyName)
	}

	// Check if company jurisdiction is allowed
	if !checkJurisdictionIsAllowed(companyProfile.Jurisdiction) {
		return fmt.Errorf("jurisdiction [%s] not permitted", companyProfile.Jurisdiction)
	}

	// Check if company status is allowed
	if !checkCompanyStatusIsAllowed(companyProfile.CompanyStatus) {
		return fmt.Errorf("company status [%s] not permitted", companyProfile.CompanyStatus)
	}

	// Check is company type is allowed
	if !checkCompanyTypeIsAllowed(companyProfile.Type) {
		return fmt.Errorf("company type [%s] not permitted", companyProfile.Type)
	}

	return nil
}

// checkJurisdictionIsAllowed checks if the provided jurisdiction of the company is allowed
func checkJurisdictionIsAllowed(providedJurisdiction string) bool {
	for _, allowedJurisdiction := range constants.AllowedJurisdictions {
		if providedJurisdiction == allowedJurisdiction {
			return true
		}
	}
	return false
}

// checkCompanyStatusIsAllowed checks if the provided company status is allowed
func checkCompanyStatusIsAllowed(providedStatus string) bool {
	for _, forbiddenStatus := range constants.ForbiddenCompanyStatus {
		if providedStatus == forbiddenStatus {
			return false
		}
	}
	return true
}

// checkCompanyTypeIsAllowed checks if the provided company type is allowed
func checkCompanyTypeIsAllowed(providedType string) bool {
	for _, forbiddenType := range constants.ForbiddenCompanyTypes {
		if providedType == forbiddenType {
			return false
		}
	}
	return true
}
