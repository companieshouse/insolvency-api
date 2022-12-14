package constants

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitIsValidAttachmentTypes(t *testing.T) {
	Convey("List of each permitted attachment type is valid", t, func() {
		tables := []struct {
			AttachmentType string
		}{
			{"resolution"},
			{"statement-of-affairs-liquidator"},
			{"statement-of-affairs-director"},
			{"statement-of-concurrence"},
			{"progress-report"},
		}

		for _, table := range tables {
			ok := IsAttachmentTypeValid(table.AttachmentType)
			So(ok, ShouldBeTrue)
		}
	})
}

func TestUnitIsAttachmentTypeIsInvalid(t *testing.T) {
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
		So(ProgressReport.String(), ShouldEqual, "progress-report")
	})
}
