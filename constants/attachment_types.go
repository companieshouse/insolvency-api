package constants

// AttachmentType Enum Type
type AttachmentType int

// Enumeration containing all permitted values for attachment
const (
	Resolution AttachmentType = 1 + iota
	StatementOfAffairsLiquidator
	StatementOfAffairsDirector
	StatementOfConcurrence
)

var attachmentTypes = [...]string{
	"resolution",
	"statement-of-affairs-liquidator",
	"statement-of-affairs-director",
	"statement-of-concurrence",
}

// String returns the correctly formatted AttachmentType
func (attachmentType AttachmentType) String() string {
	return attachmentTypes[attachmentType-1]
}

// IsAttachmentTypeValid checks if the attachmentType string supplied
// is a valid string by comparing it to the list of accepted values
func IsAttachmentTypeValid(attachmentType string) bool {
	for _, v := range attachmentTypes {
		if attachmentType == v {
			return true
		}
	}

	return false
}
