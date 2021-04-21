package models

// InsolvencyRequest is the model that should be sent to the insolvency API when creating a new request.
type InsolvencyRequest struct {
	CompanyNumber string `json:"company_number" validate:"required"`
	CompanyName   string `json:"company_name" validate:"required"`
	CaseType      string `json:"case_type" validate:"required"`
}

// PractitionerRequest is the model that should be sent when creating a new insolvency practitioner
type PractitionerRequest struct {
	IPCode    string  `json:"ip_code" validate:"required"`
	FirstName string  `json:"first_name" validate:"required,name_rule"`
	LastName  string  `json:"last_name" validate:"required,name_rule"`
	Address   Address `json:"address" validate:"required"`
	Role      string  `json:"role" validate:"required"`
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
