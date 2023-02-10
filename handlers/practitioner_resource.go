package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
		var insolvencyResource models.InsolvencyResourceDto
		// generate etag for request
		etag, err := helperService.GenerateEtag()
		if err != nil {
			log.Error(fmt.Errorf("error generating etag: [%s]", err))
			return
		}

		// Check transaction is valid
		transactionID, isValidTransaction := utils.ValidateTransaction(helperService, req, w, "practitioners", service.CheckIfTransactionClosed)
		if !isValidTransaction {
			return
		}

		// Decode the incoming request to create a list of practitioners
		var request models.PractitionerRequest
		err = json.NewDecoder(req.Body).Decode(&request)
		isValidDecoded := helperService.HandleBodyDecodedValidation(w, req, transactionID, err)
		if !isValidDecoded {
			return
		}

		// Validate all mandatory fields
		errs := utils.Validate(request)
		isValidMarshallToDB := helperService.HandleMandatoryFieldValidation(w, req, errs)
		if !isValidMarshallToDB {
			return
		}

		// Validates that the provided practitioner details are in the correct format
		validationErrs, err := service.ValidatePractitionerDetails(svc, transactionID, request)
		if err != nil {
			logError(w, req, http.StatusInternalServerError, "error", []error{err, fmt.Errorf("failed to validate the practitioner request supplied")})
			return
		}

		if validationErrs != "" {
			logError(w, req, http.StatusBadRequest, "error", []error{fmt.Errorf("invalid request - failed validation on the following: %s", validationErrs),
				fmt.Errorf("invalid request body: " + validationErrs)})
			return
		}

		// Check if practitioner role supplied is valid
		if ok := constants.IsInRoleList(request.Role); !ok {
			logError(w, req, http.StatusBadRequest, "error", []error{fmt.Errorf("invalid practitioner role"), fmt.Errorf("the practitioner role supplied is not valid %s", request.Role)})
			return
		}

		practitionerDao := transformers.PractitionerResourceRequestToDB(&request, transactionID)
		practitionerDao.Etag = etag
		practitionerDao.Kind = "insolvency#practitioner"

		// Store practitioners resource in practitioners collection
		practionersResource, statusCode, err := svc.CreatePractitionerResourceForInsolvencyCase(transactionID)
		if err != nil {
			logError(w, req, statusCode, "error", []error{err})
			return
		}

		// Check if practitioner is already assigned to this case
		extractedPractitionerIds := utils.ConvertMapToStringArray(practionersResource)

		practitionerResourceDtos, err := svc.GetPractitionersByIds(extractedPractitionerIds, transactionID)
		for _, practitionerResourceDto := range practitionerResourceDtos {
			if err == nil && practitionerDao.IPCode == practitionerResourceDto.Data.IPCode {
				logError(w, req, statusCode, "error", []error{fmt.Errorf("there was a problem handling your request for transaction %s - practitioner with IP Code %s already is already assigned to this case", transactionID, practitionerResourceDto.Data.IPCode)})
				return
			}
		}

		// Create new practitoner data to be stored
		practionersResource[practitionerDao.ID] = fmt.Sprintf(constants.TransactionsPath + transactionID + constants.PractitionersPath + string(practitionerDao.ID))

		// Create new practitioner for the insolvency
		statusCode, err = svc.CreatePractitionerResource(practitionerDao, transactionID)
		if err != nil {
			logError(w, req, statusCode, "error", []error{err})
			return
		}

		// Prepare the format of saving the new practitioner plus already existed practitioners from insolvency collection
		stringPractitionerLinks, err := utils.ConvertMapToString(practionersResource)
		if err != nil {
			logError(w, req, statusCode, "error", []error{fmt.Errorf("there was a problem handling unmarshaling insolvency practitioner with transactionId: %s ", transactionID), err})
			return
		}

		insolvencyResource.Data.Practitioners = stringPractitionerLinks

		// //Update the insolvency practitioner
		_, err = svc.UpdateInsolvencyPractitioners(insolvencyResource, transactionID)
		if err != nil {
			logError(w, req, statusCode, "error", []error{fmt.Errorf("there was a problem handling your request for transaction %s", transactionID), err})
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
		var practitionerResourceDao []models.PractitionerResourceDao
		// Check for a transaction id in request
		vars := mux.Vars(req)
		transactionID := utils.GetTransactionIDFromVars(vars)
		if transactionID == "" {
			logError(w, req, http.StatusBadRequest, "error", []error{fmt.Errorf("there is no transaction id in the url path"), fmt.Errorf("transaction id is not in the url path")})
			return
		}

		log.InfoR(req, fmt.Sprintf("start GET request for practitioners resource with transaction id: %s", transactionID))

		practitionerResources, err := svc.GetInsolvencyPractitionerByTransactionID(transactionID)
		if err != nil {
			logError(w, req, http.StatusInternalServerError, "error", []error{err})
			return
		}
		if len(practitionerResources) == 0 {
			logError(w, req, http.StatusNotFound, "error", []error{fmt.Errorf("insolvency case for transaction %s not found", transactionID),
				fmt.Errorf("there was a problem handling your request for insolvency case with transaction ID: " + transactionID + " not found")})
			return
		}
		if len(practitionerResources) == 0 {
			logError(w, req, http.StatusNotFound, "error", []error{fmt.Errorf("practitioners for insolvency case with transaction %s not found", transactionID),
				fmt.Errorf("there was a problem handling your request for insolvency case with transaction: " + transactionID + " there are no practitioners assigned to this case")})
			return
		}

		_, practitionerIds, err := utils.ConvertStringToMap(practitionerResources)
		if err != nil {
			logError(w, req, http.StatusInternalServerError, "error", []error{err})
			return
		}

		practitionerResourceDtos, _ := svc.GetPractitionersByIds(practitionerIds, transactionID)

		for _, practitionerResourceDto := range practitionerResourceDtos {
			practitionerResourceDao = append(practitionerResourceDao, practitionerResourceDto.Data)
		}

		pra := transformers.PractitionerResourceDaoListToCreatedResponseList(practitionerResourceDao)
		utils.WriteJSONWithStatus(w, req, pra, http.StatusOK)
	})
}

