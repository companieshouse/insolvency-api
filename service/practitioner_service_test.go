package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/companieshouse/insolvency-api/constants"

	mock_dao "github.com/companieshouse/insolvency-api/mocks"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/golang/mock/gomock"
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitIsValidPractitionerDetails(t *testing.T) {

	Convey("Practitioner request supplied is invalid - neither email or telephone number are supplied", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = ""
		practitioner.Email = ""

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		err, _ := ValidatePractitionerDetails(insolvencyResourceDao, transactionID, practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "either telephone_number or email are required")
	})

	Convey("Practitioner request supplied is valid - email is supplied", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = ""

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		err, _ := ValidatePractitionerDetails(insolvencyResourceDao, transactionID, practitioner)

		So(err, ShouldBeBlank)
	})

	Convey("Practitioner request supplied is valid - telephone number is supplied", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.Email = ""

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		err, _ := ValidatePractitionerDetails(insolvencyResourceDao, transactionID, practitioner)

		So(err, ShouldBeBlank)
	})

	Convey("Practitioner request supplied is invalid - telephone number is less than 10 digits", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = "07777777"

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		err, _ := ValidatePractitionerDetails(insolvencyResourceDao, transactionID, practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "telephone_number must be 10 or 11 digits long")
	})

	Convey("Practitioner request supplied is invalid - telephone number is more than 11 digits", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = "077777777777"

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		err, _ := ValidatePractitionerDetails(insolvencyResourceDao, transactionID, practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "telephone_number must be 10 or 11 digits long")
	})

	Convey("Practitioner request supplied is invalid - telephone number does not consist solely of digits", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = "077777777OO"

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		err, _ := ValidatePractitionerDetails(insolvencyResourceDao, transactionID, practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "telephone_number must start with 0 and contain only numeric characters")
	})

	Convey("Practitioner request supplied is invalid - telephone number does not consist solely of digits", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = "07777OO"

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		err, _ := ValidatePractitionerDetails(insolvencyResourceDao, transactionID, practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "telephone_number must start with 0 and contain only numeric characters")
		So(err, ShouldContainSubstring, "telephone_number must be 10 or 11 digits long")
	})

	Convey("Practitioner request supplied is invalid - telephone number contains spaces", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = "0777777 7777"

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()

		err, _ := ValidatePractitionerDetails(insolvencyResourceDao, transactionID, practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "telephone_number must not contain spaces")
	})

	Convey("Practitioner request supplied is invalid - telephone number does not begin with 0", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = "77777777777"

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		err, _ := ValidatePractitionerDetails(insolvencyResourceDao, transactionID, practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "telephone_number must start with 0 and contain only numeric characters")
	})

	Convey("Practitioner request supplied is invalid - first name does not match regex", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.FirstName = "wr0ng"

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		err, _ := ValidatePractitionerDetails(insolvencyResourceDao, transactionID, practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "the first name contains a character which is not allowed")
	})

	Convey("Practitioner request supplied is invalid - last name does not match regex", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.LastName = "wr0ng"

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		err, _ := ValidatePractitionerDetails(insolvencyResourceDao, transactionID, practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "the last name contains a character which is not allowed")
	})

	Convey("Practitioner request supplied is invalid - first and last name does not match regex", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.FirstName = "name?"
		practitioner.LastName = "wr0ng"

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		err, _ := ValidatePractitionerDetails(insolvencyResourceDao, transactionID, practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "the first name contains a character which is not allowed")
		So(err, ShouldContainSubstring, "the last name contains a character which is not allowed")
	})

	Convey("Practitioner request supplied is invalid - first and last name does not match regex and contact details missing", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.FirstName = "name?"
		practitioner.LastName = "wr0ng"
		practitioner.Email = ""
		practitioner.TelephoneNumber = ""

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		err, _ := ValidatePractitionerDetails(insolvencyResourceDao, transactionID, practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "either telephone_number or email are required")
		So(err, ShouldContainSubstring, "the first name contains a character which is not allowed")
		So(err, ShouldContainSubstring, "the last name contains a character which is not allowed")
	})

	Convey("Practitioner request supplied is invalid - role supplied is incorrect for CVL case", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.Role = constants.Receiver.String()

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		err, _ := ValidatePractitionerDetails(insolvencyResourceDao, transactionID, practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, fmt.Sprintf("the practitioner role must be "+constants.FinalLiquidator.String()+" because the insolvency case for transaction ID [%s] is of type "+constants.CVL.String(), transactionID))
	})

	Convey("Error retrieving insolvency case when validating practitioner", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()
		practitioner.Role = constants.Receiver.String()

		_, err := ValidatePractitionerDetails(&models.InsolvencyResourceDao{}, transactionID, practitioner)

		So(err, ShouldBeNil)
	})

	Convey("Practitioner request supplied is valid - both telephone number and email are supplied", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		practitioner := generatePractitioner()

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		err, _ := ValidatePractitionerDetails(insolvencyResourceDao, transactionID, practitioner)

		So(err, ShouldBeBlank)
	})
}

