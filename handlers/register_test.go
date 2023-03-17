package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/companieshouse/insolvency-api/config"
	mock_dao "github.com/companieshouse/insolvency-api/mocks"
	"github.com/companieshouse/insolvency-api/utils"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func setupTestRouter(t *testing.T) *mux.Router {
	router := mux.NewRouter()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockService := mock_dao.NewMockService(mockCtrl)
	helperService := utils.NewHelperService()
	Register(router, mockService, helperService)
	return router
}

func TestUnitRegisterRoutes(t *testing.T) {
	// Certain routes are now disabled when ENABLE_NON_LIVE_ROUTE_HANDLERS is unset or false
	cfg, _ := config.Get()

	Convey("Register routes: ENABLE_NON_LIVE_ROUTE_HANDLERS unset or false", t, func() {
		router := setupTestRouter(t)

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

		So(router.GetRoute("createStatementOfAffairs"), ShouldNotBeNil)
		So(router.GetRoute("getStatementOfAffairs"), ShouldNotBeNil)
		So(router.GetRoute("deleteStatementOfAffairs"), ShouldNotBeNil)

		So(router.GetRoute("createProgressReport"), ShouldNotBeNil)
		So(router.GetRoute("getProgressReport"), ShouldNotBeNil)
	})

	// Simulate ENABLE_NON_LIVE_ROUTE_HANDLERS feature toggle being enabled
	cfg.EnableNonLiveRouteHandlers = true
	Convey("Register routes: ENABLE_NON_LIVE_ROUTE_HANDLERS is set as true", t, func() {
		router := setupTestRouter(t)

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

		So(router.GetRoute("createStatementOfAffairs"), ShouldNotBeNil)
		So(router.GetRoute("getStatementOfAffairs"), ShouldNotBeNil)
		So(router.GetRoute("deleteStatementOfAffairs"), ShouldNotBeNil)

		So(router.GetRoute("createResolution"), ShouldNotBeNil)
		So(router.GetRoute("getResolution"), ShouldNotBeNil)
		So(router.GetRoute("deleteResolution"), ShouldNotBeNil)

		So(router.GetRoute("createStatementOfAffairs"), ShouldNotBeNil)
		So(router.GetRoute("getStatementOfAffairs"), ShouldNotBeNil)
		So(router.GetRoute("deleteStatementOfAffairs"), ShouldNotBeNil)

		So(router.GetRoute("createProgressReport"), ShouldNotBeNil)
		So(router.GetRoute("getProgressReport"), ShouldNotBeNil)
	})
}

func TestUnitHealthCheck(t *testing.T) {
	Convey("Healthcheck", t, func() {
		w := httptest.ResponseRecorder{}
		healthCheck(&w, nil)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestUnitUrlParameterValidation(t *testing.T) {
	router := setupTestRouter(t)
	matchInfo := &mux.RouteMatch{}
	Convey("Invalid Transaction ID matched to route", t, func() {
		req, _ := http.NewRequest("GET", "/transactions/x/insolvency/validation-status", nil)
		matched := router.Match(req, matchInfo)
		So(matched, ShouldEqual, false)
		So(matchInfo.MatchErr, ShouldEqual, mux.ErrNotFound)
	})
	Convey("Valid Transaction ID matched to route", t, func() {
		req, _ := http.NewRequest("GET", "/transactions/068925-439616-777491/insolvency/validation-status", nil)
		matched := router.Match(req, matchInfo)
		So(matched, ShouldEqual, true)
		So(matchInfo.MatchErr, ShouldBeNil)
	})
	Convey("Invalid Attachment ID matched to route", t, func() {
		req, _ := http.NewRequest("GET", "/transactions/068925-439616-777491/insolvency/attachments/x", nil)
		matched := router.Match(req, matchInfo)
		So(matched, ShouldEqual, false)
		So(matchInfo.MatchErr, ShouldEqual, mux.ErrNotFound)
	})
	Convey("Valid Attachment ID matched to route", t, func() {
		req, _ := http.NewRequest("GET", "/transactions/068925-439616-777491/insolvency/attachments/36bd074f-96a8-46d7-ad6b-29dd67452f0a", nil)
		matched := router.Match(req, matchInfo)
		So(matched, ShouldEqual, true)
		So(matchInfo.MatchErr, ShouldBeNil)
	})
}
