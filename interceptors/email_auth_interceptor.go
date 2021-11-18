package interceptors

import (
	"net/http"

	"github.com/companieshouse/chs.go/log"
)

// TokenPermissionsAuthIntercept checks that the user has the correct token permissions to perform the request action
func (interceptor *EmailAuthInterceptor) EmailAuthIntercept(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// format of ERIC Header: ERIC-Authorised-User: some@email.address forename*=UTF-8''%c2%a3%20A%20pound; surname=This-has-no-utf-8
		ericAuthorisedUserHeader := r.Header.Get("ERIC-Authorised-User")
		oauth2UserEmail := ericAuthorisedUserHeader.Fields(ericAuthorisedUserHeader)[0]
		log.Info(oauth2UserEmail) 

		// Now that we have the authorised token permissions and admin roles there are multiple cases that can be allowed through:
		switch {
		case hasPermissionFollowRead && isReadRequest:
			// 1) Authorized user has read permissions and this is a GET request
			log.InfoR(r, "TokenPermissionsAuthInterceptor authorised with read permission on GET request", debugMap)
			// Call the next handler
			next.ServeHTTP(w, r)
		case hasPermissionFollowUpdate && isUpdateRequest:
			// 2) Authorized user has read and update permissions
			log.InfoR(r, "TokenPermissionsAuthInterceptor authorised with update permissions", debugMap)
			// Call the next handler
			next.ServeHTTP(w, r)
		case hasAdminManageMonitorRole && isReadRequest:
			// 3) Authorized user has manage monitor role and request is GET request
			log.InfoR(r, "TokenPermissionsAuthInterceptor authorised as admin user with manage monitor role", debugMap)
			// Call the next handler
			next.ServeHTTP(w, r)
		default:
			// If none of the above conditions above are met then the request is
			// unauthorized
			w.WriteHeader(http.StatusUnauthorized)
			log.InfoR(r, "TokenPermissionsAuthInterceptor unauthorised", debugMap)
		}
	})
}