package handlers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"

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

		attachmentResponse, err := transformers.AttachmentResourceDaoToResponse(attachmentDao, header)
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
			log.ErrorR(req, fmt.Errorf("there is no file ID in the URL path"))
			m := models.NewMessageResponse("file ID is not in the URL path")
			utils.WriteJSONWithStatus(w, req, m, http.StatusBadRequest)
			return
		}

		log.InfoR(req, fmt.Sprintf("start GET request for attachment with transaction id: %s, file id: %s", transactionID, attachmentID))

		// Call File Transfer API to get attachment details
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

		partHeaders := textproto.MIMEHeader{}
		partHeaders.Set("Content-Type", GetAttachmentDetailsResponse.ContentType)

		header := multipart.FileHeader{
			Filename: GetAttachmentDetailsResponse.Name,
			Size:     GetAttachmentDetailsResponse.Size,
			Header:   partHeaders,
		}

		//-------------------
		//MONGO DB call

		attachmentDao, err := svc.GetAttachmentFromInsolvencyResource(transactionID, attachmentID)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("failed to get attachment from insolvency resource in db for transaction [%s] with attachment id of [%s]: %v", transactionID, attachmentID, err))
			m := models.NewMessageResponse("there was a problem handling your request")
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}

		attachmentResponse, err := transformers.AttachmentResourceDaoToResponse(&attachmentDao[0], &header)
		if err != nil {
			log.ErrorR(req, fmt.Errorf("error transforming dao to response: [%s]", err))
			m := models.NewMessageResponse("there was a problem handling your request")
			utils.WriteJSONWithStatus(w, req, m, http.StatusInternalServerError)
			return
		}

		utils.WriteJSONWithStatus(w, req, attachmentResponse, http.StatusOK)
	})
}
