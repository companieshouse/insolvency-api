package utils

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitIsValidDate(t *testing.T) {
	Convey("error parsing date", t, func() {
		date := "20-20-20"

		_, err := ValidateDate(date)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, fmt.Sprintf("parsing time \"%s\" as \"2006-01-02\": cannot parse", date))
	})

	Convey("date validated successfully without time", t, func() {
		date := "2020-01-20"

		isDateValid, err := ValidateDate(date)
		So(isDateValid, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})

	Convey("date validated successfully with time", t, func() {
		date := "2020-01-20 00:00:00.000Z"

		isDateValid, err := ValidateDate(date)
		So(isDateValid, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})

	Convey("date is in the future", t, func() {
		date := "2999-01-20"

		isDateValid, _ := IsDateInFuture(date)
		So(isDateValid, ShouldBeTrue)
	})

	Convey("date is not in the future", t, func() {
		date := "2000-01-20"

		isDateValid, _ := IsDateInFuture(date)
		So(isDateValid, ShouldBeFalse)
	})

	Convey("invalid date in the future", t, func() {
		date := "20-01-20"

		isDateValid, err := IsDateInFuture(date)
		So(isDateValid, ShouldBeFalse)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, fmt.Sprintf("parsing time \"%s\" as \"2006-01-02\": cannot parse", date))
	})

	Convey("later date is before early date", t, func() {
		beforeDate := "2000-01-20"
		afterDate := "2000-01-19"

		isDateValid, _ := IsDateBeforeDate(beforeDate, afterDate)
		So(isDateValid, ShouldBeFalse)
	})

	Convey("early date is before later date", t, func() {
		beforeDate := "2000-01-19"
		afterDate := "2000-01-20"

		isDateValid, _ := IsDateBeforeDate(beforeDate, afterDate)
		So(isDateValid, ShouldBeTrue)
	})

	Convey("early date is invalid format", t, func() {
		beforeDate := "20-01-19"
		afterDate := "2000-01-20"

		isDateValid, err := IsDateBeforeDate(beforeDate, afterDate)
		So(isDateValid, ShouldBeFalse)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, fmt.Sprintf("parsing time \"%s\" as \"2006-01-02\": cannot parse", beforeDate))
	})

	Convey("later date is invalid format", t, func() {
		beforeDate := "2000-01-19"
		afterDate := "20-01-20"

		isDateValid, err := IsDateBeforeDate(beforeDate, afterDate)
		So(isDateValid, ShouldBeFalse)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, fmt.Sprintf("parsing time \"%s\" as \"2006-01-02\": cannot parse", afterDate))
	})

	Convey("error parsing date", t, func() {
		date := "20-20-20"
		incorporatedOn := "2020-06-06"

		isDateValid, err := IsDateBetweenIncorporationAndNow(date, incorporatedOn)
		So(isDateValid, ShouldEqual, false)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, fmt.Sprintf("parsing time \"%s\" as \"2006-01-02\": cannot parse", date))
	})

	Convey("error parsing incorporatedOn date", t, func() {
		date := "2020-06-20"
		incorporatedOn := "20-20-20"

		isDateValid, err := IsDateBetweenIncorporationAndNow(date, incorporatedOn)
		So(isDateValid, ShouldEqual, false)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, fmt.Sprintf("parsing time \"%s\" as \"2006-01-02\": cannot parse", incorporatedOn))
	})

	Convey("date is in the future", t, func() {
		date := "9999-09-09"
		incorporatedOn := "2020-06-06"

		isDateValid, err := IsDateBetweenIncorporationAndNow(date, incorporatedOn)
		So(isDateValid, ShouldEqual, false)
		So(err, ShouldEqual, nil)
	})

	Convey("date is before incorporation date", t, func() {
		date := "2020-06-05"
		incorporatedOn := "2020-06-06"

		isDateValid, err := IsDateBetweenIncorporationAndNow(date, incorporatedOn)
		So(isDateValid, ShouldEqual, false)
		So(err, ShouldEqual, nil)
	})

	Convey("Date is between Incorporation date and Now", t, func() {
		date := "2021-06-06"
		incorporatedOn := "2020-06-06"

		isDateValid, err := IsDateBetweenIncorporationAndNow(date, incorporatedOn)
		So(isDateValid, ShouldEqual, true)
		So(err, ShouldEqual, nil)
	})
}
