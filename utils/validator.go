package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

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
