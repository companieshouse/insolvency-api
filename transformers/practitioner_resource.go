package transformers

import (
	"fmt"

	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
)

// PractitionerResourceRequestToDB transforms practitioner request model to the dao model
func PractitionerResourceRequestToDB(req *models.PractitionerRequest, transactionID string) *models.PractitionerResourceDao {

	id := utils.GenerateID()
	selfLink := fmt.Sprintf("/transactions/" + transactionID + "/insolvency/practitioners/" + id)

	dao := &models.PractitionerResourceDao{
		ID:        id,
		IPCode:    req.IPCode,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Address: models.AddressResourceDao{
			AddressLine1: req.Address.AddressLine1,
			AddressLine2: req.Address.AddressLine2,
			Country:      req.Address.Country,
			Locality:     req.Address.Locality,
			Region:       req.Address.Region,
			PostalCode:   req.Address.PostalCode,
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
		IPCode:    model.IPCode,
		FirstName: model.FirstName,
		LastName:  model.LastName,
		Address: models.CreatedAddressResource{
			AddressLine1: model.Address.AddressLine1,
			AddressLine2: model.Address.AddressLine2,
			Country:      model.Address.Country,
			Locality:     model.Address.Locality,
			Region:       model.Address.Region,
			PostalCode:   model.Address.PostalCode,
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
