package constants

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitCaseTypesString(t *testing.T) {
	Convey("provide a string for appointment made by", t, func() {
		So(CVL.String(), ShouldEqual, "creditors-voluntary-liquidation")
		So(MVL.String(), ShouldEqual, "members-voluntary-liquidation")
	})
}
