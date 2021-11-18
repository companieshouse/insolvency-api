// Package interceptors contains the interceptor middleware that checks for authorisation.
package interceptors

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/companieshouse/insolvency-api/service"
)

// EmailAuthIntercept checks that the user has a registered Insolvency Practitioner email address in Mongo to perform the request action
func EmailAuthIntercept(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// format of ERIC Header: ERIC-Authorised-User: some@email.address forename*=UTF-8''%c2%a3%20A%20pound; surname=This-has-no-utf-8
		ericAuthorisedUserHeader := r.Header.Get("ERIC-Authorised-User")
		oauth2UserEmail := strings.Fields(ericAuthorisedUserHeader)[0]

		// TODO: remove this logging comment - just here to check email is extracted
		fmt.Println(oauth2UserEmail, "<= oauth2UserEmail") 

		isUserOnEfsAllowList, err := service.IsUserOnEfsAllowList(oauth2UserEmail, r)

		// TODO: error handling
		
		fmt.Println(isUserOnEfsAllowList, "<= isUserOnEfsAllowList") 
		fmt.Println(err, "<= error") 

		// if all fine
		next.ServeHTTP(w, r)
	})
}