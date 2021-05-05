package service

import (
	"testing"

	"github.com/companieshouse/insolvency-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestIsValidPractitionerDetails(t *testing.T) {

	Convey("Practitioner request supplied is invalid - neither email or telephone number are supplied", t, func() {
		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = ""
		practitioner.Email = ""

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "either telephone_number or email are required")
	})

	Convey("Practitioner request supplied is valid - email is supplied", t, func() {
		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = ""

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldBeBlank)
	})

	Convey("Practitioner request supplied is valid - telephone number is supplied", t, func() {
		practitioner := generatePractitioner()
		practitioner.Email = ""

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldBeBlank)
	})

	Convey("Practitioner request supplied is invalid - telephone number is less than 10 digits", t, func() {
		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = "07777777"

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "telephone_number must be a valid format")
	})

	Convey("Practitioner request supplied is invalid - telephone number is more than 11 digits", t, func() {
		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = "077777777777"

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "telephone_number must be a valid format")
	})

	Convey("Practitioner request supplied is invalid - telephone number does not consist solely of digits", t, func() {
		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = "077777777OO"

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "telephone_number must be a valid format")
	})

	Convey("Practitioner request supplied is invalid - telephone number contains spaces", t, func() {
		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = "0777777 7777"

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "telephone_number must not contain spaces")
	})

	Convey("Practitioner request supplied is invalid - telephone number does not begin with 0", t, func() {
		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = "77777777777"

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "telephone_number must start with 0")
	})

	Convey("Practitioner request supplied is invalid - first name does not match regex", t, func() {
		practitioner := generatePractitioner()
		practitioner.FirstName = "wr0ng"

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "the first name contains a character which is not allowed")
	})

	Convey("Practitioner request supplied is invalid - last name does not match regex", t, func() {
		practitioner := generatePractitioner()
		practitioner.LastName = "wr0ng"

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "the last name contains a character which is not allowed")
	})

	Convey("Practitioner request supplied is invalid - first and last name does not match regex", t, func() {
		practitioner := generatePractitioner()
		practitioner.FirstName = "name?"
		practitioner.LastName = "wr0ng"

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "the first name contains a character which is not allowed")
		So(err, ShouldContainSubstring, "the last name contains a character which is not allowed")
	})

	Convey("Practitioner request supplied is invalid - first and last name does not match regex and contact details missing", t, func() {
		practitioner := generatePractitioner()
		practitioner.FirstName = "name?"
		practitioner.LastName = "wr0ng"
		practitioner.Email = ""
		practitioner.TelephoneNumber = ""

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "either telephone_number or email are required")
		So(err, ShouldContainSubstring, "the first name contains a character which is not allowed")
		So(err, ShouldContainSubstring, "the last name contains a character which is not allowed")
	})

	Convey("Practitioner request supplied is valid - both telephone number and email are supplied", t, func() {
		practitioner := generatePractitioner()
		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldBeBlank)
	})
}

func generatePractitioner() models.PractitionerRequest {
	return models.PractitionerRequest{
		IPCode:          "1234",
		FirstName:       "Joe",
		LastName:        "Bloggs",
		TelephoneNumber: "01234567890",
		Email:           "a@b.com",
		Address: models.Address{
			AddressLine1: "addressline1",
			Locality:     "locality",
		},
		Role: "role",
	}
}
