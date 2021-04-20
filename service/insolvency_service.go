package service

import (
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/dao"
)

/*
// InsolvencyService contains the DAO for db access
type InsolvencyService struct {
	DAO    dao.Service
	Config *config.Config
}
*/

// CheckPractitionerAlreadyAppointed will check with the transaction api that the provided transaction id exists
func CheckPractitionerAlreadyAppointed(svc dao.Service, transactionID string, practitionerID string, req *http.Request) (error, bool) {
	practitionerResources, err := svc.GetPractitionerResources(transactionID)
	if err != nil {
		log.ErrorR(req, fmt.Errorf("error getting pracititioner resources from DB: [%s]", err))
		return err, false
	}
	for _, v := range practitionerResources {
		if v.ID == practitionerID && v.Appointment.AppointedOn != "" {
			log.Info("practitioner ID [%s] already appointed to transaction ID [%s]")
			return nil, true
		}
	}
	return nil, false
}
