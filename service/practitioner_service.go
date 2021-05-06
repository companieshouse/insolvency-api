package service

import (
	"regexp"
	"strings"

	"github.com/companieshouse/insolvency-api/models"
)

// ValidatePractitionerDetails checks that the incoming practitioner details are valid
func ValidatePractitionerDetails(practitioner models.PractitionerRequest) string {
	var errs []string

	// Check that either the telephone number or email field are populated
	if practitioner.TelephoneNumber == "" && practitioner.Email == "" {
		errs = append(errs, "either telephone_number or email are required")
	}

	// Set allowed regexp for telephone number
	telephoneNumberRuleRegexString := `^[0]\d{9}$|^[0]\d{10}$`
	telephoneNumberRegex := regexp.MustCompile(telephoneNumberRuleRegexString)

	// Check that telephone number is a valid format
	if practitioner.TelephoneNumber != "" && !telephoneNumberRegex.MatchString(practitioner.TelephoneNumber) {
		errs = append(errs, "telephone_number must be a valid format")
	}

	// Check that telephone number does not contain spaces
	if practitioner.TelephoneNumber != "" && strings.Contains(practitioner.TelephoneNumber, " ") {
		errs = append(errs, "telephone_number must not contain spaces")
	}

	// Check that telephone number starts with 0
	if practitioner.TelephoneNumber != "" && !strings.HasPrefix(practitioner.TelephoneNumber, "0") {
		errs = append(errs, "telephone_number must start with 0")
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

	return strings.Join(errs, ", ")
}
