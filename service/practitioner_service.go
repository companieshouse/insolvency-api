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

	// Set allowed telephone number convention
	telephoneNumberRuleRegexString := `^[0][1-9]\d{9}$`
	telephoneNumberRuleRegex := regexp.MustCompile(telephoneNumberRuleRegexString)

	if practitioner.TelephoneNumber != "" && !telephoneNumberRuleRegex.MatchString(practitioner.TelephoneNumber) {
		errs = append(errs, "telephone_number must start with 0 and be 11 digits long")
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
