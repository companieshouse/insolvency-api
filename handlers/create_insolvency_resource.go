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

		// Check with company profile API if company is valid
		err, httpStatus := service.CheckCompanyInsolvencyValid(&request, req)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("company was not found valid when checking company profile API [%v]", err))
			m := models.NewMessageResponse(fmt.Sprintf("company [%s] was not found valid for insolvency: %v", request.CompanyNumber, err))
			utils.WriteJSONWithStatus(w, req, m, httpStatus)
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
