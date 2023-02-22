package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/dao"
	"github.com/companieshouse/insolvency-api/models"
)

// layout for parsing dates
const dateLayout = "2006-01-02"
const validationMessageFormat = "validation failed for insolvency ID [%s]: [%v]"

// ValidateInsolvencyDetails checks that an insolvency case is valid and ready for submission
// Any validation errors found are added to an array to be returned
func ValidateInsolvencyDetails(insolvencyResource models.InsolvencyResourceDao) *[]models.ValidationErrorResponseResource {

	validationErrors := make([]models.ValidationErrorResponseResource, 0)

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
				log.Info(validationError)
				validationErrors = addValidationError(validationErrors, validationError, "appointment")
			}
		}
	}

	hasSubmittedPractitioner := insolvencyResource.Data.Practitioners != nil && len(insolvencyResource.Data.Practitioners) > 0

	// Check if attachment type is "resolution", if not then at least one practitioner must be present
	hasResolutionAttachment := false
	resolutionArrayPosition := 0

	for i, attachment := range insolvencyResource.Data.Attachments {
		if attachment.Type == "resolution" {
			hasResolutionAttachment = true
			resolutionArrayPosition = i
			break
		}
	}

	if !hasResolutionAttachment && len(insolvencyResource.Data.Attachments) != 0 {
		if len(insolvencyResource.Data.Practitioners) == 0 || insolvencyResource.Data.Practitioners == nil {
			validationError := fmt.Sprintf("error - attachment type requires that at least one practitioner must be present for insolvency case with transaction id [%s]", insolvencyResource.TransactionID)
			log.Info(validationError)
			validationErrors = addValidationError(validationErrors, validationError, "resolution attachment type")
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
			log.Info(validationError)
			validationErrors = addValidationError(validationErrors, validationError, "statement of concurrence attachment type")
		}
	}

	// Check if attachment type is statement-of-affairs-liquidator, if true then no practitioners must be appointed, but at least one should be present
	_, hasStateOfAffairsLiquidator := attachmentTypes["statement-of-affairs-liquidator"]

	if hasStateOfAffairsLiquidator && hasAppointedPractitioner {
		validationError := fmt.Sprintf("error - no appointed practitioners can be assigned to the case when attachment type statement-of-affairs-liquidator is included with transaction id [%s]", insolvencyResource.TransactionID)
		log.Info(validationError)
		validationErrors = addValidationError(validationErrors, validationError, "statement of affairs liquidator attachment type")
	}

	// Check if attachments are present, if false then at least one appointed practitioner must be present
	hasAttachments := true
	if len(insolvencyResource.Data.Attachments) == 0 {
		hasAttachments = false
	}

	// Check if a resolution has been filed against the insolvency case
	resolutionFiled := false
	if insolvencyResource.Data.Resolution != nil {
		resolutionFiled = true
	}

	// Check if attachment_type is resolution, if true then date_of_resolution must be present
	if hasResolutionAttachment && (!resolutionFiled || (resolutionFiled && insolvencyResource.Data.Resolution.DateOfResolution == "")) {
		validationError := fmt.Sprintf("error - a date of resolution must be present as there is an attachment with type resolution for insolvency case with transaction id [%s]", insolvencyResource.TransactionID)
		log.Info(validationError)
		validationErrors = addValidationError(validationErrors, validationError, "no date of resolution")
	}

	// Check if date_of_resolution is present, then resolution attachment must be present
	if resolutionFiled && insolvencyResource.Data.Resolution.DateOfResolution != "" && !hasResolutionAttachment {
		validationError := fmt.Sprintf("error - a resolution attachment must be present as there is a date_of_resolution filed for insolvency case with transaction id [%s]", insolvencyResource.TransactionID)
		log.Info(validationError)
		validationErrors = addValidationError(validationErrors, validationError, "no resolution")
	}

	// Check that id of uploaded resolution attachment matches attachment id supplied in resolution
	if hasResolutionAttachment && resolutionFiled && (insolvencyResource.Data.Attachments[resolutionArrayPosition].ID != insolvencyResource.Data.Resolution.Attachments[0]) {
		validationError := fmt.Sprintf("error - id for uploaded resolution attachment must match the attachment id supplied when filing a resolution for insolvency case with transaction id [%s]", insolvencyResource.TransactionID)
		log.Info(validationError)
		validationErrors = addValidationError(validationErrors, validationError, "attachment ids do not match")
	}

	// Check if "statement-of-affairs-director" or "statement-of-affairs-liquidator" has been filed, if so, then a statement date must be present
	_, hasStatementOfAffairsDirector := attachmentTypes[constants.StatementOfAffairsDirector.String()]
	if (hasStatementOfAffairsDirector || hasStateOfAffairsLiquidator) && (insolvencyResource.Data.StatementOfAffairs == nil || insolvencyResource.Data.StatementOfAffairs.StatementDate == "") {
		validationError := fmt.Sprintf("error - a date of statement of affairs must be present as there is an attachment with type [%s] or [%s] for insolvency case with transaction id [%s]", constants.StatementOfAffairsDirector.String(), constants.StatementOfAffairsLiquidator.String(), insolvencyResource.TransactionID)
		log.Info(validationError)
		validationErrors = addValidationError(validationErrors, validationError, "statement-of-affairs")
	}

	// Check if SOA resource exists or statement date is not empty in DB, if not, then an SOA-D or SOA-L attachment must be filed
	if insolvencyResource.Data.StatementOfAffairs != nil && insolvencyResource.Data.StatementOfAffairs.StatementDate != "" && !(hasStatementOfAffairsDirector || hasStateOfAffairsLiquidator) {
		validationError := fmt.Sprintf("error - an attachment of type [%s] or [%s] must be present as there is a date of statement of affairs present for insolvency case with transaction id [%s]", constants.StatementOfAffairsDirector.String(), constants.StatementOfAffairsLiquidator.String(), insolvencyResource.TransactionID)
		log.Info(validationError)
		validationErrors = addValidationError(validationErrors, validationError, "statement-of-affairs")
	}

	if !hasAttachments && hasSubmittedPractitioner && !hasAppointedPractitioner {
		validationError := fmt.Sprintf("error - at least one practitioner must be appointed as there are no attachments for insolvency case with transaction id [%s]", insolvencyResource.TransactionID)
		log.Info(validationError)
		validationErrors = addValidationError(validationErrors, validationError, "no attachments")
	}

	if !hasSubmittedPractitioner && !hasResolutionAttachment {
		validationError := "error - if no practitioners are present then an attachment of the type resolution must be present"
		log.Info(fmt.Sprintf(validationMessageFormat, insolvencyResource.ID, validationError))
		validationErrors = addValidationError(validationErrors, validationError, "no practitioners and no resolution")
	}

	// Check if case has appointed practitioner and resolution attached
	// If there is, the practitioner appointed date must be the same
	// or after resolution date
	if hasAppointedPractitioner && hasResolutionAttachment {
		for _, practitioner := range insolvencyResource.Data.Practitioners {
			ok, err := checkValidAppointmentDate(practitioner.Appointment.AppointedOn, insolvencyResource.Data.Resolution.DateOfResolution)
			if err != nil {
				log.Error(fmt.Errorf("error when parsing date for insolvency ID [%s]: [%s]", insolvencyResource.ID, err))
				validationErrors = addValidationError(validationErrors, fmt.Sprint(err), "practitioner")
			}

			if !ok {
				validationError := fmt.Sprintf("error - practitioner [%s] appointed on [%s] is before the resolution date [%s]", practitioner.ID, practitioner.Appointment.AppointedOn, insolvencyResource.Data.Resolution.DateOfResolution)
				validationErrors = addValidationError(validationErrors, validationError, "practitioner")
			}
		}
	}

	// If both Statement Of Affairs Date and Resolution Date provided, validate against each other
	hasStatementOfAffairsDate := insolvencyResource.Data.StatementOfAffairs != nil && insolvencyResource.Data.StatementOfAffairs.StatementDate != ""
	hasResolutionDate := insolvencyResource.Data.Resolution != nil && insolvencyResource.Data.Resolution.DateOfResolution != ""
	if hasStatementOfAffairsDate && hasResolutionDate {
		ok, reason, errLocation, err := checkValidStatementOfAffairsDate(insolvencyResource.Data.StatementOfAffairs.StatementDate, insolvencyResource.Data.Resolution.DateOfResolution)
		if err != nil {
			log.Error(fmt.Errorf("error checking dates: %s", err))
			validationErrors = addValidationError(validationErrors, fmt.Sprint(err), errLocation)
		}
		if !ok {
			validationErrors = addValidationError(validationErrors, reason, errLocation)
		}
	}

	// If a Progress Report has been submitted then check that the from/to dates have been submitted
	_, hasProgressReport := attachmentTypes[constants.ProgressReport.String()]
	if hasProgressReport {
		if insolvencyResource.Data.ProgressReport.FromDate == "" || insolvencyResource.Data.ProgressReport.ToDate == "" {
			validationError := fmt.Sprintf("error - progress report dates must be present as there is an attachment with type progress-report for insolvency case with transaction id [%s]", insolvencyResource.TransactionID)
			log.Info(validationError)
			validationErrors = addValidationError(validationErrors, validationError, "no dates for progress report")
		}
	}

	return &validationErrors
}

