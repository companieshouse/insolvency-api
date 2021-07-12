package service

import (
	"fmt"
	"net/http"

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
				validationError := fmt.Sprintf("error - all practitioners for insolvency case with transaction id [%s] must be appointed", insolvencyResource.TransactionID)
				log.Error(fmt.Errorf(validationError))
				validationErrors = addValidationError(validationErrors, validationError, "appointment")
				return false, &validationErrors
			}
		}
	}

	// Check if attachment type is "resolution", if not then at least one practitioner must be present
	hasResolutionAttachment := false
	for _, attachment := range insolvencyResource.Data.Attachments {
		if attachment.Type == "resolution" {
			hasResolutionAttachment = true
			break
		}
	}
	if !hasResolutionAttachment && len(insolvencyResource.Data.Attachments) != 0 {
		if len(insolvencyResource.Data.Practitioners) == 0 || insolvencyResource.Data.Practitioners == nil {
			validationError := fmt.Sprintf("error - attachment type requires that at least one practitioner must be present for insolvency case with transaction id [%s]", insolvencyResource.TransactionID)
			log.Error(fmt.Errorf(validationError))
			validationErrors = addValidationError(validationErrors, validationError, "resolution attachment type")
			return false, &validationErrors
		}
	}

	// Map attachment types
	attachmentTypes := map[string]struct{}{}
	for _, attachment := range insolvencyResource.Data.Attachments {
		attachmentTypes[attachment.Type] = struct{}{}
	}

	// Check if attachment type is statement-of-concurrence, if true then statement-of-affairs-director attachment must be present
	_, hasStatementOfConcurrence := attachmentTypes["statement-of-concurrence"]

	if hasStatementOfConcurrence {
		_, hasStatementOfAffairsDirector := attachmentTypes["statement-of-affairs-director"]
		if !hasStatementOfAffairsDirector {
			validationError := fmt.Sprintf("error - attachment statement-of-concurrence must be accompanied by statement-of-affairs-director attachment for insolvency case with transaction id [%s]", insolvencyResource.TransactionID)
			log.Error(fmt.Errorf(validationError))
			validationErrors = addValidationError(validationErrors, validationError, "statement of concurrence attachment type")
			return false, &validationErrors
		}
	}

	// Check if attachment type is statement-of-affairs-liquidator, if true then no practitioners must be appointed, but at least one should be present
	_, hasStateOfAffairsLiquidator := attachmentTypes["statement-of-affairs-liquidator"]

	if hasStateOfAffairsLiquidator && hasAppointedPractitioner {
		validationError := fmt.Sprintf("error - no appointed practitioners can be assigned to the case when attachment type statement-of-affairs-liquidator is included with transaction id [%s]", insolvencyResource.TransactionID)
		log.Error(fmt.Errorf(validationError))
		validationErrors = addValidationError(validationErrors, validationError, "statement of affairs liquidator attachment type")
		return false, &validationErrors
	}

	// Check if attachments are present, if false then at least one appointed practitioner must be present
	hasAttachments := true
	if len(insolvencyResource.Data.Attachments) == 0 {
		hasAttachments = false
	}
	if !hasAttachments && !hasAppointedPractitioner {
		validationError := fmt.Sprintf("error - at least one practitioner must be appointed as there are no attachments for insolvency case with transaction id [%s]", insolvencyResource.TransactionID)
		log.Error(fmt.Errorf(validationError))
		validationErrors = addValidationError(validationErrors, validationError, "no attachments")
		return false, &validationErrors
	}

	return true, &validationErrors
}

// addValidationError adds any validation errors to an array of existing errors
func addValidationError(validationErrors []models.ValidationErrorResponseResource, validationError, errorLocation string) []models.ValidationErrorResponseResource {
	return append(validationErrors, *models.NewValidationErrorResponse(validationError, errorLocation))
}

