package dao

import (
	"context"
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//updateCollection updates documents of collections
func updateCollection(filter bson.M, updateDocument bson.M, collection *mongo.Collection) (int, error) {
	update, err := collection.UpdateOne(context.Background(), filter, updateDocument)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Return error if Mongo could not update the document
	if update.ModifiedCount == 0 {
		err = fmt.Errorf("failed to update collection")
		return http.StatusNotFound, err
	}

	return http.StatusNoContent, nil
}

//delete from Collection
func deleteCollection(filter bson.M, collection *mongo.Collection) (*mongo.DeleteResult, error) {
	opts := options.Delete().SetHint(bson.M{"_id": 1})
	result, err := collection.DeleteMany(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func getInsolvencyPractitionersDetails(practitionerLinksMap models.InsolvencyResourcePractitionersDao, transactionID string, collection *mongo.Collection) ([]models.PractitionerResourceDao, error) {
	practitionerIDs := utils.GetMapKeysAsStringSlice(practitionerLinksMap)
	if len(practitionerIDs) == 0 {
		return nil, fmt.Errorf("no practitioners to retrieve")
	}

	// make a call to fetch all practitioners from the string array
	practitionerResourceDao, err := getPractitioners(practitionerIDs, collection)
	if err != nil {
		log.Debug(err.Error(), log.Data{"transaction_id": transactionID})
		return nil, err
	}

	return practitionerResourceDao, nil
}

func getPractitioners(practitionerIDs []string, collection *mongo.Collection) ([]models.PractitionerResourceDao, error) {
	var practitionerResourceDaos []models.PractitionerResourceDao

	matchQuery := bson.D{{"$match", bson.D{{"data.practitioner_id", bson.D{{"$in", practitionerIDs}}}}}}
	lookupQuery := bson.D{{"$lookup", bson.D{{"from", AppointmentCollectionName}, {"localField", "data.practitioner_id"}, {"foreignField", "practitioner_id"}, {"as", "data.appointment"}}}}
	unwindQuery := bson.D{{"$unwind", bson.D{{"path", "$data.appointment"}, {"preserveNullAndEmptyArrays", true}}}}

	// Retrieve practitioners and appointments from DB
	practitionerCursor, err := collection.Aggregate(context.Background(), mongo.Pipeline{matchQuery, lookupQuery, unwindQuery})
	if err != nil {
		return nil, err
	}

	for practitionerCursor.Next(context.Background()) {
		var practitionerResourceDao models.PractitionerResourceDao
		err := practitionerCursor.Decode(&practitionerResourceDao)
		if err != nil {
			errMsg := fmt.Errorf("error decoding models")
			return nil, errMsg
		}

		practitionerResourceDaos = append(practitionerResourceDaos, practitionerResourceDao)
	}

	return practitionerResourceDaos, nil
}
