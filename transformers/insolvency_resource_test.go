package transformers

import (
	"fmt"
	"testing"

	"github.com/companieshouse/insolvency-api/constants"
	mock_dao "github.com/companieshouse/insolvency-api/mocks"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/golang/mock/gomock"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitInsolvencyResourceRequestToDB(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

	Convey("field mappings are correct", t, func() {

		transactionID := "987654321"

		incomingRequest := &models.InsolvencyRequest{
			CompanyNumber: "12345678",
			CaseType:      constants.CVL.String(),
			CompanyName:   "companyName",
		}

		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil)

		response := InsolvencyResourceRequestToDB(incomingRequest, transactionID, mockHelperService)

		So(response.TransactionID, ShouldEqual, transactionID)
		So(response.Data.CompanyNumber, ShouldEqual, incomingRequest.CompanyNumber)
		So(response.Data.CaseType, ShouldEqual, constants.CVL.String())
		So(response.Data.CompanyName, ShouldEqual, "companyName")
		So(response.Etag, ShouldNotBeNil)
		So(response.Kind, ShouldEqual, "insolvency-resource#insolvency-resource")
		So(response.Links.Self, ShouldEqual, fmt.Sprintf("%s", constants.TransactionsPath+transactionID+constants.InsolvencyPath))
		So(response.Links.Transaction, ShouldEqual, fmt.Sprintf("%s", constants.TransactionsPath+transactionID))
		So(response.Links.ValidationStatus, ShouldEqual, fmt.Sprintf("%s", constants.TransactionsPath+transactionID+"/insolvency/validation-status"))
	})

	Convey("Etag failed to generate", t, func() {

		transactionID := "987654321"

		incomingRequest := &models.InsolvencyRequest{
			CompanyNumber: "12345678",
			CaseType:      constants.CVL.String(),
			CompanyName:   "companyName",
		}

		mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

		mockHelperService.EXPECT().GenerateEtag().Return("", fmt.Errorf("err"))

		response := InsolvencyResourceRequestToDB(incomingRequest, transactionID, mockHelperService)

		So(response, ShouldBeNil)

	})
}

func TestUnitInsolvencyResourceDaoToCreatedResponse(t *testing.T) {
	Convey("field mappings are correct", t, func() {

		transactionID := "987654321"

		dao := &models.InsolvencyResourceDao{
			Etag: "etag123",
			Kind: "insolvency-resource#insolvency-resource",
			Data: models.InsolvencyResourceDaoData{
				CompanyName:   "companyName",
				CaseType:      constants.CVL.String(),
				CompanyNumber: "123456789",
			},
			Links: models.InsolvencyResourceLinksDao{
				Self:             constants.TransactionsPath + transactionID + constants.InsolvencyPath,
				Transaction:      fmt.Sprintf("%s", constants.TransactionsPath+transactionID),
				ValidationStatus: fmt.Sprintf("%s", constants.TransactionsPath+transactionID+constants.ValidationStatusPath),
			},
		}

		response := InsolvencyResourceDaoToCreatedResponse(dao)

		So(response.CompanyNumber, ShouldEqual, dao.Data.CompanyNumber)
		So(response.CaseType, ShouldEqual, dao.Data.CaseType)
		So(response.CompanyName, ShouldEqual, dao.Data.CompanyName)
		So(response.Etag, ShouldEqual, dao.Etag)
		So(response.Kind, ShouldEqual, dao.Kind)
		So(response.Links.Self, ShouldEqual, dao.Links.Self)
		So(response.Links.Transaction, ShouldEqual, dao.Links.Transaction)
		So(response.Links.ValidationStatus, ShouldEqual, dao.Links.ValidationStatus)
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
