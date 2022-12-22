package transformers

import (
	"fmt"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
)

// InsolvencyResourceRequestToDB will take the input request from the REST call and transform it to a dao ready for
// insertion into the database
func InsolvencyResourceRequestToDB(req *models.InsolvencyRequest, transactionID string, helperService utils.HelperService) *models.InsolvencyResourceDao {

	etag, err := helperService.GenerateEtag()
	
	if err != nil {
		log.Error(fmt.Errorf("error generating etag: [%s]", err))
		return nil
	}

	selfLink := fmt.Sprintf(constants.TransactionsPath + transactionID + constants.InsolvencyPath)
	transactionLink := fmt.Sprintf(constants.TransactionsPath + transactionID)
	validationLink := fmt.Sprintf(constants.TransactionsPath + transactionID + constants.ValidationStatusPath)

	dao := &models.InsolvencyResourceDao{
		TransactionID: transactionID,
		Data: models.InsolvencyResourceDaoData{
			CompanyNumber: req.CompanyNumber,
			CaseType:      req.CaseType,
			CompanyName:   req.CompanyName,
		},
		Etag: etag,
		Kind: "insolvency-resource#insolvency-resource",
		Links: models.InsolvencyResourceLinksDao{
			Self:             selfLink,
			Transaction:      transactionLink,
			ValidationStatus: validationLink,
		},
	}

	return dao
}

// InsolvencyResourceDaoToCreatedResponse will transform an insolvency resource dao that has successfully been created into
// a http response entity
func InsolvencyResourceDaoToCreatedResponse(model *models.InsolvencyResourceDao) *models.CreatedInsolvencyResource {
	return &models.CreatedInsolvencyResource{
		CompanyNumber: model.Data.CompanyNumber,
		CaseType:      model.Data.CaseType,
		Etag:          model.Etag,
		Kind:          model.Kind,
		CompanyName:   model.Data.CompanyName,
		Links: models.CreatedInsolvencyResourceLinks{
			Self:             model.Links.Self,
			Transaction:      model.Links.Transaction,
			ValidationStatus: model.Links.ValidationStatus,
		},
	}
}

// AppointmentResourceDaoToAppointedResponse transforms an appointment resource dao into a response entity
func AppointmentResourceDaoToAppointedResponse(model *models.AppointmentResourceDao) *models.AppointedPractitionerResource {
	return &models.AppointedPractitionerResource{
		AppointedOn: model.AppointedOn,
		MadeBy:      model.MadeBy,
		Links: models.AppointedPractitionerLinksResource{
			Self: model.Links.Self,
		},
	}
}

// AttachmentResourceDaoToResponse transforms an attachment resource dao and file attachment details into a response entity
func AttachmentResourceDaoToResponse(dao *models.AttachmentResourceDao, name string, size int64, contentType string,helperService utils.HelperService) (*models.AttachmentResource, error) {
	
	etag, err := helperService.GenerateEtag()
	
	if err != nil {
		return nil, fmt.Errorf("error generating etag")
	}

	attachmentResource := &models.AttachmentResource{
		AttachmentType: dao.Type,
		File: models.AttachmentFile{
			Name:        name,
			Size:        size,
			ContentType: contentType,
		},
		Etag:   etag,
		Kind:   "insolvency-resources#attachment",
		Status: dao.Status,
		Links: models.AttachmentLinksResource{
			Self:     dao.Links.Self,
			Download: dao.Links.Download,
		},
	}

	return attachmentResource, nil
}
