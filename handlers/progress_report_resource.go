package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/companieshouse/insolvency-api/service"
	"github.com/gorilla/mux"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/dao"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/transformers"
	"github.com/companieshouse/insolvency-api/utils"
)

// HandleCreateProgressReport receives a progress report to be stored against the Insolvency case
func HandleCreateProgressReport(svc dao.Service, helperService utils.HelperService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		incomingTransactionId := utils.GetTransactionIDFromVars(mux.Vars(req))
		transactionID, validTransactionId, httpStatusCode := helperService.HandleTransactionIdExistsValidation(w, req, incomingTransactionId)

		if !validTransactionId {
			http.Error(w, "Bad request", httpStatusCode)
			return
		}

		log.InfoR(req, fmt.Sprintf("start POST request for submit progress report with transaction id: %s", transactionID))

		isTransactionClosed, err, httpStatus := service.CheckIfTransactionClosed(transactionID, req)

		_, validTransactionNotClosed, httpStatusCodes := helperService.HandleTransactionNotClosedValidation(w, req, transactionID, isTransactionClosed, err, httpStatus)
		
		if !validTransactionNotClosed {
			http.Error(w, "Transaction closed", httpStatusCodes)
			return
		}

		//todo replace using generics when GO version 1.18+
		var request models.ProgressReport
		err = json.NewDecoder(req.Body).Decode(&request)

		isDecoded, httpStatusCode := helperService.HandleBodyDecodedValidation(w, req, transactionID, err)

		if !isDecoded {
			http.Error(w, "Server error", httpStatusCode)
			return
		}

		//todo replace using generics when GO version 1.18+
		progressReportDao := transformers.ProgressReportResourceRequestToDB(&request, helperService)

		//todo replace using generics when GO version 1.18+
		// Creates the progress report resource in mongo if all previous checks pass
		statusCode, err := svc.CreateProgressReportResource(progressReportDao, transactionID)

		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		isReportValidated, htthttpStatusCode := helperService.HandleCreateProgressReportResourceValidation(w, req, err, statusCode)

		if !isReportValidated {
			http.Error(w, "", htthttpStatusCode)
			return
		}

		utils.WriteJSONWithStatus(w, req, transformers.ProgressReportDaoToResponse(progressReportDao), http.StatusOK)
	})
}