func TestUnitIsValidAppointment(t *testing.T) {
	transactionID := "123"
	practitionerID := "456"
	apiURL := "https://api.companieshouse.gov.uk"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	Convey("error getting practitioners", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		mockService.EXPECT().GetInsolvencyAndExpandedPractitionerResources(gomock.Any()).Return(insolvencyResourceDao, nil, fmt.Errorf("err"))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateAppointmentDetails(mockService, generateAppointment(), transactionID, practitionerID, req)

		So(err.Error(), ShouldContainSubstring, "err")
		So(validationErr[0], ShouldEqual, "err")
	})

	Convey("practitioner already appointed", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService := mock_dao.NewMockService(mockCtrl)

		_, practitionerResourceDao, _ := generateInsolvencyPractitionerAppointmentResources()

		appointmentResourceDao := models.AppointmentResourceDao{}
		appointmentResourceDao.Data.AppointedOn = "2012-01-23"
		appointmentResourceDao.Data.MadeBy = "creditors"

		practitionerResourceDao.Data.Appointment = &appointmentResourceDao
		practitionerResourceDao.Data.Links = models.PractitionerResourceLinksDao{
			Self:        "/transactions/123/insolvency/practitioners/456",
			Appointment: "{\"456\":\"/transactions/123/insolvency/practitioners/456/appointment\"}",
		}

		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		mockService.EXPECT().GetInsolvencyAndExpandedPractitionerResources(transactionID).Return(insolvencyResourceDao, practitionerResourceDaos, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErrs, err := ValidateAppointmentDetails(mockService, generateAppointment(), transactionID, practitionerID, req)

		So(err, ShouldBeNil)
		So(validationErrs[0], ShouldContainSubstring, "already appointed")
	})

	Convey("error retrieving insolvency resource", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mock_dao.NewMockService(mockCtrl)

		mockService.EXPECT().GetInsolvencyAndExpandedPractitionerResources(transactionID).Return(&models.InsolvencyResourceDao{}, nil, fmt.Errorf("err"))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateAppointmentDetails(mockService, generateAppointment(), transactionID, practitionerID, req)

		So(err.Error(), ShouldContainSubstring, "err")
		So(validationErr, ShouldNotBeEmpty)
	})

	Convey("error retrieving company details", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusTeapot, ""))

		mockService := mock_dao.NewMockService(mockCtrl)

		insolvencyResourceDao, practitionerResourceDao, _ := generateInsolvencyPractitionerAppointmentResources()
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		mockService.EXPECT().GetInsolvencyAndExpandedPractitionerResources(gomock.Any()).Return(insolvencyResourceDao, practitionerResourceDaos, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateAppointmentDetails(mockService, generateAppointment(), transactionID, practitionerID, req)

		So(validationErr, ShouldBeEmpty)
		So(err.Error(), ShouldContainSubstring, "error getting company details from DB")
	})

	Convey("error parsing appointedOn date", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService := mock_dao.NewMockService(mockCtrl)

		insolvencyResourceDao, practitionerResourceDao, _ := generateInsolvencyPractitionerAppointmentResources()
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		mockService.EXPECT().GetInsolvencyAndExpandedPractitionerResources(gomock.Any()).Return(insolvencyResourceDao, practitionerResourceDaos, nil)

		appointment := generateAppointment()
		appointment.AppointedOn = "2001/1/2"

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateAppointmentDetails(mockService, appointment, transactionID, practitionerID, req)

		So(validationErr, ShouldBeEmpty)
		So(err.Error(), ShouldContainSubstring, "error parsing date")
	})

	Convey("error parsing incorporatedOn date", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("error")))

		mockService := mock_dao.NewMockService(mockCtrl)

		insolvencyResourceDao, practitionerResourceDao, _ := generateInsolvencyPractitionerAppointmentResources()
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		mockService.EXPECT().GetInsolvencyAndExpandedPractitionerResources(gomock.Any()).Return(insolvencyResourceDao, practitionerResourceDaos, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateAppointmentDetails(mockService, generateAppointment(), transactionID, practitionerID, req)

		So(validationErr, ShouldBeEmpty)
		So(err.Error(), ShouldContainSubstring, "error parsing date")
	})

	Convey("invalid appointedOn date - in the future", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService := mock_dao.NewMockService(mockCtrl)

		insolvencyResourceDao, practitionerResourceDao, _ := generateInsolvencyPractitionerAppointmentResources()
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		mockService.EXPECT().GetInsolvencyAndExpandedPractitionerResources(gomock.Any()).Return(insolvencyResourceDao, practitionerResourceDaos, nil)

		appointment := generateAppointment()
		appointment.AppointedOn = time.Now().AddDate(0, 0, 1).Format("2006-01-02")

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateAppointmentDetails(mockService, appointment, transactionID, practitionerID, req)

		So(validationErr[0], ShouldContainSubstring, "should not be in the future")
		So(err, ShouldBeNil)
	})

	Convey("invalid appointedOn date - before company was incorporated", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService := mock_dao.NewMockService(mockCtrl)

		insolvencyResourceDao, practitionerResourceDao, _ := generateInsolvencyPractitionerAppointmentResources()
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		mockService.EXPECT().GetInsolvencyAndExpandedPractitionerResources(gomock.Any()).Return(insolvencyResourceDao, practitionerResourceDaos, nil)

		appointment := generateAppointment()
		appointment.AppointedOn = "1999-01-01"

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateAppointmentDetails(mockService, appointment, transactionID, practitionerID, req)

		So(validationErr[0], ShouldContainSubstring, "before the company was incorporated")
		So(err, ShouldBeNil)
	})

	Convey("invalid appointedOn date - different from already appointed practitioners", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		practitionersResponseDao := models.PractitionerResourceDao{}

		appointmentResourceDao := models.AppointmentResourceDao{}
		appointmentResourceDao.Data.AppointedOn = "2012-01-23"

		practitionersResponseDao.Data.Appointment = &appointmentResourceDao
		practitionersResponse := append([]models.PractitionerResourceDao{}, practitionersResponseDao)

		_, practitionerResourceDao, _ := generateInsolvencyPractitionerAppointmentResources()

		practitionerResourceDao.Data.Appointment = &appointmentResourceDao
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		mockService := mock_dao.NewMockService(mockCtrl)

		insolvencyResourceDao, _, _ := generateInsolvencyPractitionerAppointmentResources()
		mockService.EXPECT().GetInsolvencyAndExpandedPractitionerResources(transactionID).Return(insolvencyResourceDao, practitionerResourceDaos, nil)

		appointment := generateAppointment()
		appointment.AppointedOn = "2012-01-24"

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateAppointmentDetails(mockService, appointment, transactionID, "111", req)

		So(validationErr[0], ShouldEqual, fmt.Sprintf("appointed_on [%s] differs from practitioner who was appointed on [%s]", appointment.AppointedOn, practitionersResponse[0].Data.Appointment.Data.AppointedOn))
		So(err, ShouldBeNil)
	})

	Convey("invalid madeBy - creditors madeBy not supplied for CVL case", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService := mock_dao.NewMockService(mockCtrl)

		insolvencyResourceDao, practitionerResourceDao, _ := generateInsolvencyPractitionerAppointmentResources()
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		mockService.EXPECT().GetInsolvencyAndExpandedPractitionerResources(gomock.Any()).Return(insolvencyResourceDao, practitionerResourceDaos, nil)

		appointment := generateAppointment()
		appointment.MadeBy = "company"

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateAppointmentDetails(mockService, appointment, transactionID, practitionerID, req)

		So(validationErr[0], ShouldEqual, fmt.Sprintf("made_by cannot be [%s] for insolvency case of type CVL", appointment.MadeBy))
		So(err, ShouldBeNil)
	})

	Convey("failed to create appointment when transactionID is not valid", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService := mock_dao.NewMockService(mockCtrl)

		insolvencyResourceDao, practitionerResourceDao, _ := generateInsolvencyPractitionerAppointmentResources()
		practitionerResourceDao.Data.Links = models.PractitionerResourceLinksDao{
			Self: "/transactions/12356/insolvency/practitioners/456",
		}
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		mockService.EXPECT().GetInsolvencyAndExpandedPractitionerResources(gomock.Any()).Return(insolvencyResourceDao, practitionerResourceDaos, nil)

		appointment := generateAppointment()
		appointment.MadeBy = "creditors"

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateAppointmentDetails(mockService, appointment, transactionID, practitionerID, req)

		So(validationErr, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
		So(validationErr[0], ShouldEqual, "practitioner ID [456] and transactionID[123] are not valid to create appointment")
	})

	Convey("create appointment when practitioner has no appointment links", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService := mock_dao.NewMockService(mockCtrl)

		insolvencyResourceDao, practitionerResourceDao, _ := generateInsolvencyPractitionerAppointmentResources()
		practitionerResourceDao.Data.Links = models.PractitionerResourceLinksDao{
			Self: "/transactions/123/insolvency/practitioners/456",
		}
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		mockService.EXPECT().GetInsolvencyAndExpandedPractitionerResources(gomock.Any()).Return(insolvencyResourceDao, practitionerResourceDaos, nil)

		appointment := generateAppointment()
		appointment.MadeBy = "creditors"

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateAppointmentDetails(mockService, appointment, transactionID, practitionerID, req)

		So(validationErr, ShouldBeEmpty)
		So(err, ShouldBeNil)
	})

	Convey("valid appointment", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		mockService := mock_dao.NewMockService(mockCtrl)

		insolvencyResourceDao, practitionerResourceDao, _ := generateInsolvencyPractitionerAppointmentResources()
		practitionerResourceDaos := append([]models.PractitionerResourceDao{}, practitionerResourceDao)

		mockService.EXPECT().GetInsolvencyAndExpandedPractitionerResources(gomock.Any()).Return(insolvencyResourceDao, practitionerResourceDaos, nil)

		appointment := generateAppointment()
		appointment.MadeBy = "creditors"

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateAppointmentDetails(mockService, appointment, transactionID, practitionerID, req)

		So(validationErr, ShouldBeEmpty)
		So(err, ShouldBeNil)
	})

}

