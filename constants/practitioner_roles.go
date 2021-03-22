package constants

// PractitionerRole Enum Type
type PractitionerRole int

// Enumeration containing all possible practitioner roles
const (
	FinalLiquidator PractitionerRole = 1 + iota
	Receiver
	ReceiverManager
	ProposedLiquidator
	ProvisionalLiquidator
	AdministrativeReceiver
	Practitioner
	InterimLiquidator
)

var practitionerRoles = [...]string{
	"final-liquidator",
	"receiver",
	"receiver-manager",
	"proposed-liquidator",
	"provisional-liquidator",
	"administrative-receiver",
	"practitioner",
	"interim-liquidator",
}

// String returns the correctly formatted practitioner role
func (practitionerRole PractitionerRole) String() string {
	return practitionerRoles[practitionerRole-1]
}

// IsInRoleList checks if the roleName string supplied is a valid string by comparing
// it to the list of accepted practitioner roles
func IsInRoleList(roleName string) bool {
	for _, v := range practitionerRoles {
		if roleName == v {
			return true
		}
	}
	return false
}
