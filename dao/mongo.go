package dao

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
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

// GetInsolvencyResource retrieves an insolvency case with the specified transactionID.
// If no insolvency resource is found then it returns nil with no error.
func (m *MongoService) GetInsolvencyResource(transactionID string) (*models.InsolvencyResourceDao, error) {

	var insolvencyResourceDao models.InsolvencyResourceDao

	collection := m.db.Collection(m.CollectionName)

	filter := bson.M{"transaction_id": transactionID}

	// Retrieve insolvency case from Mongo
	storedInsolvency := collection.FindOne(context.Background(), filter)
	err := storedInsolvency.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug(constants.MsgResourceNotFound, log.Data{"transaction_id": transactionID})
			return nil, nil
		}
		log.Error(err)
		return nil, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)
	}

	err = storedInsolvency.Decode(&insolvencyResourceDao)
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)
	}
	return &insolvencyResourceDao, nil
}

// GetInsolvencyAndExpandedPractitionerResources retrieves both the insolvency and practitioner resources,
// with the appointment details inline, for an insolvency case with the specified transactionID
// If no insolvency resource is found then it returns nil for both resources, with no error.
func (m *MongoService) GetInsolvencyAndExpandedPractitionerResources(transactionID string) (*models.InsolvencyResourceDao, []models.PractitionerResourceDao, error) {

	insolvencyResourceDao, err := m.GetInsolvencyResource(transactionID)

	if err != nil || insolvencyResourceDao == nil {
		return nil, nil, err
	}

	// make a call to get insolvency practitioner details
	if insolvencyResourceDao.Data.Practitioners != nil {
		practitionerCollection := m.db.Collection(PractitionerCollectionName)
		practitionerLinksMap := *insolvencyResourceDao.Data.Practitioners
		practitionerResourceDaos, err := getInsolvencyPractitionersDetails(practitionerLinksMap, transactionID, practitionerCollection)
		if err != nil {
			log.Error(err)
			return nil, nil, fmt.Errorf("there was a problem getting insolvency and practitioners' details for transaction [%s]", err)
		}

		return insolvencyResourceDao, practitionerResourceDaos, nil
	}

	return insolvencyResourceDao, nil, nil
}

// GetPractitionerAppointment will retrieve a practitioner appointment
func (m *MongoService) GetPractitionerAppointment(transactionID string, practitionerID string) (*models.AppointmentResourceDao, error) {

	var appointmentResourceDao models.AppointmentResourceDao

	collection := m.db.Collection(AppointmentCollectionName)

	filter := bson.M{
		"transaction_id":  transactionID,
		"practitioner_id": practitionerID,
	}

	// Retrieve practitioner appointment from Mongo
	storedAppointment := collection.FindOne(context.Background(), filter)
	err := storedAppointment.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug(constants.MsgResourceNotFound, log.Data{"transaction_id": transactionID})
			return nil, fmt.Errorf("there was a problem handling your request for transaction %s not found", transactionID)
		}
		log.Error(err)
		return nil, fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID)
	}

	err = storedAppointment.Decode(&appointmentResourceDao)
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID)
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

// AddPractitionerToInsolvencyResource will update insolvency by adding a link to a practitioner resource
func (m *MongoService) AddPractitionerToInsolvencyResource(transactionID string, practitionerID string, practitionerLink string) (int, error) {

	collection := m.db.Collection(m.CollectionName)

	filter := bson.M{"transaction_id": transactionID}
	update := bson.M{"$set": bson.M{"data.practitioners." + practitionerID: practitionerLink}}

	//Update the insolvency practitioner
	_, err := collection.UpdateOne(context.Background(), filter, update)

	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem adding a practitioner to insolvency resource for transaction %s", transactionID)
	}

	return http.StatusNoContent, nil
}

