package dao

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	PractitionerCollectionName = "practitioners"
	AppointmentCollectionName  = "appointments"
)

var client *mongo.Client

func getMongoClient(mongoDBURL string) *mongo.Client {
	if client != nil {
		return client
	}

	ctx := context.Background()

	clientOptions := options.Client().ApplyURI(mongoDBURL)
	client, err := mongo.Connect(ctx, clientOptions)

	// assume the caller of this func cannot handle the case where there is no database connection so the prog must
	// crash here as the service cannot continue.
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	// check we can connect to the mongodb instance. failure here should result in a crash.
	pingContext, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
	defer cancel()
	err = client.Ping(pingContext, nil)
	if err != nil {
		log.Error(errors.New("ping to mongodb timed out. please check the connection to mongodb and that it is running"))
		os.Exit(1)
	}

	log.Info("connected to mongodb successfully")

	return client
}

// MongoService is an implementation of the Service interface using MongoDB as the backend driver.
type MongoService struct {
	db             MongoDatabaseInterface
	CollectionName string
}

// MongoDatabaseInterface is an interface that describes the mongodb driver
type MongoDatabaseInterface interface {
	Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection
}

func getMongoDatabase(mongoDBURL, databaseName string) MongoDatabaseInterface {
	return getMongoClient(mongoDBURL).Database(databaseName)
}

// CreateInsolvencyResource will store the insolvency request into the database
func (m *MongoService) CreateInsolvencyResource(dao *models.InsolvencyResourceDao) (int, error) {

	dao.ID = primitive.NewObjectID()

	collection := m.db.Collection(m.CollectionName)

	filter := bson.M{"transaction_id": dao.TransactionID}

	// Try to retrieve existing insolvency case from Mongo
	existingInsolvency := collection.FindOne(context.Background(), filter)
	err := existingInsolvency.Err()
	if err != nil {
		// If no documents can be found then the insolvency case can be created
		if err == mongo.ErrNoDocuments {
			_, err = collection.InsertOne(context.Background(), dao)
			if err != nil {
				log.Error(err)
				return http.StatusInternalServerError, fmt.Errorf("there was a problem creating an insolvency case for this transaction id: %v", err)
			}

			return http.StatusCreated, nil
		}

		// If there is an error but it is not ErrNoDocuments then an error happened checking the existence of the insolvency case
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem creating an insolvency case for this transaction id: %v", err)
	}

	// If there is no error retrieving the insolvency case, then it already exists
	log.Info("an insolvency case already exists for this transaction id")
	return http.StatusConflict, fmt.Errorf("an insolvency case already exists for this transaction id")
}

// GetInsolvencyPractitionersResource retrieves all the data for an insolvency case with the specified transactionID
func (m *MongoService) GetInsolvencyPractitionersResource(transactionID string) (*models.InsolvencyResourceDao, []models.PractitionerResourceDao, error) {

	var insolvencyResourceDao models.InsolvencyResourceDao

	collection := m.db.Collection(m.CollectionName)
	practitionerCollection := m.db.Collection(PractitionerCollectionName)

	filter := bson.M{"transaction_id": transactionID}

	// Retrieve insolvency case from Mongo
	storedInsolvency := collection.FindOne(context.Background(), filter)

	err := storedInsolvency.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug("no insolvency resource found for transaction id", log.Data{"transaction_id": transactionID})
			return nil, nil, fmt.Errorf("there was a problem handling your request for transaction %s not found", transactionID)
		}
		log.Error(err)
		return nil, nil, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)
	}

	err = storedInsolvency.Decode(&insolvencyResourceDao)
	if err != nil {
		log.Error(err)
		return nil, nil, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)
	}

	// make a call to get insolvency practitioner details
	practitionersString := insolvencyResourceDao.Data.Practitioners
	if len(practitionersString) > 0 {
		practitionerResourceDaos, err := getInsolvencyPractitionersDetails(practitionersString, transactionID, practitionerCollection)
		if err != nil {
			log.Error(err)
			return nil, nil, fmt.Errorf("there was a problem getting insolvency and practitioners' details for transaction [%s]", err)
		}

		return &insolvencyResourceDao, practitionerResourceDaos, nil
	}

	return &insolvencyResourceDao, nil, nil
}

