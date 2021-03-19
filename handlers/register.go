// Package handlers contains the http handlers which receive requests to be processed by the API.
package handlers

import (
	"net/http"

	"github.com/companieshouse/chs.go/authentication"

	"github.com/companieshouse/insolvency-api/dao"

	"github.com/companieshouse/chs.go/log"
	"github.com/gorilla/mux"
)

// Register defines the endpoints for the API
func Register(mainRouter *mux.Router, svc dao.Service) {

	userAuthInterceptor := &authentication.UserAuthenticationInterceptor{
		AllowAPIKeyUser:                true,
		RequireElevatedAPIKeyPrivilege: false,
	}

	mainRouter.HandleFunc("/insolvency/healthcheck", healthCheck).Methods(http.MethodGet).Name("healthcheck")

	// Create a router that requires all users to be authenticated when making requests
	appRouter := mainRouter.PathPrefix("/transactions/{transaction_id}/insolvency").Subrouter()
	appRouter.Use(userAuthInterceptor.UserAuthenticationIntercept)

	// Declare endpoint URIs
	appRouter.Handle("", HandleCreateInsolvencyResource(svc)).Methods(http.MethodPost).Name("createInsolvencyResource")

	mainRouter.Use(log.Handler)
}

func healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
