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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
func (m *MongoService) CreateInsolvencyResource(dao *models.InsolvencyResourceDao) (error, int) {

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
				return fmt.Errorf("there was a problem creating an insolvency case for this transaction id: %v", err), http.StatusInternalServerError
			}

			return nil, http.StatusCreated
		}

		// If there is an error but it is not ErrNoDocuments then an error happened checking the existence of the insolvency case
		log.Error(err)
		return fmt.Errorf("there was a problem creating an insolvency case for this transaction id: %v", err), http.StatusInternalServerError
	}

	// If there is no error retrieving the insolvency case, then it already exists
	log.Info("an insolvency case already exists for this transaction id")
	return fmt.Errorf("an insolvency case already exists for this transaction id"), http.StatusConflict
}

// GetInsolvencyResource retrieves all the data for an insolvency case with the specified transactionID
func (m *MongoService) GetInsolvencyResource(transactionID string) (models.InsolvencyResourceDao, error) {
	var insolvencyResource models.InsolvencyResourceDao
	collection := m.db.Collection(m.CollectionName)

	filter := bson.M{"transaction_id": transactionID}

	// Retrieve insolvency case from Mongo
	storedInsolvency := collection.FindOne(context.Background(), filter)
	err := storedInsolvency.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug(constants.MsgResourceNotFound, log.Data{"transaction_id": transactionID})
			return models.InsolvencyResourceDao{}, fmt.Errorf("there was a problem handling your request for transaction [%s] - insolvency case not found", transactionID)
		}
		log.Error(err)
		return models.InsolvencyResourceDao{}, fmt.Errorf("there was a problem handling your request for transaction [%s]", transactionID)
	}

	err = storedInsolvency.Decode(&insolvencyResource)
	if err != nil {
		log.Error(err)
		return models.InsolvencyResourceDao{}, fmt.Errorf("there was a problem handling your request for transaction [%s]", transactionID)
	}

	return insolvencyResource, nil
}

// CreatePractitionersResource stores an incoming practitioner to the list of practitioners for the insolvency case
// with the specified transactionID
func (m *MongoService) CreatePractitionersResource(dao *models.PractitionerResourceDao, transactionID string) (error, int) {
	var insolvencyResource models.InsolvencyResourceDao
	collection := m.db.Collection(m.CollectionName)

	filter := bson.M{"transaction_id": transactionID}

	// Retrieve insolvency case from Mongo
	storedInsolvency := collection.FindOne(context.Background(), filter)
	err := storedInsolvency.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug(constants.MsgResourceNotFound, log.Data{"transaction_id": transactionID})
			return fmt.Errorf(constants.MsgReqTransactionNotFound, transactionID), http.StatusNotFound
		}
		log.Error(err)
		return fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID), http.StatusInternalServerError
	}

	err = storedInsolvency.Decode(&insolvencyResource)
	if err != nil {
		log.Error(err)
		return fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID), http.StatusInternalServerError
	}

	maxPractitioners := 5

	// Check if there are already 5 practitioners in database
	if len(insolvencyResource.Data.Practitioners) == maxPractitioners {
		err = fmt.Errorf("there was a problem handling your request for transaction %s already has 5 practitioners", transactionID)
		log.Error(err)
		return err, http.StatusBadRequest
	}

	// Check if practitioner is already assigned to this case
	for _, storedPractitioner := range insolvencyResource.Data.Practitioners {
		if dao.IPCode == storedPractitioner.IPCode {
			err = fmt.Errorf("there was a problem handling your request for transaction %s - practitioner with IP Code %s already is already assigned to this case", transactionID, dao.IPCode)
			log.Error(err)
			return err, http.StatusBadRequest
		}
	}

	insolvencyResource.Data.Practitioners = append(insolvencyResource.Data.Practitioners, *dao)

	update := bson.M{
		"$set": insolvencyResource,
	}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error(err)
		return fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID), http.StatusInternalServerError
	}

	return nil, http.StatusCreated
}

