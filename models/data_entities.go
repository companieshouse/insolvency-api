package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// InsolvencyResourceDao contains the meta-data for the insolvency resource in Mongo
type InsolvencyResourceDao struct {
	ID    primitive.ObjectID         `bson:"_id"`
	Etag  string                     `bson:"etag"`
	Kind  string                     `bson:"kind"`
	Data  InsolvencyResourceDaoData  `bson:"data"`
	Links InsolvencyResourceLinksDao `bson:"links"`
}

// InsolvencyResourceDaoData contains the data for the insolvency resource in Mongo
type InsolvencyResourceDaoData struct {
	CompanyNumber string `bson:"company_number"`
	CaseType      string `bson:"case_type"`
	CompanyName   string `bson:"company_name"`
}

// InsolvencyResourceLinksDao contains the links for the insolvency resource
type InsolvencyResourceLinksDao struct {
	Self             string `bson:"self"`
	Transaction      string `bson:"transaction"`
	ValidationStatus string `bson:"validation_status"`
}
