package subnet

import (
	"errors"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/libp2p/go-libp2p-net"
)

const (
	// DefaultMaxPeers is the maximum number of peers a peer can be in
	// communication with at any given time.
	DefaultMaxPeers = 6
)

// Subnet represents the set of peers a peer is connected to at any given time
type Subnet struct {
	peers    map[string]net.Stream
	maxPeers int
	lock     sync.RWMutex
}

// New returns a pointer to an empty Subnet with the given maxPeers.
func New(maxPeers int) *Subnet {
	p := make(map[string]net.Stream)
	sn := Subnet{maxPeers: maxPeers, peers: p, lock: sync.RWMutex{}}
	return &sn
}

// AddPeer adds a peer's multiaddress and the corresponding stream between the
// local peer and the remote peer to the subnet. If the subnet is already
// full, or the multiaddress is not valid, returns error.
func (sn *Subnet) AddPeer(sma string, stream net.Stream) error {
	sn.lock.Lock()
	if sn.Full() {
		sn.lock.Unlock()
		return errors.New("Cannot insert new mapping, Subnet is already full")
	}

	// Check if it's already in this subnet
	if sn.peers[sma] != nil {
		log.Debugf("Peer %s is already in subnet", sma)
	} else {
		log.Debugf("Adding peer %s to subnet", sma)
		sn.peers[sma] = stream
	}
	sn.lock.Unlock()
	return nil
}

// RemovePeer removes the mapping with the key mAddr from the subnet if it
// exists.
func (sn *Subnet) RemovePeer(sma string) {
	sn.lock.Lock()
	log.Debugf("Removing peer %s from subnet", sma)
	delete(sn.peers, sma)
	sn.lock.Unlock()
}

// Stream returns the stream associated with the given multiaddr in this subnet.
// Returns nil if the multiaddr is not in this subnet.
func (sn *Subnet) Stream(sma string) net.Stream {
	sn.lock.RLock()
	stream := sn.peers[sma]
	sn.lock.RUnlock()
	return stream
}

// Full returns true if the number of peers in the sunbet is at or over the
// limit set for that subnet, otherwise returns false.
func (sn *Subnet) Full() bool {
	return len(sn.peers) >= sn.maxPeers
}

// Multiaddrs returns a list of all multiaddresses contined in this subnet
func (sn *Subnet) Multiaddrs() []string {
	sn.lock.RLock()
	mas := make([]string, 0, len(sn.peers))
	for sma := range sn.peers {
		mas = append(mas, sma)
	}
	sn.lock.RUnlock()
	return mas
}
