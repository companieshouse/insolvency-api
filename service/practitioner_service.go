package service

import "github.com/companieshouse/insolvency-api/models"

// ValidatePractitionerContactDetails checks if the telephone number and email are missing
// in the request body. If they are missing, the method returns a human-readable error message.
func ValidatePractitionerContactDetails(practitioner models.PractitionerRequest) string {
	if practitioner.TelephoneNumber == "" && practitioner.Email == "" {
		return "either telephone_number or email are required"
	}

	return ""
}
