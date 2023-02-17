package dao

import (
	"github.com/companieshouse/insolvency-api/config"
	"github.com/companieshouse/insolvency-api/models"
)

// Service interface declares how to interact with the persistence layer regardless of underlying technology
type Service interface {

	// CreateInsolvencyResource will persist a newly created resource
	CreateInsolvencyResource(dao *models.InsolvencyResourceDto) (int, error)

	// GetInsolvencyResource will retrieve an Insolvency Resource
	GetInsolvencyResource(transactionID string) (models.InsolvencyResourceDao, error)

	// CreatePractitionerResource will persist a newly created practitioner resource
	CreatePractitionerResource(dao *models.PractitionerResourceDto, transactionID string) (int, error)

	// UpdateInsolvencyPractitioners will update insolvency with practitioners resource
	UpdateInsolvencyPractitioners(practitionersResource models.InsolvencyResourceDto, transactionID string) (int, error)

	// GetInsolvencyResourceData will retrieve insolvency dto object by transactionID
	GetInsolvencyResourceData(transactionID string) (*models.InsolvencyResourceDaoDataDto, error)

	// GetPractitionerAppointment will retrieve a practitioner appointment
	GetPractitionerAppointment(practitionerID string, transactionID string) (*models.AppointmentResourceDao, error)

	// GetPractitionersByIdsFromPractitioner will retrieve practitioner(s) from the insolvency resource
	GetPractitionersByIdsFromPractitioner(practitionerIDs []string, transactionID string) ([]models.PractitionerResourceDto, error)

	// DeletePractitioner will delete a practitioner from the Insolvency resource
	DeletePractitioner(practitionerID, transactionID string) (int, error)

	// CreateAppointmentResource will create appointment resource
	CreateAppointmentResource(dao *models.AppointmentResourceDto) (int, error)

	// UpdatePractitionerAppointment will update practitioner with appointment
	UpdatePractitionerAppointment(appointmentResourceDao *models.AppointmentResourceDto, transactionID string, practitionerID string) (int, error)

	// DeletePractitionerAppointment will delete the appointment for a practitioner
	DeletePractitionerAppointment(transactionID string, practitionerID string) (int, error)

	// AddAttachmentToInsolvencyResource will add an attachment to an insolvency resource
	AddAttachmentToInsolvencyResource(transactionID string, fileID string, attachmentType string) (*models.AttachmentResourceDao, error)

	// GetAttachmentFromInsolvencyResource will retrieve an attachment from an insolvency resource
	GetAttachmentFromInsolvencyResource(transactionID string, attachmentID string) (models.AttachmentResourceDao, error)

	// GetAttachmentResources retrieves all attachments filed for an Insolvency Case
	GetAttachmentResources(transactionID string) ([]models.AttachmentResourceDao, error)

	// DeleteAttachmentResource deletes an attachment in an Insolvency Case
	DeleteAttachmentResource(transactionID, attachmentID string) (int, error)

	// UpdateAttachmentStatus updates the status of an attachment for an Insolvency Case
	UpdateAttachmentStatus(transactionID, attachmentID, avStatus string) (int, error)

	// CreateStatementOfAffairsResource creates the statement of affairs resource for an Insolvency Case
	CreateStatementOfAffairsResource(dao *models.StatementOfAffairsResourceDao, transactionID string) (int, error)

	// CreateProgressReportResource creates the progress report resource for an Insolvency Case
	CreateProgressReportResource(dao *models.ProgressReportResourceDao, transactionID string) (int, error)

	// DeleteStatementOfAffairsResource deletes the statement of affairs filed for an insolvency case
	DeleteStatementOfAffairsResource(transactionID string) (int, error)

	// CreateResolutionResource creates the resolution resource for an Insolvency Case
	CreateResolutionResource(dao *models.ResolutionResourceDao, transactionID string) (int, error)

	// GetStatementOfAffairsResource retrieves the statement of affairs resource from an Insolvency Case
	GetStatementOfAffairsResource(transactionID string) (models.StatementOfAffairsResourceDao, error)

	// GetResolutionResource retrieves the resolution resource from an Insolvency Case
	GetResolutionResource(transactionID string) (models.ResolutionResourceDao, error)

	// DeleteResolutionResource deletes a resolution for an Insolvency Case
	DeleteResolutionResource(transactionID string) (int, error)
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