func (m *MongoService) GetPractitionerAppointment(practitionerID string, transactionID string) (*models.AppointmentResourceDao, error) {

	var appointmentResourceDao models.AppointmentResourceDao

	collection := m.db.Collection(AppointmentCollectionName)

	filter := bson.M{"practitioner_id": practitionerID}

	// Retrieve practitioner appointment from Mongo
	storedAppointment := collection.FindOne(context.Background(), filter)
	err := storedAppointment.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug("no appointment resource found for transaction id", log.Data{"transaction_id": transactionID})
			return nil, fmt.Errorf("there was a problem handling your request for transaction %s not found", transactionID)
		}
		log.Error(err)
		return nil, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)
	}

	err = storedAppointment.Decode(&appointmentResourceDao)
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)
	}

	return &appointmentResourceDao, nil
}

// CreatePractitionerResource stores an incoming practitioner to the practitioners collection
func (m *MongoService) CreatePractitionerResource(practitionerResourceDao *models.PractitionerResourceDao, transactionID string) (int, error) {

	collection := m.db.Collection(PractitionerCollectionName)

	_, err := collection.InsertOne(context.Background(), practitionerResourceDao)
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction %s (insert practitioner to collection)", transactionID)
	}

	return http.StatusCreated, nil
}

// CreateAppointmentResource stores a practitioner's appointment in appointment collection
func (m *MongoService) CreateAppointmentResource(appointmentResourceDao *models.AppointmentResourceDao) (int, error) {

	collection := m.db.Collection(AppointmentCollectionName)

	_, err := collection.InsertOne(context.Background(), appointmentResourceDao)
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, err
	}

	return http.StatusCreated, nil
}

// UpdateInsolvencyPractitioners updates the practitoners for an Insolvency Case
func (m *MongoService) UpdateInsolvencyPractitioners(insolvencyResource models.InsolvencyResourceDao, transactionID string) (int, error) {

	collection := m.db.Collection(m.CollectionName)

	filter := bson.M{"transaction_id": transactionID}
	update := bson.M{"$set": bson.M{"data.practitioners": insolvencyResource.Data.Practitioners}}

	//Update the insolvency practitioner
	_, err := collection.UpdateOne(context.Background(), filter, update)

	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem updating insolvency with practitioners for transaction %s", transactionID)
	}

	return http.StatusNoContent, nil
}

// GetPractitionersResource gets practitioner(s) for an practitioner collection by practitionerID(s)
func (m *MongoService) GetPractitionersResource(practitionerIDs []string) ([]models.PractitionerResourceDao, error) {

	collection := m.db.Collection(PractitionerCollectionName)

	practitionerResourceDao, err := getPractitioners(practitionerIDs, collection)
	if err != nil {
		log.Error(err)
		errMsg := fmt.Errorf("no practitioner found for practitioner id(s) %s", practitionerIDs)
		return nil, errMsg
	}

	return practitionerResourceDao, nil
}

