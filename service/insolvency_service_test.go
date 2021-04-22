package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestUnitCheckAppointmentDateValid(t *testing.T) {
	transactionID := "123"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	Convey("error getting practitioners", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetPractitionerResources(gomock.Any()).Return(nil, fmt.Errorf("err"))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		err, validDate := CheckAppointmentDateValid(mockService, transactionID, "2012-01-23", req)
		So(err.Error(), ShouldEqual, "err")
		So(validDate, ShouldBeFalse)
	})

	Convey("invalid date", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		practitionersResponse := []models.PractitionerResourceDao{
			{
				ID: "456",
				Appointment: &models.AppointmentResourceDao{
					AppointedOn: "2012-02-23",
				},
			},
		}

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetPractitionerResources(gomock.Any()).Return(practitionersResponse, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		err, validDate := CheckAppointmentDateValid(mockService, transactionID, "2012-01-23", req)
		So(err, ShouldBeNil)
		So(validDate, ShouldBeFalse)
	})

	Convey("valid date", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		practitionersResponse := []models.PractitionerResourceDao{
			{
				ID: "456",
				Appointment: &models.AppointmentResourceDao{
					AppointedOn: "2012-01-23",
				},
			},
		}

		mockService := mocks.NewMockService(mockCtrl)
		mockService.EXPECT().GetPractitionerResources(gomock.Any()).Return(practitionersResponse, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		err, validDate := CheckAppointmentDateValid(mockService, transactionID, "2012-01-23", req)
		So(err, ShouldBeNil)
		So(validDate, ShouldBeTrue)
	})
}
