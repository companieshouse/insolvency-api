package service

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitValidateAttachmentDetails(t *testing.T) {
	Convey("Valid attachment details", t, func() {
		errs := ValidateAttachmentDetails("resolution")
		So(errs, ShouldBeEmpty)
	})

	Convey("Invalid attachment details", t, func() {
		errs := ValidateAttachmentDetails("invalid")
		So(errs, ShouldEqual, "attachment_type is invalid")
	})
}
