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
	"github.com/go-playground/validator/v10"
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
		ipCodes := make(map[string]bool)

		v := validator.New()
		for _, practitioner := range request {
			// Check all required fields are populated
			if v.Struct(practitioner) != nil {
				log.ErrorR(req, fmt.Errorf("invalid request - failed validation"))
				m := models.NewMessageResponse("invalid request body")
				utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
				return
			}

			// Check if 2 of the same practitioners were supplied in the same request
			if _, exists := ipCodes[practitioner.IPCode]; exists {
				log.ErrorR(req, fmt.Errorf("invalid request - duplicate IP Code %s", practitioner.IPCode))
				m := models.NewMessageResponse("invalid request - duplicate IP Code: " + practitioner.IPCode)
				utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
				return
			}
			ipCodes[practitioner.IPCode] = true

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

		// Store practitioners resource in Mongo
		err, statusCode := svc.CreatePractitionersResource(daoList, transactionID)
		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse(err.Error())
			utils.WriteJSONWithStatus(w, req, m, statusCode)
			return
		}

		log.InfoR(req, fmt.Sprintf("successfully added practitioners resource with transaction ID: %s, to mongo", transactionID))

		utils.WriteJSONWithStatus(w, req, transformers.PractitionerResourceDaoListToCreatedResponseList(daoList), http.StatusCreated)
	})
}
