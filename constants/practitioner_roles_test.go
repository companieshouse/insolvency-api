package constants

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIsValidRole(t *testing.T) {
	Convey("role supplied is valid", t, func() {
		practitionerRole := "final-liquidator"

		ok := IsInRoleList(practitionerRole)

		So(ok, ShouldBeTrue)
	})

	Convey("role supplied is invalid", t, func() {
		practitionerRole := "error-role"

		ok := IsInRoleList(practitionerRole)

		So(ok, ShouldBeFalse)
	})
}

func TestString(t *testing.T) {
	Convey("provide a string for practitioner role", t, func() {
		practitionerRole := ReceiverManager.String()

		So(practitionerRole, ShouldEqual, "receiver-manager")
	})
}
