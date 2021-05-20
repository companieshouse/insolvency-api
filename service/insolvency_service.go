package service

import (
	"fmt"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/dao"
	"github.com/companieshouse/insolvency-api/models"
)

// ValidateInsolvencyDetails checks that an insolvency case is valid and ready for submission which is returned as a boolean
// Any validation errors found are added to an array to be returned
func ValidateInsolvencyDetails(svc dao.Service, transactionID string) (bool, *[]models.ValidationErrorResponseResource) {

	validationErrors := make([]models.ValidationErrorResponseResource, 0)

	// Retrieve details for the insolvency resource from DB
	insolvencyResource, err := svc.GetInsolvencyResource(transactionID)
	if err != nil {
		log.Error(fmt.Errorf("error getting insolvency resource from DB [%s]", err))
		validationErrors = addValidationError(validationErrors, fmt.Sprintf("error getting insolvency resource from DB: [%s]", err), "insolvency case")

		// If there is an error retrieving the insolvency resource return without running any other validations as they will all fail
		return false, &validationErrors
	}

	// Check if there is one practitioner appointed and if there is, ensure that all practitioners are appointed
	hasAppointedPractitioner := false
	for _, practitioner := range insolvencyResource.Data.Practitioners {
		if practitioner.Appointment != nil {
			hasAppointedPractitioner = true
			break
		}
	}
	if hasAppointedPractitioner {
		for _, practitioner := range insolvencyResource.Data.Practitioners {
			if practitioner.Appointment == nil || practitioner.Appointment.AppointedOn == "" {
				validationError := fmt.Sprintf("error - all practitioners for insolvency case with transaction id [%s] have been not appointed", insolvencyResource.TransactionID)
				log.Error(fmt.Errorf(validationError))
				validationErrors = addValidationError(validationErrors, validationError, "appointment")
				return false, &validationErrors
			}
		}
	}

	return true, &validationErrors
}

// addValidationError adds any validation errors to an array of existing errors
func addValidationError(validationErrors []models.ValidationErrorResponseResource, validationError, errorLocation string) []models.ValidationErrorResponseResource {
	return append(validationErrors, *models.NewValidationErrorResponse(validationError, errorLocation))
}
