package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/dao"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/service"
	"github.com/companieshouse/insolvency-api/transformers"
	"github.com/companieshouse/insolvency-api/utils"
	"github.com/gorilla/mux"
)

// HandleCreatePractitionersResource updates the insolvency resource with the
// incoming list of practitioners
func HandleCreatePractitionersResource(svc dao.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Check for a transaction id in request
		vars := mux.Vars(req)
		transactionID := utils.GetTransactionIDFromVars(vars)
		if transactionID == "" {
			log.ErrorR(req, fmt.Errorf("there is no transaction id in the url path"))
			m := models.NewMessageResponse("transaction id is not in the url path")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		log.InfoR(req, fmt.Sprintf("start POST request for practitioners resource with transaction id: %s", transactionID))

		// Decode the incoming request to create a list of practitioners
		var request models.PractitionerRequest
		err := json.NewDecoder(req.Body).Decode(&request)

		// Request body failed to get decoded
		if err != nil {
			log.ErrorR(req, fmt.Errorf("invalid request"))
			m := models.NewMessageResponse(fmt.Sprintf("failed to read request body for transaction %s", transactionID))
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// Validate all mandatory fields
		if errs := utils.Validate(request); errs != "" {
			log.ErrorR(req, fmt.Errorf("invalid request - failed validation on the following: %s", errs))
			m := models.NewMessageResponse("invalid request body: " + errs)
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// Check if practitioner role supplied is valid
		if ok := constants.IsInRoleList(request.Role); !ok {
			log.ErrorR(req, fmt.Errorf("invalid practitioner role"))
			m := models.NewMessageResponse(fmt.Sprintf("the practitioner role supplied is not valid %s", request.Role))
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		practitionerDao := transformers.PractitionerResourceRequestToDB(&request, transactionID)

		// Store practitioners resource in Mongo
		err, statusCode := svc.CreatePractitionersResource(practitionerDao, transactionID)
		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse(err.Error())
			utils.WriteJSONWithStatus(w, req, m, statusCode)
			return
		}

		log.InfoR(req, fmt.Sprintf("successfully added practitioners resource with transaction ID: %s, to mongo", transactionID))

		utils.WriteJSONWithStatus(w, req, transformers.PractitionerResourceDaoToCreatedResponse(practitionerDao), http.StatusCreated)
	})
}

// HandleGetPractitionerResources retrieves a list of practitioners for the insolvency case with
// the specified transactionID
func HandleGetPractitionerResources(svc dao.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Check for a transaction id in request
		vars := mux.Vars(req)
		transactionID := utils.GetTransactionIDFromVars(vars)
		if transactionID == "" {
			log.ErrorR(req, fmt.Errorf("there is no transaction id in the url path"))
			m := models.NewMessageResponse("transaction id is not in the url path")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		log.InfoR(req, fmt.Sprintf("start GET request for practitioners resource with transaction id: %s", transactionID))

		practitionerResources, err := svc.GetPractitionerResources(transactionID)
		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse(err.Error())
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}
		if practitionerResources == nil {
			log.ErrorR(req, fmt.Errorf("insolvency case for transaction %s not found", transactionID))
			m := models.NewMessageResponse("there was a problem handling your request for insolvency case with transaction ID: " + transactionID + " not found")
			utils.WriteJSONWithStatus(w, req, m, http.StatusNotFound)
			return
		}
		if len(practitionerResources) == 0 {
			log.ErrorR(req, fmt.Errorf("practitioners for insolvency case with transaction %s not found", transactionID))
			m := models.NewMessageResponse("there was a problem handling your request for insolvency case with transaction: " + transactionID + " there are no practitioners assigned to this case")
			utils.WriteJSONWithStatus(w, req, m, http.StatusNotFound)
			return
		}

		utils.WriteJSONWithStatus(w, req, transformers.PractitionerResourceDaoListToCreatedResponseList(practitionerResources), http.StatusOK)

	})
}

// HandleDeletePractitioner deletes a practitioner from the insolvency case with
// the specified transactionID and IPCode
func HandleDeletePractitioner(svc dao.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Check for a transaction id in request
		vars := mux.Vars(req)
		transactionID := utils.GetTransactionIDFromVars(vars)
		if transactionID == "" {
			log.ErrorR(req, fmt.Errorf("there is no transaction id in the url path"))
			m := models.NewMessageResponse("transaction id is not in the url path")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}
		// Check for a practitioner id in request
		practitionerID := utils.GetPractitionerIDFromVars(vars)
		if practitionerID == "" {
			log.ErrorR(req, fmt.Errorf("there is no practitioner id in the url path"))
			m := models.NewMessageResponse("practitioner id is not in the url path")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		log.InfoR(req, fmt.Sprintf("start DELETE request for practitioner resource with transaction id: %s and practitioner id: %s", transactionID, practitionerID))
		// Delete practitioner from Mongo
		err, statusCode := svc.DeletePractitioner(practitionerID, transactionID)
		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse(err.Error())
			utils.WriteJSONWithStatus(w, req, m, statusCode)
			return
		}

		log.InfoR(req, fmt.Sprintf("successfully deleted practitioner with transaction ID: %s and practitioner ID: %s, from mongo", transactionID, practitionerID))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
	})
}

