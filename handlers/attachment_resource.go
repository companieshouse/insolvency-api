package handlers

import (
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

// HandleSubmitAttachment receives an attachment to be stored against the Insolvency case
func HandleSubmitAttachment(svc dao.Service) http.Handler {
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

		attachmentType := req.FormValue("attachment_type")

		file, header, err := req.FormFile("file")
		if err != nil {
			log.ErrorR(req, fmt.Errorf("error reading form from request: %s", err))
			m := models.NewMessageResponse("error reading form from request")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// Validate that the provided attachment details are correct
		validationErrs, err := service.ValidateAttachmentDetails(svc, transactionID, attachmentType, header)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("error validating attachment details: [%s]", err))
			m := models.NewMessageResponse(fmt.Sprintf("there was a problem handling your request for transaction ID [%s]", transactionID))
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}
		if validationErrs != "" {
			log.ErrorR(req, fmt.Errorf("invalid request - failed validation on the following: %s", validationErrs))
			m := models.NewMessageResponse("invalid request: " + validationErrs)
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		fileID, responseType, err := service.UploadAttachment(file, header, req)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("error uploading attachment: [%v]", err), log.Data{"service_response_type": responseType.String()})

			status, err := utils.ResponseTypeToStatus(responseType.String())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(status)
			return
		}
		if responseType != service.Success {
			log.ErrorR(req, fmt.Errorf("file upload was unsuccessful"))
			status, err := utils.ResponseTypeToStatus(responseType.String())
			if err != nil {
				log.ErrorR(req, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(status)
			return
		}

		attachmentDao, err := svc.AddAttachmentToInsolvencyResource(transactionID, fileID, attachmentType)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("failed to add attachment to insolvency resource in db for transaction [%s]: %v", transactionID, err))
			m := models.NewMessageResponse("there was a problem handling your request")
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}

		attachmentResponse, err := transformers.AttachmentResourceDaoToResponse(attachmentDao, header.Filename, header.Size, header.Header.Get("Content-Type"))
		if err != nil {
			log.ErrorR(req, fmt.Errorf("error transforming dao to response: [%s]", err))
			m := models.NewMessageResponse("there was a problem handling your request")
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}

		utils.WriteJSONWithStatus(w, req, attachmentResponse, http.StatusCreated)
	})
}

// HandleSubmitAttachment receives an attachment to be stored against the Insolvency case
func HandleGetAttachmentDetails(svc dao.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		transactionID := utils.GetTransactionIDFromVars(vars)
		attachmentID := utils.GetAttachmentIDFromVars(vars)
		if transactionID == "" {
			log.ErrorR(req, fmt.Errorf("there is no transaction ID in the URL path"))
			m := models.NewMessageResponse("transaction ID is not in the URL path")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		if attachmentID == "" {
			log.ErrorR(req, fmt.Errorf("there is no attachment ID in the URL path"))
			m := models.NewMessageResponse("attachment ID is not in the URL path")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		log.InfoR(req, fmt.Sprintf("start GET request for attachment with transaction id: %s, attachment id: %s", transactionID, attachmentID))

		// Calls the database and returns attachment stored against the Insolvency case
		attachmentDao, err := svc.GetAttachmentFromInsolvencyResource(transactionID, attachmentID)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("failed to get attachment from insolvency resource in db for transaction [%s] with attachment id of [%s]: %v", transactionID, attachmentID, err))
			m := models.NewMessageResponse("there was a problem handling your request")
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}

		if attachmentDao == (models.AttachmentResourceDao{}) {
			m := models.NewMessageResponse("attachment id is not valid")
			utils.WriteJSONWithStatus(w, req, m, http.StatusNotFound)
			return
		}

		// Calls File Transfer API to get attachment details
		GetAttachmentDetailsResponse, responseType, err := service.GetAttachmentDetails(attachmentID, req)

		if err != nil {
			log.ErrorR(req, fmt.Errorf("error getting attachment details: [%v]", err), log.Data{"service_response_type": responseType.String()})

			status, err := utils.ResponseTypeToStatus(responseType.String())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(status)
			return
		}

		attachmentResponse, err := transformers.AttachmentResourceDaoToResponse(&attachmentDao, GetAttachmentDetailsResponse.Name, GetAttachmentDetailsResponse.Size, GetAttachmentDetailsResponse.ContentType)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("error transforming dao to response: [%s]", err))
			m := models.NewMessageResponse("there was a problem handling your request")
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}

		utils.WriteJSONWithStatus(w, req, attachmentResponse, http.StatusOK)
	})
}

