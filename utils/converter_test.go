package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitConverter(t *testing.T) {
	jsonPractitionersDao := `{"VM04221441":"/transactions/168570-809316-704268/insolvency/practitioners/VM04221441"}`

	stringMap := map[string]string{"VM04221441": "/transactions/168570-809316-704268/insolvency/practitioners/VM04221441"}

	Convey("GetMapKeysAsStringSlice returns required objects", t, func() {

		stringSlice := GetMapKeysAsStringSlice(stringMap)

		So(stringSlice, ShouldNotBeNil)
		So(stringSlice, ShouldResemble, []string{"VM04221441"})

	})

	Convey("CheckStringContainsElement returns true when required element is found", t, func() {

		booleanResult := CheckStringContainsElement(jsonPractitionersDao, "/", "168570-809316-704268")

		So(booleanResult, ShouldBeTrue)

	})

	Convey("CheckStringContainsElement returns false when required element is not found", t, func() {

		booleanResult := CheckStringContainsElement(jsonPractitionersDao, "/", "168570-809316-7042622")

		So(booleanResult, ShouldBeFalse)

	})

}
