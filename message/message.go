package message

import (
	"errors"
	"fmt"
)

// Message types
// NOTE: because of the way iota works, changing the order in which the
// following constants appear will change their values, which may affect the
// ability of your peer to communicate with others.
const (
	// Send the multiaddress of a peer to another peer
	PeerInfo = iota
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

// Bytes returns the given message in []byte format
func (m *Message) Bytes() []byte {
	return []byte(fmt.Sprintf("%d:%s", m.msgType, string(m.content)))
}
