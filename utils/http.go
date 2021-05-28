package utils

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/log"
)

// WriteJSONWithStatus writes the interface as a json string with the supplied status.
func WriteJSONWithStatus(w http.ResponseWriter, r *http.Request, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.ErrorR(r, fmt.Errorf("error writing response: %v", err))
	}
}

// GetTransactionIDFromVars returns the transaction id from the supplied request vars.
func GetTransactionIDFromVars(vars map[string]string) string {
	transactionID := vars["transaction_id"]
	if transactionID == "" {
		return ""
	}

	return transactionID
}

// GetPractitionerIDFromVars returns the practitioner id from the supplied request vars.
func GetPractitionerIDFromVars(vars map[string]string) string {
	practitionerID := vars["practitioner_id"]
	if practitionerID == "" {
		return ""
	}

	return practitionerID
}

// GetTransactionIDFromVars returns the transaction id from the supplied request vars.
func GetAttachmentIDFromVars(vars map[string]string) string {
	attachmentID := vars["attachment_id"]
	if attachmentID == "" {
		return ""
	}

	return attachmentID
}

// ResponseTypeToStatus converts a response type to an http status
func ResponseTypeToStatus(responseType string) (int, error) {
	switch responseType {
	case "invalid-data":
		return http.StatusBadRequest, nil
	case "error":
		return http.StatusInternalServerError, nil
	case "success":
		return http.StatusOK, nil
	default:
		return 0, fmt.Errorf("response type not recognised")
	}
}
