package transformers

import (
	"github.com/companieshouse/insolvency-api/models"
)

// StatementOfAffairsResourceRequestToDB transforms a statement of affairs request to a dao model
func StatementOfAffairsResourceRequestToDB(req *models.StatementOfAffairs) *models.StatementOfAffairsResourceDao {
	dao := &models.StatementOfAffairsResourceDao{
		StatementDate: req.StatementDate,
		Attachments:   req.Attachments,
	}

	return dao
}

// StatementOfAffairsDaoToResponse transforms a statement of affairs dao model to a response
func StatementOfAffairsDaoToResponse(statement *models.StatementOfAffairsResourceDao) *models.StatementOfAffairsResource {
	return &models.StatementOfAffairsResource{
		StatementDate: statement.StatementDate,
	}
}
