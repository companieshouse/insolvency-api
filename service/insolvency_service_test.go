package service

import (
	"errors"
	"fmt"
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
			Practitioners: []models.PractitionerResourceDao{
				{
					ID:              "1234",
					IPCode:          "1234",
					FirstName:       "Name",
					LastName:        "LastName",
					TelephoneNumber: "1234",
					Email:           "name@email.com",
					Address:         models.AddressResourceDao{},
					Role:            "final-liquidator",
					Links:           models.PractitionerResourceLinksDao{},
					Appointment: &models.AppointmentResourceDao{
						AppointedOn: "2020-01-01",
						MadeBy:      "creditors",
					},
				},
				{
					ID:              "5678",
					IPCode:          "5678",
					FirstName:       "FirstName",
					LastName:        "LastName",
					TelephoneNumber: "5678",
					Email:           "firstname@email.com",
					Address:         models.AddressResourceDao{},
					Role:            "final-liquidator",
					Links:           models.PractitionerResourceLinksDao{},
					Appointment: &models.AppointmentResourceDao{
						AppointedOn: "2020-01-01",
						MadeBy:      "creditors",
					},
				},
			},
			Attachments: []models.AttachmentResourceDao{
				{
					ID:     "id",
					Type:   "type1",
					Status: "status",
					Links: models.AttachmentResourceLinksDao{
						Self:     "self",
						Download: "download",
					},
				},
				{
					ID:     "id",
					Type:   "type2",
					Status: "status",
					Links: models.AttachmentResourceLinksDao{
						Self:     "self",
						Download: "download",
					},
				},
			},
		},
		Links: models.InsolvencyResourceLinksDao{
			Self:             "/transactions/123456789/insolvency",
			ValidationStatus: "/transactions/123456789/insolvency/validation-status",
		},
	}
}

