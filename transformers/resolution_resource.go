package transformers

import (
	"fmt"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
)

// ResolutionResourceRequestToDB transforms a resolution request to a dao model
func ResolutionResourceRequestToDB(req *models.Resolution, transactionID string, helperService utils.HelperService) *models.ResolutionResourceDao {

	etag, err := helperService.GenerateEtag()

	if err != nil {
		log.Error(fmt.Errorf("error generating etag: [%s] and etag is empty", err))
		return nil
	}

	selfLink := fmt.Sprintf("%s", constants.TransactionsPath+transactionID+"/insolvency/resolution")

	dao := &models.ResolutionResourceDao{
		Etag:             etag,
		Kind:             "insolvency-resource#resolution",
		DateOfResolution: req.DateOfResolution,
		Attachments:      req.Attachments,
		Links: models.ResolutionResourceLinksDao{
			Self: selfLink,
		},
	}

	return dao
}

// ResolutionDaoToResponse transforms a resolution dao model to a response
func ResolutionDaoToResponse(resolution *models.ResolutionResourceDao) *models.ResolutionResource {
	return &models.ResolutionResource{
		Etag:             resolution.Etag,
		Kind:             resolution.Kind,
		DateOfResolution: resolution.DateOfResolution,
		Attachments:      resolution.Attachments,
		Links:            models.ResolutionResourceLinks(resolution.Links),
	}
}
