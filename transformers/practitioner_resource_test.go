package transformers

import (
	"fmt"
	"testing"

	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/models"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitPractitionerResourceRequestToDB(t *testing.T) {
	Convey("field mappings are correct", t, func() {
		transactionID := "1234"
		practitionerID := "1234"

		incomingRequest := &models.PractitionerRequest{
			IPCode:    "1111",
			FirstName: "First",
			LastName:  "Last",
			Address: models.Address{
				AddressLine1: "addressline1",
				Locality:     "locality",
			},
			Role: constants.FinalLiquidator.String(),
		}

		response := PractitionerResourceRequestToDB(incomingRequest, practitionerID, transactionID)

		So(response.TransactionID, ShouldEqual, "1234")
		So(response.Data.IPCode, ShouldEqual, "00001111")
		So(response.Data.FirstName, ShouldEqual, incomingRequest.FirstName)
		So(response.Data.LastName, ShouldEqual, incomingRequest.LastName)
		So(response.Data.Address.AddressLine1, ShouldEqual, incomingRequest.Address.AddressLine1)
		So(response.Data.Address.Locality, ShouldEqual, incomingRequest.Address.Locality)
		So(response.Data.Role, ShouldEqual, incomingRequest.Role)
		So(response.Data.Links.Self, ShouldNotBeEmpty)
	})
}

func TestUnitPractitionerResourceDaoToCreatedResponse(t *testing.T) {
	transactionID := "1234"
	id := "123"

	Convey("field mappings are correct", t, func() {

		practitionerResourceDao := models.PractitionerResourceDao{}
		practitionerResourceDao.Data.IPCode = "1111"
		practitionerResourceDao.Data.FirstName = "First"
		practitionerResourceDao.Data.LastName = "Last"

		practitionerResourceDao.Data.Address = models.AddressResourceDao{
			AddressLine1: "addressline1",
			Locality:     "locality",
		}
		practitionerResourceDao.Data.Role = constants.FinalLiquidator.String()
		practitionerResourceDao.Data.Links = models.PractitionerResourceLinksDao{
			Self: fmt.Sprintf(constants.TransactionsPath + transactionID + constants.PractitionersPath + id),
		}

		dao := &practitionerResourceDao.Data

		response := PractitionerResourceDaoToCreatedResponse(&practitionerResourceDao)

		So(response.IPCode, ShouldEqual, dao.IPCode)
		So(response.FirstName, ShouldEqual, dao.FirstName)
		So(response.LastName, ShouldEqual, dao.LastName)
		So(response.Address.AddressLine1, ShouldEqual, dao.Address.AddressLine1)
		So(response.Address.Locality, ShouldEqual, dao.Address.Locality)
		So(response.Role, ShouldEqual, dao.Role)
		So(response.Links.Self, ShouldEqual, dao.Links.Self)
	})
}

func TestUnitPractitionerResourceDaoListToCreatedResponseList(t *testing.T) {
	Convey("field mappings are correct", t, func() {
		practitionerResourceDao := models.PractitionerResourceDao{}
		practitionerResourceDao.Data.IPCode = "1111"
		practitionerResourceDao.Data.FirstName = "First"
		practitionerResourceDao.Data.LastName = "Last"

		practitionerResourceDao.Data.Address = models.AddressResourceDao{
			AddressLine1: "addressline1",
			Locality:     "locality",
		}
		practitionerResourceDao.Data.Role = constants.FinalLiquidator.String()

		daoList := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		response := PractitionerResourceDaoListToCreatedResponseList(daoList)

		So(response[0].IPCode, ShouldEqual, daoList[0].Data.IPCode)
		So(response[0].FirstName, ShouldEqual, daoList[0].Data.FirstName)
		So(response[0].LastName, ShouldEqual, daoList[0].Data.LastName)
		So(response[0].Address.AddressLine1, ShouldEqual, daoList[0].Data.Address.AddressLine1)
		So(response[0].Address.Locality, ShouldEqual, daoList[0].Data.Address.Locality)
		So(response[0].Role, ShouldEqual, daoList[0].Data.Role)
	})
}

func TestUnitPractitionerAppointmentRequestToDB(t *testing.T) {
	Convey("field mappings are correct", t, func() {
		dao := &models.PractitionerAppointment{
			AppointedOn: "2012-02-23",
			MadeBy:      "company",
		}
		transactionID := "123"
		practitionerID := "456"

		response := PractitionerAppointmentRequestToDB(dao, transactionID, practitionerID)

		So(response.TransactionID, ShouldEqual, "123")
		So(response.Data.AppointedOn, ShouldEqual, dao.AppointedOn)
		So(response.Data.MadeBy, ShouldEqual, dao.MadeBy)
		So(response.Data.Links.Self, ShouldEqual, fmt.Sprintf(constants.TransactionsPath+transactionID+"/insolvency/practitioners/"+practitionerID+"/appointment"))
	})
}

func TestUnitPractitionerAppointmentDaoToResponse(t *testing.T) {
	Convey("field mappings are correct", t, func() {
		appointmentResourceDao := &models.AppointmentResourceDao{}
		appointmentResourceDao.Data.AppointedOn = "2012-02-23"
		appointmentResourceDao.Data.MadeBy = "company"
		appointmentResourceDao.Data.Links = models.AppointmentResourceLinksDao{
			Self: "/self/link",
		}

		response := PractitionerAppointmentDaoToResponse(appointmentResourceDao)

		So(response.AppointedOn, ShouldEqual, appointmentResourceDao.Data.AppointedOn)
		So(response.MadeBy, ShouldEqual, appointmentResourceDao.Data.MadeBy)
		So(response.Links.Self, ShouldEqual, appointmentResourceDao.Data.Links.Self)
	})
}
