package peer

import (
	"encoding/json"
	"net"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/ubclaunchpad/cumulus/msg"
)

// PeerStore is a thread-safe container for all the peers we are currently
// connected to. It maps remote peer listen addresses to Peer objects.
type PeerStore struct {
	peers                 map[string]*Peer
	ListenAddr            string
	defaultRequestHandler RequestHandler
	defaultPushHandler    PushHandler
	lock                  *sync.RWMutex
}

// NewPeerStore returns an initialized peerstore.
func NewPeerStore(la string) *PeerStore {
	return &PeerStore{
		peers:      make(map[string]*Peer, 0),
		ListenAddr: la,
		lock:       &sync.RWMutex{},
	}
}

// ConnectionHandler is called when a new connection is opened with us by a
// remote peer. It will create a dispatcher and message handlers to handle
// retrieving messages over the new connection and sending them to App.
func (ps *PeerStore) ConnectionHandler(c net.Conn) {
	// Before we can continue we must exchange listen addresses
	addr, err := exchangeListenAddrs(c, PeerSearchWaitTime, ps.ListenAddr)
	if err != nil {
		log.WithError(err).Error("Failed to retrieve peer listen address")
		return
	} else if ps.Get(addr) != nil || addr == ps.ListenAddr {
		// We are already connected to this peer (or it's us), drop the connection
		c.Close()
		return
	}
	p := New(c, ps, addr)

	// If we are already at MaxPeers, disconnect from a peer to connect to a new
	// one. This way nobody gets choked out of the network because everybody
	// happens to be fully connected.
	if p.Store.Size() >= MaxPeers {
		p.Store.RemoveRandom()
	}
	p.Store.Add(p)

	go p.Dispatch()
	log.Infof("Connected to %s", p.ListenAddr)
}

// Add synchronously adds the given peer to the peerstore
func (ps *PeerStore) Add(p *Peer) {
	ps.lock.Lock()
	defer ps.lock.Unlock()
	ps.peers[p.ListenAddr] = p
}

// Remove synchronously removes the given peer from the peerstore
func (ps *PeerStore) Remove(addr string) {
	ps.lock.Lock()
	defer ps.lock.Unlock()
	delete(ps.peers, addr)
}

// RemoveRandom removes a random peer from the given PeerStore
func (ps *PeerStore) RemoveRandom() {
	ps.lock.Lock()
	defer ps.lock.Unlock()
	for _, p := range ps.peers {
		delete(ps.peers, p.ListenAddr)
		break
	}
}

// Get synchronously retreives the peer with the given id from the peerstore
func (ps *PeerStore) Get(addr string) *Peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()
	return ps.peers[addr]
}

// SetDefaultRequestHandler will ensure that all new peers created who's
// RequestHandlers have not been set will use the given request handler by default
// until it is overridden by the call to SetRequestHandler().
func (ps *PeerStore) SetDefaultRequestHandler(rh RequestHandler) {
	ps.defaultRequestHandler = rh
}

// SetDefaultPushHandler will ensure that all new peers created who's
// PushHandlers have not been set will use the given request handler by default
// until it is overridden by the call to SetPushHandler().
func (ps *PeerStore) SetDefaultPushHandler(ph PushHandler) {
	ps.defaultPushHandler = ph
}

// Broadcast sends the given push message to all peers in the PeerStore at the
// time this function is called. Note that if we fail to write the push message
// to a peer the failure is ignored. Generally this is okay, because push
// messages sent via Broadcast() should be propagated by other peers.
func (ps *PeerStore) Broadcast(push msg.Push) {
	ps.lock.RLock()
	defer ps.lock.RUnlock()
	for _, p := range ps.peers {
		p.Push(push)
	}
}

// PeerInfoHandler will handle the response to a PeerInfo request by attempting
// to establish connections with all new peers in the given response Resource.
func (ps *PeerStore) PeerInfoHandler(res *msg.Response) {
	peers := res.Resource.([]string)
	strPeers, _ := json.Marshal(peers)
	log.Debugf("Found peers %s", string(strPeers))
	for i := 0; i < len(peers) && ps.Size() < MaxPeers; i++ {
		if ps.Get(peers[i]) != nil || peers[i] == ps.ListenAddr {
			// We are already connected to this peer. Skip it.
			continue
		}

		p, err := Connect(peers[i], ps)
		if err != nil {
			log.WithError(err).Errorf("Failed to dial peer %s", peers[i])
		}
		log.Infof("Connected to %s", p.ListenAddr)
	}
}

// Addrs returns the list of addresses of the peers in the peerstore in the form
// <IP addr>:<port>
func (ps *PeerStore) Addrs() []string {
	ps.lock.RLock()
	defer ps.lock.RUnlock()
	addrs := make([]string, 0)
	for addr := range ps.peers {
		addrs = append(addrs, addr)
	}
	return addrs
}

// Size synchornously returns the number of peers in the PeerStore
func (ps *PeerStore) Size() int {
	ps.lock.RLock()
	defer ps.lock.RUnlock()
	return len(ps.peers)
}
