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

func createInsolvencyResource() *models.InsolvencyResourceDao {
	_, practitionerResourceDao, appointmentResourceDao := generateInsolvencyPractitionerAppointmentResources()

	jsonPractitionersDao := `{
		"VM04221441": "/transactions/168570-809316-704268/insolvency/practitioners/VM04221441",
		"VM04221442": "/transactions/168570-809316-704268/insolvency/practitioners/VM04221442"
	}`

	appointmentResourceDao.Data.AppointedOn = "2021-07-07"
	appointmentResourceDao.Data.MadeBy = "creditors"

	practitionerResourceDao.Data.Appointment = &appointmentResourceDao

	practitionerResourceDao1 := models.PractitionerResourceDao{}
	practitionerResourceDao1.Data.IPCode = "5678"
	practitionerResourceDao1.Data.FirstName = "FirstName"
	practitionerResourceDao1.Data.LastName = "LastName"
	practitionerResourceDao1.Data.TelephoneNumber = "5678"
	practitionerResourceDao1.Data.Email = "firstname@email.com"
	practitionerResourceDao1.Data.Address = models.AddressResourceDao{}
	practitionerResourceDao1.Data.Role = "final-liquidator"
	practitionerResourceDao1.Data.Links = models.PractitionerResourceLinksDao{}

	appointmentResourceDao.Data.AppointedOn = "2021-07-07"
	appointmentResourceDao.Data.MadeBy = "creditors"

	practitionerResourceDao1.Data.Appointment = &appointmentResourceDao

	insolvencyResourceDao := models.InsolvencyResourceDao{
		ID:            primitive.ObjectID{},
		TransactionID: transactionID,
	}
	insolvencyResourceDao.Data.Etag = "etag1234"
	insolvencyResourceDao.Data.Kind = "insolvency"
	insolvencyResourceDao.Data.CompanyNumber = companyNumber
	insolvencyResourceDao.Data.CompanyName = companyName
	insolvencyResourceDao.Data.CaseType = "insolvency"
	insolvencyResourceDao.Data.Practitioners = jsonPractitionersDao
	insolvencyResourceDao.Data.Attachments = []models.AttachmentResourceDao{
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
			Type:   "progress-report",
			Status: "status",
			Links: models.AttachmentResourceLinksDao{
				Self:     "self",
				Download: "download",
			},
		},
	}
	insolvencyResourceDao.Data.Resolution = &models.ResolutionResourceDao{
		DateOfResolution: "2021-06-06",
		Attachments: []string{
			"id",
		},
	}
	insolvencyResourceDao.Data.StatementOfAffairs = &models.StatementOfAffairsResourceDao{
		StatementDate: "2021-06-06",
		Attachments: []string{
			"id",
		},
	}
	insolvencyResourceDao.Data.ProgressReport = &models.ProgressReportResourceDao{
		FromDate: "2021-04-14",
		ToDate:   "2022-04-13",
		Attachments: []string{
			"id",
		},
	}
	insolvencyResourceDao.Data.Links = models.InsolvencyResourceLinksDao{
		Self:             "/transactions/123456789/insolvency",
		ValidationStatus: "/transactions/123456789/insolvency/validation-status",
	}

	return &insolvencyResourceDao

}

