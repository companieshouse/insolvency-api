package transformers

import (
	"fmt"
	"testing"

	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/models"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitInsolvencyResourceRequestToDB(t *testing.T) {
	Convey("field mappings are correct", t, func() {

		transactionID := "987654321"

		incomingRequest := &models.InsolvencyRequest{
			CompanyNumber: "12345678",
			CaseType:      constants.CVL.String(),
			CompanyName:   "companyName",
		}

		response := InsolvencyResourceRequestToDB(incomingRequest, transactionID)

		So(response.Data.CompanyNumber, ShouldEqual, incomingRequest.CompanyNumber)
		So(response.Data.CaseType, ShouldEqual, constants.CVL.String())
		So(response.Data.CompanyName, ShouldEqual, "companyName")
		So(response.Etag, ShouldNotBeNil)
		So(response.Kind, ShouldEqual, "insolvency-resource#insolvency-resource")
		So(response.Links.Self, ShouldEqual, fmt.Sprintf("/transactions/"+transactionID+"/insolvency"))
		So(response.Links.Transaction, ShouldEqual, fmt.Sprintf("/transactions/"+transactionID))
		So(response.Links.ValidationStatus, ShouldEqual, fmt.Sprintf("/transactions/"+transactionID+"/insolvency/validation-status"))
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
				Self:             "/transactions/" + transactionID + "/insolvency",
				Transaction:      fmt.Sprintf("/transactions/" + transactionID),
				ValidationStatus: fmt.Sprintf("/transactions/" + transactionID + "/insolvency/validation-status"),
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
