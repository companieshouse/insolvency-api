package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// PayableResourceDao is the persisted resource for payable items
type InsolvencyResourceDao struct {
	ID            primitive.ObjectID         `bson:"_id"`
  Etag          string                     `bson:"etag"`
  Kind          string                     `bson:"kind"`
	Data          InsolvencyResourceDaoData  `bson:"data"`
  Links         InsolvencyResourceLinksDao `bson:"links"`
}

type InsolvencyResourceDaoData struct {
  CompanyNumber string                     `bson:"company_number"`
  CaseType      string                     `bson:"case_type"`
  CompanyName   string                     `bson:"company_name"`
}

type InsolvencyResourceLinksDao struct {
	Self             string `bson:"self"`
	Transaction      string `bson:"transaction"`
	ValidationStatus string `bson:"validation_status"`
}
