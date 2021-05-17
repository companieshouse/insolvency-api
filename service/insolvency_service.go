package service

import (
	"fmt"

	"github.com/companieshouse/insolvency-api/models"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/dao"
)

// ValidateInsolvencyDetails checks that an insolvency case is valid and ready for submission which is returned as a boolean
// Any validation errors found are added to an array to be returned
func ValidateInsolvencyDetails(svc dao.Service, transactionID string) (bool, *[]models.ValidationErrorResponseResource) {

	validationErrors := make([]models.ValidationErrorResponseResource, 0)

	// Retrieve details for the insolvency resource from MongoDB
	_, err := svc.GetInsolvencyResource(transactionID)
	if err != nil {
		log.Error(fmt.Errorf("error getting insolvency resource from DB [%s]", err))
		validationErrors = addValidationError(validationErrors, fmt.Sprintf("error getting insolvency resource from DB: [%s]", err), "insolvency case")

		// If there is an error retrieving the insolvency resource return without running any other validations as they will all fail
		return false, &validationErrors
	}

	//TODO: Validation checks for insolvency case

	return true, &validationErrors
}

// addValidationError adds any validation errors to an array of existing errors
func addValidationError(validationErrors []models.ValidationErrorResponseResource, validationError, errorLocation string) []models.ValidationErrorResponseResource {
	return append(validationErrors, *models.NewValidationErrorResponse(validationError, errorLocation))
}
