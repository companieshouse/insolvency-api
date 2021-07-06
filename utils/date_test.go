package utils

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitIsValidDate(t *testing.T) {
	Convey("error parsing date", t, func() {
		date := "20-20-20"
		incorporatedOn := "2020-06-06"

		isDateValid, err := IsValidDate(date, incorporatedOn)
		So(isDateValid, ShouldEqual, false)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, fmt.Sprintf("parsing time \"%s\" as \"2006-01-02\": cannot parse", date))
	})

	Convey("error parsing incorporatedOn date", t, func() {
		date := "2020-06-20"
		incorporatedOn := "20-20-20"

		isDateValid, err := IsValidDate(date, incorporatedOn)
		So(isDateValid, ShouldEqual, false)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, fmt.Sprintf("parsing time \"%s\" as \"2006-01-02\": cannot parse", incorporatedOn))
	})

	Convey("date is in the future", t, func() {
		date := "9999-09-09"
		incorporatedOn := "2020-06-06"

		isDateValid, err := IsValidDate(date, incorporatedOn)
		So(isDateValid, ShouldEqual, false)
		So(err, ShouldEqual, nil)
	})

	Convey("date is before incorporation date", t, func() {
		date := "2020-06-05"
		incorporatedOn := "2020-06-06"

		isDateValid, err := IsValidDate(date, incorporatedOn)
		So(isDateValid, ShouldEqual, false)
		So(err, ShouldEqual, nil)
	})

	Convey("valid date", t, func() {
		date := "2021-06-06"
		incorporatedOn := "2020-06-06"

		isDateValid, err := IsValidDate(date, incorporatedOn)
		So(isDateValid, ShouldEqual, true)
		So(err, ShouldEqual, nil)
	})
}
