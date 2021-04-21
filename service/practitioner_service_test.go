package service

import (
	"testing"

	"github.com/companieshouse/insolvency-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestIsValidPractitioner(t *testing.T) {
	Convey("Practitioner request supplied is valid - both telephone number and email are supplied", t, func() {
		practitioner := generatePractitioner()
		err := ValidatePractitionerContactDetails(practitioner)

		So(err, ShouldBeBlank)
	})

	Convey("Practitioner request supplied is valid - telephone number is supplied", t, func() {
		practitioner := generatePractitioner()
		practitioner.Email = ""

		err := ValidatePractitionerContactDetails(practitioner)

		So(err, ShouldBeBlank)
	})

	Convey("Practitioner request supplied is valid - email is supplied", t, func() {
		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = ""

		err := ValidatePractitionerContactDetails(practitioner)

		So(err, ShouldBeBlank)
	})

	Convey("Practitioner request supplied is invalid", t, func() {
		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = ""
		practitioner.Email = ""

		err := ValidatePractitionerContactDetails(practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "either telephone_number or email are required")
	})
}

func generatePractitioner() models.PractitionerRequest {
	return models.PractitionerRequest{
		IPCode:          "1234",
		FirstName:       "Joe",
		LastName:        "Bloggs",
		TelephoneNumber: "123456",
		Email:           "email",
		Address: models.Address{
			AddressLine1: "addressline1",
			Locality:     "locality",
		},
		Role: "role",
	}
}
