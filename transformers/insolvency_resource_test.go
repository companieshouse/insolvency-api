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

func TestUnitAppointmentResourceDaoToAppointedResponse(t *testing.T) {
	Convey("field mappings are correct", t, func() {

		appointmentResourceDao := &models.AppointmentResourceDao{}
		appointmentResourceDao.Data.AppointedOn = "2012-02-23"
		appointmentResourceDao.Data.MadeBy = "company"
		appointmentResourceDao.Data.Links = models.AppointmentResourceLinksDao{
			Self: "/self/link",
		}

		response := AppointmentResourceDaoToAppointedResponse(appointmentResourceDao)

		So(response.AppointedOn, ShouldEqual, appointmentResourceDao.Data.AppointedOn)
		So(response.MadeBy, ShouldEqual, appointmentResourceDao.Data.MadeBy)
		So(response.Links.Self, ShouldEqual, appointmentResourceDao.Data.Links.Self)
	})
}