// checkValidAppointmentData parses and checks if the appointment date is on or after the dateOfResolution
func checkValidAppointmentDate(appointedOn string, dateOfResolution string) (bool, error) {

	// Parse appointedOn to time
	appointmentDate, err := time.Parse(dateLayout, appointedOn)
	if err != nil {
		return false, err
	}

	// Parse dateOfResolution to time
	resolutionDate, err := time.Parse(dateLayout, dateOfResolution)
	if err != nil {
		return false, err
	}

	// If appointmentOn is before dateOfResolution then return false
	if appointmentDate.Before(resolutionDate) {
		return false, nil
	}

	return true, nil
}

func checkValidStatementOfAffairsDate(statementOfAffairsDate string, resolutionDate string) (bool, string, string, error) {
	soaDate, err := time.Parse(dateLayout, statementOfAffairsDate)
	if err != nil {
		return false, "", "statement of affairs date", fmt.Errorf("invalid statementOfAffairs date [%s]", statementOfAffairsDate)
	}

	resDate, err := time.Parse(dateLayout, resolutionDate)
	if err != nil {
		return false, "", "resolution date", fmt.Errorf("invalid resolution date [%s]", resolutionDate)
	}

	// Statement Of Affairs Date must not be before Resolution Date
	if soaDate.Before(resDate) {
		return false, "error - statement of affairs date must not be before resolution date", "", nil
	}

	// Statement Of Affairs Date must not be more than 7 days after resolution date
	if soaDate.Sub(resDate).Hours()/24 > 7 {
		return false, "error - statement of affairs date must be within 7 days of resolution date", "", nil
	}

	return true, "", "", nil
}

