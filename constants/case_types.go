package constants

// CaseType Enum Type
type CaseType int

// Enumeration containing all possible case types
const (
	CVL CaseType = 1 + iota
	MVL
)

// String representation of case types
var caseTypes = [...]string{
	"creditors-voluntary-liquidation",
	"members-voluntary-liquidation",
}

func (caseType CaseType) String() string {
	return caseTypes[caseType-1]
}
