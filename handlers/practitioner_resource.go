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

		var insolvencyResource models.InsolvencyResourceDao
		var practitionerResourceDto models.PractitionerResourceDao
		//var practitionersMapResource map[string]string
		practitionersMapResource := make(map[string]string)

		practitionerID := utils.GenerateID()

		// generate etag for request
		etag, err := helperService.GenerateEtag()
		if err != nil {
			log.Error(fmt.Errorf("error generating etag: [%s]", err))
			m := models.NewMessageResponse(fmt.Sprintf("error generating etag: [%s]", err))
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
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
			logErrorAndHttpResponse(w, req, http.StatusInternalServerError, "error", []error{err, fmt.Errorf("failed to validate the practitioner request supplied")})
			return
		}

		if validationErrs != "" {
			logErrorAndHttpResponse(w, req, http.StatusBadRequest, "error", []error{fmt.Errorf("invalid request - failed validation on the following: %s", validationErrs),
				fmt.Errorf("invalid request body: " + validationErrs)})
			return
		}

		// Check if practitioner role supplied is valid
		if ok := constants.IsInRoleList(request.Role); !ok {
			logErrorAndHttpResponse(w, req, http.StatusBadRequest, "error", []error{fmt.Errorf("invalid practitioner role"), fmt.Errorf("the practitioner role supplied is not valid %s", request.Role)})
			return
		}

		practitionerDao := transformers.PractitionerResourceRequestToDB(&request, practitionerID, transactionID)
		practitionerDao.Data.Etag = etag
		practitionerDao.Data.Kind = "insolvency#practitioner"
		practitionerResourceDto = *practitionerDao
		practitionerResourceDto.Data.PractitionerId = practitionerID

		// GetInsolvencyResourceData retrieves previously stored practitioners
		insolvencyResourceDao, err := svc.GetInsolvencyResourceData(transactionID)
		if err != nil {
			logErrorAndHttpResponse(w, req, http.StatusInternalServerError, "error", []error{err})
			return
		}

		maxPractitioners := 5
		//check to ensure it is not nil from the collection
		if insolvencyResourceDao != nil && len(insolvencyResourceDao.Data.Practitioners) > 0 {
			err = json.Unmarshal([]byte(insolvencyResourceDao.Data.Practitioners), &practitionersMapResource)
			if err != nil {
				logErrorAndHttpResponse(w, req, http.StatusInternalServerError, "error", []error{fmt.Errorf("there was a problem handling json Unmarshalling %s", transactionID)})
				return
			}
		}

		// Check if there are already 5 practitioners in database
		if len(practitionersMapResource) >= maxPractitioners {
			err = fmt.Errorf("there was a problem handling your request for transaction %s already has 5 practitioners", transactionID)
			logErrorAndHttpResponse(w, req, http.StatusInternalServerError, "error", []error{err})
			return
		}

		// Check if practitioner is already assigned to this case
		extractedPractitionerIds := utils.ConvertMapToStringArray(practitionersMapResource)

		practitionerResourceDaos, err := svc.GetPractitionersAppointmentResource(extractedPractitionerIds, transactionID)
		for _, practitionerResourceDao := range practitionerResourceDaos {
			if err == nil && practitionerDao.Data.IPCode == practitionerResourceDao.Data.IPCode {
				logErrorAndHttpResponse(w, req, http.StatusBadRequest, "error", []error{fmt.Errorf("there was a problem handling your request for transaction %s - practitioner with IP Code %s already is already assigned to this case", transactionID, practitionerResourceDto.Data.IPCode)})
				return
			}
		}

		// Create new practitoner data to be stored
		practitionersMapResource[practitionerID] = fmt.Sprintf(constants.TransactionsPath + transactionID + constants.PractitionersPath + string(practitionerID))

		// Create new practitioner for the insolvency
		statusCode, err := svc.CreatePractitionerResource(&practitionerResourceDto, transactionID)
		if err != nil {
			logErrorAndHttpResponse(w, req, statusCode, "error", []error{err})
			return
		}

		// Prepare the format of saving the new practitioner plus already existed practitioners from insolvency collection
		stringPractitionerLinks, err := utils.ConvertMapToString(practitionersMapResource)
		if err != nil {
			logErrorAndHttpResponse(w, req, statusCode, "error", []error{fmt.Errorf("there was a problem handling unmarshaling insolvency practitioner with transactionId: %s ", transactionID), err})
			return
		}

		insolvencyResource.Data.Practitioners = stringPractitionerLinks

		// //Update the insolvency practitioner
		statusCode, err = svc.UpdateInsolvencyPractitioners(insolvencyResource, transactionID)
		if err != nil {
			logErrorAndHttpResponse(w, req, statusCode, "error", []error{fmt.Errorf("there was a problem handling your request for transaction %s", transactionID), err})
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
			logErrorAndHttpResponse(w, req, http.StatusBadRequest, "error", []error{fmt.Errorf("there is no transaction id in the url path"), fmt.Errorf("transaction id is not in the url path")})
			return
		}

		log.InfoR(req, fmt.Sprintf("start GET request for practitioners resource with transaction id: %s", transactionID))

		insolvencyResourceDao, err := svc.GetInsolvencyResourceData(transactionID)
		if err != nil {
			logErrorAndHttpResponse(w, req, http.StatusInternalServerError, "error", []error{err})
			return
		}
		if insolvencyResourceDao == nil {
			logErrorAndHttpResponse(w, req, http.StatusNotFound, "error", []error{fmt.Errorf("insolvency case for transaction %s not found", transactionID),
				fmt.Errorf("there was a problem handling your request for insolvency case with transaction ID: " + transactionID + " not found")})
			return
		}
		if insolvencyResourceDao != nil && len(insolvencyResourceDao.Data.Practitioners) == 0 {
			logErrorAndHttpResponse(w, req, http.StatusNotFound, "error", []error{fmt.Errorf("practitioners for insolvency case with transaction %s not found", transactionID),
				fmt.Errorf("there was a problem handling your request for insolvency case with transaction: " + transactionID + " there are no practitioners assigned to this case")})
			return
		}

		_, practitionerIds, err := utils.ConvertStringToMapObjectAndStringList(insolvencyResourceDao.Data.Practitioners)
		if err != nil {
			logErrorAndHttpResponse(w, req, http.StatusInternalServerError, "error", []error{err})
			return
		}

		practitionerResourceDaos, _ := svc.GetPractitionersAppointmentResource(practitionerIds, transactionID)

		for _, practitionerDao := range practitionerResourceDaos {
			practitionerResourceDao = append(practitionerResourceDao, practitionerDao)
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
			logErrorAndHttpResponse(w, req, http.StatusBadRequest, "error", []error{err})
			return
		}

		log.InfoR(req, fmt.Sprintf("start GET request for practitioner resource with transaction id: %s and practitioner id: %s", transactionID, practitionerID))

		// Get practitioner from DB
		practitionerResourceDaos, err := svc.GetPractitionersAppointmentResource([]string{practitionerID}, transactionID)
		if err != nil {
			logErrorAndHttpResponse(w, req, http.StatusInternalServerError, "error", []error{fmt.Errorf("failed to get practitioner with id [%s]: [%s]", practitionerID, err),
				fmt.Errorf("there was a problem handling your request")})
			return
		}
		fmt.Println("practitionerResourceDaos1232======>", practitionerResourceDaos)
		// Check if practitioner returned is empty
		if len(practitionerResourceDaos) == 0 {
			logErrorAndHttpResponse(w, req, http.StatusNotFound, "debug", []error{fmt.Errorf("practitioner with ID [%s] not found", practitionerID)})
			return
		}

		for _, practitionerDao := range practitionerResourceDaos {
			practitionerResourceDao = append(practitionerResourceDao, practitionerDao)
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
			logErrorAndHttpResponse(w, req, http.StatusBadRequest, "error", []error{err})
			return
		}

		log.InfoR(req, fmt.Sprintf("start DELETE request for practitioner resource with transaction id: %s and practitioner id: %s", transactionID, practitionerID))

		// Check if transaction is closed
		isTransactionClosed, err, httpStatus := service.CheckIfTransactionClosed(transactionID, req)
		if err != nil {
			logErrorAndHttpResponse(w, req, httpStatus, "error", []error{fmt.Errorf("error checking transaction status for [%v]: [%s]", transactionID, err)})
			return
		}
		if isTransactionClosed {
			logErrorAndHttpResponse(w, req, httpStatus, "error", []error{fmt.Errorf("transaction [%v] is already closed and cannot be updated", transactionID)})
			return
		}

		// Delete practitioner from Mongo
		statusCode, err := svc.DeletePractitioner(practitionerID, transactionID)
		if err != nil {
			logErrorAndHttpResponse(w, req, statusCode, "error", []error{err})
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
			logErrorAndHttpResponse(w, req, http.StatusBadRequest, "error", []error{fmt.Errorf("invalid appointment made_by"),
				fmt.Errorf("the appointment made_by supplied is not valid: [%s]", request.MadeBy)})
			return
		}

		// Validate all appointment details are of the correct format and criteria
		validationErrs, err := service.ValidateAppointmentDetails(svc, request, transactionID, practitionerID, req)
		if err != nil {
			logErrorAndHttpResponse(w, req, http.StatusInternalServerError, "error", []error{fmt.Errorf("failed to validate appointment details: [%s]", err),
				fmt.Errorf("there was a problem handling your request for transaction ID [%s]", transactionID)})
			return
		}

		if len(validationErrs) > 0 {
			logErrorAndHttpResponse(w, req, http.StatusBadRequest, "error", []error{fmt.Errorf("invalid request - failed validation on the following: %s",
				validationErrs), fmt.Errorf("invalid request body: " + strings.Join(validationErrs, ", "))})
			return
		}

		appointmentResourceDto := transformers.PractitionerAppointmentRequestToDB(&request, transactionID, practitionerID)
		appointmentResourceDto.Data.Etag = etag
		appointmentResourceDto.Data.Kind = "insolvency#appointment"

		// CreateAppointmentResource stores appointment in DB
		statusCode, err := svc.CreateAppointmentResource(&appointmentResourceDto)

		if err != nil {
			logErrorAndHttpResponse(w, req, statusCode, "error", []error{err})
			return
		}

		// Update practitioner with appointment in DB
		statusCode, err = svc.UpdatePractitionerAppointment(&appointmentResourceDto, transactionID, practitionerID)
		if err != nil {
			logErrorAndHttpResponse(w, req, statusCode, "error", []error{err})
			return
		}

		log.InfoR(req, fmt.Sprintf("successfully added practitioner appointment with transaction ID [%s] and practitioner ID [%s] to mongo", transactionID, practitionerID))

		appointmentResourceDao, err := svc.GetPractitionerAppointment(practitionerID, transactionID)
		if err != nil {
			logErrorAndHttpResponse(w, req, http.StatusInternalServerError, "error", []error{err})
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
			logErrorAndHttpResponse(w, req, http.StatusBadRequest, "error", []error{err})
			return
		}

		log.InfoR(req, fmt.Sprintf("start GET request for appointments resource with transaction ID: [%s] and practitioner ID: [%s]", transactionID, practitionerID))

		practitionerResourceDtos, err := svc.GetPractitionersAppointmentResource([]string{practitionerID}, transactionID)
		if err != nil {
			logErrorAndHttpResponse(w, req, http.StatusInternalServerError, "error", []error{err})
			return
		}

		// Check if practitioner is empty (not found).
		if len(practitionerResourceDtos) == 0 {
			logErrorAndHttpResponse(w, req, http.StatusInternalServerError, "info", []error{fmt.Errorf("practitionerID [%s] not found for transactionID [%s]", practitionerID, transactionID)})
			return
		}

		// Check if practitioner has an appointment
		appointment := practitionerResourceDtos[0].Data.Appointment
		if appointment == nil {
			logErrorAndHttpResponse(w, req, http.StatusInternalServerError, "info", []error{fmt.Errorf("no appointment found for practitionerID [%s] an transactionID [%s]", practitionerID, transactionID)})
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
			logErrorAndHttpResponse(w, req, http.StatusBadRequest, "error", []error{err})
			return
		}

		log.InfoR(req, fmt.Sprintf("start GET request for appointments resource with transaction ID: [%s] and practitioner ID: [%s]", transactionID, practitionerID))

		// Check if transaction is closed
		isTransactionClosed, err, httpStatus := service.CheckIfTransactionClosed(transactionID, req)
		if err != nil {
			logErrorAndHttpResponse(w, req, httpStatus, "error", []error{fmt.Errorf("error checking transaction status for [%v]: [%s]", transactionID, err)})
			return
		}
		if isTransactionClosed {
			logErrorAndHttpResponse(w, req, httpStatus, "error", []error{fmt.Errorf("transaction [%v] is already closed and cannot be updated", transactionID)})
			return
		}

		statusCode, err := svc.DeletePractitionerAppointment(transactionID, practitionerID)
		if err != nil {
			logErrorAndHttpResponse(w, req, statusCode, "error", []error{err})
			return
		}

		w.WriteHeader(statusCode)
	})
}

// logErrorAndHttpResponse logs error and write to http
func logErrorAndHttpResponse(w http.ResponseWriter, req *http.Request, status int, kind string, errs []error) {
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
