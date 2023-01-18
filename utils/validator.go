package utils

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gorilla/mux"

	"github.com/companieshouse/chs.go/log"
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

func ValidateTransaction(helperService HelperService, req *http.Request, res http.ResponseWriter, startMessage string, checkIfTransactionClosedFn func(transactionID string, req *http.Request) (bool, error, int)) (string, bool) {

	// Check transaction id exists in path
	incomingTransactionId := GetTransactionIDFromVars(mux.Vars(req))
	isValidTransaction, transactionID := helperService.HandleTransactionIdExistsValidation(res, req, incomingTransactionId)
	if !isValidTransaction {
		return transactionID, isValidTransaction
	}

	log.InfoR(req, fmt.Sprintf("start POST request for submit "+startMessage+"with transaction id: %s", transactionID))

	// Check if transaction is checkIfTransactionClosedFn
	var isTransactionClosed, err, httpStatus = checkIfTransactionClosedFn(transactionID, req)
	isValidTransaction = helperService.HandleTransactionNotClosedValidation(res, req, transactionID, isTransactionClosed, httpStatus, err)
	if !isValidTransaction {
		return transactionID, isValidTransaction
	}

	return transactionID, isValidTransaction
}
