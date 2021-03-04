package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/dao"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/transformers"
	"github.com/companieshouse/insolvency-api/utils"
	"github.com/gorilla/mux"
	"net/http"
)

// HandleCreateInsolvencyResource creates an insolvency resource
func HandleCreateInsolvencyResource(svc dao.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.InfoR(req, "start POST request for insolvency resource")

		// Check for a transaction id in request
		vars := mux.Vars(req)
		transactionID, err := utils.GetTransactionIDFromVars(vars)
		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse("company number is not in request context")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// Decode the incoming request to create an insolvency resource
		var request models.InsolvencyRequest
		err = json.NewDecoder(req.Body).Decode(&request)

		// Request body failed to get decoded
		if err != nil {
			log.ErrorR(req, fmt.Errorf("invalid request"))
			m := models.NewMessageResponse("failed to read request body")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// TODO: Check company exists with company profile API

		// Check case type of incoming request is CVL
		if !(request.CaseType == constants.CVL.String()) {
			log.ErrorR(req, fmt.Errorf("only creditors-voluntary-liquidation can be filed"))
			m := models.NewMessageResponse("case type is not creditors-voluntary-liquidation")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// Add new insolvency resource to mongo
		model := transformers.InsolvencyResourceRequestToDB(&request, transactionID)

		err = svc.CreateInsolvencyResource(model)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("failed to create payable request in database"))
			m := models.NewMessageResponse("there was a problem handling your request")
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}

		utils.WriteJSONWithStatus(w, req, transformers.InsolvencyResourceDaoToCreatedResponse(model), http.StatusCreated)

		//TODO: Update transaction API with new insolvency resource
	})
}