// DeletePractitioner deletes a practitioner for an insolvency case with the specified transactionID and practitionerID
func (m *MongoService) DeletePractitioner(practitionerID string, transactionID string) (int, error) {

	var insolvencyDocumentToUpdate primitive.M
	var insolvencyResource models.InsolvencyResourceDao
	collection := m.db.Collection(m.CollectionName)

	// Choose specific transaction for insolvency case with practitioner to be removed
	filter := bson.M{"transaction_id": transactionID}

	// Check if insolvency case exists for specified transactionID
	storedInsolvency := collection.FindOne(context.Background(), filter)
	err := storedInsolvency.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Error(err)
			return http.StatusNotFound, fmt.Errorf("there was a problem handling your request for transaction id %s - insolvency case not found", transactionID)
		}
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction id %s", transactionID)
	}

	err = storedInsolvency.Decode(&insolvencyResource)
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, err
	}

	// get insolvency practitioners
	mappedInsolvencyPractitioners, _, err := utils.ConvertStringToMapObjectAndStringList(insolvencyResource.Data.Practitioners)
	if err != nil {
		log.Error(err)
		return http.StatusBadRequest, fmt.Errorf("there was a problem handling your request for transaction id %s no insolvency practitioners found", transactionID)
	}

	// check if practitionerID exists
	_, isPresent := mappedInsolvencyPractitioners[practitionerID]
	if isPresent {
		// delete slice out the practitioner to delete from the map before updating insolvency
		delete(mappedInsolvencyPractitioners, practitionerID)

		// delete practitioner appointment
		filterAppointmentToDelete := bson.M{"practitioner_id": practitionerID}
		appointmentCollection := m.db.Collection(AppointmentCollectionName)

		_, err = deleteCollection(practitionerID, filterAppointmentToDelete, appointmentCollection)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction id %s not able to delete practitioners appointment", transactionID)
		}

		// delete the practitioner after sucessfully deleted appointment.
		filterPractitionersToDelete := bson.M{"data.practitioner_id": practitionerID}
		practitionerCollection := m.db.Collection(PractitionerCollectionName)

		_, err = deleteCollection(practitionerID, filterPractitionersToDelete, practitionerCollection)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction id %s not able to delete practitioners", transactionID)
		}

		// update insolvency
		remainingPractitionerString, _ := utils.ConvertMapToString(mappedInsolvencyPractitioners)

		insolvencyToUpdate := bson.M{"transaction_id": transactionID}
		if remainingPractitionerString == "{}" {
			insolvencyDocumentToUpdate = bson.M{"$unset": bson.M{"data.practitioners": ""}}
		} else {
			insolvencyDocumentToUpdate = bson.M{"$set": bson.M{"data.practitioners": remainingPractitionerString}}
		}

		statusCode, err := updateCollection(insolvencyToUpdate, insolvencyDocumentToUpdate, collection)
		if err != nil {
			log.Error(err)
			return statusCode, fmt.Errorf("there was a problem handling your request for transaction id %s - not able to update insolvency practitioners %s", transactionID, practitionerID)
		}
	} else {
		return http.StatusNotFound, fmt.Errorf("there was a problem handling your request for transaction id %s not able to find practitioner %s to delete", transactionID, practitionerID)
	}

	return http.StatusNoContent, nil
}

// UpdatePractitionerAppointment adds appointment details into practitioner case with the specified transactionID and practitionerID
func (m *MongoService) UpdatePractitionerAppointment(appointmentResourceDao *models.AppointmentResourceDao, transactionID string, practitionerID string) (int, error) {
	//var practitionerMapAppointmentResource map[string]string
	practitionerMapAppointmentResource := make(map[string]string)

	// practitioner collection
	practitionerCollection := m.db.Collection(PractitionerCollectionName)

	// Create appointment link for a practitoner to be stored
	practitionerMapAppointmentResource[practitionerID] = appointmentResourceDao.Data.Links.Self

	// Convert the map to string
	appointmentLinks, err := utils.ConvertMapToString(practitionerMapAppointmentResource)
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction id %s - not able to convert mapped practitioner's appointment to string %s", transactionID, practitionerID)
	}

	// Choose specific practitioner to update with appointment
	practitionerToUpdate := bson.M{"data.practitioner_id": practitionerID}
	pratitionerDocumentToUpdate := bson.M{"$set": bson.M{"data.links.appointment": appointmentLinks}}

	//update practitioner collection with appointment link
	status, err := updateCollection(practitionerToUpdate, pratitionerDocumentToUpdate, practitionerCollection)
	if err != nil {
		log.Error(err)
		return status, fmt.Errorf("there was a problem handling your request for transaction id %s - not able to update practitioner's appointment %s", transactionID, practitionerID)
	}

	return status, err
}

