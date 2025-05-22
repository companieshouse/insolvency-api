package transformers

import (
	"fmt"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
)

// StatementOfAffairsResourceRequestToDB transforms a statement of affairs request to a dao model
func StatementOfAffairsResourceRequestToDB(req *models.StatementOfAffairs, transactionID string, helperService utils.HelperService) *models.StatementOfAffairsResourceDao {

	etag, err := helperService.GenerateEtag()

	if err != nil {
		log.Error(fmt.Errorf("error generating etag: [%s] and etag is empty", err))
		return nil
	}

	selfLink := fmt.Sprintf("%s", constants.TransactionsPath+transactionID+"/insolvency/statement-of-affairs")

	dao := &models.StatementOfAffairsResourceDao{
		Etag:          etag,
		Kind:          "insolvency-resource#statement-of-affairs",
		StatementDate: req.StatementDate,
		Attachments:   req.Attachments,
		Links: models.StatementOfAffairsResourceLinksDao{
			Self: selfLink,
		},
	}

	return dao
}

// StatementOfAffairsDaoToResponse transforms a statement of affairs dao model to a response
func StatementOfAffairsDaoToResponse(statement *models.StatementOfAffairsResourceDao) *models.StatementOfAffairsResource {
	return &models.StatementOfAffairsResource{
		StatementDate: statement.StatementDate,
		Attachments:   statement.Attachments,
		Etag:          statement.Etag,
		Kind:          statement.Kind,
		Links:         models.StatementOfAffairsResourceLinks(statement.Links),
	}
}
