package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/companieshouse/insolvency-api/models"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

// Validate takes in any request object and checks whether it has met
// the validation criteria according to the annotations on that object.
// If the object is invalid, the method returns a human-readable string
// which can then be added to the message response for the API user
func Validate(data interface{}) string {
	v := validator.New()
	v.RegisterTagNameFunc(extractJson)
	english := en.New()
	uni := ut.New(english, english)
	trans, _ := uni.GetTranslator("en")
	_ = en_translations.RegisterDefaultTranslations(v, trans)

	err := v.Struct(data)

	if err == nil {
		return ""
	}

	var errs []string
	validatorErrs := err.(validator.ValidationErrors)
	for _, ve := range validatorErrs {
		translatedErr := fmt.Errorf(ve.Translate(trans))
		errs = append(errs, translatedErr.Error())
	}

	return strings.Join(errs, ", ")
}

func extractJson(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

	if name == "-" {
		return ""
	}

	return name
}

// ValidatePractitionerContactDetails checks if the telephone number and email are missing
// in the request body. If they are missing, the method returns a human-readable error message.
func ValidatePractitionerContactDetails(practitioner models.PractitionerRequest) string {
	if practitioner.TelephoneNumber == "" && practitioner.Email == "" {
		return "either telephone_number or email are required"
	}

	return ""
}
