// Package interceptors contains the interceptor middleware that checks for authorisation.
package interceptors

import (
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/authentication"
	"github.com/companieshouse/chs.go/log"
)

// InsolvencyPermissionsIntercept checks that the user has the necessary token permissions for an insolvency practitioner
func InsolvencyPermissionsIntercept(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tp := &authentication.TokenPermissions{}
		err := tp.DecodeAuthorisedTokenPermissions(r)
		if err != nil {
			log.ErrorR(r, fmt.Errorf("TokenPermissionsAuthInterceptor error decoding token permissions: [%v]", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		isReadRequest := http.MethodGet == r.Method
		isUpdateRequest := http.MethodPost == r.Method || http.MethodDelete == r.Method
		hasPermissionInsolvencyRead := tp.HasPermission(authentication.PermissionKeyInsolvencyCases, authentication.PermissionValueRead)
		hasPermissionInsolvencyUpdate := tp.HasPermission(authentication.PermissionKeyInsolvencyCases, authentication.PermissionValueUpdate)

		switch {
		case hasPermissionInsolvencyRead && isReadRequest:
			next.ServeHTTP(w, r)
		case hasPermissionInsolvencyUpdate && isUpdateRequest:
			next.ServeHTTP(w, r)
		default:
			log.InfoR(r, "InsolvencyPermissionsIntercept unauthorised")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
