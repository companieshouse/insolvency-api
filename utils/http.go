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
