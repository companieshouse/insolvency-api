package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/companieshouse/api-sdk-go/companieshouseapi"
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
func HandleCreateInsolvencyResource(svc dao.Service, helperService utils.HelperService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		// generate etag for request
		etag, err := helperService.GenerateEtag()
		if err != nil {
			log.Error(fmt.Errorf("error generating etag: [%s]", err))
			m := models.NewMessageResponse(fmt.Sprintf("error generating etag: [%s]", err))
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}

		// Check transaction id exists in path
		incomingTransactionId := utils.GetTransactionIDFromVars(mux.Vars(req))
		isValidTransactionId, transactionID := helperService.HandleTransactionIdExistsValidation(w, req, incomingTransactionId)
		if !isValidTransactionId {
			return
		}

		log.InfoR(req, fmt.Sprintf("start POST request for insolvency resource with transaction id: %s", transactionID))

		// Decode the incoming request to create a list of practitioners
		var request models.InsolvencyRequest
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

		// Check with company profile API if company exists
		var companyProfile *companieshouseapi.CompanyProfile
		err, httpStatus, companyProfile = service.CheckCompanyExists(&request, req)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("company was not found valid when checking company profile API [%v]", err))
			m := models.NewMessageResponse(fmt.Sprintf("company [%s] was not found valid for insolvency: %v", request.CompanyNumber, err))
			utils.WriteJSONWithStatus(w, req, m, httpStatus)
			return
		}

		// Check with alphakey service if company name valid
		err, httpStatus = service.CheckCompanyNameAlphaKey(companyProfile.CompanyName, &request, req)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("company was not found valid when checking company profile API [%v]", err))
			m := models.NewMessageResponse(fmt.Sprintf("company [%s] was not found valid for insolvency: %v", request.CompanyNumber, err))
			utils.WriteJSONWithStatus(w, req, m, httpStatus)
			return
		}

		// Check with company profile API if other details are valid
		err = service.CheckCompanyDetailsAreValid(companyProfile)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("company was not found valid when checking company profile API [%v]", err))
			m := models.NewMessageResponse(fmt.Sprintf("company [%s] was not found valid for insolvency: %v", request.CompanyNumber, err))
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// Add new insolvency resource to mongo
		insolvencyResourceDto := transformers.InsolvencyResourceRequestToDB(&request, transactionID)
		if insolvencyResourceDto == nil {
			m := models.NewMessageResponse(fmt.Sprintf("there was a problem handling your request for transaction id [%s]", transactionID))
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}

		insolvencyResourceDto.Data.Etag = etag
		httpStatus, err = svc.CreateInsolvencyResource(insolvencyResourceDto)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("failed to create insolvency resource in database for transaction [%s]: %v", transactionID, err))
			m := models.NewMessageResponse(fmt.Sprintf("there was a problem handling your request for transaction [%s]: %v", transactionID, err))
			utils.WriteJSONWithStatus(w, req, m, httpStatus)
			return
		}

		// Patch transaction API with new insolvency resource
		err, httpStatus = service.PatchTransactionWithInsolvencyResource(transactionID, insolvencyResourceDto, req)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("error patching transaction api with insolvency resource [%s]: [%v]", insolvencyResourceDto.Data.Links.Self, err))
			m := models.NewMessageResponse(fmt.Sprintf("error patching transaction api with insolvency resource [%s]: [%v]", insolvencyResourceDto.Data.Links.Self, err))
			utils.WriteJSONWithStatus(w, req, m, httpStatus)
			return
		}

		log.InfoR(req, fmt.Sprintf("successfully added insolvency resource with transaction ID: %s, to mongo", transactionID))

		utils.WriteJSONWithStatus(w, req, transformers.InsolvencyResourceDaoToCreatedResponse(insolvencyResourceDto), http.StatusCreated)
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

		insolvencyResource, practitionerResourceDao, err := svc.GetInsolvencyPractitionersResource(transactionID)
		if err != nil {
			// Check if insolvency case was not found
			if err.Error() == fmt.Sprintf("there was a problem handling your request for transaction [%s] - insolvency case not found", transactionID) {
				message := fmt.Sprintf("insolvency case with transactionID [%s] not found", transactionID)
				log.Info(message)
				m := models.NewMessageResponse(message)
				// Returning OK instead of NOT FOUND because endpoint is specifically for validation not for insolvency case
				utils.WriteJSONWithStatus(w, req, m, http.StatusOK)
				return
			}
			log.ErrorR(req, fmt.Errorf("error getting insolvency resource from DB: [%s]", err))
			m := models.NewMessageResponse("there was a problem handling your request")
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}

		validationErrors := service.ValidateInsolvencyDetails(insolvencyResource, practitionerResourceDao)
		antivirusValidationErrors := service.ValidateAntivirus(svc, insolvencyResource, req)

		// If antivirus check has failed, set case false and append antivirus validation error to existing validation errors
		if len(*antivirusValidationErrors) > 0 {
			*validationErrors = append(*validationErrors, *antivirusValidationErrors...)
		}

		isCaseValid := true
		if len(*validationErrors) > 0 {
			log.InfoR(req, fmt.Sprintf("case for transaction id [%s] was not found valid for submission for reason(s): [%v]", transactionID, *validationErrors))
			isCaseValid = false
		}

		log.InfoR(req, fmt.Sprintf("successfully finished GET request for validating insolvency resource with transaction id: %s", transactionID))

		m := models.NewValidationStatusResponse(isCaseValid, validationErrors)
		utils.WriteJSONWithStatus(w, req, m, http.StatusOK)
	})
}

// HandleGetFilings returns the resource in filings format for the filing-resource-handler to send to CHIPS
func HandleGetFilings(svc dao.Service) http.Handler {
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

		log.InfoR(req, fmt.Sprintf("start GET request for filings resource for transaction id: %s", transactionID))

		// Check if transaction is closed before generating filings
		isTransactionClosed, err, httpStatus := service.CheckIfTransactionClosed(transactionID, req)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("error checking transaction status for [%v]: [%s]", transactionID, err))
			m := models.NewMessageResponse(fmt.Sprintf("error checking transaction status for [%v]: [%s]", transactionID, err))
			utils.WriteJSONWithStatus(w, req, m, httpStatus)
			return
		}
		if !isTransactionClosed {
			log.ErrorR(req, fmt.Errorf("transaction [%v] is not closed so the filings cannot be generated", transactionID))
			m := models.NewMessageResponse(fmt.Sprintf("transaction [%v] is not closed so the filings cannot be generated", transactionID))
			utils.WriteJSONWithStatus(w, req, m, http.StatusForbidden)
			return
		}

		filings, err := service.GenerateFilings(svc, transactionID)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("error generating filings for [%v]: [%s]", transactionID, err))
			m := models.NewMessageResponse(fmt.Sprintf("error generating filings for [%v]: [%s]", transactionID, err))
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
		}

		log.InfoR(req, fmt.Sprintf("successfully finished GET request for filings resource for transaction id: %s", transactionID))

		utils.WriteJSONWithStatus(w, req, filings, http.StatusOK)
	})
}