// addValidationError adds any validation errors to an array of existing errors
func addValidationError(validationErrors []models.ValidationErrorResponseResource, validationError, errorLocation string) []models.ValidationErrorResponseResource {
	return append(validationErrors, *models.NewValidationErrorResponse(validationError, errorLocation))
}

// ValidateAntivirus checks that attachments on an insolvency case pass the antivirus check and are ready for submission
// Any validation errors found are added to an array to be returned
func ValidateAntivirus(svc dao.Service, insolvencyResource models.InsolvencyResourceDao, req *http.Request) *[]models.ValidationErrorResponseResource {

	validationErrors := make([]models.ValidationErrorResponseResource, 0)

	// Check if the insolvency resource has attachments, if not then skip validation
	if len(insolvencyResource.Data.Attachments) != 0 {

		avStatuses := map[string]struct{}{}
		// Check the antivirus status of each attachment type and update with the appropriate status in mongodb
		for _, attachment := range insolvencyResource.Data.Attachments {
			// Calls File Transfer API to get attachment details
			attachmentDetailsResponse, responseType, err := GetAttachmentDetails(attachment.ID, req)
			if err != nil {
				log.ErrorR(req, fmt.Errorf("error getting attachment details for attachment ID [%s]: [%v]", attachment.ID, err), log.Data{"service_response_type": responseType.String()})
			}

			// If antivirus check has not passed, update insolvency resource with "integrity_failed" status
			if attachmentDetailsResponse.AVStatus != "clean" {
				svc.UpdateAttachmentStatus(insolvencyResource.TransactionID, attachment.ID, "integrity_failed")
				avStatuses[attachmentDetailsResponse.AVStatus] = struct{}{}
				continue
			}
			// If antivirus has passed, update insolvency resource with "processed" status
			svc.UpdateAttachmentStatus(insolvencyResource.TransactionID, attachment.ID, "processed")
			avStatuses[attachmentDetailsResponse.AVStatus] = struct{}{}
		}
		// Check avStatuses map to see if status "not-scanned" exists
		_, attachmentNotScanned := avStatuses["not-scanned"]
		if attachmentNotScanned {
			validationError := fmt.Sprintf("error - antivirus check has failed on insolvency case with transaction id [%s], attachments have not been scanned", insolvencyResource.TransactionID)
			log.Info(fmt.Sprintf(validationMessageFormat, insolvencyResource.ID, validationError))
			validationErrors = addValidationError(validationErrors, validationError, "antivirus incomplete")
		}
		// Check avStatuses map to see if status "infected" exists
		_, attachmentInfected := avStatuses["infected"]
		if attachmentInfected {
			validationError := fmt.Sprintf("error - antivirus check has failed on insolvency case with transaction id [%s], virus detected", insolvencyResource.TransactionID)
			log.Info(fmt.Sprintf(validationMessageFormat, insolvencyResource.ID, validationError))
			validationErrors = addValidationError(validationErrors, validationError, "antivirus failure")
		}
	}

	return &validationErrors
}

