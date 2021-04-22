package constants

// AppointmentMadeBy Enum Type
type AppointmentMadeBy int

// Enumeration containing all permitted values for AppointedBy
const (
	Company AppointmentMadeBy = 1 + iota
	Creditors
)

var appointmentMadeByTypes = [...]string{
	"company",
	"creditors",
}

// String returns the correctly formatted AppointmentMadeBy
func (appointmentMadeBy AppointmentMadeBy) String() string {
	return appointmentMadeByTypes[appointmentMadeBy-1]
}

// IsAppointmentMadeByInList checks if the appointmentMadeBy string supplied
// is a valid string by comparing it to the list of accepted values
func IsAppointmentMadeByInList(appointmentMadeBy string) bool {
	for _, v := range appointmentMadeByTypes {
		if appointmentMadeBy == v {
			return true
		}
	}
	return false
}