func generatePractitioner() models.PractitionerRequest {
	return models.PractitionerRequest{
		IPCode:          "1234",
		FirstName:       "Joe",
		LastName:        "Bloggs",
		TelephoneNumber: "01234567890",
		Email:           "a@b.com",
		Address: models.Address{
			AddressLine1: "addressline1",
			Locality:     "locality",
		},
		Role: constants.FinalLiquidator.String(),
	}
}

func generateAppointment() models.PractitionerAppointment {
	return models.PractitionerAppointment{
		AppointedOn: "2012-01-23",
		MadeBy:      "creditors",
	}
}

func generateInsolvencyPractitionerAppointmentResources() (*models.InsolvencyResourceDao, models.PractitionerResourceDao, models.AppointmentResourceDao) {

	practitionerID := "456"
	practitionerResourceDao := models.PractitionerResourceDao{}

	practitionerResourceDao.Data.PractitionerId = practitionerID
	practitionerResourceDao.Data.IPCode = "1111"
	practitionerResourceDao.Data.FirstName = "First"
	practitionerResourceDao.Data.LastName = "First"
	practitionerResourceDao.Data.TelephoneNumber = "TelephoneNumber"
	practitionerResourceDao.Data.Email = "email@email.com"
	practitionerResourceDao.Data.Role = "role"
	practitionerResourceDao.Data.Appointment = &models.AppointmentResourceDao{}
	practitionerResourceDao.Data.Links = models.PractitionerResourceLinksDao{
		Self: "/transactions/123/insolvency/practitioners/456",
	}

	appointmentResourceDao := models.AppointmentResourceDao{}
	appointmentResourceDao.Data.AppointedOn = "2012-01-23"
	appointmentResourceDao.Data.MadeBy = "MadeBy"
	appointmentResourceDao.Data.Links = models.AppointmentResourceLinksDao{}
	appointmentResourceDao.PractitionerId = "PractitionerID"

	insolvencyResourceDaoData := models.InsolvencyResourceDao{}
	insolvencyResourceDaoData.Data.CompanyNumber = "1234"
	insolvencyResourceDaoData.Data.CaseType = "creditors-voluntary-liquidation"
	insolvencyResourceDaoData.Data.CompanyName = "Company"

	return &insolvencyResourceDaoData, practitionerResourceDao, appointmentResourceDao
}

func companyProfileDateResponse(dateOfCreation string) string {
	return `
{
 "company_name": "companyName",
 "company_number": "01234567",
 "jurisdiction": "england-wales",
 "company_status": "active",
 "type": "private-shares-exemption-30",
 "date_of_creation": "` + dateOfCreation + `",
 "registered_office_address" : {
   "postal_code" : "CF14 3UZ",
   "address_line_2" : "Cardiff",
   "address_line_1" : "1 Crown Way"
  }
}
`

}
