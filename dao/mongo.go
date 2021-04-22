package dao

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/companieshouse/chs.go/log"
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
func (m *MongoService) CreateInsolvencyResource(dao *models.InsolvencyResourceDao) error {

	dao.ID = primitive.NewObjectID()

	collection := m.db.Collection(m.CollectionName)
	_, err := collection.InsertOne(context.Background(), dao)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
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
			log.Debug("no insolvency resource found for transaction id", log.Data{"transaction_id": transactionID})
			return fmt.Errorf("there was a problem handling your request for transaction %s not found", transactionID), http.StatusNotFound
		}
		log.Error(err)
		return fmt.Errorf("there was a problem handling your request for transaction %s", transactionID), http.StatusInternalServerError
	}

	err = storedInsolvency.Decode(&insolvencyResource)
	if err != nil {
		log.Error(err)
		return fmt.Errorf("there was a problem handling your request for transaction %s", transactionID), http.StatusInternalServerError
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
		return fmt.Errorf("there was a problem handling your request for transaction %s", transactionID), http.StatusInternalServerError
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
			log.Debug("no insolvency case found for transaction id", log.Data{"transaction_id": transactionID})
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

// GetPractitionerResource gets a single practitioners for an insolvency case with the specified transactionID and practitionerID
func (m *MongoService) GetPractitionerResource(practitionerID string, transactionID string) (models.PractitionerResourceDao, error) {
	practitioners, err := m.GetPractitionerResources(transactionID)
	if err != nil {
		log.Error(err)
		return models.PractitionerResourceDao{}, err
	}

	for _, practitioner := range practitioners {
		if practitioner.ID == practitionerID {
			return practitioner, nil
		}
	}

	return models.PractitionerResourceDao{}, nil
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

	update, err := collection.UpdateOne(context.Background(), filter, updateDocument)
	if err != nil {
		log.Error(err)
		return fmt.Errorf("there was a problem handling your request - could not add practitioner appointment to id %s", practitionerID), http.StatusInternalServerError
	}
	// Check if Mongo updated the collection
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
