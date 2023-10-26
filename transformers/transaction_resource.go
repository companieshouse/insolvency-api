package transformers

import (
	"github.com/companieshouse/api-sdk-go/apicore"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/private-api-sdk-go/companieshouseapi"
)

// InsolvencyResourceDaoToTransactionResource takes the dao for an insolvency request and converts it to
// a transaction resource
func InsolvencyResourceDaoToTransactionResource(req *models.InsolvencyResourceDao) *companieshouseapi.Transaction {

	// Generate insolvency resource for the transaction
	transactionResource := make(map[string]*companieshouseapi.Resource)
	transactionResource[req.Data.Links.Self] = &companieshouseapi.Resource{
		Kind: req.Data.Kind,
		Links: companieshouseapi.Links{
			Resource:         req.Data.Links.Self,
			ValidationStatus: req.Data.Links.ValidationStatus,
		},
		Marshal: apicore.Marshal{},
	}

	transaction := &companieshouseapi.Transaction{
		Resources: transactionResource,
	}

	return transaction
}
