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

// HandleCreateStatementOfAffairs receives a statement of affairs to be stored against the Insolvency case
func HandleCreateStatementOfAffairs(svc dao.Service, helperService utils.HelperService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		// Check transaction is valid
		transactionID, isValidTransaction := utils.ValidateTransaction(helperService, req, w, "statement of affairs", service.CheckIfTransactionClosed)
		if !isValidTransaction {
			return
		}

		// Decode Request body
		var request models.StatementOfAffairs
		err := json.NewDecoder(req.Body).Decode(&request)
		isValidDecoded := helperService.HandleBodyDecodedValidation(w, req, transactionID, err)
		if !isValidDecoded {
			return
		}

		statementDao := transformers.StatementOfAffairsResourceRequestToDB(&request)

		// Validate all mandatory fields
		errs := utils.Validate(request)
		isValidMarshallToDB := helperService.HandleMandatoryFieldValidation(w, req, errs)
		if !isValidMarshallToDB {
			return
		}

		// Validate the provided statement details are in the correct format
		validationErrs, err := service.ValidateStatementDetails(svc, statementDao, transactionID, req)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("failed to validate statement of affairs: [%s]", err))
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
		attachment, err := svc.GetAttachmentFromInsolvencyResource(transactionID, statementDao.Attachments[0])
		isValidAttachment := helperService.HandleAttachmentValidation(w, req, transactionID, attachment, err)
		if !isValidAttachment {
			return
		}

		// Validate the supplied attachment is a valid type
		if attachment.Type != "statement-of-affairs-director" && attachment.Type != "statement-of-affairs-liquidator" {
			err := fmt.Errorf("attachment id [%s] is an invalid type for this request: %v", statementDao.Attachments[0], attachment.Type)
			responseMessage := "attachment is not a statement-of-affairs-director or statement-of-affairs-liquidator"

			helperService.HandleAttachmentTypeValidation(w, req, responseMessage, err)
			return
		}

		// Creates the statement of affairs resource in mongo if all previous checks pass
		statusCode, err := svc.CreateStatementOfAffairsResource(statementDao, transactionID)
		isValidCreateResource := helperService.HandleCreateResourceValidation(w, req, statusCode, err)
		if !isValidCreateResource {
			return
		}

		daoResponse := transformers.StatementOfAffairsDaoToResponse(statementDao)

		log.InfoR(req, fmt.Sprintf("successfully added statement of affairs resource with transaction ID: %s, to mongo", transactionID))

		utils.WriteJSONWithStatus(w, req, daoResponse, http.StatusCreated)
	})
}

// HandleGetStatementOfAffairs retrieves a statement of affairs stored against the Insolvency Case
func HandleGetStatementOfAffairs(svc dao.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		transactionID := utils.GetTransactionIDFromVars(vars)
		if transactionID == "" {
			log.ErrorR(req, fmt.Errorf("there is no transaction ID in the URL path"))
			m := models.NewMessageResponse("transaction ID is not in the URL path")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		log.InfoR(req, fmt.Sprintf("start GET request for get statement of affairs with transaction id: %s", transactionID))

		statementOfAffairs, err := svc.GetStatementOfAffairsResource(transactionID)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("failed to get statement of affairs from insolvency resource in db for transaction [%s]: %v", transactionID, err))
			m := models.NewMessageResponse("there was a problem handling your request")
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}
		if statementOfAffairs.StatementDate == "" {
			m := models.NewMessageResponse(fmt.Sprintf("statement of affairs not found on transaction with ID: [%s]", transactionID))
			utils.WriteJSONWithStatus(w, req, m, http.StatusNotFound)
			return
		}

		log.InfoR(req, fmt.Sprintf("successfully retrieved statement of affairs resource with transaction ID: %s, from mongo", transactionID))

		utils.WriteJSONWithStatus(w, req, statementOfAffairs, http.StatusOK)
	})
}

// HandleDeleteStatementOfAffairs deletes a statement of affairs resource from an insolvency case
func HandleDeleteStatementOfAffairs(svc dao.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		transactionID := utils.GetTransactionIDFromVars(vars)
		if transactionID == "" {
			log.ErrorR(req, fmt.Errorf("there is no transaction ID in the URL path"))
			m := models.NewMessageResponse("transaction ID is not in the URL path")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		log.InfoR(req, fmt.Sprintf("start DELETE request for submit statement of affairs with transaction id: %s", transactionID))

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

		// Delete SOA from DB
		statusCode, err := svc.DeleteStatementOfAffairsResource(transactionID)
		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse(err.Error())
			utils.WriteJSONWithStatus(w, req, m, statusCode)
			return
		}

		log.InfoR(req, fmt.Sprintf("successfully deleted statement of affairs from insolvency case with transaction ID: %s", transactionID))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
	})
}
