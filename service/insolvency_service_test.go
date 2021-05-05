package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/companieshouse/insolvency-api/mocks"
	"github.com/companieshouse/insolvency-api/models"
	"github.com/golang/mock/gomock"
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitCheckPractitionerAlreadyAppointed(t *testing.T) {
	transactionID := "123"
	practitionerID := "456"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	Convey("error getting practitioners", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetPractitionerResources(gomock.Any()).Return(nil, fmt.Errorf("err"))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		err, appointed := CheckPractitionerAlreadyAppointed(mockService, transactionID, practitionerID, req)
		So(err.Error(), ShouldEqual, "err")
		So(appointed, ShouldBeFalse)
	})

	Convey("practitioner already appointed", t, func() {
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

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		err, appointed := CheckPractitionerAlreadyAppointed(mockService, transactionID, practitionerID, req)
		So(err, ShouldBeNil)
		So(appointed, ShouldBeTrue)
	})

	Convey("practitioner not already appointed", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		practitionersResponse := []models.PractitionerResourceDao{{ID: practitionerID}}
		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetPractitionerResources(gomock.Any()).Return(practitionersResponse, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		err, appointed := CheckPractitionerAlreadyAppointed(mockService, transactionID, practitionerID, req)
		So(err, ShouldBeNil)
		So(appointed, ShouldBeFalse)
	})
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

func TestUnitCheckAppointmentDateValid(t *testing.T) {
	transactionID := "123"
	apiURL := "https://api.companieshouse.gov.uk"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	Convey("error getting practitioners", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(models.InsolvencyResourceDao{}, fmt.Errorf("err"))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		err, validDate := CheckAppointmentDateValid(mockService, transactionID, "2012-01-23", req)
		So(err.Error(), ShouldEqual, "err")
		So(validDate, ShouldBeFalse)
	})

	Convey("error getting company details", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusTeapot, ""))

		insolvencyResponse := generateInsolvencyResource()

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(insolvencyResponse, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		err, validDate := CheckAppointmentDateValid(mockService, transactionID, "2012-01-23", req)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, `error communicating with the company profile api`)
		So(validDate, ShouldBeFalse)
	})

	Convey("error parsing incorporatedOn date", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-2600:00:00.000Z")))

		insolvencyResponse := generateInsolvencyResource()

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(insolvencyResponse, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		err, validDate := CheckAppointmentDateValid(mockService, transactionID, "2012-01-23", req)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, `error parsing date`)
		So(validDate, ShouldBeFalse)
	})

	Convey("error parsing appointedOn date", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2013-06-26 00:00:00.000Z")))

		insolvencyResponse := generateInsolvencyResource()

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(insolvencyResponse, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		err, validDate := CheckAppointmentDateValid(mockService, transactionID, "2012-01-2300:00:00.000Z", req)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, `error parsing date`)
		So(validDate, ShouldBeFalse)
	})

	Convey("invalid date - date supplied is before company incorporation date", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2013-06-26 00:00:00.000Z")))

		insolvencyResponse := generateInsolvencyResource()

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(insolvencyResponse, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		err, validDate := CheckAppointmentDateValid(mockService, transactionID, "2012-01-23", req)
		So(err, ShouldBeNil)
		So(validDate, ShouldBeFalse)
	})

	Convey("invalid date - date supplied is in the future", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26")))

		insolvencyResponse := generateInsolvencyResource()
		appointedOn := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(insolvencyResponse, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		err, validDate := CheckAppointmentDateValid(mockService, transactionID, appointedOn, req)
		So(err, ShouldBeNil)
		So(validDate, ShouldBeFalse)
	})

	Convey("invalid date - appointment date supplied is different from appointment date of already appointed practitioner", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		insolvencyResponse := generateInsolvencyResource()
		appointment := models.AppointmentResourceDao{
			AppointedOn: "2013-01-23",
			MadeBy:      "creditors",
			Links:       models.AppointmentResourceLinksDao{},
		}
		insolvencyResponse.Data.Practitioners[0].Appointment = &appointment

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(insolvencyResponse, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		err, validDate := CheckAppointmentDateValid(mockService, transactionID, "2012-01-23", req)
		So(err, ShouldBeNil)
		So(validDate, ShouldBeFalse)
	})

	Convey("valid date", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		defer httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, apiURL+"/company/1234", httpmock.NewStringResponder(http.StatusOK, companyProfileDateResponse("2000-06-26 00:00:00.000Z")))

		insolvencyResponse := generateInsolvencyResource()
		appointment := models.AppointmentResourceDao{
			AppointedOn: "2012-01-23",
			MadeBy:      "creditors",
			Links:       models.AppointmentResourceLinksDao{},
		}
		insolvencyResponse.Data.Practitioners[0].Appointment = &appointment

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetInsolvencyResource(gomock.Any()).Return(insolvencyResponse, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		err, validDate := CheckAppointmentDateValid(mockService, transactionID, "2012-01-23", req)
		So(err, ShouldBeNil)
		So(validDate, ShouldBeTrue)
	})
}

func generateInsolvencyResource() models.InsolvencyResourceDao {
	return models.InsolvencyResourceDao{
		Data: models.InsolvencyResourceDaoData{
			CompanyNumber: "1234",
			CaseType:      "CVL",
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
