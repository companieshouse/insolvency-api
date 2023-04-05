package transformers

import (
	"fmt"

	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/models"
)

// PractitionerResourceRequestToDB transforms practitioner request model to the dao model
func PractitionerResourceRequestToDB(req *models.PractitionerRequest, practitionerID string, transactionID string) *models.PractitionerResourceDao {

	selfLink := fmt.Sprintf(constants.TransactionsPath + transactionID + constants.PractitionersPath + practitionerID)

	// Pad IP Code with leading zeros
	req.IPCode = fmt.Sprintf("%08s", req.IPCode)

	practitionerResourceDao := models.PractitionerResourceDao{}
	practitionerResourceDao.TransactionID = transactionID
	practitionerResourceDao.Data.IPCode = req.IPCode
	practitionerResourceDao.Data.FirstName = req.FirstName
	practitionerResourceDao.Data.LastName = req.LastName
	practitionerResourceDao.Data.TelephoneNumber = req.TelephoneNumber
	practitionerResourceDao.Data.Email = req.Email
	practitionerResourceDao.Data.Address = models.AddressResourceDao{
		Premises:     req.Address.Premises,
		AddressLine1: req.Address.AddressLine1,
		AddressLine2: req.Address.AddressLine2,
		Country:      req.Address.Country,
		Locality:     req.Address.Locality,
		Region:       req.Address.Region,
		PostalCode:   req.Address.PostalCode,
		POBox:        req.Address.POBox,
	}
	practitionerResourceDao.Data.Role = req.Role
	practitionerResourceDao.Data.Links = models.PractitionerResourceLinksDao{
		Self: selfLink,
	}

	return &practitionerResourceDao
}

// PractitionerResourceDaoToCreatedResponse transforms the dao model to the created response model
func PractitionerResourceDaoToCreatedResponse(model *models.PractitionerResourceDao) *models.CreatedPractitionerResource {
	practitionerResourceDao := model.Data

	practitionerResource := &models.CreatedPractitionerResource{
		PractitionerId:  practitionerResourceDao.PractitionerId,
		IPCode:          practitionerResourceDao.IPCode,
		FirstName:       practitionerResourceDao.FirstName,
		LastName:        practitionerResourceDao.LastName,
		Email:           practitionerResourceDao.Email,
		TelephoneNumber: practitionerResourceDao.TelephoneNumber,
		Etag:            practitionerResourceDao.Etag,
		Kind:            practitionerResourceDao.Kind,
		Address: models.CreatedAddressResource{
			Premises:     practitionerResourceDao.Address.Premises,
			AddressLine1: practitionerResourceDao.Address.AddressLine1,
			AddressLine2: practitionerResourceDao.Address.AddressLine2,
			Country:      practitionerResourceDao.Address.Country,
			Locality:     practitionerResourceDao.Address.Locality,
			Region:       practitionerResourceDao.Address.Region,
			PostalCode:   practitionerResourceDao.Address.PostalCode,
			POBox:        practitionerResourceDao.Address.POBox,
		},
		Role: practitionerResourceDao.Role,
	}
	if len(practitionerResourceDao.Links.Appointment) > 0 {
		practitionerResource.Links.Appointment = &practitionerResourceDao.Links.Appointment
	}

	if len(practitionerResourceDao.Links.Self) > 0 {
		practitionerResource.Links.Self = &practitionerResourceDao.Links.Self
	}

	return practitionerResource

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
func PractitionerAppointmentRequestToDB(req *models.PractitionerAppointment, transactionID string, practitionerID string) models.AppointmentResourceDao {
	selfLink := fmt.Sprintf(constants.TransactionsPath + transactionID + constants.PractitionersPath + practitionerID + "/appointment")

	appointmentResourceDao := models.AppointmentResourceDao{}
	appointmentResourceDao.TransactionID = transactionID
	appointmentResourceDao.Data.AppointedOn = req.AppointedOn
	appointmentResourceDao.Data.MadeBy = req.MadeBy
	appointmentResourceDao.Data.Links = models.AppointmentResourceLinksDao{
		Self: selfLink,
	}

	appointmentResourceDao.PractitionerId = practitionerID

	return appointmentResourceDao
}

// PractitionerAppointmentDaoToResponse transforms an appointment dao model to a response
func PractitionerAppointmentDaoToResponse(appointment *models.AppointmentResourceDao) models.AppointedPractitionerResource {
	return models.AppointedPractitionerResource{
		AppointedOn: appointment.Data.AppointedOn,
		MadeBy:      appointment.Data.MadeBy,
		Links: models.AppointedPractitionerLinksResource{
			Self: appointment.Data.Links.Self,
		},
		Etag: appointment.Data.Etag,
		Kind: appointment.Data.Kind,
	}
}