// HandleDownloadAttachment download an attachment which is stored against an Insolvency case
func HandleDownloadAttachment(svc dao.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		transactionID := utils.GetTransactionIDFromVars(vars)
		if transactionID == "" {
			log.ErrorR(req, fmt.Errorf("there is no transaction ID in the URL path"))
			m := models.NewMessageResponse("transaction ID is not in the URL path")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		attachmentID := utils.GetAttachmentIDFromVars(vars)
		if attachmentID == "" {
			log.ErrorR(req, fmt.Errorf("there is no attachment ID in the URL path"))
			m := models.NewMessageResponse("attachment ID is not in the URL path")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// Get attachment data from DB to check if attachment ID is valid
		attachmentResource, err := svc.GetAttachmentFromInsolvencyResource(transactionID, attachmentID)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("failed to get attachment from insolvency db resource for transaction [%s] with attachment id [%s]: %v", transactionID, attachmentID, err))
			m := models.NewMessageResponse("there was a problem handling your request")
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}
		if attachmentResource == (models.AttachmentResourceDao{}) {
			m := models.NewMessageResponse("attachment id is not valid")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		// get data from File Transfer API to check antivirus is complete
		attachmentDetails, responseType, err := service.GetAttachmentDetails(attachmentID, req)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("error getting attachment details: [%v]", err), log.Data{"service_response_type": responseType.String()})

			status, err := utils.ResponseTypeToStatus(responseType.String())
			if err != nil {
				log.ErrorR(req, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(status)
			return
		}
		if attachmentDetails.AVStatus != "clean" {
			log.ErrorR(req, fmt.Errorf("antivirus status not clean for attachment ID [%s]", attachmentID))
			m := models.NewMessageResponse("attachment unavailable for download")
			utils.WriteJSONWithStatus(w, req, m, http.StatusForbidden)
			return
		}

		responseType, err = service.DownloadAttachment(attachmentID, req, w)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("error downloading attachment: [%v]", err), log.Data{"service_response_type": responseType.String()})

			status, err := utils.ResponseTypeToStatus(responseType.String())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(status)
			return
		}
		if responseType != service.Success {
			log.ErrorR(req, fmt.Errorf("file download was unsuccessful"))
			status, err := utils.ResponseTypeToStatus(responseType.String())
			if err != nil {
				log.ErrorR(req, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(status)
			return
		}
	})
}

// HandleDeleteAttachment deletes an attachment resource from the DB and deletes the stored file
func HandleDeleteAttachment(svc dao.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		transactionID := utils.GetTransactionIDFromVars(vars)
		attachmentID := utils.GetAttachmentIDFromVars(vars)
		if transactionID == "" {
			log.ErrorR(req, fmt.Errorf("there is no transaction ID in the URL path"))
			m := models.NewMessageResponse("transaction ID is not in the URL path")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		if attachmentID == "" {
			log.ErrorR(req, fmt.Errorf("there is no attachment ID in the URL path"))
			m := models.NewMessageResponse("attachment ID is not in the URL path")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		log.InfoR(req, fmt.Sprintf("start DELETE request for attachment with transaction id: %s, attachment id: %s", transactionID, attachmentID))

		responseType, err := service.DeleteAttachment(attachmentID, req)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("error deleting attachment: [%v]", err), log.Data{"service_response_type": responseType.String()})

			status, err := utils.ResponseTypeToStatus(responseType.String())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(status)
			return
		}

		// Delete attachment from DB
		statusCode, err := svc.DeleteAttachmentResource(transactionID, attachmentID)
		if err != nil {
			log.ErrorR(req, err)
			m := models.NewMessageResponse(err.Error())
			utils.WriteJSONWithStatus(w, req, m, statusCode)
			return
		}

		utils.WriteJSONWithStatus(w, req, "", http.StatusNoContent)
	})
}
