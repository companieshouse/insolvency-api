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

// HandleCreateInsolvencyResource creates an insolvency resource
func HandleCreateInsolvencyResource(svc dao.Service) http.Handler {
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

		log.InfoR(req, fmt.Sprintf("start POST request for insolvency resource with transaction id: %s", transactionID))

		// Decode the incoming request to create an insolvency resource
		var request models.InsolvencyRequest
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

		// Check case type of incoming request is CVL
		if !(request.CaseType == constants.CVL.String()) {
			log.ErrorR(req, fmt.Errorf("only creditors-voluntary-liquidation can be filed"))
			m := models.NewMessageResponse(fmt.Sprintf("case type is not creditors-voluntary-liquidation for transaction %s", transactionID))
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// Check with transaction API that provided transaction ID exists
		err, httpStatus := service.CheckTransactionID(transactionID, req)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("transaction id [%s] was not found valid for insolvency request against company [%s] when checking transaction api: [%v]",
				transactionID, request.CompanyNumber, err))
			m := models.NewMessageResponse(fmt.Sprintf("transaction id [%s] was not found valid for insolvency: %v", transactionID, err))
			utils.WriteJSONWithStatus(w, req, m, httpStatus)
			return
		}

		// Check with company profile API if company is valid
		err, httpStatus = service.CheckCompanyInsolvencyValid(&request, req)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("company was not found valid when checking company profile API [%v]", err))
			m := models.NewMessageResponse(fmt.Sprintf("company [%s] was not found valid for insolvency: %v", request.CompanyNumber, err))
			utils.WriteJSONWithStatus(w, req, m, httpStatus)
			return
		}

		// Add new insolvency resource to mongo
		model := transformers.InsolvencyResourceRequestToDB(&request, transactionID)

		err, httpStatus = svc.CreateInsolvencyResource(model)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("failed to create insolvency resource in database for transaction [%s]: %v", transactionID, err))
			m := models.NewMessageResponse(fmt.Sprintf("there was a problem handling your request for transaction [%s]: %v", transactionID, err))
			utils.WriteJSONWithStatus(w, req, m, httpStatus)
			return
		}

		// Patch transaction API with new insolvency resource
		err, httpStatus = service.PatchTransactionWithInsolvencyResource(transactionID, model, req)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("error patching transaction api with insolvency resource [%s]: [%v]", model.Links.Self, err))
			m := models.NewMessageResponse(fmt.Sprintf("error patching transaction api with insolvency resource [%s]: [%v]", model.Links.Self, err))
			utils.WriteJSONWithStatus(w, req, m, httpStatus)
			return
		}

		log.InfoR(req, fmt.Sprintf("successfully added insolvency resource with transaction ID: %s, to mongo", transactionID))

		utils.WriteJSONWithStatus(w, req, transformers.InsolvencyResourceDaoToCreatedResponse(model), http.StatusCreated)
	})
}

// HandleSubmitAttachment receives an attachment to be stored against the Insolvency case
func HandleSubmitAttachment(svc dao.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		transactionID := utils.GetTransactionIDFromVars(vars)
		if transactionID == "" {
			log.ErrorR(req, fmt.Errorf("there is no transaction ID in the URL path"))
			m := models.NewMessageResponse("transaction ID is not in the URL path")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		log.InfoR(req, fmt.Sprintf("start POST request for submit attachment with transaction id: %s", transactionID))

		attachmentType := req.FormValue("attachment_type")

		file, header, err := req.FormFile("file")
		if err != nil {
			log.ErrorR(req, fmt.Errorf("error reading form from request: %s", err))
			m := models.NewMessageResponse("error reading form from request")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// Validate that the provided attachment details are correct
		if errs := service.ValidateAttachmentDetails(attachmentType); errs != "" {
			log.ErrorR(req, fmt.Errorf("invalid request - failed validation on the following: %s", errs))
			m := models.NewMessageResponse("invalid request: " + errs)
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		fileID, responseType, err := service.UploadAttachment(file, header, req)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("error uploading attachment: [%v]", err), log.Data{"service_response_type": responseType.String()})

			status, err := utils.ResponseTypeToStatus(responseType.String())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(status)
			return
		}
		if responseType != service.Success {
			status, err := utils.ResponseTypeToStatus(responseType.String())
			if err != nil {
				log.ErrorR(req, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(status)
			return
		}

		attachmentDao, err := svc.AddAttachmentToInsolvencyResource(transactionID, fileID, attachmentType)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("failed to add attachment to insolvency resource in db for transaction [%s]: %v", transactionID, err))
			m := models.NewMessageResponse("there was a problem handling your request")
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}

		attachmentResponse, err := transformers.AttachmentResourceDaoToResponse(attachmentDao, header)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("error transforming dao to response: [%s]", err))
			m := models.NewMessageResponse("there was a problem handling your request")
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
		}

		utils.WriteJSONWithStatus(w, req, attachmentResponse, http.StatusCreated)
	})
}

// HandleGetValidationStatus returns whether a created insolvency case is acceptable to be closed by the transaction API
func HandleGetValidationStatus(svc dao.Service) http.Handler {
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

		log.InfoR(req, fmt.Sprintf("start GET request for validating insolvency resource with transaction id: %s", transactionID))

		isCaseValid, validationErrors := service.ValidateInsolvencyDetails(svc, transactionID)
		if !isCaseValid {
			log.ErrorR(req, fmt.Errorf("case for transaction id [%s] was not found valid for submission for reason(s): [%v]", transactionID, validationErrors))
		}

		log.InfoR(req, fmt.Sprintf("successfully finished GET request for validating insolvency resource with transaction id: %s", transactionID))

		m := models.NewValidationStatusResponse(isCaseValid, validationErrors)
		utils.WriteJSONWithStatus(w, req, m, http.StatusOK)
	})
}
