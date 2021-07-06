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

// ValidateResolutionRequest checks that the incoming resolution request is valid
func ValidateResolutionRequest(resolution models.Resolution) string {
	var errs []string

	// Check that the attachment has been submitted correctly
	if len(resolution.Attachments) == 0 || len(resolution.Attachments) > 1 {
		errs = append(errs, "please supply only one attachment")
	}
	return strings.Join(errs, ", ")
}

// ValidateResolutionDate checks that the incoming resolution date is valid
func ValidateResolutionDate(svc dao.Service, resolution *models.ResolutionResourceDao, transactionID string, req *http.Request) (string, error) {
	var errs []string

	// Check if resolution date supplied is in the future or before company was incorporated
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

	ok, err := utils.IsValidDate(resolution.DateOfResolution, incorporatedOn)
	if err != nil {
		err = fmt.Errorf("error parsing date: [%s]", err)
		log.ErrorR(req, err)
		return "", err
	}
	if !ok {
		errs = append(errs, fmt.Sprintf("date_of_resolution [%s] should not be in the future or before the company was incorporated", resolution.DateOfResolution))
	}

	return strings.Join(errs, ", "), nil
}
