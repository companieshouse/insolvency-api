package transformers

import (
	"fmt"

	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
)

// PractitionerResourceRequestToDB transforms practitioner request model to the dao model
func PractitionerResourceRequestToDB(req *models.PractitionerRequest, transactionID string) *models.PractitionerResourceDao {

	id := utils.GenerateID()
	selfLink := fmt.Sprintf("%s", constants.TransactionsPath+transactionID+constants.PractitionersPath+id)

	// Pad IP Code with leading zeros
	req.IPCode = fmt.Sprintf("%08s", req.IPCode)

	dao := &models.PractitionerResourceDao{
		ID:              id,
		IPCode:          req.IPCode,
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		TelephoneNumber: req.TelephoneNumber,
		Email:           req.Email,
		Address: models.AddressResourceDao{
			Premises:     req.Address.Premises,
			AddressLine1: req.Address.AddressLine1,
			AddressLine2: req.Address.AddressLine2,
			Country:      req.Address.Country,
			Locality:     req.Address.Locality,
			Region:       req.Address.Region,
			PostalCode:   req.Address.PostalCode,
			POBox:        req.Address.POBox,
		},
		Role: req.Role,
		Links: models.PractitionerResourceLinksDao{
			Self: selfLink,
		},
	}

	return dao
}

// PractitionerResourceDaoToCreatedResponse transforms the dao model to the created response model
func PractitionerResourceDaoToCreatedResponse(model *models.PractitionerResourceDao) *models.CreatedPractitionerResource {
	return &models.CreatedPractitionerResource{
		IPCode:          model.IPCode,
		FirstName:       model.FirstName,
		LastName:        model.LastName,
		TelephoneNumber: model.TelephoneNumber,
		Email:           model.Email,
		Address: models.CreatedAddressResource{
			Premises:     model.Address.Premises,
			AddressLine1: model.Address.AddressLine1,
			AddressLine2: model.Address.AddressLine2,
			Country:      model.Address.Country,
			Locality:     model.Address.Locality,
			Region:       model.Address.Region,
			PostalCode:   model.Address.PostalCode,
			POBox:        model.Address.POBox,
		},
		Role: model.Role,
		Links: models.CreatedPractitionerLinksResource{
			Self: model.Links.Self,
		},
	}
}

// PractitionerResourceDaoListToCreatedResponseList transforms a list of practitioner dao models to
// a list of the created response model
func PractitionerResourceDaoListToCreatedResponseList(practitionerList []models.PractitionerResourceDao) []models.CreatedPractitionerResource {
	var createdPractitioners []models.CreatedPractitionerResource

	for _, practitioner := range practitionerList {
		createdPractitioners = append(createdPractitioners, *PractitionerResourceDaoToCreatedResponse(&practitioner))
	}

	return createdPractitioners
}

// PractitionerAppointmentRequestToDB transforms an appointment request to a dao model
func PractitionerAppointmentRequestToDB(req *models.PractitionerAppointment, transactionID string, practitionerID string) *models.AppointmentResourceDao {

	selfLink := fmt.Sprintf("%s", constants.TransactionsPath+transactionID+constants.PractitionersPath+practitionerID+"/appointment")
	dao := &models.AppointmentResourceDao{
		AppointedOn: req.AppointedOn,
		MadeBy:      req.MadeBy,
		Links: models.AppointmentResourceLinksDao{
			Self: selfLink,
		},
	}

	return dao
}

// PractitionerAppointmentDaoToResponse transforms an appointment dao model to a response
func PractitionerAppointmentDaoToResponse(appointment models.AppointmentResourceDao) models.AppointedPractitionerResource {
	return models.AppointedPractitionerResource{
		AppointedOn: appointment.AppointedOn,
		MadeBy:      appointment.MadeBy,
		Links: models.AppointedPractitionerLinksResource{
			Self: appointment.Links.Self,
		},
	}
}
