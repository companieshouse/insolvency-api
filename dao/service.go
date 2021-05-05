package dao

import (
	"github.com/companieshouse/insolvency-api/config"
	"github.com/companieshouse/insolvency-api/models"
)

// Service interface declares how to interact with the persistence layer regardless of underlying technology
type Service interface {
	// CreateInsolvencyResource will persist a newly created resource
	CreateInsolvencyResource(dao *models.InsolvencyResourceDao) error

	// CreatePractitionersResource will persist a newly created practitioner resource
	CreatePractitionersResource(dao *models.PractitionerResourceDao, transactionID string) (error, int)

	// GetPractitionerResources will retrieve a list of persisted practitioners
	GetPractitionerResources(transactionID string) ([]models.PractitionerResourceDao, error)

	// GetPractitionerResource will retrieve a practitioner from the Insolvency resource
	GetPractitionerResource(practitionerID string, transactionID string) (models.PractitionerResourceDao, error)

	// DeletePractitioner will delete a practitioner from the Insolvency resource
	DeletePractitioner(practitionerID, transactionID string) (error, int)

	// AppointPractitioner will appoint add appointment details to a practitioner resource
	AppointPractitioner(dao *models.AppointmentResourceDao, transactionID string, practitionerID string) (error, int)

	// DeletePractitionerAppointment will delete the appointment for a practitioner
	DeletePractitionerAppointment(transactionID string, practitionerID string) (error, int)
}

// NewDAOService will create a new instance of the Service interface. All details about its implementation and the
// database driver will be hidden from outside of this package
func NewDAOService(cfg *config.Config) Service {
	database := getMongoDatabase(cfg.MongoDBURL, cfg.Database)
	return &MongoService{
		db:             database,
		CollectionName: cfg.MongoCollection,
	}
}