// DeletePractitionerAppointment deletes an appointment for the specified transactionID and practitionerID
func (m *MongoService) DeletePractitionerAppointment(transactionID string, practitionerID string) (int, error) {

	var practitionerDocumentToUpdate primitive.M
	var practitionerResourceDao models.PractitionerResourceDao

	practitionerCollection := m.db.Collection(PractitionerCollectionName)
	appointmentCollection := m.db.Collection(AppointmentCollectionName)

	// get practitioner with specified practitionerID
	filter := bson.M{"data.practitioner_id": practitionerID}

	storedPractitioners := practitionerCollection.FindOne(context.Background(), filter)
	err := storedPractitioners.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Error(err)
			return http.StatusNotFound, fmt.Errorf("there was a problem handling your request for transaction id %s - practitioner not found", transactionID)
		}
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction id %s", transactionID)
	}

	err = storedPractitioners.Decode(&practitionerResourceDao)
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, err
	}

	// get practitioner appointment(s)
	mappedPractitionerAppointment, _, err := utils.ConvertStringToMapObjectAndStringList(practitionerResourceDao.Data.Links.Appointment)
	if err != nil {
		log.Error(err)
		return http.StatusNotFound, fmt.Errorf("there was a problem handling your request for transaction id %s - no practitioner's appointment found", transactionID)
	}

	// check if practitionerID exists
	_, isPresent := mappedPractitionerAppointment[practitionerID]
	if isPresent {
		// remove unwanted appointment from the slice
		delete(mappedPractitionerAppointment, practitionerID)

		//delete appointment
		filterAppointmentToDelete := bson.M{"practitioner_id": practitionerID}
		_, err = deleteCollection(practitionerID, filterAppointmentToDelete, appointmentCollection)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction id %s - not able to delete practitioners appointment", transactionID)
		}

		//update practitioners
		practitionerAppointmentString, _ := utils.ConvertMapToString(mappedPractitionerAppointment)

		practitionerToUpdate := bson.M{"data.practitioner_id": practitionerID}
		if practitionerAppointmentString == "{}" {
			practitionerDocumentToUpdate = bson.M{"$unset": bson.M{"data.links.appointment": ""}}
		} else {
			practitionerDocumentToUpdate = bson.M{"$set": bson.M{"data.links.appointment": practitionerAppointmentString}}
		}

		statusCode, err := updateCollection(practitionerToUpdate, practitionerDocumentToUpdate, practitionerCollection)
		if err != nil {
			log.Error(err)
			return statusCode, fmt.Errorf("there was a problem handling your request for transaction id %s - not able to update insolvency practitioners %s", transactionID, practitionerID)
		}
	} else {
		return http.StatusNotFound, fmt.Errorf("there was a problem handling your request for transaction id %s - not able to find practitioner %s to delete appointment", transactionID, practitionerID)
	}

	return http.StatusNoContent, nil
}

func (m *MongoService) AddAttachmentToInsolvencyResource(transactionID string, fileID string, attachmentType string) (*models.AttachmentResourceDao, error) {
	collection := m.db.Collection(m.CollectionName)

	filter := bson.M{"transaction_id": transactionID}

	attachmentDao := models.AttachmentResourceDao{
		ID:     fileID,
		Type:   attachmentType,
		Status: "submitted",
		Links: models.AttachmentResourceLinksDao{
			Self:     constants.TransactionsPath + transactionID + constants.AttachmentsPath + fileID,
			Download: constants.TransactionsPath + transactionID + constants.AttachmentsPath + fileID + "/download",
		},
	}

	update := bson.M{
		"$push": bson.M{
			"data.attachments": attachmentDao,
		},
	}

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return nil, fmt.Errorf("error updating mongo for transaction [%s]: [%s]", transactionID, err)
	}

	if result.MatchedCount != 1 || result.ModifiedCount != 1 {
		return nil, fmt.Errorf("no documents updated")
	}

	return &attachmentDao, nil
}

// GetAttachmentResources retrieves all attachments filed for an Insolvency Case
func (m *MongoService) GetAttachmentResources(transactionID string) ([]models.AttachmentResourceDao, error) {

	var insolvencyResource models.InsolvencyResourceDao
	collection := m.db.Collection(m.CollectionName)

	filter := bson.M{"transaction_id": transactionID}

	// Retrieve attachments from Mongo
	opts := options.FindOne().SetProjection(bson.M{"_id": 0, "data.attachments": 1})
	storedAttachments := collection.FindOne(context.Background(), filter, opts)
	err := storedAttachments.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug(fmt.Sprintf("no insolvency case found for transaction id: [%s]", transactionID))
			return nil, nil
		}
		log.Error(err)
		return nil, err
	}

	err = storedAttachments.Decode(&insolvencyResource)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// Return an empty array instead of nil to distinguish from insolvency case
	// not found
	if insolvencyResource.Data.Attachments == nil {
		return make([]models.AttachmentResourceDao, 0), nil
	}

	return insolvencyResource.Data.Attachments, nil
}