func TestUnitValidateInsolvencyDetails(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	Convey("error - one practitioner is appointed but not all practitioners have been appointed", t, func() {
		insolvencyCase := createInsolvencyResource()

		// Remove appointment for one practitioner
		practitionerResourceDao := models.PractitionerResourceDao{}
		practitionerResourceDao.Data.Appointment = &models.AppointmentResourceDao{}
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)

		So(validationErrors, ShouldHaveLength, 3)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - all practitioners for insolvency case with transaction id [%s] must be appointed", insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "appointment")
	})

	Convey("error - one practitioner is appointed but not all practitioners have been appointed - missing date", t, func() {
		insolvencyCase := createInsolvencyResource()

		appointmentResourceDao := models.AppointmentResourceDao{}

		appointmentResourceDao.Data.AppointedOn = "2021-07-07"
		practitionerResourceDao := models.PractitionerResourceDao{}
		practitionerResourceDao.Data.Appointment = &appointmentResourceDao

		practitionerResourceDao1 := models.PractitionerResourceDao{}
		appointmentResourceDao1 := models.AppointmentResourceDao{}
		practitionerResourceDao1.Data.Appointment = &appointmentResourceDao1

		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao, practitionerResourceDao1)

		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)

		So(validationErrors, ShouldHaveLength, 3)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - all practitioners for insolvency case with transaction id [%s] must be appointed", insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "appointment")
	})

	Convey("successful validation of practitioner appointments - all practitioners appointed", t, func() {

		practitionerResourceDao := models.PractitionerResourceDao{}
		appointmentResourceDao := models.AppointmentResourceDao{}
		appointmentResourceDao.Data.AppointedOn = "2012-01-23"
		appointmentResourceDao.Data.MadeBy = "MadeBy"
		appointmentResourceDao.Data.Links = models.AppointmentResourceLinksDao{}
		appointmentResourceDao.PractitionerId = "PractitionerID"
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		validationErrors := ValidateInsolvencyDetails(*createInsolvencyResource(), practitionerResourceDaos)
		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("successful validation of practitioner appointments - no practitioners are appointed", t, func() {

		insolvencyCase := createInsolvencyResource()

		// Remove appointment details for all practitioners
		practitionerResourceDao := models.PractitionerResourceDao{}
		practitionerResourceDao.Data.Appointment = nil
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)

		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("error - attachment type is not resolution and practitioners key is absent", t, func() {
		// Expect GetInsolvencyPractitionersResource to be called once and return a valid insolvency case
		insolvencyCase := models.InsolvencyResourceDao{}
		insolvencyCase.Data.Attachments = []models.AttachmentResourceDao{
			{
				Type: "type",
			},
		}

		validationErrors := ValidateInsolvencyDetails(insolvencyCase, nil)

		So(validationErrors, ShouldHaveLength, 2)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - attachment type requires that at least one practitioner must be present for insolvency case with transaction id [%s]", insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "resolution attachment type")
	})

	Convey("error - attachment type is not resolution and practitioners object is empty", t, func() {
		insolvencyCase := models.InsolvencyResourceDao{}
		insolvencyCase.Data.Attachments = []models.AttachmentResourceDao{
			{
				Type: "type",
			},
		}

		validationErrors := ValidateInsolvencyDetails(insolvencyCase, nil)

		So(validationErrors, ShouldHaveLength, 2)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - attachment type requires that at least one practitioner must be present for insolvency case with transaction id [%s]", insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "resolution attachment type")
	})

	Convey("successful validation of attachment type - attachment type is not resolution and practitioner present", t, func() {
		insolvencyCase := createInsolvencyResource()

		practitionerResourceDao := models.PractitionerResourceDao{}

		appointmentResourceDao := models.AppointmentResourceDao{}
		appointmentResourceDao.Data.AppointedOn = "2022-10-10"
		appointmentResourceDao.Data.MadeBy = "MadeBy"
		appointmentResourceDao.Data.Links = models.AppointmentResourceLinksDao{}
		appointmentResourceDao.PractitionerId = "PractitionerID"

		practitionerResourceDao.Data.Appointment = &appointmentResourceDao
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)

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

		practitionerResourceDao := models.PractitionerResourceDao{}
		appointmentResourceDao := models.AppointmentResourceDao{}
		appointmentResourceDao.Data.AppointedOn = "2022-01-23"
		appointmentResourceDao.Data.MadeBy = "MadeBy"
		appointmentResourceDao.Data.Links = models.AppointmentResourceLinksDao{}
		appointmentResourceDao.PractitionerId = "PractitionerID"
		practitionerResourceDao.Data.Appointment = &appointmentResourceDao
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)

		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("successful validation of resolution attachment - attachment type is resolution and practitioners key is absent", t, func() {
		insolvencyCase := models.InsolvencyResourceDao{}
		insolvencyCase.Data.Attachments = []models.AttachmentResourceDao{
			{
				Type: "resolution",
				ID:   "1234",
			},
		}
		insolvencyCase.Data.Resolution = &models.ResolutionResourceDao{
			DateOfResolution: "2021-06-06",
			Attachments: []string{
				"1234",
			},
		}

		validationErrors := ValidateInsolvencyDetails(insolvencyCase, nil)

		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("successful validation of resolution attachment - attachment type is resolution and practitioners object empty", t, func() {
		insolvencyCase := models.InsolvencyResourceDao{}
		insolvencyCase.Data.Attachments = []models.AttachmentResourceDao{
			{
				Type: "resolution",
				ID:   "1234",
			},
		}
		insolvencyCase.Data.Resolution = &models.ResolutionResourceDao{
			DateOfResolution: "2021-06-06",
			Attachments: []string{
				"1234",
			},
		}

		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, models.PractitionerResourceDao{})

		validationErrors := ValidateInsolvencyDetails(insolvencyCase, practitionerResourceDaos)

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

		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, models.PractitionerResourceDao{})

		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)

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

		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, models.PractitionerResourceDao{})

		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)
		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("error - attachment type is statement-of-affairs-liquidator and a practitioner is appointed", t, func() {
		insolvencyCase := createInsolvencyResource()

		// Set attachment type to "statement-of-concurrence"
		insolvencyCase.Data.Attachments[0].Type = "statement-of-affairs-liquidator"

		practitionerResourceDao := models.PractitionerResourceDao{}

		appointmentResourceDao := models.AppointmentResourceDao{}
		appointmentResourceDao.Data.AppointedOn = "2021-06-23"
		appointmentResourceDao.Data.MadeBy = "MadeBy"
		appointmentResourceDao.Data.Links = models.AppointmentResourceLinksDao{}
		appointmentResourceDao.PractitionerId = "PractitionerID"

		practitionerResourceDao.Data.Appointment = &appointmentResourceDao
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)

		So(validationErrors, ShouldHaveLength, 2)
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
		practitionerResourceDao := models.PractitionerResourceDao{}
		practitionerResourceDao.Data.Appointment = nil
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)
		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("error - no attachments present and no appointed practitioners on insolvency case", t, func() {

		practitionerResourceDao := models.PractitionerResourceDao{}
		practitionerResourceDao.Data.FirstName = "Bob"
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		insolvencyCase := models.InsolvencyResourceDao{}

		validationErrors := ValidateInsolvencyDetails(insolvencyCase, practitionerResourceDaos)

		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - at least one practitioner must be appointed as there are no attachments for insolvency case with transaction id [%s]", insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "no attachments")
	})

	Convey("error - no resolution and no submitted practitioners on insolvency case", t, func() {

		insolvencyCase := models.InsolvencyResourceDao{}
		insolvencyCase.Data.Resolution = nil

		practitionerResourceDao := models.PractitionerResourceDao{}

		appointmentResourceDao := models.AppointmentResourceDao{}
		appointmentResourceDao.Data.AppointedOn = ""
		appointmentResourceDao.Data.MadeBy = "MadeBy"
		appointmentResourceDao.Data.Links = models.AppointmentResourceLinksDao{}
		practitionerResourceDao.Data.Appointment = &appointmentResourceDao

		validationErrors := ValidateInsolvencyDetails(insolvencyCase, nil)

		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, "error - if no practitioners are present then an attachment of the type resolution must be present")
		So((*validationErrors)[0].Location, ShouldContainSubstring, "no practitioners and no resolution")
	})

	Convey("successful validation - no attachments present but at least one appointed practitioner is present on insolvency case", t, func() {

		_, practitionerResourceDao, appointmentResourceDao := generateInsolvencyPractitionerAppointmentResources()

		practitionerResourceDao.Data.Appointment = &appointmentResourceDao
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		insolvencyCase := models.InsolvencyResourceDao{}

		validationErrors := ValidateInsolvencyDetails(insolvencyCase, practitionerResourceDaos)

		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("error - resolution attachment present and no date of resolution filed for insolvency case", t, func() {
		insolvencyCase := createInsolvencyResource()
		insolvencyCase.Data.Attachments[0].Type = "resolution"
		insolvencyCase.Data.Resolution = &models.ResolutionResourceDao{
			DateOfResolution: "",
			Attachments:      []string{"123"},
		}
		practitionerResourceDao := models.PractitionerResourceDao{}

		appointmentResourceDao := models.AppointmentResourceDao{}
		appointmentResourceDao.Data.AppointedOn = ""
		appointmentResourceDao.Data.MadeBy = "MadeBy"
		appointmentResourceDao.Data.Links = models.AppointmentResourceLinksDao{}
		practitionerResourceDao.Data.Appointment = &appointmentResourceDao

		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)

		So(validationErrors, ShouldHaveLength, 5)
		So((*validationErrors)[1].Error, ShouldContainSubstring, fmt.Sprintf("error - a date of resolution must be present as there is an attachment with type resolution for insolvency case with transaction id [%s]", insolvencyCase.TransactionID))
		So((*validationErrors)[1].Location, ShouldContainSubstring, "no date of resolution")
	})

	Convey("error - resolution attachment present and no resolution details filed for insolvency case", t, func() {
		insolvencyCase := createInsolvencyResource()
		insolvencyCase.Data.Attachments[0].Type = "resolution"

		insolvencyCase.Data.Resolution = nil

		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, nil)

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

		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, models.PractitionerResourceDao{})
		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)

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

		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, models.PractitionerResourceDao{})
		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)

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

		practitionerResourceDaos := append([]models.PractitionerResourceDao{})

		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)
		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("error - statement-of-affairs-director filed but no statement date exists in DB", t, func() {

		// Create insolvency case and remove SOA date
		insolvencyCase := createInsolvencyResource()
		insolvencyCase.Data.StatementOfAffairs = &models.StatementOfAffairsResourceDao{
			StatementDate: "",
		}

		practitionerResourceDaos := append([]models.PractitionerResourceDao{})
		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)

		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - a date of statement of affairs must be present as there is an attachment with type [%s] or [%s] for insolvency case with transaction id [%s]", constants.StatementOfAffairsDirector.String(), constants.StatementOfAffairsLiquidator.String(), insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "statement-of-affairs")
	})

	Convey("error - statement-of-affairs-liquidator filed but no statement resource exists in DB", t, func() {

		// Create insolvency case
		insolvencyCase := createInsolvencyResource()

		// Change attachment type to SOA-L
		insolvencyCase.Data.Attachments[1].Type = "statement-of-affairs-liquidator"

		// Make statement date an empty string
		insolvencyCase.Data.StatementOfAffairs = &models.StatementOfAffairsResourceDao{
			StatementDate: "",
		}

		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, nil)

		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - a date of statement of affairs must be present as there is an attachment with type [%s] or [%s] for insolvency case with transaction id [%s]", constants.StatementOfAffairsDirector.String(), constants.StatementOfAffairsLiquidator.String(), insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "statement-of-affairs")
	})

	Convey("error - statement-of-affairs-director filed but no statement resource exists in DB", t, func() {

		// Create insolvency case and remove SOA date
		insolvencyCase := createInsolvencyResource()
		insolvencyCase.Data.StatementOfAffairs = nil

		practitionerResourceDaos := append([]models.PractitionerResourceDao{})

		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)

		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - a date of statement of affairs must be present as there is an attachment with type [%s] or [%s] for insolvency case with transaction id [%s]", constants.StatementOfAffairsDirector.String(), constants.StatementOfAffairsLiquidator.String(), insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "statement-of-affairs")
	})

	Convey("error - statement-of-affairs-liquidator filed but no statement resource exists in DB", t, func() {

		// Create insolvency case
		insolvencyCase := createInsolvencyResource()

		// Change attachment type to SOA-L
		insolvencyCase.Data.Attachments[1].Type = "statement-of-affairs-liquidator"

		// Remove statement resource
		insolvencyCase.Data.StatementOfAffairs = nil

		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, nil)

		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - a date of statement of affairs must be present as there is an attachment with type [%s] or [%s] for insolvency case with transaction id [%s]", constants.StatementOfAffairsDirector.String(), constants.StatementOfAffairsLiquidator.String(), insolvencyCase.TransactionID))
		So((*validationErrors)[0].Location, ShouldContainSubstring, "statement-of-affairs")
	})

	Convey("error - statement resource exists in DB but no statement-of-affairs attachment filed", t, func() {

		// Create insolvency case and remove SOA date
		insolvencyCase := createInsolvencyResource()
		insolvencyCase.Data.Attachments[1].Type = "random"

		practitionerResourceDaos := append([]models.PractitionerResourceDao{})
		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)

		So(validationErrors, ShouldHaveLength, 1)
		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - an attachment of type [%s] or [%s] must be present as there is a date of statement of affairs present for insolvency case with transaction id [%s]", constants.StatementOfAffairsDirector.String(), constants.StatementOfAffairsLiquidator.String(), insolvencyCase.TransactionID))
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
		practitionerResourceDao := models.PractitionerResourceDao{}
		appointmentResourceDao := models.AppointmentResourceDao{}
		appointmentResourceDao.Data.AppointedOn = "2021-05-05"
		practitionerResourceDao.Data.Appointment = &appointmentResourceDao
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)

		So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - practitioner appointed on [%s] is before the resolution date [%s]",
			practitionerResourceDaos[0].Data.Appointment.Data.AppointedOn, insolvencyCase.Data.Resolution.DateOfResolution))
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
		practitionerResourceDao := models.PractitionerResourceDao{}
		appointmentResourceDao := models.AppointmentResourceDao{}
		appointmentResourceDao.Data.AppointedOn = "date"
		practitionerResourceDao.Data.Appointment = &appointmentResourceDao
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)

		So((*validationErrors)[0].Error, ShouldContainSubstring, "cannot parse")
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
		practitionerResourceDao := models.PractitionerResourceDao{}
		appointmentResourceDao := models.AppointmentResourceDao{}
		appointmentResourceDao.Data.AppointedOn = "date"
		practitionerResourceDao.Data.Appointment = &appointmentResourceDao
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)

		So((*validationErrors)[0].Error, ShouldContainSubstring, "cannot parse")
		So((*validationErrors)[0].Location, ShouldContainSubstring, "practitioner")
	})

	Convey("Validate statement date and resolution date", t, func() {
		Convey("Invalid statement date", func() {
			insolvencyCase := createInsolvencyResource()
			insolvencyCase.Data.StatementOfAffairs.StatementDate = "invalid"

			practitionerResourceDao := models.PractitionerResourceDao{}
			practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

			validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)
			So((*validationErrors)[0].Error, ShouldContainSubstring, "invalid statementOfAffairs date")
		})

		Convey("Invalid resolution date", func() {
			insolvencyCase := createInsolvencyResource()
			insolvencyCase.Data.Resolution.DateOfResolution = "invalid"

			validationErrors := ValidateInsolvencyDetails(*insolvencyCase, nil)
			So((*validationErrors)[0].Error, ShouldContainSubstring, "invalid resolution date")
		})

		Convey("Statement date before resolution date", func() {
			insolvencyCase := createInsolvencyResource()
			insolvencyCase.Data.StatementOfAffairs.StatementDate = "2021-07-20"
			insolvencyCase.Data.Resolution.DateOfResolution = "2021-07-21"

			validationErrors := ValidateInsolvencyDetails(*insolvencyCase, nil)
			So((*validationErrors)[0].Error, ShouldContainSubstring, "error - statement of affairs date must not be before resolution date")
		})

		Convey("Statement date > 7 days after resolution date", func() {
			insolvencyCase := createInsolvencyResource()
			insolvencyCase.Data.StatementOfAffairs.StatementDate = "2021-07-29"
			insolvencyCase.Data.Resolution.DateOfResolution = "2021-07-21"

			validationErrors := ValidateInsolvencyDetails(*insolvencyCase, nil)
			So((*validationErrors)[0].Error, ShouldContainSubstring, "error - statement of affairs date must be within 7 days of resolution date")
		})

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

		practitionerResourceDao := models.PractitionerResourceDao{}
		appointmentResourceDao := models.AppointmentResourceDao{}
		appointmentResourceDao.Data.AppointedOn = "2022-06-06"
		practitionerResourceDao.Data.Appointment = &appointmentResourceDao
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)
		So(validationErrors, ShouldHaveLength, 0)
	})

	Convey("Validate progress report from and to dates", t, func() {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		Convey("valid submission of progress-report", func() {
			insolvencyCase := createInsolvencyResource()
			insolvencyCase.Data.ProgressReport = &models.ProgressReportResourceDao{
				FromDate: "2021-04-14",
				ToDate:   "2022-04-13",
				Attachments: []string{
					"id",
				},
			}

			practitionerResourceDaos := append([]models.PractitionerResourceDao{}, models.PractitionerResourceDao{})

			validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)

			So(validationErrors, ShouldHaveLength, 0)
		})

		Convey("progress-report attachment present and from date blank", func() {
			insolvencyCase := createInsolvencyResource()
			insolvencyCase.Data.ProgressReport = &models.ProgressReportResourceDao{
				FromDate: "",
				ToDate:   "2022-04-13",
				Attachments: []string{
					"id",
				},
			}

			practitionerResourceDaos := append([]models.PractitionerResourceDao{})

			validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)

			So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - progress report dates must be present as there is an attachment with type progress-report for insolvency case with transaction id [%s]", insolvencyCase.TransactionID))
			So((*validationErrors)[0].Location, ShouldContainSubstring, "no dates for progress report")
		})

		Convey("progress-report attachment present and to date blank", func() {
			insolvencyCase := createInsolvencyResource()
			insolvencyCase.Data.ProgressReport = &models.ProgressReportResourceDao{
				FromDate: "2021-04-14",
				ToDate:   "",
				Attachments: []string{
					"id",
				},
			}

			practitionerResourceDaos := append([]models.PractitionerResourceDao{}, models.PractitionerResourceDao{})

			validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)

			So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - progress report dates must be present as there is an attachment with type progress-report for insolvency case with transaction id [%s]", insolvencyCase.TransactionID))
			So((*validationErrors)[0].Location, ShouldContainSubstring, "no dates for progress report")
		})

		Convey("progress-report attachment present and all dates blank", func() {
			insolvencyCase := createInsolvencyResource()
			insolvencyCase.Data.ProgressReport = &models.ProgressReportResourceDao{
				FromDate: "",
				ToDate:   "",
				Attachments: []string{
					"id",
				},
			}

			practitionerResourceDaos := append([]models.PractitionerResourceDao{}, models.PractitionerResourceDao{})

			validationErrors := ValidateInsolvencyDetails(*insolvencyCase, practitionerResourceDaos)

			So((*validationErrors)[0].Error, ShouldContainSubstring, fmt.Sprintf("error - progress report dates must be present as there is an attachment with type progress-report for insolvency case with transaction id [%s]", insolvencyCase.TransactionID))
			So((*validationErrors)[0].Location, ShouldContainSubstring, "no dates for progress report")
		})
	})

}

