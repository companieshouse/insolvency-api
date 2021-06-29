package transformers

import (
	"github.com/companieshouse/insolvency-api/models"
)

// ResolutionResourceRequestToDB transforms a resolution request to a dao model
func ResolutionResourceRequestToDB(req *models.Resolution) *models.ResolutionResourceDao {
	dao := &models.ResolutionResourceDao{
		DateOfResolution: req.DateOfResolution,
		Attachments:      req.Attachments,
	}

	return dao
}

// ResolutionDaoToResponse transforms a response dao model to a response
func ResolutionDaoToResponse(resolution *models.ResolutionResourceDao) *models.ResolutionResource {
	return &models.ResolutionResource{
		DateOfResolution: resolution.DateOfResolution,
	}
}
