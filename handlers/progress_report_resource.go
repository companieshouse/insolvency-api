package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/companieshouse/insolvency-api/service"
	"github.com/gorilla/mux"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/dao"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/transformers"
	"github.com/companieshouse/insolvency-api/utils"
)

// HandleCreateProgressReport receives a progress report to be stored against the Insolvency case
func HandleCreateProgressReport(svc dao.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		transactionID, validTransactionId := utils.HandleTransactionIdExistsValidation(w, req, utils.GetTransactionIDFromVars(mux.Vars(req)))
		if !validTransactionId {
			return
		}

		log.InfoR(req, fmt.Sprintf("start POST request for submit progress report with transaction id: %s", transactionID))

		isTransactionClosed, err, httpStatus := service.CheckIfTransactionClosed(transactionID, req)
		err, validTransactionNotClosed := utils.HandleTransactionNotClosedValidation(w, req, transactionID, isTransactionClosed, err, httpStatus)
		if !validTransactionNotClosed {
			return
		}

		//todo replace using generics when GO version 1.18+
		var request models.ProgressReport
		err = json.NewDecoder(req.Body).Decode(&request)

		// Request body failed to get decoded
		if !utils.HandleBodyDecodedValidation(w, req, transactionID, err) {
			return
		}

		//todo replace using generics when GO version 1.18+
		progressReportDao := transformers.ProgressReportResourceRequestToDB(&request)

		//todo replace using generics when GO version 1.18+
		// Creates the progress report resource in mongo if all previous checks pass
		statusCode, err := svc.CreateProgressReportResource(progressReportDao, transactionID)
		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse(err.Error())
			utils.WriteJSONWithStatus(w, req, m, statusCode)
			return
		}

		utils.WriteJSONWithStatus(w, req, transformers.ProgressReportDaoToResponse(progressReportDao), http.StatusOK)
	})
}
