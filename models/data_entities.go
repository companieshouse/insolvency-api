package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// PayableResourceDao is the persisted resource for payable items
type InsolvencyResourceDao struct {
	ID            primitive.ObjectID         `bson:"_id"`
	CompanyNumber string                     `bson:"company_number"`
	CaseType      string                     `bson:"case_type"`
	Etag          string                     `bson:"etag"`
	Kind          string                     `bson:"kind"`
	CompanyName   string                     `bson:"company_name"`
	Links         InsolvencyResourceLinksDao `bson:"links"`
}

type InsolvencyResourceLinksDao struct {
	Self             string `bson:"self"`
	Transaction      string `bson:"transaction"`
	ValidationStatus string `bson:"validation_status"`
}
