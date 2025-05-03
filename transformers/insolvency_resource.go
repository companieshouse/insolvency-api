package transformers

import (
	"fmt"

	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
)

// InsolvencyResourceRequestToDB will take the input request from the REST call and transform it to a dao ready for
// insertion into the database
func InsolvencyResourceRequestToDB(req *models.InsolvencyRequest, transactionID string) *models.InsolvencyResourceDao {

	selfLink := fmt.Sprintf(constants.TransactionsPath + transactionID + constants.InsolvencyPath)
	transactionLink := fmt.Sprintf(constants.TransactionsPath + transactionID)
	validationLink := fmt.Sprintf(constants.TransactionsPath + transactionID + constants.ValidationStatusPath)

	insolvencyResourceDao := &models.InsolvencyResourceDao{
		TransactionID: transactionID,
	}

	insolvencyResourceDao.Data.Kind = "insolvency-resource#insolvency-resource"
	insolvencyResourceDao.Data.CompanyNumber = req.CompanyNumber
	insolvencyResourceDao.Data.CaseType = req.CaseType
	insolvencyResourceDao.Data.CompanyName = req.CompanyName
	insolvencyResourceDao.Data.Links = models.InsolvencyResourceLinksDao{
		Self:             selfLink,
		Transaction:      transactionLink,
		ValidationStatus: validationLink,
	}

	return insolvencyResourceDao
}

// InsolvencyResourceDaoToCreatedResponse will transform an insolvency resource dao that has successfully been created into
// a http response entity
func InsolvencyResourceDaoToCreatedResponse(insolvencyResourceDao *models.InsolvencyResourceDao) *models.CreatedInsolvencyResource {
	return &models.CreatedInsolvencyResource{
		CompanyNumber: insolvencyResourceDao.Data.CompanyNumber,
		CaseType:      insolvencyResourceDao.Data.CaseType,
		Etag:          insolvencyResourceDao.Data.Etag,
		Kind:          insolvencyResourceDao.Data.Kind,
		CompanyName:   insolvencyResourceDao.Data.CompanyName,
		Links: models.CreatedInsolvencyResourceLinks{
			Self:             insolvencyResourceDao.Data.Links.Self,
			Transaction:      insolvencyResourceDao.Data.Links.Transaction,
			ValidationStatus: insolvencyResourceDao.Data.Links.ValidationStatus,
		},
	}
}

// PractitionerResourceDaosToPractitionerFilingsResponse transforms a slice of
// PractitionerResourceDaos (including appointment details) into a slice of
// CreatedPractitionerResources (also including appointment details)
func PractitionerResourceDaosToPractitionerFilingsResponse(practitionerResourceDaos []models.PractitionerResourceDao) []models.CreatedPractitionerResource {

	var practitionerResponses []models.CreatedPractitionerResource
	var appointedPractitionerResource models.AppointedPractitionerResource

	for _, practitioner := range practitionerResourceDaos {

		practitionerResponse := models.CreatedPractitionerResource{}
		practitionerResourceLinksDao := models.PractitionerResourceLinksDao{}

		practitionerResourceLinksDao.Appointment = practitioner.Data.Links.Appointment
		practitionerResourceLinksDao.Self = practitioner.Data.Links.Self

		practitionerResponse.PractitionerId = practitioner.Data.PractitionerId
		practitionerResponse.IPCode = practitioner.Data.IPCode
		practitionerResponse.FirstName = practitioner.Data.FirstName
		practitionerResponse.LastName = practitioner.Data.LastName
		practitionerResponse.Email = practitioner.Data.Email
		practitionerResponse.TelephoneNumber = practitioner.Data.TelephoneNumber
		practitionerResponse.Address = models.CreatedAddressResource{
			Premises:     practitioner.Data.Address.Premises,
			AddressLine1: practitioner.Data.Address.AddressLine1,
			AddressLine2: practitioner.Data.Address.AddressLine2,
			Country:      practitioner.Data.Address.Country,
			Locality:     practitioner.Data.Address.Locality,
			Region:       practitioner.Data.Address.Region,
			PostalCode:   practitioner.Data.Address.PostalCode,
			POBox:        practitioner.Data.Address.POBox,
		}
		practitionerResponse.Role = practitioner.Data.Role

		if practitioner.Data.Appointment != nil {
			appointedPractitionerResource = PractitionerAppointmentDaoToResponse(practitioner.Data.Appointment)
			appointedPractitionerResource.Etag = ""
			appointedPractitionerResource.Kind = ""

			practitionerResponse.Appointment = &appointedPractitionerResource
		}

		practitionerResponse.Links.Self = practitionerResourceLinksDao.Self

		practitionerResponses = append(practitionerResponses, practitionerResponse)
	}

	return practitionerResponses
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

// AttachmentResourceDaoToResponse transforms an attachment resource dao and file attachment details into a response entity
func AttachmentResourceDaoToFilingsResponse(dao *models.AttachmentResourceDao) *models.AttachmentFilingsResource {

	attachmentResource := &models.AttachmentFilingsResource{
		ID:     dao.ID,
		Type:   dao.Type,
		Status: dao.Status,
		Links: models.AttachmentLinksResource{
			Self:     dao.Links.Self,
			Download: dao.Links.Download,
		},
	}

	return attachmentResource
}
