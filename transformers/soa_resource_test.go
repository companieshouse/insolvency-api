package transformers

import (
	"testing"

	"github.com/companieshouse/insolvency-api/models"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitStatementOfAffairsResourceRequestToDB(t *testing.T) {
	Convey("field mappings are correct", t, func() {
		dao := &models.StatementOfAffairs{
			StatementDate: "2021-06-06",
			Attachments: []string{
				"1234567890",
			},
		}

		response := StatementOfAffairsResourceRequestToDB(dao)

		So(response.StatementDate, ShouldEqual, dao.StatementDate)
		So(response.Attachments, ShouldResemble, dao.Attachments)
	})
}

func TestUnitStatementOfAffairsDaoToResponse(t *testing.T) {
	Convey("field mappings are correct", t, func() {
		dao := &models.StatementOfAffairsResourceDao{
			StatementDate: "2021-06-06",
			Attachments: []string{
				"1234567890",
			},
		}

		response := StatementOfAffairsDaoToResponse(dao)

		So(response.StatementDate, ShouldEqual, dao.StatementDate)
	})
}
