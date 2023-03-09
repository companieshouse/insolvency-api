// Package handlers contains the http handlers which receive requests to be processed by the API.
package handlers

import (
	"net/http"

	"github.com/companieshouse/chs.go/authentication"
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/config"
	"github.com/companieshouse/insolvency-api/dao"
	"github.com/companieshouse/insolvency-api/interceptors"
	"github.com/companieshouse/insolvency-api/utils"
	"github.com/gorilla/mux"
)

const (
	digitsAndDashRegex     = "[0-9-]+"
	insolvencyPath         = "/{transaction_id:" + digitsAndDashRegex + "}/insolvency"
	uuidCharsRegex         = "[a-f0-9-]+"
	attachmentsPath        = insolvencyPath + "/attachments"
	specificAttachmentPath = attachmentsPath + "/{attachment_id:" + uuidCharsRegex + "}"
)

// Register defines the endpoints for the API
func Register(mainRouter *mux.Router, svc dao.Service, helperService utils.HelperService) {

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
	publicAppRouter.Handle(insolvencyPath, HandleCreateInsolvencyResource(svc, helperService)).Methods(http.MethodPost).Name("createInsolvencyResource")

	publicAppRouter.Handle(insolvencyPath+"/validation-status", HandleGetValidationStatus(svc)).Methods(http.MethodGet).Name("getValidationStatus")

	publicAppRouter.Handle(insolvencyPath+"/practitioners", HandleCreatePractitionersResource(svc, helperService)).Methods(http.MethodPost).Name("createPractitionersResource")
	publicAppRouter.Handle(insolvencyPath+"/practitioners", HandleGetPractitionerResources(svc)).Methods(http.MethodGet).Name("getPractitionerResources")
	publicAppRouter.Handle(insolvencyPath+"/practitioners/{practitioner_id}", HandleDeletePractitioner(svc)).Methods(http.MethodDelete).Name("deletePractitioner")
	publicAppRouter.Handle(insolvencyPath+"/practitioners/{practitioner_id}", HandleGetPractitionerResource(svc)).Methods(http.MethodGet).Name("getPractitionerResource")

	publicAppRouter.Handle(insolvencyPath+"/practitioners/{practitioner_id}/appointment", HandleAppointPractitioner(svc, helperService)).Methods(http.MethodPost).Name("appointPractitioner")
	publicAppRouter.Handle(insolvencyPath+"/practitioners/{practitioner_id}/appointment", HandleGetPractitionerAppointment(svc)).Methods(http.MethodGet).Name("getPractitionerAppointment")
	publicAppRouter.Handle(insolvencyPath+"/practitioners/{practitioner_id}/appointment", HandleDeletePractitionerAppointment(svc)).Methods(http.MethodDelete).Name("deletePractitionerAppointment")

	publicAppRouter.Handle(attachmentsPath, HandleSubmitAttachment(svc, helperService)).Methods(http.MethodPost).Name("submitAttachment")
	publicAppRouter.Handle(specificAttachmentPath, HandleGetAttachmentDetails(svc, helperService)).Methods(http.MethodGet).Name("getAttachmentDetails")
	publicAppRouter.Handle(specificAttachmentPath+"/download", HandleDownloadAttachment(svc)).Methods(http.MethodGet).Name("downloadAttachment")
	publicAppRouter.Handle(specificAttachmentPath, HandleDeleteAttachment(svc)).Methods(http.MethodDelete).Name("deleteAttachment")

	publicAppRouter.Handle(insolvencyPath+"/resolution", HandleCreateResolution(svc, helperService)).Methods(http.MethodPost).Name("createResolution")
	publicAppRouter.Handle(insolvencyPath+"/resolution", HandleGetResolution(svc)).Methods(http.MethodGet).Name("getResolution")
	publicAppRouter.Handle(insolvencyPath+"/resolution", HandleDeleteResolution(svc)).Methods(http.MethodDelete).Name("deleteResolution")

	publicAppRouter.Handle(insolvencyPath+"/statement-of-affairs", HandleCreateStatementOfAffairs(svc, helperService)).Methods(http.MethodPost).Name("createStatementOfAffairs")
	publicAppRouter.Handle(insolvencyPath+"/statement-of-affairs", HandleGetStatementOfAffairs(svc)).Methods(http.MethodGet).Name("getStatementOfAffairs")
	publicAppRouter.Handle(insolvencyPath+"/statement-of-affairs", HandleDeleteStatementOfAffairs(svc)).Methods(http.MethodDelete).Name("deleteStatementOfAffairs")

	// Get environment config - only required whilst feature flag in use to disable
	// non-live form handling routes unless set to true
	cfg, err := config.Get()
	// Check environment variable to enable non-live form endpoints if set to true
	// and if so, block enable those handlers
	if err != nil {
		log.Info("Failed to get config for EnableNonLiveRouteHandlers")
	} else if cfg.EnableNonLiveRouteHandlers {
		publicAppRouter.Handle(insolvencyPath+"/progress-report", HandleCreateProgressReport(svc, helperService)).Methods(http.MethodPost).Name("createProgressReport")
		publicAppRouter.Handle(insolvencyPath+"/progress-report", HandleGetProgressReport(svc)).Methods(http.MethodGet).Name("getProgressReport")
		publicAppRouter.Handle(insolvencyPath+"/progress-report", HandleDeleteProgressReport(svc, helperService)).Methods(http.MethodDelete).Name("deleteProgressReport")
	} else {
		log.Info("Non-live endpoints blocked")
	}

	// Create a private router that requires all users to be authenticated when making requests
	privateAppRouter := mainRouter.PathPrefix("/private").Subrouter()
	privateAppRouter.Use(privateUserAuthInterceptor.UserAuthenticationIntercept)

	privateAppRouter.Handle("/transactions"+insolvencyPath+"/filings", HandleGetFilings(svc)).Methods(http.MethodGet).Name("getFilings")

	mainRouter.Use(log.Handler)
	mainRouter.Use(RecoveryHandler)
}

func healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
