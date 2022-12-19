package transformers

import (
	"testing"

	"github.com/companieshouse/insolvency-api/models"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitProgressReportResourceRequestToDB(t *testing.T) {
	Convey("field mappings are correct", t, func() {
		dao := &models.ProgressReport{
			FromDate: "2021-06-06",
			ToDate:   "2021-06-07",
			Attachments: []string{
				"1234567890",
			},
		}

		response := ProgressReportResourceRequestToDB(dao)

		So(response.FromDate, ShouldEqual, dao.FromDate)
		So(response.ToDate, ShouldEqual, dao.ToDate)
		So(response.Attachments, ShouldResemble, dao.Attachments)
	})

	Convey("field mappings are correct", t, func() {
		dao := &models.ProgressReport{
			FromDate: "2021-06-06",
			ToDate:   "2021-06-07",
			Attachments: []string{
				"1234567890",
			},
		}

		response := ProgressReportResourceRequestToDB(dao)

		So(response.FromDate, ShouldEqual, dao.FromDate)
		So(response.ToDate, ShouldEqual, dao.ToDate)
		So(response.Attachments, ShouldResemble, dao.Attachments)
	})
}

func TestUnitProgressReportDaoToResponse(t *testing.T) {
	Convey("field mappings are correct", t, func() {
		dao := &models.ProgressReportResourceDao{
			FromDate: "2021-06-06",
			ToDate:   "2021-06-07",
			Attachments: []string{
				"1234567890",
			},
			Etag: "123",
			Kind: "abc",
		}

		response := ProgressReportDaoToResponse(dao)

		So(response.FromDate, ShouldEqual, dao.FromDate)
		So(response.ToDate, ShouldEqual, dao.ToDate)
		So(response.Attachments, ShouldResemble, dao.Attachments)
		So(response.Etag, ShouldEqual, dao.Etag)
		So(response.Kind, ShouldEqual, dao.Kind)
	})
}
