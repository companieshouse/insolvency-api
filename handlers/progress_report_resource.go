package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/dao"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/service"
	"github.com/companieshouse/insolvency-api/transformers"
	"github.com/companieshouse/insolvency-api/utils"
)

// HandleCreateProgressReport receives a progress report to be stored against the Insolvency case
func HandleCreateProgressReport(svc dao.Service, helperService utils.HelperService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		// Check transaction is valid
		transactionID, isValidTransaction := utils.ValidateTransaction(helperService, req, w, "progress report", service.CheckIfTransactionClosed)
		if !isValidTransaction {
			return
		}

		// Decode Request body
		var request models.ProgressReport
		err := json.NewDecoder(req.Body).Decode(&request)
		isValidDecoded := helperService.HandleBodyDecodedValidation(w, req, transactionID, err)
		if !isValidDecoded {
			return
		}

		progressReportDao := transformers.ProgressReportResourceRequestToDB(&request, helperService)

		// Validate all mandatory fields
		errs := utils.Validate(request)
		isValidMarshallToDB := helperService.HandleMandatoryFieldValidation(w, req, errs)
		if !isValidMarshallToDB {
			return
		}

		// Validate the provided statement details are in the correct format
		validationErrs, err := service.ValidateProgressReportDetails(svc, progressReportDao, transactionID, req)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("failed to validate progress report: [%s]", err))
			m := models.NewMessageResponse(fmt.Sprintf("there was a problem handling your request for transaction ID [%s]", transactionID))
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}
		if validationErrs != "" {
			log.ErrorR(req, fmt.Errorf("invalid request - failed validation on the following: %s", validationErrs))
			m := models.NewMessageResponse("invalid request body: " + validationErrs)
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// Validate if supplied attachment matches attachments associated with supplied transactionID in mongo db
		attachment, err := svc.GetAttachmentFromInsolvencyResource(transactionID, progressReportDao.Attachments[0])
		isValidAttachment := helperService.HandleAttachmentValidation(w, req, transactionID, attachment, err)
		if !isValidAttachment {
			return
		}

		// Validate the supplied attachment is a valid type
		if attachment.Type != "progress-report" {
			err := fmt.Errorf("attachment id [%s] is an invalid type for this request: %v", progressReportDao.Attachments[0], attachment.Type)
			responseMessage := "attachment is not a progress-report"

			helperService.HandleAttachmentTypeValidation(w, req, responseMessage, err)
			return
		}

		// Creates the progress report resource in mongo if all previous checks pass
		statusCode, err := svc.CreateProgressReportResource(progressReportDao, transactionID)
		isValidCreateResource := helperService.HandleCreateResourceValidation(w, req, statusCode, err)

		if !isValidCreateResource {
			return
		}

		daoResponse := transformers.ProgressReportDaoToResponse(progressReportDao)

		log.InfoR(req, fmt.Sprintf("successfully added statement of progress report with transaction ID: %s, to mongo", transactionID))

		utils.WriteJSONWithStatus(w, req, daoResponse, http.StatusOK)
	})
}
