// Package interceptors contains the interceptor middleware that checks for authorisation.
package interceptors

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/service"
)

// EmailAuthIntercept checks that the user has a registered Insolvency Practitioner email address in Mongo to perform the request action
func EmailAuthIntercept(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// format of ERIC Header: ERIC-Authorised-User: some@email.address forename*=UTF-8''%c2%a3%20A%20pound; surname=This-has-no-utf-8
		ericAuthorisedUserHeader := r.Header.Get("ERIC-Authorised-User")
		oauth2UserEmail := strings.Fields(ericAuthorisedUserHeader)[0]

		isUserOnEfsAllowList, err := service.IsUserOnEfsAllowList(oauth2UserEmail, r)

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
