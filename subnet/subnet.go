package subnet

import (
	"errors"

	net "github.com/libp2p/go-libp2p-net"
	ma "github.com/multiformats/go-multiaddr"
	msg "github.com/ubclaunchpad/cumulus/message"
)

const (
	// DefaultMaxPeers is the maximum number of peers a peer can be in
	// communication with at any given time.
	DefaultMaxPeers = 6
)

// Subnet represents the set of peers a peer is connected to at any given time
type Subnet struct {
	peers    map[ma.Multiaddr]net.Stream
	maxPeers uint16
	numPeers uint16
}

// New returns a pointer to an empty Subnet with the given maxPeers.
func New(maxPeers uint16) *Subnet {
	p := make(map[ma.Multiaddr]net.Stream)
	sn := Subnet{maxPeers: maxPeers, numPeers: 0, peers: p}
	return &sn
}

// AddPeer adds a peer's multiaddress and the corresponding stream between the
// local peer and the remote peer to the subnet. If the subnet is already
// full, or the multiaddress is not valid, returns error.
func (sn *Subnet) AddPeer(mAddr ma.Multiaddr, stream net.Stream) error {
	if sn.isFull() {
		return errors.New("Cannot insert new mapping, Subnet is already full")
	}

	// Validate the multiaddress
	mAddr, err := ma.NewMultiaddr(mAddr.String())
	if err != nil {
		return err
	}

	sn.peers[mAddr] = stream
	sn.numPeers++
	return nil
}

// RemovePeer removes the mapping with the key mAddr from the subnet if it
// exists.
func (sn *Subnet) RemovePeer(mAddr ma.Multiaddr) {
	delete(sn.peers, mAddr)
	sn.numPeers--
}

// isFull returns true if the number of peers in the sunbet is at or over the
// limit set for that subnet, otherwise returns false.
func (sn *Subnet) isFull() bool {
	return sn.numPeers >= sn.maxPeers
}

// Broadcast sends information to all peers we are currently connected to
func (sn *Subnet) Broadcast(m msg.Message) error {
	return errors.New("Function not implemented")
}

// Listen listens to all peers we are currently connected to
// Call appropriate routines in response to new messages
func (sn *Subnet) Listen() {
	panic("Function not implemented")
}
