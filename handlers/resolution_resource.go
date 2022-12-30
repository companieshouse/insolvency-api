package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/dao"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/service"
	"github.com/companieshouse/insolvency-api/transformers"
	"github.com/companieshouse/insolvency-api/utils"
	"github.com/gorilla/mux"
)

// HandleCreateResolution receives a resolution to be stored against the Insolvency case
func HandleCreateResolution(svc dao.Service, helperService utils.HelperService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		// Check transaction id exists in path
		incomingTransactionId := utils.GetTransactionIDFromVars(mux.Vars(req))
		transactionID, isValidTransactionId, httpStatusCode := helperService.HandleTransactionIdExistsValidation(w, req, incomingTransactionId)
		if !isValidTransactionId {
			http.Error(w, "Bad request", httpStatusCode)
			return
		}

		log.InfoR(req, fmt.Sprintf("start POST request for submit resolution with transaction id: %s", transactionID))

		// Check if transaction is closed
		isTransactionClosed, err, httpStatus := service.CheckIfTransactionClosed(transactionID, req)
		isValidTransactionNotClosed, httpStatusCode, _ := helperService.HandleTransactionNotClosedValidation(w, req, transactionID, isTransactionClosed, httpStatus, err)
		if !isValidTransactionNotClosed {
			http.Error(w, "Transaction closed", httpStatusCode)
			return
		}

		// Decode Request body
		var request models.Resolution
		err = json.NewDecoder(req.Body).Decode(&request)
		isValidDecoded, httpStatusCode := helperService.HandleBodyDecodedValidation(w, req, transactionID, err)
		if !isValidDecoded {
			http.Error(w, fmt.Sprintf("failed to read request body for transaction %s", transactionID), httpStatusCode)
			return
		}

		resolutionDao := transformers.ResolutionResourceRequestToDB(&request, transactionID, helperService)

		// Validate all mandatory fields
		errs := utils.Validate(request)
		isValidMarshallToDB, httpStatusCode := helperService.HandleMandatoryFieldValidation(w, req, errs)
		if !isValidMarshallToDB {
			http.Error(w, errs, httpStatusCode)
			return
		}

		// Validate the provided statement details are in the correct format
		if errs := service.ValidateResolutionRequest(request); errs != "" {
			log.ErrorR(req, fmt.Errorf("invalid request - failed validation on the following: %s", errs))
			m := models.NewMessageResponse("invalid request body: " + errs)
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// Validate the provided resolution date is in the correct format
		validationErrs, err := service.ValidateResolutionDate(svc, resolutionDao, transactionID, req)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("failed to validate resolution: [%s]", err))
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

		// Validate if supplied attachment matches attachments associated with supplied transactionID in mongo db
		attachment, err := svc.GetAttachmentFromInsolvencyResource(transactionID, resolutionDao.Attachments[0])
		isValidAttachment, httpStatusCode := helperService.HandleAttachmentValidation(w, req, transactionID, attachment, err)
		if !isValidAttachment {
			http.Error(w, "attachment not found on transaction", httpStatusCode)
			return
		}

		// Validate the supplied attachment is a valid type
		if attachment.Type != "resolution" {
			err := fmt.Errorf("attachment id [%s] is an invalid type for this request: %v", resolutionDao.Attachments[0], attachment.Type)
			responseMessage := "attachment is not a resolution"

			httpStatusCode := helperService.HandleAttachmentTypeValidation(w, req, responseMessage, err)
			http.Error(w, responseMessage, httpStatusCode)
			return
		}

		// Creates the statement of affairs resource in mongo if all previous checks pass
		statusCode, err := svc.CreateResolutionResource(resolutionDao, transactionID)
		isValidCreateResource, httpStatusCode := helperService.HandleCreateResourceValidation(w, req, statusCode, err)
		if !isValidCreateResource {
			http.Error(w, "Server error", httpStatusCode)
			return
		}

		daoResponse := transformers.ResolutionDaoToResponse(resolutionDao)

		log.InfoR(req, fmt.Sprintf("successfully added resolution resource with transaction ID: %s, to mongo", transactionID))

		utils.WriteJSONWithStatus(w, req, daoResponse, http.StatusOK)
	})
}

// HandleGetResolution retrieves a resolution stored against the Insolvency Case
func HandleGetResolution(svc dao.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		transactionID := utils.GetTransactionIDFromVars(vars)
		if transactionID == "" {
			log.ErrorR(req, fmt.Errorf("there is no transaction ID in the URL path"))
			m := models.NewMessageResponse("transaction ID is not in the URL path")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		log.InfoR(req, fmt.Sprintf("start GET request for get resolution with transaction id: %s", transactionID))

		resolution, err := svc.GetResolutionResource(transactionID)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("failed to get resolution from insolvency resource in db for transaction [%s]: %v", transactionID, err))
			m := models.NewMessageResponse("there was a problem handling your request")
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}
		if resolution.DateOfResolution == "" {
			m := models.NewMessageResponse("resolution not found on transaction")
			utils.WriteJSONWithStatus(w, req, m, http.StatusNotFound)
			return
		}

		log.InfoR(req, fmt.Sprintf("successfully retrieved resolution resource with transaction ID: %s, from mongo", transactionID))

		utils.WriteJSONWithStatus(w, req, resolution, http.StatusOK)
	})
}

// HandleDeleteResolution deletes a resolution stored against the Insolvency Case
func HandleDeleteResolution(svc dao.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		vars := mux.Vars(req)
		transactionID := utils.GetTransactionIDFromVars(vars)

		if transactionID == "" {
			log.ErrorR(req, fmt.Errorf("there is no transaction ID in the URL path"))
			m := models.NewMessageResponse("transaction ID is not in the URL path")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		log.InfoR(req, fmt.Sprintf("start DELETE request for get resolution with transaction id: %s", transactionID))

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

		// Delete resolution from Mongo
		statusCode, err := svc.DeleteResolutionResource(transactionID)
		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse(err.Error())
			utils.WriteJSONWithStatus(w, req, m, statusCode)
			return
		}

		log.InfoR(req, fmt.Sprintf("successfully deleted resolution with transaction ID: %s from mongo", transactionID))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
	})
}