// GetAttachmentFromInsolvencyResource retrieves an attachment filed for an Insolvency Case
func (m *MongoService) GetAttachmentFromInsolvencyResource(transactionID string, fileID string) (models.AttachmentResourceDao, error) {

	var insolvencyResource models.InsolvencyResourceDao
	collection := m.db.Collection(m.CollectionName)

	filter := bson.M{
		"transaction_id":      transactionID,
		"data.attachments.id": fileID,
	}

	// Retrieve attachment from Mongo
	opts := options.FindOne().SetProjection(bson.M{"_id": 0, "data.attachments.$": 1})
	storedAttachment := collection.FindOne(context.Background(), filter, opts)
	err := storedAttachment.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug("no insolvency case found for transaction id", log.Data{"transaction_id": transactionID})
			return models.AttachmentResourceDao{}, nil
		}

		log.Error(err)
		return models.AttachmentResourceDao{}, err
	}

	err = storedAttachment.Decode(&insolvencyResource)
	if err != nil {
		log.Error(err)
		return models.AttachmentResourceDao{}, err
	}

	return insolvencyResource.Data.Attachments[0], nil
}

// DeleteAttachmentResource deletes an attachment filed for an Insolvency Case
func (m *MongoService) DeleteAttachmentResource(transactionID, attachmentID string) (int, error) {
	collection := m.db.Collection(m.CollectionName)

	// Choose specific transaction for insolvency case with attachment to be removed
	filter := bson.M{"transaction_id": transactionID}

	// Check if insolvency case exists for specified transactionID
	storedInsolvency := collection.FindOne(context.Background(), filter)
	err := storedInsolvency.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Error(err)
			return http.StatusNotFound, fmt.Errorf("there was a problem handling your request for transaction id [%s] - insolvency case not found", transactionID)
		}
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction id [%s]", transactionID)
	}

	// Choose specific attachment to delete
	pullQuery := bson.M{"data.attachments": bson.M{"id": attachmentID}}

	update, err := collection.UpdateOne(context.Background(), filter, bson.M{"$pull": pullQuery})
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction id [%s] - could not delete attachment with id [%s]", transactionID, attachmentID)
	}

	// Return error if Mongo could not update the document
	if update.ModifiedCount == 0 {
		err = fmt.Errorf("there was a problem handling your request for transaction id [%s] - attachment with id [%s] not found", transactionID, attachmentID)
		log.Error(err)
		return http.StatusNotFound, err
	}

	return http.StatusNoContent, nil
}

// UpdateAttachmentStatus updates the status of an attachment filed for an Insolvency Case
func (m *MongoService) UpdateAttachmentStatus(transactionID, attachmentID string, avStatus string) (int, error) {

	var insolvencyResource models.InsolvencyResourceDao
	collection := m.db.Collection(m.CollectionName)

	// Choose specific transaction for insolvency case with attachment status to be updated
	filter := bson.M{
		"transaction_id":      transactionID,
		"data.attachments.id": attachmentID,
	}

	// Retrieve attachment from Mongo
	opts := options.FindOne().SetProjection(bson.M{"_id": 0, "data.attachments.$": 1})
	storedAttachment := collection.FindOne(context.Background(), filter, opts)
	err := storedAttachment.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Error(err)
			return http.StatusNotFound, fmt.Errorf("there was a problem handling your request for transaction id [%s] - insolvency case not found", transactionID)
		}
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction id [%s]", transactionID)
	}

	err = storedAttachment.Decode(&insolvencyResource)
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, err
	}

	if insolvencyResource.Data.Attachments[0].Status != "processed" && insolvencyResource.Data.Attachments[0].Status != avStatus {
		update := bson.M{"$set": bson.M{
			"data.attachments.$.status": avStatus,
		},
		}

		// Choose specific attachment status to update
		result, err := collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction id [%s] - could not update status of attachment with id [%s]", transactionID, attachmentID)
		}

		// Return error if Mongo could not update the document
		if result.ModifiedCount == 0 {
			err = fmt.Errorf("there was a problem handling your request for transaction id [%s] - attachment with id [%s] not found", transactionID, attachmentID)
			log.Error(err)
			return http.StatusNotFound, err
		}
	}

	return http.StatusNoContent, nil
}