// GetSinglePractitionerResource gets a specific practitioner by transactionID & practitionerID.
// It first checks the insolvency collection to ensure the practitioner data is linked to by the
// insolvency resource with the specified transactionID (to avoid retrieving orphaned records),
// and also checks the transactionID in the practitioners collection.
// If no insolvency case is found for the transactionID, or the insolvency case does not link to
// the practitionerID, then the function returns nil with no error to indicate a 'not found' result.
func (m *MongoService) GetSinglePractitionerResource(transactionID string, practitionerID string) (*models.PractitionerResourceDao, error) {

	// check that the practitionerID is referenced from an insolvency case with the transactionID
	matchedIDs, err := checkIDsMatch(transactionID, practitionerID, m)
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID)
	}
	if !matchedIDs {
		log.Info(fmt.Sprintf("practitioner id [%s] not found for transaction id [%s]", practitionerID, transactionID))
		return nil, nil
	}

	var practitionerResourceDao models.PractitionerResourceDao
	practitionerCollection := m.db.Collection(PractitionerCollectionName)

	filter := bson.M{
		"transaction_id":       transactionID,
		"data.practitioner_id": practitionerID,
	}

	// Retrieve practitioner from Mongo
	storedPractitioner := practitionerCollection.FindOne(context.Background(), filter)
	log.Info(fmt.Sprint(storedPractitioner))
	err = storedPractitioner.Err()
	if err != nil {
		// not checking for ErrNoDocuments separately as it should be there based on checkIDsMatch
		log.Error(err)
		return nil, fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID)
	}

	err = storedPractitioner.Decode(&practitionerResourceDao)
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)
	}

	return &practitionerResourceDao, nil
}

// GetAllPractitionerResourcesForTransactionID gets practitioner(s) for an insolvency case by transactionID
// It queries the insolvency collection to find only practitionerIDs linked to by the insolvency resource
// with the specified transactionID as well as checking the transactionID in the practitioners collection.
// If no insolvency case is found for the transactionID, or the insolvency case contains no practitioner
// references, then the function returns nil with no error to indicate a 'not found' result.
func (m *MongoService) GetAllPractitionerResourcesForTransactionID(transactionID string) ([]models.PractitionerResourceDao, error) {

	// check that practitionerID is referenced from insolvency case with transactionID
	insolvencyDao, err := m.GetInsolvencyResource(transactionID)
	if err != nil {
		log.Error(fmt.Errorf("error getting insolvency resource for transaction id [%s]: [%s]", transactionID, err))
		return nil, fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID)
	}
	if insolvencyDao == nil || insolvencyDao.Data.Practitioners == nil {
		msg := fmt.Sprintf("no practitioners found for transaction id [%s]", transactionID)
		log.Info(msg)
		return nil, nil
	}
	practitionerLinksMap := *insolvencyDao.Data.Practitioners
	pracIDs := utils.GetMapKeysAsStringSlice(practitionerLinksMap)

	var practitionerResourceDaos []models.PractitionerResourceDao
	practitionerCollection := m.db.Collection(PractitionerCollectionName)

	filter := bson.M{
		"transaction_id":       transactionID,
		"data.practitioner_id": bson.M{"$in": pracIDs},
	}

	// Retrieve practitioners from Mongo
	storedPractitionerCursor, err := practitionerCollection.Find(context.Background(), filter)
	if err != nil {
		log.Error(fmt.Errorf("error getting practitioner resources for transaction id [%s]: [%s]", transactionID, err))
		return nil, fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID)
	}

	// decode all the retrieved data into practitionerResourceDaos
	err = storedPractitionerCursor.All(context.Background(), &practitionerResourceDaos)
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf("there was a problem handling your request for transaction %s", transactionID)
	}

	return practitionerResourceDaos, nil
}

