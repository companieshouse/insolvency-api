package dao

import (
	"github.com/companieshouse/insolvency-api/config"
	"github.com/companieshouse/insolvency-api/models"
)

// Service interface declares how to interact with the persistence layer regardless of underlying technology
type Service interface {

	// CreateInsolvencyResource will persist a newly created resource
	CreateInsolvencyResource(dao *models.InsolvencyResourceDao) (int, error)

	// GetInsolvencyResource will retrieve an Insolvency Resource with the specified transactionID.
	// If no insolvency resource is found then it returns nil with no error.
	GetInsolvencyResource(transactionID string) (*models.InsolvencyResourceDao, error)

	// GetInsolvencyAndExpandedPractitionerResources retrieves both the insolvency and practitioner resources,
	// with the appointment details inline, for an insolvency case with the specified transactionID.
	// If no insolvency resource is found then it returns nil for both resources, with no error.
	GetInsolvencyAndExpandedPractitionerResources(transactionID string) (*models.InsolvencyResourceDao, []models.PractitionerResourceDao, error)

	// CreatePractitionerResource will persist a newly created practitioner resource
	CreatePractitionerResource(dao *models.PractitionerResourceDao, transactionID string) (int, error)

	// AddPractitionerToInsolvencyResource will update insolvency by adding a link to a practitioner resource
	AddPractitionerToInsolvencyResource(transactionID string, practitionerID string, practitionerLink string) (int, error)

	// GetPractitionerAppointment will retrieve a practitioner appointment
	GetPractitionerAppointment(transactionID string, practitionerID string) (*models.AppointmentResourceDao, error)

	// GetSinglePractitionerResource gets a specific practitioner by transactionID & practitionerID.
	// If no insolvency case is found for the transactionID, or the insolvency case does not link to
	// the practitionerID, then the function returns nil with no error to indicate a 'not found' result.
	GetSinglePractitionerResource(transactionID string, practitionerID string) (*models.PractitionerResourceDao, error)

	// GetAllPractitionerResourcesForTransactionID gets all practitioner resources belonging to the insolvency case with the given transactionID.
	// If no insolvency case is found for the transactionID, or the insolvency case contains no practitioner
	// references, then the function returns nil with no error to indicate a 'not found' result.
	GetAllPractitionerResourcesForTransactionID(transactionID string) ([]models.PractitionerResourceDao, error)

	// DeletePractitioner will delete a practitioner from the Insolvency resource with the specified transactionID and practitionerID.
	// Any appointment data for the practitioner is also deleted. If successful it returns http.StatusNoContent.
	// If no insolvency case is found for the transactionID, or the insolvency case does not link to
	// the practitionerID, then the function returns http.StatusNotFound with an error.
	DeletePractitioner(transactionID, practitionerID string) (int, error)

	// CreateAppointmentResource will create appointment resource
	CreateAppointmentResource(dao *models.AppointmentResourceDao) (int, error)

	// UpdatePractitionerAppointment adds an appointment reference into the practitioner resource with the specified transactionID and practitionerID
	UpdatePractitionerAppointment(appointmentResourceDao *models.AppointmentResourceDao, transactionID string, practitionerID string) (int, error)

	// DeletePractitionerAppointment will delete the appointment for a practitioner
	// and remove the appointment reference from the practitioner resource.
	// If successful it returns http.StatusNoContent.
	// If no insolvency case is found for the transactionID, the insolvency case does not link to
	// the practitionerID, or the practitionerID has no appointment, then the function returns http.StatusNotFound with an error.
	DeletePractitionerAppointment(transactionID, practitionerID, etag string) (int, error)

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

	//GetProgressReportResource retrieves the progress report resource from an Insolvency case
	GetProgressReportResource(transactionID string) (*models.ProgressReportResourceDao, error)

	//DeleteProgressReportResource deletes a progress report for an insolvency case
	DeleteProgressReportResource(transactionID string) (int, error)
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
