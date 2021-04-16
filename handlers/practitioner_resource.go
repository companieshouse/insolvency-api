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

		// Check all required fields are populated
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
