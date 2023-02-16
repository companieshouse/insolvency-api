package transformers

import (
	"fmt"
	"testing"

	mock_dao "github.com/companieshouse/insolvency-api/mocks"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/companieshouse/insolvency-api/utils"
	"github.com/golang/mock/gomock"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitProgressReportResourceRequestToDB(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	Convey("field mappings are correct", t, func() {

		req := &models.ProgressReport{
			FromDate: "2021-06-07",
			ToDate:   "2022-06-06",
			Attachments: []string{
				"1234567890",
			},
		}

		dao := ProgressReportResourceRequestToDB(req, "transactionID", utils.NewHelperService())

		So(dao.FromDate, ShouldEqual, req.FromDate)
		So(dao.ToDate, ShouldEqual, req.ToDate)
		So(dao.Attachments, ShouldResemble, req.Attachments)
		So(dao.Etag, ShouldNotBeNil)
		So(dao.Kind, ShouldEqual, "insolvency-resource#progress-report")
		So(dao.Links.Self, ShouldNotBeNil)
	})

	Convey("Etag failed to generate", t, func() {

		mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

		req := &models.ProgressReport{
			FromDate: "2021-06-06",
			ToDate:   "2021-06-07",
			Attachments: []string{
				"1234567890",
			},
		}

		mockHelperService.EXPECT().GenerateEtag().Return("", fmt.Errorf("err"))

		dao := ProgressReportResourceRequestToDB(req, "transactionID", mockHelperService)

		So(dao, ShouldBeNil)

	})

	Convey("Etag generated not validated", t, func() {

		mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

		req := &models.ProgressReport{
			FromDate: "2021-06-06",
			ToDate:   "2021-06-07",
			Attachments: []string{
				"1234567890",
			},
		}

		mockHelperService.EXPECT().GenerateEtag().Return("", fmt.Errorf("err"))
		mockHelperService.EXPECT().HandleEtagGenerationValidation(gomock.Any()).Return(false).AnyTimes()

		dao := ProgressReportResourceRequestToDB(req, "transactionID", mockHelperService)

		So(dao, ShouldBeNil)

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
			Links: models.ProgressReportResourceLinksDao{
				Self: "transactions/1234567890/insolvency/progress-report",
			},
		}

		response := ProgressReportDaoToResponse(dao)

		So(response.FromDate, ShouldEqual, dao.FromDate)
		So(response.ToDate, ShouldEqual, dao.ToDate)
		So(response.Attachments, ShouldResemble, dao.Attachments)
		So(response.Etag, ShouldEqual, dao.Etag)
		So(response.Kind, ShouldEqual, dao.Kind)
		So(response.Links.Self, ShouldEqual, dao.Links.Self)
	})
}
