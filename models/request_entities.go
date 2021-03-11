package models

// InsolvencyRequest is the model that should be sent to the insolvency API when creating a new request.
type InsolvencyRequest struct {
	CompanyNumber string `json:"company_number" validate:"required"`
	CompanyName   string `json:"company_name" validate:"required"`
	CaseType      string `json:"case_type" validate:"required"`
}
