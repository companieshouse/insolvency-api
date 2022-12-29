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
		transactionID, isValidTransactionId, httpStatusCode := helperService.HandleTransactionIdExistsValidation(w, req, incomingTransactionId)

		if !isValidTransactionId {
			http.Error(w, "Bad request", httpStatusCode)
			return
		}

		log.InfoR(req, fmt.Sprintf("start POST request for submit progress report with transaction id: %s", transactionID))

		isTransactionClosed, err, httpStatus := service.CheckIfTransactionClosed(transactionID, req)

		isValidTransactionNotClosed, httpStatusCodes, _ := helperService.HandleTransactionNotClosedValidation(w, req, transactionID, isTransactionClosed, httpStatus, err)

		if !isValidTransactionNotClosed {
			http.Error(w, "Transaction closed", httpStatusCodes)
			return
		}

		var request models.ProgressReport
		err = json.NewDecoder(req.Body).Decode(&request)

		isValidDecoded, httpStatusCode := helperService.HandleBodyDecodedValidation(w, req, transactionID, err)

		if !isValidDecoded {
			http.Error(w, fmt.Sprintf("failed to read request body for transaction %s", transactionID), httpStatusCode)
			return
		}

		progressReportDao := transformers.ProgressReportResourceRequestToDB(&request, helperService)

		// Creates the progress report resource in mongo if all previous checks pass
		statusCode, err := svc.CreateProgressReportResource(progressReportDao, transactionID)

		if err != nil {
			http.Error(w, "Server error", statusCode)
			return
		}

		isValidCreateResource, httpStatusCode := helperService.HandleCreateResourceValidation(w, req, statusCode, err)

		if !isValidCreateResource {
			http.Error(w, "", httpStatusCode)
			return
		}

		daoResponse := transformers.ProgressReportDaoToResponse(progressReportDao)

		utils.WriteJSONWithStatus(w, req, daoResponse, http.StatusOK)
	})
}
