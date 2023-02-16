package transformers

import (
	"fmt"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
)

// ProgressReportResourceRequestToDB transforms a progress report request to a dao model
func ProgressReportResourceRequestToDB(req *models.ProgressReport, transactionID string, helperService utils.HelperService) *models.ProgressReportResourceDao {

	etag, err := helperService.GenerateEtag()

	if err != nil {
		log.Error(fmt.Errorf("error generating etag: [%s] and etag is empty", err))
		return nil
	}

	isEtagValidated := helperService.HandleEtagGenerationValidation(err)

	if !isEtagValidated {
		return nil
	}

	selfLink := fmt.Sprintf(constants.TransactionsPath + transactionID + "insolvency/progress-report")

	dao := &models.ProgressReportResourceDao{
		FromDate:    req.FromDate,
		ToDate:      req.ToDate,
		Attachments: req.Attachments,
		Etag:        etag,
		Kind:        "insolvency-resource#progress-report",
		Links:       models.ProgressReportResourceLinksDao{Self: selfLink},
	}

	return dao
}

// ProgressReportDaoToResponse transforms a progress report dao model to a response
func ProgressReportDaoToResponse(progressReport *models.ProgressReportResourceDao) *models.ProgressReportResource {
	return &models.ProgressReportResource{
		FromDate:    progressReport.FromDate,
		ToDate:      progressReport.ToDate,
		Attachments: progressReport.Attachments,
		Etag:        progressReport.Etag,
		Kind:        progressReport.Kind,
		Links:       models.ProgressReportResourceLinks(progressReport.Links),
	}
}
