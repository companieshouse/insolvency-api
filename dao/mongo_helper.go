package dao

import (
	"context"
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/log"
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

func Find(filter bson.M, collection *mongo.Collection, model interface{}, errs []error) (*mongo.SingleResult, interface{}, error) {
	// Retrieve case from Mongo
	storedCursor := collection.FindOne(context.Background(), filter)
	err := storedCursor.Err()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Debug(errs[0].Error())
			return nil, nil, errs[1]
		}
		log.Error(err)
		return nil, nil, errs[2]
	}

	if model == nil {
		return storedCursor, nil, nil
	}

	err = storedCursor.Decode(&model)
	if err != nil {
		log.Error(err)
		return nil, nil, errs[2]
	}

	return storedCursor, model, nil
}
