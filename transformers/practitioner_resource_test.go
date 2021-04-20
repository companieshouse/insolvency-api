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

		response := PractitionerResourceRequestToDB(incomingRequest, transactionID)

		So(response.ID, ShouldNotBeBlank)
		So(response.IPCode, ShouldEqual, incomingRequest.IPCode)
		So(response.FirstName, ShouldEqual, incomingRequest.FirstName)
		So(response.LastName, ShouldEqual, incomingRequest.LastName)
		So(response.Address.AddressLine1, ShouldEqual, incomingRequest.Address.AddressLine1)
		So(response.Address.Locality, ShouldEqual, incomingRequest.Address.Locality)
		So(response.Role, ShouldEqual, incomingRequest.Role)
		So(response.Links.Self, ShouldEqual, fmt.Sprintf("/transactions/"+transactionID+"/insolvency/practitioners/"+response.ID))
	})
}

func TestUnitPractitionerResourceDaoToCreatedResponse(t *testing.T) {
	transactionID := "1234"
	id := "123"

	Convey("field mappings are correct", t, func() {
		dao := &models.PractitionerResourceDao{
			ID:        id,
			IPCode:    "1111",
			FirstName: "First",
			LastName:  "Last",
			Address: models.AddressResourceDao{
				AddressLine1: "addressline1",
				Locality:     "locality",
			},
			Role: constants.FinalLiquidator.String(),
			Links: models.PractitionerResourceLinksDao{
				Self: fmt.Sprintf("/transactions/" + transactionID + "/insolvency/practitioners/" + id),
			},
		}

		response := PractitionerResourceDaoToCreatedResponse(dao)

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
		daoList := []models.PractitionerResourceDao{
			{
				IPCode:    "1111",
				FirstName: "First",
				LastName:  "Last",
				Address: models.AddressResourceDao{
					AddressLine1: "addressline1",
					Locality:     "locality",
				},
				Role: constants.FinalLiquidator.String(),
			},
		}

		response := PractitionerResourceDaoListToCreatedResponseList(daoList)

		So(response[0].IPCode, ShouldEqual, daoList[0].IPCode)
		So(response[0].FirstName, ShouldEqual, daoList[0].FirstName)
		So(response[0].LastName, ShouldEqual, daoList[0].LastName)
		So(response[0].Address.AddressLine1, ShouldEqual, daoList[0].Address.AddressLine1)
		So(response[0].Address.Locality, ShouldEqual, daoList[0].Address.Locality)
		So(response[0].Role, ShouldEqual, daoList[0].Role)
	})
}
