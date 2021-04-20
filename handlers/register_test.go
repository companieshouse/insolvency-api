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
		So(router.GetRoute("createPractitionersResource"), ShouldNotBeNil)
		So(router.GetRoute("getPractitionerResources"), ShouldNotBeNil)
		So(router.GetRoute("deletePractitioner"), ShouldNotBeNil)
	})
}

func TestUnitHealthCheck(t *testing.T) {
	Convey("Healthcheck", t, func() {
		w := httptest.ResponseRecorder{}
		healthCheck(&w, nil)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}
