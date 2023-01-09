package service

import (
	"net/http"
	"strings"

	"github.com/companieshouse/insolvency-api/dao"
	"github.com/companieshouse/insolvency-api/models"
)

// ValidateProgressReportDetails checks that the incoming statement details are valid
func ValidateProgressReportDetails(svc dao.Service, progressReportStatementDao *models.ProgressReportResourceDao, transactionID string, req *http.Request) (string, error) {
	var errs []string

	// Check that the attachment has been submitted correctly
	if len(progressReportStatementDao.Attachments) == 0 || len(progressReportStatementDao.Attachments) > 1 {
		errs = append(errs, "please supply only one attachment")
	}

	return strings.Join(errs, ", "), nil
}
