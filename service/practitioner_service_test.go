package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/companieshouse/insolvency-api/mocks"
	"github.com/golang/mock/gomock"
	"github.com/jarcoal/httpmock"

	"github.com/companieshouse/insolvency-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestIsValidPractitionerDetails(t *testing.T) {

	Convey("Practitioner request supplied is invalid - neither email or telephone number are supplied", t, func() {
		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = ""
		practitioner.Email = ""

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "either telephone_number or email are required")
	})

	Convey("Practitioner request supplied is valid - email is supplied", t, func() {
		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = ""

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldBeBlank)
	})

	Convey("Practitioner request supplied is valid - telephone number is supplied", t, func() {
		practitioner := generatePractitioner()
		practitioner.Email = ""

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldBeBlank)
	})

	Convey("Practitioner request supplied is invalid - telephone number is less than 10 digits", t, func() {
		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = "07777777"

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "telephone_number must be a valid format")
	})

	Convey("Practitioner request supplied is invalid - telephone number is more than 11 digits", t, func() {
		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = "077777777777"

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "telephone_number must be a valid format")
	})

	Convey("Practitioner request supplied is invalid - telephone number does not consist solely of digits", t, func() {
		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = "077777777OO"

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "telephone_number must be a valid format")
	})

	Convey("Practitioner request supplied is invalid - telephone number contains spaces", t, func() {
		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = "0777777 7777"

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "telephone_number must not contain spaces")
	})

	Convey("Practitioner request supplied is invalid - telephone number does not begin with 0", t, func() {
		practitioner := generatePractitioner()
		practitioner.TelephoneNumber = "77777777777"

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "telephone_number must start with 0")
	})

	Convey("Practitioner request supplied is invalid - first name does not match regex", t, func() {
		practitioner := generatePractitioner()
		practitioner.FirstName = "wr0ng"

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "the first name contains a character which is not allowed")
	})

	Convey("Practitioner request supplied is invalid - last name does not match regex", t, func() {
		practitioner := generatePractitioner()
		practitioner.LastName = "wr0ng"

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "the last name contains a character which is not allowed")
	})

	Convey("Practitioner request supplied is invalid - first and last name does not match regex", t, func() {
		practitioner := generatePractitioner()
		practitioner.FirstName = "name?"
		practitioner.LastName = "wr0ng"

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "the first name contains a character which is not allowed")
		So(err, ShouldContainSubstring, "the last name contains a character which is not allowed")
	})

	Convey("Practitioner request supplied is invalid - first and last name does not match regex and contact details missing", t, func() {
		practitioner := generatePractitioner()
		practitioner.FirstName = "name?"
		practitioner.LastName = "wr0ng"
		practitioner.Email = ""
		practitioner.TelephoneNumber = ""

		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldNotBeBlank)
		So(err, ShouldContainSubstring, "either telephone_number or email are required")
		So(err, ShouldContainSubstring, "the first name contains a character which is not allowed")
		So(err, ShouldContainSubstring, "the last name contains a character which is not allowed")
	})

	Convey("Practitioner request supplied is valid - both telephone number and email are supplied", t, func() {
		practitioner := generatePractitioner()
		err := ValidatePractitionerDetails(practitioner)

		So(err, ShouldBeBlank)
	})
}

