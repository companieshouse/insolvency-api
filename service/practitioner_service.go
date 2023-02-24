package service

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/dao"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
)

// ValidatePractitionerDetails checks that the incoming practitioner details are valid
func ValidatePractitionerDetails(svc dao.Service, transactionID string, practitioner models.PractitionerRequest) (string, error) {
	var errs []string

	// Check that either the telephone number or email field are populated
	if practitioner.TelephoneNumber == "" && practitioner.Email == "" {
		errs = append(errs, "either telephone_number or email are required")
	}

	// Set allowed regexp for telephone number
	telephoneNumberRuleRegexString := `^\d+$`
	telephoneNumberRegex := regexp.MustCompile(telephoneNumberRuleRegexString)

	// Check that telephone number starts with 0 and only contains digits
	if practitioner.TelephoneNumber != "" && (!strings.HasPrefix(practitioner.TelephoneNumber, "0") || !telephoneNumberRegex.MatchString(practitioner.TelephoneNumber)) {
		errs = append(errs, "telephone_number must start with 0 and contain only numeric characters")
	}

	// Check that telephone number is the correct length
	if practitioner.TelephoneNumber != "" && !((len(practitioner.TelephoneNumber) == 10) || (len(practitioner.TelephoneNumber) == 11)) {
		errs = append(errs, "telephone_number must be 10 or 11 digits long")
	}

	// Check that telephone number does not contain spaces
	if practitioner.TelephoneNumber != "" && strings.Contains(practitioner.TelephoneNumber, " ") {
		errs = append(errs, "telephone_number must not contain spaces")
	}

	// Set allowed naming conventions for names
	nameRuleRegexString := `^[a-zA-ZàáâäãåąčćęèéêëėįìíîïłńòóôöõøùúûüųūÿýżźñçčšžÀÁÂÄÃÅĄĆČĖĘÈÉÊËÌÍÎÏĮŁŃÒÓÔÖÕØÙÚÛÜŲŪŸÝŻŹÑßÇŒÆČŠŽ∂ð ,.'-]+$`
	nameRuleRegex := regexp.MustCompile(nameRuleRegexString)

	// Check that the first name matches naming conventions
	if !nameRuleRegex.MatchString(practitioner.FirstName) {
		errs = append(errs, "the first name contains a character which is not allowed")
	}

	// Check that the last name matches naming conventions
	if !nameRuleRegex.MatchString(practitioner.LastName) {
		errs = append(errs, "the last name contains a character which is not allowed")
	}

	// Get insolvency case from DB
	insolvencyCase, _, err := svc.GetInsolvencyPractitionersResource(transactionID)
	if err != nil {
		log.Error(fmt.Errorf("error getting insolvency case from DB: [%s]", err))
		return "", err
	}

	// Check if insolvency case is of type CVL and practitioner role is of type final liquidator
	if insolvencyCase.Data.CaseType == constants.CVL.String() && practitioner.Role != constants.FinalLiquidator.String() {
		errs = append(errs, fmt.Sprintf("the practitioner role must be "+constants.FinalLiquidator.String()+" because the insolvency case for transaction ID [%s] is of type "+constants.CVL.String(), transactionID))
	}

	return strings.Join(errs, ", "), nil
}

// ValidateAppointmentDetails checks that the incoming appointment details are valid
func ValidateAppointmentDetails(svc dao.Service, appointment models.PractitionerAppointment, transactionID string, practitionerID string, req *http.Request) ([]string, error) {

	var errs []string

	insolvencyResource, practitionerResourceDaos, err := svc.GetInsolvencyPractitionersResource(transactionID)
	if err != nil {
		errs = append(errs, err.Error())
		err = fmt.Errorf("error getting practitioner and insolvency resources from DB: [%s]", err)
		log.ErrorR(req, err)
		return errs, err
	}

	for _, practitioner := range practitionerResourceDaos {
		if practitioner.Data.PractitionerId == practitionerID && practitioner.Data.Appointment != nil && practitioner.Data.Appointment.Data.AppointedOn != "" {
			msg := fmt.Sprintf("practitioner ID [%s] already appointed to transaction ID [%s]", practitionerID, transactionID)
			log.Info(msg)
			errs = append(errs, msg)
		}
	}

	// Retrieve company incorporation date
	incorporatedOn, err := GetCompanyIncorporatedOn(insolvencyResource.Data.CompanyNumber, req)
	if err != nil {
		err = fmt.Errorf("error getting company details from DB: [%s]", err)
		log.ErrorR(req, err)
		return errs, err
	}

	ok, err := utils.IsDateBetweenIncorporationAndNow(appointment.AppointedOn, incorporatedOn)
	if err != nil {
		err = fmt.Errorf("error parsing date: [%s]", err)
		log.ErrorR(req, err)
		return errs, err
	}
	if !ok {
		errs = append(errs, fmt.Sprintf("appointed_on [%s] should not be in the future or before the company was incorporated", appointment.AppointedOn))
	}

	// // Check if appointment date supplied is different from stored appointment dates in DB
	for _, practitioner := range practitionerResourceDaos {
		if practitioner.Data.Appointment != nil && practitioner.Data.Appointment.Data.AppointedOn != "" && practitioner.Data.Appointment.Data.AppointedOn != appointment.AppointedOn {
			errs = append(errs, fmt.Sprintf("appointed_on [%s] differs from practitioner who was appointed on [%s]", appointment.AppointedOn, practitioner.Data.Appointment.Data.AppointedOn))
		}
	}

	// Check that a CVL case is only made by creditors
	if appointment.MadeBy != "" {
		if insolvencyResource.Data.CaseType == constants.CVL.String() && appointment.MadeBy != constants.Creditors.String() {
			errs = append(errs, fmt.Sprintf("made_by cannot be [%s] for insolvency case of type CVL", appointment.MadeBy))
		}
	}

	return errs, nil
}
