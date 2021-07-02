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
	"github.com/gorilla/mux"
)

// HandleCreateResolution receives a resolution to be stored against the Insolvency case
func HandleCreateResolution(svc dao.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		transactionID := utils.GetTransactionIDFromVars(vars)
		if transactionID == "" {
			log.ErrorR(req, fmt.Errorf("there is no transaction ID in the URL path"))
			m := models.NewMessageResponse("transaction ID is not in the URL path")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		log.InfoR(req, fmt.Sprintf("start POST request for submit attachment with transaction id: %s", transactionID))

		// Check if transaction is closed
		isTransactionClosed, err, httpStatus := service.CheckIfTransactionClosed(transactionID, req)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("error checking transaction status for [%v]: [%s]", transactionID, err))
			m := models.NewMessageResponse(fmt.Sprintf("error checking transaction status for [%v]: [%s]", transactionID, err))
			utils.WriteJSONWithStatus(w, req, m, httpStatus)
			return
		}
		if isTransactionClosed {
			log.ErrorR(req, fmt.Errorf("transaction [%v] is already closed and cannot be updated", transactionID))
			m := models.NewMessageResponse(fmt.Sprintf("transaction [%v] is already closed and cannot be updated", transactionID))
			utils.WriteJSONWithStatus(w, req, m, httpStatus)
			return
		}

		var request models.Resolution
		err = json.NewDecoder(req.Body).Decode(&request)

		// Request body failed to get decoded
		if err != nil {
			log.ErrorR(req, fmt.Errorf("invalid request"))
			m := models.NewMessageResponse(fmt.Sprintf("failed to read request body for transaction %s", transactionID))
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		resolutionDao := transformers.ResolutionResourceRequestToDB(&request)

		// Validate all mandatory fields
		if errs := utils.Validate(request); errs != "" {
			log.ErrorR(req, fmt.Errorf("invalid request - failed validation on the following: %s", errs))
			m := models.NewMessageResponse("invalid request body: " + errs)
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}
		// Validate if an attachment has been supplied
		if len(resolutionDao.Attachments) == 0 {
			log.ErrorR(req, fmt.Errorf("invalid attachment"))
			m := models.NewMessageResponse(fmt.Sprintf("no attachment has been supplied"))
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}
		// Validate if more than one attachment has been supplied
		if len(resolutionDao.Attachments) > 1 {
			log.ErrorR(req, fmt.Errorf("invalid attachments"))
			m := models.NewMessageResponse(fmt.Sprintf("only one attachment can be supplied: %s", resolutionDao.Attachments))
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		isAttachmentValid := true
		attachment, err := svc.GetAttachmentFromInsolvencyResource(transactionID, resolutionDao.Attachments[0])

		// Validate if supplied attachment matches attachments associated with supplied transactionID
		if attachment == (models.AttachmentResourceDao{}) {
			isAttachmentValid = false
			log.ErrorR(req, fmt.Errorf("failed to get attachment from insolvency resource in db for transaction [%s] with attachment id of [%s]: %v", transactionID, resolutionDao.Attachments[0], err))
			m := models.NewMessageResponse("attachment not found on transaction")
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}

		// Validate the supplied attachment is a valid type
		if attachment.Type != "resolution" {
			isAttachmentValid = false
			log.ErrorR(req, fmt.Errorf("attachment id [%s] is an invalid type for this request: %v", resolutionDao.Attachments[0], err))
			m := models.NewMessageResponse("attachment is not a resolution")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// Creates the resolution resource in mongo if all previous checks pass
		if isAttachmentValid {
			statusCode, err := svc.CreateResolutionResource(resolutionDao, transactionID)
			if err != nil {
				log.ErrorR(req, err)
				m := models.NewMessageResponse(err.Error())
				utils.WriteJSONWithStatus(w, req, m, statusCode)
				return
			}

			log.InfoR(req, fmt.Sprintf("successfully added resolution resource with transaction ID: %s, to mongo", transactionID))

			utils.WriteJSONWithStatus(w, req, transformers.ResolutionDaoToResponse(resolutionDao), http.StatusOK)
		}
	})
}