func TestUnitValidateAntivirus(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

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

		mockService.EXPECT().UpdateAttachmentStatus(transactionID, insolvencyCase.Data.Attachments[0].ID, "integrity_failed").Return(http.StatusNoContent, nil).Times(3)

		validationErrors := ValidateAntivirus(mockService, *insolvencyCase, req)

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

		mockService.EXPECT().UpdateAttachmentStatus(transactionID, insolvencyCase.Data.Attachments[0].ID, "integrity_failed").Return(http.StatusNoContent, nil).Times(3)

		validationErrors := ValidateAntivirus(mockService, *insolvencyCase, req)

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

		mockService.EXPECT().UpdateAttachmentStatus(transactionID, insolvencyCase.Data.Attachments[0].ID, "processed").Return(http.StatusNoContent, nil).Times(3)

		validationErrors := ValidateAntivirus(mockService, *insolvencyCase, req)
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

		// Expect GetInsolvencyPractitionersResource to be called once and return an error for the insolvency case
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(createInsolvencyResource(), nil, errors.New("insolvency case does not exist")).Times(1)

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

		_, practitionerResourceDao, appointmentResourceDao := generateInsolvencyPractitionerAppointmentResources()

		practitionerResourceDao.Data.Appointment = &appointmentResourceDao
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao, practitionerResourceDao)

		// Expect GetInsolvencyPractitionersResource to be called once and return a valid insolvency case
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResource, practitionerResourceDaos, nil).Times(1)

		filings, err := GenerateFilings(mockService, transactionID)

		So(len(filings), ShouldEqual, 1)

		So(filings[0].Kind, ShouldEqual, "insolvency#600")
		So(filings[0].DescriptionIdentifier, ShouldEqual, "600")
		So(filings[0].Data, ShouldContainKey, "practitioners")
		So(filings[0].Data, ShouldNotContainKey, "attachments")

		So(err, ShouldBeNil)
	})

	Convey("Generate filing for LRESEX case with resolution attachment and no practitioners", t, func() {
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
		}

		// Expect GetInsolvencyPractitionersResource to be called once and return a valid insolvency case
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResource, nil, nil).Times(1)

		filings, err := GenerateFilings(mockService, transactionID)

		So(len(filings), ShouldEqual, 1)

		So(filings[0].Kind, ShouldEqual, "insolvency#LRESEX")
		So(filings[0].DescriptionIdentifier, ShouldEqual, "LRESEX")
		So(filings[0].Data, ShouldNotContainKey, "practitioners")
		So(len(filings[0].Data["attachments"].([]*models.AttachmentResourceDao)), ShouldEqual, 1)
		So(filings[0].Data["attachments"].([]*models.AttachmentResourceDao)[0].Type, ShouldEqual, "resolution")

		So(err, ShouldBeNil)
	})

	Convey("Generate filing for LIQ02 case with statement-of-affairs-director attachment and two practitioners", t, func() {
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

		practitionerResourceDao := models.PractitionerResourceDao{}
		practitionerResourceDao.Data.Appointment = nil
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao, practitionerResourceDao)

		// Expect GetInsolvencyPractitionersResource to be called once and return a valid insolvency case
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResource, practitionerResourceDaos, nil).Times(1)

		filings, err := GenerateFilings(mockService, transactionID)

		So(len(filings), ShouldEqual, 1)

		So(filings[0].Kind, ShouldEqual, "insolvency#LIQ02")
		So(filings[0].DescriptionIdentifier, ShouldEqual, "LIQ02")
		So(filings[0].Data, ShouldContainKey, "practitioners")
		So(len(filings[0].Data["attachments"].([]*models.AttachmentResourceDao)), ShouldEqual, 1)
		So(filings[0].Data["attachments"].([]*models.AttachmentResourceDao)[0].Type, ShouldEqual, "statement-of-affairs-director")

		So(err, ShouldBeNil)
	})

	Convey("Generate filing for LIQ02 case with statement-of-affairs-liquidator attachment and two practitioners", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect the transaction api to be called and return a closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		insolvencyResource := createInsolvencyResource()
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

		practitionerResourceDao := models.PractitionerResourceDao{}
		practitionerResourceDao.Data.Appointment = nil
		practitionerResourceDao1 := models.PractitionerResourceDao{}
		practitionerResourceDao1.Data.Appointment = nil
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao, practitionerResourceDao)

		// Expect GetInsolvencyPractitionersResource to be called once and return a valid insolvency case
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResource, practitionerResourceDaos, nil).Times(1)

		filings, err := GenerateFilings(mockService, transactionID)

		So(len(filings), ShouldEqual, 1)

		So(filings[0].Kind, ShouldEqual, "insolvency#LIQ02")
		So(filings[0].DescriptionIdentifier, ShouldEqual, "LIQ02")
		So(filings[0].Data, ShouldContainKey, "practitioners")
		So(len(filings[0].Data["attachments"].([]*models.AttachmentResourceDao)), ShouldEqual, 1)
		So(filings[0].Data["attachments"].([]*models.AttachmentResourceDao)[0].Type, ShouldEqual, "statement-of-affairs-liquidator")

		So(err, ShouldBeNil)
	})

	Convey("Generate filing for LIQ02 case with statement-of-affairs-director and statement-of-concurrence attachments and two practitioners", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect the transaction api to be called and return a closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		insolvencyResource := createInsolvencyResource()
		practitionerResourceDao := models.PractitionerResourceDao{}
		practitionerResourceDao.Data.Appointment = nil
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao, practitionerResourceDao)
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
			{
				ID:     "id",
				Type:   "statement-of-concurrence",
				Status: "status",
				Links: models.AttachmentResourceLinksDao{
					Self:     "self",
					Download: "download",
				},
			},
		}

		// Expect GetInsolvencyPractitionersResource to be called once and return a valid insolvency case
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResource, practitionerResourceDaos, nil).Times(1)

		filings, err := GenerateFilings(mockService, transactionID)

		So(len(filings), ShouldEqual, 1)

		So(filings[0].Kind, ShouldEqual, "insolvency#LIQ02")
		So(filings[0].DescriptionIdentifier, ShouldEqual, "LIQ02")
		So(filings[0].Data, ShouldContainKey, "practitioners")
		So(len(filings[0].Data["attachments"].([]*models.AttachmentResourceDao)), ShouldEqual, 2)
		So(filings[0].Data["attachments"].([]*models.AttachmentResourceDao)[0].Type, ShouldEqual, "statement-of-affairs-director")
		So(filings[0].Data["attachments"].([]*models.AttachmentResourceDao)[1].Type, ShouldEqual, "statement-of-concurrence")

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
		_, practitionerResourceDao, appointmentResourceDao := generateInsolvencyPractitionerAppointmentResources()

		practitionerResourceDao.Data.Appointment = &appointmentResourceDao
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao, practitionerResourceDao)

		// Expect GetInsolvencyPractitionersResource to be called once and return a valid insolvency case
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResource, practitionerResourceDaos, nil).Times(1)

		filings, err := GenerateFilings(mockService, transactionID)

		So(len(filings), ShouldEqual, 2)

		So(filings[0].Kind, ShouldEqual, "insolvency#600")
		So(filings[0].DescriptionIdentifier, ShouldEqual, "600")
		So(filings[0].Data, ShouldContainKey, "practitioners")
		So(filings[0].Data, ShouldNotContainKey, "attachments")

		So(filings[1].Kind, ShouldEqual, "insolvency#LIQ02")
		So(filings[1].DescriptionIdentifier, ShouldEqual, "LIQ02")
		So(filings[1].Data, ShouldContainKey, "practitioners")
		So(len(filings[1].Data["attachments"].([]*models.AttachmentResourceDao)), ShouldEqual, 1)
		So(filings[1].Data["attachments"].([]*models.AttachmentResourceDao)[0].Type, ShouldEqual, "statement-of-affairs-director")

		So(err, ShouldBeNil)
	})

	Convey("Generate filing for 600, LRESEX, and LIQ02 case with statement-of-affairs-director and statement-of-concurrence attachments and two practitioners", t, func() {
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
				Type:   "statement-of-concurrence",
				Status: "status",
				Links: models.AttachmentResourceLinksDao{
					Self:     "self",
					Download: "download",
				},
			},
		}

		_, practitionerResourceDao, appointmentResourceDao := generateInsolvencyPractitionerAppointmentResources()

		practitionerResourceDao.Data.Appointment = &appointmentResourceDao
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao, practitionerResourceDao)

		// Expect GetInsolvencyPractitionersResource to be called once and return a valid insolvency case
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResource, practitionerResourceDaos, nil).Times(1)

		filings, err := GenerateFilings(mockService, transactionID)

		So(len(filings), ShouldEqual, 3)

		So(filings[0].Kind, ShouldEqual, "insolvency#600")
		So(filings[0].DescriptionIdentifier, ShouldEqual, "600")
		So(filings[0].Data, ShouldContainKey, "practitioners")
		So(filings[0].Data, ShouldNotContainKey, "attachments")

		So(filings[1].Kind, ShouldEqual, "insolvency#LRESEX")
		So(filings[1].DescriptionIdentifier, ShouldEqual, "LRESEX")
		So(filings[1].Data, ShouldNotContainKey, "practitioners")
		So(len(filings[1].Data["attachments"].([]*models.AttachmentResourceDao)), ShouldEqual, 1)
		So(filings[1].Data["attachments"].([]*models.AttachmentResourceDao)[0].Type, ShouldEqual, "resolution")

		So(filings[2].Kind, ShouldEqual, "insolvency#LIQ02")
		So(filings[2].DescriptionIdentifier, ShouldEqual, "LIQ02")
		So(filings[2].Data, ShouldContainKey, "practitioners")
		So(len(filings[2].Data["attachments"].([]*models.AttachmentResourceDao)), ShouldEqual, 2)
		So(filings[2].Data["attachments"].([]*models.AttachmentResourceDao)[0].Type, ShouldEqual, "statement-of-affairs-director")
		So(filings[2].Data["attachments"].([]*models.AttachmentResourceDao)[1].Type, ShouldEqual, "statement-of-concurrence")

		So(err, ShouldBeNil)
	})

	Convey("Generate filing for LIQ03 case with progress-report attachment and one practitioner", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect the transaction api to be called and return a closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		_, practitionerResourceDao, _ := generateInsolvencyPractitionerAppointmentResources()
		insolvencyResource := createInsolvencyResource()

		practitionerResourceDao.Data.Appointment = nil
		practitionerResourceDao.Data.Address = models.AddressResourceDao{}
		practitionerResourceDao.Data.Role = "final-liquidator"
		practitionerResourceDao.Data.Links = models.PractitionerResourceLinksDao{}
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao, practitionerResourceDao)
		insolvencyResource.Data.Attachments = []models.AttachmentResourceDao{
			{
				ID:     "id",
				Type:   "progress-report",
				Status: "status",
				Links: models.AttachmentResourceLinksDao{
					Self:     "self",
					Download: "download",
				},
			},
		}

		// Expect GetInsolvencyPractitionersResource to be called once and return a valid insolvency case
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResource, practitionerResourceDaos, nil).Times(1)

		filings, err := GenerateFilings(mockService, transactionID)

		So(len(filings), ShouldEqual, 1)

		So(filings[0].Kind, ShouldEqual, "insolvency#LIQ03")
		So(filings[0].DescriptionIdentifier, ShouldEqual, "LIQ03")
		So(filings[0].Data, ShouldContainKey, "practitioners")
		So(len(filings[0].Data["attachments"].([]*models.AttachmentResourceDao)), ShouldEqual, 1)
		So(filings[0].Data["attachments"].([]*models.AttachmentResourceDao)[0].Type, ShouldEqual, "progress-report")

		So(err, ShouldBeNil)
	})

	Convey("Generate filing for 600 and LIQ03 case with progress-report attachment and one practitioner", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mocks.NewMockService(mockCtrl)

		// Expect the transaction api to be called and return a closed transaction
		httpmock.RegisterResponder(http.MethodGet, "https://api.companieshouse.gov.uk/transactions/12345678", httpmock.NewStringResponder(http.StatusOK, transactionProfileResponseClosed))

		_, practitionerResourceDao, appointmentResourceDao := generateInsolvencyPractitionerAppointmentResources()
		insolvencyResource := createInsolvencyResource()

		practitionerResourceDao.Data.Appointment = &appointmentResourceDao
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)
		insolvencyResource.Data.Attachments = []models.AttachmentResourceDao{
			{
				ID:     "id",
				Type:   "progress-report",
				Status: "status",
				Links: models.AttachmentResourceLinksDao{
					Self:     "self",
					Download: "download",
				},
			},
		}

		// Expect GetInsolvencyPractitionersResource to be called once and return a valid insolvency case
		mockService.EXPECT().GetInsolvencyPractitionersResource(transactionID).Return(insolvencyResource, practitionerResourceDaos, nil).Times(1)

		filings, err := GenerateFilings(mockService, transactionID)

		So(len(filings), ShouldEqual, 2)

		So(filings[0].Kind, ShouldEqual, "insolvency#600")
		So(filings[0].DescriptionIdentifier, ShouldEqual, "600")
		So(filings[0].Data, ShouldContainKey, "practitioners")
		So(filings[0].Data, ShouldNotContainKey, "attachments")

		So(filings[1].Kind, ShouldEqual, "insolvency#LIQ03")
		So(filings[1].DescriptionIdentifier, ShouldEqual, "LIQ03")
		So(filings[1].Data, ShouldContainKey, "practitioners")
		So(len(filings[1].Data["attachments"].([]*models.AttachmentResourceDao)), ShouldEqual, 1)
		So(filings[1].Data["attachments"].([]*models.AttachmentResourceDao)[0].Type, ShouldEqual, "progress-report")

		So(err, ShouldBeNil)
	})

}
