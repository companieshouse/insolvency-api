package models

// CreatedInsolvencyResource is the entity returned in a successful creation of an insolvency resource
type CreatedInsolvencyResource struct {
	CompanyNumber string                         `json:"company_number"`
	CaseType      string                         `json:"case_type"`
	Etag          string                         `json:"etag"`
	Kind          string                         `json:"kind"`
	CompanyName   string                         `json:"company_name"`
	Links         CreatedInsolvencyResourceLinks `json:"links"`
}

// CreatedInsolvencyResourceLinks contains the links for the created insolvency resource
type CreatedInsolvencyResourceLinks struct {
	Self             string `json:"self"`
	Transaction      string `json:"transaction"`
	ValidationStatus string `json:"validation_status"`
}

// CreatedPractitionerResource is the entity returned in a successful creation of an practitioner resource
type CreatedPractitionerResource struct {
	IPCode          string                           `json:"ip_code"`
	FirstName       string                           `json:"first_name"`
	LastName        string                           `json:"last_name"`
	TelephoneNumber string                           `json:"telephone_number"`
	Email           string                           `json:"email"`
	Address         CreatedAddressResource           `json:"address"`
	Role            string                           `json:"role"`
	Links           CreatedPractitionerLinksResource `json:"links"`
}

// CreatedAddressResource contains the address fields for the created practitioner resource
type CreatedAddressResource struct {
	AddressLine1 string `json:"address_line_1"`
	AddressLine2 string `json:"address_line_2"`
	Country      string `json:"country"`
	Locality     string `json:"locality"`
	Region       string `json:"region"`
	PostalCode   string `json:"postal_code"`
}

type CreatedPractitionerLinksResource struct {
	Self string `json:"self"`
}

// ResponseResource is the object returned in an error case
type ResponseResource struct {
	Message string `json:"message"`
}

// NewMessageResponse - convenience function for creating a response resource
func NewMessageResponse(message string) *ResponseResource {
	return &ResponseResource{Message: message}
}
