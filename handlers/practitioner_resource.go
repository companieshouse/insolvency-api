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
func HandleCreatePractitionersResource(svc dao.Service, helperService utils.HelperService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Check transaction id exists in path
		incomingTransactionId := utils.GetTransactionIDFromVars(mux.Vars(req))
		transactionID, isValidTransactionId, httpStatusCode := helperService.HandleTransactionIdExistsValidation(w, req, incomingTransactionId)
		if !isValidTransactionId {
			http.Error(w, "Bad request", httpStatusCode)
			return
		}

		log.InfoR(req, fmt.Sprintf("start POST request for practitioners resource with transaction id: %s", transactionID))

		// Check if transaction is closed
		isTransactionClosed, err, httpStatus := service.CheckIfTransactionClosed(transactionID, req)
		isValidTransactionNotClosed, httpStatusCode, _ := helperService.HandleTransactionNotClosedValidation(w, req, transactionID, isTransactionClosed, httpStatus, err)
		if !isValidTransactionNotClosed {
			http.Error(w, "Transaction closed", httpStatusCode)
			return
		}

		// Decode the incoming request to create a list of practitioners
		var request models.PractitionerRequest
		err = json.NewDecoder(req.Body).Decode(&request)
		isValidDecoded, httpStatusCode := helperService.HandleBodyDecodedValidation(w, req, transactionID, err)
		if !isValidDecoded {
			http.Error(w, fmt.Sprintf("failed to read request body for transaction %s", transactionID), httpStatusCode)
			return
		}

		// Validate all mandatory fields
		errs := utils.Validate(request)
		isValidMarshallToDB, httpStatusCode := helperService.HandleMandatoryFieldValidation(w, req, errs)
		if !isValidMarshallToDB {
			http.Error(w, errs, httpStatusCode)
			return
		}

		// Validates that the provided practitioner details are in the correct format
		validationErrs, err := service.ValidatePractitionerDetails(svc, transactionID, request)
		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse("failed to validate the practitioner request supplied")
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}
		if validationErrs != "" {
			log.ErrorR(req, fmt.Errorf("invalid request - failed validation on the following: %s", validationErrs))
			m := models.NewMessageResponse("invalid request body: " + validationErrs)
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

// HandleGetPractitionerResource retrieves a practitioner with the specified practitionerID
// on the insolvency case with the specified transactionID
func HandleGetPractitionerResource(svc dao.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Check for a transaction id in request
		vars := mux.Vars(req)
		transactionID, practitionerID, err := getTransactionIDAndPractitionerIDFromVars(vars)
		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse(err.Error())
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		log.InfoR(req, fmt.Sprintf("start GET request for practitioner resource with transaction id: %s and practitioner id: %s", transactionID, practitionerID))

		// Get practitioner from DB
		practitioner, err := svc.GetPractitionerResource(practitionerID, transactionID)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("failed to get practitioner with id [%s]: [%s]", practitionerID, err))
			m := models.NewMessageResponse("there was a problem handling your request")
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}

		// Check if practitioner returned is empty
		if practitioner == (models.PractitionerResourceDao{}) {
			message := fmt.Sprintf("practitioner with ID [%s] not found", practitionerID)
			log.Debug(message)
			m := models.NewMessageResponse(message)
			utils.WriteJSONWithStatus(w, req, m, http.StatusNotFound)
			return
		}

		// Successfully retrieved practitioner
		utils.WriteJSONWithStatus(w, req, transformers.PractitionerResourceDaoToCreatedResponse(&practitioner), http.StatusOK)
	})
}

