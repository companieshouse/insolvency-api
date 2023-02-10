package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitConverter(t *testing.T) {
	jsonPractitionersDao := `{"VM04221441":"/transactions/168570-809316-704268/insolvency/practitioners/VM04221441"}`

	stringMapArray := map[string]string{"VM04221441": "/transactions/168570-809316-704268/insolvency/practitioners/VM04221441"}

	Convey("ConvertStringToMap returns required objects", t, func() {

		mapObject, arrayString, err := ConvertStringToMap(jsonPractitionersDao)

		So(mapObject, ShouldNotBeNil)
		So(arrayString, ShouldNotBeNil)
		So(len(arrayString), ShouldEqual, 1)
		So(err, ShouldBeNil)
	})

	Convey("ConvertMapToString returns required objects", t, func() {

		arrayString, err := ConvertMapToString(stringMapArray)

		So(arrayString, ShouldNotBeNil)
		So(arrayString, ShouldEqual, jsonPractitionersDao)
		So(err, ShouldBeNil)
	})

	Convey("ConvertMapToStringArray returns required objects", t, func() {

		arrayString := ConvertMapToStringArray(stringMapArray)

		So(arrayString, ShouldNotBeNil)
		So(arrayString, ShouldResemble, []string{"VM04221441"})
	 
	})

}
