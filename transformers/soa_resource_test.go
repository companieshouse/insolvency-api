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

func TestUnitStatementOfAffairsResourceRequestToDB(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	Convey("field mappings are correct", t, func() {
		req := &models.StatementOfAffairs{
			StatementDate: "2021-06-06",
			Attachments: []string{
				"1234567890",
			},
		}

		dao := StatementOfAffairsResourceRequestToDB(req, "transactionID", utils.NewHelperService())

		So(dao.StatementDate, ShouldEqual, req.StatementDate)
		So(dao.Attachments, ShouldResemble, req.Attachments)
		So(dao.Etag, ShouldNotBeNil)
		So(dao.Kind, ShouldEqual, "insolvency-resource#statement-of-affairs")
		So(dao.Links.Self, ShouldNotBeNil)
	})

	Convey("Etag failed to generate", t, func() {

		mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

		req := &models.StatementOfAffairs{
			StatementDate: "2021-06-06",
			Attachments: []string{
				"1234567890",
			},
		}

		mockHelperService.EXPECT().GenerateEtag().Return("", fmt.Errorf("err"))

		dao := StatementOfAffairsResourceRequestToDB(req, "transactionID", mockHelperService)

		So(dao, ShouldBeNil)

	})

	Convey("Etag generated not validated", t, func() {

		mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

		req := &models.StatementOfAffairs{
			StatementDate: "2021-06-06",
			Attachments: []string{
				"1234567890",
			},
		}

		mockHelperService.EXPECT().GenerateEtag().Return("", fmt.Errorf("err"))
		mockHelperService.EXPECT().HandleEtagGenerationValidation(gomock.Any()).Return(false).AnyTimes()

		dao := StatementOfAffairsResourceRequestToDB(req, "transactionID", mockHelperService)

		So(dao, ShouldBeNil)

	})
}

func TestUnitStatementOfAffairsDaoToResponse(t *testing.T) {
	Convey("field mappings are correct", t, func() {
		dao := &models.StatementOfAffairsResourceDao{
			Etag:          "6f143c1f8109d834263eb764c5f020a0ae3ff78ee1789477179cb80f",
			Kind:          "insolvency-resource#statement-of-affairs",
			StatementDate: "2021-06-06",
			Attachments: []string{
				"1234567890",
			},
			Links: models.StatementOfAffairsResourceLinksDao{
				Self: "/transactions/12345678/insolvency/statement-of-affairs",
			},
		}

		response := StatementOfAffairsDaoToResponse(dao)

		So(response.StatementDate, ShouldEqual, dao.StatementDate)
		So(response.Attachments, ShouldResemble, dao.Attachments)
		So(response.Etag, ShouldEqual, dao.Etag)
		So(response.Kind, ShouldEqual, dao.Kind)
		So(response.Links.Self, ShouldEqual, dao.Links.Self)
	})
}
