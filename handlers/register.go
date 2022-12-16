// Package handlers contains the http handlers which receive requests to be processed by the API.
package handlers

import (
	"net/http"

	"github.com/companieshouse/chs.go/authentication"
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/config"
	"github.com/companieshouse/insolvency-api/dao"
	"github.com/companieshouse/insolvency-api/interceptors"
	"github.com/gorilla/mux"
)

// Register defines the endpoints for the API
func Register(mainRouter *mux.Router, svc dao.Service) {

	userAuthInterceptor := &authentication.UserAuthenticationInterceptor{
		AllowAPIKeyUser:                false,
		RequireElevatedAPIKeyPrivilege: false,
	}

	privateUserAuthInterceptor := &authentication.UserAuthenticationInterceptor{
		AllowAPIKeyUser:                true,
		RequireElevatedAPIKeyPrivilege: true,
	}

	mainRouter.HandleFunc("/insolvency/healthcheck", healthCheck).Methods(http.MethodGet).Name("healthcheck")

	// Create a public router that requires all users to be authenticated when making requests
	publicAppRouter := mainRouter.PathPrefix("/transactions").Subrouter()
	publicAppRouter.Use(userAuthInterceptor.UserAuthenticationIntercept, interceptors.EmailAuthIntercept, interceptors.InsolvencyPermissionsIntercept)

	// Declare endpoint URIs
	publicAppRouter.Handle("/{transaction_id}/insolvency", HandleCreateInsolvencyResource(svc)).Methods(http.MethodPost).Name("createInsolvencyResource")

	publicAppRouter.Handle("/{transaction_id}/insolvency/validation-status", HandleGetValidationStatus(svc)).Methods(http.MethodGet).Name("getValidationStatus")

	publicAppRouter.Handle("/{transaction_id}/insolvency/practitioners", HandleCreatePractitionersResource(svc)).Methods(http.MethodPost).Name("createPractitionersResource")
	publicAppRouter.Handle("/{transaction_id}/insolvency/practitioners", HandleGetPractitionerResources(svc)).Methods(http.MethodGet).Name("getPractitionerResources")
	publicAppRouter.Handle("/{transaction_id}/insolvency/practitioners/{practitioner_id}", HandleDeletePractitioner(svc)).Methods(http.MethodDelete).Name("deletePractitioner")
	publicAppRouter.Handle("/{transaction_id}/insolvency/practitioners/{practitioner_id}", HandleGetPractitionerResource(svc)).Methods(http.MethodGet).Name("getPractitionerResource")

	publicAppRouter.Handle("/{transaction_id}/insolvency/practitioners/{practitioner_id}/appointment", HandleAppointPractitioner(svc)).Methods(http.MethodPost).Name("appointPractitioner")
	publicAppRouter.Handle("/{transaction_id}/insolvency/practitioners/{practitioner_id}/appointment", HandleGetPractitionerAppointment(svc)).Methods(http.MethodGet).Name("getPractitionerAppointment")
	publicAppRouter.Handle("/{transaction_id}/insolvency/practitioners/{practitioner_id}/appointment", HandleDeletePractitionerAppointment(svc)).Methods(http.MethodDelete).Name("deletePractitionerAppointment")

	publicAppRouter.Handle("/{transaction_id}/insolvency/attachments", HandleSubmitAttachment(svc)).Methods(http.MethodPost).Name("submitAttachment")
	publicAppRouter.Handle("/{transaction_id}/insolvency/attachments/{attachment_id}", HandleGetAttachmentDetails(svc)).Methods(http.MethodGet).Name("getAttachmentDetails")
	publicAppRouter.Handle("/{transaction_id}/insolvency/attachments/{attachment_id}/download", HandleDownloadAttachment(svc)).Methods(http.MethodGet).Name("downloadAttachment")
	publicAppRouter.Handle("/{transaction_id}/insolvency/attachments/{attachment_id}", HandleDeleteAttachment(svc)).Methods(http.MethodDelete).Name("deleteAttachment")

	publicAppRouter.Handle("/{transaction_id}/insolvency/resolution", HandleCreateResolution(svc)).Methods(http.MethodPost).Name("createResolution")
	publicAppRouter.Handle("/{transaction_id}/insolvency/resolution", HandleGetResolution(svc)).Methods(http.MethodGet).Name("getResolution")
	publicAppRouter.Handle("/{transaction_id}/insolvency/resolution", HandleDeleteResolution(svc)).Methods(http.MethodDelete).Name("deleteResolution")

	// Get environment config - only required whilst feature flag in use to disable
	// non-live form handling routes unless set to true
	cfg, err := config.Get()
	// Check environment variable to enable non-live form endpoints if set to true
	// and if so, block enable those handlers
	if err != nil {
		log.Info("Failed to get config for EnableNonLiveRouteHandlers")
	} else if cfg.EnableNonLiveRouteHandlers {
		publicAppRouter.Handle("/{transaction_id}/insolvency/statement-of-affairs", HandleCreateStatementOfAffairs(svc)).Methods(http.MethodPost).Name("createStatementOfAffairs")
		publicAppRouter.Handle("/{transaction_id}/insolvency/statement-of-affairs", HandleGetStatementOfAffairs(svc)).Methods(http.MethodGet).Name("getStatementOfAffairs")
		publicAppRouter.Handle("/{transaction_id}/insolvency/statement-of-affairs", HandleDeleteStatementOfAffairs(svc)).Methods(http.MethodDelete).Name("deleteStatementOfAffairs")

		publicAppRouter.Handle("/{transaction_id}/insolvency/progress-report", HandleCreateProgressReport(svc)).Methods(http.MethodPost).Name("createProgressReport")
	} else {
		log.Info("Non-live endpoints blocked")
	}

	// Create a private router that requires all users to be authenticated when making requests
	privateAppRouter := mainRouter.PathPrefix("/private").Subrouter()
	privateAppRouter.Use(privateUserAuthInterceptor.UserAuthenticationIntercept)

	privateAppRouter.Handle("/transactions/{transaction_id}/insolvency/filings", HandleGetFilings(svc)).Methods(http.MethodGet).Name("getFilings")

	mainRouter.Use(log.Handler)
}

func healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
