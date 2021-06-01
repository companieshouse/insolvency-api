package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/companieshouse/insolvency-api/models"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitWriteJSONWithStatus(t *testing.T) {
	Convey("Failure to marshal json", t, func() {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		// causes an UnsupportedTypeError
		WriteJSONWithStatus(w, r, make(chan int), http.StatusInternalServerError)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")
		So(w.Body.String(), ShouldEqual, "")
	})

	Convey("contents are written as json", t, func() {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		WriteJSONWithStatus(w, r, &models.CreatedInsolvencyResourceLinks{}, http.StatusCreated)

		So(w.Code, ShouldEqual, http.StatusCreated)
		So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")
		So(w.Body.String(), ShouldEqual, "{\"self\":\"\",\"transaction\":\"\",\"validation_status\":\"\"}\n")
	})
}

func TestUnitGetTransactionIDFromVars(t *testing.T) {
	Convey("Get Transaction ID", t, func() {
		vars := map[string]string{
			"transaction_id": "12345",
		}
		transactionID := GetTransactionIDFromVars(vars)
		So(transactionID, ShouldEqual, "12345")
	})

	Convey("No Transaction ID", t, func() {
		vars := map[string]string{}
		transactionID := GetTransactionIDFromVars(vars)
		So(transactionID, ShouldBeEmpty)
	})
}

func TestUnitGetPractitionerIDFromVars(t *testing.T) {
	Convey("Get Practitioner ID", t, func() {
		vars := map[string]string{
			"practitioner_id": "67890",
		}
		PractitionerID := GetPractitionerIDFromVars(vars)
		So(PractitionerID, ShouldEqual, "67890")
	})

	Convey("No Practitioner ID", t, func() {
		vars := map[string]string{}
		PractitionerID := GetPractitionerIDFromVars(vars)
		So(PractitionerID, ShouldBeEmpty)
	})
}

func TestUnitGetAttachmentIDFromVars(t *testing.T) {
	Convey("Get Attachment ID", t, func() {
		vars := map[string]string{
			"attachment_id": "67890",
		}
		PractitionerID := GetAttachmentIDFromVars(vars)
		So(PractitionerID, ShouldEqual, "67890")
	})

	Convey("No Attachment ID", t, func() {
		vars := map[string]string{}
		PractitionerID := GetAttachmentIDFromVars(vars)
		So(PractitionerID, ShouldBeEmpty)
	})
}

func TestUnitResponseTypeToStatus(t *testing.T) {
	Convey("Response Type to Status", t, func() {
		r, err := ResponseTypeToStatus("invalid-data")
		So(r, ShouldEqual, http.StatusBadRequest)
		So(err, ShouldBeNil)

		r, err = ResponseTypeToStatus("error")
		So(r, ShouldEqual, http.StatusInternalServerError)
		So(err, ShouldBeNil)

		r, err = ResponseTypeToStatus("success")
		So(r, ShouldEqual, http.StatusOK)
		So(err, ShouldBeNil)

		r, err = ResponseTypeToStatus("default")
		So(r, ShouldEqual, 0)
		So(err.Error(), ShouldEqual, "response type not recognised")
	})
}
