package transformers

import (
	"fmt"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
)

// InsolvencyResourceRequestToDB will take the input request from the REST call and transform it to a dao ready for
// insertion into the database
func InsolvencyResourceRequestToDB(req *models.InsolvencyRequest, transactionID string) *models.InsolvencyResourceDao {

	kind := "insolvency-resource#insolvency-resource"

	etag, err := utils.GenerateEtag()
	if err != nil {
		log.Error(fmt.Errorf("error generating etag: [%s]", err))
	}

	selfLink := fmt.Sprintf("/transactions/" + transactionID + "/insolvency")
	transactionLink := fmt.Sprintf("/transactions/" + transactionID)
	validationLink := fmt.Sprintf("/transactions/" + transactionID + "/insolvency/validation-status")

	dao := &models.InsolvencyResourceDao{
		Data: models.InsolvencyResourceDaoData{
			CompanyNumber: req.CompanyNumber,
			CaseType:      req.CaseType,
			CompanyName:   req.CompanyName,
		},
		Etag: etag,
		Kind: kind,
		Links: models.InsolvencyResourceLinksDao{
			Self:             selfLink,
			Transaction:      transactionLink,
			ValidationStatus: validationLink,
		},
	}

	return dao
}

// InsolvencyResourceDaoToCreatedResponse will transform an insolvency resource dao that has successfully been created into
// a http response entity
func InsolvencyResourceDaoToCreatedResponse(model *models.InsolvencyResourceDao) *models.CreatedInsolvencyResource {
	return &models.CreatedInsolvencyResource{
		CompanyNumber: model.Data.CompanyNumber,
		CaseType:      model.Data.CaseType,
		Etag:          model.Etag,
		Kind:          model.Kind,
		CompanyName:   model.Data.CompanyName,
		Links: models.CreatedInsolvencyResourceLinks{
			Self:             model.Links.Self,
			Transaction:      model.Links.Transaction,
			ValidationStatus: model.Links.ValidationStatus,
		},
	}
}
