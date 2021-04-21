package constants

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitIsValidAppointmentMadeBy(t *testing.T) {
	Convey("appointment made by supplied is valid", t, func() {
		ok := IsAppointmentMadeByInList("company")
		So(ok, ShouldBeTrue)
	})

	Convey("appointment made by supplied is invalid", t, func() {
		ok := IsAppointmentMadeByInList("invalid")
		So(ok, ShouldBeFalse)
	})
}

func TestUnitAppointmentMadeByString(t *testing.T) {
	Convey("provide a string for appointment made by", t, func() {
		So(Company.String(), ShouldEqual, "company")
		So(Creditors.String(), ShouldEqual, "creditors")
	})
}