// HandleAppointPractitioner adds appointment details to a practitioner resource on a transaction
func HandleAppointPractitioner(svc dao.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Check for a transaction id in request
		vars := mux.Vars(req)
		transactionID := utils.GetTransactionIDFromVars(vars)
		if transactionID == "" {
			log.ErrorR(req, fmt.Errorf("there is no Transaction ID in the URL path"))
			m := models.NewMessageResponse("Transaction ID is not in the URL path")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// Check for a practitioner ID in request
		practitionerID := utils.GetPractitionerIDFromVars(vars)
		if practitionerID == "" {
			log.ErrorR(req, fmt.Errorf("there is no Practitioner ID in the URL path"))
			m := models.NewMessageResponse("Practitioner ID is not in the URL path")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		log.InfoR(req, fmt.Sprintf("start POST request for practitioner appointment with transaction ID: [%s] and practitioner ID: [%s]", transactionID, practitionerID))

		// Decode the incoming request
		var request models.PractitionerAppointment
		err := json.NewDecoder(req.Body).Decode(&request)
		// Request body failed to get decoded
		if err != nil {
			log.ErrorR(req, fmt.Errorf("invalid request"))
			m := models.NewMessageResponse(fmt.Sprintf("failed to read request body for transactionID [%s] and practitionerID [%s]", transactionID, practitionerID))
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// Validate all mandatory fields
		if errs := utils.Validate(request); errs != "" {
			log.ErrorR(req, fmt.Errorf("invalid request - failed validation on the following: %s", errs))
			m := models.NewMessageResponse("invalid request body: " + errs)
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// Check if made_by supplied is valid
		if ok := constants.IsAppointmentMadeByInList(request.MadeBy); !ok {
			log.ErrorR(req, fmt.Errorf("invalid appointment made_by"))
			m := models.NewMessageResponse(fmt.Sprintf("the appointment made_by supplied is not valid: [%s]", request.MadeBy))
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// Check whether practitioner has already been appointed
		err, alreadyAppointed := service.CheckPractitionerAlreadyAppointed(svc, transactionID, practitionerID, req)
		if err != nil {
			m := models.NewMessageResponse("error checking practitioner details")
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}
		if alreadyAppointed {
			msg := fmt.Sprintf("practitioner ID [%s] already appointed for transaction ID [%s]", practitionerID, transactionID)
			log.Info(msg)
			m := models.NewMessageResponse(msg)
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// Check appointment date valid
		err, validAppointmentDate := service.CheckAppointmentDateValid(svc, transactionID, request.AppointedOn, req)
		if err != nil {
			m := models.NewMessageResponse("error checking practitioner details")
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}
		if !validAppointmentDate {
			msg := fmt.Sprintf("appointment Date [%s] differs for practitioner ID [%s] and transaction ID [%s]", request.AppointedOn, practitionerID, transactionID)
			log.Info(msg)
			m := models.NewMessageResponse(msg)
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		practitionerAppointmentDao := transformers.PractitionerAppointmentRequestToDB(&request, transactionID, practitionerID)

		// Store appointment in DB
		err, statusCode := svc.AppointPractitioner(practitionerAppointmentDao, transactionID, practitionerID)
		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse(err.Error())
			utils.WriteJSONWithStatus(w, req, m, statusCode)
			return
		}

		log.InfoR(req, fmt.Sprintf("successfully added practitioner appointment with transaction ID [%s] and practitioner ID [%s] to mongo", transactionID, practitionerID))

		practitioner, err := svc.GetPractitionerResource(practitionerID, transactionID)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Check if practitioner is empty (not found). This should not happen, as appointment has just been added.
		if practitioner == (models.PractitionerResourceDao{}) {
			log.ErrorR(req, fmt.Errorf("practitionerID [%s] not found for transactionID [%s]", practitionerID, transactionID))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		appointmentResponse := transformers.PractitionerAppointmentDaoToResponse(practitioner.Appointment)

		utils.WriteJSONWithStatus(w, req, appointmentResponse, http.StatusOK)
	})
}