// CreateResolutionResource stores the resolution for the insolvency case
// with the specified transactionID
func (m *MongoService) CreateResolutionResource(dao *models.ResolutionResourceDao, transactionID string) (int, error) {

	var insolvencyResource models.InsolvencyResourceDao
	collection := m.db.Collection(m.CollectionName)

	filter := bson.M{"transaction_id": transactionID}

	resolutionDao := models.ResolutionResourceDao{
		DateOfResolution: dao.DateOfResolution,
		Attachments:      dao.Attachments,
		Kind:             dao.Kind,
		Etag:             dao.Etag,
		Links:            dao.Links,
	}

	// Retrieve insolvency case from Mongo
	storedInsolvency := collection.FindOne(context.Background(), filter)
	err := storedInsolvency.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug("no insolvency resource found for transaction id", log.Data{"transaction_id": transactionID})
			return http.StatusNotFound, fmt.Errorf("there was a problem handling your request for transaction %s not found", transactionID)
		}
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)
	}

	err = storedInsolvency.Decode(&insolvencyResource)
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)
	}

	update := bson.M{
		"$set": bson.M{
			"data.resolution": resolutionDao,
		},
	}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)
	}

	return http.StatusCreated, nil
}

// CreateStatementOfAffairsResource stores the statement of affairs resource for the insolvency case
// with the specified transactionID
func (m *MongoService) CreateStatementOfAffairsResource(dao *models.StatementOfAffairsResourceDao, transactionID string) (int, error) {

	var insolvencyResource models.InsolvencyResourceDao
	collection := m.db.Collection(m.CollectionName)

	filter := bson.M{"transaction_id": transactionID}

	statementDao := models.StatementOfAffairsResourceDao{
		StatementDate: dao.StatementDate,
		Attachments:   dao.Attachments,
	}

	// Retrieve insolvency case from Mongo
	storedInsolvency := collection.FindOne(context.Background(), filter)
	err := storedInsolvency.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug("no insolvency resource found for transaction id", log.Data{"transaction_id": transactionID})
			return http.StatusNotFound, fmt.Errorf("there was a problem handling your request for transaction %s not found", transactionID)
		}
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)
	}

	err = storedInsolvency.Decode(&insolvencyResource)
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)
	}

	update := bson.M{
		"$set": bson.M{
			"data.statement-of-affairs": statementDao,
		},
	}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)
	}

	return http.StatusCreated, nil
}

// GetStatementOfAffairsResource retrieves the statement of affairs filed for an Insolvency Case
func (m *MongoService) GetStatementOfAffairsResource(transactionID string) (models.StatementOfAffairsResourceDao, error) {

	var insolvencyResource models.InsolvencyResourceDao
	collection := m.db.Collection(m.CollectionName)

	filter := bson.M{
		"transaction_id": transactionID,
	}

	// Retrieve statement of affairs from Mongo
	storedStatementOfAffairs := collection.FindOne(context.Background(), filter)
	err := storedStatementOfAffairs.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug("no insolvency case found for transaction id", log.Data{"transaction_id": transactionID})
			return models.StatementOfAffairsResourceDao{}, nil
		}

		log.Error(err)
		return models.StatementOfAffairsResourceDao{}, err
	}

	err = storedStatementOfAffairs.Decode(&insolvencyResource)
	if err != nil {
		log.Error(err)
		return models.StatementOfAffairsResourceDao{}, err
	}

	return *insolvencyResource.Data.StatementOfAffairs, nil
}

