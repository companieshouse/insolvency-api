package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	mock_dao "github.com/companieshouse/insolvency-api/mocks"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitRegisterRoutes(t *testing.T) {
	Convey("Register routes", t, func() {
		router := mux.NewRouter()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockService := mock_dao.NewMockService(mockCtrl)
		Register(router, mockService)

		So(router.GetRoute("healthcheck"), ShouldNotBeNil)

		So(router.GetRoute("createInsolvencyResource"), ShouldNotBeNil)
		So(router.GetRoute("getValidationStatus"), ShouldNotBeNil)
		So(router.GetRoute("getFilings"), ShouldNotBeNil)

		So(router.GetRoute("createPractitionersResource"), ShouldNotBeNil)
		So(router.GetRoute("getPractitionerResources"), ShouldNotBeNil)
		So(router.GetRoute("getPractitionerResource"), ShouldNotBeNil)
		So(router.GetRoute("deletePractitioner"), ShouldNotBeNil)

		So(router.GetRoute("appointPractitioner"), ShouldNotBeNil)
		So(router.GetRoute("getPractitionerAppointment"), ShouldNotBeNil)
		So(router.GetRoute("deletePractitionerAppointment"), ShouldNotBeNil)

		So(router.GetRoute("submitAttachment"), ShouldNotBeNil)
		So(router.GetRoute("getAttachmentDetails"), ShouldNotBeNil)
		So(router.GetRoute("downloadAttachment"), ShouldNotBeNil)
		So(router.GetRoute("deleteAttachment"), ShouldNotBeNil)

		So(router.GetRoute("createResolution"), ShouldNotBeNil)
		So(router.GetRoute("getResolution"), ShouldNotBeNil)
		So(router.GetRoute("deleteResolution"), ShouldNotBeNil)
	})
}

func TestUnitHealthCheck(t *testing.T) {
	Convey("Healthcheck", t, func() {
		w := httptest.ResponseRecorder{}
		healthCheck(&w, nil)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}
