package service

import (
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/go-sdk-manager/manager"
	"github.com/companieshouse/insolvency-api/models"
	"strings"
)

func CheckCompanyNameAlphaKey(companyProfileCompanyName string, insolvencyRequest *models.InsolvencyRequest, req *http.Request) (error, int) {

	api, err := manager.GetPrivateSDK(req)
	if err != nil {
		return fmt.Errorf("error creating private SDK to call alphakeyservice: [%v]", err.Error()), http.StatusInternalServerError
	}

	requestAlphaKeyResponse, err := api.AlphaKey.Get(insolvencyRequest.CompanyName).Do()
	if err != nil {
		log.ErrorR(req, fmt.Errorf("error communicating with alphakey service [%v]", err))
		return fmt.Errorf("error verifying company name"), requestAlphaKeyResponse.HTTPStatusCode
	}

	insolvencyRequestAlphaKey := requestAlphaKeyResponse.SameAsAlphaKey

	profileAlphaKeyResponse, err := api.AlphaKey.Get(companyProfileCompanyName).Do()
	if err != nil {
		log.ErrorR(req, fmt.Errorf("error communicating with alphakey service [%v]", err))
		return fmt.Errorf("error verifying company name"), requestAlphaKeyResponse.HTTPStatusCode
	}

	profileAlphaKey := profileAlphaKeyResponse.SameAsAlphaKey

	if !strings.EqualFold(insolvencyRequestAlphaKey, profileAlphaKey) {
		return fmt.Errorf("company names do not match - provided: [%s], expected: [%s]", insolvencyRequest.CompanyName, companyProfileCompanyName), http.StatusBadRequest
	}

	//no error
	return nil, requestAlphaKeyResponse.HTTPStatusCode
}