// GetPractitionerResources gets a list of all practitioners for an insolvency case with the specified transactionID
func (m *MongoService) GetPractitionerResources(transactionID string) ([]models.PractitionerResourceDao, error) {
	var insolvencyResource models.InsolvencyResourceDao
	collection := m.db.Collection(m.CollectionName)

	filter := bson.M{"transaction_id": transactionID}

	// Retrieve insolvency case from Mongo
	opts := options.FindOne().SetProjection(bson.M{"_id": 0, "data.practitioners": 1})
	storedPractitioners := collection.FindOne(context.Background(), filter, opts)
	err := storedPractitioners.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug(constants.MsgCaseNotFound, log.Data{"transaction_id": transactionID})
			return nil, nil
		}
		log.Error(err)
		return nil, err
	}

	err = storedPractitioners.Decode(&insolvencyResource)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// Return an empty array instead of nil so the handler can check
	// that there are no practitioners
	if insolvencyResource.Data.Practitioners == nil {
		return make([]models.PractitionerResourceDao, 0), nil
	}

	return insolvencyResource.Data.Practitioners, nil
}

// GetPractitionerResource gets a single practitioner for an insolvency case with the specified transactionID and practitionerID
func (m *MongoService) GetPractitionerResource(practitionerID string, transactionID string) (models.PractitionerResourceDao, error) {

	var insolvencyResource models.InsolvencyResourceDao
	collection := m.db.Collection(m.CollectionName)

	filter := bson.M{
		"transaction_id":  transactionID,
		"data.practitioners.id":   practitionerID,
	}

	projection := bson.M{"_id": 0, "data.practitioners.$": 1}

	// Retrieve insolvency case from Mongo
	opts := options.FindOne().SetProjection(projection)
	practitioner := collection.FindOne(context.Background(), filter, opts)
	err := practitioner.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug(constants.MsgCaseNotFound, log.Data{"transaction_id": transactionID})
			return models.PractitionerResourceDao{}, nil
		}

		log.Error(err)
		return models.PractitionerResourceDao{}, err
	}
	err = practitioner.Decode(&insolvencyResource)
	if err != nil {
		log.Error(err)
		return models.PractitionerResourceDao{}, err
	}

	return insolvencyResource.Data.Practitioners[0], nil
}

// DeletePractitioner deletes a practitioner for an insolvency case with the specified transactionID and practitionerID
func (m *MongoService) DeletePractitioner(practitionerID string, transactionID string) (error, int) {
	collection := m.db.Collection(m.CollectionName)

	// Choose specific transaction for insolvency case with practitioner to be removed
	filter := bson.M{"transaction_id": transactionID}

	// Check if insolvency case exists for specified transactionID
	storedInsolvency := collection.FindOne(context.Background(), filter)
	err := storedInsolvency.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Error(err)
			return fmt.Errorf("there was a problem handling your request for transaction id %s - insolvency case not found", transactionID), http.StatusNotFound
		}
		log.Error(err)
		return fmt.Errorf("there was a problem handling your request for transaction id %s", transactionID), http.StatusInternalServerError
	}

	// Choose specific practitioner to delete
	pullQuery := bson.M{"data.practitioners": bson.M{"id": practitionerID}}

	update, err := collection.UpdateOne(context.Background(), filter, bson.M{"$pull": pullQuery})
	if err != nil {
		log.Error(err)
		return fmt.Errorf("there was a problem handling your request for transaction id %s - could not delete practitioner with id %s", transactionID, practitionerID), http.StatusInternalServerError
	}

	// Return error if Mongo could not update the document
	if update.ModifiedCount == 0 {
		err = fmt.Errorf("there was a problem handling your request for transaction id %s - practitioner with id %s not found", transactionID, practitionerID)
		log.Error(err)
		return err, http.StatusNotFound
	}

	return nil, http.StatusNoContent
}

