package service

import (
	"errors"
	"testing"

	"github.com/companieshouse/insolvency-api/mocks"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/golang/mock/gomock"
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var transactionID = "12345678"
var companyNumber = "01234567"
var companyName = "companyName"

func createInsolvencyResource() models.InsolvencyResourceDao {
	return models.InsolvencyResourceDao{
		ID:            primitive.ObjectID{},
		TransactionID: transactionID,
		Etag:          "etag1234",
		Kind:          "insolvency",
		Data: models.InsolvencyResourceDaoData{
			CompanyNumber: companyNumber,
			CompanyName:   companyName,
			CaseType:      "insolvency",
		},
		Links: models.InsolvencyResourceLinksDao{
			Self:             "/transactions/123456789/insolvency",
			ValidationStatus: "/transactions/123456789/insolvency/validation-status",
		},
	}
}

func TestValidateInsolvencyDetails(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	Convey("error getting insolvency resource from database", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect GetInsolvencyResource to be called once and return an error for the insolvency case
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(createInsolvencyResource(), errors.New("insolvency case does not exist")).Times(1)

		isValid, validationErrors := ValidateInsolvencyDetails(mockService, transactionID)

		So(isValid, ShouldBeFalse)
		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, "error getting insolvency resource from DB: [insolvency case does not exist]")
		So((*validationErrors)[0].Location, ShouldContainSubstring, "insolvency case")
	})

	Convey("successfully returned valid insolvency resource", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(createInsolvencyResource(), nil).Times(1)

		isValid, validationErrors := ValidateInsolvencyDetails(mockService, transactionID)

		So(isValid, ShouldBeTrue)
		So(validationErrors, ShouldHaveLength, 0)
	})
}
