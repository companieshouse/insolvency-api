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
)

//UpdateCollection updates documents of collections
func UpdateCollection(transactionID string, practitionerID string, filter bson.M, updateDocument bson.M, collection *mongo.Collection) (int, error) {
	_, err := collection.UpdateOne(context.Background(), filter, updateDocument)
	if err != nil {
		errMsg := fmt.Errorf("could not update practitioner appointment for practitionerID %s: %s", practitionerID, err)
		log.Error(errMsg)
		return http.StatusInternalServerError, errMsg
	}

	return http.StatusNoContent, nil
}

func GetInsolvencyPractitionersDetails(practitionersString string, transactionID string, collection *mongo.Collection) ([]models.PractitionerResourceDao, error) {
	_, practitionerIDs, err := utils.ConvertStringToMapObjectAndStringList(practitionersString)
	if err != nil {
		return nil, err
	}

	// make a call to fetch all practitioners from the string array
	var practitionerResourceDao []models.PractitionerResourceDao
	practitionerResourceDao, err = getPractitioners(practitionerIDs, transactionID, collection)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return practitionerResourceDao, nil
}

func getPractitioners(practitionerIDs []string, transactionID string, collection *mongo.Collection) ([]models.PractitionerResourceDao, error) {
	var practitionerResourceDaos []models.PractitionerResourceDao
	var practitionerResourceDao models.PractitionerResourceDao

	matchQuery := bson.D{{"$match", bson.D{{"data.practitioner_id", bson.D{{"$in", practitionerIDs}}}}}}
	lookupQuery := bson.D{{"$lookup", bson.D{{"from", AppointmentCollectionName}, {"localField", "data.practitioner_id"}, {"foreignField", "practitioner_id"}, {"as", "data.appointment"}}}}
	unwindQuery := bson.D{{"$unwind", bson.D{{"path", "$data.appointment"}, {"preserveNullAndEmptyArrays", true}}}}

	// Retrieve practitioners and appointments from DB
	practitionerCursor, err := collection.Aggregate(context.Background(), mongo.Pipeline{lookupQuery, matchQuery, unwindQuery})
	if err != nil {
		log.Debug("no practitioner found for transaction id", log.Data{"transaction_id": transactionID})
		return nil, err
	}

	for practitionerCursor.Next(context.Background()) {
		err := practitionerCursor.Decode(&practitionerResourceDao)
		if err != nil {
			log.Debug("error decoding practitioners", log.Data{"transaction_id": transactionID})
			return nil, err
		}

		practitionerResourceDaos = append(practitionerResourceDaos, practitionerResourceDao)
	}

	return practitionerResourceDaos, nil
}
