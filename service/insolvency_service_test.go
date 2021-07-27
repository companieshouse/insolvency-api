package service

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/companieshouse/insolvency-api/constants"
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
var req = httptest.NewRequest(http.MethodPut, "/test", nil)

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
						AppointedOn: "2021-07-07",
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
						AppointedOn: "2021-07-07",
						MadeBy:      "creditors",
					},
				},
			},
			Attachments: []models.AttachmentResourceDao{
				{
					ID:     "id",
					Type:   "resolution",
					Status: "status",
					Links: models.AttachmentResourceLinksDao{
						Self:     "self",
						Download: "download",
					},
				},
				{
					ID:     "id",
					Type:   "statement-of-affairs-director",
					Status: "status",
					Links: models.AttachmentResourceLinksDao{
						Self:     "self",
						Download: "download",
					},
				},
			},
			Resolution: &models.ResolutionResourceDao{
				DateOfResolution: "2021-06-06",
				Attachments: []string{
					"id",
				},
			},
			StatementOfAffairs: &models.StatementOfAffairsResourceDao{
				StatementDate: "2021-06-06",
				Attachments: []string{
					"id",
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

	Convey("error - one practitioner is appointed but not all practitioners have been appointed", t, func() {
		insolvencyCase := createInsolvencyResource()

		// Remove appointment for one practitioner
		insolvencyCase.Data.Practitioners[1].Appointment = nil

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)

		So(isValid, ShouldBeFalse)
		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - all practitioners for insolvency case with transaction id [%s] must be appointed", insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "appointment")
	})

	Convey("error - one practitioner is appointed but not all practitioners have been appointed - missing date", t, func() {
		insolvencyCase := createInsolvencyResource()

		// Remove appointment for one practitioner
		insolvencyCase.Data.Practitioners[1].Appointment.AppointedOn = ""

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)

		So(isValid, ShouldBeFalse)
		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - all practitioners for insolvency case with transaction id [%s] must be appointed", insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "appointment")
	})

	Convey("successful validation of practitioner appointments - all practitioners appointed", t, func() {
		isValid, validationErrors := ValidateInsolvencyDetails(createInsolvencyResource())

		So(isValid, ShouldBeTrue)
		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("successful validation of practitioner appointments - no practitioners are appointed", t, func() {
		insolvencyCase := createInsolvencyResource()

		// Remove appointment details for all practitioners
		insolvencyCase.Data.Practitioners[0].Appointment = nil
		insolvencyCase.Data.Practitioners[1].Appointment = nil

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)

		So(isValid, ShouldBeTrue)
		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("error - attachment type is not resolution and practitioners key is absent", t, func() {
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

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)

		So(isValid, ShouldBeFalse)
		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - attachment type requires that at least one practitioner must be present for insolvency case with transaction id [%s]", insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "resolution attachment type")
	})

	Convey("error - attachment type is not resolution and practitioners object is empty", t, func() {
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

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)

		So(isValid, ShouldBeFalse)
		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - attachment type requires that at least one practitioner must be present for insolvency case with transaction id [%s]", insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "resolution attachment type")
	})

	Convey("successful validation of attachment type - attachment type is not resolution and practitioner present", t, func() {
		insolvencyCase := createInsolvencyResource()

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)

		So(isValid, ShouldBeTrue)
		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("successful validation of resolution attachment - attachment type is resolution and practitioner present", t, func() {
		insolvencyCase := createInsolvencyResource()
		// Set attachment type to "resolution"
		insolvencyCase.Data.Attachments[0] = models.AttachmentResourceDao{
			Type: "resolution",
			ID:   "1234",
		}

		insolvencyCase.Data.Resolution = &models.ResolutionResourceDao{
			DateOfResolution: "2021-06-06",
			Attachments: []string{
				"1234",
			},
		}

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)

		So(isValid, ShouldBeTrue)
		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("successful validation of resolution attachment - attachment type is resolution and practitioners key is absent", t, func() {
		insolvencyCase := models.InsolvencyResourceDao{
			Data: models.InsolvencyResourceDaoData{
				Attachments: []models.AttachmentResourceDao{
					{
						Type: "resolution",
						ID:   "1234",
					},
				},
				Resolution: &models.ResolutionResourceDao{
					DateOfResolution: "2021-06-06",
					Attachments: []string{
						"1234",
					},
				},
			},
		}

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)

		So(isValid, ShouldBeTrue)
		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("successful validation of resolution attachment - attachment type is resolution and practitioners object empty", t, func() {
		insolvencyCase := models.InsolvencyResourceDao{
			Data: models.InsolvencyResourceDaoData{
				Attachments: []models.AttachmentResourceDao{
					{
						Type: "resolution",
						ID:   "1234",
					},
				},
				Resolution: &models.ResolutionResourceDao{
					DateOfResolution: "2021-06-06",
					Attachments: []string{
						"1234",
					},
				},
			},
		}

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)

		So(isValid, ShouldBeTrue)
		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("error - attachment type is statement-of-concurrence and attachment type statement-of-affairs-director is not present", t, func() {
		insolvencyCase := createInsolvencyResource()
		// Set attachment type to "statement-of-concurrence"
		insolvencyCase.Data.Attachments = append(insolvencyCase.Data.Attachments, models.AttachmentResourceDao{Type: "statement-of-concurrence"})
		// Remove SOA director
		insolvencyCase.Data.Attachments[1].Type = "type"
		// Remove SOA
		insolvencyCase.Data.StatementOfAffairs = nil

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)

		So(isValid, ShouldBeFalse)
		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - attachment statement-of-concurrence must be accompanied by statement-of-affairs-director attachment for insolvency case with transaction id [%s]", insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "statement of concurrence attachment type")
	})

	Convey("successful validation of statement-of-concurrence attachment - attachment type is statement-of-concurrence and statement-of-affairs-director are present", t, func() {
		insolvencyCase := createInsolvencyResource()
		insolvencyCase.Data.Resolution = nil

		// Set attachment type to "statement-of-concurrence"
		insolvencyCase.Data.Attachments[0].Type = "statement-of-concurrence"
		insolvencyCase.Data.Attachments[1].Type = "statement-of-affairs-director"

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)

		So(isValid, ShouldBeTrue)
		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("error - attachment type is statement-of-affairs-liquidator and a practitioner is appointed", t, func() {
		insolvencyCase := createInsolvencyResource()

		// Set attachment type to "statement-of-concurrence"
		insolvencyCase.Data.Attachments[0].Type = "statement-of-affairs-liquidator"

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)

		So(isValid, ShouldBeFalse)
		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - no appointed practitioners can be assigned to the case when attachment type statement-of-affairs-liquidator is included with transaction id [%s]", insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "statement of affairs liquidator attachment type")
	})

	Convey("successful validation of statement-of-affairs-liquidator - attachment type is statement-of-affairs-liquidator and at least one practitioner is present but not appointed", t, func() {
		insolvencyCase := createInsolvencyResource()

		// Set attachment type to "statement-of-affairs-liquidator"
		insolvencyCase.Data.Attachments[0].Type = "statement-of-affairs-liquidator"

		// Remove resolution from insolvency case
		insolvencyCase.Data.Resolution = nil

		// Remove appointment details for all practitioners
		insolvencyCase.Data.Practitioners[0].Appointment = nil
		insolvencyCase.Data.Practitioners[1].Appointment = nil

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)

		So(isValid, ShouldBeTrue)
		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("error - no attachments present and no appointed practitioners on insolvency case", t, func() {
		insolvencyCase := models.InsolvencyResourceDao{
			Data: models.InsolvencyResourceDaoData{},
		}

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)

		So(isValid, ShouldBeFalse)
		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - at least one practitioner must be appointed as there are no attachments for insolvency case with transaction id [%s]", insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "no attachments")
	})

	Convey("successful validation - no attachments present but at least one appointed practitioner is present on insolvency case", t, func() {
		insolvencyCase := models.InsolvencyResourceDao{
			Data: models.InsolvencyResourceDaoData{
				Practitioners: []models.PractitionerResourceDao{
					{
						Appointment: &models.AppointmentResourceDao{
							AppointedOn: "2020-01-01",
							MadeBy:      "creditors",
						},
					},
				},
			},
		}

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)

		So(isValid, ShouldBeTrue)
		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("error - resolution attachment present and no date of resolution filed for insolvency case", t, func() {
		insolvencyCase := createInsolvencyResource()
		insolvencyCase.Data.Attachments[0].Type = "resolution"
		insolvencyCase.Data.Resolution = &models.ResolutionResourceDao{
			DateOfResolution: "",
		}

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)

		So(isValid, ShouldBeFalse)
		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - a date of resolution must be present as there is an attachment with type resolution for insolvency case with transaction id [%s]", insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "no date of resolution")
	})

	Convey("error - date_of_resolution present and no resolution filed for insolvency case", t, func() {
		insolvencyCase := createInsolvencyResource()
		insolvencyCase.Data.Attachments[0].Type = "test"

		insolvencyCase.Data.Resolution = &models.ResolutionResourceDao{
			DateOfResolution: "2021-06-06",
		}

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)

		So(isValid, ShouldBeFalse)
		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - a resolution attachment must be present as there is a date_of_resolution filed for insolvency case with transaction id [%s]", insolvencyCase.TransactionID))

		So((*validationErrors)[0].Location, ShouldContainSubstring, "no resolution")
	})

	Convey("error - id for uploaded resolution attachment does not match id supplied with resolution filed for insolvency case", t, func() {
		insolvencyCase := createInsolvencyResource()
		insolvencyCase.Data.Attachments[0] = models.AttachmentResourceDao{
			Type: "resolution",
			ID:   "1234",
		}
		insolvencyCase.Data.Resolution = &models.ResolutionResourceDao{
			DateOfResolution: "2021-06-06",
			Attachments: []string{
				"0234",
			},
		}

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)

		So(isValid, ShouldBeFalse)
		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - id for uploaded resolution attachment must match the attachment id supplied when filing a resolution for insolvency case with transaction id [%s]", insolvencyCase.TransactionID))

		So((*validationErrors)[0].Location, ShouldContainSubstring, "attachment ids do not match")
	})

	Convey("successful validation - resolution attachment present and date of resolution filed for insolvency case", t, func() {
		insolvencyCase := createInsolvencyResource()
		insolvencyCase.Data.Attachments[0] = models.AttachmentResourceDao{
			Type: "resolution",
			ID:   "1234",
		}
		insolvencyCase.Data.Resolution = &models.ResolutionResourceDao{
			DateOfResolution: "2021-06-06",
			Attachments: []string{
				"1234",
			},
		}
		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)

		So(isValid, ShouldBeTrue)
		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("error - statement-of-affairs-director filed but no statement date exists in DB", t, func() {

		// Create insolvency case and remove SOA date
		insolvencyCase := createInsolvencyResource()
		insolvencyCase.Data.StatementOfAffairs = &models.StatementOfAffairsResourceDao{
			StatementDate: "",
		}

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)

		So(isValid, ShouldBeFalse)
		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - a date of statement of affairs must be present as there is an attachment with type [%s] for insolvency case with transaction id [%s]", constants.StatementOfAffairsDirector.String(), insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "statement-of-affairs")
	})

	Convey("error - practitioner appointment is before date of resolution", t, func() {
		// Add resolution to insolvency case
		insolvencyCase := createInsolvencyResource()
		insolvencyCase.Data.Attachments[0] = models.AttachmentResourceDao{
			Type: "resolution",
			ID:   "1234",
		}
		insolvencyCase.Data.Resolution = &models.ResolutionResourceDao{
			DateOfResolution: "2021-06-06",
			Attachments: []string{
				"1234",
			},
		}

		// Appoint practitioner before resolution
		insolvencyCase.Data.Practitioners[0].Appointment.AppointedOn = "2021-05-05"

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)
		So(isValid, ShouldBeFalse)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - practitioner [%s] appointed on [%s] is before the resolution date [%s]", insolvencyCase.Data.Practitioners[0].ID, insolvencyCase.Data.Practitioners[0].Appointment.AppointedOn, insolvencyCase.Data.Resolution.DateOfResolution))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "practitioner")
	})

	Convey("error parsing appointment date", t, func() {
		// Add resolution to insolvency case
		insolvencyCase := createInsolvencyResource()
		insolvencyCase.Data.Attachments[0] = models.AttachmentResourceDao{
			Type: "resolution",
			ID:   "1234",
		}
		insolvencyCase.Data.Resolution = &models.ResolutionResourceDao{
			DateOfResolution: "2021-06-06",
			Attachments: []string{
				"1234",
			},
		}

		// Appoint practitioner before resolution
		insolvencyCase.Data.Practitioners[0].Appointment.AppointedOn = "date"

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)
		So(isValid, ShouldBeFalse)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("cannot parse"))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "practitioner")
	})

	Convey("error parsing resolution date", t, func() {
		// Add resolution to insolvency case
		insolvencyCase := createInsolvencyResource()
		insolvencyCase.Data.Attachments[0] = models.AttachmentResourceDao{
			Type: "resolution",
			ID:   "1234",
		}
		insolvencyCase.Data.Resolution = &models.ResolutionResourceDao{
			DateOfResolution: "date",
			Attachments: []string{
				"1234",
			},
		}
		// Appoint practitioner before resolution
		insolvencyCase.Data.Practitioners[0].Appointment.AppointedOn = "2021-05-05"

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)
		So(isValid, ShouldBeFalse)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("cannot parse"))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "practitioner")
	})

	Convey("valid insolvency case - appointment date is after resolution date", t, func() {
		// Add resolution to insolvency case
		insolvencyCase := createInsolvencyResource()
		insolvencyCase.Data.Attachments[0] = models.AttachmentResourceDao{
			Type: "resolution",
			ID:   "1234",
		}
		insolvencyCase.Data.Resolution = &models.ResolutionResourceDao{
			DateOfResolution: "2021-06-06",
			Attachments: []string{
				"1234",
			},
		}

		isValid, validationErrors := ValidateInsolvencyDetails(insolvencyCase)
		So(isValid, ShouldBeTrue)
		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("error - antivirus check has not been completed", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		insolvencyCase := createInsolvencyResource()

		attachment := `{
			"name": "file",
			"size": 1000,
			"content_type": "test",
			"av_status": "not-scanned"
			}`

		// Expect GetAttachmentDetails to be called once and return the attachment
		httpmock.RegisterResponder(http.MethodGet, `=~.*`, httpmock.NewStringResponder(http.StatusOK, attachment))

		mockService.EXPECT().UpdateAttachmentStatus(transactionID, insolvencyCase.Data.Attachments[0].ID, "integrity_failed").Return(http.StatusNoContent, nil).Times(2)

		isValid, validationErrors := ValidateAntivirus(mockService, insolvencyCase, req)

		So(isValid, ShouldBeFalse)
		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - antivirus check has failed on insolvency case with transaction id [%s], attachments have not been scanned", insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "antivirus incomplete")
	})

	Convey("error - antivirus check has failed, attachment is infected", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		insolvencyCase := createInsolvencyResource()

		attachment := `{
			"name": "file",
			"size": 1000,
			"content_type": "test",
			"av_status": "infected"
			}`

		// Expect GetAttachmentDetails to be called once and return the attachment
		httpmock.RegisterResponder(http.MethodGet, `=~.*`, httpmock.NewStringResponder(http.StatusOK, attachment))

		mockService.EXPECT().UpdateAttachmentStatus(transactionID, insolvencyCase.Data.Attachments[0].ID, "integrity_failed").Return(http.StatusNoContent, nil).Times(2)

		isValid, validationErrors := ValidateAntivirus(mockService, insolvencyCase, req)

		So(isValid, ShouldBeFalse)
		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - antivirus check has failed on insolvency case with transaction id [%s], virus detected", insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "antivirus failure")
	})

	Convey("successful validation - antivirus check has passed, attachment is clean", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		insolvencyCase := createInsolvencyResource()

		attachment := `{
			"name": "file",
			"size": 1000,
			"content_type": "test",
			"av_status": "clean"
			}`

		// Expect GetAttachmentDetails to be called once and return the attachment
		httpmock.RegisterResponder(http.MethodGet, `=~.*`, httpmock.NewStringResponder(http.StatusOK, attachment))

		mockService.EXPECT().UpdateAttachmentStatus(transactionID, insolvencyCase.Data.Attachments[0].ID, "processed").Return(http.StatusNoContent, nil).Times(2)

		isValid, validationErrors := ValidateAntivirus(mockService, insolvencyCase, req)

		So(isValid, ShouldBeTrue)
		So(validationErrors, ShouldHaveLength, 0)
	})
}

