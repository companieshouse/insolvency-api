package service

import (
	"fmt"
	"net/http"

	"github.com/companieshouse/api-sdk-go/companieshouseapi"
	"github.com/companieshouse/go-sdk-manager/manager"
	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/models"
)

// CheckCompanyExists will check that the company exists against the company profile api to make a valid insolvency
func CheckCompanyExists(insolvencyRequest *models.InsolvencyRequest, req *http.Request) (error, int, *companieshouseapi.CompanyProfile) {

	// Create SDK session
	api, err := manager.GetSDK(req)
	if err != nil {
		return fmt.Errorf("error creating SDK to call company profile: [%v]", err.Error()), http.StatusInternalServerError, nil
	}

	// Call company profile api to retrieve company details
	companyProfile, err := api.Profile.Get(insolvencyRequest.CompanyNumber).Do()
	if err != nil {
		// If 404 then return that company not found
		if companyProfile.HTTPStatusCode == http.StatusNotFound {
			return fmt.Errorf("company not found"), http.StatusNotFound, nil
		}
		// Else there has been an error contacting the company profile api
		return fmt.Errorf("error communicating with the company profile api"), companyProfile.HTTPStatusCode, nil
	}

	// If no errors then the company exists
	return nil, companyProfile.HTTPStatusCode, companyProfile

}

// GetCompanyIncorporatedOn retrieves the date that the company was created
func GetCompanyIncorporatedOn(companyNumber string, req *http.Request) (string, error) {
	// Create SDK session
	api, err := manager.GetSDK(req)
	if err != nil {
		return "", fmt.Errorf("error creating SDK to call company profile: [%v]", err.Error())
	}

	// Call company profile api to retrieve company details
	companyProfile, err := api.Profile.Get(companyNumber).Do()
	if err != nil {
		// If 404 then return that company not found
		if companyProfile.HTTPStatusCode == http.StatusNotFound {
			return "", fmt.Errorf("company not found")
		}
		// Else there has been an error contacting the company profile api
		return "", fmt.Errorf("error communicating with the company profile api")
	}

	return companyProfile.DateOfCreation, nil
}

// checkCompanyDetailsAreValid checks the incoming company profile to see if it's valid for insolvency
func CheckCompanyDetailsAreValid(companyProfile *companieshouseapi.CompanyProfile, insolvencyRequest *models.InsolvencyRequest) error {

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
