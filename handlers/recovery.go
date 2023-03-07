package handlers

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
)

// RecoveryHandler is a handler wrapper that catches runtime panics and returns a 500 Internal Server Error
func RecoveryHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				log.ErrorR(req, fmt.Errorf("runtime error [%s]: %s", err, debug.Stack()))
				m := models.NewMessageResponse("there was a problem handling your request")
				utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, req)
	})
}
