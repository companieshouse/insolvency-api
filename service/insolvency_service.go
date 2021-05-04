package service

import (
	"fmt"
	"net/http"
	"strings"
	"time"

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
	// Retrieve insolvency resource from DB
	insolvencyResource, err := svc.GetInsolvencyResource(transactionID)
	if err != nil {
		log.ErrorR(req, fmt.Errorf("error getting pracititioner resources from DB: [%s]", err))
		return err, false
	}

	// Check if appointment date supplied is before to company incorporation date
	incorporatedOn, err := GetCompanyIncorporatedOn(insolvencyResource.Data.CompanyNumber, req)
	if err != nil {
		log.ErrorR(req, fmt.Errorf("error getting company details from DB: [%s]", err))
		return err, false
	}

	// Check if appointment date supplied is in the future or before company was incorporated
	err, ok := isValidAppointmentDate(appointedOn, incorporatedOn)
	if err != nil {
		err = fmt.Errorf("error parsing date: [%s]", err)
		log.ErrorR(req, err)
		return err, false
	}
	if !ok {
		return nil, false
	}

	// Check if appointment date supplied is different from stored appointment dates in DB
	for _, practitioner := range insolvencyResource.Data.Practitioners {
		if practitioner.Appointment != nil && practitioner.Appointment.AppointedOn != "" && practitioner.Appointment.AppointedOn != appointedOn {
			return nil, false
		}
	}
	return nil, true
}

// isValidAppointmentDate is a helper function to check if the appointment date
// supplied is after today or before the company was incorporated
func isValidAppointmentDate(appointedOn string, incorporatedOn string) (error, bool) {
	layout := "2006-01-02"
	today := time.Now()
	appointedDate, err := time.Parse(layout, appointedOn)
	if err != nil {
		log.Error(fmt.Errorf("error when parsing appointedOn to date: [%s]", err))
		return err, false
	}

	// Retrieve only the date portion of the incorporatedOn datetime string
	if idx := strings.Index(incorporatedOn, " "); idx != -1 {
		incorporatedOn = incorporatedOn[:idx]
	}

	incorporatedDate, err := time.Parse(layout, incorporatedOn)
	if err != nil {
		log.Error(fmt.Errorf("error when parsing incorporatedOn to date: [%s]", err))
		return err, false
	}

	// Check if appointedOn is in the future
	if today.Before(appointedDate) {
		return nil, false
	}

	// Check if appointedOn is before company was incorporated
	if appointedDate.Before(incorporatedDate) {
		return nil, false
	}

	return nil, true
}
