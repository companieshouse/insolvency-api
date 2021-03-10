// Package handlers contains the http handlers which receive requests to be processed by the API.
package handlers

import (
	"net/http"

	"github.com/companieshouse/insolvency-api/dao"

	"github.com/companieshouse/chs.go/log"
	"github.com/gorilla/mux"
)

// Register defines the endpoints for the API
func Register(mainRouter *mux.Router, svc dao.Service) {

	mainRouter.HandleFunc("/insolvency/healthcheck", healthCheck).Methods(http.MethodGet).Name("healthcheck")

	mainRouter.Handle("/transactions/{transaction_id}/insolvency", HandleCreateInsolvencyResource(svc)).Methods(http.MethodPost).Name("createInsolvencyResource")
	mainRouter.Use(log.Handler)
}

func healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
