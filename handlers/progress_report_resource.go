package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/companieshouse/insolvency-api/service"
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
		transactionID, validTransactionId := service.TransactionIdExists(w, req)
		if !validTransactionId {
			return
		}

		log.InfoR(req, fmt.Sprintf("start POST request for submit progress report with transaction id: %s", transactionID))

		err, validTransactionNotClosed := service.ValidateTransactionNotClosed(w, req, transactionID)
		if !validTransactionNotClosed {
			return
		}

		//todo replace using generics when GO version 1.18+
		var request models.ProgressReport
		err = json.NewDecoder(req.Body).Decode(&request)

		// Request body failed to get decoded
		if !service.ValidBodyDecode(w, req, err, transactionID) {
			return
		}

		//todo replace using generics when GO version 1.18+
		progressReportDao := transformers.ProgressReportResourceRequestToDB(&request)

		//todo replace using generics when GO version 1.18+
		// Creates the statement of affairs resource in mongo if all previous checks pass
		statusCode, err := svc.CreateProgressReportResource(progressReportDao, transactionID)
		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse(err.Error())
			utils.WriteJSONWithStatus(w, req, m, statusCode)
			return
		}

		utils.WriteJSONWithStatus(w, req, req.Body, http.StatusOK)
	})
}