// HandleDeletePractitioner deletes a practitioner from the insolvency case with
// the specified transactionID and IPCode
func HandleDeletePractitioner(svc dao.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Check for a transaction id in request
		vars := mux.Vars(req)
		transactionID, practitionerID, err := getTransactionIDAndPractitionerIDFromVars(vars)
		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse(err.Error())
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		log.InfoR(req, fmt.Sprintf("start DELETE request for practitioner resource with transaction id: %s and practitioner id: %s", transactionID, practitionerID))

		// Check if transaction is closed
		isTransactionClosed, err, httpStatus := service.CheckIfTransactionClosed(transactionID, req)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("error checking transaction status for [%v]: [%s]", transactionID, err))
			m := models.NewMessageResponse(fmt.Sprintf("error checking transaction status for [%v]: [%s]", transactionID, err))
			utils.WriteJSONWithStatus(w, req, m, httpStatus)
			return
		}
		if isTransactionClosed {
			log.ErrorR(req, fmt.Errorf("transaction [%v] is already closed and cannot be updated", transactionID))
			m := models.NewMessageResponse(fmt.Sprintf("transaction [%v] is already closed and cannot be updated", transactionID))
			utils.WriteJSONWithStatus(w, req, m, httpStatus)
			return
		}

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
func HandleAppointPractitioner(svc dao.Service, helperService utils.HelperService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Check transaction id exists in path
		incomingTransactionId, practitionerID, err := getTransactionIDAndPractitionerIDFromVars(mux.Vars(req))
		transactionID, isValidTransactionId, httpStatusCode := helperService.HandleTransactionIdExistsValidation(w, req, incomingTransactionId)
		if !isValidTransactionId {
			http.Error(w, "Bad request", httpStatusCode)
			return
		}

		log.InfoR(req, fmt.Sprintf("start POST request for practitioner appointment with transaction ID: [%s] and practitioner ID: [%s]", transactionID, practitionerID))

		// Check if transaction is closed
		isTransactionClosed, err, httpStatus := service.CheckIfTransactionClosed(transactionID, req)
		isValidTransactionNotClosed, httpStatusCode, _ := helperService.HandleTransactionNotClosedValidation(w, req, transactionID, isTransactionClosed, httpStatus, err)
		if !isValidTransactionNotClosed {
			http.Error(w, "Transaction closed", httpStatusCode)
			return
		}

		// Decode the incoming request to create a list of practitioners
		var request models.PractitionerAppointment
		err = json.NewDecoder(req.Body).Decode(&request)
		isValidDecoded, httpStatusCode := helperService.HandleBodyDecodedValidation(w, req, transactionID, err)
		if !isValidDecoded {
			http.Error(w, fmt.Sprintf("failed to read request body for transaction %s", transactionID), httpStatusCode)
			return
		}

		// Validate all mandatory fields
		errs := utils.Validate(request)
		isValidMarshallToDB, httpStatusCode := helperService.HandleMandatoryFieldValidation(w, req, errs)
		if !isValidMarshallToDB {
			http.Error(w, errs, httpStatusCode)
			return
		}

		// Check if made_by supplied is valid
		if ok := constants.IsAppointmentMadeByInList(request.MadeBy); !ok {
			log.ErrorR(req, fmt.Errorf("invalid appointment made_by"))
			m := models.NewMessageResponse(fmt.Sprintf("the appointment made_by supplied is not valid: [%s]", request.MadeBy))
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// Validate all appointment details are of the correct format and criteria
		validationErrs, err := service.ValidateAppointmentDetails(svc, request, transactionID, practitionerID, req)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("failed to validate appointment details: [%s]", err))
			m := models.NewMessageResponse(fmt.Sprintf("there was a problem handling your request for transaction ID [%s]", transactionID))
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}
		if validationErrs != "" {
			log.ErrorR(req, fmt.Errorf("invalid request - failed validation on the following: %s", validationErrs))
			m := models.NewMessageResponse("invalid request body: " + validationErrs)
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
			log.ErrorR(req, err)
			m := models.NewMessageResponse(err.Error())
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}

		// Check if practitioner is empty (not found). This should not happen, as appointment has just been added.
		if practitioner == (models.PractitionerResourceDao{}) {
			msg := fmt.Sprintf("practitionerID [%s] not found for transactionID [%s]", practitionerID, transactionID)
			log.InfoR(req, msg)
			m := models.NewMessageResponse(msg)
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}

		appointmentResponse := transformers.PractitionerAppointmentDaoToResponse(*practitioner.Appointment)

		utils.WriteJSONWithStatus(w, req, appointmentResponse, http.StatusCreated)
	})
}

