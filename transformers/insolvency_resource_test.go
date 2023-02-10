package transformers

import (
	"fmt"
	"testing"

	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/golang/mock/gomock"
	"go.mongodb.org/mongo-driver/bson/primitive"

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

		dao := &models.InsolvencyResourceDto{
			Data: models.InsolvencyResourceDaoDataDto{
				Etag:          "etag123",
				Kind:          "insolvency-resource#insolvency-resource",
				CompanyName:   "companyName",
				CaseType:      constants.CVL.String(),
				CompanyNumber: "123456789",
				Links: models.InsolvencyResourceLinksDao{
					Self:             constants.TransactionsPath + transactionID + constants.InsolvencyPath,
					Transaction:      fmt.Sprintf(constants.TransactionsPath + transactionID),
					ValidationStatus: fmt.Sprintf(constants.TransactionsPath + transactionID + constants.ValidationStatusPath),
				},
			},
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
		dao := &models.AppointmentResourceDao{
			AppointedOn: "2012-02-23",
			MadeBy:      "company",
			Links: models.AppointmentResourceLinksDao{
				Self: "/self/link",
			},
		}

		response := AppointmentResourceDaoToAppointedResponse(dao)
		So(response.AppointedOn, ShouldEqual, dao.AppointedOn)
		So(response.MadeBy, ShouldEqual, dao.MadeBy)
		So(response.Links.Self, ShouldEqual, dao.Links.Self)
	})
}

func TestUnitInsolvencyResourceDtoToInsolvencyResourceDao(t *testing.T) {
	Convey("field mappings are correct", t, func() {
		insolvencyResourceDto := models.InsolvencyResourceDto{
			ID:            primitive.NewObjectID(),
			TransactionID: "transaction_id",
			Data: models.InsolvencyResourceDaoDataDto{
				CompanyNumber: "company_number",
				CaseType:      "case_type",
				CompanyName:   "company_name",
				Etag:          "etag",
				Kind:          "kind",
				Practitioners: "practitioners,omitempty",
			},
		}

		response := InsolvencyResourceDtoToInsolvencyResourceDao(insolvencyResourceDto)
		So(response.ID, ShouldEqual, insolvencyResourceDto.ID)
		So(response.TransactionID, ShouldEqual, insolvencyResourceDto.TransactionID)
		So(response.Data.CompanyNumber, ShouldEqual, insolvencyResourceDto.Data.CompanyNumber)
	})
}
