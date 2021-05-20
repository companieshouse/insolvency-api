package service

// ResponseType enumerates the types of authentication supported
type ResponseType int

const (
	// InvalidData response
	InvalidData ResponseType = iota

	// Error response
	Error

	// Forbidden response
	Forbidden

	// Success response
	Success
)

var vals = [...]string{
	"invalid-data",
	"error",
	"forbidden",
	"success",
}

// String representation of `ResponseType`
func (a ResponseType) String() string {
	return vals[a]
}