// DeletePractitioner deletes a practitioner for an insolvency case with the specified transactionID and practitionerID
// Any appointment data for the practitioner is also deleted. If successful it returns http.StatusNoContent.
// If no insolvency case is found for the transactionID, or the insolvency case does not link to
// the practitionerID, then the function returns http.StatusNotFound with an error.
func (m *MongoService) DeletePractitioner(transactionID string, practitionerID string) (int, error) {

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

	if insolvencyResource.Data.Practitioners == nil {
		log.Debug("Cannot delete practitioner - no practitioners found for insolvency case", log.Data{"transaction_id": transactionID})
		return http.StatusNotFound, fmt.Errorf("there was a problem handling your request for transaction id %s no insolvency practitioners found", transactionID)
	}

	// get insolvency practitioner links
	practitionerLinksMap := *insolvencyResource.Data.Practitioners

	// check if practitionerID exists
	_, isPresent := practitionerLinksMap[practitionerID]
	if isPresent {

		// delete practitioner appointment
		filterAppointmentToDelete := bson.M{
			"transaction_id":  transactionID,
			"practitioner_id": practitionerID,
		}
		appointmentCollection := m.db.Collection(AppointmentCollectionName)

		_, err = deleteCollection(filterAppointmentToDelete, appointmentCollection)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction id %s not able to delete practitioners appointment", transactionID)
		}

		// delete the practitioner after sucessfully deleted appointment.
		filterPractitionerToDelete := bson.M{
			"transaction_id":       transactionID,
			"data.practitioner_id": practitionerID,
		}
		practitionerCollection := m.db.Collection(PractitionerCollectionName)

		_, err = deleteCollection(filterPractitionerToDelete, practitionerCollection)
		if err != nil {
			log.Error(err)
			return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction id %s not able to delete practitioners", transactionID)
		}

		// update insolvency
		insolvencyToUpdate := bson.M{"transaction_id": transactionID}
		if len(practitionerLinksMap) == 1 {
			insolvencyDocumentToUpdate = bson.M{"$unset": bson.M{"data.practitioners": ""}}
		} else {
			insolvencyDocumentToUpdate = bson.M{"$unset": bson.M{"data.practitioners." + practitionerID: ""}}
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
func (m *MongoService) UpdatePractitionerAppointment(appointmentResourceDao *models.AppointmentResourceDao, transactionID, practitionerID string) (int, error) {

	// practitioner collection
	practitionerCollection := m.db.Collection(PractitionerCollectionName)

	// Select specific practitioner and specify appointment link to add
	practitionerToUpdate := bson.M{
		"transaction_id":       transactionID,
		"data.practitioner_id": practitionerID,
	}
	practitionerDocumentToUpdate := bson.M{
		"$set": bson.M{
			"data.etag":              appointmentResourceDao.Data.Etag,
			"data.links.appointment": appointmentResourceDao.Data.Links.Self,
		},
	}

	//update practitioner collection with appointment link
	status, err := updateCollection(practitionerToUpdate, practitionerDocumentToUpdate, practitionerCollection)
	if err != nil {
		log.Error(err)
		return status, fmt.Errorf("there was a problem handling your request for transaction id %s - not able to update practitioner's appointment %s", transactionID, practitionerID)
	}

	return status, err
}

// DeletePractitionerAppointment deletes an appointment for the specified transactionID and practitionerID and removes
// the appointment reference from the practitioner collection. If successful it returns http.StatusNoContent.
// If no insolvency case is found for the transactionID, the insolvency case does not link to
// the practitionerID, or the practitionerID has no appointment, then the function returns http.StatusNotFound with an error.
func (m *MongoService) DeletePractitionerAppointment(transactionID, practitionerID, etag string) (int, error) {

	var practitionerDocumentToUpdate primitive.M
	var practitionerResourceDao models.PractitionerResourceDao

	practitionerCollection := m.db.Collection(PractitionerCollectionName)
	appointmentCollection := m.db.Collection(AppointmentCollectionName)

	// check that the practitionerID is referenced from an insolvency case with the transactionID
	matchedIDs, err := checkIDsMatch(transactionID, practitionerID, m)
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID)
	}
	if !matchedIDs {
		msg := fmt.Sprintf("practitioner id [%s] not found for transaction id [%s]", practitionerID, transactionID)
		log.Info(msg)
		return http.StatusNotFound, fmt.Errorf(msg)
	}

	// get practitioner with specified transactionID & practitionerID
	filter := bson.M{
		"transaction_id":       transactionID,
		"data.practitioner_id": practitionerID,
	}

	storedPractitioners := practitionerCollection.FindOne(context.Background(), filter)
	err = storedPractitioners.Err()
	if err != nil {
		// not treating not found case separately here as should exist based on checks above
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction id %s", transactionID)
	}

	err = storedPractitioners.Decode(&practitionerResourceDao)
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, err
	}

	// get practitioner appointment(s)
	if practitionerResourceDao.Data.Links.Appointment == "" {
		log.Debug("Cannot delete practitioner appointment - no appointment found for practitioner", log.Data{"transaction_id": transactionID})
		return http.StatusNotFound, fmt.Errorf("there was a problem handling your request for transaction id %s - no practitioner's appointment found", transactionID)
	}

	//delete appointment
	filterAppointmentToDelete := bson.M{
		"transaction_id":  transactionID,
		"practitioner_id": practitionerID,
	}
	_, err = deleteCollection(filterAppointmentToDelete, appointmentCollection)
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction id %s - not able to delete practitioners appointment", transactionID)
	}

	//remove appointment link from practitioner
	practitionerToUpdate := bson.M{
		"transaction_id":       transactionID,
		"data.practitioner_id": practitionerID,
	}
	practitionerDocumentToUpdate = bson.M{
		"$unset": bson.M{"data.links.appointment": ""},
		"$set":   bson.M{"data.etag": etag},
	}

	statusCode, err := updateCollection(practitionerToUpdate, practitionerDocumentToUpdate, practitionerCollection)
	if err != nil {
		log.Error(err)
		return statusCode, fmt.Errorf("there was a problem handling your request for transaction id %s - not able to update insolvency practitioners %s", transactionID, practitionerID)
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

// GetAttachmentResources retrieves all attachments filed for an Insolvency Case.
// If no insolvency case is found for the transaction id, returns a nil result.
// If an insolvency case is found but it has no attachments, returns an empty array.
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
			log.Debug(constants.MsgCaseNotFound, log.Data{"transaction_id": transactionID})
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
			return http.StatusNotFound, fmt.Errorf(constants.MsgCaseForTransactionNotFound, transactionID)
		}
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID)
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
			return http.StatusNotFound, fmt.Errorf(constants.MsgCaseForTransactionNotFound, transactionID)
		}
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID)
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
			log.Debug(constants.MsgResourceNotFound, log.Data{"transaction_id": transactionID})
			return http.StatusNotFound, fmt.Errorf(constants.MsgReqTransactionNotFound, transactionID)
		}
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID)
	}

	err = storedInsolvency.Decode(&insolvencyResource)
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID)
	}

	update := bson.M{
		"$set": bson.M{
			"data.resolution": resolutionDao,
		},
	}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID)
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
		Kind:          dao.Kind,
		Etag:          dao.Etag,
		Links:         dao.Links,
	}

	// Retrieve insolvency case from Mongo
	storedInsolvency := collection.FindOne(context.Background(), filter)
	err := storedInsolvency.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug(constants.MsgResourceNotFound, log.Data{"transaction_id": transactionID})
			return http.StatusNotFound, fmt.Errorf(constants.MsgReqTransactionNotFound, transactionID)
		}
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID)
	}

	err = storedInsolvency.Decode(&insolvencyResource)
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID)
	}

	update := bson.M{
		"$set": bson.M{
			"data.statement-of-affairs": statementDao,
		},
	}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID)
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

	// Retrieve insolvency resource from Mongo
	storedInsolvency := collection.FindOne(context.Background(), filter)
	err := storedInsolvency.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug(constants.MsgCaseNotFound, log.Data{"transaction_id": transactionID})
			return models.StatementOfAffairsResourceDao{}, nil
		}

		log.Error(err)
		return models.StatementOfAffairsResourceDao{}, err
	}

	err = storedInsolvency.Decode(&insolvencyResource)
	if err != nil {
		log.Error(err)
		return models.StatementOfAffairsResourceDao{}, err
	}
	if insolvencyResource.Data.StatementOfAffairs == nil {
		return models.StatementOfAffairsResourceDao{}, nil
	}

	return *insolvencyResource.Data.StatementOfAffairs, nil
}

