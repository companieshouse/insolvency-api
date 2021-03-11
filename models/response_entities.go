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

// ResponseResource is the object returned in an error case
type ResponseResource struct {
	Message string `json:"message"`
}

// NewMessageResponse - convenience function for creating a response resource
func NewMessageResponse(message string) *ResponseResource {
	return &ResponseResource{Message: message}
}
