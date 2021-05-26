package service

import (
	"mime/multipart"
	"net/textproto"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitValidateAttachmentDetails(t *testing.T) {

	Convey("Invalid attachment details - invalid attachment type", t, func() {
		validationErrs := ValidateAttachmentDetails("invalid", createHeader())
		So(validationErrs, ShouldEqual, "attachment_type is invalid")
	})

	Convey("Invalid attachment details - invalid file format in header and name", t, func() {
		header := createHeader()
		header.Header.Set("Content-Type", "invalid")
		header.Filename = "test.txt"
		validationErrs := ValidateAttachmentDetails("resolution", header)
		So(validationErrs, ShouldEqual, "attachment file format should be pdf")
	})

	Convey("Invalid attachment details - invalid file format in header and name", t, func() {
		header := createHeader()
		header.Size = 10000000000
		validationErrs := ValidateAttachmentDetails("resolution", header)
		So(validationErrs, ShouldEqual, "attachment file size is too large to be processed")
	})

	Convey("Valid attachment details", t, func() {
		validationErrs := ValidateAttachmentDetails("resolution", createHeader())
		So(validationErrs, ShouldBeEmpty)
	})

}

func createHeader() *multipart.FileHeader {
	return &multipart.FileHeader{
		Filename: "test.pdf",
		Header: textproto.MIMEHeader{
			"Content-Type": []string{
				"application/pdf",
			},
		},
		Size: 1,
	}
}
