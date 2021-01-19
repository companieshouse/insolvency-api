package handlers

import (
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/gorilla/mux"
)

func Register(mainRouter *mux.Router) {
	mainRouter.HandleFunc("/insolvency-service/healthcheck", healthCheck).Methods(http.MethodGet).Name("healthcheck")
	mainRouter.Use(log.Handler)
}

func healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
