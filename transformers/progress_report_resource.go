package transformers

import (
	"github.com/companieshouse/insolvency-api/models"
)

// ProgressReportResourceRequestToDB transforms a progress report request to a dao model
func ProgressReportResourceRequestToDB(req *models.ProgressReport) *models.ProgressReportResourceDao {
	dao := &models.ProgressReportResourceDao{
		ProgressReportFromDate: req.ProgressReportFromDate,
		ProgressReportToDate:   req.ProgressReportToDate,
		Attachments:            req.Attachments,
	}

	return dao
}

// ProgressReportDaoToResponse transforms a progress report dao model to a response
func ProgressReportDaoToResponse(progressReport *models.ProgressReportResourceDao) *models.ProgressReportResource {
	return &models.ProgressReportResource{
		ProgressReportFromDate: progressReport.ProgressReportFromDate,
		ProgressReportToDate:   progressReport.ProgressReportToDate,
	}
}
