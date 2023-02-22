package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// // InsolvencyResourceDto contains the meta-data for the insolvency resource in Mongo with links rather than real data
// type InsolvencyResourceDto struct {
// 	ID            primitive.ObjectID           `bson:"_id"`
// 	TransactionID string                       `bson:"transaction_id"`
// 	Data          InsolvencyResourceDaoDataDto `bson:"data"`
// }

// // InsolvencyResourceDaoDataDto contains the data for the insolvency resource in Mongo with links rather than real data
// type InsolvencyResourceDaoDataDto struct {
// 	CompanyNumber      string                         `bson:"company_number"`
// 	CaseType           string                         `bson:"case_type"`
// 	CompanyName        string                         `bson:"company_name"`
// 	Etag               string                         `bson:"etag"`
// 	Kind               string                         `bson:"kind"`
// 	Practitioners      string                         `bson:"practitioners,omitempty"`
// 	Links              InsolvencyResourceLinksDao     `bson:"links,omitempty"`
// 	Attachments        []AttachmentResourceDao        `bson:"attachments,omitempty"`
// 	Resolution         *ResolutionResourceDao         `bson:"resolution,omitempty"`
// 	StatementOfAffairs *StatementOfAffairsResourceDao `bson:"statement-of-affairs,omitempty"`
// 	ProgressReport     *ProgressReportResourceDao     `bson:"progress-report,omitempty"`
// }

// InsolvencyResourceDao contains the meta-data for the insolvency resource in Mongo
type InsolvencyResourceDao struct {
	ID            primitive.ObjectID `bson:"_id"`
	TransactionID string             `bson:"transaction_id"`
	Data          struct {
		CompanyNumber      string                         `bson:"company_number"`
		CaseType           string                         `bson:"case_type"`
		CompanyName        string                         `bson:"company_name"`
		Etag               string                         `bson:"etag"`
		Kind               string                         `bson:"kind"`
		Practitioners      string                         `bson:"practitioners,omitempty"`
		Links              InsolvencyResourceLinksDao     `bson:"links,omitempty"`
		Attachments        []AttachmentResourceDao        `bson:"attachments,omitempty"`
		Resolution         *ResolutionResourceDao         `bson:"resolution,omitempty"`
		StatementOfAffairs *StatementOfAffairsResourceDao `bson:"statement-of-affairs,omitempty"`
		ProgressReport     *ProgressReportResourceDao     `bson:"progress-report,omitempty"`
	}
}

// InsolvencyResourceDaoData contains the data for the insolvency resource in Mongo
type InsolvencyResourceDaoData struct {
	ID            primitive.ObjectID `bson:"_id"`
	TransactionID string             `bson:"transaction_id"`
	Data          struct {
		CompanyNumber      string                         `bson:"company_number"`
		CaseType           string                         `bson:"case_type"`
		CompanyName        string                         `bson:"company_name"`
		Etag               string                         `bson:"etag"`
		Kind               string                         `bson:"kind"`
		Practitioners      []PractitionerResourceDao      `bson:"practitioners,omitempty"`
		Links              InsolvencyResourceLinksDao     `bson:"links,omitempty"`
		Attachments        []AttachmentResourceDao        `bson:"attachments,omitempty"`
		Resolution         *ResolutionResourceDao         `bson:"resolution,omitempty"`
		StatementOfAffairs *StatementOfAffairsResourceDao `bson:"statement-of-affairs,omitempty"`
		ProgressReport     *ProgressReportResourceDao     `bson:"progress-report,omitempty"`
	}
}

// InsolvencyResourceLinksDao contains the links for the insolvency resource
type InsolvencyResourceLinksDao struct {
	Self             string `bson:"self"`
	Transaction      string `bson:"transaction"`
	ValidationStatus string `bson:"validation_status"`
}

// PractitionerResourceDao contains the data for the practitioner resource in Mongo
type PractitionerResourceDao struct {
	Data struct {
		PractitionerId  string                       `bson:"practitioner_id"`
		IPCode          string                       `bson:"ip_code"`
		FirstName       string                       `bson:"first_name"`
		LastName        string                       `bson:"last_name"`
		TelephoneNumber string                       `bson:"telephone_number,omitempty"`
		Email           string                       `bson:"email,omitempty"`
		Address         AddressResourceDao           `bson:"address"`
		Role            string                       `bson:"role"`
		Etag            string                       `bson:"etag"`
		Kind            string                       `bson:"kind"`
		Links           PractitionerResourceLinksDao `bson:"links"`
		Appointment     *AppointmentResourceDao      `bson:"appointment,omitempty"`
	}
}

// AppointmentResourceDao contains the data for the appointment resource in Mongo
type AppointmentResourceDao struct {
	PractitionerId string `bson:"practitioner_id"`
	Data           struct {
		AppointedOn string                      `bson:"appointed_on,omitempty"`
		MadeBy      string                      `bson:"made_by,omitempty"`
		Links       AppointmentResourceLinksDao `bson:"links,omitempty"`
		Etag        string                      `bson:"etag"`
		Kind        string                      `bson:"kind"`
	}
}

// AppointmentResourceLinksDao contains the Links data for an appointment
type AppointmentResourceLinksDao struct {
	Self string `bson:"self,omitempty"`
}

// AddressResourceDao contains the data for any addresses in Mongo
type AddressResourceDao struct {
	Premises     string `bson:"premises"`
	AddressLine1 string `bson:"address_line_1"`
	AddressLine2 string `bson:"address_line_2"`
	Country      string `bson:"country"`
	Locality     string `bson:"locality"`
	Region       string `bson:"region"`
	PostalCode   string `bson:"postal_code"`
	POBox        string `bson:"po_box"`
}

// PractitionerResourceLinksDao contains the Links data for a practitioner
type PractitionerResourceLinksDao struct {
	Self        string `bson:"self"`
	Appointment string `bson:"appointment,omitempty"`
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
	Etag             string                     `bson:"etag"`
	Kind             string                     `bson:"kind"`
	DateOfResolution string                     `bson:"date_of_resolution"`
	Attachments      []string                   `bson:"attachments"`
	Links            ResolutionResourceLinksDao `bson:"links"`
}

// ResolutionResourceLinksDao contains the Links data for a resolution
type ResolutionResourceLinksDao struct {
	Self string `bson:"self,omitempty"`
}

// StatementOfAffairsResourceDao contains the data for the statement of affairs DB resource
type StatementOfAffairsResourceDao struct {
	StatementDate string   `bson:"statement_date"`
	Attachments   []string `bson:"attachments"`
}

type ProgressReportResourceDao struct {
	FromDate    string   `bson:"from_date"`
	ToDate      string   `bson:"to_date"`
	Attachments []string `bson:"attachments"`
	Etag        string   `bson:"etag"`
	Kind        string   `bson:"kind"`
}
