package service

import (
	"fmt"
	"net/http"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/insolvency-api/dao"
)

// CheckPractitionerAlreadyAppointed will check with the transaction api that the provided transaction id exists
func CheckPractitionerAlreadyAppointed(svc dao.Service, transactionID string, practitionerID string, req *http.Request) (error, bool) {
	practitionerResources, err := svc.GetPractitionerResources(transactionID)
	if err != nil {
		log.ErrorR(req, fmt.Errorf("error getting pracititioner resources from DB: [%s]", err))
		return err, false
	}
	for _, practitioner := range practitionerResources {
		if practitioner.ID == practitionerID && practitioner.Appointment != nil && practitioner.Appointment.AppointedOn != "" {
			log.Info("practitioner ID [%s] already appointed to transaction ID [%s]")
			return nil, true
		}
	}
	return nil, false
}

// CheckAppointmentDateValid checks that the date is the same for all appointments
func CheckAppointmentDateValid(svc dao.Service, transactionID string, appointedOn string, req *http.Request) (error, bool) {
	practitionerResources, err := svc.GetPractitionerResources(transactionID)
	if err != nil {
		log.ErrorR(req, fmt.Errorf("error getting pracititioner resources from DB: [%s]", err))
		return err, false
	}
	for _, practitioner := range practitionerResources {
		if practitioner.Appointment != nil && practitioner.Appointment.AppointedOn != "" && practitioner.Appointment.AppointedOn != appointedOn {
			return nil, false
		}
	}
	return nil, true
}
