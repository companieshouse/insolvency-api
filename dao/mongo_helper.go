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