// DeleteStatementOfAffairsResource deletes the statement of affairs filed for an insolvency case
func (m *MongoService) DeleteStatementOfAffairsResource(transactionID string) (int, error) {

	collection := m.db.Collection(m.CollectionName)

	// Choose specific transaction for insolvency case with attachment to be removed
	filter := bson.M{"transaction_id": transactionID}

	// Check if insolvency case exists for specified transactionID
	storedInsolvency := collection.FindOne(context.Background(), filter)
	err := storedInsolvency.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Error(err)
			return http.StatusNotFound, fmt.Errorf("there was a problem handling your request for transaction id [%s] - insolvency case not found", transactionID)
		}
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction id [%s]", transactionID)
	}

	query := bson.M{"data.statement-of-affairs": ""}

	update, err := collection.UpdateOne(context.Background(), filter, bson.M{"$unset": query})
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction id [%s] - could not delete statement of affairs", transactionID)
	}

	// Return error if Mongo could not update the document
	if update.ModifiedCount == 0 {
		err = fmt.Errorf("there was a problem handling your request for transaction id [%s] - statement of affairs not found", transactionID)
		log.Error(err)
		return http.StatusNotFound, err
	}

	return http.StatusNoContent, nil
}

// CreateProgressReportResource stores the statement of affairs resource for the insolvency case
// with the specified transactionID
func (m *MongoService) CreateProgressReportResource(dao *models.ProgressReportResourceDao, transactionID string) (int, error) {

	var insolvencyResource models.InsolvencyResourceDao
	collection := m.db.Collection(m.CollectionName)

	filter := bson.M{"transaction_id": transactionID}

	progessReportDao := models.ProgressReportResourceDao{
		FromDate:    dao.FromDate,
		ToDate:      dao.ToDate,
		Attachments: dao.Attachments,
		Etag:        dao.Etag,
		Kind:        dao.Kind,
	}

	// Retrieve insolvency case from Mongo
	storedInsolvency := collection.FindOne(context.Background(), filter)
	err := storedInsolvency.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug("no insolvency resource found for transaction id", log.Data{"transaction_id": transactionID})
			return http.StatusNotFound, fmt.Errorf("there was a problem handling your request for transaction %s not found", transactionID)
		}
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)
	}

	err = storedInsolvency.Decode(&insolvencyResource)
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)
	}

	update := bson.M{
		"$set": bson.M{
			"data.progress-report": progessReportDao,
		},
	}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)
	}

	return http.StatusCreated, nil
}

// GetResolutionResource retrieves the resolution filed for an Insolvency Case
func (m *MongoService) GetResolutionResource(transactionID string) (models.ResolutionResourceDao, error) {

	var insolvencyResource models.InsolvencyResourceDao
	collection := m.db.Collection(m.CollectionName)

	filter := bson.M{
		"transaction_id": transactionID,
	}

	// Retrieve resolution from Mongo
	storedResolution := collection.FindOne(context.Background(), filter)
	err := storedResolution.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug("no insolvency case found for transaction id", log.Data{"transaction_id": transactionID})
			return models.ResolutionResourceDao{}, nil
		}

		log.Error(err)
		return models.ResolutionResourceDao{}, err
	}

	err = storedResolution.Decode(&insolvencyResource)
	if err != nil {
		log.Error(err)
		return models.ResolutionResourceDao{}, err
	}

	return *insolvencyResource.Data.Resolution, nil
}

// DeleteResolutionResource deletes an resource filed for an Insolvency Case
func (m *MongoService) DeleteResolutionResource(transactionID string) (int, error) {

	collection := m.db.Collection(m.CollectionName)

	// Choose specific transaction for insolvency case with attachment to be removed
	filter := bson.M{"transaction_id": transactionID}

	// Check if insolvency case exists for specified transactionID
	storedInsolvency := collection.FindOne(context.Background(), filter)
	err := storedInsolvency.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Error(err)
			return http.StatusNotFound, fmt.Errorf("there was a problem handling your request for transaction id [%s] - insolvency case not found", transactionID)
		}
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction id [%s]", transactionID)
	}

	// Choose specific attachment to delete

	query := bson.M{"data.resolution": ""}

	update, err := collection.UpdateOne(context.Background(), filter, bson.M{"$unset": query})
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction id [%s] - could not delete resolution", transactionID)
	}

	// Return error if Mongo could not update the document
	if update.ModifiedCount == 0 {
		err = fmt.Errorf("there was a problem handling your request for transaction id [%s] - resolution not found", transactionID)
		log.Error(err)
		return http.StatusNotFound, err
	}

	return http.StatusNoContent, nil
}