// DeleteStatementOfAffairsResource deletes the statement of affairs filed for an insolvency case
func (m *MongoService) DeleteStatementOfAffairsResource(transactionID string) (int, error) {

	httpStatus, err := m.DeleteResource(transactionID, "statement-of-affairs")
	return httpStatus, err

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
		Links:       dao.Links,
	}

	// Retrieve insolvency case from Mongo
	storedInsolvency := collection.FindOne(context.Background(), filter)
	err := storedInsolvency.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug(constants.MsgResourceNotFound, log.Data{"transaction_id": transactionID})
			return http.StatusNotFound, fmt.Errorf(constants.MsgReqTransactionNotFound, transactionID)
		}
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID)
	}

	err = storedInsolvency.Decode(&insolvencyResource)
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID)
	}

	update := bson.M{
		"$set": bson.M{
			"data.progress-report": progessReportDao,
		},
	}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID)
	}

	return http.StatusCreated, nil
}

// GetProgressReportResource retrieves the progress report filed for an Insolvency Case
func (m *MongoService) GetProgressReportResource(transactionID string) (*models.ProgressReportResourceDao, error) {

	var insolvencyResource models.InsolvencyResourceDao
	collection := m.db.Collection(m.CollectionName)

	filter := bson.M{
		"transaction_id": transactionID,
	}

	// Retrieve insolvency resource from Mongo
	storedInsolvency := collection.FindOne(context.Background(), filter)
	err := storedInsolvency.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug(constants.MsgCaseNotFound, log.Data{"transaction_id": transactionID})
			return &models.ProgressReportResourceDao{}, nil
		}

		log.Error(err)
		return &models.ProgressReportResourceDao{}, err
	}

	err = storedInsolvency.Decode(&insolvencyResource)
	if err != nil {
		log.Error(err)
		return &models.ProgressReportResourceDao{}, err
	}
	if insolvencyResource.Data.ProgressReport == nil {
		return &models.ProgressReportResourceDao{}, nil
	}

	return insolvencyResource.Data.ProgressReport, nil
}

