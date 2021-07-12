package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// InsolvencyResourceDao contains the meta-data for the insolvency resource in Mongo
type InsolvencyResourceDao struct {
	ID            primitive.ObjectID         `bson:"_id"`
	TransactionID string                     `bson:"transaction_id"`
	Etag          string                     `bson:"etag"`
	Kind          string                     `bson:"kind"`
	Data          InsolvencyResourceDaoData  `bson:"data"`
	Links         InsolvencyResourceLinksDao `bson:"links"`
}

// InsolvencyResourceDaoData contains the data for the insolvency resource in Mongo
type InsolvencyResourceDaoData struct {
	CompanyNumber string                    `bson:"company_number"`
	CaseType      string                    `bson:"case_type"`
	CompanyName   string                    `bson:"company_name"`
	Practitioners []PractitionerResourceDao `bson:"practitioners,omitempty"`
	Attachments   []AttachmentResourceDao   `bson:"attachments,omitempty"`
	Resolution    ResolutionResourceDao     `bson:"resolution,omitempty"`
}

// InsolvencyResourceLinksDao contains the links for the insolvency resource
type InsolvencyResourceLinksDao struct {
	Self             string `bson:"self"`
	Transaction      string `bson:"transaction"`
	ValidationStatus string `bson:"validation_status"`
}

// PractitionerResourceDao contains the data for the practitioner resource in Mongo
type PractitionerResourceDao struct {
	ID              string                       `bson:"id"`
	IPCode          string                       `bson:"ip_code"`
	FirstName       string                       `bson:"first_name"`
	LastName        string                       `bson:"last_name"`
	TelephoneNumber string                       `bson:"telephone_number,omitempty"`
	Email           string                       `bson:"email,omitempty"`
	Address         AddressResourceDao           `bson:"address"`
	Role            string                       `bson:"role"`
	Links           PractitionerResourceLinksDao `bson:"links"`
	Appointment     *AppointmentResourceDao      `bson:"appointment,omitempty"`
}

// AppointmentResourceDao contains the appointment data for a practitioner
type AppointmentResourceDao struct {
	AppointedOn string                      `bson:"appointed_on,omitempty"`
	MadeBy      string                      `bson:"made_by,omitempty"`
	Links       AppointmentResourceLinksDao `bson:"links,omitempty"`
}

// AppointmentResourceLinksDao contains the Links data for an appointment
type AppointmentResourceLinksDao struct {
	Self string `bson:"self,omitempty"`
}

// AddressResourceDao contains the data for any addresses in Mongo
type AddressResourceDao struct {
	AddressLine1 string `bson:"address_line_1"`
	AddressLine2 string `bson:"address_line_2"`
	Country      string `bson:"country"`
	Locality     string `bson:"locality"`
	Region       string `bson:"region"`
	PostalCode   string `bson:"postal_code"`
}

// PractitionerResourceLinksDao contains the Links data for a practitioner
type PractitionerResourceLinksDao struct {
	Self string `bson:"self"`
}

// AttachmentResourceDao contains the data for the attachment DB resource
type AttachmentResourceDao struct {
	ID     string                     `bson:"id"`
	Type   string                     `bson:"type"`
	Status string                     `bson:"status"`
	Links  AttachmentResourceLinksDao `bson:"links"`
}

// AttachmentResourceLinksDao contains the Links data for an attachment
type AttachmentResourceLinksDao struct {
	Self     string `bson:"self"`
	Download string `bson:"download"`
}

// ResolutionResourceDao contains the data for the resolution DB resource
type ResolutionResourceDao struct {
	DateOfResolution string   `bson:"date_of_resolution"`
	Attachments      []string `bson:"attachments"`
}

// StatementOfAffairsResourceDao contains the data for the statement of affairs DB resource
type StatementOfAffairsResourceDao struct {
	StatementDate string   `bson:"statement_date"`
	Attachments   []string `bson:"attachments"`
}
