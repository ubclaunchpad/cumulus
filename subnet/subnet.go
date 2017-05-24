package subnet

import (
	"errors"

	log "github.com/Sirupsen/logrus"
	net "github.com/libp2p/go-libp2p-net"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	// DefaultMaxPeers is the maximum number of peers a peer can be in
	// communication with at any given time.
	DefaultMaxPeers = 6
)

// Subnet represents the set of peers a peer is connected to at any given time
type Subnet struct {
	peers    map[ma.Multiaddr]net.Stream
	maxPeers int
}

// New returns a pointer to an empty Subnet with the given maxPeers.
func New(maxPeers int) *Subnet {
	p := make(map[ma.Multiaddr]net.Stream)
	sn := Subnet{maxPeers: maxPeers, peers: p}
	return &sn
}

// AddPeer adds a peer's multiaddress and the corresponding stream between the
// local peer and the remote peer to the subnet. If the subnet is already
// full, or the multiaddress is not valid, returns error.
func (sn *Subnet) AddPeer(mAddr ma.Multiaddr, stream net.Stream) error {
	if sn.Full() {
		return errors.New("Cannot insert new mapping, Subnet is already full")
	}

	// Validate the multiaddress
	mAddr, err := ma.NewMultiaddr(mAddr.String())
	if err != nil {
		return err
	}

	// Check if it's already in this subnet
	if sn.peers[mAddr] != nil {
		log.Debugf("Peer %s is already in subnet", mAddr.String())
	} else {
		log.Debugf("Adding peer %s to subnet", mAddr.String())
		sn.peers[mAddr] = stream
	}
	return nil
}

// RemovePeer removes the mapping with the key mAddr from the subnet if it
// exists.
func (sn *Subnet) RemovePeer(mAddr ma.Multiaddr) {
	log.Debugf("Removing peer %s from subnet", mAddr.String())
	delete(sn.peers, mAddr)
}

// Full returns true if the number of peers in the sunbet is at or over the
// limit set for that subnet, otherwise returns false.
func (sn *Subnet) Full() bool {
	return len(sn.peers) >= sn.maxPeers
}

// Multiaddrs returns a list of all multiaddresses contined in this subnet
func (sn *Subnet) Multiaddrs() []ma.Multiaddr {
	mAddrs := make([]ma.Multiaddr, 0, len(sn.peers))
	for mAddr := range sn.peers {
		mAddrs = append(mAddrs, mAddr)
	}
	return mAddrs
}

// StringMultiaddrs returns a list of all multiaddrs in this subnet as strings
func (sn *Subnet) StringMultiaddrs() []string {
	mAddrs := make([]string, 0, len(sn.peers))
	for mAddr := range sn.peers {
		mAddrs = append(mAddrs, mAddr.String())
	}
	return mAddrs
}
