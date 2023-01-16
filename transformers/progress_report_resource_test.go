package transformers

import (
	"fmt"
	"testing"

	mock_dao "github.com/companieshouse/insolvency-api/mocks"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/golang/mock/gomock"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitProgressReportResourceRequestToDB(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	Convey("field mappings are correct", t, func() {

		mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

		dao := &models.ProgressReport{
			FromDate: "2021-06-06",
			ToDate:   "2021-06-07",
			Attachments: []string{
				"1234567890",
			},
		}

		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()
		mockHelperService.EXPECT().HandleEtagGenerationValidation(gomock.Any()).Return(true).AnyTimes()

		response := ProgressReportResourceRequestToDB(dao, mockHelperService)

		So(response.FromDate, ShouldEqual, dao.FromDate)
		So(response.ToDate, ShouldEqual, dao.ToDate)
		So(response.Attachments, ShouldResemble, dao.Attachments)
	})

	Convey("Etag failed to generate", t, func() {

		mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

		dao := &models.ProgressReport{
			FromDate: "2021-06-06",
			ToDate:   "2021-06-07",
			Attachments: []string{
				"1234567890",
			},
		}

		mockHelperService.EXPECT().GenerateEtag().Return("", fmt.Errorf("err"))

		response := ProgressReportResourceRequestToDB(dao, mockHelperService)

		So(response, ShouldBeNil)

	})

	Convey("Etag generated not validated", t, func() {

		mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

		dao := &models.ProgressReport{
			FromDate: "2021-06-06",
			ToDate:   "2021-06-07",
			Attachments: []string{
				"1234567890",
			},
		}

		mockHelperService.EXPECT().GenerateEtag().Return("", fmt.Errorf("err"))
		mockHelperService.EXPECT().HandleEtagGenerationValidation(gomock.Any()).Return(false).AnyTimes()

		response := ProgressReportResourceRequestToDB(dao, mockHelperService)

		So(response, ShouldBeNil)

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
