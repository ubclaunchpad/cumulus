package message

import (
	"encoding/gob"
	"io"

	"github.com/google/uuid"
	ma "github.com/multiformats/go-multiaddr"
)

type (
	// Type specifies the type of a message.
	Type int
	// ResourceType specifies the type of a resource in a message.
	ResourceType int
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

// Message is a container for messages, containing a type and either a Request,
// Response, or Push in the payload.
type Message struct {
	Type    Type
	Payload interface{}
}

// New returns a new Message.
func New(t Type, payload interface{}) *Message {
	return &Message{
		Type:    t,
		Payload: payload,
	}
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

// NewRequestMessage returns a new Message continaing a request with the given
// parameters.
func NewRequestMessage(rt ResourceType, p map[string]interface{}) *Message {
	req := Request{
		ID:           uuid.New().String(),
		ResourceType: rt,
		Params:       p,
	}
	return New(MessageRequest, req)
}

// Response is a container for a response payload, containing the unique request
// ID of the request prompting it, an Error (if one occurred), and the requested
// resource (if no error occurred).
type Response struct {
	ID       string
	Error    error
	Resource interface{}
}

// NewResponseMessage returns a new Message continaing a response with the given
// parameters. ResponseMessages should have the same ID as that of the request
// message they are a response to.
func NewResponseMessage(id string, err error, resource interface{}) *Message {
	res := &Response{
		ID:       id,
		Error:    err,
		Resource: resource,
	}
	return New(MessageResponse, res)
}

// Push is a container for a push payload, containing a resource proactively sent
// to us by another peer.
type Push struct {
	ResourceType ResourceType
	Resource     interface{}
}

// NewPushMessage returns a new Message continaing a response with the given
// parameters.
func NewPushMessage(rt ResourceType, resource interface{}) *Message {
	res := &Push{
		ResourceType: rt,
		Resource:     resource,
	}
	return New(MessageResponse, res)
}

// Write encodes and writes the Message into the given Writer.
func (m *Message) Write(w io.Writer) error {
	return gob.NewEncoder(w).Encode(m)
}

// Read decodes a message from a Reader and returns it.
func Read(r io.Reader) (*Message, error) {
	var m Message
	err := gob.NewDecoder(r).Decode(&m)
	return &m, err
}

// RegisterGobTypes registers all the types used by gob in reading and writing
// messages. This should only be called once during initializaiton.
func init() {
	dummy, _ := ma.NewMultiaddr("")
	gob.Register(dummy)
	gob.Register(Request{})
	gob.Register(Response{})
	gob.Register(Push{})
}
