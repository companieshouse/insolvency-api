package utils

import (
	"regexp"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerateID(t *testing.T) {
	Convey("test generation of an ID", t, func() {
		id := GenerateID()

		match, _ := regexp.MatchString("[A-Z]{2}[0-9]{8}", id)

		So(id, ShouldNotBeNil)
		So(id, ShouldHaveLength, 10)
		So(match, ShouldBeTrue)
	})
}