// AppointPractitioner adds appointment details insolvency case with the specified transactionID and practitionerID
func (m *MongoService) AppointPractitioner(dao *models.AppointmentResourceDao, transactionID string, practitionerID string) (error, int) {

	collection := m.db.Collection(m.CollectionName)

	// Choose specific practitioner to update
	filter := bson.M{"transaction_id": transactionID, "data.practitioners.id": practitionerID}

	updateDocument := bson.M{"$set": bson.M{"data.practitioners.$.appointment": dao}}

	err, status := updatePractitioner(transactionID, practitionerID, filter, updateDocument, collection)

	return err, status
}

// DeletePractitionerAppointment deletes an appointment for the specified transactionID and practitionerID
func (m *MongoService) DeletePractitionerAppointment(transactionID string, practitionerID string) (error, int) {
	collection := m.db.Collection(m.CollectionName)

	// Choose specific practitioner to update
	filter := bson.M{"transaction_id": transactionID, "data.practitioners.id": practitionerID}

	updateDocument := bson.M{"$unset": bson.M{"data.practitioners.$.appointment": ""}}

	err, status := updatePractitioner(transactionID, practitionerID, filter, updateDocument, collection)

	return err, status
}

func updatePractitioner(transactionID string, practitionerID string, filter bson.M, updateDocument bson.M, collection *mongo.Collection) (error, int) {
	update, err := collection.UpdateOne(context.Background(), filter, updateDocument)
	if err != nil {
		errMsg := fmt.Errorf("could not update practitioner appointment for practitionerID %s: %s", practitionerID, err)
		log.Error(errMsg)
		return errMsg, http.StatusInternalServerError
	}
	// Check if a match was found
	if update.MatchedCount == 0 {
		err = fmt.Errorf("item with transaction id %s or practitioner id %s does not exist", transactionID, practitionerID)
		log.Error(err)
		return err, http.StatusNotFound
	}
	// Check if Mongo updated the collection
	if update.ModifiedCount == 0 {
		err = fmt.Errorf("item with transaction id %s or practitioner id %s not updated", transactionID, practitionerID)
		log.Error(err)
		return err, http.StatusNotFound
	}

	return nil, http.StatusNoContent
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

	// Retrieve statement of affairs from Mongo
	storedStatementOfAffairs := collection.FindOne(context.Background(), filter)
	err := storedStatementOfAffairs.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug(constants.MsgCaseNotFound, log.Data{"transaction_id": transactionID})
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
			return http.StatusNotFound, fmt.Errorf(constants.MsgCaseForTransactionNotFound, transactionID)
		}
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID)
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
func (m *MongoService) GetProgressReportResource(transactionID string) (models.ProgressReportResourceDao, error) {

	var insolvencyResource models.InsolvencyResourceDao
	collection := m.db.Collection(m.CollectionName)

	filter := bson.M{
		"transaction_id": transactionID,
	}

	// Retrieve progress report from Mongo
	storedProgressReport := collection.FindOne(context.Background(), filter)
	err := storedProgressReport.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug(constants.MsgCaseNotFound, log.Data{"transaction_id": transactionID})
			return models.ProgressReportResourceDao{}, nil
		}

		log.Error(err)
		return models.ProgressReportResourceDao{}, err
	}

	err = storedProgressReport.Decode(&insolvencyResource)
	if err != nil {
		log.Error(err)
		return models.ProgressReportResourceDao{}, err
	}

	return *insolvencyResource.Data.ProgressReport, nil
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
			log.Debug(constants.MsgCaseNotFound, log.Data{"transaction_id": transactionID})
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
			return http.StatusNotFound, fmt.Errorf(constants.MsgCaseForTransactionNotFound, transactionID)
		}
		log.Error(err)
		return http.StatusInternalServerError, fmt.Errorf(constants.MsgHandleReqTransactionId, transactionID)
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
