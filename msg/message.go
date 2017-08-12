package msg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type (
	// ResourceType specifies the type of a resource in a message.
	ResourceType int
	// ErrorCode is a code associated with an error
	ErrorCode int
)

const (
	// PushMessage fills the Type field in the Message struct for Push messages
	PushMessage = "Push"
	// RequestMessage fills the Type field in the Message struct for Request messages
	RequestMessage = "Request"
	// ResponseMessage fills the Type field in the Message struct for Response messages
	ResponseMessage = "Response"
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
	// BadRequest occurs when a malformatted request is received
	BadRequest = 400
	// InvalidResourceType occurs when a request is received with an unknown
	// ResourceType value.
	InvalidResourceType = 401
	// RequestTimeout occurs when a peer does not respond to a request within
	// some predefined period of time (see peer.DefaultRequestTimeout)
	RequestTimeout = 408
	// ResourceNotFound occurs when a node reports the requested resource missing.
	ResourceNotFound = 404
	// NotImplemented occurs when a message or request is received whos response
	// requires functionality that does not yet exist.
	NotImplemented = 501
)

// ProtocolError is an error that occured during a request.
type ProtocolError struct {
	Code    ErrorCode
	Message string
}

// NewProtocolError returns a new error struct.
func NewProtocolError(code ErrorCode, msg string) *ProtocolError {
	return &ProtocolError{code, msg}
}

// Error returns the error message; to implement `error`.
func (e *ProtocolError) Error() string {
	return e.Message
}

// Message is a wrapper for requests, responses, and pushes.
// Type must be one of a "Request", "Response", or "Push"
// Payload must be a marshalled representation of a Request, Response, or Push
// when the message is sent.
type Message struct {
	Type    string
	Payload []byte
}

// MessagePayload is an interface that is implemented by Request, Response, and
// Push. It is used to generally refer to these 3 payload types so we can
// return only a single value from Read().
type MessagePayload interface {
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

// Response is a container for a response payload, containing the unique request
// ID of the request prompting it, an Error (if one occurred), and the requested
// resource (if no error occurred).
type Response struct {
	ID       string
	Error    *ProtocolError
	Resource interface{}
}

// Push is a container for a push payload, containing a resource proactively sent
// to us by another peer.
type Push struct {
	ResourceType ResourceType
	Resource     interface{}
}

// Write encodes and writes the Message into the given Writer.
func (r *Request) Write(w io.Writer) error {
	payload, err := json.Marshal(r)
	if err != nil {
		return err
	}
	msg := Message{
		Type:    RequestMessage,
		Payload: payload,
	}
	return json.NewEncoder(w).Encode(msg)
}

func (r *Response) Write(w io.Writer) error {
	payload, err := json.Marshal(r)
	if err != nil {
		return err
	}
	msg := Message{
		Type:    ResponseMessage,
		Payload: payload,
	}
	return json.NewEncoder(w).Encode(msg)
}

func (p *Push) Write(w io.Writer) error {
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}
	msg := Message{
		Type:    PushMessage,
		Payload: payload,
	}
	return json.NewEncoder(w).Encode(msg)
}

// Read decodes a message from a Reader and returns the message payload, or an
// error if the read fails. On success, the payload returned will be either a
// Request, Response, or Push. Resource fields will contain the appropriate
// type chosen by the json.Decode() function.
func Read(r io.Reader) (MessagePayload, error) {
	var m Message
	err := json.NewDecoder(r).Decode(&m)
	if err != nil {
		return nil, err
	}

	var returnPayload MessagePayload
	dec := json.NewDecoder(bytes.NewReader(m.Payload))
	dec.UseNumber() // So big numbers aren't turned into float64

	// Check the message type and use it to unmarshal the payload
	switch m.Type {
	case RequestMessage:
		var req Request
		err = dec.Decode(&req)
		if err == nil {
			returnPayload = &req
		}
	case ResponseMessage:
		var res Response
		err = dec.Decode(&res)
		if err == nil {
			returnPayload = &res
		}
	case PushMessage:
		var push Push
		err = dec.Decode(&push)
		if err == nil {
			returnPayload = &push
		}
	default:
		err = fmt.Errorf("Received message with invalid type %s", m.Type)
	}

	return returnPayload, err
}
