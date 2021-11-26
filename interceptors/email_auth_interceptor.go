// Package interceptors contains the interceptor middleware that checks for authorisation.
package interceptors

import (
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/authentication"
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/service"
)

// EmailAuthIntercept checks that the user has a registered Insolvency Practitioner email address in Mongo to perform the request action
func EmailAuthIntercept(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user details from context
		userDetails, ok := r.Context().Value(authentication.ContextKeyUserDetails).(authentication.AuthUserDetails)
		if !ok {
			log.ErrorR(r, fmt.Errorf("email auth interceptor error: invalid AuthUserDetails from context"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		isUserOnEfsAllowList, err := service.IsUserOnEfsAllowList(userDetails.Email, r)

		if err != nil {
			log.ErrorR(r, fmt.Errorf("error checking EFS allow list: [%s]", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !isUserOnEfsAllowList {
			log.ErrorR(r, fmt.Errorf("user not on EFS allow list"))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
