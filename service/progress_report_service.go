package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/dao"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
)

// ValidateProgressReportDetails checks that the incoming statement details are valid
func ValidateProgressReportDetails(svc dao.Service, progressReportStatementDao *models.ProgressReportResourceDao, transactionID string, req *http.Request) (string, error) {
	var errs []string

	// Check that the attachment has been submitted correctly
	if len(progressReportStatementDao.Attachments) == 0 || len(progressReportStatementDao.Attachments) > 1 {
		errs = append(errs, "please supply only one attachment")
	}

	// Check if statement date supplied is in the future or before company was incorporated
	insolvencyResource, err := svc.GetInsolvencyResource(transactionID)
	if err != nil {
		err = fmt.Errorf("error getting insolvency resource from DB: [%s]", err)
		log.ErrorR(req, err)
		return "", err
	}

	// Retrieve company incorporation date
	incorporatedOn, err := GetCompanyIncorporatedOn(insolvencyResource.Data.CompanyNumber, req)
	if err != nil {
		err = fmt.Errorf("error getting company details from DB: [%s]", err)
		log.ErrorR(req, err)
		return "", err
	}

	ok, err := utils.IsDateBetweenIncorporationAndNow(progressReportStatementDao.FromDate, incorporatedOn)
	if err != nil {
		err = fmt.Errorf("error parsing date: [%s]", err)
		log.ErrorR(req, err)
		return "", err
	}
	if !ok {
		errs = append(errs, fmt.Sprintf("from_date [%s] should not be in the future or before the company was incorporated", progressReportStatementDao.FromDate))
	}

	ok, err = utils.IsDateBetweenIncorporationAndNow(progressReportStatementDao.ToDate, incorporatedOn)
	if err != nil {
		err = fmt.Errorf("error parsing date: [%s]", err)
		log.ErrorR(req, err)
		return "", err
	}
	if !ok {
		errs = append(errs, fmt.Sprintf("to_date [%s] should not be in the future or before the company was incorporated", progressReportStatementDao.FromDate))
	}

	// Check if from date is after to date
	ok, err = utils.IsDateBeforeDate(progressReportStatementDao.FromDate, progressReportStatementDao.ToDate)
	if !ok {
		errs = append(errs, fmt.Sprintf("to_date [%s] should not be before from_date [%s]", progressReportStatementDao.ToDate, progressReportStatementDao.FromDate))
	}

	return strings.Join(errs, ", "), nil
}
