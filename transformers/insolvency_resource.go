package transformers

import (
	"fmt"

	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
)

// InsolvencyResourceRequestToDB will take the input request from the REST call and transform it to a dao ready for
// insertion into the database
func InsolvencyResourceRequestToDB(req *models.InsolvencyRequest, transactionID string) *models.InsolvencyResourceDto {

	selfLink := fmt.Sprintf(constants.TransactionsPath + transactionID + constants.InsolvencyPath)
	transactionLink := fmt.Sprintf(constants.TransactionsPath + transactionID)
	validationLink := fmt.Sprintf(constants.TransactionsPath + transactionID + constants.ValidationStatusPath)

	insolvencyResourceDto := &models.InsolvencyResourceDto{
		TransactionID: transactionID,
		Data: models.InsolvencyResourceDaoDataDto{
			Kind:          "insolvency-resource#insolvency-resource",
			CompanyNumber: req.CompanyNumber,
			CaseType:      req.CaseType,
			CompanyName:   req.CompanyName,
		},

		Links: models.InsolvencyResourceLinksDao{
			Self:             selfLink,
			Transaction:      transactionLink,
			ValidationStatus: validationLink,
		},
	}

	return insolvencyResourceDto
}

// InsolvencyResourceDaoToCreatedResponse will transform an insolvency resource dao that has successfully been created into
// a http response entity
func InsolvencyResourceDaoToCreatedResponse(insolvencyResourceDto *models.InsolvencyResourceDto) *models.CreatedInsolvencyResource {
	return &models.CreatedInsolvencyResource{
		CompanyNumber: insolvencyResourceDto.Data.CompanyNumber,
		CaseType:      insolvencyResourceDto.Data.CaseType,
		Etag:          insolvencyResourceDto.Data.Etag,
		Kind:          insolvencyResourceDto.Data.Kind,
		CompanyName:   insolvencyResourceDto.Data.CompanyName,
		Links: models.CreatedInsolvencyResourceLinks{
			Self:             insolvencyResourceDto.Links.Self,
			Transaction:      insolvencyResourceDto.Links.Transaction,
			ValidationStatus: insolvencyResourceDto.Links.ValidationStatus,
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
func AttachmentResourceDaoToResponse(dao *models.AttachmentResourceDao, name string, size int64, contentType string, helperService utils.HelperService) (*models.AttachmentResource, error) {

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
