package models

// InsolvencyRequest is the model that should be sent to the insolvency API when creating a new request.
type InsolvencyRequest struct {
	CompanyNumber string `json:"company_number" validate:"required"`
	CompanyName   string `json:"company_name" validate:"required"`
	CaseType      string `json:"case_type" validate:"required"`
}

// PractitionerRequest is the model that should be sent when creating a new insolvency practitioner
type PractitionerRequest struct {
	IPCode          string  `json:"ip_code" validate:"required,number,max=8"`
	FirstName       string  `json:"first_name" validate:"required"`
	LastName        string  `json:"last_name" validate:"required"`
	TelephoneNumber string  `json:"telephone_number" validate:"omitempty"`
	Email           string  `json:"email" validate:"omitempty,email"`
	Address         Address `json:"address" validate:"required"`
	Role            string  `json:"role" validate:"required"`
}

// Address is the model to represent any addresses within the insolvency service
type Address struct {
	AddressLine1 string `json:"address_line_1" validate:"required"`
	AddressLine2 string `json:"address_line_2"`
	Country      string `json:"country"`
	Locality     string `json:"locality" validate:"required"`
	Region       string `json:"region"`
	PostalCode   string `json:"postal_code"`
}

// PractitionerAppointment is the model to represent appointment data for a practitioner
type PractitionerAppointment struct {
	AppointedOn string `json:"appointed_on" validate:"required,datetime=2006-01-02"`
	MadeBy      string `json:"made_by" validate:"required"`
}

// Attachment is the model to represent an attachment for an insolvency case
type Attachment struct {
	AttachmentType string `json:"attachment_type"`
	File           string `json:"file"`
}

// Resolution is the model to represent a resolution for an insolvency case
type Resolution struct {
	DateOfResolution string   `json:"date_of_resolution" validate:"required,datetime=2006-01-02"`
	Attachments      []string `json:"attachments" validate:"required"`
}
