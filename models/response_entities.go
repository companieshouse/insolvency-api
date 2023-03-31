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
	PractitionerId  string                 `json:"practitioner_id"`
	IPCode          string                 `json:"ip_code"`
	FirstName       string                 `json:"first_name"`
	LastName        string                 `json:"last_name"`
	Email           string                 `json:"email"`
	TelephoneNumber string                 `json:"telephone_number"`
	Address         CreatedAddressResource `json:"address"`
	Role            string                 `json:"role"`
	Etag            string                 `json:"etag"`
	Kind            string                 `json:"kind"`
	Links           struct {
		Self        *string `json:"self,omitempty"`
		Appointment *string `json:"appointment,omitempty"`
	} `json:"links"`
	Appointment *AppointedPractitionerResource `json:"appointment,omitempty"`
}

// CreatedAddressResource contains the address fields for the created practitioner resource
type CreatedAddressResource struct {
	Premises     string `json:"premises"`
	AddressLine1 string `json:"address_line_1"`
	AddressLine2 string `json:"address_line_2"`
	Country      string `json:"country"`
	Locality     string `json:"locality"`
	Region       string `json:"region"`
	PostalCode   string `json:"postal_code"`
	POBox        string `json:"po_box"`
}

// CreatedPractitionerLinksResource contains the links details for a practitioner
type CreatedPractitionerLinksResource struct {
	Self string `json:"self"`
}

// AppointedPractitionerResource contains the details of an appointed practitioner
type AppointedPractitionerResource struct {
	AppointedOn string                             `json:"appointed_on"`
	MadeBy      string                             `json:"made_by"`
	Links       AppointedPractitionerLinksResource `json:"links"`
	Etag        string                             `json:"etag"`
	Kind        string                             `json:"kind"`
}

// AppointedPractitionerLinksResource contains the links details for a practitioner appointment
type AppointedPractitionerLinksResource struct {
	Self string `json:"self"`
}

// AttachmentResource contains the details of an attachment
type AttachmentResource struct {
	AttachmentType string                  `json:"attachment_type"`
	File           AttachmentFile          `json:"file"`
	Etag           string                  `json:"etag"`
	Kind           string                  `json:"kind"`
	Status         string                  `json:"status"`
	Links          AttachmentLinksResource `json:"links"`
}

// AttachmentFile contains the details of an attachment file
type AttachmentFile struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	AVStatus    string `json:"av_status,omitempty"`
}

// AttachmentLinksResource contains the details of the links associated with an attachment
type AttachmentLinksResource struct {
	Self     string `json:"self"`
	Download string `json:"download"`
}

// ResolutionResource contains the details of the resolution resource
type ResolutionResource struct {
	DateOfResolution string                  `json:"date_of_resolution"`
	Attachments      []string                `json:"attachments"`
	Etag             string                  `json:"etag"`
	Kind             string                  `json:"kind"`
	Links            ResolutionResourceLinks `json:"links"`
}

// ResolutionResourceLinks contains the links details for a resolution
type ResolutionResourceLinks struct {
	Self string `json:"self"`
}

// StatementOfAffairsResource contains the details of the statement of affairs resource

type StatementOfAffairsResource struct {
	StatementDate string                          `json:"statement_date"`
	Attachments   []string                        `json:"attachments"`
	Etag          string                          `json:"etag"`
	Kind          string                          `json:"kind"`
	Links         StatementOfAffairsResourceLinks `json:"links"`
}

// StatementOfAffairsResourceLinks contains the links details for a statement of affairs
type StatementOfAffairsResourceLinks struct {
	Self string `json:"self"`
}

// ProgressReportResource contains the details of the progress report resource
type ProgressReportResource struct {
	FromDate    string                      `json:"from_date"`
	ToDate      string                      `json:"to_date"`
	Attachments []string                    `json:"attachments"`
	Etag        string                      `json:"etag"`
	Kind        string                      `json:"kind"`
	Links       ProgressReportResourceLinks `json:"links"`
}

// ProgressReportResourceLinks contains the link details associated with a progress report
type ProgressReportResourceLinks struct {
	Self string `json:"self"`
}

// ValidationStatusResponse is the object returned when checking the validation of a case
type ValidationStatusResponse struct {
	IsValid bool                              `json:"is_valid"`
	Errors  []ValidationErrorResponseResource `json:"errors"`
}

// NewValidationStatusResponse - convenience function for creating a validation response resource
func NewValidationStatusResponse(isValid bool, errors *[]ValidationErrorResponseResource) *ValidationStatusResponse {
	return &ValidationStatusResponse{IsValid: isValid, Errors: *errors}
}

// ValidationErrorResponseResource contains the details of an error when checking the validation for closing a case - as expected by transaction api
type ValidationErrorResponseResource struct {
	Error        string `json:"error"`
	Location     string `json:"location"`
	LocationType string `json:"location_type"`
	Type         string `json:"type"`
}

// NewValidationErrorResponse - convenience function for creating validation error responses
func NewValidationErrorResponse(validationError, location string) *ValidationErrorResponseResource {
	return &ValidationErrorResponseResource{
		Error:        validationError,
		Location:     location,
		LocationType: "json-path",
		Type:         "ch:validation",
	}
}

// Filing represents filing details to be returned to the filing resource handler
type Filing struct {
	Data                  map[string]interface{} `json:"data"`
	Description           string                 `json:"description"`
	DescriptionIdentifier string                 `json:"description_identifier"`
	DescriptionValues     map[string]string      `json:"description_values"`
	Kind                  string                 `json:"kind"`
}

// NewFiling - convenience function for creating a filing resource
func NewFiling(data map[string]interface{}, description, descriptionIdentifier, kind string) *Filing {
	return &Filing{
		Data:                  data,
		Description:           description,
		DescriptionIdentifier: descriptionIdentifier,
		DescriptionValues:     nil,
		Kind:                  kind,
	}
}

// ResponseResource is the object returned in an error case
type ResponseResource struct {
	Message string `json:"message"`
}

// NewMessageResponse - convenience function for creating a response resource
func NewMessageResponse(message string) *ResponseResource {
	return &ResponseResource{Message: message}
}
