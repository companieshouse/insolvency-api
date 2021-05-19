package constants

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitIsAttachmentTypeValid(t *testing.T) {
	Convey("attachment type is valid", t, func() {
		ok := IsAttachmentTypeValid("resolution")
		So(ok, ShouldBeTrue)
	})

	Convey("attachment type is invalid", t, func() {
		ok := IsAttachmentTypeValid("invalid")
		So(ok, ShouldBeFalse)
	})
}

func TestUnitAttachmentTypeString(t *testing.T) {
	Convey("provide a string for attachmentType", t, func() {
		So(Resolution.String(), ShouldEqual, "resolution")
		So(StatementOfAffairsLiquidator.String(), ShouldEqual, "statement-of-affairs-liquidator")
		So(StatementOfAffairsDirector.String(), ShouldEqual, "statement-of-affairs-director")
		So(StatementOfConcurrence.String(), ShouldEqual, "statement-of-concurrence")
	})
}
