package message

import (
	"encoding/json"
	"errors"
	"strings"
)

// Message types
// NOTE: because of the way iota works, changing the order in which the
// following constants appear will change their values, which may affect the
// ability of your peer to communicate with others.
const (
	// Send the multiaddress of a peer to another peer
	PeerInfo = iota
	// Request addressess of peers in the remote peer's subnet
	RequestPeerInfo = iota
	// Send information about a block that was just hashed
	NewBlock = iota
	// Request chunk of the blockchain from peer
	RequestChunk = iota
	// Advertise that we have a chunk of the blockchain
	AdvertiseChunk = iota
	// Send information about a new transaction to another peer
	Transaction = iota
)

// Message is a container for information and its type that is
// sent between Cumulus peers.
type Message struct {
	msgType int
	content []byte
}

// New returns a pointer to a message initialized with a byte array
// of content and a message type, or an error if the type is not one
// of those defined above.
func New(c []byte, t int) (*Message, error) {
	switch t {
	case PeerInfo:
	case NewBlock:
	case RequestChunk:
	case AdvertiseChunk:
	case Transaction:
		break
	default:
		return nil, errors.New("Invalid message type")
	}

	m := &Message{msgType: t, content: c}
	return m, nil
}

// Bytes returns JSON representation of message as a byte array, or error if
// message cannot be marshalled.
func (m *Message) Bytes() ([]byte, error) {
	return json.Marshal(m)
}

// FromString parses a message in the form of a string and returns a pointer
// to a new Message struct made from the contents of the string. Returns error
// if string is malformed.
func FromString(s string) (*Message, error) {
	var msg Message
	s = strings.TrimSpace(s)
	err := json.Unmarshal([]byte(s), &msg)
	return &msg, err
}

// Type returns msgType for message
func (m *Message) Type() int {
	return m.msgType
}
