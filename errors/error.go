package errors

import "fmt"

const (
	// InvalidResourceType occurs when a request is received with an unknown
	// ResourceType value.
	InvalidResourceType = 401
	// SubnetFull occurs when a stream is opened with a peer who's Subnet is
	// already full.
	SubnetFull = 501
	// NotImplemented occurs when a message or request is received whos response
	// requires functionality that does not yet exist.
	NotImplemented = 502
)

const (
	invalidResourceTypeMsg = "Invalid resource type"
	subnetFullMsg          = "Failed to add peer to subnet (peer subnet full)"
	notImplementedMsg      = "Functionality not yet implemented"
)

// ProtocolError is a container for error messages returned in Peer
// Response bodies.
type ProtocolError struct {
	Code    int
	Message string
}

// New returns new ProtocolError with the given parameters.
// code argument should be one of the error codes defined above.
func New(code int) *ProtocolError {
	var msg string

	switch code {
	case SubnetFull:
		msg = subnetFullMsg
		break
	case InvalidResourceType:
		msg = invalidResourceTypeMsg
		break
	case NotImplemented:
		msg = notImplementedMsg
		break
	default:
		panic("Attempt to create error with invalid error code")
	}

	return &ProtocolError{
		Code:    code,
		Message: msg,
	}
}

// Error returns the string representation of the given ProtocolError
func (e ProtocolError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}