// HandleGetPractitionerAppointment retrieves appointment details
// for the specified transactionID and practitionerID
func HandleGetPractitionerAppointment(svc dao.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		vars := mux.Vars(req)
		transactionID, practitionerID, err := getTransactionIDAndPractitionerIDFromVars(vars)
		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse(err.Error())
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		log.InfoR(req, fmt.Sprintf("start GET request for appointments resource with transaction ID: [%s] and practitioner ID: [%s]", transactionID, practitionerID))

		practitioner, err := svc.GetPractitionerResource(practitionerID, transactionID)
		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse(err.Error())
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}

		// Check if practitioner is empty (not found).
		if practitioner == (models.PractitionerResourceDao{}) {
			msg := fmt.Sprintf("practitionerID [%s] not found for transactionID [%s]", practitionerID, transactionID)
			log.InfoR(req, msg)
			m := models.NewMessageResponse(msg)
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}

		// Check if practitioner has an appointment
		if practitioner.Appointment == nil {
			msg := fmt.Sprintf("No appointment found for practitionerID [%s] an transactionID [%s]", practitionerID, transactionID)
			log.InfoR(req, msg)
			m := models.NewMessageResponse(msg)
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}

		appointmentResponse := transformers.PractitionerAppointmentDaoToResponse(*practitioner.Appointment)

		utils.WriteJSONWithStatus(w, req, appointmentResponse, http.StatusOK)
	})
}

// HandleDeletePractitionerAppointment deletes an appointment
// for the specified transactionID and practitionerID
func HandleDeletePractitionerAppointment(svc dao.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		transactionID, practitionerID, err := getTransactionIDAndPractitionerIDFromVars(vars)
		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse(err.Error())
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		log.InfoR(req, fmt.Sprintf("start GET request for appointments resource with transaction ID: [%s] and practitioner ID: [%s]", transactionID, practitionerID))

		// Check if transaction is closed
		isTransactionClosed, err, httpStatus := service.CheckIfTransactionClosed(transactionID, req)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("error checking transaction status for [%v]: [%s]", transactionID, err))
			m := models.NewMessageResponse(fmt.Sprintf("error checking transaction status for [%v]: [%s]", transactionID, err))
			utils.WriteJSONWithStatus(w, req, m, httpStatus)
			return
		}
		if isTransactionClosed {
			log.ErrorR(req, fmt.Errorf("transaction [%v] is already closed and cannot be updated", transactionID))
			m := models.NewMessageResponse(fmt.Sprintf("transaction [%v] is already closed and cannot be updated", transactionID))
			utils.WriteJSONWithStatus(w, req, m, httpStatus)
			return
		}

		err, statusCode := svc.DeletePractitionerAppointment(transactionID, practitionerID)
		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse(err.Error())
			utils.WriteJSONWithStatus(w, req, m, statusCode)
			return
		}

		w.WriteHeader(statusCode)
	})
}

func getTransactionIDAndPractitionerIDFromVars(vars map[string]string) (transactionID string, practitionerID string, err error) {
	transactionID = utils.GetTransactionIDFromVars(vars)
	if transactionID == "" {
		err = fmt.Errorf("there is no Transaction ID in the URL path")
		return
	}

	// Check for a practitioner ID in request
	practitionerID = utils.GetPractitionerIDFromVars(vars)
	if practitionerID == "" {
		err = fmt.Errorf("there is no Practitioner ID in the URL path")
		return
	}
	return transactionID, practitionerID, nil
}
