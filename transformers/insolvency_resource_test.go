package transformers

import (
	"fmt"
	"testing"

	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/golang/mock/gomock"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitInsolvencyResourceRequestToDB(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	Convey("field mappings are correct", t, func() {

		transactionID := "987654321"

		incomingRequest := &models.InsolvencyRequest{
			CompanyNumber: "12345678",
			CaseType:      constants.CVL.String(),
			CompanyName:   "companyName",
		}

		response := InsolvencyResourceRequestToDB(incomingRequest, transactionID)

		So(response.TransactionID, ShouldEqual, transactionID)
		So(response.Data.CompanyNumber, ShouldEqual, incomingRequest.CompanyNumber)
		So(response.Data.CaseType, ShouldEqual, constants.CVL.String())
		So(response.Data.CompanyName, ShouldEqual, "companyName")
		So(response.Data.Etag, ShouldNotBeNil)
		So(response.Data.Kind, ShouldEqual, "insolvency-resource#insolvency-resource")
		So(response.Data.Links.Self, ShouldEqual, fmt.Sprintf(constants.TransactionsPath+transactionID+constants.InsolvencyPath))
		So(response.Data.Links.Transaction, ShouldEqual, fmt.Sprintf(constants.TransactionsPath+transactionID))
		So(response.Data.Links.ValidationStatus, ShouldEqual, fmt.Sprintf(constants.TransactionsPath+transactionID+"/insolvency/validation-status"))
	})
}

func TestUnitInsolvencyResourceDaoToCreatedResponse(t *testing.T) {
	Convey("field mappings are correct", t, func() {

		transactionID := "987654321"

		dao := &models.InsolvencyResourceDao{}
		dao.Data.Etag = "etag123"
		dao.Data.Kind = "insolvency-resource#insolvency-resource"
		dao.Data.CompanyName = "companyName"
		dao.Data.CaseType = constants.CVL.String()
		dao.Data.CompanyNumber = "123456789"
		dao.Data.Links = models.InsolvencyResourceLinksDao{
			Self:             constants.TransactionsPath + transactionID + constants.InsolvencyPath,
			Transaction:      fmt.Sprintf(constants.TransactionsPath + transactionID),
			ValidationStatus: fmt.Sprintf(constants.TransactionsPath + transactionID + constants.ValidationStatusPath),
		}

		response := InsolvencyResourceDaoToCreatedResponse(dao)

		So(response.CompanyNumber, ShouldEqual, dao.Data.CompanyNumber)
		So(response.CaseType, ShouldEqual, dao.Data.CaseType)
		So(response.CompanyName, ShouldEqual, dao.Data.CompanyName)
		So(response.Etag, ShouldEqual, dao.Data.Etag)
		So(response.Kind, ShouldEqual, dao.Data.Kind)
		So(response.Links.Self, ShouldEqual, dao.Data.Links.Self)
		So(response.Links.Transaction, ShouldEqual, dao.Data.Links.Transaction)
		So(response.Links.ValidationStatus, ShouldEqual, dao.Data.Links.ValidationStatus)
	})
}

func TestUnitPractitionerResourceDaosToPractitionerFilingsResponse(t *testing.T) {
	Convey("field mappings are correct - no appointment", t, func() {

		practitionerResourceDao := models.PractitionerResourceDao{}
		practitionerResourceDao.Data.IPCode = "1111"
		practitionerResourceDao.Data.FirstName = "First"
		practitionerResourceDao.Data.LastName = "Last"
		practitionerResourceDao.Data.Etag = "etag123"
		practitionerResourceDao.Data.Kind = "insolvency-resource#insolvency-resource"
		practitionerResourceDao.Data.Links = models.PractitionerResourceLinksDao{
			Self: "/self/link",
		}
		practitionerResourceDao.Data.Address = models.AddressResourceDao{
			AddressLine1: "addressline1",
			Locality:     "locality",
		}
		practitionerResourceDao.Data.Role = constants.FinalLiquidator.String()
		daoList := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		responses := PractitionerResourceDaosToPractitionerFilingsResponse(daoList)

		So(responses[0].IPCode, ShouldEqual, daoList[0].Data.IPCode)
		So(responses[0].FirstName, ShouldEqual, daoList[0].Data.FirstName)
		So(responses[0].LastName, ShouldEqual, daoList[0].Data.LastName)
		So(responses[0].Address.AddressLine1, ShouldEqual, daoList[0].Data.Address.AddressLine1)
		So(responses[0].Address.Locality, ShouldEqual, daoList[0].Data.Address.Locality)
		So(responses[0].Role, ShouldEqual, daoList[0].Data.Role)
		So(responses[0].Etag, ShouldEqual, daoList[0].Data.Etag)
		So(responses[0].Kind, ShouldEqual, daoList[0].Data.Kind)
		So(responses[0].Links.Self, ShouldEqual, daoList[0].Data.Links.Self)
		So(responses[0].Links.Appointment, ShouldEqual, "")
	})

	Convey("field mappings are correct - with appointment", t, func() {

		practitionerResourceDao := models.PractitionerResourceDao{}
		practitionerResourceDao.Data.IPCode = "1111"
		practitionerResourceDao.Data.FirstName = "First"
		practitionerResourceDao.Data.LastName = "Last"
		practitionerResourceDao.Data.Etag = "etag123"
		practitionerResourceDao.Data.Kind = "insolvency-resource#insolvency-resource"
		practitionerResourceDao.Data.Links = models.PractitionerResourceLinksDao{
			Self:        "/self/link",
			Appointment: "/appointment/link",
		}
		practitionerResourceDao.Data.Address = models.AddressResourceDao{
			AddressLine1: "addressline1",
			Locality:     "locality",
		}
		practitionerResourceDao.Data.Role = constants.FinalLiquidator.String()
		appointmentResourceDao := &models.AppointmentResourceDao{}
		appointmentResourceDao.Data.AppointedOn = "2012-02-23"
		appointmentResourceDao.Data.MadeBy = "company"
		appointmentResourceDao.Data.Links = models.AppointmentResourceLinksDao{
			Self: "/appt/self/link",
		}
		practitionerResourceDao.Data.Appointment = appointmentResourceDao
		daoList := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		responses := PractitionerResourceDaosToPractitionerFilingsResponse(daoList)

		So(responses[0].IPCode, ShouldEqual, daoList[0].Data.IPCode)
		So(responses[0].FirstName, ShouldEqual, daoList[0].Data.FirstName)
		So(responses[0].LastName, ShouldEqual, daoList[0].Data.LastName)
		So(responses[0].Address.AddressLine1, ShouldEqual, daoList[0].Data.Address.AddressLine1)
		So(responses[0].Address.Locality, ShouldEqual, daoList[0].Data.Address.Locality)
		So(responses[0].Role, ShouldEqual, daoList[0].Data.Role)
		So(responses[0].Etag, ShouldEqual, daoList[0].Data.Etag)
		So(responses[0].Kind, ShouldEqual, daoList[0].Data.Kind)
		So(responses[0].Links.Self, ShouldEqual, daoList[0].Data.Links.Self)
		So(responses[0].Links.Appointment, ShouldEqual, daoList[0].Data.Links.Appointment)
		So(responses[0].Appointment.AppointedOn, ShouldEqual, appointmentResourceDao.Data.AppointedOn)
		So(responses[0].Appointment.MadeBy, ShouldEqual, appointmentResourceDao.Data.MadeBy)
		So(responses[0].Appointment.Links.Self, ShouldEqual, appointmentResourceDao.Data.Links.Self)

	})
}