// GenerateFilings generates an array of filings for this insolvency resource to be used by the filing resource handler
func GenerateFilings(svc dao.Service, transactionID string) ([]models.Filing, error) {

	// Retrieve details for the insolvency resource from DB
	insolvencyResource, err := svc.GetInsolvencyResource(transactionID)
	if err != nil {
		message := fmt.Errorf("error getting insolvency resource from DB [%s]", err)
		return nil, message
	}

	var filings []models.Filing

	// Check for an appointed practitioner to determine if there's a 600 insolvency form
	for _, practitioner := range insolvencyResource.Data.Practitioners {
		if practitioner.Appointment != nil {
			// If a filing is a 600 add a generated filing to the array of filings
			newFiling := generateNewFiling(&insolvencyResource, nil, "600")
			filings = append(filings, *newFiling)
			break
		}
	}

	// Map attachments to filing types
	attachmentsLRESEX := []*models.AttachmentResourceDao{}
	attachmentsLIQ02 := []*models.AttachmentResourceDao{}
	attachmentsLIQ03 := []*models.AttachmentResourceDao{}
	// using range index to allow passing reference not value
	for i := range insolvencyResource.Data.Attachments {
		switch insolvencyResource.Data.Attachments[i].Type {
		case "resolution":
			attachmentsLRESEX = append(attachmentsLRESEX, &insolvencyResource.Data.Attachments[i])
		case "statement-of-affairs-director", "statement-of-affairs-liquidator", "statement-of-concurrence":
			attachmentsLIQ02 = append(attachmentsLIQ02, &insolvencyResource.Data.Attachments[i])
		case "progress-report":
			attachmentsLIQ03 = append(attachmentsLIQ03, &insolvencyResource.Data.Attachments[i])
		}
	}
	if len(attachmentsLRESEX) > 0 {
		newFiling := generateNewFiling(&insolvencyResource, attachmentsLRESEX, "LRESEX")
		filings = append(filings, *newFiling)
	}
	if len(attachmentsLIQ02) > 0 {
		newFiling := generateNewFiling(&insolvencyResource, attachmentsLIQ02, "LIQ02")
		filings = append(filings, *newFiling)
	}
	if len(attachmentsLIQ03) > 0 {
		newFiling := generateNewFiling(&insolvencyResource, attachmentsLIQ03, "LIQ03")
		filings = append(filings, *newFiling)
	}
	return filings, nil
}

// generateDataBlockForFiling generates the block of data to be included with a filing
func generateNewFiling(insolvencyResource *models.InsolvencyResourceDao, attachments []*models.AttachmentResourceDao, filingType string) *models.Filing {

	dataBlock := map[string]interface{}{
		"company_number": &insolvencyResource.Data.CompanyNumber,
		"case_type":      &insolvencyResource.Data.CaseType,
		"company_name":   &insolvencyResource.Data.CompanyName,
		"practitioners":  &insolvencyResource.Data.Practitioners,
	}

	switch filingType {
	case "LRESEX":
		if insolvencyResource.Data.Resolution != nil {
			dataBlock["case_date"] = &insolvencyResource.Data.Resolution.DateOfResolution
		}
		delete(dataBlock, "practitioners")
	case "LIQ02":
		if insolvencyResource.Data.StatementOfAffairs != nil {
			dataBlock["soa_date"] = &insolvencyResource.Data.StatementOfAffairs.StatementDate
		}
	case "LIQ03":
		if insolvencyResource.Data.ProgressReport != nil {
			dataBlock["from_date"] = &insolvencyResource.Data.ProgressReport.FromDate
			dataBlock["to_date"] = &insolvencyResource.Data.ProgressReport.ToDate
		}
	}
	if attachments != nil {
		dataBlock["attachments"] = attachments
	}

	newFiling := models.NewFiling(
		dataBlock,
		fmt.Sprintf("%s insolvency case for %v", filingType, insolvencyResource.Data.CompanyNumber),
		filingType,
		fmt.Sprintf("insolvency#%s", filingType))
	return newFiling
}
