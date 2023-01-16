package transformers

import (
	"fmt"
	"testing"

	mock_dao "github.com/companieshouse/insolvency-api/mocks"

	"github.com/companieshouse/insolvency-api/constants"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/golang/mock/gomock"

	. "github.com/smartystreets/goconvey/convey"
)

const kind = "insolvency-resource#resolution"

func TestUnitResolutionResourceRequestToDB(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	transactionID := "1234"

	req := &models.Resolution{
		DateOfResolution: "2021-06-06",
		Attachments: []string{
			"1234567890",
		},
	}

	Convey("field mappings are correct", t, func() {

		mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)
		
		mockHelperService.EXPECT().GenerateEtag().Return("etag", nil).AnyTimes()

		response := ResolutionResourceRequestToDB(req, transactionID, mockHelperService)

		So(response.Etag, ShouldNotBeNil)
		So(response.Kind, ShouldEqual, kind)
		So(response.DateOfResolution, ShouldEqual, req.DateOfResolution)
		So(response.Attachments, ShouldResemble, req.Attachments)
		So(response.Links.Self, ShouldEqual, fmt.Sprintf(constants.TransactionsPath+transactionID+"/insolvency/resolution"))
	})

	Convey("Etag failed to generate", t, func() {

		mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)
		mockHelperService.EXPECT().GenerateEtag().Return("", fmt.Errorf("err"))

		response := ResolutionResourceRequestToDB(req, transactionID, mockHelperService)

		So(response, ShouldBeNil)

	})

	Convey("Etag generated not validated", t, func() {

		mockHelperService := mock_dao.NewHelperMockHelperService(mockCtrl)

		mockHelperService.EXPECT().GenerateEtag().Return("", fmt.Errorf("err"))
		mockHelperService.EXPECT().HandleEtagGenerationValidation(gomock.Any()).Return(false).AnyTimes()

		response := ResolutionResourceRequestToDB(req, transactionID, mockHelperService)

		So(response, ShouldBeNil)

	})
}

func TestUnitResolutionDaoToResponse(t *testing.T) {
	Convey("field mappings are correct", t, func() {
		transactionID := "987654321"

		dao := &models.ResolutionResourceDao{
			Etag:             "etag123",
			Kind:             kind,
			DateOfResolution: "2021-06-06",
			Attachments: []string{
				"1234567890",
			},
			Links: models.ResolutionResourceLinksDao{
				Self: constants.TransactionsPath + transactionID + "/insolvency/resolution",
			},
		}

		response := ResolutionDaoToResponse(dao)

		So(response.Etag, ShouldNotBeNil)
		So(response.Kind, ShouldEqual, kind)
		So(response.DateOfResolution, ShouldEqual, dao.DateOfResolution)
		So(response.Links.Self, ShouldEqual, fmt.Sprintf(constants.TransactionsPath+transactionID+"/insolvency/resolution"))
	})
}
