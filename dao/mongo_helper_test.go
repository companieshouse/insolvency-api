package dao

import (
	"fmt"
	"testing"

	"github.com/companieshouse/insolvency-api/mocks"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestUnit_getInsolvencyPractitionersDetails(t *testing.T) {
	t.Parallel()

	t.Run("no practitioners to retrieve", func(t *testing.T) {
		got, err := getInsolvencyPractitionersDetails(nil, "12345", nil)
		assert.Equal(t, fmt.Errorf("no practitioners to retrieve"), err)
		assert.Nil(t, got)
	})
}

func TestUnit_checkIDsMatch(t *testing.T) {

	insolvencyDao := models.InsolvencyResourceDao{}
	insolvencyDao.Data.CompanyNumber = "1234"
	insolvencyDao.Data.CaseType = "CVL"
	insolvencyDao.Data.CompanyName = "Company"

	Convey("error getting insolvency resource", t, func() {
		// setup
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// create mock response
		mockService.EXPECT().GetInsolvencyResource("transactionID").Return(nil, fmt.Errorf("mocked error"))

		result, err := checkIDsMatch("transactionID", "practitionerID", mockService)

		So(result, ShouldBeFalse)
		So(err.Error(), ShouldContainSubstring, "mocked error")
	})

	Convey("no insolvency resource found", t, func() {
		// setup
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// create mock response
		mockService.EXPECT().GetInsolvencyResource("transactionID").Return(nil, nil)

		result, err := checkIDsMatch("transactionID", "practitionerID", mockService)

		So(result, ShouldBeFalse)
		So(err, ShouldBeNil)
	})

	Convey("insolvency resource has no practitioners", t, func() {
		// setup
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// create mock response
		mockService.EXPECT().GetInsolvencyResource("transactionID").Return(&insolvencyDao, nil)
		result, err := checkIDsMatch("transactionID", "practitionerID", mockService)

		So(result, ShouldBeFalse)
		So(err, ShouldBeNil)
	})

	Convey("insolvency resource has practitioners but not the correct one", t, func() {
		// setup
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// create mock response
		insolvencyDao.Data.Practitioners = &models.InsolvencyResourcePractitionersDao{
			"wrongPracID":  "wrongPracLink",
			"wrongPracID2": "wrongPracLink2",
		}

		mockService.EXPECT().GetInsolvencyResource("transactionID").Return(&insolvencyDao, nil)

		result, err := checkIDsMatch("transactionID", "practitionerID", mockService)

		So(result, ShouldBeFalse)
		So(err, ShouldBeNil)
	})

	Convey("success - practitioner id matched in insolvency resource for transaction id", t, func() {
		// setup
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// create mock response
		insolvencyDao.Data.Practitioners = &models.InsolvencyResourcePractitionersDao{
			"wrongPracID":    "wrongPracLink",
			"practitionerID": "practitionerLink",
		}
		mockService.EXPECT().GetInsolvencyResource("transactionID").Return(&insolvencyDao, nil)

		result, err := checkIDsMatch("transactionID", "practitionerID", mockService)

		So(result, ShouldBeTrue)
		So(err, ShouldBeNil)
	})

}