// DeleteProgressReportResource deletes the progress report filed for an insolvency case
func (m *MongoService) DeleteProgressReportResource(transactionID string) (int, error) {

	httpStatus, err := m.DeleteResource(transactionID, "progress-report")
	return httpStatus, err

}

// GetResolutionResource retrieves the resolution filed for an Insolvency Case
func (m *MongoService) GetResolutionResource(transactionID string) (models.ResolutionResourceDao, error) {

	var insolvencyResource models.InsolvencyResourceDao
	collection := m.db.Collection(m.CollectionName)

	filter := bson.M{
		"transaction_id": transactionID,
	}

	// Retrieve insolvency resource from Mongo
	storedInsolvency := collection.FindOne(context.Background(), filter)
	err := storedInsolvency.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug(constants.MsgCaseNotFound, log.Data{"transaction_id": transactionID})
			return models.ResolutionResourceDao{}, nil
		}

		log.Error(err)
		return models.ResolutionResourceDao{}, err
	}

	err = storedInsolvency.Decode(&insolvencyResource)
	if err != nil {
		log.Error(err)
		return models.ResolutionResourceDao{}, err
	}
	if insolvencyResource.Data.Resolution == nil {
		return models.ResolutionResourceDao{}, nil
	}

	return *insolvencyResource.Data.Resolution, nil
}

// DeleteResolutionResource deletes a resolution resource filed for an Insolvency Case
func (m *MongoService) DeleteResolutionResource(transactionID string) (int, error) {

	httpStatus, err := m.DeleteResource(transactionID, "resolution")
	return httpStatus, err

}

func (m *MongoService) DeleteResource(transactionID string, resType string) (int, error) {
	collection := m.db.Collection(m.CollectionName)

	// Choose specific transaction for insolvency case with attachment to be removed
	filter := bson.M{"transaction_id": transactionID}

	// Check if insolvency case exists for specified transactionID
	storedInsolvency := collection.FindOne(context.Background(), filter)
	err := storedInsolvency.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Error(err)
			return http.StatusNotFound, fmt.Errorf(constants.MsgCaseForTransactionNotFound, transactionID)
		}
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID)
	}

	// Choose specific attachment to delete
	query := bson.M{"data." + resType: ""}

	update, err := collection.UpdateOne(context.Background(), filter, bson.M{"$unset": query})
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf("there was a problem handling your request for transaction id [%s] - could not delete %v", transactionID, strings.ReplaceAll(resType, "-", " "))
	}

	// Return error if Mongo could not update the document
	if update.ModifiedCount == 0 {
		err = fmt.Errorf("there was a problem handling your request for transaction id [%s] - %v not found", transactionID, strings.ReplaceAll(resType, "-", " "))
		log.Error(err)
		return http.StatusNotFound, err
	}

	return http.StatusNoContent, nil

}
