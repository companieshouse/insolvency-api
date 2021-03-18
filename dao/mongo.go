package dao

import (
	"context"
	"errors"
	"fmt"
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

var (
	ErrorNotFound                    error
	ErrorPractitionerLimitReached    error
	ErrorPractitionerLimitWillExceed error
)

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

func (m *MongoService) CreatePractitionersResource(dao []models.PractitionerResourceDao, transactionID string) error {
	var insolvencyResource models.InsolvencyResourceDao
	collection := m.db.Collection(m.CollectionName)

	id, err := primitive.ObjectIDFromHex(transactionID)
	if err != nil {
		log.Error(err)
		return err
	}

	storedInsolvency := collection.FindOne(context.Background(), bson.M{"_id": id})
	err = storedInsolvency.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ErrorNotFound = fmt.Errorf(fmt.Sprintf("insolvency case with transactionID %s not found", transactionID))
			log.Debug("no insolvency resource found for transaction id", log.Data{"transaction_id": transactionID})
			return ErrorNotFound
		}
		log.Error(err)
		return err
	}

	err = storedInsolvency.Decode(&insolvencyResource)
	if err != nil {
		log.Error(err)
		return err
	}

	// Check if there are already 5 practitioners in database
	if len(insolvencyResource.Data.Practitioners) == 5 {
		ErrorPractitionerLimitReached = fmt.Errorf(fmt.Sprintf("insolvency case with transactionID %s has reached the max capacity of 5 practitioners", transactionID))
		log.Error(err)
		return ErrorPractitionerLimitReached
	}

	// Check if number of stored practitioners + number of incoming practitioners
	// is greater than 5
	if len(insolvencyResource.Data.Practitioners)+len(dao) > 5 {
		ErrorPractitionerLimitWillExceed = fmt.Errorf(fmt.Sprintf("insolvency case with transactionID %s will exceed the max capacity of 5 practitioners", transactionID))
		log.Error(err)
		return ErrorPractitionerLimitWillExceed
	}

	insolvencyResource.Data.Practitioners = append(insolvencyResource.Data.Practitioners, dao...)

	update := bson.M{
		"$set": insolvencyResource,
	}

	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": id}, update)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}
