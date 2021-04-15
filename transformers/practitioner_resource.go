package transformers

import (
	"github.com/companieshouse/insolvency-api/models"
)

// PractitionerResourceRequestToDB transforms practitioner request model to the dao model
func PractitionerResourceRequestToDB(req *models.PractitionerRequest, transactionID string) *models.PractitionerResourceDao {

	dao := &models.PractitionerResourceDao{
		IPCode:          req.IPCode,
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		TelephoneNumber: req.TelephoneNumber,
		Email:           req.Email,
		Address: models.AddressResourceDao{
			AddressLine1: req.Address.AddressLine1,
			AddressLine2: req.Address.AddressLine2,
			Country:      req.Address.Country,
			Locality:     req.Address.Locality,
			Region:       req.Address.Region,
			PostalCode:   req.Address.PostalCode,
		},
		Role: req.Role,
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
			AddressLine1: model.Address.AddressLine1,
			AddressLine2: model.Address.AddressLine2,
			Country:      model.Address.Country,
			Locality:     model.Address.Locality,
			Region:       model.Address.Region,
			PostalCode:   model.Address.PostalCode,
		},
		Role: model.Role,
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