var transactionProfileResponseClosed = `
{
 "status": "closed"
}
`

func TestUnitGenerateFilings(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	Convey("error getting insolvency resource from database", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect GetInsolvencyResource to be called once and return an error for the insolvency case
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(createInsolvencyResource(), errors.New("insolvency case does not exist")).Times(1)

		filings, err := GenerateFilings(mockService, transactionID)

		So(filings, ShouldBeNil)
		So(err.Error(), ShouldContainSubstring, "insolvency case does not exist")
	})

	Convey("Generate filing for 600 case with two practitioners", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect the transaction api to be called and return a closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		insolvencyResource := createInsolvencyResource()
		insolvencyResource.Data.Attachments = []models.AttachmentResourceDao{}

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyResource, nil).Times(1)

		filings, err := GenerateFilings(mockService, transactionID)

		So(filings[0].Kind, ShouldEqual, "insolvency#600")
		So(filings[0].DescriptionIdentifier, ShouldEqual, "600")
		So(err, ShouldBeNil)
	})

	Convey("Generate filing for LRESEX case with resolution attachment and no practitioners", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect the transaction api to be called and return a closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		insolvencyResource := createInsolvencyResource()
		insolvencyResource.Data.Practitioners = []models.PractitionerResourceDao{}
		insolvencyResource.Data.Attachments = []models.AttachmentResourceDao{
			{
				ID:     "id",
				Type:   "resolution",
				Status: "status",
				Links: models.AttachmentResourceLinksDao{
					Self:     "self",
					Download: "download",
				},
			},
		}

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyResource, nil).Times(1)

		filings, err := GenerateFilings(mockService, transactionID)

		So(filings[0].Kind, ShouldEqual, "insolvency#LRESEX")
		So(filings[0].DescriptionIdentifier, ShouldEqual, "LRESEX")
		So(err, ShouldBeNil)
	})

	Convey("Generate filing for LIQ02 case with statement-of-affairs-director attachment and two practitioners", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect the transaction api to be called and return a closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		insolvencyResource := createInsolvencyResource()
		insolvencyResource.Data.Practitioners[0].Appointment = nil
		insolvencyResource.Data.Practitioners[1].Appointment = nil
		insolvencyResource.Data.Attachments = []models.AttachmentResourceDao{
			{
				ID:     "id",
				Type:   "statement-of-affairs-director",
				Status: "status",
				Links: models.AttachmentResourceLinksDao{
					Self:     "self",
					Download: "download",
				},
			},
		}

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyResource, nil).Times(1)

		filings, err := GenerateFilings(mockService, transactionID)

		So(filings[0].Kind, ShouldEqual, "insolvency#LIQ02")
		So(filings[0].DescriptionIdentifier, ShouldEqual, "LIQ02")
		So(err, ShouldBeNil)
	})

	Convey("Generate filing for LIQ02 case with statement-of-affairs-liquidator attachment and two practitioners", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect the transaction api to be called and return a closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		insolvencyResource := createInsolvencyResource()
		insolvencyResource.Data.Practitioners[0].Appointment = nil
		insolvencyResource.Data.Practitioners[1].Appointment = nil
		insolvencyResource.Data.Attachments = []models.AttachmentResourceDao{
			{
				ID:     "id",
				Type:   "statement-of-affairs-liquidator",
				Status: "status",
				Links: models.AttachmentResourceLinksDao{
					Self:     "self",
					Download: "download",
				},
			},
		}

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyResource, nil).Times(1)

		filings, err := GenerateFilings(mockService, transactionID)

		So(filings[0].Kind, ShouldEqual, "insolvency#LIQ02")
		So(filings[0].DescriptionIdentifier, ShouldEqual, "LIQ02")
		So(err, ShouldBeNil)
	})

	Convey("Generate filing for LIQ02 case with statement-of-affairs-liquidator and statement-of-affairs-director attachments and two practitioners", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect the transaction api to be called and return a closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		insolvencyResource := createInsolvencyResource()
		insolvencyResource.Data.Practitioners[0].Appointment = nil
		insolvencyResource.Data.Practitioners[1].Appointment = nil
		insolvencyResource.Data.Attachments = []models.AttachmentResourceDao{
			{
				ID:     "id",
				Type:   "statement-of-affairs-liquidator",
				Status: "status",
				Links: models.AttachmentResourceLinksDao{
					Self:     "self",
					Download: "download",
				},
			},
			{
				ID:     "id",
				Type:   "statement-of-affairs-director",
				Status: "status",
				Links: models.AttachmentResourceLinksDao{
					Self:     "self",
					Download: "download",
				},
			},
		}

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyResource, nil).Times(1)

		filings, err := GenerateFilings(mockService, transactionID)

		So(filings[0].Kind, ShouldEqual, "insolvency#LIQ02")
		So(filings[0].DescriptionIdentifier, ShouldEqual, "LIQ02")
		So(err, ShouldBeNil)
	})

	Convey("Generate filing for 600 and LIQ02 case with statement-of-affairs-director attachment and two practitioners", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect the transaction api to be called and return a closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		insolvencyResource := createInsolvencyResource()
		insolvencyResource.Data.Attachments = []models.AttachmentResourceDao{
			{
				ID:     "id",
				Type:   "statement-of-affairs-director",
				Status: "status",
				Links: models.AttachmentResourceLinksDao{
					Self:     "self",
					Download: "download",
				},
			},
		}

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyResource, nil).Times(1)

		filings, err := GenerateFilings(mockService, transactionID)

		So(filings[0].Kind, ShouldEqual, "insolvency#600")
		So(filings[0].DescriptionIdentifier, ShouldEqual, "600")
		So(filings[1].Kind, ShouldEqual, "insolvency#LIQ02")
		So(filings[1].DescriptionIdentifier, ShouldEqual, "LIQ02")
		So(err, ShouldBeNil)
	})

	Convey("Generate filing for 600, LRESEX, and LIQ02 case with statement-of-affairs-director and statement-of-affairs-liquidator attachments and two practitioners", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect the transaction api to be called and return a closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		insolvencyResource := createInsolvencyResource()
		insolvencyResource.Data.Attachments = []models.AttachmentResourceDao{
			{
				ID:     "id",
				Type:   "resolution",
				Status: "status",
				Links: models.AttachmentResourceLinksDao{
					Self:     "self",
					Download: "download",
				},
			},
			{
				ID:     "id",
				Type:   "statement-of-affairs-director",
				Status: "status",
				Links: models.AttachmentResourceLinksDao{
					Self:     "self",
					Download: "download",
				},
			},
			{
				ID:     "id",
				Type:   "statement-of-affairs-liquidator",
				Status: "status",
				Links: models.AttachmentResourceLinksDao{
					Self:     "self",
					Download: "download",
				},
			},
		}

		// Expect GetInsolvencyResource to be called once and return a valid insolvency case
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(insolvencyResource, nil).Times(1)

		filings, err := GenerateFilings(mockService, transactionID)

		So(filings[0].Kind, ShouldEqual, "insolvency#600")
		So(filings[0].DescriptionIdentifier, ShouldEqual, "600")
		So(filings[1].Kind, ShouldEqual, "insolvency#LRESEX")
		So(filings[1].DescriptionIdentifier, ShouldEqual, "LRESEX")
		So(filings[2].Kind, ShouldEqual, "insolvency#LIQ02")
		So(filings[2].DescriptionIdentifier, ShouldEqual, "LIQ02")
		So(err, ShouldBeNil)
	})
}
