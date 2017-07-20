package msg

import (
	"encoding/json"
	"fmt"
	"io"

	log "github.com/Sirupsen/logrus"
)

type (
	// ResourceType specifies the type of a resource in a message.
	ResourceType int
	// ErrorCode is a code associated with an error
	ErrorCode int
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
	// ResourceNotFound occurs when a node reports the requested resource missing.
	ResourceNotFound = 404
	// NotImplemented occurs when a message or request is received whos response
	// requires functionality that does not yet exist.
	NotImplemented = 501
	// SubnetFull occurs when a stream is opened with a peer whose Subnet is
	// already full.
	SubnetFull = 503
)

// ProtocolError is an error that occured during a request.
type ProtocolError struct {
	Code    ErrorCode `json:"Code"`
	Message string    `json:"Message"`
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
	Type    string `json:"Type"`
	Payload []byte `json:"Payload"`
}

// Request is a container for a request payload, containing a unique request ID,
// the resource type we are requesting, and a Params field specifying request
// parameters. PeerInfo requests should send all info of all peers. Block requests
// should specify block number in parameters.
type Request struct {
	ID           string                 `json:"ID"`
	ResourceType ResourceType           `json:""`
	Params       map[string]interface{} `json:"Params"`
}

// Response is a container for a response payload, containing the unique request
// ID of the request prompting it, an Error (if one occurred), and the requested
// resource (if no error occurred).
type Response struct {
	ID       string         `json:"ID"`
	Error    *ProtocolError `json:"Error"`
	Resource interface{}    `json:"Resource"`
}

// Push is a container for a push payload, containing a resource proactively sent
// to us by another peer.
type Push struct {
	ResourceType ResourceType `json:"ResourceType"`
	Resource     interface{}  `json:"Resource"`
}

// Write encodes and writes the Message into the given Writer.
func (r *Request) Write(w io.Writer) error {
	payload, err := json.Marshal(r)
	if err != nil {
		return err
	}
	msg := Message{
		Type:    "Request",
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
		Type:    "Response",
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
		Type:    "Push",
		Payload: payload,
	}
	return json.NewEncoder(w).Encode(msg)
}

// Read decodes a message from a Reader and returns the message, or an error
// if the read fails. On success, the message payload will be either a Request,
// Response, or Push.
func Read(r io.Reader) (*Request, *Response, *Push, error) {
	var m Message
	err := json.NewDecoder(r).Decode(&m)
	if err != nil {
		return nil, nil, nil, err
	}

	switch m.Type {
	case "Request":
		var req Request
		err = json.Unmarshal([]byte(m.Payload), &req)
		if err == nil {
			log.Debug("Read request ", req)
			return &req, nil, nil, nil
		}
	case "Response":
		var res Response
		err = json.Unmarshal([]byte(m.Payload), &res)
		if err == nil {
			log.Debug("Read response ", res)
			return nil, &res, nil, nil
		}
	case "Push":
		var push Push
		err = json.Unmarshal([]byte(m.Payload), &push)
		if err == nil {
			log.Debug("Read push ", push)
			return nil, nil, &push, nil
		}
	default:
		err = fmt.Errorf("Received message with invalid type %s", m.Type)
	}

	return nil, nil, nil, err
}