func TestUnitValidateInsolvencyDetails(t *testing.T) {
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

	Convey("error - one practitioner is appointed but not all practitioners have been appointed", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		insolvencyCase := createInsolvencyResource()
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyCase, nil).Times(1)

		// Remove appointment for one practitioner
		insolvencyCase.Data.Practitioners[1].Appointment = nil

		isValid, validationErrors := ValidateInsolvencyDetails(mockService, transactionID)

		So(isValid, ShouldBeFalse)
		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - all practitioners for insolvency case with transaction id [%s] must be appointed", insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "appointment")
	})

	Convey("error - one practitioner is appointed but not all practitioners have been appointed - missing date", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		insolvencyCase := createInsolvencyResource()
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyCase, nil).Times(1)

		// Remove appointment for one practitioner
		insolvencyCase.Data.Practitioners[1].Appointment.AppointedOn = ""

		isValid, validationErrors := ValidateInsolvencyDetails(mockService, transactionID)

		So(isValid, ShouldBeFalse)
		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - all practitioners for insolvency case with transaction id [%s] must be appointed", insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "appointment")
	})

	Convey("successful validation of practitioner appointments - all practitioners appointed", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		insolvencyCase := createInsolvencyResource()
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyCase, nil).Times(1)

		isValid, validationErrors := ValidateInsolvencyDetails(mockService, transactionID)

		So(isValid, ShouldBeTrue)
		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("successful validation of practitioner appointments - no practitioners are appointed", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		insolvencyCase := createInsolvencyResource()
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyCase, nil).Times(1)

		// Remove appointment details for all practitioners
		insolvencyCase.Data.Practitioners[0].Appointment = nil
		insolvencyCase.Data.Practitioners[1].Appointment = nil

		isValid, validationErrors := ValidateInsolvencyDetails(mockService, transactionID)

		So(isValid, ShouldBeTrue)
		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("error - attachment type is not resolution and practitioners key is absent", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		insolvencyCase := models.InsolvencyResourceDao{
			Data: models.InsolvencyResourceDaoData{
				Attachments: []models.AttachmentResourceDao{
					{
						Type: "type",
					},
				},
			},
		}
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyCase, nil).Times(1)

		isValid, validationErrors := ValidateInsolvencyDetails(mockService, transactionID)

		So(isValid, ShouldBeFalse)
		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - attachment type requires that at least one practitioner must be present for insolvency case with transaction id [%s]", insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "attachment type")
	})

	Convey("error - attachment type is not resolution and practitioners object is empty", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		insolvencyCase := models.InsolvencyResourceDao{
			Data: models.InsolvencyResourceDaoData{
				Practitioners: []models.PractitionerResourceDao{},
				Attachments: []models.AttachmentResourceDao{
					{
						Type: "type",
					},
				},
			},
		}
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyCase, nil).Times(1)

		isValid, validationErrors := ValidateInsolvencyDetails(mockService, transactionID)

		So(isValid, ShouldBeFalse)
		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - attachment type requires that at least one practitioner must be present for insolvency case with transaction id [%s]", insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "attachment type")
	})

	Convey("successful validation of attachment type - attachment type is not resolution and practitioner present", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		insolvencyCase := createInsolvencyResource()
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyCase, nil).Times(1)

		isValid, validationErrors := ValidateInsolvencyDetails(mockService, transactionID)

		So(isValid, ShouldBeTrue)
		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("successful validation of resolution attachment - attachment type is resolution and practitioner present", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		insolvencyCase := createInsolvencyResource()
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyCase, nil).Times(1)

		// Set attachment type to "resolution"
		insolvencyCase.Data.Attachments[0].Type = "resolution"

		isValid, validationErrors := ValidateInsolvencyDetails(mockService, transactionID)

		So(isValid, ShouldBeTrue)
		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("successful validation of resolution attachment - attachment type is resolution and practitioners key is absent", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		insolvencyCase := models.InsolvencyResourceDao{
			Data: models.InsolvencyResourceDaoData{
				Attachments: []models.AttachmentResourceDao{
					{
						Type: "test",
					},
				},
			},
		}
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyCase, nil).Times(1)

		// Set attachment type to "resolution"
		insolvencyCase.Data.Attachments[0].Type = "resolution"

		isValid, validationErrors := ValidateInsolvencyDetails(mockService, transactionID)

		So(isValid, ShouldBeTrue)
		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("successful validation of resolution attachment - attachment type is resolution and practitioners object empty", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		insolvencyCase := models.InsolvencyResourceDao{

			Data: models.InsolvencyResourceDaoData{
				Practitioners: []models.PractitionerResourceDao{},
				Attachments: []models.AttachmentResourceDao{
					{
						Type: "test",
					},
				},
			},
		}
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyCase, nil).Times(1)

		// Set attachment type to "resolution"
		insolvencyCase.Data.Attachments[0].Type = "resolution"

		isValid, validationErrors := ValidateInsolvencyDetails(mockService, transactionID)

		So(isValid, ShouldBeTrue)
		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("error - attachment type is statement-of-concurrence and attachment type statement-of-affairs-director is not present", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		insolvencyCase := createInsolvencyResource()
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyCase, nil).Times(1)

		// Set attachment type to "statement-of-concurrence"
		insolvencyCase.Data.Attachments[0].Type = "statement-of-concurrence"

		isValid, validationErrors := ValidateInsolvencyDetails(mockService, transactionID)

		So(isValid, ShouldBeFalse)
		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - attachment statement-of-concurrence must be accompanied by statement-of-affairs-director for insolvency case with transaction id [%s]", insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "statement of concurrence attachment type")
	})

	Convey("successful validation of statement-of-concurrence attachment - attachment type is statement-of-concurrence and statement-of-affairs-director are present", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		insolvencyCase := createInsolvencyResource()
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyCase, nil).Times(1)

		// Set attachment type to "statement-of-concurrence"
		insolvencyCase.Data.Attachments[0].Type = "statement-of-concurrence"
		insolvencyCase.Data.Attachments[1].Type = "statement-of-affairs-director"

		isValid, validationErrors := ValidateInsolvencyDetails(mockService, transactionID)

		So(isValid, ShouldBeTrue)
		So(validationErrors, ShouldHaveLength, 0)
	})
}
