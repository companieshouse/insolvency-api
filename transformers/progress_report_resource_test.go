package transformers

import (
	"testing"

	"github.com/companieshouse/insolvency-api/models"
	"github.com/smartystreets/goconvey/convey"
)

func TestUnitProgressReportResourceRequestToDB(t *testing.T) {
	convey.Convey("field mappings are correct", t, func() {
		dao := &models.ProgressReport{
			FromDate: "2021-06-06",
			ToDate:   "2021-06-07",
			Attachments: []string{
				"1234567890",
			},
		}

		response := ProgressReportResourceRequestToDB(dao)

		convey.So(response.FromDate, convey.ShouldEqual, dao.FromDate)
		convey.So(response.ToDate, convey.ShouldEqual, dao.ToDate)
		convey.So(response.Attachments, convey.ShouldResemble, dao.Attachments)
	})

	convey.Convey("field mappings are correct", t, func() {
		dao := &models.ProgressReport{
			FromDate: "2021-06-06",
			ToDate:   "2021-06-07",
			Attachments: []string{
				"1234567890",
			},
		}

		response := ProgressReportResourceRequestToDB(dao)

		convey.So(response.FromDate, convey.ShouldEqual, dao.FromDate)
		convey.So(response.ToDate, convey.ShouldEqual, dao.ToDate)
		convey.So(response.Attachments, convey.ShouldResemble, dao.Attachments)
	})
}

func TestUnitProgressReportDaoToResponse(t *testing.T) {
	convey.Convey("field mappings are correct", t, func() {
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

		convey.So(response.FromDate, convey.ShouldEqual, dao.FromDate)
		convey.So(response.ToDate, convey.ShouldEqual, dao.ToDate)
		convey.So(response.Attachments, convey.ShouldResemble, dao.Attachments)
		convey.So(response.Etag, convey.ShouldEqual, dao.Etag)
		convey.So(response.Kind, convey.ShouldEqual, dao.Kind)
	})
}