func TestIsValidAppointment(t *testing.T) {
	transactionID := "123"
	practitionerID := "456"
	apiURL := "https://api.companieshouse.gov.uk"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	Convey("error getting practitioners", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetPractitionerResources(gomock.Any()).Return(nil, fmt.Errorf("err"))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateAppointmentDetails(mockService, generateAppointment(), transactionID, practitionerID, req)
		So(err.Error(), ShouldContainSubstring, "err")
		So(validationErr, ShouldBeEmpty)
	})

	Convey("practitioner already appointed", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		practitionersResponse := []models.PractitionerResourceDao{
			{
				ID: practitionerID,
				Appointment: &models.AppointmentResourceDao{
					AppointedOn: "2012-01-23",
				},
			},
		}
		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetPractitionerResources(gomock.Any()).Return(practitionersResponse, nil)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErrs, err := ValidateAppointmentDetails(mockService, generateAppointment(), transactionID, practitionerID, req)
		So(err, ShouldBeNil)
		So(validationErrs, ShouldContainSubstring, "already appointed")
	})

	Convey("error retrieving insolvency resource", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		practitionersResponse := []models.PractitionerResourceDao{
			{
				ID: practitionerID,
				Appointment: &models.AppointmentResourceDao{
					AppointedOn: "2012-01-23",
				},
			},
		}
		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetPractitionerResources(gomock.Any()).Return(practitionersResponse, nil)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(models.InsolvencyResourceDao{}, fmt.Errorf("err"))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateAppointmentDetails(mockService, generateAppointment(), transactionID, practitionerID, req)
		So(err.Error(), ShouldContainSubstring, "err")
		So(validationErr, ShouldBeEmpty)
	})

	Convey("error retrieving company details", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusTeapot, ""))

		practitionersResponse := []models.PractitionerResourceDao{
			{
				ID: practitionerID,
				Appointment: &models.AppointmentResourceDao{
					AppointedOn: "2012-01-23",
				},
			},
		}
		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetPractitionerResources(gomock.Any()).Return(practitionersResponse, nil)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)

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

		practitionersResponse := []models.PractitionerResourceDao{
			{
				ID: practitionerID,
				Appointment: &models.AppointmentResourceDao{
					AppointedOn: "2012-01-23",
				},
			},
		}
		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetPractitionerResources(gomock.Any()).Return(practitionersResponse, nil)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)

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

		practitionersResponse := []models.PractitionerResourceDao{
			{
				ID: practitionerID,
				Appointment: &models.AppointmentResourceDao{
					AppointedOn: "2012-01-23",
				},
			},
		}
		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetPractitionerResources(gomock.Any()).Return(practitionersResponse, nil)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)

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

		practitionersResponse := []models.PractitionerResourceDao{
			{
				ID: practitionerID,
				Appointment: &models.AppointmentResourceDao{
					AppointedOn: "2012-01-23",
				},
			},
		}
		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetPractitionerResources(gomock.Any()).Return(practitionersResponse, nil)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)

		appointment := generateAppointment()
		appointment.AppointedOn = time.Now().AddDate(0, 0, 1).Format("2006-01-02")

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateAppointmentDetails(mockService, appointment, transactionID, "111", req)
		So(validationErr, ShouldContainSubstring, "should not be in the future")
		So(err, ShouldBeNil)
	})

	Convey("invalid appointedOn date - before company was incorporated", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		practitionersResponse := []models.PractitionerResourceDao{
			{
				ID: practitionerID,
				Appointment: &models.AppointmentResourceDao{
					AppointedOn: "2012-01-23",
				},
			},
		}
		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetPractitionerResources(gomock.Any()).Return(practitionersResponse, nil)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)

		appointment := generateAppointment()
		appointment.AppointedOn = "1999-01-01"

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateAppointmentDetails(mockService, appointment, transactionID, "111", req)
		So(validationErr, ShouldContainSubstring, "before the company was incorporated")
		So(err, ShouldBeNil)
	})

	Convey("invalid appointedOn date - different from already appointed practitioners", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		practitionersResponse := []models.PractitionerResourceDao{
			{
				ID: practitionerID,
				Appointment: &models.AppointmentResourceDao{
					AppointedOn: "2012-01-23",
				},
			},
		}
		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetPractitionerResources(gomock.Any()).Return(practitionersResponse, nil)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)

		appointment := generateAppointment()
		appointment.AppointedOn = "2012-01-24"

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateAppointmentDetails(mockService, appointment, transactionID, "111", req)
		So(validationErr, ShouldEqual, fmt.Sprintf("appointed_on [%s] differs from practitioner ID [%s] who was appointed on [%s]", appointment.AppointedOn, practitionerID, practitionersResponse[0].Appointment.AppointedOn))
		So(err, ShouldBeNil)
	})

	Convey("invalid madeBy - creditors madeBy not supplied for CVL case", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		practitionersResponse := []models.PractitionerResourceDao{
			{
				ID: practitionerID,
				Appointment: &models.AppointmentResourceDao{
					AppointedOn: "2012-01-23",
				},
			},
		}
		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetPractitionerResources(gomock.Any()).Return(practitionersResponse, nil)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)

		appointment := generateAppointment()
		appointment.MadeBy = "company"

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateAppointmentDetails(mockService, appointment, transactionID, "111", req)
		So(validationErr, ShouldEqual, fmt.Sprintf("made_by cannot be [%s] for insolvency case of type CVL", appointment.MadeBy))
		So(err, ShouldBeNil)
	})

	Convey("valid appointment", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		practitionersResponse := []models.PractitionerResourceDao{
			{
				ID: practitionerID,
				Appointment: &models.AppointmentResourceDao{
					AppointedOn: "2012-01-23",
				},
			},
		}
		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetPractitionerResources(gomock.Any()).Return(practitionersResponse, nil)
		mockService.EXPECT().GetInsolvencyResource(transactionID).Return(generateInsolvencyResource(), nil)

		appointment := generateAppointment()
		appointment.MadeBy = "creditors"

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		validationErr, err := ValidateAppointmentDetails(mockService, appointment, transactionID, "111", req)
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
		Role: "role",
	}
}

func generateAppointment() models.PractitionerAppointment {
	return models.PractitionerAppointment{
		AppointedOn: "2012-01-23",
		MadeBy:      "creditors",
	}
}

func generateInsolvencyResource() models.InsolvencyResourceDao {
	return models.InsolvencyResourceDao{
		Data: models.InsolvencyResourceDaoData{
			CompanyNumber: "1234",
			CaseType:      "creditors-voluntary-liquidation",
			CompanyName:   "Company",
			Practitioners: []models.PractitionerResourceDao{
				{
					ID:              "1234",
					IPCode:          "1111",
					FirstName:       "First",
					LastName:        "Last",
					TelephoneNumber: "12345678901",
					Email:           "email@email.com",
					Address:         models.AddressResourceDao{},
					Role:            "role",
					Links:           models.PractitionerResourceLinksDao{},
					Appointment:     nil,
				},
			},
		},
	}
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
