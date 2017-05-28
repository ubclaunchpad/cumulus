package message

import (
	"encoding/gob"
	"io"
)

type (
	// Type specifies the type of a message.
	Type int
	// ResourceType specifies the type of a resource in a message.
	ResourceType int
	// ErrorCode is a code associated with an error
	ErrorCode int
)

const (
	// MessageRequest messages ask a peer for a resource.
	MessageRequest Type = iota
	// MessageResponse messages repond to a request message with an error or a resource.
	MessageResponse
	// MessagePush messages proactively send a resource to a peer.
	MessagePush
)

const (
	// ResourcePeerInfo resources contain a list of peers.
	ResourcePeerInfo ResourceType = iota
	// ResourceBlock resources contain a block in the blockchain.
	ResourceBlock
	// ResourceTransaction resources contain a transaction to add to the blockchain.
	ResourceTransaction
)

const (
	// InvalidResourceType occurs when a request is received with an unknown
	// ResourceType value.
	InvalidResourceType = 401
	// NotImplemented occurs when a message or request is received whos response
	// requires functionality that does not yet exist.
	NotImplemented = 501
	// SubnetFull occurs when a stream is opened with a peer who's Subnet is
	// already full.
	SubnetFull = 503
)

// Initializes all the types we need to encode.
func init() {
	gob.Register(&Request{})
	gob.Register(&Response{})
	gob.Register(&Push{})
}

// Error is an error that occured during a request.
type Error struct {
	Code    ErrorCode
	Message string
}

// NewError returns a new error struct.
func NewError(code ErrorCode, msg string) *Error {
	return &Error{code, msg}
}

// Error returns the error message; to implement `error`.
func (e *Error) Error() string {
	return e.Message
}

// Message is a container for messages, containing a type and either a Request,
// Response, or Push in the payload.
type Message interface {
	Type() Type
	Write(io.Writer) error
}

// Request is a container for a request payload, containing a unique request ID,
// the resource type we are requesting, and a Params field specifying request
// parameters. PeerInfo requests should send all info of all peers. Block requests
// should specify block number in parameters.
type Request struct {
	ID           string
	ResourceType ResourceType
	Params       map[string]interface{}
}

// Type returns the message type
func (r *Request) Type() Type {
	return MessageRequest
}

// Response is a container for a response payload, containing the unique request
// ID of the request prompting it, an Error (if one occurred), and the requested
// resource (if no error occurred).
type Response struct {
	ID       string
	Error    *Error
	Resource interface{}
}

// Type returns the message type
func (r *Response) Type() Type {
	return MessageResponse
}

// Push is a container for a push payload, containing a resource proactively sent
// to us by another peer.
type Push struct {
	ResourceType ResourceType
	Resource     interface{}
}

// Type returns the message type
func (p *Push) Type() Type {
	return MessagePush
}

// Write encodes and writes the Message into the given Writer.
func (r *Request) Write(w io.Writer) error {
	var m Message = r
	return gob.NewEncoder(w).Encode(&m)
}

func (r *Response) Write(w io.Writer) error {
	var m Message = r
	return gob.NewEncoder(w).Encode(&m)
}

func (p *Push) Write(w io.Writer) error {
	var m Message = p
	return gob.NewEncoder(w).Encode(&m)
}

// Read decodes a message from a Reader and returns it.
func Read(r io.Reader) (Message, error) {
	var m Message
	err := gob.NewDecoder(r).Decode(&m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
