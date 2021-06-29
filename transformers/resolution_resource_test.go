package transformers

import (
	"testing"

	"github.com/companieshouse/insolvency-api/models"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitResolutionResourceRequestToDB(t *testing.T) {
	Convey("field mappings are correct", t, func() {
		dao := &models.Resolution{
			DateOfResolution: "2021-06-06",
			Attachments: []string{
				"1234567890",
			},
		}

		response := ResolutionResourceRequestToDB(dao)

		So(response.DateOfResolution, ShouldEqual, dao.DateOfResolution)
		So(response.Attachments, ShouldResemble, dao.Attachments)
	})
}

func TestUnitResolutionDaoToResponse(t *testing.T) {
	Convey("field mappings are correct", t, func() {
		dao := &models.ResolutionResourceDao{
			DateOfResolution: "2021-06-06",
			Attachments: []string{
				"1234567890",
			},
		}

		response := ResolutionDaoToResponse(dao)

		So(response.DateOfResolution, ShouldEqual, dao.DateOfResolution)
	})
}