// HandleGetPractitionerResource retrieves a practitioner with the specified practitionerID
// on the insolvency case with the specified transactionID
func HandleGetPractitionerResource(svc dao.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var practitionerResourceDao []models.PractitionerResourceDao
		// Check for a transaction id in request
		vars := mux.Vars(req)
		transactionID, practitionerID, err := getTransactionIDAndPractitionerIDFromVars(vars)
		if err != nil {
			logError(w, req, http.StatusBadRequest, "error", []error{err})
			return
		}

		log.InfoR(req, fmt.Sprintf("start GET request for practitioner resource with transaction id: %s and practitioner id: %s", transactionID, practitionerID))

		// Get practitioner from DB
		practitionerResourceDtos, err := svc.GetPractitionersByIds([]string{practitionerID}, transactionID)
		if err != nil {
			logError(w, req, http.StatusInternalServerError, "error", []error{fmt.Errorf("failed to get practitioner with id [%s]: [%s]", practitionerID, err),
				fmt.Errorf("there was a problem handling your request")})
			return
		}

		// Check if practitioner returned is empty
		if len(practitionerResourceDtos) == 0 {
			logError(w, req, http.StatusNotFound, "debug", []error{fmt.Errorf("practitioner with ID [%s] not found", practitionerID)})
			return
		}

		for _, practitionerResourceDto := range practitionerResourceDtos {
			practitionerResourceDao = append(practitionerResourceDao, practitionerResourceDto.Data)
		}

		// Successfully retrieved practitioner
		utils.WriteJSONWithStatus(w, req, transformers.PractitionerResourceDaoToCreatedResponse(&practitionerResourceDao[0]), http.StatusOK)
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
			logError(w, req, http.StatusBadRequest, "error", []error{err})
			return
		}

		log.InfoR(req, fmt.Sprintf("start DELETE request for practitioner resource with transaction id: %s and practitioner id: %s", transactionID, practitionerID))

		// Check if transaction is closed
		isTransactionClosed, err, httpStatus := service.CheckIfTransactionClosed(transactionID, req)
		if err != nil {
			logError(w, req, httpStatus, "error", []error{fmt.Errorf("error checking transaction status for [%v]: [%s]", transactionID, err)})
			return
		}
		if isTransactionClosed {
			logError(w, req, httpStatus, "error", []error{fmt.Errorf("transaction [%v] is already closed and cannot be updated", transactionID)})
			return
		}

		// Delete practitioner from Mongo
		statusCode, err := svc.DeletePractitioner(practitionerID, transactionID)
		if err != nil {
			logError(w, req, statusCode, "error", []error{err})
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
		var appointmentResponse models.AppointedPractitionerResource
		// generate etag for request
		etag, err := helperService.GenerateEtag()
		if err != nil {
			log.Error(fmt.Errorf("error generating etag: [%s]", err))
			return
		}

		// Check transaction id exists in path
		incomingTransactionId, practitionerID, _ := getTransactionIDAndPractitionerIDFromVars(mux.Vars(req))
		isValidTransactionId, transactionID := helperService.HandleTransactionIdExistsValidation(w, req, incomingTransactionId)
		if !isValidTransactionId {
			return
		}

		log.InfoR(req, fmt.Sprintf("start POST request for practitioner appointment with transaction ID: [%s] and practitioner ID: [%s]", transactionID, practitionerID))

		// Check if transaction is closed
		isTransactionClosed, err, httpStatus := service.CheckIfTransactionClosed(transactionID, req)
		isValidTransactionNotClosed := helperService.HandleTransactionNotClosedValidation(w, req, transactionID, isTransactionClosed, httpStatus, err)
		if !isValidTransactionNotClosed {
			return
		}

		// Decode the incoming request to create a list of practitioners
		var request models.PractitionerAppointment
		err = json.NewDecoder(req.Body).Decode(&request)
		isValidDecoded := helperService.HandleBodyDecodedValidation(w, req, transactionID, err)
		if !isValidDecoded {
			return
		}

		// Validate all mandatory fields
		errs := utils.Validate(request)
		isValidMarshallToDB := helperService.HandleMandatoryFieldValidation(w, req, errs)
		if !isValidMarshallToDB {
			return
		}

		// Check if made_by supplied is valid
		if ok := constants.IsAppointmentMadeByInList(request.MadeBy); !ok {
			logError(w, req, http.StatusBadRequest, "error", []error{fmt.Errorf("invalid appointment made_by"),
				fmt.Errorf("the appointment made_by supplied is not valid: [%s]", request.MadeBy)})
			return
		}

		// Validate all appointment details are of the correct format and criteria
		validationErrs, err := service.ValidateAppointmentDetails(svc, request, transactionID, practitionerID, req)
		if err != nil {
			logError(w, req, http.StatusInternalServerError, "error", []error{fmt.Errorf("failed to validate appointment details: [%s]", err),
				fmt.Errorf("there was a problem handling your request for transaction ID [%s]", transactionID)})
			return
		}

		if len(validationErrs) > 0 {
			logError(w, req, http.StatusBadRequest, "error", []error{fmt.Errorf("invalid request - failed validation on the following: %s",
				validationErrs), fmt.Errorf("invalid request body: " + strings.Join(validationErrs, ", "))})
			return
		}

		practitionerAppointmentDao := transformers.PractitionerAppointmentRequestToDB(&request, transactionID, practitionerID)
		practitionerAppointmentDao.Etag = etag
		practitionerAppointmentDao.Kind = "insolvency#appointment"

		// Store appointment in DB
		statusCode, err := svc.CreateAppointmentResource(practitionerAppointmentDao)
		if err != nil {
			logError(w, req, statusCode, "error", []error{err})
			return
		}

		// Update insolvency with practitioner appointment in DB
		statusCode, err = svc.UpdateInsolvencyPractitionerAppointment(practitionerAppointmentDao, transactionID, practitionerID)
		if err != nil {
			logError(w, req, statusCode, "error", []error{err})
			return
		}

		// Update practitioner with appointment in DB
		statusCode, err = svc.UpdatePractitionerAppointment(practitionerAppointmentDao, transactionID, practitionerID)
		if err != nil {
			logError(w, req, statusCode, "error", []error{err})
			return
		}

		log.InfoR(req, fmt.Sprintf("successfully added practitioner appointment with transaction ID [%s] and practitioner ID [%s] to mongo", transactionID, practitionerID))

		appointmentResourceDao, err := svc.GetPractitionerAppointment(practitionerID, transactionID)
		if err != nil {
			logError(w, req, http.StatusInternalServerError, "error", []error{err})
			return
		}

		if appointmentResourceDao != nil {
			appointmentResponse = transformers.PractitionerAppointmentDaoToResponse(appointmentResourceDao)
		}

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
			logError(w, req, http.StatusBadRequest, "error", []error{err})
			return
		}

		log.InfoR(req, fmt.Sprintf("start GET request for appointments resource with transaction ID: [%s] and practitioner ID: [%s]", transactionID, practitionerID))

		practitionerResourceDtos, err := svc.GetPractitionersByIds([]string{practitionerID}, transactionID)
		if err != nil {
			logError(w, req, http.StatusInternalServerError, "error", []error{err})
			return
		}

		// Check if practitioner is empty (not found).
		if len(practitionerResourceDtos) == 0 {
			logError(w, req, http.StatusInternalServerError, "info", []error{fmt.Errorf("practitionerID [%s] not found for transactionID [%s]", practitionerID, transactionID)})
			return
		}

		// Check if practitioner has an appointment
		appointment := practitionerResourceDtos[0].Data.Appointment
		if appointment == nil {
			logError(w, req, http.StatusInternalServerError, "info", []error{fmt.Errorf("no appointment found for practitionerID [%s] an transactionID [%s]", practitionerID, transactionID)})
			return
		}

		appointmentResponse := transformers.PractitionerAppointmentDaoToResponse(appointment)

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
			logError(w, req, http.StatusBadRequest, "error", []error{err})
			return
		}

		log.InfoR(req, fmt.Sprintf("start GET request for appointments resource with transaction ID: [%s] and practitioner ID: [%s]", transactionID, practitionerID))

		// Check if transaction is closed
		isTransactionClosed, err, httpStatus := service.CheckIfTransactionClosed(transactionID, req)
		if err != nil {
			logError(w, req, httpStatus, "error", []error{fmt.Errorf("error checking transaction status for [%v]: [%s]", transactionID, err)})
			return
		}
		if isTransactionClosed {
			logError(w, req, httpStatus, "error", []error{fmt.Errorf("transaction [%v] is already closed and cannot be updated", transactionID)})
			return
		}

		statusCode, err := svc.DeletePractitionerAppointment(transactionID, practitionerID)
		if err != nil {
			logError(w, req, statusCode, "error", []error{err})
			return
		}

		w.WriteHeader(statusCode)
	})
}

func logError(w http.ResponseWriter, req *http.Request, status int, kind string, errs []error) {
	switch kind {
	case "error":
		log.ErrorR(req, errs[0])
	case "info":
		log.ErrorR(req, errs[0])
	case "debug":
		log.Debug(errs[0].Error())
	}

	m := models.NewMessageResponse(errs[0].Error())
	if len(errs) > 1 {
		m = models.NewMessageResponse(errs[1].Error())
	}

	utils.WriteJSONWithStatus(w, req, m, status)
}

func getTransactionIDAndPractitionerIDFromVars(vars map[string]string) (transactionID string, practitionerID string, err error) {
	transactionID = utils.GetTransactionIDFromVars(vars)
	if transactionID == "" {
		err = fmt.Errorf("there is no Transaction ID in the URL path")
		return "", "", err
	}

	// Check for a practitioner ID in request
	practitionerID = utils.GetPractitionerIDFromVars(vars)
	if practitionerID == "" {
		err = fmt.Errorf("there is no Practitioner ID in the URL path")
		return "", "", err
	}
	return transactionID, practitionerID, nil
}
