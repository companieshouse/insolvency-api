package transformers

import (
	"testing"

	"github.com/companieshouse/insolvency-api/models"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitInsolvencyResourceDaoToTransactionResource(t *testing.T) {
	Convey("field mappings are correct", t, func() {

		linksModels := models.InsolvencyResourceLinksDao{
			Self:             "/transactions/87654321/insolvency",
			ValidationStatus: "/transactions/87654321/insolvency/validation-status",
		}

		incomingRequest := &models.InsolvencyResourceDao{}
		incomingRequest.Data.Links = linksModels

		response := InsolvencyResourceDaoToTransactionResource(incomingRequest)

		So(response.Resources, ShouldHaveLength, 1)
	})
}