// ValidateAntivirus checks that attachments on an insolvency case pass the antivirus check and are ready for submission which is returned as a boolean
// Any validation errors found are added to an array to be returned
func ValidateAntivirus(svc dao.Service, transactionID string, req *http.Request) (bool, *[]models.ValidationErrorResponseResource) {

	validationErrors := make([]models.ValidationErrorResponseResource, 0)

	// Retrieve details for the insolvency resource from DB
	insolvencyResource, err := svc.GetInsolvencyResource(transactionID)
	if err != nil {
		log.Error(fmt.Errorf("error getting insolvency resource from DB [%s]", err))
		validationErrors = addValidationError(validationErrors, fmt.Sprintf("error getting insolvency resource from DB: [%s]", err), "insolvency case")

		// If there is an error retrieving the insolvency resource return without running any other validation as it will fail
		return false, &validationErrors
	}
	// Check if the insolvency resource has attachments, if not then skip validation
	if len(insolvencyResource.Data.Attachments) != 0 {

		AvStatuses := map[string]struct{}{}
		// Check the antivirus status of each attachment type and update with the appropriate status in mongodb
		for _, attachment := range insolvencyResource.Data.Attachments {
			// Calls File Transfer API to get attachment details
			attachmentDetailsResponse, responseType, err := GetAttachmentDetails(attachment.ID, req)
			if err != nil {
				log.ErrorR(req, fmt.Errorf("error getting attachment details: [%v]", err), log.Data{"service_response_type": responseType.String()})
			}
			// If antivirus check has not passed, update insolvency resource with "integrity_failed" status
			if attachmentDetailsResponse.AVStatus != "clean" {
				svc.UpdateAttachmentStatus(insolvencyResource.TransactionID, attachment.ID, "integrity_failed")
				AvStatuses[attachmentDetailsResponse.AVStatus] = struct{}{}
				continue
			}
			// If antivirus has passed, update insolvency resource with "processed" status
			svc.UpdateAttachmentStatus(insolvencyResource.TransactionID, attachment.ID, "processed")
			AvStatuses[attachmentDetailsResponse.AVStatus] = struct{}{}
		}
		// Check AvStatuses map to see if status "not-scanned" exists
		_, attachmentNotScanned := AvStatuses["not-scanned"]
		if attachmentNotScanned {
			validationError := fmt.Sprintf("error - antivirus check has failed on insolvency case with transaction id [%s], attachments have not been scanned", insolvencyResource.TransactionID)
			log.Error(fmt.Errorf(validationError))
			validationErrors = addValidationError(validationErrors, validationError, "antivirus incomplete")
			return false, &validationErrors
		}
		// Check AvStatuses map to see if status "infected" exists
		_, attachmentInfected := AvStatuses["infected"]
		if attachmentInfected {
			validationError := fmt.Sprintf("error - antivirus check has failed on insolvency case with transaction id [%s], virus detected", insolvencyResource.TransactionID)
			log.Error(fmt.Errorf(validationError))
			validationErrors = addValidationError(validationErrors, validationError, "antivirus failure")
			return false, &validationErrors
		}
	}

	return true, &validationErrors
}

// GenerateFilings generates an array of filings for this insolvency resource to be used by the filing resource handler
func GenerateFilings(svc dao.Service, transactionID string) ([]models.Filing, error) {

	// Retrieve details for the insolvency resource from DB
	insolvencyResource, err := svc.GetInsolvencyResource(transactionID)
	if err != nil {
		message := fmt.Errorf("error getting insolvency resource from DB [%s]", err)
		log.Error(message)
		return nil, message
	}

	var filings []models.Filing

	// Check for an appointed practitioner to determine if there's a 600 insolvency form
	for _, practitioner := range insolvencyResource.Data.Practitioners {
		if practitioner.Appointment != nil {
			// If a filing is a 600 add a generated filing to the array of filings
			filing600 := models.NewFiling(
				generateDataBlockForFiling(&insolvencyResource, "600"),
				fmt.Sprintf("600 insolvency case for %v", insolvencyResource.Data.CompanyNumber),
				"600",
				"insolvency#600")
			filings = append(filings, *filing600)
			break
		}
	}

	// Map attachment types
	attachmentTypes := map[string]struct{}{}
	for _, attachment := range insolvencyResource.Data.Attachments {
		attachmentTypes[attachment.Type] = struct{}{}
	}

	// Check if a "resolution" attachment type is present to determine if there's a LRESEX form
	if _, hasResolution := attachmentTypes["resolution"]; hasResolution {
		// If a filing is an LRESEX add a generated filing to the array of filings
		filingLRESEX := models.NewFiling(
			generateDataBlockForFiling(&insolvencyResource, "LRESEX"),
			fmt.Sprintf("LRESEX insolvency case for %v", insolvencyResource.Data.CompanyNumber),
			"LRESEX",
			"insolvency#LRESEX")
		filings = append(filings, *filingLRESEX)
	}

	// Check if a "statement-of-affairs-director" or "statement-of-affairs-liquidator" is present to determine if there's a LIQ02 form
	_, hasStatementOfAffairsDirector := attachmentTypes["statement-of-affairs-director"]
	_, hasStatementOfAffairsLiquidator := attachmentTypes["statement-of-affairs-liquidator"]
	if hasStatementOfAffairsDirector || hasStatementOfAffairsLiquidator {
		// If a filing is a LIQ02 add a generated filing to the array of filings
		filingLIQ02 := models.NewFiling(
			generateDataBlockForFiling(&insolvencyResource, "LIQ02"),
			fmt.Sprintf("LIQ02 insolvency case for %v", insolvencyResource.Data.CompanyNumber),
			"LIQ02",
			"insolvency#LIQ02")
		filings = append(filings, *filingLIQ02)
	}

	return filings, nil
}

// generateDataBlockForFiling generates the block of data to be included with a filing
func generateDataBlockForFiling(insolvencyResource *models.InsolvencyResourceDao, form string) map[string]interface{} {

	// If the form is a 600 do not include attachment details, otherwise do include them
	if form == "600" {
		return map[string]interface{}{
			"company_number": &insolvencyResource.Data.CompanyNumber,
			"case_type":      &insolvencyResource.Data.CaseType,
			"company_name":   &insolvencyResource.Data.CompanyName,
			"practitioners":  &insolvencyResource.Data.Practitioners,
		}
	} else if form == "LRESEX" || form == "LIQ02" {
		return map[string]interface{}{
			"company_number": &insolvencyResource.Data.CompanyNumber,
			"case_type":      &insolvencyResource.Data.CaseType,
			"company_name":   &insolvencyResource.Data.CompanyName,
			"practitioners":  &insolvencyResource.Data.Practitioners,
			"attachments":    &insolvencyResource.Data.Attachments,
		}
	}

	return nil
}
