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
	update, err := collection.UpdateOne(context.Background(), filter, updateDocument)
	if err != nil {
		errMsg := fmt.Errorf("could not update practitioner appointment for practitionerID %s: %s", practitionerID, err)
		log.Error(errMsg)
		return http.StatusInternalServerError, errMsg
	}
	// Check if a match was found
	if update.MatchedCount == 0 {
		err = fmt.Errorf("item with transaction id %s or practitioner id %s does not exist", transactionID, practitionerID)
		log.Error(err)
		return http.StatusNotFound, err
	}
	// Check if Mongo updated the collection
	if update.ModifiedCount == 0 {
		err = fmt.Errorf("item with transaction id %s or practitioner id %s not updated", transactionID, practitionerID)
		log.Error(err)
		return http.StatusNotFound, err
	}

	return http.StatusNoContent, nil 
}
