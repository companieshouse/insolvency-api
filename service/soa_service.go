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

// ValidateStatementDetails checks that the incoming statement details are valid
func ValidateStatementDetails(svc dao.Service, statementDao *models.StatementOfAffairsResourceDao, transactionID string, req *http.Request) (string, error) {
	var errs []string

	if statementDao == nil {
		err := fmt.Errorf("nil DAO passed to service for validation")
		log.ErrorR(req, err)
		return "", err
	}

	// Check that the attachment has been submitted correctly
	if len(statementDao.Attachments) == 0 || len(statementDao.Attachments) > 1 {
		errs = append(errs, "please supply only one attachment")
	}

	// Check if statement date supplied is in the future or before company was incorporated
	insolvencyResource, _, err := svc.GetInsolvencyPractitionersResource(transactionID)
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

	ok, err := utils.IsDateBetweenIncorporationAndNow(statementDao.StatementDate, incorporatedOn)
	if err != nil {
		err = fmt.Errorf("error parsing date: [%s]", err)
		log.ErrorR(req, err)
		return "", err
	}
	if !ok {
		errs = append(errs, fmt.Sprintf("statement_date [%s] should not be in the future or before the company was incorporated", statementDao.StatementDate))
	}

	return strings.Join(errs, ", "), nil
}
