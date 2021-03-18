package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/dao"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/transformers"
	"github.com/companieshouse/insolvency-api/utils"
	"github.com/gorilla/mux"
	"gopkg.in/go-playground/validator.v9"
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

		// Check all required fields are populated
		v := validator.New()
		if v.Struct(request) != nil {
			log.ErrorR(req, fmt.Errorf("invalid request - failed validation"))
			m := models.NewMessageResponse("invalid request body")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// TODO: Check company exists with company profile API

		// Check case type of incoming request is CVL
		if !(request.CaseType == constants.CVL.String()) {
			log.ErrorR(req, fmt.Errorf("only creditors-voluntary-liquidation can be filed"))
			m := models.NewMessageResponse(fmt.Sprintf("case type is not creditors-voluntary-liquidation for transaction %s", transactionID))
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// Add new insolvency resource to mongo
		model := transformers.InsolvencyResourceRequestToDB(&request, transactionID)

		err = svc.CreateInsolvencyResource(model)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("failed to create insolvency resource in database"))
			m := models.NewMessageResponse(fmt.Sprintf("there was a problem handling your request for transaction %s", transactionID))
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}

		log.InfoR(req, fmt.Sprintf("successfully added insolvency resource with transaction ID: %s, to mongo", transactionID))

		utils.WriteJSONWithStatus(w, req, transformers.InsolvencyResourceDaoToCreatedResponse(model), http.StatusCreated)

		// TODO: Update transaction API with new insolvency resource
	})
}

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
		var request []models.PractitionerRequest
		err := json.NewDecoder(req.Body).Decode(&request)

		// Request body failed to get decoded
		if err != nil {
			log.ErrorR(req, fmt.Errorf("invalid request"))
			m := models.NewMessageResponse(fmt.Sprintf("failed to read request body for transaction %s", transactionID))
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		var daoList []models.PractitionerResourceDao

		for _, practitioner := range request {
			// Check all required fields are populated
			v := validator.New()
			if v.Struct(practitioner) != nil {
				log.ErrorR(req, fmt.Errorf("invalid request - failed validation"))
				m := models.NewMessageResponse("invalid request body")
				utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
				return
			}

			// Check if practitioner role supplied is valid
			if ok := constants.IsInRoleList(practitioner.Role); !ok {
				log.ErrorR(req, fmt.Errorf("invalid practitioner role"))
				m := models.NewMessageResponse(fmt.Sprintf("the practitioner role supplied is not valid %s", practitioner.Role))
				utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
				return
			}

			practitionerDao := transformers.PractitionerResourceRequestToDB(&practitioner, transactionID)
			daoList = append(daoList, *practitionerDao)
		}

		// Store practitioner resources in Mongo
		err = svc.CreatePractitionersResource(daoList, transactionID)
		if err != nil {
			switch err {
			case dao.ErrorNotFound:
				log.ErrorR(req, fmt.Errorf("failed to create practitioner resource in database - transaction %s not found", transactionID))
				m := models.NewMessageResponse(fmt.Sprintf("there was a problem handling your request for transaction %s not found", transactionID))
				utils.WriteJSONWithStatus(w, req, m, http.StatusNotFound)
				return
			case dao.ErrorPractitionerLimitReached:
				log.ErrorR(req, fmt.Errorf("failed to create practitioner resource in database - transaction %s already has 5 practitioners", transactionID))
				m := models.NewMessageResponse(fmt.Sprintf("there was a problem handling your request for transaction %s already has 5 practitioners", transactionID))
				utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
				return
			case dao.ErrorPractitionerLimitWillExceed:
				log.ErrorR(req, fmt.Errorf("failed to create practitioner resource in database - transaction %s will have more than 5 practitioners", transactionID))
				m := models.NewMessageResponse(fmt.Sprintf("there was a problem handling your request for transaction %s will have more than 5 practitioners", transactionID))
				utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
				return
			default:
				log.ErrorR(req, fmt.Errorf("failed to create practitioner resource in database"))
				m := models.NewMessageResponse(fmt.Sprintf("there was a problem handling your request for transaction %s", transactionID))
				utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
				return
			}
		}

		log.InfoR(req, fmt.Sprintf("successfully added practitioners resource with transaction ID: %s, to mongo", transactionID))

		utils.WriteJSONWithStatus(w, req, transformers.PractitionerResourceDaoListToCreatedResponseList(daoList), http.StatusCreated)
	})
}
